package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

func CreateTablesSQL(path string) *sql.DB {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Fatalf("Failed to create DB folder: %v", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("Failed to open SQLite: %v", err)
	}

	schemaFile := "schema.sql"
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema file %s: %v", schemaFile, err)
	}

	if _, err := db.Exec(string(schemaBytes)); err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}
	log.Println("created db schema")
	return db
}
