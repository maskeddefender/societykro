package service

import "errors"

// Sentinel errors for the auth service. Handlers map these to HTTP status codes.
var (
	ErrTooManyOTPRequests  = errors.New("too many OTP requests, try again in 15 minutes")
	ErrOTPExpired          = errors.New("OTP expired or not found, request a new one")
	ErrInvalidOTP          = errors.New("invalid OTP")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrUserNotFound        = errors.New("user not found")
	ErrSocietyNotFound     = errors.New("society not found")
	ErrAlreadyMember       = errors.New("user is already a member of this society")
	ErrFlatNotFound        = errors.New("flat not found in this society")
)
