package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/builderwire/lumber-now/backend/internal/domain"
)

// ---------------------------------------------------------------------------
// InputType.Valid() — table-driven
// ---------------------------------------------------------------------------

func TestInputTypeValid(t *testing.T) {
	tests := []struct {
		name  string
		input domain.InputType
		valid bool
	}{
		{"text is valid", domain.InputText, true},
		{"voice is valid", domain.InputVoice, true},
		{"image is valid", domain.InputImage, true},
		{"pdf is valid", domain.InputPDF, true},
		{"empty string is invalid", domain.InputType(""), false},
		{"unknown type csv", domain.InputType("csv"), false},
		{"unknown type video", domain.InputType("video"), false},
		{"unknown type audio", domain.InputType("audio"), false},
		{"uppercase TEXT", domain.InputType("TEXT"), false},
		{"mixed case Text", domain.InputType("Text"), false},
		{"whitespace prefix", domain.InputType(" text"), false},
		{"whitespace suffix", domain.InputType("text "), false},
		{"null-like string", domain.InputType("null"), false},
		{"sql injection attempt", domain.InputType("text'; DROP TABLE requests;--"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.Valid(); got != tt.valid {
				t.Errorf("InputType(%q).Valid() = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// InputType constant string values
// ---------------------------------------------------------------------------

func TestInputTypeConstantValues(t *testing.T) {
	tests := []struct {
		constant domain.InputType
		expected string
	}{
		{domain.InputText, "text"},
		{domain.InputVoice, "voice"},
		{domain.InputImage, "image"},
		{domain.InputPDF, "pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("InputType constant = %q, want %q", string(tt.constant), tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RequestStatus — verify all expected status values exist
// ---------------------------------------------------------------------------

func TestRequestStatusValues(t *testing.T) {
	statuses := []struct {
		name   string
		status domain.RequestStatus
		value  string
	}{
		{"pending", domain.StatusPending, "pending"},
		{"processing", domain.StatusProcessing, "processing"},
		{"parsed", domain.StatusParsed, "parsed"},
		{"confirmed", domain.StatusConfirmed, "confirmed"},
		{"sent", domain.StatusSent, "sent"},
		{"failed", domain.StatusFailed, "failed"},
	}

	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status == "" {
				t.Error("found empty status constant")
			}
			if string(tt.status) != tt.value {
				t.Errorf("RequestStatus %q = %q, want %q", tt.name, string(tt.status), tt.value)
			}
		})
	}
}

func TestRequestStatusCount(t *testing.T) {
	// There should be exactly 6 defined request statuses.
	statuses := []domain.RequestStatus{
		domain.StatusPending,
		domain.StatusProcessing,
		domain.StatusParsed,
		domain.StatusConfirmed,
		domain.StatusSent,
		domain.StatusFailed,
	}
	if len(statuses) != 6 {
		t.Errorf("expected 6 request statuses, got %d", len(statuses))
	}
}

// ---------------------------------------------------------------------------
// RequestStatus uniqueness
// ---------------------------------------------------------------------------

func TestRequestStatusUniqueness(t *testing.T) {
	statuses := []domain.RequestStatus{
		domain.StatusPending,
		domain.StatusProcessing,
		domain.StatusParsed,
		domain.StatusConfirmed,
		domain.StatusSent,
		domain.StatusFailed,
	}

	seen := make(map[domain.RequestStatus]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("duplicate status value: %q", s)
		}
		seen[s] = true
	}
}

// ---------------------------------------------------------------------------
// CreateRequestInput struct fields
// ---------------------------------------------------------------------------

func TestCreateRequestInput_ZeroValue(t *testing.T) {
	var input CreateRequestInput

	if input.DealerID != uuid.Nil {
		t.Error("expected zero-value DealerID to be uuid.Nil")
	}
	if input.ContractorID != uuid.Nil {
		t.Error("expected zero-value ContractorID to be uuid.Nil")
	}
	if input.InputType != "" {
		t.Error("expected zero-value InputType to be empty")
	}
	if input.RawText != "" {
		t.Error("expected zero-value RawText to be empty")
	}
	if input.MediaURL != "" {
		t.Error("expected zero-value MediaURL to be empty")
	}
}

func TestCreateRequestInput_PopulatedFields(t *testing.T) {
	dealerID := uuid.New()
	contractorID := uuid.New()

	input := CreateRequestInput{
		DealerID:     dealerID,
		ContractorID: contractorID,
		InputType:    domain.InputText,
		RawText:      "2x4 lumber, 100 pieces",
		MediaURL:     "",
	}

	if input.DealerID != dealerID {
		t.Errorf("DealerID mismatch: got %v, want %v", input.DealerID, dealerID)
	}
	if input.ContractorID != contractorID {
		t.Errorf("ContractorID mismatch: got %v, want %v", input.ContractorID, contractorID)
	}
	if input.InputType != domain.InputText {
		t.Errorf("InputType mismatch: got %q, want %q", input.InputType, domain.InputText)
	}
	if input.RawText != "2x4 lumber, 100 pieces" {
		t.Errorf("RawText mismatch: got %q", input.RawText)
	}
}

func TestCreateRequestInput_MediaURLForImage(t *testing.T) {
	input := CreateRequestInput{
		DealerID:     uuid.New(),
		ContractorID: uuid.New(),
		InputType:    domain.InputImage,
		RawText:      "",
		MediaURL:     "https://cdn.example.com/uploads/order-photo.jpg",
	}

	if input.InputType != domain.InputImage {
		t.Errorf("expected InputType=image, got %q", input.InputType)
	}
	if input.MediaURL == "" {
		t.Error("MediaURL should not be empty for image input")
	}
}

func TestCreateRequestInput_MediaURLForPDF(t *testing.T) {
	input := CreateRequestInput{
		DealerID:     uuid.New(),
		ContractorID: uuid.New(),
		InputType:    domain.InputPDF,
		RawText:      "",
		MediaURL:     "https://cdn.example.com/uploads/order.pdf",
	}

	if input.InputType != domain.InputPDF {
		t.Errorf("expected InputType=pdf, got %q", input.InputType)
	}
	if input.MediaURL == "" {
		t.Error("MediaURL should not be empty for PDF input")
	}
}

func TestCreateRequestInput_VoiceType(t *testing.T) {
	input := CreateRequestInput{
		DealerID:     uuid.New(),
		ContractorID: uuid.New(),
		InputType:    domain.InputVoice,
		MediaURL:     "https://cdn.example.com/uploads/voice-memo.m4a",
	}

	if !input.InputType.Valid() {
		t.Error("voice InputType should be valid")
	}
}

// ---------------------------------------------------------------------------
// CreateRequestInput — InputType validation coverage
// ---------------------------------------------------------------------------

func TestCreateRequestInput_InvalidInputTypeIsDetected(t *testing.T) {
	input := CreateRequestInput{
		DealerID:     uuid.New(),
		ContractorID: uuid.New(),
		InputType:    domain.InputType("spreadsheet"),
		RawText:      "some data",
	}

	if input.InputType.Valid() {
		t.Error("InputType 'spreadsheet' should not be valid")
	}
}

func TestCreateRequestInput_EmptyInputTypeIsInvalid(t *testing.T) {
	input := CreateRequestInput{
		DealerID:     uuid.New(),
		ContractorID: uuid.New(),
		InputType:    domain.InputType(""),
		RawText:      "some data",
	}

	if input.InputType.Valid() {
		t.Error("empty InputType should not be valid")
	}
}

// ---------------------------------------------------------------------------
// Domain error sentinels
// ---------------------------------------------------------------------------

func TestDomainErrorsExist(t *testing.T) {
	errors := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", domain.ErrNotFound},
		{"ErrForbidden", domain.ErrForbidden},
		{"ErrUnauthorized", domain.ErrUnauthorized},
		{"ErrConflict", domain.ErrConflict},
		{"ErrBadRequest", domain.ErrBadRequest},
		{"ErrInternal", domain.ErrInternal},
		{"ErrInvalidInput", domain.ErrInvalidInput},
		{"ErrInvalidRole", domain.ErrInvalidRole},
		{"ErrInvalidStatus", domain.ErrInvalidStatus},
		{"ErrTenantMissing", domain.ErrTenantMissing},
	}

	for _, tt := range errors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("domain error %s should not be nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("domain error %s should have a non-empty message", tt.name)
			}
		})
	}
}

func TestDomainErrorMessages(t *testing.T) {
	tests := []struct {
		err     error
		message string
	}{
		{domain.ErrNotFound, "resource not found"},
		{domain.ErrForbidden, "access forbidden"},
		{domain.ErrUnauthorized, "unauthorized"},
		{domain.ErrConflict, "resource already exists"},
		{domain.ErrBadRequest, "bad request"},
		{domain.ErrInternal, "internal server error"},
		{domain.ErrInvalidInput, "invalid input"},
		{domain.ErrInvalidRole, "invalid role"},
		{domain.ErrInvalidStatus, "invalid status transition"},
		{domain.ErrTenantMissing, "tenant ID missing"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			if tt.err.Error() != tt.message {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.message)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StructuredItem — JSON serialization
// ---------------------------------------------------------------------------

func TestStructuredItemJSON(t *testing.T) {
	item := domain.StructuredItem{
		SKU:        "LBR-2X4-8",
		Name:       "2x4 Lumber 8ft",
		Quantity:   100,
		Unit:       "pieces",
		Confidence: 0.95,
		Matched:    true,
		Notes:      "standard grade",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal StructuredItem: %v", err)
	}

	var decoded domain.StructuredItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal StructuredItem: %v", err)
	}

	if decoded.SKU != item.SKU {
		t.Errorf("SKU mismatch: got %q, want %q", decoded.SKU, item.SKU)
	}
	if decoded.Name != item.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, item.Name)
	}
	if decoded.Quantity != item.Quantity {
		t.Errorf("Quantity mismatch: got %v, want %v", decoded.Quantity, item.Quantity)
	}
	if decoded.Unit != item.Unit {
		t.Errorf("Unit mismatch: got %q, want %q", decoded.Unit, item.Unit)
	}
	if decoded.Confidence != item.Confidence {
		t.Errorf("Confidence mismatch: got %v, want %v", decoded.Confidence, item.Confidence)
	}
	if decoded.Matched != item.Matched {
		t.Errorf("Matched mismatch: got %v, want %v", decoded.Matched, item.Matched)
	}
	if decoded.Notes != item.Notes {
		t.Errorf("Notes mismatch: got %q, want %q", decoded.Notes, item.Notes)
	}
}

func TestStructuredItemJSON_OmitsEmptyNotes(t *testing.T) {
	item := domain.StructuredItem{
		SKU:        "PLY-3/4",
		Name:       "3/4 Plywood",
		Quantity:   20,
		Unit:       "sheets",
		Confidence: 0.88,
		Matched:    false,
		Notes:      "", // Should be omitted due to omitempty tag.
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal StructuredItem: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if _, exists := raw["notes"]; exists {
		t.Error("expected 'notes' field to be omitted when empty")
	}
}

func TestStructuredItemJSON_IncludesNonEmptyNotes(t *testing.T) {
	item := domain.StructuredItem{
		SKU:        "PLY-3/4",
		Name:       "3/4 Plywood",
		Quantity:   20,
		Unit:       "sheets",
		Confidence: 0.88,
		Matched:    false,
		Notes:      "special order",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal StructuredItem: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if _, exists := raw["notes"]; !exists {
		t.Error("expected 'notes' field to be present when non-empty")
	}
}

func TestStructuredItemSliceJSON(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "A", Name: "Item A", Quantity: 1, Unit: "ea", Confidence: 0.9, Matched: true},
		{SKU: "B", Name: "Item B", Quantity: 2, Unit: "ft", Confidence: 0.8, Matched: false},
		{SKU: "C", Name: "Item C", Quantity: 3, Unit: "lb", Confidence: 0.7, Matched: true},
	}

	data, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("failed to marshal items slice: %v", err)
	}

	var decoded []domain.StructuredItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal items slice: %v", err)
	}

	if len(decoded) != 3 {
		t.Fatalf("expected 3 items, got %d", len(decoded))
	}

	for i, item := range decoded {
		if item.SKU != items[i].SKU {
			t.Errorf("item[%d] SKU mismatch: got %q, want %q", i, item.SKU, items[i].SKU)
		}
	}
}

