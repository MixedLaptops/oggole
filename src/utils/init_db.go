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

	// Insert admin user with bcrypt hashed password
	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		"admin", "keamonk1@stud.kea.dk", "$2a$10$B4J5hK1kdHTOOmTXNCuaquqCe17t/2tB6.7gl3fAgURhHPX2TIQuC")
	if err != nil {
		log.Fatal(err)
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
