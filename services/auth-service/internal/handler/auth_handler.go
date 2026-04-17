package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/auth-service/internal/service"
)

// AuthHandler handles all authentication-related HTTP endpoints.
type AuthHandler struct {
	svc *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// SendOTP sends a 6-digit OTP to the given phone number.
// POST /api/v1/auth/otp/send
func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Phone == "" || len(req.Phone) < 10 {
		return response.BadRequest(c, "Valid phone number required (e.g. +919876543210)")
	}

	if err := h.svc.SendOTP(c.Context(), req.Phone); err != nil {
		if errors.Is(err, service.ErrTooManyOTPRequests) {
			return response.TooManyRequests(c, err.Error())
		}
		return response.InternalError(c, "Failed to send OTP")
	}

	return response.OKMessage(c, "OTP sent successfully")
}

// VerifyOTP verifies the OTP and returns JWT tokens + user profile.
// POST /api/v1/auth/otp/verify
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Phone == "" || req.OTP == "" {
		return response.BadRequest(c, "Phone and OTP are required")
	}

	resp, err := h.svc.VerifyOTP(c.Context(), req.Phone, req.OTP)
	if err != nil {
		if errors.Is(err, service.ErrOTPExpired) || errors.Is(err, service.ErrInvalidOTP) {
			return response.Unauthorized(c, err.Error())
		}
		return response.InternalError(c, "Verification failed")
	}

	return response.OK(c, resp)
}

// RefreshToken issues new token pair using a valid refresh token.
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.RefreshToken == "" {
		return response.BadRequest(c, "Refresh token is required")
	}

	tokens, err := h.svc.RefreshTokens(c.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			return response.Unauthorized(c, err.Error())
		}
		return response.InternalError(c, "Token refresh failed")
	}

	return response.OK(c, tokens)
}

// Logout invalidates the user's refresh token.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Not authenticated")
	}

	if err := h.svc.Logout(c.Context(), userID); err != nil {
		return response.InternalError(c, "Logout failed")
	}

	return response.OKMessage(c, "Logged out successfully")
}

// GetProfile returns the authenticated user's profile with memberships.
// GET /api/v1/auth/me
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Not authenticated")
	}

	profile, err := h.svc.GetProfile(c.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch profile")
	}

	return response.OK(c, profile)
}

// UpdateFCMToken updates the push notification token for the authenticated user.
// PUT /api/v1/auth/me/fcm-token
func (h *AuthHandler) UpdateFCMToken(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Not authenticated")
	}

	var req struct {
		FCMToken string `json:"fcm_token"`
	}
	if err := c.BodyParser(&req); err != nil || req.FCMToken == "" {
		return response.BadRequest(c, "fcm_token is required")
	}

	if err := h.svc.UpdateFCMToken(c.Context(), userID, req.FCMToken); err != nil {
		return response.InternalError(c, "Failed to update token")
	}

	return response.OKMessage(c, "FCM token updated")
}
