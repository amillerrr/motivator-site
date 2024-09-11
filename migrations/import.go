package migrations

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"

	_ "github.com/lib/pq"
)

// Struct without the ID field
type Data struct {
	Author  string `json:"author"`
	Message string `json:"message"`
}

// insertDataFromJSON reads a JSON file, unmarshals its contents, and inserts data into PostgreSQL
func insertDataFromJSON(db *sql.DB, fsys fs.FS, filePath string, category string) error {
	// 1. Read the JSON file
	fileBytes, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return fmt.Errorf("error reading embedded JSON file (%s): %w", filePath, err)
	}

	// 2. Unmarshal JSON into Go struct
	var jsonData []Data
	if err := json.Unmarshal(fileBytes, &jsonData); err != nil {
		return fmt.Errorf("error unmarshalling JSON from file (%s): %w", filePath, err)
	}

	// 3. Insert data into PostgreSQL with category
	for _, entry := range jsonData {
		_, err := db.Exec("INSERT INTO quotes (author, message, category) VALUES ($1, $2, $3)", entry.Author, entry.Message, category)
		if err != nil {
			return fmt.Errorf("error inserting data from file (%s): %w", filePath, err)
		}
	}

	log.Printf("Data from %s inserted successfully under category %s!\n", filePath, category)
	return nil
}

func ImportGrit(db *sql.DB, fsys fs.FS) error {
	return insertDataFromJSON(db, fsys, "data/grit.json", "grit")
}

func ImportGratitude(db *sql.DB, fsys fs.FS) error {
	return insertDataFromJSON(db, fsys, "data/gratitude.json", "gratitude")
}

func ImportPerseverance(db *sql.DB, fsys fs.FS) error {
	return insertDataFromJSON(db, fsys, "data/perseverance.json", "perseverance")
}
