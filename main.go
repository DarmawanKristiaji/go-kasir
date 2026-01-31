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

	// Dependency Injection - Product
	if db != nil {
		defer db.Close()
		productRepo := repositories.NewProductRepository(db)
		productService := services.NewProductService(productRepo)
		productHandler := handlers.NewProductHandler(productService)

		// Setup routes
		http.HandleFunc("/api/produk", productHandler.HandleProducts)
		http.HandleFunc("/api/produk/", productHandler.HandleProductByID)

		// Dependency Injection - Category
		categoryRepo := repositories.NewCategoryRepository(db)
		categoryService := services.NewCategoryService(categoryRepo)
		categoryHandler := handlers.NewCategoryHandler(categoryService)

		// Setup category routes
		http.HandleFunc("/categories", categoryHandler.HandleCategories)
		http.HandleFunc("/categories/", categoryHandler.HandleCategoryByID)
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
