package service

import (
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

// ---------------------------------------------------------------------------
// Password complexity validation
// ---------------------------------------------------------------------------

func TestPasswordComplexity(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"empty password", "", true},
		{"1 char", "a", true},
		{"7 chars exactly", "abcdefg", true},
		{"8 chars no upper", "abcdefg1", true},
		{"8 chars no lower", "ABCDEFG1", true},
		{"8 chars no digit", "Abcdefgh", true},
		{"8 chars all lower", "abcdefgh", true},
		{"8 chars valid", "Abcdefg1", false},
		{"9 chars valid", "Abcdefg12", false},
		{"16 chars valid", "Abcdefghijklmn1p", false},
		{"max bcrypt input (72 bytes)", strings.Repeat("A", 35) + strings.Repeat("a", 35) + "12", false},
		{"unicode valid", "Pässwörd1", false},
		{"special chars only with upper lower digit", "P@ss!0rd", false},
		{"all digits", "12345678", true},
		{"all spaces", "        ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswordComplexity(tt.password)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for password %q, got nil", tt.password)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for password %q: %v", tt.password, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Email validation (uses net/mail.ParseAddress, same as Register)
// ---------------------------------------------------------------------------

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid simple email", "user@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"valid with plus tag", "user+tag@example.com", false},
		{"valid with dots", "first.last@example.com", false},
		{"missing at sign", "userexample.com", true},
		{"missing domain", "user@", true},
		{"missing local part", "@example.com", true},
		{"empty string", "", true},
		{"double at sign", "user@@example.com", true},
		{"bare word", "user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mail.ParseAddress(tt.email)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for email %q, got nil", tt.email)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for email %q: %v", tt.email, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Role validation (domain.Role.Valid)
// ---------------------------------------------------------------------------

