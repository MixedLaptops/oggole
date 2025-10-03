package utils

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func SeedData() {
	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "data/oggole.db"
	}

	db, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	now := time.Now().Format("2006-01-02 15:04:05")

	samplePages := []struct {
		title       string
		url         string
		language    string
		content     string
	}{
		{
			title:    "Go Programming Tutorial",
			url:      "https://example.com/go-tutorial",
			language: "en",
			content:  "Learn Go programming language with this comprehensive tutorial. Go is a modern programming language developed by Google. It features garbage collection, memory safety, and excellent performance.",
		},
		{
			title:    "Golang Web Development",
			url:      "https://example.com/golang-web",
			language: "en",
			content:  "Build web applications using Golang. This guide covers HTTP servers, routing, templates, and database integration. Golang makes web development simple and efficient.",
		},
		{
			title:    "Database with Go",
			url:      "https://example.com/go-database",
			language: "en",
			content:  "Working with databases in Go programming. Learn about SQL drivers, prepared statements, and best practices for database operations in Go applications.",
		},
		{
			title:    "Go Programmering Guide",
			url:      "https://example.dk/go-guide",
			language: "da",
			content:  "Lær Go programmering på dansk. Denne guide dækker grundlæggende koncepter og avancerede teknikker i Go sproget. Go er et kraftfuldt og moderne programmeringssprog.",
		},
		{
			title:    "React Frontend Development",
			url:      "https://example.com/react-guide",
			language: "en",
			content:  "Master React for frontend development. Learn components, hooks, state management, and modern React patterns. Build interactive user interfaces with React.",
		},
		{
			title:    "Python Data Science",
			url:      "https://example.com/python-data",
			language: "en",
			content:  "Python for data science and machine learning. Explore pandas, numpy, matplotlib, and scikit-learn. Analyze data and build predictive models with Python.",
		},
	}

	for _, page := range samplePages {
		_, err := db.Exec(
			"INSERT INTO pages (title, url, language, last_updated, content) VALUES (?, ?, ?, ?, ?) ON CONFLICT (title) DO UPDATE SET url = excluded.url, language = excluded.language, last_updated = excluded.last_updated, content = excluded.content",
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