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
	"github.com/societykro/go-common/logger"
	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/auth-service/internal/handler"
	"github.com/societykro/auth-service/internal/repository"
	"github.com/societykro/auth-service/internal/service"
)

func main() {
	cfg := config.Load()

	port := cfg.App.Port
	if p := os.Getenv("AUTH_SERVICE_PORT"); p != "" {
		port = p
	}

	// Logger
	logger.Init(cfg.App.Env)
	log := &logger.Log
	log.Info().Str("service", "auth-service").Str("env", cfg.App.Env).Msg("Starting")

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

	// JWT Manager
	jwtMgr, err := auth.NewJWTManager(
		cfg.JWT.PrivateKeyPath,
		cfg.JWT.PublicKeyPath,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("JWT manager initialization failed")
	}

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	societyRepo := repository.NewSocietyRepository(pool)

	// Service
	authSvc := service.NewAuthService(userRepo, societyRepo, rdb, jwtMgr, cfg)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	societyH := handler.NewSocietyHandler(authSvc)

	// Fiber
	app := fiber.New(fiber.Config{
		AppName:       "SocietyKro Auth Service",
		ServerHeader:  "SocietyKro",
		CaseSensitive: true,
		StrictRouting: true,
		BodyLimit:     4 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error().Err(err).Str("path", c.Path()).Str("method", c.Method()).Msg("Unhandled error")
			return response.InternalError(c, "Something went wrong")
		},
	})

	// Global middleware
	app.Use(fiberRecover.New())
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",
	}))

	// Health check (public)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "auth-service",
			"version": "0.2.0",
		})
	})

	api := app.Group("/api/v1")

	// --- Public routes (no auth required) ---
	authRoutes := api.Group("/auth")
	authRoutes.Post("/otp/send", authH.SendOTP)
	authRoutes.Post("/otp/verify", authH.VerifyOTP)
	authRoutes.Post("/refresh", authH.RefreshToken)

	// --- Protected routes (JWT required) ---
	protected := api.Use(middleware.JWTMiddleware(jwtMgr, rdb))

	// Auth - protected
	protectedAuth := protected.Group("/auth")
	protectedAuth.Get("/me", authH.GetProfile)
	protectedAuth.Put("/me/fcm-token", authH.UpdateFCMToken)
	protectedAuth.Post("/logout", authH.Logout)

	// Society - protected
	societies := protected.Group("/societies")
	societies.Post("/", societyH.CreateSociety)
	societies.Post("/join", societyH.JoinSociety)
	societies.Get("/:id", societyH.GetSociety)
	societies.Get("/:id/flats", societyH.ListFlats)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("Shutting down auth-service...")
		_ = app.Shutdown()
	}()

	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("addr", addr).Msg("auth-service listening")
	if err := app.Listen(addr); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
