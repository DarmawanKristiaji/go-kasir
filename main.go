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

func main() {
	// Load environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	// Get config from viper
	config := Config{
		Port:   viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	// Setup database
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
		log.Println("Starting server without database connection...")
		// Continue without database for now
	}

	// Only setup product and category endpoints if DB is available
	if db != nil {
		defer db.Close()

		// Dependency Injection - Product
		productRepo := repositories.NewProductRepository(db)
		productService := services.NewProductService(productRepo)
		productHandler := handlers.NewProductHandler(productService)

		// Setup routes for products - handle both with and without trailing slash
		http.HandleFunc("/api/produk", func(w http.ResponseWriter, r *http.Request) {
			// Strip trailing slash for consistency
			path := strings.TrimSuffix(r.URL.Path, "/")
			
			if path == "/api/produk" {
				productHandler.HandleProducts(w, r)
			} else {
				productHandler.HandleProductByID(w, r)
			}
		})

		// Dependency Injection - Category
		categoryRepo := repositories.NewCategoryRepository(db)
		categoryService := services.NewCategoryService(categoryRepo)
		categoryHandler := handlers.NewCategoryHandler(categoryService)

		// Setup routes for categories - handle both with and without trailing slash
		http.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
			// Strip trailing slash for consistency
			path := strings.TrimSuffix(r.URL.Path, "/")
			
			if path == "/categories" {
				categoryHandler.HandleCategories(w, r)
			} else {
				categoryHandler.HandleCategoryByID(w, r)
			}
		})
	} else {
		log.Println("Warning: Product and Category endpoints disabled (no database)")
	}

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"OK","message":"API Running"}`)
	})

	// Start server
	addr := "0.0.0.0:" + config.Port
	fmt.Printf("Server running di %s\n", addr)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("gagal running server", err)
	}
}
