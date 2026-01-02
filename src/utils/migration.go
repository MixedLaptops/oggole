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

	// Add content_tsv column
	_, err = db.Exec(`ALTER TABLE pages ADD COLUMN IF NOT EXISTS content_tsv tsvector;`)
	if err != nil {
		log.Fatal(err)
	}

	// Create trigger function
	_, err = db.Exec(`
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
		log.Fatal(err)
	}

	// Create trigger
	_, err = db.Exec(`
		DROP TRIGGER IF EXISTS tsvector_update ON pages;
		CREATE TRIGGER tsvector_update BEFORE INSERT OR UPDATE
		ON pages FOR EACH ROW EXECUTE FUNCTION update_content_tsv();
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Populate existing rows
	_, err = db.Exec(`
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
		log.Fatal(err)
	}

	// Create GIN index
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS content_tsv_idx ON pages USING GIN(content_tsv);`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Migration completed successfully")
}
