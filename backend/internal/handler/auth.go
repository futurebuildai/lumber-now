package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
)

type AuthHandler struct {
	authSvc *service.AuthService
	store   *store.Store
}

func NewAuthHandler(authSvc *service.AuthService, s *store.Store) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, store: s}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	dealerID, err := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tenant required"})
	}

	tokens, err := h.authSvc.Login(c.Context(), service.LoginInput{
		DealerID: dealerID,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	return c.JSON(tokens)
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	dealerID, err := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tenant required"})
	}

	role := domain.Role(req.Role)
	if req.Role == "" {
		role = domain.RoleContractor
	}

	user, err := h.authSvc.Register(c.Context(), service.RegisterInput{
		DealerID: dealerID,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
		Role:     role,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":        user.ID,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.Role,
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	claims, err := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	user, err := h.store.Queries.GetUser(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	var assignedRepID *uuid.UUID
	if user.AssignedRepID.Valid {
		id := uuid.UUID(user.AssignedRepID.Bytes)
		assignedRepID = &id
	}

	return c.JSON(fiber.Map{
		"id":              user.ID,
		"dealer_id":       user.DealerID,
		"email":           user.Email,
		"full_name":       user.FullName,
		"phone":           user.Phone,
		"role":            user.Role,
		"assigned_rep_id": assignedRepID,
		"active":          user.Active,
		"created_at":      user.CreatedAt,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tokens, err := h.authSvc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	return c.JSON(tokens)
}
