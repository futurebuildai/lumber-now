package handler

import (
	"testing"
)

func TestAllowedContentTypes(t *testing.T) {
	expected := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"application/pdf",
		"audio/mpeg",
		"audio/wav",
		"audio/ogg",
		"audio/webm",
		"text/csv",
	}
	for _, ct := range expected {
		if !allowedContentTypes[ct] {
			t.Errorf("expected %q to be allowed", ct)
		}
	}
}

func TestAllowedContentTypesCount(t *testing.T) {
	// Ensure no unexpected content types were added without test coverage
	if len(allowedContentTypes) != 9 {
		t.Errorf("expected 9 allowed content types, got %d", len(allowedContentTypes))
	}
}

func TestDisallowedContentTypes(t *testing.T) {
	disallowed := []string{
		"",
		"text/html",
		"application/javascript",
		"text/xml",
		"application/zip",
		"video/mp4",
		"application/x-sh",
		"application/x-executable",
		"text/plain",
	}
	for _, ct := range disallowed {
		if allowedContentTypes[ct] {
			t.Errorf("expected %q to be disallowed", ct)
		}
	}
}

func TestMaxUploadSize(t *testing.T) {
	if maxUploadSize != 10*1024*1024 {
		t.Errorf("maxUploadSize = %d, want %d", maxUploadSize, 10*1024*1024)
	}
}

func TestMaxUploadSizeInMB(t *testing.T) {
	mb := maxUploadSize / (1024 * 1024)
	if mb != 10 {
		t.Errorf("maxUploadSize in MB = %d, want 10", mb)
	}
}

func TestMaxUploadSizeIsPositive(t *testing.T) {
	if maxUploadSize <= 0 {
		t.Errorf("maxUploadSize should be positive, got %d", maxUploadSize)
	}
}

func TestNewMediaHandler(t *testing.T) {
	h := NewMediaHandler(nil)
	if h == nil {
		t.Fatal("NewMediaHandler(nil) returned nil")
	}
	if h.mediaSvc != nil {
		t.Error("expected mediaSvc field to be nil when constructed with nil")
	}
}
