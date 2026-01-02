package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file from root directory
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	// Get database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("✓ Database connection successful")

	// Check if pages table exists and has data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pages").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query pages table: %v", err)
	}
	fmt.Printf("✓ Pages table exists with %d rows\n", count)

	// Check if content_tsv column is populated
	var nullCount int
	err = db.QueryRow("SELECT COUNT(*) FROM pages WHERE content_tsv IS NULL").Scan(&nullCount)
	if err != nil {
		log.Fatalf("Failed to check content_tsv: %v", err)
	}
	fmt.Printf("✓ content_tsv: %d rows are NULL (should be 0)\n", nullCount)

	// Test the actual search query
	query := "test"
	language := "en"
	tsConfig := "english"
	searchPattern := "%" + query + "%"

	sqlQuery := fmt.Sprintf(`
		SELECT title, url, language, last_updated, content
		FROM pages
		WHERE language = $1
		  AND (content_tsv @@ plainto_tsquery('%s'::regconfig, $2)
		       OR title ILIKE $3
		       OR content ILIKE $3)
		ORDER BY ts_rank(content_tsv, plainto_tsquery('%s'::regconfig, $2)) DESC
		LIMIT 50
	`, tsConfig, tsConfig)

	fmt.Println("\nTesting search query...")
	fmt.Printf("Query: %s\nLanguage: %s\nPattern: %s\n\n", query, language, searchPattern)

	rows, err := db.Query(sqlQuery, language, query, searchPattern)
	if err != nil {
		log.Fatalf("✗ Search query failed: %v", err)
	}
	defer rows.Close()

	resultCount := 0
	for rows.Next() {
		var title, url, lang, lastUpdated, content string
		if err := rows.Scan(&title, &url, &lang, &lastUpdated, &content); err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		resultCount++
		fmt.Printf("  - %s (%s)\n", title, url)
	}

	fmt.Printf("\n✓ Search query successful: %d results\n", resultCount)
}
