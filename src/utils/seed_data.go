package utils

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func SeedData() {
	// Get database URL from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	now := time.Now().Format("2006-01-02 15:04:05")

	// No seed data - using actual crawled pages
	samplePages := []struct {
		title       string
		url         string
		language    string
		content     string
	}{}

	for _, page := range samplePages {
		_, err := db.Exec(
			"INSERT INTO pages (title, url, language, last_updated, content) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (title) DO UPDATE SET url = excluded.url, language = excluded.language, last_updated = excluded.last_updated, content = excluded.content",
			page.title, page.url, page.language, now, page.content,
		)
		if err != nil {
			log.Printf("Error inserting page '%s': %v", page.title, err)
		} else {
			log.Printf("Inserted page: %s (%s)", page.title, page.language)
		}
	}

	// Check final count
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pages").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Database now has %d pages", count)
}