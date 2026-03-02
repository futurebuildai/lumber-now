package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestInputType_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input InputType
		want  bool
	}{
		{
			name:  "valid text type",
			input: InputText,
			want:  true,
		},
		{
			name:  "valid voice type",
			input: InputVoice,
			want:  true,
		},
		{
			name:  "valid image type",
			input: InputImage,
			want:  true,
		},
		{
			name:  "valid PDF type",
			input: InputPDF,
			want:  true,
		},
		{
			name:  "invalid empty string",
			input: InputType(""),
			want:  false,
		},
		{
			name:  "invalid random string",
			input: InputType("invalid"),
			want:  false,
		},
		{
			name:  "invalid video type",
			input: InputType("video"),
			want:  false,
		},
		{
			name:  "invalid audio type",
			input: InputType("audio"),
			want:  false,
		},
		{
			name:  "case sensitivity - uppercase TEXT",
			input: InputType("TEXT"),
			want:  false,
		},
		{
			name:  "case sensitivity - mixed case Voice",
			input: InputType("Voice"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.Valid(); got != tt.want {
				t.Errorf("InputType.Valid() = %v, want %v for input %q", got, tt.want, tt.input)
			}
		})
	}
}

func TestRole_Valid(t *testing.T) {
	tests := []struct {
		name  string
		role  Role
		want  bool
	}{
		{
			name: "valid platform admin role",
			role: RolePlatformAdmin,
			want: true,
		},
		{
			name: "valid dealer admin role",
			role: RoleDealerAdmin,
			want: true,
		},
		{
			name: "valid sales rep role",
			role: RoleSalesRep,
			want: true,
		},
		{
			name: "valid contractor role",
			role: RoleContractor,
			want: true,
		},
		{
			name: "invalid empty string",
			role: Role(""),
			want: false,
		},
		{
			name: "invalid random role",
			role: Role("invalid_role"),
			want: false,
		},
		{
			name: "invalid admin role",
			role: Role("admin"),
			want: false,
		},
		{
			name: "invalid user role",
			role: Role("user"),
			want: false,
		},
		{
			name: "case sensitivity - uppercase PLATFORM_ADMIN",
			role: Role("PLATFORM_ADMIN"),
			want: false,
		},
		{
			name: "case sensitivity - mixed case Contractor",
			role: Role("Contractor"),
			want: false,
		},
		{
			name: "partial match - platform",
			role: Role("platform"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.Valid(); got != tt.want {
				t.Errorf("Role.Valid() = %v, want %v for role %q", got, tt.want, tt.role)
			}
		})
	}
}

func TestClaimsFromLocals(t *testing.T) {
	validClaims := &JWTClaims{
		UserID:   uuid.New(),
		DealerID: uuid.New(),
		Email:    "test@example.com",
		Role:     RoleContractor,
	}

	tests := []struct {
		name    string
		input   interface{}
		want    *JWTClaims
		wantErr error
	}{
		{
			name:    "valid claims",
			input:   validClaims,
			want:    validClaims,
			wantErr: nil,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantErr: ErrUnauthorized,
		},
		{
			name:    "wrong type - string",
			input:   "not a claims object",
			want:    nil,
			wantErr: ErrUnauthorized,
		},
		{
			name:    "wrong type - int",
			input:   42,
			want:    nil,
			wantErr: ErrUnauthorized,
		},
		{
			name:    "wrong type - map",
			input:   map[string]string{"key": "value"},
			want:    nil,
			wantErr: ErrUnauthorized,
		},
		{
			name:    "wrong type - struct instead of pointer",
			input:   JWTClaims{UserID: uuid.New(), DealerID: uuid.New()},
			want:    nil,
			wantErr: ErrUnauthorized,
		},
		{
			name: "valid claims with all fields populated",
			input: &JWTClaims{
				UserID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				DealerID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				Email:    "admin@example.com",
				Role:     RolePlatformAdmin,
			},
			want: &JWTClaims{
				UserID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				DealerID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				Email:    "admin@example.com",
				Role:     RolePlatformAdmin,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ClaimsFromLocals(tt.input)

			if err != tt.wantErr {
				t.Errorf("ClaimsFromLocals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("ClaimsFromLocals() should return nil when error occurs, got %v", got)
				}
				return
			}

			if got == nil && tt.want != nil {
				t.Errorf("ClaimsFromLocals() = nil, want %v", tt.want)
				return
			}

			if got != nil && tt.want != nil {
				if got.UserID != tt.want.UserID {
					t.Errorf("ClaimsFromLocals() UserID = %v, want %v", got.UserID, tt.want.UserID)
				}
				if got.DealerID != tt.want.DealerID {
					t.Errorf("ClaimsFromLocals() DealerID = %v, want %v", got.DealerID, tt.want.DealerID)
				}
				if got.Email != tt.want.Email {
					t.Errorf("ClaimsFromLocals() Email = %v, want %v", got.Email, tt.want.Email)
				}
				if got.Role != tt.want.Role {
					t.Errorf("ClaimsFromLocals() Role = %v, want %v", got.Role, tt.want.Role)
				}
			}
		})
	}
}

