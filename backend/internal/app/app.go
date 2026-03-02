package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/builderwire/lumber-now/backend/internal/handler"
	"github.com/builderwire/lumber-now/backend/internal/middleware"
	"github.com/builderwire/lumber-now/backend/internal/platform/anthropic"
	"github.com/builderwire/lumber-now/backend/internal/platform/email"
	"github.com/builderwire/lumber-now/backend/internal/platform/gcloud"
	s3platform "github.com/builderwire/lumber-now/backend/internal/platform/s3"
	"github.com/builderwire/lumber-now/backend/internal/router"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
)

type App struct {
	fiber   *fiber.App
	config  *Config
	store   *store.Store
	closers []io.Closer
	reqSvc  *service.RequestService
}

func New(cfg *Config, s *store.Store) (*App, error) {
	f := fiber.New(fiber.Config{
		AppName:      "LumberNow API",
		BodyLimit:    50 * 1024 * 1024, // 50MB for file uploads
		ErrorHandler: globalErrorHandler,
	})

	// Services
	authSvc := service.NewAuthService(s, cfg.JWTSecret)

	aiClient := anthropic.NewClient(cfg.AnthropicAPIKey)
	invSvc := service.NewInventoryService(s)

	s3Client, err := s3platform.NewClient(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey)
	if err != nil {
		slog.Warn("S3 client init failed, media uploads disabled", "error", err)
		s3Client = nil
	}

	var mediaSvc *service.MediaService
	if s3Client != nil {
		mediaSvc = service.NewMediaService(s3Client, cfg.S3Bucket)
	}

	// Optional: Google Cloud Speech-to-Text
	var closers []io.Closer
	var transcriber service.Transcriber
	if cfg.GCloudCredentialsFile != "" {
		speechClient, err := gcloud.NewSpeechClient(context.Background(), cfg.GCloudCredentialsFile)
		if err != nil {
			slog.Warn("Google Cloud STT init failed, voice transcription disabled", "error", err)
		} else {
			transcriber = speechClient
			closers = append(closers, speechClient)
			slog.Info("Google Cloud Speech-to-Text enabled")
		}
	}

	// Optional: Resend email client
	var emailClient service.EmailSender
	if cfg.ResendAPIKey != "" {
		emailClient = email.NewClient(cfg.ResendAPIKey, cfg.EmailFrom)
		slog.Info("Resend email client enabled")
	}

	reqSvc := service.NewRequestService(s, aiClient, transcriber, emailClient, mediaSvc)

	// Handlers
	healthHandler := handler.NewHealthHandler(s)
	healthHandler.RegisterCircuit("ai", aiClient)
	if s3Client != nil {
		healthHandler.RegisterCircuit("s3", s3Client)
	}

	handlers := router.Handlers{
		Health:    healthHandler,
		Auth:      handler.NewAuthHandler(authSvc, s),
		Tenant:    handler.NewTenantHandler(s),
		Request:   handler.NewRequestHandler(reqSvc, s),
		Inventory: handler.NewInventoryHandler(invSvc, s),
		Media:     handler.NewMediaHandler(mediaSvc),
		Admin:     handler.NewAdminHandler(s),
		Platform:  handler.NewPlatformHandler(s, authSvc, mediaSvc),
	}

	metrics := middleware.NewMetrics()
	handlers.Request.SetMetrics(metrics)
	metrics.SetPoolStatsFunc(func() middleware.PoolStats {
		stat := s.Pool.Stat()
		return middleware.PoolStats{
			TotalConns:    stat.TotalConns(),
			IdleConns:     stat.IdleConns(),
			AcquiredConns: stat.AcquiredConns(),
			MaxConns:      stat.MaxConns(),
		}
	})
	router.Setup(f, s, authSvc, handlers, cfg.CORSOrigins, metrics)

	return &App{
		fiber:   f,
		config:  cfg,
		store:   s,
		closers: closers,
		reqSvc:  reqSvc,
	}, nil
}

func (a *App) Start() error {
	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		slog.Info("shutting down server...")
		// Wait for in-flight email goroutines
		if a.reqSvc != nil {
			a.reqSvc.Close()
		}
		for _, c := range a.closers {
			c.Close()
		}
		if err := a.fiber.ShutdownWithTimeout(30 * time.Second); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	addr := fmt.Sprintf(":%s", a.config.Port)
	slog.Info("starting server", "addr", addr)
	return a.fiber.Listen(addr)
}

func globalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}
	slog.Error("unhandled error", "status", code, "error", err.Error(), "path", c.Path())
	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
