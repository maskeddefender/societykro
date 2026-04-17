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

	"github.com/societykro/notice-service/internal/handler"
	"github.com/societykro/notice-service/internal/repository"
	"github.com/societykro/notice-service/internal/service"
)

func main() {
	cfg := config.Load()
	port := os.Getenv("NOTICE_SERVICE_PORT")
	if port == "" {
		port = "8085"
	}

	logger.Init(cfg.App.Env)
	log := &logger.Log
	log.Info().Str("service", "notice-service").Msg("Starting")

	pool, err := database.NewPostgresPool(cfg.Database.DSN(), cfg.Database.MaxConnections)
	if err != nil {
		log.Fatal().Err(err).Msg("PostgreSQL connection failed")
	}
	defer pool.Close()

	rdb, err := database.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Redis connection failed")
	}
	defer rdb.Close()

	jwtMgr, err := auth.NewJWTManager(
		cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
		cfg.JWT.AccessTokenExpiry, cfg.JWT.RefreshTokenExpiry,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("JWT manager failed")
	}

	bus, err := events.NewBus(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("NATS connection failed")
	}
	defer bus.Close()
	bus.EnsureStream(events.StreamSocietyKro, events.AllSubjects())

	repo := repository.NewNoticeRepository(pool)
	svc := service.NewNoticeService(repo, bus)
	h := handler.NewNoticeHandler(svc)

	app := fiber.New(fiber.Config{
		AppName:      "SocietyKro Notice Service",
		ServerHeader: "SocietyKro",
		BodyLimit:    4 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error().Err(err).Str("path", c.Path()).Msg("Unhandled error")
			return response.InternalError(c, "Something went wrong")
		},
	})

	app.Use(fiberRecover.New())
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "notice-service"})
	})

	api := app.Group("/api/v1", middleware.JWTMiddleware(jwtMgr, rdb))
	notices := api.Group("/notices")
	notices.Post("/", h.Create)
	notices.Get("/", h.List)
	notices.Get("/:id", h.GetByID)
	notices.Post("/:id/read", h.MarkRead)
	notices.Delete("/:id", h.Delete)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-quit; log.Info().Msg("Shutting down..."); app.Shutdown() }()

	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("addr", addr).Msg("notice-service listening")
	if err := app.Listen(addr); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
