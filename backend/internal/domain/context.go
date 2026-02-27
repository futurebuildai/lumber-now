package domain

import "github.com/google/uuid"

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
