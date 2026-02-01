package database

import (
	"context"
	"database/sql"
	"log"
	"net"
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
			log.Printf("Detected URL format connection string with user: %s\n", parsedURL.User.Username())
			// pq driver handles URL decoding automatically
		}
	}

	// Force IPv4 by resolving hostname first
	if strings.Contains(connectionString, "host=") {
		// For DSN format, try to resolve hostname to IPv4
		parts := strings.Split(connectionString, " ")
		for i, part := range parts {
			if strings.HasPrefix(part, "host=") {
				hostname := strings.TrimPrefix(part, "host=")
				log.Printf("Attempting to resolve hostname: %s\n", hostname)
				
				// Use net.LookupIP with IPv4 preference
				ips, err := net.LookupIP(hostname)
				if err == nil && len(ips) > 0 {
					// Find IPv4 address
					for _, ip := range ips {
						if ipv4 := ip.To4(); ipv4 != nil {
							log.Printf("Resolved to IPv4: %s\n", ipv4.String())
							parts[i] = "host=" + ipv4.String()
							connectionString = strings.Join(parts, " ")
							break
						}
					}
				}
				break
			}
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
