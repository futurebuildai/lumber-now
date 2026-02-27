package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RolePlatformAdmin Role = "platform_admin"
	RoleDealerAdmin   Role = "dealer_admin"
	RoleSalesRep      Role = "sales_rep"
	RoleContractor    Role = "contractor"
)

func (r Role) Valid() bool {
	switch r {
	case RolePlatformAdmin, RoleDealerAdmin, RoleSalesRep, RoleContractor:
		return true
	}
	return false
}

type RequestStatus string

const (
	StatusPending    RequestStatus = "pending"
	StatusProcessing RequestStatus = "processing"
	StatusParsed     RequestStatus = "parsed"
	StatusConfirmed  RequestStatus = "confirmed"
	StatusSent       RequestStatus = "sent"
	StatusFailed     RequestStatus = "failed"
)

type InputType string

const (
	InputText  InputType = "text"
	InputVoice InputType = "voice"
	InputImage InputType = "image"
	InputPDF   InputType = "pdf"
)

func (t InputType) Valid() bool {
	switch t {
	case InputText, InputVoice, InputImage, InputPDF:
		return true
	}
	return false
}

type Dealer struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Subdomain      string    `json:"subdomain"`
	LogoURL        string    `json:"logo_url"`
	PrimaryColor   string    `json:"primary_color"`
	SecondaryColor string    `json:"secondary_color"`
	ContactEmail   string    `json:"contact_email"`
	ContactPhone   string    `json:"contact_phone"`
	Address        string    `json:"address"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type User struct {
	ID            uuid.UUID  `json:"id"`
	DealerID      uuid.UUID  `json:"dealer_id"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"`
	FullName      string     `json:"full_name"`
	Phone         string     `json:"phone"`
	Role          Role       `json:"role"`
	AssignedRepID *uuid.UUID `json:"assigned_rep_id,omitempty"`
	Active        bool       `json:"active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type InventoryItem struct {
	ID          uuid.UUID       `json:"id"`
	DealerID    uuid.UUID       `json:"dealer_id"`
	SKU         string          `json:"sku"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Unit        string          `json:"unit"`
	Price       string          `json:"price"`
	InStock     bool            `json:"in_stock"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Request struct {
	ID              uuid.UUID       `json:"id"`
	DealerID        uuid.UUID       `json:"dealer_id"`
	ContractorID    uuid.UUID       `json:"contractor_id"`
	AssignedRepID   *uuid.UUID      `json:"assigned_rep_id,omitempty"`
	Status          RequestStatus   `json:"status"`
	InputType       InputType       `json:"input_type"`
	RawText         string          `json:"raw_text"`
	MediaURL        string          `json:"media_url"`
	StructuredItems json.RawMessage `json:"structured_items"`
	AIConfidence    string          `json:"ai_confidence"`
	Notes           string          `json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type StructuredItem struct {
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit"`
	Confidence float64 `json:"confidence"`
	Matched    bool    `json:"matched"`
	Notes      string  `json:"notes,omitempty"`
}

type TenantConfig struct {
	DealerID       uuid.UUID `json:"dealer_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	LogoURL        string    `json:"logo_url"`
	PrimaryColor   string    `json:"primary_color"`
	SecondaryColor string    `json:"secondary_color"`
	ContactEmail   string    `json:"contact_email"`
	ContactPhone   string    `json:"contact_phone"`
}

type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	DealerID uuid.UUID `json:"dealer_id"`
	Email    string    `json:"email"`
	Role     Role      `json:"role"`
}
