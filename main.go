package main

import (
	"log"
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
	if cfg.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET is required!")
	}

	// Initialize database
	db, err := database.NewSQLiteDB(cfg.DBDSN)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWT_SECRET)
	urlHandler := handlers.NewURLHandler(db)
	resetHandler := handlers.NewPasswordResetHandler(db)

	// Setup routes
	r := routes.SetupRouter(
		urlHandler,
		authHandler,
		resetHandler,
		middleware.JWTAuthMiddleware(cfg.JWT_SECRET),
	)

	// Start server
	log.Printf("Server running on %s", cfg.BASE_URL)
	log.Fatal(r.Run(cfg.BASE_URL))
}