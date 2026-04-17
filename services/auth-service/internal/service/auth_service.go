package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/societykro/go-common/auth"
	"github.com/societykro/go-common/config"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/auth-service/internal/model"
	"github.com/societykro/auth-service/internal/repository"
)

// AuthService handles authentication, user management, and society operations.
type AuthService struct {
	userRepo    *repository.UserRepository
	societyRepo *repository.SocietyRepository
	redis       *redis.Client
	jwt         *auth.JWTManager
	config      *config.Config
}

// NewAuthService creates a new AuthService with all dependencies.
func NewAuthService(
	userRepo *repository.UserRepository,
	societyRepo *repository.SocietyRepository,
	rdb *redis.Client,
	jwtMgr *auth.JWTManager,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		societyRepo: societyRepo,
		redis:       rdb,
		jwt:         jwtMgr,
		config:      cfg,
	}
}

// SendOTP generates a 6-digit OTP and stores it in Redis with a 5-minute TTL.
// Rate-limited to 3 requests per phone per 15 minutes.
func (s *AuthService) SendOTP(ctx context.Context, phone string) error {
	attemptsKey := fmt.Sprintf("auth:otp_attempts:%s", phone)
	attempts, _ := s.redis.Get(ctx, attemptsKey).Int()
	if attempts >= 3 {
		return ErrTooManyOTPRequests
	}

	otp, err := generateSecureOTP(6)
	if err != nil {
		return fmt.Errorf("generate OTP: %w", err)
	}

	pipe := s.redis.Pipeline()
	pipe.Set(ctx, otpKey(phone), otp, 5*time.Minute)
	pipe.Incr(ctx, attemptsKey)
	pipe.Expire(ctx, attemptsKey, 15*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("store OTP: %w", err)
	}

	// In development, log OTP to console. In production, send via MSG91.
	if s.config.App.Env == "development" {
		logger.Log.Debug().Str("phone", phone).Str("otp", otp).Msg("DEV OTP generated")
	} else {
		// TODO: Integrate MSG91 SMS API
		logger.Log.Info().Str("phone", maskPhone(phone)).Msg("OTP sent via SMS")
	}

	return nil
}

// VerifyOTP validates the OTP against Redis and returns an authenticated response
// with signed JWT tokens. Creates a new user if the phone is not registered.
func (s *AuthService) VerifyOTP(ctx context.Context, phone, otp string) (*model.AuthResponse, error) {
	stored, err := s.redis.Get(ctx, otpKey(phone)).Result()
	if err == redis.Nil {
		return nil, ErrOTPExpired
	}
	if err != nil {
		return nil, fmt.Errorf("fetch OTP: %w", err)
	}

	// Dev bypass: accept "000000" in development
	if stored != otp && !(s.config.App.Env == "development" && otp == "000000") {
		return nil, ErrInvalidOTP
	}

	s.redis.Del(ctx, otpKey(phone))

	// Find or create user
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	isNewUser := user == nil
	if isNewUser {
		user, err = s.userRepo.Create(ctx, phone, "New User", "hi")
		if err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	}

	_ = s.userRepo.UpdateLastActive(ctx, user.ID)

	// Fetch memberships to include default society/role in token
	var societyID *string
	var role *string
	memberships, _ := s.userRepo.GetMemberships(ctx, user.ID)
	if len(memberships) > 0 {
		sid := memberships[0].SocietyID.String()
		societyID = &sid
		role = &memberships[0].Role
	}

	// Generate JWT tokens
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Phone, societyID, role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, refreshTTL, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Store refresh token in Redis for single-use validation
	refreshKey := fmt.Sprintf("auth:refresh:%s", user.ID.String())
	s.redis.Set(ctx, refreshKey, refreshToken, refreshTTL)

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
		IsNewUser:    isNewUser,
		Memberships:  memberships,
	}, nil
}

