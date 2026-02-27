package handler

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

type InventoryHandler struct {
	invSvc *service.InventoryService
	store  *store.Store
}

func NewInventoryHandler(invSvc *service.InventoryService, s *store.Store) *InventoryHandler {
	return &InventoryHandler{invSvc: invSvc, store: s}
}

func (h *InventoryHandler) List(c *fiber.Ctx) error {
	dealerID, err := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tenant required"})
	}

	limit := int32(c.QueryInt("limit", 50))
	offset := int32(c.QueryInt("offset", 0))

	search := c.Query("search")
	if search != "" {
		items, err := h.store.Queries.SearchInventory(c.Context(), db.SearchInventoryParams{
			DealerID:         dealerID,
			PlaintoTsquery: search,
			Limit:            limit,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "search failed"})
		}
		return c.JSON(fiber.Map{"items": items})
	}

	items, err := h.store.Queries.ListInventory(c.Context(), db.ListInventoryParams{
		DealerID: dealerID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list inventory"})
	}

	return c.JSON(fiber.Map{"items": items})
}

type createInventoryBody struct {
	SKU         string          `json:"sku"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Unit        string          `json:"unit"`
	Price       string          `json:"price"`
	InStock     bool            `json:"in_stock"`
	Metadata    json.RawMessage `json:"metadata"`
}

func (h *InventoryHandler) Create(c *fiber.Ctx) error {
	var body createInventoryBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	metadata := body.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}

	item, err := h.store.Queries.CreateInventoryItem(c.Context(), db.CreateInventoryItemParams{
		DealerID:    dealerID,
		Sku:         body.SKU,
		Name:        body.Name,
		Description: body.Description,
		Category:    body.Category,
		Unit:        body.Unit,
		Price:       body.Price,
		InStock:     body.InStock,
		Metadata:    metadata,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *InventoryHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid item ID"})
	}

	var body createInventoryBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	metadata := body.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}

	item, err := h.store.Queries.UpdateInventoryItem(c.Context(), db.UpdateInventoryItemParams{
		ID:          id,
		Name:        body.Name,
		Description: body.Description,
		Category:    body.Category,
		Unit:        body.Unit,
		Price:       body.Price,
		InStock:     body.InStock,
		Metadata:    metadata,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update item"})
	}

	return c.JSON(item)
}

func (h *InventoryHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid item ID"})
	}

	if err := h.store.Queries.DeleteInventoryItem(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete item"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *InventoryHandler) ImportCSV(c *fiber.Ctx) error {
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

	result, err := h.invSvc.ImportCSV(c.Context(), dealerID, f)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}
