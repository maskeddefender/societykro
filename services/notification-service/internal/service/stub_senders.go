package service

import (
	"context"

	"github.com/societykro/go-common/logger"
)

// StubPushSender logs push notifications instead of sending them.
// Replace with real FCM implementation in production.
type StubPushSender struct{}

func (s *StubPushSender) Send(_ context.Context, token, title, body string, _ map[string]string) error {
	logger.Log.Info().Str("token", token[:20]+"...").Str("title", title).Msg("[STUB] Push notification")
	return nil
}

// StubWhatsAppSender logs WhatsApp messages instead of sending them.
// Replace with Meta Cloud API implementation in production.
type StubWhatsAppSender struct{}

func (s *StubWhatsAppSender) Send(_ context.Context, phone, message string, _ map[string]string) error {
	logger.Log.Info().Str("phone", phone[:4]+"****").Str("msg", message[:50]).Msg("[STUB] WhatsApp message")
	return nil
}

// StubTelegramSender logs Telegram messages instead of sending them.
// Replace with Telegram Bot API implementation in production.
type StubTelegramSender struct{}

func (s *StubTelegramSender) Send(_ context.Context, chatID, text string) error {
	logger.Log.Info().Str("chat_id", chatID).Str("text", text[:50]).Msg("[STUB] Telegram message")
	return nil
}

// StubSMSSender logs SMS messages instead of sending them.
// Replace with MSG91 implementation in production.
type StubSMSSender struct{}

func (s *StubSMSSender) Send(_ context.Context, phone, message string) error {
	logger.Log.Info().Str("phone", phone[:4]+"****").Str("msg", message[:50]).Msg("[STUB] SMS")
	return nil
}
