package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// setupTestDB initializes a test database connection
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL environment variable required for integration tests")
	}

	testDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err = testDB.Ping(); err != nil {
		testDB.Close()
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Set the global db variable for handlers to use
	db = testDB

	cleanup := func() {
		// Clean up test data
		testDB.Exec("DELETE FROM sessions WHERE username LIKE 'testuser%'")
		testDB.Exec("DELETE FROM users WHERE username LIKE 'testuser%'")
		testDB.Close()
	}

	return testDB, cleanup
}

// TestSearchHandler_Integration verifies search functionality
func TestSearchHandler_Integration(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/search?q=test&language=en", nil)
	w := httptest.NewRecorder()

	search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var pages []Page
	if err := json.NewDecoder(w.Body).Decode(&pages); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should return array (even if empty)
	if pages == nil {
		t.Errorf("expected array response, got nil")
	}
}

// TestLoginHandler_InvalidPassword verifies authentication fails correctly
func TestLoginHandler_InvalidPassword(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	_, err := testDB.Exec(`
		INSERT INTO users (username, email, password, registration_ip)
		VALUES ($1, $2, $3, $4)
	`, "testuser_auth", "test@example.com", string(hashedPassword), "127.0.0.1")

	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Try with wrong password
	form := url.Values{}
	form.Set("username", "testuser_auth")
	form.Set("password", "wrongpassword")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestLogoutHandler_Integration verifies logout clears session
func TestLogoutHandler_Integration(t *testing.T) {
	testDB, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test user and session
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	_, err := testDB.Exec(`
		INSERT INTO users (username, email, password, registration_ip)
		VALUES ($1, $2, $3, $4)
	`, "testuser_logout", "logout@example.com", string(hashedPassword), "127.0.0.1")

	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create session
	token := "test-session-token-12345"
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = testDB.Exec(`
		INSERT INTO sessions (token, username, expires_at)
		VALUES ($1, $2, $3)
	`, token, "testuser_logout", expiresAt)

	if err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}

	// Logout
	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})
	w := httptest.NewRecorder()

	logout(w, req)

	// Should redirect to login
	if w.Code != http.StatusSeeOther {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusSeeOther)
	}

	// Verify session was deleted from database
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = $1", token).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}

	if count != 0 {
		t.Errorf("session still exists in database, expected deletion")
	}
}