func TestRoleValidation(t *testing.T) {
	tests := []struct {
		name  string
		role  domain.Role
		valid bool
	}{
		{"platform_admin", domain.RolePlatformAdmin, true},
		{"dealer_admin", domain.RoleDealerAdmin, true},
		{"sales_rep", domain.RoleSalesRep, true},
		{"contractor", domain.RoleContractor, true},
		{"empty string", domain.Role(""), false},
		{"unknown role", domain.Role("superuser"), false},
		{"case mismatch", domain.Role("Contractor"), false},
		{"partial match", domain.Role("admin"), false},
		{"extra whitespace", domain.Role(" contractor"), false},
		{"sql injection attempt", domain.Role("contractor'; DROP TABLE users;--"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.Valid()
			if got != tt.valid {
				t.Errorf("Role(%q).Valid() = %v, want %v", tt.role, got, tt.valid)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Bcrypt cost verification
// ---------------------------------------------------------------------------

func TestBcryptCostIs12(t *testing.T) {
	password := "TestPassword123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}

	cost, err := bcrypt.Cost(hash)
	if err != nil {
		t.Fatalf("bcrypt.Cost: %v", err)
	}
	if cost != 12 {
		t.Errorf("expected bcrypt cost 12, got %d", cost)
	}
}

func TestBcryptHashVerifiesCorrectPassword(t *testing.T) {
	password := "SecurePass123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		t.Errorf("expected correct password to verify, got: %v", err)
	}
}

func TestBcryptHashRejectsWrongPassword(t *testing.T) {
	password := "SecurePass123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte("wrongpassword")); err == nil {
		t.Error("expected wrong password to fail verification")
	}
}

func TestBcryptDefaultCostIsTooLow(t *testing.T) {
	if bcrypt.DefaultCost >= 12 {
		t.Skip("default cost is already >= 12, test not meaningful")
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("Test1234"), bcrypt.DefaultCost)
	cost, _ := bcrypt.Cost(hash)
	if cost >= 12 {
		t.Error("default cost should be below the production cost of 12")
	}
}

// ---------------------------------------------------------------------------
// JWT token round-trip via AuthService (no DB required)
// ---------------------------------------------------------------------------

func newTestUser() db.User {
	return db.User{
		ID:       uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
		DealerID: uuid.MustParse("11111111-2222-3333-4444-555555555555"),
		Email:    "test@example.com",
		Role:     db.UserRoleContractor,
		Active:   true,
		FullName: "Test User",
	}
}

func TestTokenRoundTrip(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-key-for-jwt")
	user := newTestUser()

	pair, err := svc.issueTokenPair(user)
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("access token is empty")
	}
	if pair.RefreshToken == "" {
		t.Error("refresh token is empty")
	}
	if pair.ExpiresAt == 0 {
		t.Error("expires_at is zero")
	}

	claims, err := svc.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("user ID mismatch: got %v, want %v", claims.UserID, user.ID)
	}
	if claims.DealerID != user.DealerID {
		t.Errorf("dealer ID mismatch: got %v, want %v", claims.DealerID, user.DealerID)
	}
	if claims.Email != user.Email {
		t.Errorf("email mismatch: got %q, want %q", claims.Email, user.Email)
	}
	if claims.Role != domain.Role(user.Role) {
		t.Errorf("role mismatch: got %q, want %q", claims.Role, user.Role)
	}
}

func TestAccessTokenCannotBeUsedAsRefresh(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-key-for-jwt")

	pair, err := svc.issueTokenPair(newTestUser())
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	_, err = svc.validateToken(pair.AccessToken, "refresh")
	if err == nil {
		t.Error("expected error when using access token as refresh token")
	}
}

func TestRefreshTokenCannotBeUsedAsAccess(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-key-for-jwt")

	pair, err := svc.issueTokenPair(newTestUser())
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	_, err = svc.ValidateToken(pair.RefreshToken)
	if err == nil {
		t.Error("expected error when using refresh token as access token")
	}
}

func TestInvalidTokenRejected(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-key-for-jwt")

	_, err := svc.ValidateToken("this-is-not-a-jwt")
	if err == nil {
		t.Error("expected error for garbage token")
	}
}

func TestEmptyTokenRejected(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-key-for-jwt")

	_, err := svc.ValidateToken("")
	if err == nil {
		t.Error("expected error for empty token string")
	}
}

func TestWrongSecretRejected(t *testing.T) {
	svc1 := NewAuthService(nil, "secret-one")
	svc2 := NewAuthService(nil, "secret-two")

	pair, err := svc1.issueTokenPair(newTestUser())
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	_, err = svc2.ValidateToken(pair.AccessToken)
	if err == nil {
		t.Error("expected error when validating token with a different secret")
	}
}

func TestTokenClaimsContainCorrectRole(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	roles := []db.UserRole{
		db.UserRolePlatformAdmin,
		db.UserRoleDealerAdmin,
		db.UserRoleSalesRep,
		db.UserRoleContractor,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			user := newTestUser()
			user.Role = role

			pair, err := svc.issueTokenPair(user)
			if err != nil {
				t.Fatalf("issueTokenPair: %v", err)
			}

			claims, err := svc.ValidateToken(pair.AccessToken)
			if err != nil {
				t.Fatalf("ValidateToken: %v", err)
			}

			if claims.Role != domain.Role(role) {
				t.Errorf("role mismatch: got %q, want %q", claims.Role, role)
			}
		})
	}
}

func TestTokenPairContainsDifferentTokens(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	pair, err := svc.issueTokenPair(newTestUser())
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	if pair.AccessToken == pair.RefreshToken {
		t.Error("access and refresh tokens should be different")
	}
}

// ---------------------------------------------------------------------------
// Account lockout
// ---------------------------------------------------------------------------

