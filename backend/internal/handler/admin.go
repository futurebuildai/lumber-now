package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

type AdminHandler struct {
	store *store.Store
}

func NewAdminHandler(s *store.Store) *AdminHandler {
	return &AdminHandler{store: s}
}

func (h *AdminHandler) ListRequests(c *fiber.Ctx) error {
	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	limit := clampLimit(int32(c.QueryInt("limit", 50)))
	offset := clampOffset(int32(c.QueryInt("offset", 0)))

	status := c.Query("status")
	if status != "" {
		// Validate status enum
		validStatuses := map[string]bool{
			"pending": true, "processing": true, "parsed": true,
			"confirmed": true, "sent": true, "failed": true,
		}
		if !validStatuses[status] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid status filter"})
		}
		requests, err := h.store.Queries.ListRequestsByStatus(c.Context(), db.ListRequestsByStatusParams{
			DealerID: dealerID,
			Status:   db.RequestStatus(status),
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list requests"})
		}
		total, _ := h.store.Queries.CountRequestsByStatus(c.Context(), db.CountRequestsByStatusParams{
			DealerID: dealerID,
			Status:   db.RequestStatus(status),
		})
		return c.JSON(fiber.Map{"requests": requests, "total": total})
	}

	requests, err := h.store.Queries.ListRequestsByDealer(c.Context(), db.ListRequestsByDealerParams{
		DealerID: dealerID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list requests"})
	}

	total, _ := h.store.Queries.CountRequestsByDealer(c.Context(), dealerID)

	return c.JSON(fiber.Map{"requests": requests, "total": total})
}

func clampLimit(v int32) int32 {
	if v < 1 {
		return 1
	}
	if v > 100 {
		return 100
	}
	return v
}

func clampOffset(v int32) int32 {
	if v < 0 {
		return 0
	}
	return v
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	limit := clampLimit(int32(c.QueryInt("limit", 50)))
	offset := clampOffset(int32(c.QueryInt("offset", 0)))

	role := c.Query("role")
	if role != "" {
		users, err := h.store.Queries.ListUsersByRole(c.Context(), db.ListUsersByRoleParams{
			DealerID: dealerID,
			Role:     db.UserRole(role),
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list users"})
		}
		return c.JSON(fiber.Map{"users": users})
	}

	users, err := h.store.Queries.ListUsersByDealer(c.Context(), db.ListUsersByDealerParams{
		DealerID: dealerID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list users"})
	}

	return c.JSON(fiber.Map{"users": users})
}

func (h *AdminHandler) AssignContractorToRep(c *fiber.Ctx) error {
	contractorID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user ID"})
	}

	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	var body struct {
		RepID string `json:"rep_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	repID, err := uuid.Parse(body.RepID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rep ID"})
	}

	err = h.store.Queries.AssignContractorToRepByDealer(c.Context(), db.AssignContractorToRepByDealerParams{
		ID:            contractorID,
		DealerID:      dealerID,
		AssignedRepID: pgtype.UUID{Bytes: repID, Valid: true},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to assign"})
	}

	return c.JSON(fiber.Map{"status": "assigned"})
}

func (h *AdminHandler) UpdateRouting(c *fiber.Ctx) error {
	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	var body struct {
		Assignments []struct {
			ContractorID string `json:"contractor_id"`
			RepID        string `json:"rep_id"`
		} `json:"assignments"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Execute all assignments within a single transaction
	var failed int
	if err := h.store.WithTx(c.Context(), func(qtx *db.Queries) error {
		for _, a := range body.Assignments {
			contractorID, err := uuid.Parse(a.ContractorID)
			if err != nil {
				failed++
				continue
			}
			repID, err := uuid.Parse(a.RepID)
			if err != nil {
				failed++
				continue
			}
			if err := qtx.AssignContractorToRepByDealer(c.Context(), db.AssignContractorToRepByDealerParams{
				ID:            contractorID,
				DealerID:      dealerID,
				AssignedRepID: pgtype.UUID{Bytes: repID, Valid: true},
			}); err != nil {
				slog.Error("failed to assign contractor", "contractor_id", contractorID, "error", err)
				failed++
			}
		}
		return nil
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "routing update failed"})
	}

	if failed > 0 {
		return c.Status(fiber.StatusMultiStatus).JSON(fiber.Map{"status": "partial", "failed": failed})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

func (h *AdminHandler) GetSettings(c *fiber.Ctx) error {
	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	dealer, err := h.store.Queries.GetDealer(c.Context(), dealerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dealer not found"})
	}

	return c.JSON(fiber.Map{
		"name":            dealer.Name,
		"logo_url":        dealer.LogoUrl,
		"primary_color":   dealer.PrimaryColor,
		"secondary_color": dealer.SecondaryColor,
		"contact_email":   dealer.ContactEmail,
		"contact_phone":   dealer.ContactPhone,
		"address":         dealer.Address,
	})
}

func (h *AdminHandler) UpdateSettings(c *fiber.Ctx) error {
	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	var body struct {
		Name           string `json:"name"`
		LogoURL        string `json:"logo_url"`
		PrimaryColor   string `json:"primary_color"`
		SecondaryColor string `json:"secondary_color"`
		ContactEmail   string `json:"contact_email"`
		ContactPhone   string `json:"contact_phone"`
		Address        string `json:"address"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	dealer, err := h.store.Queries.UpdateDealer(c.Context(), db.UpdateDealerParams{
		ID:             dealerID,
		Name:           body.Name,
		LogoUrl:        body.LogoURL,
		PrimaryColor:   body.PrimaryColor,
		SecondaryColor: body.SecondaryColor,
		ContactEmail:   body.ContactEmail,
		ContactPhone:   body.ContactPhone,
		Address:        body.Address,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update settings"})
	}

	return c.JSON(dealer)
}
