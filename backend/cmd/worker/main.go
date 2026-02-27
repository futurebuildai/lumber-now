package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/builderwire/lumber-now/backend/internal/app"
	"github.com/builderwire/lumber-now/backend/internal/platform/anthropic"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
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
	aiClient := anthropic.NewClient(cfg.AnthropicAPIKey)
	reqSvc := service.NewRequestService(s, aiClient)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("worker started, polling for pending requests...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutting down")
			return
		case <-ticker.C:
			processPendingRequests(ctx, s, reqSvc)
		}
	}
}

func processPendingRequests(ctx context.Context, s *store.Store, reqSvc *service.RequestService) {
	// Get all active dealers
	dealers, err := s.Queries.ListActiveDealers(ctx)
	if err != nil {
		slog.Error("failed to list dealers", "error", err)
		return
	}

	for _, dealer := range dealers {
		requests, err := s.Queries.ListRequestsByStatus(ctx, db.ListRequestsByStatusParams{
			DealerID: dealer.ID,
			Status:   db.RequestStatusPending,
			Limit:    10,
			Offset:   0,
		})
		if err != nil {
			continue
		}

		for _, req := range requests {
			slog.Info("processing request", "id", req.ID, "dealer", dealer.Slug, "type", req.InputType)
			if _, err := reqSvc.Process(ctx, req.ID); err != nil {
				slog.Error("failed to process request", "id", req.ID, "error", err)
			}
		}
	}
}
