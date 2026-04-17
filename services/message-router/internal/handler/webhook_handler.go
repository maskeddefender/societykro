package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/societykro/go-common/logger"
	"github.com/societykro/go-common/response"

	"github.com/societykro/message-router/internal/model"
	"github.com/societykro/message-router/internal/service"
)

// WebhookHandler handles incoming webhooks from WhatsApp and Telegram.
type WebhookHandler struct {
	router *service.Router
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(router *service.Router) *WebhookHandler {
	return &WebhookHandler{router: router}
}

// WhatsAppVerify handles the GET webhook verification from Meta.
// Meta sends hub.mode, hub.verify_token, hub.challenge.
// GET /webhook/whatsapp
func (h *WebhookHandler) WhatsAppVerify(c *fiber.Ctx) error {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	verifyToken := os.Getenv("META_WA_VERIFY_TOKEN")
	if mode == "subscribe" && token == verifyToken {
		logger.Log.Info().Msg("WhatsApp webhook verified")
		return c.SendString(challenge)
	}

	return c.Status(fiber.StatusForbidden).SendString("Forbidden")
}

// WhatsAppWebhook receives incoming WhatsApp messages.
// POST /webhook/whatsapp
func (h *WebhookHandler) WhatsAppWebhook(c *fiber.Ctx) error {
	// Verify signature from Meta
	if !h.verifyWhatsAppSignature(c) {
		logger.Log.Warn().Msg("Invalid WhatsApp webhook signature")
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid signature")
	}

	var payload model.WhatsAppWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to parse WA webhook")
		return c.SendStatus(fiber.StatusOK) // Always 200 to Meta
	}

	// Process async — always return 200 quickly to Meta
	go func() {
		if err := h.router.HandleWhatsApp(c.Context(), payload); err != nil {
			logger.Log.Error().Err(err).Msg("WA message handling failed")
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// TelegramWebhook receives incoming Telegram updates.
// POST /webhook/telegram
func (h *WebhookHandler) TelegramWebhook(c *fiber.Ctx) error {
	// Verify secret token header
	secret := c.Get("X-Telegram-Bot-Api-Secret-Token")
	expectedSecret := os.Getenv("TELEGRAM_WEBHOOK_SECRET")
	if expectedSecret != "" && secret != expectedSecret {
		logger.Log.Warn().Msg("Invalid Telegram webhook secret")
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid secret")
	}

	var update model.TelegramUpdate
	if err := c.BodyParser(&update); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to parse TG webhook")
		return c.SendStatus(fiber.StatusOK)
	}

	go func() {
		if err := h.router.HandleTelegram(c.Context(), update); err != nil {
			logger.Log.Error().Err(err).Msg("TG message handling failed")
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// SendTestMessage is a dev endpoint to simulate an incoming message.
// POST /webhook/test
func (h *WebhookHandler) SendTestMessage(c *fiber.Ctx) error {
	var msg model.IncomingMessage
	if err := c.BodyParser(&msg); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.router.HandleWhatsApp(c.Context(), model.WhatsAppWebhookPayload{}); err != nil {
		return response.InternalError(c, "Failed to process test message")
	}

	return response.OKMessage(c, "Test message processed")
}

// verifyWhatsAppSignature validates the X-Hub-Signature-256 header from Meta.
func (h *WebhookHandler) verifyWhatsAppSignature(c *fiber.Ctx) bool {
	appSecret := os.Getenv("META_WA_APP_SECRET")
	if appSecret == "" {
		return true // Skip verification in dev if no secret configured
	}

	signature := c.Get("X-Hub-Signature-256")
	if signature == "" {
		return false
	}

	// Remove "sha256=" prefix
	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(c.Body())
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}
