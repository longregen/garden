package main

import (
	"context"
	"log"

	"garden3/internal/adapter/primary/http"
	"garden3/internal/adapter/primary/http/handler"
	"garden3/internal/adapter/secondary/postgres"
	"garden3/internal/adapter/secondary/postgres/repository"
	"garden3/internal/domain/service"
)

func main() {
	ctx := context.Background()

	// Initialize database
	db, err := postgres.NewDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	configRepo := repository.NewConfigurationRepository(db.Pool)

	// Initialize services
	configService := service.NewConfigurationService(configRepo)

	// Initialize HTTP server
	server := http.NewServer()

	// Initialize handlers and register routes
	configHandler := handler.NewConfigurationHandler(configService)
	configHandler.RegisterRoutes(server.Router())

	// Start server
	log.Println("Starting API server...")
	if err := server.Start("8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