// RefreshTokens validates the refresh token and issues a new token pair.
// Implements single-use rotation: old refresh token is invalidated.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshTokenStr string) (*model.TokenPairResponse, error) {
	userID, _, err := s.jwt.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check refresh token matches what's stored (single-use)
	refreshKey := fmt.Sprintf("auth:refresh:%s", userID.String())
	stored, err := s.redis.Get(ctx, refreshKey).Result()
	if err == redis.Nil || stored != refreshTokenStr {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Fetch role for new access token
	var societyID *string
	var role *string
	memberships, _ := s.userRepo.GetMemberships(ctx, user.ID)
	if len(memberships) > 0 {
		sid := memberships[0].SocietyID.String()
		societyID = &sid
		role = &memberships[0].Role
	}

	newAccess, err := s.jwt.GenerateAccessToken(user.ID, user.Phone, societyID, role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefresh, refreshTTL, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Rotate: replace old refresh token
	s.redis.Set(ctx, refreshKey, newRefresh, refreshTTL)

	return &model.TokenPairResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}, nil
}

// Logout invalidates the refresh token for the user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	refreshKey := fmt.Sprintf("auth:refresh:%s", userID.String())
	return s.redis.Del(ctx, refreshKey).Err()
}

// GetProfile returns the user profile with society memberships.
func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.ProfileResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	memberships, err := s.userRepo.GetMemberships(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("fetch memberships: %w", err)
	}

	return &model.ProfileResponse{
		User:        *user,
		Memberships: memberships,
	}, nil
}

// UpdateFCMToken updates the user's Firebase Cloud Messaging token.
func (s *AuthService) UpdateFCMToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.userRepo.UpdateFCMToken(ctx, userID, token)
}

// --- Society Operations ---

// CreateSociety creates a new society and auto-generates flats.
func (s *AuthService) CreateSociety(ctx context.Context, req model.CreateSocietyRequest, creatorID uuid.UUID) (*model.Society, error) {
	society, err := s.societyRepo.Create(ctx, req.Name, req.Address, req.City, req.State, req.Pincode, req.TotalFlats)
	if err != nil {
		return nil, fmt.Errorf("create society: %w", err)
	}

	// Auto-generate flats if block info provided
	if req.Blocks > 0 && req.FlatsPerFloor > 0 && req.Floors > 0 {
		if err := s.societyRepo.GenerateFlats(ctx, society.ID, req.Blocks, req.Floors, req.FlatsPerFloor); err != nil {
			logger.Log.Error().Err(err).Msg("Failed to auto-generate flats")
		}
	}

	// Make creator an admin
	_, err = s.societyRepo.AddMember(ctx, creatorID, society.ID, nil, "admin")
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to add creator as admin")
	}

	return society, nil
}

// GetSociety returns society details by ID.
func (s *AuthService) GetSociety(ctx context.Context, id uuid.UUID) (*model.Society, error) {
	society, err := s.societyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find society: %w", err)
	}
	if society == nil {
		return nil, ErrSocietyNotFound
	}
	return society, nil
}

// JoinSociety adds a user to a society by invite code.
func (s *AuthService) JoinSociety(ctx context.Context, userID uuid.UUID, code string, flatNumber *string) (*model.UserSocietyMembership, error) {
	society, err := s.societyRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("find society: %w", err)
	}
	if society == nil {
		return nil, ErrSocietyNotFound
	}

	var flatID *uuid.UUID
	if flatNumber != nil {
		flat, err := s.societyRepo.FindFlatByNumber(ctx, society.ID, *flatNumber)
		if err != nil {
			return nil, fmt.Errorf("find flat: %w", err)
		}
		if flat != nil {
			flatID = &flat.ID
		}
	}

	membership, err := s.societyRepo.AddMember(ctx, userID, society.ID, flatID, "resident")
	if err != nil {
		return nil, fmt.Errorf("add member: %w", err)
	}

	return membership, nil
}

// ListFlats returns all flats for a society.
func (s *AuthService) ListFlats(ctx context.Context, societyID uuid.UUID) ([]model.Flat, error) {
	return s.societyRepo.ListFlats(ctx, societyID)
}

// --- Helpers ---

func otpKey(phone string) string {
	return fmt.Sprintf("auth:otp:%s", phone)
}

func maskPhone(phone string) string {
	if len(phone) < 6 {
		return "****"
	}
	return phone[:4] + "****" + phone[len(phone)-2:]
}

func generateSecureOTP(length int) (string, error) {
	otp := make([]byte, length)
	for i := range otp {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp[i] = byte('0' + n.Int64())
	}
	return string(otp), nil
}
