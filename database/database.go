package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents a database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return &DB{db}, nil
}

// CreateTable creates a new table with the given name and columns
func (db *DB) CreateTable(tableName string, columns []string) error {
	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, columns)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}
	return nil
}

// InsertRecord inserts a single record into the specified table
func (db *DB) InsertRecord(tableName string, columns []string, values []interface{}) error {
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	_, err := db.Exec(insertSQL, values...)
	if err != nil {
		return fmt.Errorf("error inserting record: %v", err)
	}
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}