package model

import (
	"time"

	"github.com/google/uuid"
)

// Vendor represents an external service provider (plumber, electrician, etc.).
type Vendor struct {
	ID                 uuid.UUID `json:"id"`
	SocietyID          uuid.UUID `json:"society_id"`
	Name               string    `json:"name"`
	CompanyName        *string   `json:"company_name,omitempty"`
	Phone              string    `json:"phone"`
	WhatsappPhone      *string   `json:"whatsapp_phone,omitempty"`
	Category           string    `json:"category"`
	SubCategory        *string   `json:"sub_category,omitempty"`
	Address            *string   `json:"address,omitempty"`
	AvgRating          float64   `json:"avg_rating"`
	TotalJobs          int       `json:"total_jobs"`
	CompletedJobs      int       `json:"completed_jobs"`
	ResponseTimeAvgHrs float64   `json:"response_time_avg_hrs"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// DomesticHelp represents a domestic helper (maid, cook, driver, etc.).
type DomesticHelp struct {
	ID          uuid.UUID `json:"id"`
	SocietyID   uuid.UUID `json:"society_id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	PhotoURL    *string   `json:"photo_url,omitempty"`
	Role        string    `json:"role"`
	IsVerified  bool      `json:"is_verified"`
	AvgRating   float64   `json:"avg_rating"`
	EntryMethod string    `json:"entry_method"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// DomesticHelpFlat links a domestic helper to a flat with pay and schedule details.
type DomesticHelpFlat struct {
	ID              uuid.UUID `json:"id"`
	DomesticHelpID  uuid.UUID `json:"domestic_help_id"`
	FlatID          uuid.UUID `json:"flat_id"`
	MonthlyPay      float64   `json:"monthly_pay"`
	WorkingDays     *string   `json:"working_days,omitempty"`
	IsActive        bool      `json:"is_active"`
}

// DomesticHelpAttendance records entry/exit of domestic help at a society.
type DomesticHelpAttendance struct {
	ID             uuid.UUID  `json:"id"`
	DomesticHelpID uuid.UUID  `json:"domestic_help_id"`
	SocietyID      uuid.UUID  `json:"society_id"`
	FlatID         uuid.UUID  `json:"flat_id"`
	EntryAt        time.Time  `json:"entry_at"`
	ExitAt         *time.Time `json:"exit_at,omitempty"`
	Date           time.Time  `json:"date"`
}

// --- Request DTOs ---

// CreateVendorRequest is the input for registering a new vendor.
type CreateVendorRequest struct {
	Name          string  `json:"name"`
	CompanyName   *string `json:"company_name,omitempty"`
	Phone         string  `json:"phone"`
	WhatsappPhone *string `json:"whatsapp_phone,omitempty"`
	Category      string  `json:"category"`
	SubCategory   *string `json:"sub_category,omitempty"`
	Address       *string `json:"address,omitempty"`
}

// UpdateVendorRequest is the input for updating an existing vendor.
type UpdateVendorRequest struct {
	Name          *string `json:"name,omitempty"`
	CompanyName   *string `json:"company_name,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	WhatsappPhone *string `json:"whatsapp_phone,omitempty"`
	Category      *string `json:"category,omitempty"`
	SubCategory   *string `json:"sub_category,omitempty"`
	Address       *string `json:"address,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// CreateDomesticHelpRequest is the input for registering domestic help.
type CreateDomesticHelpRequest struct {
	Name        string  `json:"name"`
	Phone       string  `json:"phone"`
	PhotoURL    *string `json:"photo_url,omitempty"`
	Role        string  `json:"role"`
	EntryMethod string  `json:"entry_method,omitempty"`
}

// LogAttendanceRequest records a domestic help entry at a flat.
type LogAttendanceRequest struct {
	FlatID uuid.UUID `json:"flat_id"`
}

// LinkFlatRequest links domestic help to a flat with pay info.
type LinkFlatRequest struct {
	FlatID     uuid.UUID `json:"flat_id"`
	MonthlyPay float64   `json:"monthly_pay"`
}

// VendorListFilter specifies filters for listing vendors.
type VendorListFilter struct {
	SocietyID   uuid.UUID
	Category    *string
	SubCategory *string
	IsActive    *bool
	Cursor      *uuid.UUID
	Limit       int
}
