package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationTarget represents a user to notify and their available channels.
type NotificationTarget struct {
	UserID         uuid.UUID
	Phone          string
	Name           string
	FCMToken       *string
	WhatsappOptIn  bool
	TelegramChatID *string
	Language       string
}

// Notification represents a notification to be sent.
type Notification struct {
	ID          string    `json:"id"`
	SocietyID   string    `json:"society_id"`
	UserID      string    `json:"user_id"`
	Channel     string    `json:"channel"` // push, whatsapp, telegram, sms, email
	EventType   string    `json:"event_type"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Data        map[string]string `json:"data,omitempty"`
	Status      string    `json:"status"` // pending, sent, failed
	SentAt      *time.Time `json:"sent_at,omitempty"`
	FailReason  *string   `json:"fail_reason,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationTemplate defines the template for each event type.
type NotificationTemplate struct {
	EventType string
	TitleTpl  string
	BodyTpl   string
	Channels  []string // which channels to use for this event type
	Priority  string   // normal, high (high = SOS, emergency)
}
