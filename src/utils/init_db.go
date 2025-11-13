package utils

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() {
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

	// Drop tables if they exist
	_, err = db.Exec("DROP TABLE IF EXISTS sessions")
	if err != nil {
		log.Fatal(err)
	}
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
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	// Kode til at sikre os at vores admin user ikke bliver hardcodet men får info fra env variabler
	// Insert admin user only if explicitly requested
	initAdmin := os.Getenv("INIT_ADMIN")
	env := os.Getenv("ENV")

	if initAdmin == "true" {
		// Check production safety
		if env == "production" {
			confirmProd := os.Getenv("CONFIRM_PRODUCTION_ADMIN")
			if confirmProd != "true" {
				log.Fatal("ERROR: Refusing to create admin in production without CONFIRM_PRODUCTION_ADMIN=true")
			}
		}

		// Read admin credentials from environment
		adminUsername := os.Getenv("ADMIN_USERNAME")
		adminEmail := os.Getenv("ADMIN_EMAIL")
		adminPasswordHash := os.Getenv("ADMIN_PASSWORD_HASH")

		// Validate all required variables are present
		if adminUsername == "" || adminEmail == "" || adminPasswordHash == "" {
			log.Fatal("ERROR: INIT_ADMIN=true but missing required environment variables:\n" +
				"  ADMIN_USERNAME, ADMIN_EMAIL, ADMIN_PASSWORD_HASH\n" +
				"  Generate hash with: go run your_hash_generator.go")
		}

		// Insert admin user
		_, err = db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
			adminUsername, adminEmail, adminPasswordHash)
		if err != nil {
			log.Fatalf("ERROR: Failed to create admin user: %v", err)
		}

		log.Printf("Admin user created: %s (%s)", adminUsername, adminEmail)
	} else {
		log.Println("Skipping admin user creation (INIT_ADMIN not set to 'true')")
	}

	// Create pages table
	pagesSchema := `
	CREATE TABLE IF NOT EXISTS pages (
		title TEXT PRIMARY KEY,
		url TEXT NOT NULL UNIQUE,
		language TEXT NOT NULL CHECK(language IN ('en', 'da')) DEFAULT 'en',
		last_updated TIMESTAMP DEFAULT NOW(),
		content TEXT NOT NULL
	);`

	_, err = db.Exec(pagesSchema)
	if err != nil {
		log.Fatal(err)
	}

	// Create sessions table
	sessionsSchema := `
	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);`

	_, err = db.Exec(sessionsSchema)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database initialized successfully")
}
