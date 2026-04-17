package service

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/notification-service/internal/model"
)

// EventHandler processes NATS events and dispatches notifications.
type EventHandler struct {
	pool       *pgxpool.Pool
	dispatcher *Dispatcher
}

// NewEventHandler creates a new EventHandler.
func NewEventHandler(pool *pgxpool.Pool, dispatcher *Dispatcher) *EventHandler {
	return &EventHandler{pool: pool, dispatcher: dispatcher}
}

// Handle processes a single event from NATS.
func (h *EventHandler) Handle(evt events.Event) error {
	logger.Log.Info().Str("type", evt.Type).Str("id", evt.ID).Msg("Processing event")

	tmpl, exists := Templates[evt.Type]
	if !exists {
		logger.Log.Debug().Str("type", evt.Type).Msg("No template for event type, skipping")
		return nil
	}

	// Extract society_id from payload to find notification targets
	var payload map[string]interface{}
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to unmarshal event payload")
		return nil // don't retry on bad payload
	}

	societyID, _ := payload["society_id"].(string)
	if societyID == "" {
		logger.Log.Warn().Str("type", evt.Type).Msg("Event missing society_id")
		return nil
	}

	// Determine who to notify based on event type
	targets := h.getTargets(context.Background(), evt.Type, societyID, payload)
	if len(targets) == 0 {
		logger.Log.Debug().Str("type", evt.Type).Msg("No notification targets")
		return nil
	}

	// Build title/body from template (simplified — use text/template in production)
	title := tmpl.TitleTpl
	body := tmpl.BodyTpl

	// Dispatch to each target
	for _, target := range targets {
		h.dispatcher.Dispatch(context.Background(), target, title, body, tmpl.Channels, nil)
	}

	logger.Log.Info().Str("type", evt.Type).Int("targets", len(targets)).Msg("Notifications dispatched")
	return nil
}

// getTargets returns the users who should be notified for a given event.
func (h *EventHandler) getTargets(ctx context.Context, eventType, societyID string, payload map[string]interface{}) []model.NotificationTarget {
	var query string

	switch {
	case eventType == "sos.triggered" || eventType == "complaint.emergency":
		// Notify ALL members + guards
		query = `SELECT u.id, u.phone, u.name, u.fcm_token, u.whatsapp_opted_in, u.telegram_chat_id, u.preferred_language
			FROM app_user u
			JOIN user_society_membership usm ON usm.user_id = u.id
			WHERE usm.society_id = $1 AND usm.is_active = true AND u.is_active = true`

	case eventType == "visitor.logged":
		// Notify flat residents only
		flatID, _ := payload["flat_id"].(string)
		if flatID == "" {
			return nil
		}
		query = `SELECT u.id, u.phone, u.name, u.fcm_token, u.whatsapp_opted_in, u.telegram_chat_id, u.preferred_language
			FROM app_user u
			JOIN user_society_membership usm ON usm.user_id = u.id
			WHERE usm.society_id = $1 AND usm.flat_id = '` + flatID + `' AND usm.is_active = true AND u.is_active = true`

	default:
		// Notify society admins
		query = `SELECT u.id, u.phone, u.name, u.fcm_token, u.whatsapp_opted_in, u.telegram_chat_id, u.preferred_language
			FROM app_user u
			JOIN user_society_membership usm ON usm.user_id = u.id
			WHERE usm.society_id = $1 AND usm.is_active = true AND u.is_active = true
			AND usm.role IN ('admin', 'secretary', 'treasurer', 'president')`
	}

	rows, err := h.pool.Query(ctx, query, societyID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to query notification targets")
		return nil
	}
	defer rows.Close()

	var targets []model.NotificationTarget
	for rows.Next() {
		var t model.NotificationTarget
		if err := rows.Scan(&t.UserID, &t.Phone, &t.Name, &t.FCMToken, &t.WhatsappOptIn, &t.TelegramChatID, &t.Language); err != nil {
			logger.Log.Error().Err(err).Msg("Failed to scan target")
			continue
		}
		targets = append(targets, t)
	}
	return targets
}
