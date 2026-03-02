package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
)

const testJWTSecret = "test-secret-key-that-is-at-least-32-chars-long"

// testAuthService wraps service.AuthService using a known secret.
func newTestAuthService() *service.AuthService {
	return service.NewAuthService(nil, testJWTSecret)
}

// makeToken creates a JWT token with the given claims for testing.
func makeToken(t *testing.T, userID, dealerID uuid.UUID, email string, role domain.Role, tokenType string, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":       userID.String(),
		"dealer_id": dealerID.String(),
		"email":     email,
		"role":      string(role),
		"type":      tokenType,
		"jti":       uuid.New().String(),
		"exp":       exp.Unix(),
		"iat":       time.Now().Unix(),
		"nbf":       time.Now().Unix(),
		"iss":       "lumber-now",
		"aud":       "lumber-now-api",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func bodyJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	return m
}

// ---------------------------------------------------------------------------
// RequestID middleware
// ---------------------------------------------------------------------------

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		rid, _ := c.Locals(domain.LocalsRequestID).(string)
		return c.SendString(rid)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	rid := string(body)
	if rid == "" {
		t.Error("expected a generated request ID, got empty string")
	}
	// Should be a valid UUID
	if _, err := uuid.Parse(rid); err != nil {
		t.Errorf("expected valid UUID, got %q: %v", rid, err)
	}

	// Should also be in the response header
	headerRID := resp.Header.Get("X-Request-ID")
	if headerRID != rid {
		t.Errorf("expected X-Request-ID header %q, got %q", rid, headerRID)
	}
}

func TestRequestID_PassesThroughExisting(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		rid, _ := c.Locals(domain.LocalsRequestID).(string)
		return c.SendString(rid)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "my-custom-id-123")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "my-custom-id-123" {
		t.Errorf("expected my-custom-id-123, got %q", string(body))
	}
	if resp.Header.Get("X-Request-ID") != "my-custom-id-123" {
		t.Errorf("expected X-Request-ID header my-custom-id-123, got %q", resp.Header.Get("X-Request-ID"))
	}
}

// ---------------------------------------------------------------------------
// RequireRole middleware
// ---------------------------------------------------------------------------

