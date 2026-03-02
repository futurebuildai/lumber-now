package handler

import (
	"testing"
)

// Test clampLimit with table-driven tests
func TestClampLimit(t *testing.T) {
	tests := []struct {
		name string
		in   int32
		want int32
	}{
		{"normal value", 50, 50},
		{"minimum boundary", 1, 1},
		{"maximum boundary", 100, 100},
		{"below minimum", 0, 1},
		{"negative value", -10, 1},
		{"above maximum", 150, 100},
		{"way above maximum", 1000, 100},
		{"int32 max", 2147483647, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampLimit(tt.in); got != tt.want {
				t.Errorf("clampLimit(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

// Test clampOffset with table-driven tests
func TestClampOffset(t *testing.T) {
	tests := []struct {
		name string
		in   int32
		want int32
	}{
		{"zero", 0, 0},
		{"positive", 50, 50},
		{"negative clamped to zero", -1, 0},
		{"large negative", -1000, 0},
		{"large positive", 99999, 99999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampOffset(tt.in); got != tt.want {
				t.Errorf("clampOffset(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

// TestClampLimitAndOffsetCombined verifies that a realistic page request
// produces clamped values that make sense for pagination.
func TestClampLimitAndOffsetCombined(t *testing.T) {
	// Simulate a request with out-of-range pagination values
	rawLimit := int32(-50)
	rawOffset := int32(-25)

	limit := clampLimit(rawLimit)
	offset := clampOffset(rawOffset)

	if limit < 1 {
		t.Errorf("clamped limit should be >= 1, got %d", limit)
	}
	if offset < 0 {
		t.Errorf("clamped offset should be >= 0, got %d", offset)
	}
}

// TestNewAdminHandler verifies that the constructor stores the provided store.
func TestNewAdminHandler(t *testing.T) {
	h := NewAdminHandler(nil)
	if h == nil {
		t.Fatal("NewAdminHandler(nil) returned nil")
	}
	if h.store != nil {
		t.Error("expected store field to be nil when constructed with nil")
	}
}
