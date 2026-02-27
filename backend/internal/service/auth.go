package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

type AuthService struct {
	store     *store.Store
	jwtSecret []byte
}

func NewAuthService(s *store.Store, jwtSecret string) *AuthService {
	return &AuthService{
		store:     s,
		jwtSecret: []byte(jwtSecret),
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

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
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
	user, err := s.store.Queries.GetUserByEmail(ctx, db.GetUserByEmailParams{
		DealerID: input.DealerID,
		Email:    input.Email,
	})
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if !user.Active {
		return nil, domain.ErrForbidden
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.issueTokenPair(user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.validateToken(refreshToken, "refresh")
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

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
	expiresAt := time.Now().Add(24 * time.Hour)
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
	claims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"dealer_id": user.DealerID.String(),
		"email":     user.Email,
		"role":      string(user.Role),
		"type":      tokenType,
		"exp":       expiresAt.Unix(),
		"iat":       time.Now().Unix(),
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

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	dealerID, err := uuid.Parse(claims["dealer_id"].(string))
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return &domain.JWTClaims{
		UserID:   userID,
		DealerID: dealerID,
		Email:    claims["email"].(string),
		Role:     domain.Role(claims["role"].(string)),
	}, nil
}
