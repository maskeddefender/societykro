package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/societykro/go-common/auth"
	"github.com/societykro/go-common/config"
	"github.com/societykro/go-common/database"
	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"
	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/complaint-service/internal/handler"
	"github.com/societykro/complaint-service/internal/repository"
	"github.com/societykro/complaint-service/internal/service"
)

func main() {
	cfg := config.Load()
	port := getEnvOr("COMPLAINT_SERVICE_PORT", "8082")

	logger.Init(cfg.App.Env)
	log := &logger.Log
	log.Info().Str("service", "complaint-service").Str("env", cfg.App.Env).Msg("Starting")

	// PostgreSQL
	pool, err := database.NewPostgresPool(cfg.Database.DSN(), cfg.Database.MaxConnections)
	if err != nil {
		log.Fatal().Err(err).Msg("PostgreSQL connection failed")
	}
	defer pool.Close()

	// Redis
	rdb, err := database.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Redis connection failed")
	}
	defer rdb.Close()

	// JWT (for middleware validation only — no token generation here)
	jwtMgr, err := auth.NewJWTManager(
		cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
		cfg.JWT.AccessTokenExpiry, cfg.JWT.RefreshTokenExpiry,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("JWT manager failed")
	}

	// NATS
	bus, err := events.NewBus(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("NATS connection failed")
	}
	defer bus.Close()

	if err := bus.EnsureStream(events.StreamSocietyKro, events.AllSubjects()); err != nil {
		log.Fatal().Err(err).Msg("NATS stream creation failed")
	}

	// Layers
	repo := repository.NewComplaintRepository(pool)
	svc := service.NewComplaintService(repo, bus)
	h := handler.NewComplaintHandler(svc)

	// Fiber
	app := fiber.New(fiber.Config{
		AppName:      "SocietyKro Complaint Service",
		ServerHeader: "SocietyKro",
		BodyLimit:    10 * 1024 * 1024, // 10MB (images)
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error().Err(err).Str("path", c.Path()).Msg("Unhandled error")
			return response.InternalError(c, "Something went wrong")
		},
	})

	app.Use(fiberRecover.New())
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "complaint-service"})
	})

	// All routes require JWT
	api := app.Group("/api/v1", middleware.JWTMiddleware(jwtMgr, rdb))

	complaints := api.Group("/complaints")
	complaints.Post("/", h.Create)
	complaints.Get("/", h.List)
	complaints.Get("/analytics", h.GetStats)
	complaints.Get("/:id", h.GetByID)
	complaints.Put("/:id/status", h.UpdateStatus)
	complaints.Put("/:id/assign", h.AssignVendor)
	complaints.Post("/:id/rate", h.Rate)
	complaints.Post("/:id/comments", h.AddComment)
	complaints.Get("/:id/comments", h.ListComments)

	// Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("Shutting down complaint-service...")
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("addr", addr).Msg("complaint-service listening")
	if err := app.Listen(addr); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
