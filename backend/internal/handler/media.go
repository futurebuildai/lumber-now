package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
)

type MediaHandler struct {
	mediaSvc *service.MediaService
}

func NewMediaHandler(mediaSvc *service.MediaService) *MediaHandler {
	return &MediaHandler{mediaSvc: mediaSvc}
}

func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	dealerID, err := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tenant required"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer f.Close()

	filename := fmt.Sprintf("%s-%s", uuid.New().String()[:8], file.Filename)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	key, err := h.mediaSvc.Upload(c.Context(), dealerID, filename, f, contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "upload failed"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"key": key,
	})
}

func (h *MediaHandler) Download(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "key required"})
	}

	// Use wildcard to handle nested paths
	key = c.Params("*1", key)

	url, err := h.mediaSvc.GetPresignedURL(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate URL"})
	}

	return c.JSON(fiber.Map{"url": url})
}
