package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/http"
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
			ipv4 := ""
			if hostname != "" {
				log.Printf("Attempting to resolve hostname: %s\n", hostname)
				resolved, resolveErr := resolveIPv4(hostname)
				if resolveErr == nil && resolved != "" {
					ipv4 = resolved
					log.Printf("Resolved to IPv4: %s\n", ipv4)
				}
			}

			query := parsedURL.Query()
			if query.Get("sslmode") == "" {
				query.Set("sslmode", "require")
			}

			connectionString = buildDSNFromURL(parsedURL, ipv4, query)
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

	if ipv4 := lookupIPv4WithResolver(ctx, net.DefaultResolver, hostname); ipv4 != "" {
		return ipv4, nil
	}

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "udp4", "1.1.1.1:53")
		},
	}

	if ipv4 := lookupIPv4WithResolver(ctx, resolver, hostname); ipv4 != "" {
		return ipv4, nil
	}

	if ipv4 := lookupIPv4WithDoH(ctx, hostname); ipv4 != "" {
		return ipv4, nil
	}

	return "", nil
}

func lookupIPv4WithResolver(ctx context.Context, resolver *net.Resolver, hostname string) string {
	ips, err := resolver.LookupIP(ctx, "ip4", hostname)
	if err != nil {
		return ""
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}

	return ""
}

func lookupIPv4WithDoH(ctx context.Context, hostname string) string {
	endpoint := "https://cloudflare-dns.com/dns-query?name=" + url.QueryEscape(hostname) + "&type=A"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/dns-json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	var parsed struct {
		Answer []struct {
			Data string `json:"data"`
			Type int    `json:"type"`
		} `json:"Answer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return ""
	}

	for _, ans := range parsed.Answer {
		if ans.Type == 1 && net.ParseIP(ans.Data) != nil {
			return ans.Data
		}
	}

	return ""
}

func buildDSNFromURL(parsedURL *url.URL, hostaddr string, query url.Values) string {
	username := ""
	password := ""
	if parsedURL.User != nil {
		username = parsedURL.User.Username()
		if pwd, ok := parsedURL.User.Password(); ok {
			password = pwd
		}
	}

	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "5432"
	}

	dbname := strings.TrimPrefix(parsedURL.Path, "/")
	if dbname == "" {
		dbname = "postgres"
	}

	parts := []string{}
	if host != "" {
		parts = append(parts, "host="+host)
	}
	if hostaddr != "" {
		parts = append(parts, "hostaddr="+hostaddr)
	}
	if port != "" {
		parts = append(parts, "port="+port)
	}
	if username != "" {
		parts = append(parts, "user="+username)
	}
	if password != "" {
		parts = append(parts, "password="+password)
	}
	if dbname != "" {
		parts = append(parts, "dbname="+dbname)
	}

	sslmode := query.Get("sslmode")
	if sslmode == "" {
		sslmode = "require"
	}
	parts = append(parts, "sslmode="+sslmode)

	return strings.Join(parts, " ")
}
