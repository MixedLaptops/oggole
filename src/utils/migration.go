package utils

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func Migration() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Add content_tsv column
	_, err = tx.Exec(`ALTER TABLE pages ADD COLUMN IF NOT EXISTS content_tsv tsvector;`)
	if err != nil {
		log.Fatalf("Failed to add content_tsv column: %v", err)
	}

	// Create trigger function
	_, err = tx.Exec(`
		CREATE OR REPLACE FUNCTION update_content_tsv() RETURNS trigger AS $$
		BEGIN
			NEW.content_tsv := to_tsvector(
				CASE NEW.language
					WHEN 'en' THEN 'english'::regconfig
					WHEN 'da' THEN 'danish'::regconfig
					ELSE 'english'::regconfig
				END,
				NEW.content
			);
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		log.Fatalf("Failed to create trigger function: %v", err)
	}

	// Create trigger
	_, err = tx.Exec(`
		DROP TRIGGER IF EXISTS tsvector_update ON pages;
		CREATE TRIGGER tsvector_update BEFORE INSERT OR UPDATE
		ON pages FOR EACH ROW EXECUTE FUNCTION update_content_tsv();
	`)
	if err != nil {
		log.Fatalf("Failed to create trigger: %v", err)
	}

	// Populate existing rows
	_, err = tx.Exec(`
		UPDATE pages
		SET content_tsv = to_tsvector(
			CASE language
				WHEN 'en' THEN 'english'::regconfig
				WHEN 'da' THEN 'danish'::regconfig
				ELSE 'english'::regconfig
			END,
			content
		)
		WHERE content_tsv IS NULL;
	`)
	if err != nil {
		log.Fatalf("Failed to populate tsvector: %v", err)
	}

	// Create GIN index
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS content_tsv_idx ON pages USING GIN(content_tsv);`)
	if err != nil {
		log.Fatalf("Failed to create GIN index: %v", err)
	}

	if err = tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("Migration completed successfully")
}
