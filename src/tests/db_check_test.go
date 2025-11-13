package main

import (
	"database/sql"
	"os"
	"testing"
	_ "github.com/lib/pq"
)

func TestDatabaseConnection(t *testing.T) {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL environment variable is required")
	}

	// Test database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Try a simple query
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	t.Logf("Database works! Found %d users", count)
}
