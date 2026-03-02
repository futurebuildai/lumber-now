package domain

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrForbidden     = errors.New("access forbidden")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrConflict      = errors.New("resource already exists")
	ErrBadRequest    = errors.New("bad request")
	ErrInternal      = errors.New("internal server error")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInvalidRole   = errors.New("invalid role")
	ErrInvalidStatus = errors.New("invalid status transition")
	ErrTenantMissing   = errors.New("tenant ID missing")
	ErrAccountLocked   = errors.New("account temporarily locked")
	ErrVersionConflict = errors.New("version conflict: resource was modified")
)
