package main

import (
	"database/sql"
	"testing"
	_ "modernc.org/sqlite"
)

func TestDatabaseConnection(t *testing.T) {
	// Test database connection
	db, err := sql.Open("sqlite", "../whoknows.db")
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
