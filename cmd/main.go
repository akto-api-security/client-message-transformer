package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"client-message-transformer/internal/config"
	"client-message-transformer/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create service
	svc, err := service.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	// Start Kafka transformer service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = svc.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Received shutdown signal...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	err = svc.Stop(shutdownCtx)
	if err != nil {
		log.Fatalf("Error during shutdown: %v", err)
	}

	log.Println("🎉 Kafka Transformer Service stopped successfully")
}
