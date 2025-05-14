package main

import (
	"log"
	"fmt"
	"goshort-api/configs"
	"goshort-api/internal/database"
	"goshort-api/internal/handlers"
	"goshort-api/internal/routes"
	"goshort-api/internal/middleware"
)

func main() {
	// Load configuration
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Validate JWT secret in production
	if cfg.AppEnv == "production" && cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required in production")
	}

	// Initialize database
	db, err := database.NewSQLiteDB(cfg.DBDSN)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	urlHandler := handlers.NewURLHandler(db)

	// Setup routes
	r := routes.SetupRouter(
		urlHandler,
		authHandler,
		middleware.JWTAuthMiddleware(cfg.JWTSecret),
	)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server running on port %d in %s mode", cfg.Port, cfg.AppEnv)
	log.Fatal(r.Run(addr))
}