func TestRequireRole_AllowsMatchingRole(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(domain.LocalsClaims, &domain.JWTClaims{
			UserID:   uuid.New(),
			DealerID: uuid.New(),
			Email:    "admin@test.com",
			Role:     domain.RoleDealerAdmin,
		})
		return c.Next()
	})
	app.Get("/test", RequireRole(domain.RoleDealerAdmin, domain.RoleSalesRep), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequireRole_BlocksNonMatchingRole(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(domain.LocalsClaims, &domain.JWTClaims{
			UserID:   uuid.New(),
			DealerID: uuid.New(),
			Email:    "user@test.com",
			Role:     domain.RoleContractor,
		})
		return c.Next()
	})
	app.Get("/test", RequireRole(domain.RoleDealerAdmin), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 403 {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
	m := bodyJSON(t, resp)
	if m["error"] != "insufficient permissions" {
		t.Errorf("expected 'insufficient permissions', got %v", m["error"])
	}
}

func TestRequireRole_BlocksUnauthenticated(t *testing.T) {
	app := fiber.New()
	// No claims set in locals
	app.Get("/test", RequireRole(domain.RoleDealerAdmin), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRequireRole_AllowsMultipleRoles(t *testing.T) {
	roles := []domain.Role{domain.RoleSalesRep, domain.RoleDealerAdmin, domain.RolePlatformAdmin}
	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals(domain.LocalsClaims, &domain.JWTClaims{
					UserID:   uuid.New(),
					DealerID: uuid.New(),
					Email:    "user@test.com",
					Role:     role,
				})
				return c.Next()
			})
			app.Get("/test", RequireRole(domain.RoleSalesRep, domain.RoleDealerAdmin, domain.RolePlatformAdmin), func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()

			if resp.StatusCode != 200 {
				t.Errorf("expected 200 for role %s, got %d", role, resp.StatusCode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Auth middleware
// ---------------------------------------------------------------------------

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()
	app.Use(Auth(authSvc))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	m := bodyJSON(t, resp)
	if m["error"] != "missing authorization header" {
		t.Errorf("expected 'missing authorization header', got %v", m["error"])
	}
}

func TestAuth_InvalidHeaderFormat(t *testing.T) {
	cases := []struct {
		name   string
		header string
	}{
		{"no_bearer_prefix", "Token abc123"},
		{"single_word", "Bearer"},
		{"empty_after_bearer", "Basic abc123"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			authSvc := newTestAuthService()
			app := fiber.New()
			app.Use(Auth(authSvc))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != 401 {
				t.Errorf("expected 401, got %d", resp.StatusCode)
			}
		})
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()
	app.Use(Auth(authSvc))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	m := bodyJSON(t, resp)
	if m["error"] != "invalid or expired token" {
		t.Errorf("expected 'invalid or expired token', got %v", m["error"])
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()
	app.Use(Auth(authSvc))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	token := makeToken(t, uuid.New(), uuid.New(), "user@test.com", domain.RoleSalesRep, "access", time.Now().Add(-1*time.Hour))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_RefreshTokenRejected(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()
	app.Use(Auth(authSvc))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Auth middleware expects "access" type tokens
	token := makeToken(t, uuid.New(), uuid.New(), "user@test.com", domain.RoleSalesRep, "refresh", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("expected 401 for refresh token used as access, got %d", resp.StatusCode)
	}
}

func TestAuth_ValidTokenSetsClaims(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()
	app.Use(Auth(authSvc))

	userID := uuid.New()
	dealerID := uuid.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		claims, err := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))
		if err != nil {
			return c.Status(500).SendString("no claims")
		}
		return c.JSON(fiber.Map{
			"user_id":   claims.UserID.String(),
			"dealer_id": claims.DealerID.String(),
			"email":     claims.Email,
			"role":      string(claims.Role),
		})
	})

	token := makeToken(t, userID, dealerID, "user@test.com", domain.RoleSalesRep, "access", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	m := bodyJSON(t, resp)
	if m["user_id"] != userID.String() {
		t.Errorf("expected user_id %s, got %v", userID, m["user_id"])
	}
	if m["dealer_id"] != dealerID.String() {
		t.Errorf("expected dealer_id %s, got %v", dealerID, m["dealer_id"])
	}
	if m["email"] != "user@test.com" {
		t.Errorf("expected email user@test.com, got %v", m["email"])
	}
	if m["role"] != "sales_rep" {
		t.Errorf("expected role sales_rep, got %v", m["role"])
	}
}

func TestAuth_DealerMismatch_NonAdmin_Forbidden(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()

	tokenDealerID := uuid.New()
	tenantDealerID := uuid.New()

	// Simulated tenant middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(domain.LocalsDealerID, tenantDealerID)
		return c.Next()
	})
	app.Use(Auth(authSvc))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	token := makeToken(t, uuid.New(), tokenDealerID, "user@test.com", domain.RoleSalesRep, "access", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 403 {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
	m := bodyJSON(t, resp)
	if m["error"] != "token dealer does not match tenant" {
		t.Errorf("expected 'token dealer does not match tenant', got %v", m["error"])
	}
}

func TestAuth_DealerMismatch_PlatformAdmin_Allowed(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()

	tokenDealerID := uuid.New()
	tenantDealerID := uuid.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals(domain.LocalsDealerID, tenantDealerID)
		return c.Next()
	})
	app.Use(Auth(authSvc))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	token := makeToken(t, uuid.New(), tokenDealerID, "admin@test.com", domain.RolePlatformAdmin, "access", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200, got %d; body: %s", resp.StatusCode, string(body))
	}
}

func TestAuth_DealerMatch_Allowed(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()

	dealerID := uuid.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals(domain.LocalsDealerID, dealerID)
		return c.Next()
	})
	app.Use(Auth(authSvc))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	token := makeToken(t, uuid.New(), dealerID, "user@test.com", domain.RoleSalesRep, "access", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuth_CaseInsensitiveBearerPrefix(t *testing.T) {
	authSvc := newTestAuthService()
	app := fiber.New()
	app.Use(Auth(authSvc))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	token := makeToken(t, uuid.New(), uuid.New(), "user@test.com", domain.RoleSalesRep, "access", time.Now().Add(1*time.Hour))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "BEARER "+token)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200 for case-insensitive bearer, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Logging middleware
// ---------------------------------------------------------------------------

func TestLogging_DoesNotPanicOnSuccess(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Use(Logging())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLogging_DoesNotPanicOnError(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Use(Logging())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.Status(500).SendString("error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 500 {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestLogging_WithClaims(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(domain.LocalsClaims, &domain.JWTClaims{
			UserID:   uuid.New(),
			DealerID: uuid.New(),
			Email:    "test@test.com",
			Role:     domain.RoleContractor,
		})
		return c.Next()
	})
	app.Use(Logging())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLogging_WithoutRequestID(t *testing.T) {
	app := fiber.New()
	// Intentionally skip RequestID middleware
	app.Use(Logging())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Integration: Auth + RequireRole
// ---------------------------------------------------------------------------

func TestAuthAndRequireRole_Integration(t *testing.T) {
	authSvc := newTestAuthService()
	dealerID := uuid.New()

	app := fiber.New()
	app.Use(Auth(authSvc))
	app.Get("/admin", RequireRole(domain.RoleDealerAdmin), func(c *fiber.Ctx) error {
		return c.SendString("admin ok")
	})

	t.Run("dealer_admin_allowed", func(t *testing.T) {
		token := makeToken(t, uuid.New(), dealerID, "admin@test.com", domain.RoleDealerAdmin, "access", time.Now().Add(1*time.Hour))
		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("contractor_blocked", func(t *testing.T) {
		token := makeToken(t, uuid.New(), dealerID, "user@test.com", domain.RoleContractor, "access", time.Now().Add(1*time.Hour))
		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 403 {
			t.Errorf("expected 403, got %d", resp.StatusCode)
		}
	})

	t.Run("no_token_unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 401 {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})
}

// ---------------------------------------------------------------------------
// All roles enumeration for RequireRole
// ---------------------------------------------------------------------------

func TestRequireRole_AllRoles(t *testing.T) {
	allRoles := []domain.Role{
		domain.RolePlatformAdmin,
		domain.RoleDealerAdmin,
		domain.RoleSalesRep,
		domain.RoleContractor,
	}

	for _, allowedRole := range allRoles {
		for _, userRole := range allRoles {
			name := fmt.Sprintf("allowed_%s_user_%s", allowedRole, userRole)
			t.Run(name, func(t *testing.T) {
				app := fiber.New()
				app.Use(func(c *fiber.Ctx) error {
					c.Locals(domain.LocalsClaims, &domain.JWTClaims{
						UserID:   uuid.New(),
						DealerID: uuid.New(),
						Email:    "test@test.com",
						Role:     userRole,
					})
					return c.Next()
				})
				app.Get("/test", RequireRole(allowedRole), func(c *fiber.Ctx) error {
					return c.SendString("ok")
				})

				req := httptest.NewRequest("GET", "/test", nil)
				resp, err := app.Test(req, -1)
				if err != nil {
					t.Fatal(err)
				}
				resp.Body.Close()

				if userRole == allowedRole {
					if resp.StatusCode != 200 {
						t.Errorf("expected 200, got %d", resp.StatusCode)
					}
				} else {
					if resp.StatusCode != 403 {
						t.Errorf("expected 403, got %d", resp.StatusCode)
					}
				}
			})
		}
	}
}

// ---------------------------------------------------------------------------
// Tenant middleware (header-based only; subdomain/slug need DB)
// ---------------------------------------------------------------------------

func TestTenant_ValidHeaderSetsLocals(t *testing.T) {
	app := fiber.New()
	dealerID := uuid.New()
	app.Use(Tenant(nil))
	app.Get("/test", func(c *fiber.Ctx) error {
		val := c.Locals(domain.LocalsDealerID)
		if val == nil {
			return c.Status(500).SendString("no dealer ID")
		}
		return c.SendString(val.(uuid.UUID).String())
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", dealerID.String())
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != dealerID.String() {
		t.Errorf("expected %s, got %s", dealerID, string(body))
	}
}

func TestTenant_InvalidHeaderUUID(t *testing.T) {
	app := fiber.New()
	app.Use(Tenant(nil))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "not-a-uuid")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestTenant_MissingAllIdentifiers(t *testing.T) {
	app := fiber.New()
	// Pass nil store - header path doesn't need DB
	app.Use(Tenant(nil))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		resp.Body.Close()
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	m := bodyJSON(t, resp)
	errMsg, _ := m["error"].(string)
	if !strings.Contains(errMsg, "tenant identification required") {
		t.Errorf("expected tenant error message, got %v", m["error"])
	}
}

// ---------------------------------------------------------------------------
// RequestTimeout middleware
// ---------------------------------------------------------------------------

func TestRequestTimeout_ReturnsSuccessForFastHandler(t *testing.T) {
	app := fiber.New()
	app.Use(RequestTimeout(5 * time.Second))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequestTimeout_DoesNotPanicWithNilContext(t *testing.T) {
	// Ensure the middleware wraps without error
	handler := RequestTimeout(30 * time.Second)
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}

// ---------------------------------------------------------------------------
// Logging middleware (additional coverage)
// ---------------------------------------------------------------------------

func TestLogging_CallsNextAndReturnsCorrectStatus(t *testing.T) {
	handlerCalled := false
	app := fiber.New()
	app.Use(RequestID())
	app.Use(Logging())
	app.Get("/test", func(c *fiber.Ctx) error {
		handlerCalled = true
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if !handlerCalled {
		t.Error("expected handler to be called via c.Next()")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLogging_PropagatesHandlerError(t *testing.T) {
	app := fiber.New()
	app.Use(Logging())
	app.Get("/err", func(c *fiber.Ctx) error {
		return c.Status(400).SendString("bad request")
	})

	req := httptest.NewRequest("GET", "/err", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// RequestID middleware (additional coverage)
// ---------------------------------------------------------------------------

func TestRequestID_StoredInUserContext(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		// Verify it is also stored in context.Context via WithRequestID
		rid := domain.RequestIDFromContext(c.UserContext())
		return c.SendString(rid)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	rid := string(body)
	if rid == "" {
		t.Error("expected request ID in user context, got empty string")
	}
	if _, err := uuid.Parse(rid); err != nil {
		t.Errorf("expected valid UUID in user context, got %q: %v", rid, err)
	}
}

func TestRequestID_CustomHeaderPreservedInContext(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		rid := domain.RequestIDFromContext(c.UserContext())
		return c.SendString(rid)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "custom-id-abc")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "custom-id-abc" {
		t.Errorf("expected custom-id-abc in user context, got %q", string(body))
	}
}

// ---------------------------------------------------------------------------
// RequestTimeout middleware (additional coverage)
// ---------------------------------------------------------------------------

func TestRequestTimeout_WrapsHandler(t *testing.T) {
	handlerCalled := false
	app := fiber.New()
	app.Use(RequestTimeout(5 * time.Second))
	app.Get("/test", func(c *fiber.Ctx) error {
		handlerCalled = true
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if !handlerCalled {
		t.Error("expected handler to be called through timeout middleware")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequestTimeout_DifferentDurations(t *testing.T) {
	durations := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		30 * time.Second,
		1 * time.Minute,
	}
	for _, d := range durations {
		t.Run(d.String(), func(t *testing.T) {
			handler := RequestTimeout(d)
			if handler == nil {
				t.Errorf("expected non-nil handler for duration %v", d)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AuditLog middleware
// ---------------------------------------------------------------------------

func TestAuditLog_SkipsGET(t *testing.T) {
	handlerCalled := false
	app := fiber.New()
	app.Use(AuditLog())
	app.Get("/test", func(c *fiber.Ctx) error {
		handlerCalled = true
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuditLog_SkipsHEAD(t *testing.T) {
	app := fiber.New()
	app.Use(AuditLog())
	app.Head("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("HEAD", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuditLog_SkipsOPTIONS(t *testing.T) {
	app := fiber.New()
	app.Use(AuditLog())
	app.Options("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestAuditLog_LogsPOST(t *testing.T) {
	handlerCalled := false
	app := fiber.New()
	app.Use(AuditLog())
	app.Post("/test", func(c *fiber.Ctx) error {
		handlerCalled = true
		return c.Status(201).SendString("created")
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if !handlerCalled {
		t.Error("expected handler to be called for POST")
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAuditLog_LogsPUT(t *testing.T) {
	app := fiber.New()
	app.Use(AuditLog())
	app.Put("/test", func(c *fiber.Ctx) error {
		return c.SendString("updated")
	})

	req := httptest.NewRequest("PUT", "/test", strings.NewReader(`{}`))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuditLog_LogsDELETE(t *testing.T) {
	app := fiber.New()
	app.Use(AuditLog())
	app.Delete("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestAuditLog_LogsPATCH(t *testing.T) {
	app := fiber.New()
	app.Use(AuditLog())
	app.Patch("/test", func(c *fiber.Ctx) error {
		return c.SendString("patched")
	})

	req := httptest.NewRequest("PATCH", "/test", strings.NewReader(`{}`))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuditLog_WithClaims(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(domain.LocalsClaims, &domain.JWTClaims{
			UserID:   uuid.New(),
			DealerID: uuid.New(),
			Email:    "admin@test.com",
			Role:     domain.RoleDealerAdmin,
		})
		return c.Next()
	})
	app.Use(AuditLog())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.Status(201).SendString("created")
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAuditLog_WithRequestID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Use(AuditLog())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.Status(201).SendString("created")
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAuditLog_WithXRequestIDHeader(t *testing.T) {
	app := fiber.New()
	// Do NOT use RequestID middleware, but set the header directly
	app.Use(AuditLog())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.Status(201).SendString("created")
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("X-Request-ID", "header-based-id")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// CSRFProtection middleware (additional edge cases)
// ---------------------------------------------------------------------------

func TestCSRFProtection_AllowsPATCHWithHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Patch("/test", func(c *fiber.Ctx) error {
		return c.SendString("patched")
	})

	req := httptest.NewRequest("PATCH", "/test", strings.NewReader(`{}`))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("PATCH with X-Requested-With: status = %d, want 200", resp.StatusCode)
	}
}

func TestCSRFProtection_BlocksPATCHWithoutHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Patch("/test", func(c *fiber.Ctx) error {
		return c.SendString("patched")
	})

	req := httptest.NewRequest("PATCH", "/test", strings.NewReader(`{}`))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Errorf("PATCH without X-Requested-With: status = %d, want 403", resp.StatusCode)
	}
}

func TestCSRFProtection_AllowsPUTWithHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Put("/test", func(c *fiber.Ctx) error {
		return c.SendString("updated")
	})

	req := httptest.NewRequest("PUT", "/test", strings.NewReader(`{}`))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("PUT with X-Requested-With: status = %d, want 200", resp.StatusCode)
	}
}

func TestCSRFProtection_ErrorResponseContainsMessage(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	m := bodyJSON(t, resp)
	errMsg, _ := m["error"].(string)
	if errMsg != "missing X-Requested-With header" {
		t.Errorf("expected 'missing X-Requested-With header', got %q", errMsg)
	}
}

func TestCSRFProtection_AnyHeaderValueAccepted(t *testing.T) {
	headerValues := []string{
		"XMLHttpRequest",
		"fetch",
		"anything",
		"1",
		"true",
	}

	for _, val := range headerValues {
		t.Run(val, func(t *testing.T) {
			app := fiber.New()
			app.Use(CSRFProtection())
			app.Post("/test", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
			req.Header.Set("X-Requested-With", val)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()

			if resp.StatusCode != 200 {
				t.Errorf("POST with X-Requested-With=%q: status = %d, want 200", val, resp.StatusCode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Idempotency middleware (additional edge cases)
// ---------------------------------------------------------------------------

func TestIdempotency_PUTWithKey_Passthrough(t *testing.T) {
	cache := NewIdempotencyCache(5 * time.Minute)
	callCount := 0
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Put("/test", func(c *fiber.Ctx) error {
		callCount++
		return c.SendString("updated")
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("PUT", "/test", strings.NewReader(`{}`))
		req.Header.Set("Idempotency-Key", "same-key")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	if callCount != 3 {
		t.Errorf("PUT with idempotency key should not be cached: handler called %d times, want 3", callCount)
	}
}

func TestIdempotency_DELETEWithKey_Passthrough(t *testing.T) {
	cache := NewIdempotencyCache(5 * time.Minute)
	callCount := 0
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Delete("/test", func(c *fiber.Ctx) error {
		callCount++
		return c.SendStatus(204)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("DELETE", "/test", nil)
		req.Header.Set("Idempotency-Key", "delete-key")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	if callCount != 2 {
		t.Errorf("DELETE with idempotency key should not be cached: handler called %d times, want 2", callCount)
	}
}

func TestIdempotency_POSTWithKey_CachesCorrectStatus(t *testing.T) {
	cache := NewIdempotencyCache(5 * time.Minute)
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.Status(202).SendString("accepted")
	})

	key := "status-cache-key"

	// First request
	req1 := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req1.Header.Set("Idempotency-Key", key)
	resp1, err := app.Test(req1, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()

	if resp1.StatusCode != 202 {
		t.Errorf("first request status = %d, want 202", resp1.StatusCode)
	}

	// Second request - should return cached 202
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req2.Header.Set("Idempotency-Key", key)
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()

	if resp2.StatusCode != 202 {
		t.Errorf("cached request status = %d, want 202", resp2.StatusCode)
	}
	if resp2.Header.Get("X-Idempotency-Replay") != "true" {
		t.Error("expected X-Idempotency-Replay header on cached response")
	}
}

func TestIdempotency_GETWithoutKey_Passthrough(t *testing.T) {
	cache := NewIdempotencyCache(5 * time.Minute)
	callCount := 0
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Get("/test", func(c *fiber.Ctx) error {
		callCount++
		return c.SendString("ok")
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	if callCount != 3 {
		t.Errorf("GET without key: handler called %d times, want 3", callCount)
	}
}