func TestDealerIDFromLocals(t *testing.T) {
	validDealerID := uuid.New()

	tests := []struct {
		name    string
		input   interface{}
		want    uuid.UUID
		wantErr error
	}{
		{
			name:    "valid dealer ID",
			input:   validDealerID,
			want:    validDealerID,
			wantErr: nil,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    uuid.Nil,
			wantErr: ErrTenantMissing,
		},
		{
			name:    "wrong type - string",
			input:   "not a uuid",
			want:    uuid.Nil,
			wantErr: ErrTenantMissing,
		},
		{
			name:    "wrong type - int",
			input:   123,
			want:    uuid.Nil,
			wantErr: ErrTenantMissing,
		},
		{
			name:    "wrong type - string UUID",
			input:   "123e4567-e89b-12d3-a456-426614174000",
			want:    uuid.Nil,
			wantErr: ErrTenantMissing,
		},
		{
			name:    "wrong type - byte slice",
			input:   []byte{1, 2, 3, 4},
			want:    uuid.Nil,
			wantErr: ErrTenantMissing,
		},
		{
			name:    "valid specific UUID",
			input:   uuid.MustParse("323e4567-e89b-12d3-a456-426614174000"),
			want:    uuid.MustParse("323e4567-e89b-12d3-a456-426614174000"),
			wantErr: nil,
		},
		{
			name:    "wrong type - pointer to UUID",
			input:   &validDealerID,
			want:    uuid.Nil,
			wantErr: ErrTenantMissing,
		},
		{
			name:    "wrong type - map",
			input:   map[string]interface{}{"id": validDealerID},
			want:    uuid.Nil,
			wantErr: ErrTenantMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DealerIDFromLocals(tt.input)

			if err != tt.wantErr {
				t.Errorf("DealerIDFromLocals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("DealerIDFromLocals() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRoleConstants ensures role constants match expected values
func TestRoleConstants(t *testing.T) {
	tests := []struct {
		role     Role
		expected string
	}{
		{RolePlatformAdmin, "platform_admin"},
		{RoleDealerAdmin, "dealer_admin"},
		{RoleSalesRep, "sales_rep"},
		{RoleContractor, "contractor"},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("Role constant %v = %q, want %q", tt.role, string(tt.role), tt.expected)
			}
		})
	}
}

// TestInputTypeConstants ensures input type constants match expected values
func TestInputTypeConstants(t *testing.T) {
	tests := []struct {
		inputType InputType
		expected  string
	}{
		{InputText, "text"},
		{InputVoice, "voice"},
		{InputImage, "image"},
		{InputPDF, "pdf"},
	}

	for _, tt := range tests {
		t.Run(string(tt.inputType), func(t *testing.T) {
			if string(tt.inputType) != tt.expected {
				t.Errorf("InputType constant %v = %q, want %q", tt.inputType, string(tt.inputType), tt.expected)
			}
		})
	}
}

// TestRequestStatusConstants ensures status constants exist and have expected values
func TestRequestStatusConstants(t *testing.T) {
	tests := []struct {
		status   RequestStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusProcessing, "processing"},
		{StatusParsed, "parsed"},
		{StatusConfirmed, "confirmed"},
		{StatusSent, "sent"},
		{StatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("RequestStatus constant %v = %q, want %q", tt.status, string(tt.status), tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error variable existence checks
// ---------------------------------------------------------------------------

func TestErrorVariables_NotNil(t *testing.T) {
	errors := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrForbidden", ErrForbidden},
		{"ErrUnauthorized", ErrUnauthorized},
		{"ErrConflict", ErrConflict},
		{"ErrBadRequest", ErrBadRequest},
		{"ErrInternal", ErrInternal},
		{"ErrInvalidInput", ErrInvalidInput},
		{"ErrInvalidRole", ErrInvalidRole},
		{"ErrInvalidStatus", ErrInvalidStatus},
		{"ErrTenantMissing", ErrTenantMissing},
		{"ErrAccountLocked", ErrAccountLocked},
		{"ErrVersionConflict", ErrVersionConflict},
	}

	for _, tt := range errors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
		})
	}
}

func TestErrorVariables_HaveMessages(t *testing.T) {
	errors := []struct {
		name     string
		err      error
		contains string
	}{
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrForbidden", ErrForbidden, "forbidden"},
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
		{"ErrConflict", ErrConflict, "already exists"},
		{"ErrBadRequest", ErrBadRequest, "bad request"},
		{"ErrInternal", ErrInternal, "internal"},
		{"ErrInvalidInput", ErrInvalidInput, "invalid input"},
		{"ErrInvalidRole", ErrInvalidRole, "invalid role"},
		{"ErrInvalidStatus", ErrInvalidStatus, "invalid status"},
		{"ErrTenantMissing", ErrTenantMissing, "tenant"},
		{"ErrAccountLocked", ErrAccountLocked, "locked"},
		{"ErrVersionConflict", ErrVersionConflict, "version conflict"},
	}

	for _, tt := range errors {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Errorf("%s.Error() should not be empty", tt.name)
			}
			found := false
			for i := 0; i <= len(msg)-len(tt.contains); i++ {
				if msg[i:i+len(tt.contains)] == tt.contains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s.Error() = %q, expected to contain %q", tt.name, msg, tt.contains)
			}
		})
	}
}

func TestErrorVariables_AreDistinct(t *testing.T) {
	allErrors := []error{
		ErrNotFound,
		ErrForbidden,
		ErrUnauthorized,
		ErrConflict,
		ErrBadRequest,
		ErrInternal,
		ErrInvalidInput,
		ErrInvalidRole,
		ErrInvalidStatus,
		ErrTenantMissing,
		ErrAccountLocked,
		ErrVersionConflict,
	}

	for i := 0; i < len(allErrors); i++ {
		for j := i + 1; j < len(allErrors); j++ {
			if allErrors[i] == allErrors[j] {
				t.Errorf("error %d and %d are the same: %v", i, j, allErrors[i])
			}
		}
	}
}
