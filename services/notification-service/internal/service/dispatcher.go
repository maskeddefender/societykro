package service

import (
	"context"
	"fmt"

	"github.com/societykro/go-common/logger"

	"github.com/societykro/notification-service/internal/model"
)

// Dispatcher sends notifications through various channels.
// Each channel (push, WhatsApp, Telegram, SMS) has a sender interface.
type Dispatcher struct {
	push     PushSender
	whatsapp WhatsAppSender
	telegram TelegramSender
	sms      SMSSender
}

// PushSender sends push notifications via FCM.
type PushSender interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
}

// WhatsAppSender sends WhatsApp messages via Meta Cloud API.
type WhatsAppSender interface {
	Send(ctx context.Context, phone, template string, params map[string]string) error
}

// TelegramSender sends Telegram messages via Bot API.
type TelegramSender interface {
	Send(ctx context.Context, chatID, text string) error
}

// SMSSender sends SMS via MSG91 or Twilio.
type SMSSender interface {
	Send(ctx context.Context, phone, message string) error
}

// NewDispatcher creates a new Dispatcher. Pass nil for channels not yet implemented.
func NewDispatcher(push PushSender, wa WhatsAppSender, tg TelegramSender, sms SMSSender) *Dispatcher {
	return &Dispatcher{push: push, whatsapp: wa, telegram: tg, sms: sms}
}

// Dispatch sends a notification to a target through the specified channels.
func (d *Dispatcher) Dispatch(ctx context.Context, target model.NotificationTarget, title, body string, channels []string, data map[string]string) {
	for _, ch := range channels {
		switch ch {
		case "push":
			if target.FCMToken != nil && d.push != nil {
				if err := d.push.Send(ctx, *target.FCMToken, title, body, data); err != nil {
					logger.Log.Error().Err(err).Str("user", target.UserID.String()).Msg("Push send failed")
				} else {
					logger.Log.Debug().Str("user", target.Name).Msg("Push sent")
				}
			}

		case "whatsapp":
			if target.WhatsappOptIn && d.whatsapp != nil {
				if err := d.whatsapp.Send(ctx, target.Phone, body, nil); err != nil {
					logger.Log.Error().Err(err).Str("user", target.UserID.String()).Msg("WhatsApp send failed")
				} else {
					logger.Log.Debug().Str("user", target.Name).Msg("WhatsApp sent")
				}
			}

		case "telegram":
			if target.TelegramChatID != nil && d.telegram != nil {
				if err := d.telegram.Send(ctx, *target.TelegramChatID, fmt.Sprintf("*%s*\n%s", title, body)); err != nil {
					logger.Log.Error().Err(err).Str("user", target.UserID.String()).Msg("Telegram send failed")
				} else {
					logger.Log.Debug().Str("user", target.Name).Msg("Telegram sent")
				}
			}

		case "sms":
			if d.sms != nil {
				if err := d.sms.Send(ctx, target.Phone, body); err != nil {
					logger.Log.Error().Err(err).Str("user", target.UserID.String()).Msg("SMS send failed")
				} else {
					logger.Log.Debug().Str("user", target.Name).Msg("SMS sent")
				}
			}
		}
	}
}
