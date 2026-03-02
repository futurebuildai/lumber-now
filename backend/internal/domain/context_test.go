package domain

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// WithRequestID / RequestIDFromContext
// ---------------------------------------------------------------------------

func TestWithRequestID_RoundTrip(t *testing.T) {
	ctx := context.Background()
	rid := "test-request-id-123"
	ctx = WithRequestID(ctx, rid)

	got := RequestIDFromContext(ctx)
	if got != rid {
		t.Errorf("RequestIDFromContext() = %q, want %q", got, rid)
	}
}

func TestRequestIDFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	got := RequestIDFromContext(ctx)
	if got != "" {
		t.Errorf("RequestIDFromContext(empty) = %q, want empty string", got)
	}
}

func TestWithRequestID_OverwritesPrevious(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "first")
	ctx = WithRequestID(ctx, "second")

	got := RequestIDFromContext(ctx)
	if got != "second" {
		t.Errorf("RequestIDFromContext() = %q, want %q", got, "second")
	}
}

func TestWithRequestID_EmptyString(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "")

	got := RequestIDFromContext(ctx)
	if got != "" {
		t.Errorf("RequestIDFromContext() = %q, want empty string", got)
	}
}

func TestWithRequestID_UUID(t *testing.T) {
	ctx := context.Background()
	rid := uuid.New().String()
	ctx = WithRequestID(ctx, rid)

	got := RequestIDFromContext(ctx)
	if got != rid {
		t.Errorf("RequestIDFromContext() = %q, want %q", got, rid)
	}
	// Verify it's a valid UUID
	if _, err := uuid.Parse(got); err != nil {
		t.Errorf("expected valid UUID, got %q: %v", got, err)
	}
}

func TestRequestIDFromContext_WrongType(t *testing.T) {
	// Inject a non-string value under the same key
	ctx := context.WithValue(context.Background(), RequestIDKey, 12345)
	got := RequestIDFromContext(ctx)
	if got != "" {
		t.Errorf("RequestIDFromContext(wrong type) = %q, want empty string", got)
	}
}

func TestRequestIDFromContext_NilContext(t *testing.T) {
	// context.Background() is never nil, but WithValue with a nil-ish context
	// would panic; test that our function works with a plain background context.
	got := RequestIDFromContext(context.Background())
	if got != "" {
		t.Errorf("RequestIDFromContext(background) = %q, want empty string", got)
	}
}

// ---------------------------------------------------------------------------
// DealerIDFromLocals (tested here for the context package; also in types_test)
// ---------------------------------------------------------------------------

func TestDealerIDFromLocals_Nil(t *testing.T) {
	_, err := DealerIDFromLocals(nil)
	if err != ErrTenantMissing {
		t.Errorf("DealerIDFromLocals(nil) error = %v, want ErrTenantMissing", err)
	}
}

func TestDealerIDFromLocals_WrongType(t *testing.T) {
	_, err := DealerIDFromLocals("not-a-uuid")
	if err != ErrTenantMissing {
		t.Errorf("DealerIDFromLocals(string) error = %v, want ErrTenantMissing", err)
	}
}

func TestDealerIDFromLocals_ValidUUID(t *testing.T) {
	id := uuid.New()
	got, err := DealerIDFromLocals(id)
	if err != nil {
		t.Errorf("DealerIDFromLocals(valid) error = %v, want nil", err)
	}
	if got != id {
		t.Errorf("DealerIDFromLocals(valid) = %v, want %v", got, id)
	}
}

// ---------------------------------------------------------------------------
// ClaimsFromLocals (tested here for the context package; also in types_test)
// ---------------------------------------------------------------------------

func TestClaimsFromLocals_Nil(t *testing.T) {
	_, err := ClaimsFromLocals(nil)
	if err != ErrUnauthorized {
		t.Errorf("ClaimsFromLocals(nil) error = %v, want ErrUnauthorized", err)
	}
}

func TestClaimsFromLocals_WrongType(t *testing.T) {
	_, err := ClaimsFromLocals("not-claims")
	if err != ErrUnauthorized {
		t.Errorf("ClaimsFromLocals(string) error = %v, want ErrUnauthorized", err)
	}
}

func TestClaimsFromLocals_ValidClaims(t *testing.T) {
	claims := &JWTClaims{
		UserID:   uuid.New(),
		DealerID: uuid.New(),
		Email:    "test@example.com",
		Role:     RoleSalesRep,
	}
	got, err := ClaimsFromLocals(claims)
	if err != nil {
		t.Errorf("ClaimsFromLocals(valid) error = %v, want nil", err)
	}
	if got != claims {
		t.Errorf("ClaimsFromLocals(valid) returned different pointer")
	}
	if got.UserID != claims.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, claims.UserID)
	}
	if got.Email != claims.Email {
		t.Errorf("Email = %v, want %v", got.Email, claims.Email)
	}
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestLocalsConstants(t *testing.T) {
	if LocalsDealerID != "dealer_id" {
		t.Errorf("LocalsDealerID = %q, want %q", LocalsDealerID, "dealer_id")
	}
	if LocalsClaims != "claims" {
		t.Errorf("LocalsClaims = %q, want %q", LocalsClaims, "claims")
	}
	if LocalsRequestID != "request_id" {
		t.Errorf("LocalsRequestID = %q, want %q", LocalsRequestID, "request_id")
	}
}

func TestRequestIDKey_IsContextKey(t *testing.T) {
	if RequestIDKey != contextKey("request_id") {
		t.Errorf("RequestIDKey = %v, want contextKey(\"request_id\")", RequestIDKey)
	}
}