// ---------------------------------------------------------------------------
// Role.Valid() — cross-package verification from service layer
// ---------------------------------------------------------------------------

func TestRoleValid(t *testing.T) {
	tests := []struct {
		name  string
		role  domain.Role
		valid bool
	}{
		{"platform_admin", domain.RolePlatformAdmin, true},
		{"dealer_admin", domain.RoleDealerAdmin, true},
		{"sales_rep", domain.RoleSalesRep, true},
		{"contractor", domain.RoleContractor, true},
		{"empty", domain.Role(""), false},
		{"unknown", domain.Role("manager"), false},
		{"uppercase", domain.Role("CONTRACTOR"), false},
		{"with space", domain.Role("sales rep"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.Valid(); got != tt.valid {
				t.Errorf("Role(%q).Valid() = %v, want %v", tt.role, got, tt.valid)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// NewRequestService constructor
// ---------------------------------------------------------------------------

func TestNewRequestService_NilDependencies(t *testing.T) {
	svc := NewRequestService(nil, nil, nil, nil, nil)
	if svc == nil {
		t.Fatal("NewRequestService should not return nil even with nil dependencies")
	}
}

func TestNewRequestService_StoresFields(t *testing.T) {
	svc := NewRequestService(nil, nil, nil, nil, nil)
	if svc.store != nil {
		t.Error("expected store to be nil")
	}
	if svc.aiClient != nil {
		t.Error("expected aiClient to be nil")
	}
	if svc.transcriber != nil {
		t.Error("expected transcriber to be nil")
	}
	if svc.emailClient != nil {
		t.Error("expected emailClient to be nil")
	}
	if svc.mediaSvc != nil {
		t.Error("expected mediaSvc to be nil")
	}
}

// ---------------------------------------------------------------------------
// StructuredItem zero values
// ---------------------------------------------------------------------------

func TestStructuredItemZeroValue(t *testing.T) {
	var item domain.StructuredItem

	if item.SKU != "" {
		t.Error("zero-value SKU should be empty")
	}
	if item.Name != "" {
		t.Error("zero-value Name should be empty")
	}
	if item.Quantity != 0 {
		t.Error("zero-value Quantity should be 0")
	}
	if item.Unit != "" {
		t.Error("zero-value Unit should be empty")
	}
	if item.Confidence != 0 {
		t.Error("zero-value Confidence should be 0")
	}
	if item.Matched != false {
		t.Error("zero-value Matched should be false")
	}
	if item.Notes != "" {
		t.Error("zero-value Notes should be empty")
	}
}

// ---------------------------------------------------------------------------
// JSON tag verification via round-trip
// ---------------------------------------------------------------------------

func TestStructuredItemJSONTagNames(t *testing.T) {
	item := domain.StructuredItem{
		SKU:        "TEST-SKU",
		Name:       "Test Name",
		Quantity:   5.5,
		Unit:       "boards",
		Confidence: 0.99,
		Matched:    true,
		Notes:      "rush order",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	expectedKeys := []string{"sku", "name", "quantity", "unit", "confidence", "matched", "notes"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q to be present", key)
		}
	}
}

// ---------------------------------------------------------------------------
// wrapVersionErr — maps pgx.ErrNoRows to domain.ErrVersionConflict
// ---------------------------------------------------------------------------

func TestWrapVersionErr_NoRows(t *testing.T) {
	err := wrapVersionErr(pgx.ErrNoRows)
	if !errors.Is(err, domain.ErrVersionConflict) {
		t.Errorf("wrapVersionErr(pgx.ErrNoRows) = %v, want domain.ErrVersionConflict", err)
	}
}

func TestWrapVersionErr_OtherError(t *testing.T) {
	original := fmt.Errorf("some database error")
	err := wrapVersionErr(original)
	if err != original {
		t.Errorf("wrapVersionErr(other) = %v, want original error %v", err, original)
	}
}

func TestWrapVersionErr_NilError(t *testing.T) {
	err := wrapVersionErr(nil)
	if err != nil {
		t.Errorf("wrapVersionErr(nil) = %v, want nil", err)
	}
}

func TestWrapVersionErr_WrappedNoRows(t *testing.T) {
	wrapped := fmt.Errorf("query failed: %w", pgx.ErrNoRows)
	err := wrapVersionErr(wrapped)
	if !errors.Is(err, domain.ErrVersionConflict) {
		t.Errorf("wrapVersionErr(wrapped pgx.ErrNoRows) = %v, want domain.ErrVersionConflict", err)
	}
}
