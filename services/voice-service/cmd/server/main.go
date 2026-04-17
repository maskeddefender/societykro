// Package main is the entry point for the SocietyKro voice service.
// It wires configuration, JWT auth, Redis, the Bhashini client and HTTP
// handlers into a Fiber application.
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

	"github.com/societykro/voice-service/internal/handler"
	"github.com/societykro/voice-service/internal/service"
)

func main() {
	cfg := config.Load()

	port := os.Getenv("VOICE_SERVICE_PORT")
	if port == "" {
		port = "8090"
	}

	logger.Init(cfg.App.Env)
	log := &logger.Log
	log.Info().Str("service", "voice-service").Msg("Starting")

	// Redis (used for JWT blacklist checks in auth middleware).
	rdb, err := database.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Redis connection failed")
	}
	defer rdb.Close()

	// JWT manager for token validation.
	jwtMgr, err := auth.NewJWTManager(
		cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
		cfg.JWT.AccessTokenExpiry, cfg.JWT.RefreshTokenExpiry,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("JWT manager failed")
	}

	// Bhashini API client.
	bhashiniUserID := os.Getenv("BHASHINI_USER_ID")
	bhashiniAPIKey := os.Getenv("BHASHINI_API_KEY")
	bhashiniPipelineURL := os.Getenv("BHASHINI_PIPELINE_URL")

	bhashiniClient := service.NewBhashiniClient(bhashiniUserID, bhashiniAPIKey, bhashiniPipelineURL)
	voiceSvc := service.NewVoiceService(bhashiniClient)
	h := handler.NewVoiceHandler(voiceSvc)

	// Fiber app.
	app := fiber.New(fiber.Config{
		AppName:      "SocietyKro Voice Service",
		ServerHeader: "SocietyKro",
		BodyLimit:    10 * 1024 * 1024, // 10 MB for audio payloads
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error().Err(err).Str("path", c.Path()).Msg("Unhandled error")
			return response.InternalError(c, "Something went wrong")
		},
	})

	app.Use(fiberRecover.New())
	app.Use(requestid.New())
	app.Use(cors.New())

	// Health check (public).
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "voice-service"})
	})

	api := app.Group("/api/v1")

	// Public endpoint -- no JWT required.
	api.Get("/languages", h.GetSupportedLanguages)

	// Authenticated endpoints.
	voice := api.Group("/voice", middleware.JWTMiddleware(jwtMgr, rdb))
	voice.Post("/transcribe", h.Transcribe)
	voice.Post("/translate", h.Translate)
	voice.Post("/detect-language", h.DetectLanguage)

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-quit; log.Info().Msg("Shutting down..."); _ = app.Shutdown() }()

	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("addr", addr).Msg("voice-service listening")
	if err := app.Listen(addr); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
