package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/societykro/go-common/auth"
	"github.com/societykro/go-common/response"
)

type contextKey string

const (
	CtxUserID    contextKey = "user_id"
	CtxPhone     contextKey = "phone"
	CtxSocietyID contextKey = "society_id"
	CtxRole      contextKey = "role"
	CtxClaims    contextKey = "claims"
)

// JWTMiddleware validates the Bearer token and injects claims into context
func JWTMiddleware(jwtMgr *auth.JWTManager, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return response.Unauthorized(c, "Missing Authorization header")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return response.Unauthorized(c, "Invalid Authorization format. Use: Bearer <token>")
		}

		tokenStr := parts[1]

		// Check if token is blacklisted (logout)
		blacklistKey := "auth:blacklist:" + tokenStr[:16] // use prefix as key
		if rdb != nil {
			exists, _ := rdb.Exists(c.Context(), blacklistKey).Result()
			if exists > 0 {
				return response.Unauthorized(c, "Token has been revoked")
			}
		}

		claims, err := jwtMgr.ValidateAccessToken(tokenStr)
		if err != nil {
			return response.Unauthorized(c, "Invalid or expired token")
		}

		// Inject into Fiber locals (accessible via c.Locals)
		c.Locals(string(CtxUserID), claims.UserID)
		c.Locals(string(CtxPhone), claims.Phone)
		c.Locals(string(CtxClaims), claims)

		if claims.SocietyID != nil {
			c.Locals(string(CtxSocietyID), *claims.SocietyID)
		}
		if claims.Role != nil {
			c.Locals(string(CtxRole), *claims.Role)
		}

		return c.Next()
	}
}

// RequireRole checks if the user has one of the required roles
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals(string(CtxRole)).(string)
		if !ok || userRole == "" {
			return response.Forbidden(c, "No role assigned")
		}

		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}

		return response.Forbidden(c, "Insufficient permissions")
	}
}

// RequireAdmin is shorthand for RequireRole("admin", "secretary", "treasurer", "president")
func RequireAdmin() fiber.Handler {
	return RequireRole("admin", "secretary", "treasurer", "president")
}

// GetUserID extracts user_id from Fiber context
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(string(CtxUserID)).(uuid.UUID)
	return id, ok
}

// GetSocietyID extracts society_id from Fiber context
func GetSocietyID(c *fiber.Ctx) (string, bool) {
	id, ok := c.Locals(string(CtxSocietyID)).(string)
	return id, ok
}

// GetRole extracts role from Fiber context
func GetRole(c *fiber.Ctx) (string, bool) {
	role, ok := c.Locals(string(CtxRole)).(string)
	return role, ok
}

// SetSocietyContext is used by services that receive society_id from URL params
// and need to validate user belongs to that society
func SetSocietyContext(c *fiber.Ctx) context.Context {
	ctx := c.Context()
	return ctx
}
