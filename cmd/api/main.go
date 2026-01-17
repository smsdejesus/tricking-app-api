package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tricking-api/internal/config"
	"tricking-api/internal/database"
	"tricking-api/internal/handlers"
	"tricking-api/internal/repository"
	"tricking-api/internal/routes"
	"tricking-api/internal/services"
)

func main() {
	// STEP 1: Load Configuration
	cfg, err := config.Load()
	if err != nil {
		// log.Fatalf prints the error and exits the program with status code 1
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// STEP 2: Initialize Database Connection Pool
	dbPool, err := database.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// defer ensures this runs when main() exits, cleaning up resources
	defer dbPool.Close()

	// STEP 3: Initialize Application Layers (Dependency Injection)
	// Create repositories (data access layer)
	trickRepo := repository.NewTrickRepository(dbPool)
	videoRepo := repository.NewVideoRepository(dbPool)
	categoryRepo := repository.NewCategoryRepository(dbPool)
	userRepo := repository.NewUserRepository(dbPool)
	//comboRepo := repository.NewComboRepository(dbPool)

	// Create services (business logic layer)
	// Services receive repositories as dependencies
	trickService := services.NewTrickService(trickRepo, videoRepo)
	comboService := services.NewComboService(trickRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	userService := services.NewUserService(userRepo)
	// Create handlers (HTTP layer)
	// Handlers receive services as dependencies
	trickHandler := handlers.NewTrickHandler(trickService)
	comboHandler := handlers.NewComboHandler(comboService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	userHandler := handlers.NewUserHandler(userService)

	// STEP 4: Setup Router and Routes
	router := routes.NewRouter(cfg, trickHandler, comboHandler, categoryHandler, userHandler)

	// STEP 5: Create HTTP Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port, // e.g., ":8080"
		Handler: router,         // Our Gin router handles all requests
		// Timeouts prevent slow clients from holding connections indefinitely
		ReadTimeout:  15 * time.Second, // Max time to read request
		WriteTimeout: 15 * time.Second, // Max time to write response
		IdleTimeout:  60 * time.Second, // Max time for keep-alive connections
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		// ListenAndServe blocks until the server stops
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// STEP 7: Graceful Shutdown
	// We listen for interrupt signals (Ctrl+C) or termination signals (from Docker/K8s)
	quit := make(chan os.Signal, 1)
	// SIGINT = Ctrl+C, SIGTERM = kill command or container orchestrator
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until we receive a signal

	log.Println("Shutting down server...")

	// Create a deadline for shutdown - give requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
