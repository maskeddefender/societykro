package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents an app user stored in the app_user table.
type User struct {
	ID                uuid.UUID  `json:"id"`
	Phone             string     `json:"phone"`
	PhoneHash         string     `json:"-"`
	Name              string     `json:"name"`
	Email             *string    `json:"email,omitempty"`
	AvatarURL         *string    `json:"avatar_url,omitempty"`
	PreferredLanguage string     `json:"preferred_language"`
	IsSeniorCitizen   bool       `json:"is_senior_citizen"`
	WhatsappOptedIn   bool       `json:"whatsapp_opted_in"`
	TelegramChatID    *string    `json:"telegram_chat_id,omitempty"`
	FCMToken          *string    `json:"-"`
	LastActiveAt      *time.Time `json:"last_active_at,omitempty"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Society represents a housing society.
type Society struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Code              string    `json:"code"`
	Address           string    `json:"address"`
	City              string    `json:"city"`
	State             string    `json:"state"`
	Pincode           string    `json:"pincode"`
	TotalFlats        int       `json:"total_flats"`
	Subscription      string    `json:"subscription"`
	DefaultLanguage   string    `json:"default_language"`
	MaintenanceAmount float64   `json:"maintenance_amount"`
	MaintenanceDueDay int       `json:"maintenance_due_day"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

// Flat represents a unit/flat in a society.
type Flat struct {
	ID         uuid.UUID `json:"id"`
	SocietyID  uuid.UUID `json:"society_id"`
	FlatNumber string    `json:"flat_number"`
	Block      *string   `json:"block,omitempty"`
	Floor      *int      `json:"floor,omitempty"`
	FlatType   string    `json:"flat_type"`
	IsOccupied bool      `json:"is_occupied"`
	Occupancy  *string   `json:"occupancy,omitempty"`
}

// UserSocietyMembership represents a user's membership in a society.
type UserSocietyMembership struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	SocietyID       uuid.UUID  `json:"society_id"`
	FlatID          *uuid.UUID `json:"flat_id,omitempty"`
	Role            string     `json:"role"`
	IsPrimaryMember bool       `json:"is_primary_member"`
	JoinedAt        time.Time  `json:"joined_at"`
	IsActive        bool       `json:"is_active"`
}

// --- Request DTOs ---

// CreateSocietyRequest is the input for creating a new society.
type CreateSocietyRequest struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Pincode       string `json:"pincode"`
	TotalFlats    int    `json:"total_flats"`
	Blocks        int    `json:"blocks,omitempty"`
	Floors        int    `json:"floors,omitempty"`
	FlatsPerFloor int    `json:"flats_per_floor,omitempty"`
}

// JoinSocietyRequest is the input for joining a society by invite code.
type JoinSocietyRequest struct {
	Code       string  `json:"code"`
	FlatNumber *string `json:"flat_number,omitempty"`
}

// UpdateProfileRequest is the input for updating user profile.
type UpdateProfileRequest struct {
	Name              *string `json:"name,omitempty"`
	Email             *string `json:"email,omitempty"`
	PreferredLanguage *string `json:"preferred_language,omitempty"`
	IsSeniorCitizen   *bool   `json:"is_senior_citizen,omitempty"`
}

// --- Response DTOs ---

// AuthResponse is returned after successful OTP verification.
type AuthResponse struct {
	AccessToken  string                  `json:"access_token"`
	RefreshToken string                  `json:"refresh_token"`
	User         User                    `json:"user"`
	IsNewUser    bool                    `json:"is_new_user"`
	Memberships  []UserSocietyMembership `json:"memberships"`
}

// TokenPairResponse is returned after token refresh.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ProfileResponse includes user details and their society memberships.
type ProfileResponse struct {
	User        User                    `json:"user"`
	Memberships []UserSocietyMembership `json:"memberships"`
}
