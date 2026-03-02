package app

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	JWTSecret            string
	AnthropicAPIKey      string
	S3Endpoint           string
	S3Bucket             string
	S3Region             string
	S3AccessKey          string
	S3SecretKey          string
	CORSOrigins          string
	GCloudCredentialsFile string
	ResendAPIKey         string
	EmailFrom            string
	LogLevel             string
	DBMaxConns           int32
	DBMinConns           int32
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		S3Endpoint:      getEnv("S3_ENDPOINT", ""),
		S3Bucket:        getEnv("S3_BUCKET", "lumber-now"),
		S3Region:        getEnv("S3_REGION", "us-east-1"),
		S3AccessKey:     os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:     os.Getenv("S3_SECRET_KEY"),
		CORSOrigins:          getEnv("CORS_ORIGINS", ""),
		GCloudCredentialsFile: os.Getenv("GCLOUD_CREDENTIALS_FILE"),
		ResendAPIKey:         os.Getenv("RESEND_API_KEY"),
		EmailFrom:            getEnv("EMAIL_FROM", "LumberNow <onboarding@resend.dev>"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		DBMaxConns:           getEnvInt32("DB_MAX_CONNS", 25),
		DBMinConns:           getEnvInt32("DB_MIN_CONNS", 5),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.CORSOrigins == "" {
		return nil, fmt.Errorf("CORS_ORIGINS is required (set to explicit allowed origins)")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt32(key string, fallback int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return int32(n)
		}
	}
	return fallback
}
