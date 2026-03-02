package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"sync"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

const (
	jwtIssuer        = "lumber-now"
	jwtAudience      = "lumber-now-api"
	maxLoginAttempts = 5
	lockoutDuration  = 15 * time.Minute

	// blacklistSweepInterval controls how often a full blacklist cleanup
	// is performed lazily (every N calls to RevokeToken).
	blacklistSweepInterval = 100
)

type loginAttemptInfo struct {
	count       int
	lockedUntil time.Time
}

type AuthService struct {
	store     *store.Store
	jwtSecret []byte

	loginMu       sync.Mutex
	loginAttempts map[string]*loginAttemptInfo

	// Token blacklist for revocation (in-memory, cleared on restart).
	blacklistMu    sync.RWMutex
	blacklist      map[string]time.Time // jti -> expiry time
	revokeCounter  uint64               // tracks calls to RevokeToken for lazy sweep
}

func NewAuthService(s *store.Store, jwtSecret string) *AuthService {
	return &AuthService{
		store:         s,
		jwtSecret:     []byte(jwtSecret),
		loginAttempts: make(map[string]*loginAttemptInfo),
		blacklist:     make(map[string]time.Time),
	}
}

type LoginInput struct {
	DealerID uuid.UUID
	Email    string
	Password string
}

type RegisterInput struct {
	DealerID uuid.UUID
	Email    string
	Password string
	FullName string
	Phone    string
	Role     domain.Role
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*db.User, error) {
	if !input.Role.Valid() {
		return nil, domain.ErrInvalidRole
	}

	// Email validation
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return nil, domain.ErrInvalidInput
	}

	// Password strength validation (length + complexity)
	if err := validatePasswordComplexity(input.Password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.store.Queries.CreateUser(ctx, db.CreateUserParams{
		DealerID:     input.DealerID,
		Email:        input.Email,
		PasswordHash: string(hash),
		FullName:     input.FullName,
		Phone:        input.Phone,
		Role:         db.UserRole(input.Role),
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	lockoutKey := fmt.Sprintf("%s:%s", input.DealerID, input.Email)

	// Check DB-backed lockout first (survives restarts, works multi-instance)
	if s.store != nil {
		if err := s.checkDBLockout(ctx, input.DealerID, input.Email); err != nil {
			return nil, err
		}
	} else {
		// Fallback to in-memory lockout (for tests with nil store)
		if err := s.checkLockout(lockoutKey); err != nil {
			return nil, err
		}
	}

	user, err := s.store.Queries.GetUserByEmail(ctx, db.GetUserByEmailParams{
		DealerID: input.DealerID,
		Email:    input.Email,
	})
	if err != nil {
		s.recordLoginFailureDB(ctx, input.DealerID, input.Email, lockoutKey)
		return nil, domain.ErrUnauthorized
	}

	if !user.Active {
		return nil, domain.ErrForbidden
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		s.recordLoginFailureDB(ctx, input.DealerID, input.Email, lockoutKey)
		return nil, domain.ErrUnauthorized
	}

	// Successful login - reset attempts in both DB and memory
	s.resetLoginAttemptsDB(ctx, input.DealerID, input.Email, lockoutKey)
	return s.issueTokenPair(user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.validateToken(refreshToken, "refresh")
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Blacklist the old refresh token to prevent reuse.
	s.RevokeToken(refreshToken)

	user, err := s.store.Queries.GetUser(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if !user.Active {
		return nil, domain.ErrForbidden
	}

	return s.issueTokenPair(user)
}

func (s *AuthService) ValidateToken(tokenStr string) (*domain.JWTClaims, error) {
	return s.validateToken(tokenStr, "access")
}

func (s *AuthService) issueTokenPair(user db.User) (*TokenPair, error) {
	expiresAt := time.Now().Add(15 * time.Minute)
	accessToken, err := s.createToken(user, "access", expiresAt)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.createToken(user, "refresh", time.Now().Add(7*24*time.Hour))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func (s *AuthService) createToken(user db.User, tokenType string, expiresAt time.Time) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"dealer_id": user.DealerID.String(),
		"email":     user.Email,
		"role":      string(user.Role),
		"type":      tokenType,
		"jti":       uuid.New().String(),
		"exp":       expiresAt.Unix(),
		"iat":       now.Unix(),
		"nbf":       now.Unix(),
		"iss":       jwtIssuer,
		"aud":       jwtAudience,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) validateToken(tokenStr, expectedType string) (*domain.JWTClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != expectedType {
		return nil, domain.ErrUnauthorized
	}

	// Check token blacklist (revoked tokens).
	jti, _ := claims["jti"].(string)
	if jti != "" && s.isTokenBlacklisted(jti) {
		return nil, domain.ErrUnauthorized
	}

	// Validate issuer
	iss, _ := claims["iss"].(string)
	if iss != jwtIssuer {
		return nil, domain.ErrUnauthorized
	}

	// Validate audience
	aud, _ := claims["aud"].(string)
	if aud != jwtAudience {
		return nil, domain.ErrUnauthorized
	}

	subStr, ok := claims["sub"].(string)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	userID, err := uuid.Parse(subStr)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	dealerStr, ok := claims["dealer_id"].(string)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	dealerID, err := uuid.Parse(dealerStr)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	emailStr, _ := claims["email"].(string)
	roleStr, _ := claims["role"].(string)

	return &domain.JWTClaims{
		UserID:   userID,
		DealerID: dealerID,
		Email:    emailStr,
		Role:     domain.Role(roleStr),
	}, nil
}

// validatePasswordComplexity checks length and character class requirements.
func validatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}

	return nil
}

// ---------------------------------------------------------------------------
// Token blacklist
// ---------------------------------------------------------------------------

// RevokeToken adds a token's JTI to the blacklist until it naturally expires.
// It uses ParseUnverified so that even expired or otherwise-invalid tokens can
// be blacklisted (e.g. during logout the token may be near expiry).
func (s *AuthService) RevokeToken(tokenStr string) error {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil // silently ignore unparseable tokens
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	jti, _ := claims["jti"].(string)
	if jti == "" {
		return nil
	}

	expFloat, _ := claims["exp"].(float64)
	expiry := time.Unix(int64(expFloat), 0)

	s.blacklistMu.Lock()
	s.blacklist[jti] = expiry
	s.revokeCounter++
	counter := s.revokeCounter
	s.blacklistMu.Unlock()

	// Lazy sweep: every blacklistSweepInterval calls, clean up expired entries.
	if counter%blacklistSweepInterval == 0 {
		s.cleanupBlacklist()
	}

	return nil
}

// isTokenBlacklisted checks if a token's JTI has been revoked.
func (s *AuthService) isTokenBlacklisted(jti string) bool {
	s.blacklistMu.RLock()
	defer s.blacklistMu.RUnlock()
	expiry, exists := s.blacklist[jti]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

// cleanupBlacklist removes expired entries from the blacklist.
func (s *AuthService) cleanupBlacklist() {
	s.blacklistMu.Lock()
	defer s.blacklistMu.Unlock()
	now := time.Now()
	for jti, expiry := range s.blacklist {
		if now.After(expiry) {
			delete(s.blacklist, jti)
		}
	}
}

// ---------------------------------------------------------------------------
// Account lockout (in-memory)
// ---------------------------------------------------------------------------

// checkLockout returns ErrAccountLocked if the account is temporarily locked.
func (s *AuthService) checkLockout(key string) error {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()

	info, exists := s.loginAttempts[key]
	if !exists {
		return nil
	}

	if !info.lockedUntil.IsZero() && time.Now().Before(info.lockedUntil) {
		return domain.ErrAccountLocked
	}

	// Lockout expired, reset
	if !info.lockedUntil.IsZero() {
		delete(s.loginAttempts, key)
	}

	return nil
}

// recordLoginFailure increments the failure count and locks after maxLoginAttempts.
func (s *AuthService) recordLoginFailure(key string) {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()

	info, exists := s.loginAttempts[key]
	if !exists {
		info = &loginAttemptInfo{}
		s.loginAttempts[key] = info
	}

	info.count++
	if info.count >= maxLoginAttempts {
		info.lockedUntil = time.Now().Add(lockoutDuration)
	}
}

// resetLoginAttempts clears the failure count after a successful login.
func (s *AuthService) resetLoginAttempts(key string) {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	delete(s.loginAttempts, key)
}

// checkDBLockout queries the database for lockout status.
func (s *AuthService) checkDBLockout(ctx context.Context, dealerID uuid.UUID, email string) error {
	status, err := s.store.Queries.GetUserLockoutStatus(ctx, db.GetUserLockoutStatusParams{
		DealerID: dealerID,
		Email:    email,
	})
	if err != nil {
		// User not found is not a lockout error
		return nil
	}

	if status.LockedUntil.Valid && status.LockedUntil.Time.After(time.Now()) {
		return domain.ErrAccountLocked
	}
	return nil
}

// recordLoginFailureDB records a login failure in both DB and in-memory.
func (s *AuthService) recordLoginFailureDB(ctx context.Context, dealerID uuid.UUID, email, memKey string) {
	// Always update in-memory (fast path, catches rapid bursts)
	s.recordLoginFailure(memKey)

	// Also update DB (persistent, multi-instance)
	if s.store != nil {
		if _, err := s.store.Queries.IncrementLoginFailures(ctx, db.IncrementLoginFailuresParams{
			DealerID: dealerID,
			Email:    email,
		}); err != nil {
			slog.Warn("failed to record DB login failure", "error", err, "email", email)
		}
	}
}

// resetLoginAttemptsDB resets attempts in both DB and in-memory.
func (s *AuthService) resetLoginAttemptsDB(ctx context.Context, dealerID uuid.UUID, email, memKey string) {
	s.resetLoginAttempts(memKey)

	if s.store != nil {
		if err := s.store.Queries.ResetLoginFailures(ctx, db.ResetLoginFailuresParams{
			DealerID: dealerID,
			Email:    email,
		}); err != nil {
			slog.Warn("failed to reset DB login failures", "error", err, "email", email)
		}
	}
}
