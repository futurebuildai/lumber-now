package domain

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

const (
	LocalsDealerID = "dealer_id"
	LocalsClaims   = "claims"
	LocalsRequestID = "request_id"
)

func DealerIDFromLocals(val interface{}) (uuid.UUID, error) {
	if val == nil {
		return uuid.Nil, ErrTenantMissing
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrTenantMissing
	}
	return id, nil
}

func ClaimsFromLocals(val interface{}) (*JWTClaims, error) {
	if val == nil {
		return nil, ErrUnauthorized
	}
	claims, ok := val.(*JWTClaims)
	if !ok {
		return nil, ErrUnauthorized
	}
	return claims, nil
}