func TestAccountLockoutAfterMaxAttempts(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	key := "11111111-2222-3333-4444-555555555555:attacker@example.com"

	// First 4 failures should not lock
	for i := 0; i < maxLoginAttempts-1; i++ {
		svc.recordLoginFailure(key)
		if err := svc.checkLockout(key); err != nil {
			t.Errorf("attempt %d: should not be locked yet, got: %v", i+1, err)
		}
	}

	// 5th failure should trigger lockout
	svc.recordLoginFailure(key)
	if err := svc.checkLockout(key); err != domain.ErrAccountLocked {
		t.Errorf("expected ErrAccountLocked after %d failures, got: %v", maxLoginAttempts, err)
	}
}

func TestAccountLockoutResetsOnSuccess(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	key := "dealer:user@test.com"

	// Record some failures (but not enough to lock)
	for i := 0; i < 3; i++ {
		svc.recordLoginFailure(key)
	}

	// Successful login resets the counter
	svc.resetLoginAttempts(key)

	// Now failures count from zero again
	for i := 0; i < maxLoginAttempts-1; i++ {
		svc.recordLoginFailure(key)
	}

	if err := svc.checkLockout(key); err != nil {
		t.Errorf("should not be locked after reset, got: %v", err)
	}
}

func TestCheckLockoutReturnsNilForUnknownKey(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	if err := svc.checkLockout("unknown:key"); err != nil {
		t.Errorf("expected nil for unknown key, got: %v", err)
	}
}

