package model

import (
	"time"

	"github.com/google/uuid"
)

// Complaint represents a complaint in the system.
type Complaint struct {
	ID                  uuid.UUID  `json:"id"`
	SocietyID           uuid.UUID  `json:"society_id"`
	FlatID              *uuid.UUID `json:"flat_id,omitempty"`
	RaisedBy            uuid.UUID  `json:"raised_by"`
	RaisedByName        string     `json:"raised_by_name,omitempty"`
	TicketNumber        string     `json:"ticket_number"`
	Category            string     `json:"category"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	DescriptionOriginal *string    `json:"description_original,omitempty"`
	DescriptionEnglish  *string    `json:"description_english,omitempty"`
	OriginalLanguage    *string    `json:"original_language,omitempty"`
	VoiceURL            *string    `json:"voice_url,omitempty"`
	ImageURLs           []string   `json:"image_urls"`
	Status              string     `json:"status"`
	Priority            string     `json:"priority"`
	IsEmergency         bool       `json:"is_emergency"`
	IsCommonArea        bool       `json:"is_common_area"`
	AssignedVendorID    *uuid.UUID `json:"assigned_vendor_id,omitempty"`
	AssignedVendorName  string     `json:"assigned_vendor_name,omitempty"`
	AssignedAt          *time.Time `json:"assigned_at,omitempty"`
	ResolvedAt          *time.Time `json:"resolved_at,omitempty"`
	ClosedAt            *time.Time `json:"closed_at,omitempty"`
	ResolutionRating    *int       `json:"resolution_rating,omitempty"`
	ResolutionFeedback  *string    `json:"resolution_feedback,omitempty"`
	Source              string     `json:"source"`
	EscalationCount     int        `json:"escalation_count"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// Comment represents a comment on a complaint thread.
type Comment struct {
	ID             uuid.UUID  `json:"id"`
	ComplaintID    uuid.UUID  `json:"complaint_id"`
	UserID         uuid.UUID  `json:"user_id"`
	UserName       string     `json:"user_name,omitempty"`
	Comment        string     `json:"comment"`
	ImageURL       *string    `json:"image_url,omitempty"`
	IsInternal     bool       `json:"is_internal"`
	IsStatusChange bool       `json:"is_status_change"`
	OldStatus      *string    `json:"old_status,omitempty"`
	NewStatus      *string    `json:"new_status,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// --- Request DTOs ---

// CreateComplaintRequest is the input for raising a new complaint.
type CreateComplaintRequest struct {
	Category     string   `json:"category"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ImageURLs    []string `json:"image_urls,omitempty"`
	Priority     string   `json:"priority,omitempty"`
	IsEmergency  bool     `json:"is_emergency,omitempty"`
	IsCommonArea bool     `json:"is_common_area,omitempty"`
	Source       string   `json:"source,omitempty"`
}

// UpdateStatusRequest changes the status of a complaint.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// AssignVendorRequest assigns a vendor to a complaint.
type AssignVendorRequest struct {
	VendorID string `json:"vendor_id"`
}

// AddCommentRequest adds a comment to a complaint.
type AddCommentRequest struct {
	Comment    string  `json:"comment"`
	ImageURL   *string `json:"image_url,omitempty"`
	IsInternal bool    `json:"is_internal,omitempty"`
}

// RateResolutionRequest rates and gives feedback on complaint resolution.
type RateResolutionRequest struct {
	Rating   int     `json:"rating"`
	Feedback *string `json:"feedback,omitempty"`
}

// ComplaintListFilter specifies filters for listing complaints.
type ComplaintListFilter struct {
	SocietyID uuid.UUID
	Status    *string
	Category  *string
	FlatID    *uuid.UUID
	Cursor    *uuid.UUID
	Limit     int
}
