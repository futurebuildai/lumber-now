package handler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
)

const maxUploadSize = 10 * 1024 * 1024 // 10MB

var allowedContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
	"audio/mpeg":      true,
	"audio/wav":       true,
	"audio/ogg":       true,
	"audio/webm":      true,
	"text/csv":        true,
}

type MediaHandler struct {
	mediaSvc *service.MediaService
}

func NewMediaHandler(mediaSvc *service.MediaService) *MediaHandler {
	return &MediaHandler{mediaSvc: mediaSvc}
}

func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	if h.mediaSvc == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "media service not configured"})
	}

	dealerID, err := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tenant required"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required"})
	}

	// Enforce file size limit
	if file.Size > maxUploadSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file too large (max 10MB)"})
	}

	// Validate content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" || !allowedContentTypes[contentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported file type"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer f.Close()

	// Sanitize filename: keep only base name, strip path traversal
	safeName := filepath.Base(file.Filename)
	safeName = strings.ReplaceAll(safeName, "..", "")
	filename := fmt.Sprintf("%s-%s", uuid.New().String()[:8], safeName)

	key, err := h.mediaSvc.Upload(c.Context(), dealerID, filename, f, contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "upload failed"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"key": key,
	})
}

func (h *MediaHandler) Download(c *fiber.Ctx) error {
	if h.mediaSvc == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "media service not configured"})
	}

	dealerID, err := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tenant required"})
	}

	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "key required"})
	}

	// Use wildcard to handle nested paths
	key = c.Params("*1", key)

	// Tenant isolation: verify the S3 key belongs to this dealer.
	// Keys are stored as {dealer_id}/YYYY/MM/DD/filename.
	if !strings.HasPrefix(key, dealerID.String()+"/") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	url, err := h.mediaSvc.GetPresignedURL(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate URL"})
	}

	return c.JSON(fiber.Map{"url": url})
}
