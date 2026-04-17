package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/message-router/internal/model"
)

// Router normalizes incoming messages from WhatsApp/Telegram,
// resolves user identity, and publishes to the NATS event bus
// for the chatbot-service to process.
type Router struct {
	pool *pgxpool.Pool
	bus  *events.Bus
}

// NewRouter creates a new message Router.
func NewRouter(pool *pgxpool.Pool, bus *events.Bus) *Router {
	return &Router{pool: pool, bus: bus}
}

// HandleWhatsApp parses a WhatsApp webhook payload and routes the message.
func (r *Router) HandleWhatsApp(ctx context.Context, payload model.WhatsAppWebhookPayload) error {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				incoming := model.IncomingMessage{
					ID:          msg.ID,
					Source:      "whatsapp",
					SenderID:    msg.From,
					SenderPhone: "+" + msg.From,
					Timestamp:   parseTimestamp(msg.Timestamp),
				}

				switch msg.Type {
				case "text":
					if msg.Text != nil {
						incoming.MessageType = "text"
						incoming.Text = msg.Text.Body
					}
				case "audio":
					incoming.MessageType = "voice"
					if msg.Audio != nil {
						incoming.VoiceURL = msg.Audio.ID // Media ID, download later
					}
				case "image":
					incoming.MessageType = "image"
					if msg.Image != nil {
						incoming.ImageURL = msg.Image.ID
						if msg.Image.Caption != "" {
							incoming.Text = msg.Image.Caption
						}
					}
				case "interactive":
					incoming.MessageType = "text"
					if msg.Interactive != nil {
						if msg.Interactive.ButtonReply != nil {
							incoming.Text = msg.Interactive.ButtonReply.ID
						} else if msg.Interactive.ListReply != nil {
							incoming.Text = msg.Interactive.ListReply.ID
						}
					}
				default:
					logger.Log.Debug().Str("type", msg.Type).Msg("Unsupported WA message type")
					continue
				}

				if err := r.resolveAndPublish(ctx, &incoming); err != nil {
					logger.Log.Error().Err(err).Str("from", msg.From).Msg("Failed to route WA message")
				}
			}
		}
	}
	return nil
}

// HandleTelegram parses a Telegram update and routes the message.
func (r *Router) HandleTelegram(ctx context.Context, update model.TelegramUpdate) error {
	if update.Message == nil && update.CallbackQuery == nil {
		return nil
	}

	incoming := model.IncomingMessage{
		Source:    "telegram",
		Timestamp: time.Now().UTC(),
	}

	if update.Message != nil {
		incoming.ID = fmt.Sprintf("tg_%d", update.Message.MessageID)
		incoming.SenderID = fmt.Sprintf("%d", update.Message.Chat.ID)

		if update.Message.Text != "" {
			incoming.MessageType = "text"
			incoming.Text = update.Message.Text
		} else if update.Message.Voice != nil {
			incoming.MessageType = "voice"
			incoming.VoiceURL = update.Message.Voice.FileID
		} else if len(update.Message.Photo) > 0 {
			incoming.MessageType = "image"
			incoming.ImageURL = update.Message.Photo[len(update.Message.Photo)-1].FileID
		} else {
			return nil
		}
	}

	if update.CallbackQuery != nil {
		incoming.ID = update.CallbackQuery.ID
		incoming.MessageType = "text"
		incoming.Text = update.CallbackQuery.Data
		if update.CallbackQuery.Message != nil {
			incoming.SenderID = fmt.Sprintf("%d", update.CallbackQuery.Message.Chat.ID)
		}
	}

	return r.resolveAndPublish(ctx, &incoming)
}

// resolveAndPublish looks up the user by phone/telegram_chat_id,
// enriches the message with user context, and publishes to NATS.
func (r *Router) resolveAndPublish(ctx context.Context, msg *model.IncomingMessage) error {
	// Resolve user identity from DB
	var userID, societyID, flatID, role string

	switch msg.Source {
	case "whatsapp":
		err := r.pool.QueryRow(ctx,
			`SELECT u.id, usm.society_id, usm.flat_id, usm.role
			 FROM app_user u
			 JOIN user_society_membership usm ON usm.user_id = u.id AND usm.is_active = true
			 WHERE u.phone = $1 AND u.is_active = true
			 LIMIT 1`, msg.SenderPhone,
		).Scan(&userID, &societyID, &flatID, &role)
		if err != nil {
			logger.Log.Warn().Str("phone", msg.SenderPhone[:4]+"****").Msg("Unknown WhatsApp user")
			// Still publish — chatbot can handle unknown users with onboarding flow
		}

	case "telegram":
		err := r.pool.QueryRow(ctx,
			`SELECT u.id, usm.society_id, usm.flat_id, usm.role
			 FROM app_user u
			 JOIN user_society_membership usm ON usm.user_id = u.id AND usm.is_active = true
			 WHERE u.telegram_chat_id = $1 AND u.is_active = true
			 LIMIT 1`, msg.SenderID,
		).Scan(&userID, &societyID, &flatID, &role)
		if err != nil {
			logger.Log.Warn().Str("chat_id", msg.SenderID).Msg("Unknown Telegram user")
		}
	}

	msg.UserID = userID
	msg.SocietyID = societyID
	msg.FlatID = flatID
	msg.Role = role

	// Publish to NATS for chatbot-service to pick up
	if err := r.bus.Publish("message.received", "message.received", msg); err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	logger.Log.Info().
		Str("source", msg.Source).
		Str("type", msg.MessageType).
		Str("user_id", userID).
		Msg("Message routed to chatbot")

	return nil
}

func parseTimestamp(ts string) time.Time {
	unix, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Now().UTC()
	}
	return time.Unix(unix, 0).UTC()
}
