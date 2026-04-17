package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DatabaseConfig struct {
	Host           string
	Port           int
	Name           string
	User           string
	Password       string
	SSLMode        string
	MaxConnections int
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	PrivateKeyPath     string
	PublicKeyPath      string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	NatsURL  string
}

func Load() *Config {
	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "societykro"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:           getEnv("DATABASE_HOST", "localhost"),
			Port:           getEnvInt("DATABASE_PORT", 5432),
			Name:           getEnv("DATABASE_NAME", "societykro"),
			User:           getEnv("DATABASE_USER", "societykro"),
			Password:       getEnv("DATABASE_PASSWORD", "societykro_dev"),
			SSLMode:        getEnv("DATABASE_SSL_MODE", "disable"),
			MaxConnections: getEnvInt("DATABASE_MAX_CONNECTIONS", 25),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			PrivateKeyPath:     getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
			PublicKeyPath:      getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
			AccessTokenExpiry:  getEnvDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_TOKEN_EXPIRY", 720*time.Hour),
		},
		NatsURL: getEnv("NATS_URL", "nats://localhost:4222"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}
