package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/adapter/primary/http/handler"
	"garden3/internal/adapter/secondary/ai"
	"garden3/internal/adapter/secondary/contentprocessor"
	"garden3/internal/adapter/secondary/embedding"
	"garden3/internal/adapter/secondary/httpfetch"
	"garden3/internal/adapter/secondary/llm"
	"garden3/internal/adapter/secondary/postgres"
	"garden3/internal/adapter/secondary/postgres/repository"
	"garden3/internal/adapter/secondary/social"
	"garden3/internal/domain/service"
)

func main() {
	ctx := context.Background()

	// Initialize database
	log.Println("Connecting to database...")
	db, err := postgres.NewDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// Initialize repositories
	configRepo := repository.NewConfigurationRepository(db.Pool)
	contactRepo := repository.NewContactRepository(db.Pool)
	roomRepo := repository.NewRoomRepository(db.Pool)
	messageRepo := repository.NewMessageRepository(db.Pool)
	sessionRepo := repository.NewSessionRepository(db.Pool)
	noteRepo := repository.NewNoteRepository(db.Pool)
	itemRepo := repository.NewItemRepository(db.Pool)
	bookmarkRepo := repository.NewBookmarkRepository(db.Pool)
	entityRepo := repository.NewEntityRepository(db.Pool)
	categoryRepo := repository.NewCategoryRepository(db.Pool)
	socialPostRepo := repository.NewSocialPostRepository(db.Pool)
	observationRepo := repository.NewObservationRepository(db.Pool)
	dashboardRepo := repository.NewDashboardRepository(db.Pool)
	browserHistoryRepo := repository.NewBrowserHistoryRepository(db.Pool)
	searchRepo := repository.NewSearchRepository(db.Pool)
	tagRepo := repository.NewTagRepository(db.Pool)

	// Initialize external service adapters
	// Get Ollama configuration for embeddings
	ollamaEmbedURL := os.Getenv("OLLAMA_EMBED_API_URL")
	if ollamaEmbedURL == "" {
		ollamaEmbedURL = os.Getenv("OLLAMA_API_URL") // Fall back to main Ollama URL
	}
	ollamaEmbedModel := os.Getenv("OLLAMA_EMBED_MODEL")
	if ollamaEmbedModel == "" {
		ollamaEmbedModel = "nomic-embed-text:latest"
	}

	embeddingService := embedding.NewOllamaEmbeddingService(ollamaEmbedURL, ollamaEmbedModel)
	embeddingsService := embedding.NewOllamaEmbeddingsService(ollamaEmbedURL, ollamaEmbedModel)
	socialMediaService := social.NewService(configRepo)
	httpFetcher := httpfetch.NewFetcher()

	// Initialize AI service (for summary generation)
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = os.Getenv("OLLAMA_API_URL") // Fall back to Ollama URL
	}
	aiServiceKey := os.Getenv("AI_SERVICE_KEY")
	aiService := ai.NewService(aiServiceURL, aiServiceKey)

	contentProcessor := contentprocessor.NewProcessor()

	// Initialize LLM service (Ollama)
	ollamaURL := os.Getenv("OLLAMA_API_URL")
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	llmService := llm.NewOllamaService(ollamaURL, ollamaModel)

	// Initialize domain services
	configService := service.NewConfigurationService(configRepo)
	contactService := service.NewContactService(contactRepo)
	roomService := service.NewRoomService(roomRepo)
	messageService := service.NewMessageService(messageRepo)
	sessionService := service.NewSessionService(sessionRepo, embeddingService)
	noteService := service.NewNoteService(noteRepo, embeddingsService)
	itemService := service.NewItemService(itemRepo)
	bookmarkService := service.NewBookmarkService(bookmarkRepo, httpFetcher, embeddingsService, aiService, contentProcessor)
	entityService := service.NewEntityService(entityRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	socialPostService := service.NewSocialPostService(socialPostRepo, socialMediaService)
	observationService := service.NewObservationService(observationRepo)
	dashboardService := service.NewDashboardService(dashboardRepo)
	browserHistoryService := service.NewBrowserHistoryService(browserHistoryRepo)
	searchService := service.NewSearchService(searchRepo, embeddingService, llmService, configService)
	utilityService := service.NewUtilityService(sessionRepo, messageRepo, configRepo, db.Pool)
	logseqSyncService := service.NewLogseqSyncService(configService, entityRepo)
	tagService := service.NewTagService(tagRepo)

	// Initialize HTTP handlers
	configHandler := handler.NewConfigurationHandler(configService)
	contactHandler := handler.NewContactHandler(contactService)
	roomHandler := handler.NewRoomHandler(roomService)
	messageHandler := handler.NewMessageHandler(messageService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	noteHandler := handler.NewNoteHandler(noteService)
	itemHandler := handler.NewItemHandler(itemService, tagService)
	bookmarkHandler := handler.NewBookmarkHandler(bookmarkService)
	entityHandler := handler.NewEntityHandler(entityService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	socialPostHandler := handler.NewSocialPostHandler(socialPostService)
	observationHandler := handler.NewObservationHandler(observationService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	browserHistoryHandler := handler.NewBrowserHistoryHandler(browserHistoryService)
	searchHandler := handler.NewSearchHandler(searchService)
	utilityHandler := handler.NewUtilityHandler(utilityService)
	logseqHandler := handler.NewLogseqHandler(logseqSyncService, entityRepo)
	tagHandler := handler.NewTagHandler(tagService)

	// Initialize HTTP server
	server := httpAdapter.NewServer()
	router := server.Router()

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpAdapter.JSON(w, 200, map[string]string{"status": "ok"})
	})

	// Register all routes
	configHandler.RegisterRoutes(router)
	contactHandler.RegisterRoutes(router)
	roomHandler.RegisterRoutes(router)
	messageHandler.RegisterRoutes(router)
	sessionHandler.RegisterRoutes(router)
	noteHandler.RegisterRoutes(router)
	itemHandler.RegisterRoutes(router)
	bookmarkHandler.RegisterRoutes(router)
	entityHandler.RegisterRoutes(router)
	categoryHandler.RegisterRoutes(router)
	socialPostHandler.RegisterRoutes(router)
	observationHandler.RegisterRoutes(router)
	dashboardHandler.RegisterRoutes(router)
	browserHistoryHandler.RegisterRoutes(router)
	searchHandler.RegisterRoutes(router)
	utilityHandler.RegisterRoutes(router)
	logseqHandler.RegisterRoutes(router)
	tagHandler.RegisterRoutes(router)

	log.Println("Routes registered")

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Println("Server starting on :8080")
		if err := server.Start(""); err != nil {
			serverErrors <- err
		}
	}()

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Received signal %v, shutting down gracefully...", sig)

		// Give outstanding requests 30 seconds to complete
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server (this would need to be implemented in the Server type)
		// For now, just close the database
		_ = shutdownCtx // TODO: use this context when implementing graceful shutdown
		db.Close()

		log.Println("Shutdown complete")
	}
}
