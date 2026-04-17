package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Phone     string    `json:"phone"`
	SocietyID *string   `json:"society_id,omitempty"`
	Role      *string   `json:"role,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	privateKey         *rsa.PrivateKey
	publicKey          *rsa.PublicKey
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewJWTManager(privateKeyPath, publicKeyPath string, accessExpiry, refreshExpiry time.Duration) (*JWTManager, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}
	// Try PKCS1 first, fall back to PKCS8
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		keyInterface, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse private key (tried PKCS1 and PKCS8): %w", err2)
		}
		var ok bool
		privKey, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not RSA")
		}
	}

	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	pubBlock, _ := pem.Decode(pubBytes)
	if pubBlock == nil {
		return nil, fmt.Errorf("failed to decode public key PEM")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return &JWTManager{
		privateKey:         privKey,
		publicKey:          pubKey,
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
	}, nil
}

func (j *JWTManager) GenerateAccessToken(userID uuid.UUID, phone string, societyID *string, role *string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Phone:     phone,
		SocietyID: societyID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "societykro",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

func (j *JWTManager) GenerateRefreshToken(userID uuid.UUID) (string, time.Duration, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "societykro",
		Subject:   userID.String(),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(j.privateKey)
	return signed, j.refreshTokenExpiry, err
}

func (j *JWTManager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func (j *JWTManager) ValidateRefreshToken(tokenStr string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.publicKey, nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", fmt.Errorf("invalid refresh token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid user ID in token")
	}

	return userID, claims.ID, nil
}

func (j *JWTManager) GetPublicKey() *rsa.PublicKey {
	return j.publicKey
}
