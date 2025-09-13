package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

func main() {
	// Test database connection
	db, err := sql.Open("sqlite", "../whoknows.db")
	if err != nil {
		fmt.Println("❌ Database connection failed:", err)
		return
	}
	defer db.Close()

	// Try a simple query
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		fmt.Println("❌ Query failed:", err)
		return
	}

	fmt.Printf("✅ Database works! Found %d users\n", count)
}