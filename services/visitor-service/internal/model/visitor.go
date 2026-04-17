package model

import (
	"time"

	"github.com/google/uuid"
)

// Visitor represents a visitor entry at a society gate.
type Visitor struct {
	ID            uuid.UUID  `json:"id"`
	SocietyID     uuid.UUID  `json:"society_id"`
	FlatID        uuid.UUID  `json:"flat_id"`
	FlatNumber    string     `json:"flat_number,omitempty"`
	Name          string     `json:"name"`
	Phone         *string    `json:"phone,omitempty"`
	Purpose       string     `json:"purpose"`
	VehicleNumber *string    `json:"vehicle_number,omitempty"`
	PhotoURL      *string    `json:"photo_url,omitempty"`
	Status        string     `json:"status"` // pending, pre_approved, approved, denied, checked_in, checked_out
	OTPCode       *string    `json:"otp_code,omitempty"`
	OTPExpiresAt  *time.Time `json:"otp_expires_at,omitempty"`
	ApprovedBy    *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedVia   *string    `json:"approved_via,omitempty"` // app, otp, pass
	DenyReason    *string    `json:"deny_reason,omitempty"`
	LoggedBy      *uuid.UUID `json:"logged_by,omitempty"`
	CheckedInAt   *time.Time `json:"checked_in_at,omitempty"`
	CheckedOutAt  *time.Time `json:"checked_out_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// VisitorPass represents a recurring visitor pass for a flat.
type VisitorPass struct {
	ID            uuid.UUID  `json:"id"`
	FlatID        uuid.UUID  `json:"flat_id"`
	SocietyID     uuid.UUID  `json:"society_id"`
	VisitorName   string     `json:"visitor_name"`
	Phone         *string    `json:"phone,omitempty"`
	Purpose       string     `json:"purpose"`
	VehicleNumber *string    `json:"vehicle_number,omitempty"`
	ValidFrom     time.Time  `json:"valid_from"`
	ValidUntil    time.Time  `json:"valid_until"`
	CreatedBy     uuid.UUID  `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
}

// --- Request DTOs ---

// CreateVisitorRequest is the input for logging a visitor at the gate.
type CreateVisitorRequest struct {
	Name          string `json:"name"`
	FlatID        string `json:"flat_id"`
	Purpose       string `json:"purpose"`
	VehicleNumber string `json:"vehicle_number,omitempty"`
	Phone         string `json:"phone,omitempty"`
}

// PreApproveRequest is the input for a resident pre-approving a visitor.
type PreApproveRequest struct {
	Name          string `json:"name"`
	FlatID        string `json:"flat_id,omitempty"`
	Purpose       string `json:"purpose"`
	VehicleNumber string `json:"vehicle_number,omitempty"`
	Phone         string `json:"phone,omitempty"`
}

// ApproveRequest is the input for approving a visitor.
type ApproveRequest struct {
	Via string `json:"via,omitempty"` // app, otp
}

// DenyRequest is the input for denying a visitor.
type DenyRequest struct {
	Reason string `json:"reason"`
}

// CheckoutRequest is the input for checking out a visitor.
type CheckoutRequest struct{}

// CreatePassRequest is the input for creating a visitor pass.
type CreatePassRequest struct {
	VisitorName   string `json:"visitor_name"`
	Phone         string `json:"phone,omitempty"`
	Purpose       string `json:"purpose"`
	VehicleNumber string `json:"vehicle_number,omitempty"`
	ValidFrom     string `json:"valid_from"`
	ValidUntil    string `json:"valid_until"`
}

// VerifyOTPRequest is the input for verifying a visitor OTP at the gate.
type VerifyOTPRequest struct {
	OTPCode string `json:"otp_code"`
}

// VisitorListFilter specifies filters for listing visitors.
type VisitorListFilter struct {
	SocietyID uuid.UUID
	Status    *string
	FlatID    *uuid.UUID
	Cursor    *uuid.UUID
	Limit     int
}
