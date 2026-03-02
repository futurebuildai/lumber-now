package handler

import (
	"testing"
)

func TestValidateMediaURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// --- Valid external URLs (use IP-based URLs to avoid DNS resolution
	// failures in test environments) ---
		{
			name:    "valid https URL with public IP",
			url:     "https://8.8.8.8/image.png",
			wantErr: false,
		},
		{
			name:    "valid http URL with public IP",
			url:     "http://1.1.1.1/photo.jpg",
			wantErr: false,
		},
		{
			name:    "valid URL with port",
			url:     "https://203.0.113.50:8443/resource",
			wantErr: false,
		},
		{
			name:    "valid URL with query string",
			url:     "https://198.51.100.1/file?token=abc123",
			wantErr: false,
		},
		{
			name:    "valid URL with path segments",
			url:     "https://203.0.113.10/a/b/c/d/image.webp",
			wantErr: false,
		},
		{
			name:    "empty scheme treated as relative (allowed with public IP)",
			url:     "//8.8.4.4/image.png",
			wantErr: false,
		},

		// --- Blocked schemes (SSRF) ---
		{
			name:    "file:// scheme blocked",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "ftp:// scheme blocked",
			url:     "ftp://evil.com/payload",
			wantErr: true,
		},
		{
			name:    "gopher:// scheme blocked",
			url:     "gopher://evil.com:70/_payload",
			wantErr: true,
		},
		{
			name:    "data: scheme blocked",
			url:     "data:text/html,<h1>hello</h1>",
			wantErr: true,
		},
		{
			name:    "javascript: scheme blocked",
			url:     "javascript:alert(1)",
			wantErr: true,
		},

		// --- Blocked hosts (SSRF - internal/private) ---
		{
			name:    "localhost blocked",
			url:     "http://localhost/admin",
			wantErr: true,
		},
		{
			name:    "127.0.0.1 blocked",
			url:     "http://127.0.0.1/internal",
			wantErr: true,
		},
		{
			name:    "0.0.0.0 blocked",
			url:     "http://0.0.0.0/",
			wantErr: true,
		},
		{
			name:    "AWS metadata endpoint blocked",
			url:     "http://169.254.169.254/latest/meta-data/",
			wantErr: true,
		},
		{
			name:    "GCP metadata endpoint blocked",
			url:     "http://metadata.google.internal/computeMetadata/v1/",
			wantErr: true,
		},
		{
			name:    "IPv6 loopback blocked",
			url:     "http://[::1]/",
			wantErr: true,
		},
		{
			name:    "IPv4-mapped IPv6 loopback blocked",
			url:     "http://[::ffff:127.0.0.1]/",
			wantErr: true,
		},
		{
			name:    "IPv4-mapped IPv6 metadata blocked",
			url:     "http://[::ffff:169.254.169.254]/",
			wantErr: true,
		},
		{
			name:    "IPv4-mapped IPv6 private blocked",
			url:     "http://[::ffff:10.0.0.1]/",
			wantErr: true,
		},

		// --- Blocked private IP ranges ---
		{
			name:    "10.x.x.x range blocked",
			url:     "http://10.0.0.1/internal",
			wantErr: true,
		},
		{
			name:    "10.255.255.255 blocked",
			url:     "http://10.255.255.255/secret",
			wantErr: true,
		},
		{
			name:    "192.168.x.x range blocked",
			url:     "http://192.168.1.1/admin",
			wantErr: true,
		},
		{
			name:    "192.168.0.100 blocked",
			url:     "http://192.168.0.100/",
			wantErr: true,
		},
		{
			name:    "172.16.x.x blocked",
			url:     "http://172.16.0.1/internal",
			wantErr: true,
		},
		{
			name:    "172.31.255.255 blocked",
			url:     "http://172.31.255.255/",
			wantErr: true,
		},

		// --- Edge cases ---
		{
			name:    "localhost with HTTPS",
			url:     "https://localhost/secret",
			wantErr: true,
		},
		{
			name:    "localhost with port",
			url:     "http://localhost:8080/admin",
			wantErr: true,
		},
		{
			name:    "127.0.0.1 with port",
			url:     "http://127.0.0.1:3000/api",
			wantErr: true,
		},
		{
			name:    "metadata with HTTPS",
			url:     "https://169.254.169.254/latest/",
			wantErr: true,
		},
		{
			name:    "case insensitive localhost",
			url:     "http://LOCALHOST/admin",
			wantErr: true,
		},
		{
			name:    "case insensitive metadata",
			url:     "http://METADATA.GOOGLE.INTERNAL/",
			wantErr: true,
		},
		{
			name:    "mixed case scheme with public IP",
			url:     "HTTP://8.8.8.8/image.png",
			wantErr: false,
		},
		{
			name:    "FTP mixed case blocked",
			url:     "FTP://evil.com/payload",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMediaURL(tt.url)
			if tt.wantErr && err == nil {
				t.Errorf("validateMediaURL(%q) = nil, want error", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateMediaURL(%q) = %v, want nil", tt.url, err)
			}
		})
	}
}

// TestValidateMediaURL_PrivateRangePrefix verifies that the 172.x check
// covers the entire 172.16.0.0/12 range using prefix matching on "172.".
func TestValidateMediaURL_PrivateRangePrefix(t *testing.T) {
	// The implementation blocks any host starting with "172." which is broader
	// than the actual private range (172.16-31.x.x). These should all be blocked.
	blockedIPs := []string{
		"172.0.0.1",
		"172.16.0.1",
		"172.20.10.5",
		"172.31.255.255",
	}

	for _, ip := range blockedIPs {
		url := "http://" + ip + "/path"
		if err := validateMediaURL(url); err == nil {
			t.Errorf("expected %q to be blocked, but it was allowed", ip)
		}
	}
}

// TestValidateMediaURL_EmptyString verifies that an empty URL string is
// accepted (the caller in Create checks for empty before calling validate).
func TestValidateMediaURL_EmptyString(t *testing.T) {
	err := validateMediaURL("")
	if err != nil {
		t.Errorf("validateMediaURL(\"\") = %v, expected nil (caller handles empty check)", err)
	}
}

// TestValidateMediaURL_PublicIPsAllowed ensures that non-private public IPs
// are not blocked.
func TestValidateMediaURL_PublicIPsAllowed(t *testing.T) {
	publicURLs := []string{
		"http://8.8.8.8/image.png",
		"http://1.1.1.1/file.pdf",
		"https://203.0.113.50/upload.jpg",
	}

	for _, u := range publicURLs {
		if err := validateMediaURL(u); err != nil {
			t.Errorf("validateMediaURL(%q) = %v, expected nil for public IP", u, err)
		}
	}
}
