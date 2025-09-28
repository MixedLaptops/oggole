package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	schema := `
	DROP TABLE IF EXISTS users CASCADE;
	DROP TABLE IF EXISTS pages CASCADE;

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);

	INSERT INTO users (username, email, password)
	VALUES ('admin', 'keamonk1@stud.kea.dk', '5f4dcc3b5aa765d61d8327deb882cf99');

	CREATE TABLE IF NOT EXISTS pages (
		title TEXT PRIMARY KEY,
		url TEXT NOT NULL UNIQUE,
		language TEXT NOT NULL CHECK(language IN ('en', 'da')) DEFAULT 'en',
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		content TEXT NOT NULL
	);`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database initialized successfully")
}
