package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// Test data constants
const (
	testUsername = "testuser"
	testPassword = "testpassword123"
)

// TestGenerateToken_ReturnsValidBase64 verifies token is valid URL-safe base64
func TestGenerateToken_ReturnsValidBase64(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken() returned error: %v", err)
	}

	// Base64 encoding of 32 bytes should be 43 characters (without padding)
	// or 44 characters (with padding)
	if len(token) < 43 {
		t.Errorf("token length = %d, want >= 43", len(token))
	}

	// Verify it's valid base64 by checking for valid characters
	// URL-safe base64 uses: A-Z, a-z, 0-9, -, _
	for _, c := range token {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '=') {
			t.Errorf("token contains invalid base64 character: %c", c)
		}
	}
}

// TestGenerateToken_ReturnsUniqueTokens verifies tokens are random
func TestGenerateToken_ReturnsUniqueTokens(t *testing.T) {
	token1, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken() first call returned error: %v", err)
	}

	token2, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken() second call returned error: %v", err)
	}

	if token1 == token2 {
		t.Errorf("generateToken() returned identical tokens, expected unique values")
	}
}

// TestGetClientIP verifies IP extraction logic
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:          "uses X-Forwarded-For when present",
			xForwardedFor: "192.168.1.1",
			remoteAddr:    "10.0.0.1:12345",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "falls back to RemoteAddr when no header",
			xForwardedFor: "",
			remoteAddr:    "10.0.0.1:12345",
			expectedIP:    "10.0.0.1:12345",
		},
		{
			name:          "uses X-Forwarded-For even if empty string",
			xForwardedFor: "",
			remoteAddr:    "10.0.0.1:12345",
			expectedIP:    "10.0.0.1:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			req.RemoteAddr = tt.remoteAddr

			got := getClientIP(req)
			if got != tt.expectedIP {
				t.Errorf("getClientIP() = %v, want %v", got, tt.expectedIP)
			}
		})
	}
}

// TestGetCookieSecure verifies cookie security settings
func TestGetCookieSecure(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "defaults to true when not set",
			envValue: "",
			want:     true,
		},
		{
			name:     "returns false when explicitly set to false",
			envValue: "false",
			want:     false,
		},
		{
			name:     "returns false for FALSE (case insensitive)",
			envValue: "FALSE",
			want:     false,
		},
		{
			name:     "returns true when set to true",
			envValue: "true",
			want:     true,
		},
		{
			name:     "returns true for any other value",
			envValue: "yes",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env value
			original := os.Getenv("COOKIE_SECURE")
			defer os.Setenv("COOKIE_SECURE", original)

			// Set test env value
			if tt.envValue == "" {
				os.Unsetenv("COOKIE_SECURE")
			} else {
				os.Setenv("COOKIE_SECURE", tt.envValue)
			}

			got := getCookieSecure()
			if got != tt.want {
				t.Errorf("getCookieSecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPasswordHashing verifies bcrypt password hashing works correctly
func TestPasswordHashing(t *testing.T) {
	password := testPassword

	// Generate hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() returned error: %v", err)
	}

	// Verify correct password matches
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		t.Errorf("bcrypt.CompareHashAndPassword() failed for correct password: %v", err)
	}

	// Verify incorrect password fails
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	if err == nil {
		t.Errorf("bcrypt.CompareHashAndPassword() succeeded for incorrect password, expected error")
	}
}

// TestSetSessionCookie verifies session cookie is set with correct attributes
func TestSetSessionCookie(t *testing.T) {
	// Save and restore original env value
	original := os.Getenv("COOKIE_SECURE")
	defer os.Setenv("COOKIE_SECURE", original)

	// Set to false for testing (easier to verify in test environment)
	os.Setenv("COOKIE_SECURE", "false")

	w := httptest.NewRecorder()
	token := "test-token-12345"

	setSessionCookie(w, token)

	// Get the cookie from response
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies set in response")
	}

	cookie := cookies[0]

	// Verify cookie attributes
	if cookie.Name != "session_token" {
		t.Errorf("cookie name = %v, want session_token", cookie.Name)
	}
	if cookie.Value != token {
		t.Errorf("cookie value = %v, want %v", cookie.Value, token)
	}
	if !cookie.HttpOnly {
		t.Errorf("cookie HttpOnly = false, want true")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie SameSite = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
	if cookie.MaxAge != 86400 {
		t.Errorf("cookie MaxAge = %v, want 86400", cookie.MaxAge)
	}
	if cookie.Path != "/" {
		t.Errorf("cookie Path = %v, want /", cookie.Path)
	}
}
