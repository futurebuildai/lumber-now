package app

import (
	"os"
	"testing"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"PORT", "DATABASE_URL", "JWT_SECRET", "ANTHROPIC_API_KEY",
		"S3_ENDPOINT", "S3_BUCKET", "S3_REGION", "S3_ACCESS_KEY", "S3_SECRET_KEY",
		"CORS_ORIGINS", "GCLOUD_CREDENTIALS_FILE", "RESEND_API_KEY", "EMAIL_FROM", "LOG_LEVEL",
	} {
		orig, exists := os.LookupEnv(key)
		if exists {
			t.Cleanup(func() { os.Setenv(key, orig) })
		} else {
			t.Cleanup(func() { os.Unsetenv(key) })
		}
		os.Unsetenv(key)
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "this-is-a-very-long-jwt-secret-for-testing-purposes")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")
}

func TestLoadConfig_MissingDatabaseURL(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("JWT_SECRET", "this-is-a-very-long-jwt-secret-for-testing-purposes")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for missing DATABASE_URL")
	}
}

func TestLoadConfig_MissingJWTSecret(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for missing JWT_SECRET")
	}
}

func TestLoadConfig_JWTSecretTooShort(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "short")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for short JWT_SECRET")
	}
}

func TestLoadConfig_MissingCORSOrigins(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "this-is-a-very-long-jwt-secret-for-testing-purposes")

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for missing CORS_ORIGINS")
	}
}

func TestLoadConfig_Valid(t *testing.T) {
	clearConfigEnv(t)
	setRequiredEnv(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test" {
		t.Errorf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.CORSOrigins != "http://localhost:3000" {
		t.Errorf("unexpected CORSOrigins: %s", cfg.CORSOrigins)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	clearConfigEnv(t)
	setRequiredEnv(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.S3Bucket != "lumber-now" {
		t.Errorf("expected default S3 bucket lumber-now, got %s", cfg.S3Bucket)
	}
	if cfg.S3Region != "us-east-1" {
		t.Errorf("expected default S3 region us-east-1, got %s", cfg.S3Region)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level info, got %s", cfg.LogLevel)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	clearConfigEnv(t)
	setRequiredEnv(t)
	os.Setenv("PORT", "9090")
	os.Setenv("S3_BUCKET", "custom-bucket")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("RESEND_API_KEY", "re_test")
	os.Setenv("EMAIL_FROM", "Test <test@example.com>")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.S3Bucket != "custom-bucket" {
		t.Errorf("expected custom-bucket, got %s", cfg.S3Bucket)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected debug, got %s", cfg.LogLevel)
	}
	if cfg.ResendAPIKey != "re_test" {
		t.Errorf("expected re_test, got %s", cfg.ResendAPIKey)
	}
	if cfg.EmailFrom != "Test <test@example.com>" {
		t.Errorf("expected custom email from, got %s", cfg.EmailFrom)
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_KEY_FOR_TEST")
	result := getEnv("NONEXISTENT_KEY_FOR_TEST", "fallback_value")
	if result != "fallback_value" {
		t.Errorf("expected fallback_value, got %s", result)
	}
}

func TestGetEnv_EnvSet(t *testing.T) {
	os.Setenv("TEST_KEY_FOR_GETENV", "actual_value")
	t.Cleanup(func() { os.Unsetenv("TEST_KEY_FOR_GETENV") })

	result := getEnv("TEST_KEY_FOR_GETENV", "fallback_value")
	if result != "actual_value" {
		t.Errorf("expected actual_value, got %s", result)
	}
}

// ---------------------------------------------------------------------------
// Error message content validation
// ---------------------------------------------------------------------------

func TestLoadConfig_MissingDatabaseURL_ErrorMessage(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("JWT_SECRET", "this-is-a-very-long-jwt-secret-for-testing-purposes")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
	if err.Error() != "DATABASE_URL is required" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestLoadConfig_MissingJWTSecret_ErrorMessage(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET")
	}
	if err.Error() != "JWT_SECRET is required" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestLoadConfig_ShortJWTSecret_ErrorMessage(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "short")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for short JWT_SECRET")
	}
	if err.Error() != "JWT_SECRET must be at least 32 characters" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestLoadConfig_MissingCORSOrigins_ErrorMessage(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "this-is-a-very-long-jwt-secret-for-testing-purposes")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing CORS_ORIGINS")
	}
	if err.Error() != "CORS_ORIGINS is required (set to explicit allowed origins)" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestLoadConfig_JWTSecretExactly31Chars(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "1234567890123456789012345678901") // 31 chars
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for 31-char JWT_SECRET")
	}
}

func TestLoadConfig_JWTSecretExactly32Chars(t *testing.T) {
	clearConfigEnv(t)
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET", "12345678901234567890123456789012") // 32 chars
	os.Setenv("CORS_ORIGINS", "http://localhost:3000")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error for 32-char JWT_SECRET: %v", err)
	}
	if cfg.JWTSecret != "12345678901234567890123456789012" {
		t.Errorf("unexpected JWTSecret: %s", cfg.JWTSecret)
	}
}

// ---------------------------------------------------------------------------
// getEnv edge cases
// ---------------------------------------------------------------------------

func TestGetEnv_EmptyFallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_KEY_EMPTY_FALLBACK")
	result := getEnv("NONEXISTENT_KEY_EMPTY_FALLBACK", "")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestGetEnv_EmptyEnvValue_ReturnsFallback(t *testing.T) {
	// os.Setenv with empty string means getEnv's os.Getenv returns ""
	// so it falls back to the default
	os.Setenv("TEST_EMPTY_ENV_VAL", "")
	t.Cleanup(func() { os.Unsetenv("TEST_EMPTY_ENV_VAL") })

	result := getEnv("TEST_EMPTY_ENV_VAL", "fallback")
	if result != "fallback" {
		t.Errorf("expected fallback for empty env var, got %q", result)
	}
}

func TestLoadConfig_ReturnsNonNilConfig(t *testing.T) {
	clearConfigEnv(t)
	setRequiredEnv(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Error("expected non-nil config")
	}
}

func TestLoadConfig_DefaultEmailFrom(t *testing.T) {
	clearConfigEnv(t)
	setRequiredEnv(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.EmailFrom != "LumberNow <onboarding@resend.dev>" {
		t.Errorf("expected default EmailFrom, got %q", cfg.EmailFrom)
	}
}