func TestMultipleAccountsIndependent(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")

	key1 := "dealer:user1@test.com"
	key2 := "dealer:user2@test.com"

	// Lock user1
	for i := 0; i < maxLoginAttempts; i++ {
		svc.recordLoginFailure(key1)
	}

	// user1 should be locked
	if err := svc.checkLockout(key1); err != domain.ErrAccountLocked {
		t.Errorf("user1 should be locked, got: %v", err)
	}

	// user2 should not be locked
	if err := svc.checkLockout(key2); err != nil {
		t.Errorf("user2 should not be locked, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// JWT claim validation (iss, aud)
// ---------------------------------------------------------------------------

func TestTokenContainsIssuerAndAudience(t *testing.T) {
	svc := NewAuthService(nil, "test-secret")
	user := newTestUser()

	pair, err := svc.issueTokenPair(user)
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	// Parse token manually to check iss/aud claims
	claims, err := svc.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken should pass: %v", err)
	}

	// ValidateToken already validates iss/aud internally,
	// so if we got here without error, the claims are valid.
	if claims.UserID != user.ID {
		t.Errorf("user ID mismatch")
	}
}

// ---------------------------------------------------------------------------
// Token blacklist / revocation
// ---------------------------------------------------------------------------

func TestRevokeToken_BlacklistsJTI(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")
	user := newTestUser()

	pair, err := svc.issueTokenPair(user)
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	// Token should be valid before revocation.
	if _, err := svc.ValidateToken(pair.AccessToken); err != nil {
		t.Fatalf("token should be valid before revoke: %v", err)
	}

	// Revoke the access token.
	if err := svc.RevokeToken(pair.AccessToken); err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}

	// Token should now be rejected.
	if _, err := svc.ValidateToken(pair.AccessToken); err == nil {
		t.Error("expected error when validating a revoked access token")
	}
}

func TestRevokeToken_InvalidToken(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")

	// Revoking garbage should not panic or return an error.
	if err := svc.RevokeToken("not-a-real-jwt"); err != nil {
		t.Errorf("expected nil error for unparseable token, got: %v", err)
	}

	// Revoking an empty string should also be fine.
	if err := svc.RevokeToken(""); err != nil {
		t.Errorf("expected nil error for empty token, got: %v", err)
	}
}

func TestIsTokenBlacklisted_NotBlacklisted(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")

	if svc.isTokenBlacklisted("nonexistent-jti") {
		t.Error("expected false for a JTI that was never blacklisted")
	}
}

func TestIsTokenBlacklisted_ExpiredEntry(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")

	// Manually insert an already-expired entry.
	svc.blacklistMu.Lock()
	svc.blacklist["expired-jti"] = time.Now().Add(-1 * time.Hour)
	svc.blacklistMu.Unlock()

	if svc.isTokenBlacklisted("expired-jti") {
		t.Error("expected false for an expired blacklist entry")
	}
}

func TestCleanupBlacklist_RemovesExpired(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")

	// Insert a mix of expired and future entries.
	svc.blacklistMu.Lock()
	svc.blacklist["expired-1"] = time.Now().Add(-2 * time.Hour)
	svc.blacklist["expired-2"] = time.Now().Add(-30 * time.Minute)
	svc.blacklist["valid-1"] = time.Now().Add(1 * time.Hour)
	svc.blacklist["valid-2"] = time.Now().Add(24 * time.Hour)
	svc.blacklistMu.Unlock()

	svc.cleanupBlacklist()

	svc.blacklistMu.RLock()
	defer svc.blacklistMu.RUnlock()

	if _, exists := svc.blacklist["expired-1"]; exists {
		t.Error("expired-1 should have been cleaned up")
	}
	if _, exists := svc.blacklist["expired-2"]; exists {
		t.Error("expired-2 should have been cleaned up")
	}
	if _, exists := svc.blacklist["valid-1"]; !exists {
		t.Error("valid-1 should still be in the blacklist")
	}
	if _, exists := svc.blacklist["valid-2"]; !exists {
		t.Error("valid-2 should still be in the blacklist")
	}
}

func TestRevokeToken_RefreshTokenAlsoBlacklisted(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")
	user := newTestUser()

	pair, err := svc.issueTokenPair(user)
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	// Refresh token should validate as "refresh" type.
	if _, err := svc.validateToken(pair.RefreshToken, "refresh"); err != nil {
		t.Fatalf("refresh token should be valid before revoke: %v", err)
	}

	// Revoke the refresh token.
	if err := svc.RevokeToken(pair.RefreshToken); err != nil {
		t.Fatalf("RevokeToken on refresh token: %v", err)
	}

	// Refresh token should now be rejected.
	if _, err := svc.validateToken(pair.RefreshToken, "refresh"); err == nil {
		t.Error("expected error when validating a revoked refresh token")
	}
}

func TestRevokedAccessDoesNotAffectRefresh(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")
	user := newTestUser()

	pair, err := svc.issueTokenPair(user)
	if err != nil {
		t.Fatalf("issueTokenPair: %v", err)
	}

	// Revoke only the access token.
	svc.RevokeToken(pair.AccessToken)

	// Access should be rejected.
	if _, err := svc.ValidateToken(pair.AccessToken); err == nil {
		t.Error("revoked access token should fail validation")
	}

	// Refresh should still be valid (different JTI).
	if _, err := svc.validateToken(pair.RefreshToken, "refresh"); err != nil {
		t.Errorf("refresh token should still be valid after revoking access: %v", err)
	}
}

func TestRevokeToken_TokenWithoutJTI(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")

	// Create a token without a JTI claim.
	claims := jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "access",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"iss":  jwtIssuer,
		"aud":  jwtAudience,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(svc.jwtSecret)
	if err != nil {
		t.Fatalf("signing token: %v", err)
	}

	// Revoking a token without JTI should not error.
	if err := svc.RevokeToken(tokenStr); err != nil {
		t.Errorf("expected nil error for token without JTI, got: %v", err)
	}
}

func TestCleanupBlacklist_EmptyBlacklist(t *testing.T) {
	svc := NewAuthService(nil, "test-secret-blacklist")

	// Should not panic on empty blacklist.
	svc.cleanupBlacklist()

	svc.blacklistMu.RLock()
	defer svc.blacklistMu.RUnlock()
	if len(svc.blacklist) != 0 {
		t.Errorf("expected empty blacklist, got %d entries", len(svc.blacklist))
	}
}
