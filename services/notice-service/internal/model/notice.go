package model

import (
	"time"

	"github.com/google/uuid"
)

// Notice represents a society notice/announcement.
type Notice struct {
	ID                 uuid.UUID  `json:"id"`
	SocietyID          uuid.UUID  `json:"society_id"`
	CreatedBy          uuid.UUID  `json:"created_by"`
	CreatedByName      string     `json:"created_by_name,omitempty"`
	Title              string     `json:"title"`
	Body               string     `json:"body"`
	Category           string     `json:"category"`
	IsPinned           bool       `json:"is_pinned"`
	BroadcastWhatsapp  bool       `json:"broadcast_whatsapp"`
	BroadcastTelegram  bool       `json:"broadcast_telegram"`
	AttachmentURLs     []string   `json:"attachment_urls"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	ReadCount          int        `json:"read_count,omitempty"`
	TotalMembers       int        `json:"total_members,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// CreateNoticeRequest is the input for posting a notice.
type CreateNoticeRequest struct {
	Title             string   `json:"title"`
	Body              string   `json:"body"`
	Category          string   `json:"category,omitempty"`
	IsPinned          bool     `json:"is_pinned,omitempty"`
	BroadcastWhatsapp bool     `json:"broadcast_whatsapp,omitempty"`
	BroadcastTelegram bool     `json:"broadcast_telegram,omitempty"`
	AttachmentURLs    []string `json:"attachment_urls,omitempty"`
}
