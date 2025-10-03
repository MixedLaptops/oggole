package utils

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() {
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

	// Drop tables if they exist
	_, err = db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("DROP TABLE IF EXISTS pages")
	if err != nil {
		log.Fatal(err)
	}

	// Create users table
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	// Insert admin user
	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		"admin", "keamonk1@stud.kea.dk", "5f4dcc3b5aa765d61d8327deb882cf99")
	if err != nil {
		log.Fatal(err)
	}

	// Create pages table
	pagesSchema := `
	CREATE TABLE IF NOT EXISTS pages (
		title TEXT PRIMARY KEY,
		url TEXT NOT NULL UNIQUE,
		language TEXT NOT NULL CHECK(language IN ('en', 'da')) DEFAULT 'en',
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		content TEXT NOT NULL
	);`

	_, err = db.Exec(pagesSchema)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database initialized successfully")
}
