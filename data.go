package main

import (
	"crypto/sha512"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func AddVisited(db *sqlx.DB, url string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO visited (url) VALUES (?)", url)
	return err
}

func IsVisited(db *sqlx.DB, url string) (bool, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM visited WHERE url = ?", url)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func AddPageVisit(db *sqlx.DB, url string, statusCode int, contentID string) (string, error) {
	id := uuid.New().String()
	date := time.Now().Unix()
	_, err := db.Exec(
		"INSERT INTO page_visit (id, date, url, status_code, content_id) VALUES (?, ?, ?, ?, ?)",
		id, date, url, statusCode, contentID,
	)
	return id, err
}

func AddContent(db *sqlx.DB, content []byte) (string, error) {
	hash := sha512.Sum512(content)
	hashStr := hex.EncodeToString(hash[:])

	// Check if page with this hash already exists
	var existingID string
	err := db.Get(&existingID, "SELECT id FROM content WHERE sha512_hash = ?", hashStr)
	if err == nil {
		// Page already exists, return existing id
		return existingID, nil
	}

	// Page doesn't exist, insert new one
	id := uuid.New().String()
	_, err = db.Exec(
		"INSERT INTO content (id, content, sha512_hash) VALUES (?, ?, ?)",
		id, content, hashStr,
	)
	return id, err
}

func EmptyVisited(db *sqlx.DB) error {
	_, err := db.Exec("DELETE FROM visited")
	return err
}
