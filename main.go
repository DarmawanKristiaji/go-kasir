package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
	"kasir-api/database"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
)

type Config struct {
	Port   string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
}

// maskConnectionString hides sensitive info from logs
func maskConnectionString(connStr string) string {
	if len(connStr) < 50 {
		return "***MASKED***"
	}
	return connStr[:20] + "***MASKED***" + connStr[len(connStr)-20:]
}

func main() {
	// Load environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	// Get config from viper - try multiple sources
	config := Config{
		Port:   viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}

	// Fallback: try reading directly from os.Getenv if viper didn't find it
	if config.DBConn == "" {
		config.DBConn = os.Getenv("DB_CONN")
		if config.DBConn != "" {
			log.Println("DB_CONN loaded from os.Getenv (viper fallback)")
		}
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	// Log database connection status
	if config.DBConn == "" {
		log.Println("ERROR: DB_CONN environment variable not set")
		log.Println("Available env vars: PORT =", config.Port)
	} else {
		log.Printf("DB_CONN found with length: %d\n", len(config.DBConn))
		log.Printf("Attempting database connection to: %s\n", maskConnectionString(config.DBConn))
	}

	// Setup database
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Printf("WARNING: Failed to initialize database: %v", err)
		log.Printf("Connection string length: %d", len(config.DBConn))
		log.Println("Starting server without database connection...")
		// Continue without database for now
		db = nil
	}

	if db != nil {
		log.Println("Database connected successfully")
	} else {
		log.Println("WARNING: No database connection - routes may not be registered")
	}

	// Health check - register first
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		dbStatus := "not connected"
		dbConnLen := 0
		if db != nil {
			dbStatus = "connected"
		}
		if config.DBConn != "" {
			dbConnLen = len(config.DBConn)
		}
		fmt.Fprintf(w, `{"status":"OK","message":"API Running - Go Kasir POS System","version":"1.0","database":"%s","db_conn_length":%d}`, dbStatus, dbConnLen)
	})

	// Debug endpoint (remove in production)
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		dbConnSet := config.DBConn != ""
		dbConnSample := ""
		if dbConnSet && len(config.DBConn) > 20 {
			dbConnSample = config.DBConn[:20] + "..."
		}
		fmt.Fprintf(w, `{"db_conn_set":%t,"db_conn_sample":"%s","port":"%s"}`, dbConnSet, dbConnSample, config.Port)
	})

	// Only setup product and category endpoints if DB is available
	if db != nil {
		defer db.Close()

		// Dependency Injection - Product
		productRepo := repositories.NewProductRepository(db)
		productService := services.NewProductService(productRepo)
		productHandler := handlers.NewProductHandler(productService)

		// Setup routes for products - register handler for both paths
		productRouter := func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/produk/" || r.URL.Path == "/api/produk" {
				productHandler.HandleProducts(w, r)
			} else {
				productHandler.HandleProductByID(w, r)
			}
		}
		http.HandleFunc("/api/produk", productRouter)
		http.HandleFunc("/api/produk/", productRouter)

		// Dependency Injection - Category
		categoryRepo := repositories.NewCategoryRepository(db)
		categoryService := services.NewCategoryService(categoryRepo)
		categoryHandler := handlers.NewCategoryHandler(categoryService)

		// Setup routes for categories - register handler for both paths
		categoryRouter := func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/categories/" || r.URL.Path == "/categories" {
				categoryHandler.HandleCategories(w, r)
			} else {
				categoryHandler.HandleCategoryByID(w, r)
			}
		}
		http.HandleFunc("/categories", categoryRouter)
		http.HandleFunc("/categories/", categoryRouter)
	} else {
		log.Println("WARNING: No database connection - routes disabled")
		// Still register placeholder endpoints for debugging
		http.HandleFunc("/api/produk", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"error","message":"Database not connected"}`)
		})
		http.HandleFunc("/api/produk/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"error","message":"Database not connected"}`)
		})
		http.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"error","message":"Database not connected"}`)
		})
		http.HandleFunc("/categories/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"error","message":"Database not connected"}`)
		})
	}

	// Start server
	addr := "0.0.0.0:" + config.Port
	fmt.Printf("Server running di %s\n", addr)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("gagal running server", err)
	}
}
