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

	"github.com/societykro/vendor-service/internal/handler"
	"github.com/societykro/vendor-service/internal/repository"
	"github.com/societykro/vendor-service/internal/service"
)

func main() {
	cfg := config.Load()
	port := getEnvOr("VENDOR_SERVICE_PORT", "8086")

	logger.Init(cfg.App.Env)
	log := &logger.Log
	log.Info().Str("service", "vendor-service").Str("env", cfg.App.Env).Msg("Starting")

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
	repo := repository.NewVendorRepository(pool)
	svc := service.NewVendorService(repo, bus)
	h := handler.NewVendorHandler(svc)

	// Fiber
	app := fiber.New(fiber.Config{
		AppName:      "SocietyKro Vendor Service",
		ServerHeader: "SocietyKro",
		BodyLimit:    10 * 1024 * 1024, // 10MB
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error().Err(err).Str("path", c.Path()).Msg("Unhandled error")
			return response.InternalError(c, "Something went wrong")
		},
	})

	app.Use(fiberRecover.New())
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "vendor-service"})
	})

	// All routes require JWT
	api := app.Group("/api/v1", middleware.JWTMiddleware(jwtMgr, rdb))

	// Vendor routes — CRUD is admin-only, read is open to authenticated users
	vendors := api.Group("/vendors")
	vendors.Get("/", h.ListVendors)
	vendors.Get("/:id", h.GetVendor)
	vendors.Post("/", middleware.RequireAdmin(), h.CreateVendor)
	vendors.Put("/:id", middleware.RequireAdmin(), h.UpdateVendor)
	vendors.Delete("/:id", middleware.RequireAdmin(), h.DeleteVendor)

	// Domestic help routes
	domestic := api.Group("/domestic-help")
	domestic.Post("/", h.CreateDomesticHelp)
	domestic.Get("/", h.ListDomesticHelp)
	domestic.Get("/:id", h.GetDomesticHelp)
	domestic.Post("/:id/link-flat", h.LinkToFlat)
	domestic.Post("/:id/attendance", h.LogAttendance)
	domestic.Put("/attendance/:id/exit", h.LogExit)
	domestic.Get("/:id/attendance", h.GetAttendance)

	// Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("Shutting down vendor-service...")
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("addr", addr).Msg("vendor-service listening")
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
