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

	"github.com/societykro/go-common/config"
	"github.com/societykro/go-common/database"
	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/message-router/internal/handler"
	"github.com/societykro/message-router/internal/service"
)

func main() {
	cfg := config.Load()
	port := os.Getenv("MESSAGE_ROUTER_PORT")
	if port == "" {
		port = "8089"
	}

	logger.Init(cfg.App.Env)
	log := &logger.Log
	log.Info().Str("service", "message-router").Msg("Starting")

	// PostgreSQL (for user phone/chat_id lookups)
	pool, err := database.NewPostgresPool(cfg.Database.DSN(), cfg.Database.MaxConnections)
	if err != nil {
		log.Fatal().Err(err).Msg("PostgreSQL connection failed")
	}
	defer pool.Close()

	// NATS (publish normalized messages for chatbot)
	bus, err := events.NewBus(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("NATS connection failed")
	}
	defer bus.Close()

	// Ensure message stream exists
	bus.EnsureStream("MESSAGES", []string{"message.*"})

	// Layers
	router := service.NewRouter(pool, bus)
	webhookH := handler.NewWebhookHandler(router)

	// Fiber (no JWT — webhooks use their own signature verification)
	app := fiber.New(fiber.Config{
		AppName:      "SocietyKro Message Router",
		ServerHeader: "SocietyKro",
		BodyLimit:    4 * 1024 * 1024,
	})

	app.Use(fiberRecover.New())
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "message-router"})
	})

	// WhatsApp webhooks (Meta requires specific verify endpoint)
	app.Get("/webhook/whatsapp", webhookH.WhatsAppVerify)
	app.Post("/webhook/whatsapp", webhookH.WhatsAppWebhook)

	// Telegram webhook
	app.Post("/webhook/telegram", webhookH.TelegramWebhook)

	// Dev test endpoint
	if cfg.App.Env == "development" {
		app.Post("/webhook/test", webhookH.SendTestMessage)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-quit; log.Info().Msg("Shutting down..."); app.Shutdown() }()

	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("addr", addr).Msg("message-router listening")
	if err := app.Listen(addr); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
