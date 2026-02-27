package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/builderwire/lumber-now/backend/internal/app"
	"github.com/builderwire/lumber-now/backend/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	godotenv.Load()

	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := store.NewPool(context.Background(), cfg.DatabaseURL)
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
