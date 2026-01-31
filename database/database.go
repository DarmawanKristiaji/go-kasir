package database

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func InitDB(connectionString string) (*sql.DB, error) {
	if connectionString == "" {
		log.Println("Connection string is empty")
		return nil, sql.ErrNoRows
	}

	// Handle URL-encoded passwords - decode if it's a URL format
	if strings.HasPrefix(connectionString, "postgresql://") || strings.HasPrefix(connectionString, "postgres://") {
		// Parse URL and unescape password
		parsedURL, err := url.Parse(connectionString)
		if err == nil && parsedURL.User != nil {
			password, _ := parsedURL.User.Password()
			log.Printf("Detected URL format connection string with user: %s\n", parsedURL.User.Username())
			// Reconstruct with properly decoded password (pq driver handles this)
		}
	}

	// Open database with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Open database
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Failed to open database: %v\n", err)
		return nil, err
	}

	// Test connection with timeout
	err = db.PingContext(ctx)
	if err != nil {
		log.Printf("Database ping failed: %v\n", err)
		db.Close()
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connected successfully")
	return db, nil
}
