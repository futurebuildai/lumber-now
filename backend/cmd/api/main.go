package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/builderwire/lumber-now/backend/internal/app"
	"github.com/builderwire/lumber-now/backend/internal/store"
)

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	godotenv.Load()

	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(cfg.LogLevel)})))

	pool, err := store.NewPool(context.Background(), cfg.DatabaseURL, store.PoolConfig{
		MaxConns: cfg.DBMaxConns,
		MinConns: cfg.DBMinConns,
	})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	s := store.New(pool)

	application, err := app.New(cfg, s)
	if err != nil {
		slog.Error("failed to create app", "error", err)
		os.Exit(1)
	}

	if err := application.Start(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
