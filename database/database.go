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

		if err == nil {
			hostname := parsedURL.Hostname()
			if hostname != "" {
				log.Printf("Attempting to resolve hostname: %s\n", hostname)
				ipv4, resolveErr := resolveIPv4(hostname)
				if resolveErr == nil && ipv4 != "" {
					log.Printf("Resolved to IPv4: %s\n", ipv4)
					query := parsedURL.Query()
					query.Set("hostaddr", ipv4)
					parsedURL.RawQuery = query.Encode()
				}
			}

			query := parsedURL.Query()
			if query.Get("sslmode") == "" {
				query.Set("sslmode", "require")
				parsedURL.RawQuery = query.Encode()
			}

			connectionString = parsedURL.String()
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

				ipv4, err := resolveIPv4(hostname)
				if err == nil && ipv4 != "" {
					log.Printf("Resolved to IPv4: %s\n", ipv4)
					parts[i] = "host=" + hostname
					parts = append(parts, "hostaddr="+ipv4)
					connectionString = strings.Join(parts, " ")
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

func resolveIPv4(hostname string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", hostname)
	if err != nil {
		return "", err
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	return "", nil
}
