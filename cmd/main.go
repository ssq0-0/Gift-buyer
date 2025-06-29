// Package main provides the entry point for the Gift Buyer application.
// Gift Buyer is an automated system for monitoring and purchasing Telegram Star Gifts
// based on configurable criteria such as price range, supply limits, and total star cap.
//
// The application connects to Telegram API, monitors available gifts, validates them
// against user-defined criteria, and automatically purchases eligible gifts.
//
// Usage:
//
//	go run cmd/main.go
//
// Configuration is loaded from internal/config/config.json file.
package main

import (
	"context"
	"gift-buyer/internal/config"
	"gift-buyer/internal/usecase"
	"gift-buyer/pkg/logger"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

// main is the entry point of the Gift Buyer application.
// It initializes the logger, loads configuration, creates the gift service,
// and handles graceful shutdown on system signals.
func main() {
	logger.Init("debug")

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	configPath := filepath.Join(basepath, "..", "internal", "config", "config.json")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.GlobalLogger.Fatalf("Failed to load config: %v", err)
	}

	logLevel := logger.ParseLevel(cfg.LoggerLevel)
	logger.Init(logLevel)

	service, err := usecase.NewFactory(&cfg.SoftConfig).CreateSystem()
	if err != nil {
		logger.GlobalLogger.Fatalf("Failed to init telegram client: %v", err)
	}

	if err = service.SetIds(context.Background()); err != nil {
		logger.GlobalLogger.Fatalf("Failed to set IDs: %v", err)
	}

	go func() {
		logger.GlobalLogger.Info("Starting gift service...")
		service.Start()
	}()

	go func() {
		logger.GlobalLogger.Info("Starting update checker...")
		service.CheckForUpdates()
	}()

	logger.GlobalLogger.Info("Gift buyer service started. Press Ctrl+C to stop.")
	gracefulShutdown(service)
	logger.GlobalLogger.Info("Application terminated")
}

// gracefulShutdown handles the graceful shutdown of the gift service.
// It listens for SIGINT and SIGTERM signals and provides a 30-second timeout
// for the service to stop gracefully before forcing termination.
//
// Parameters:
//   - service: The GiftService instance to be stopped gracefully
func gracefulShutdown(service usecase.UseCase) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.GlobalLogger.Info("Received shutdown signal, stopping service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		service.Stop()
		close(done)
	}()

	select {
	case <-done:
		logger.GlobalLogger.Info("Service stopped gracefully")
	case <-shutdownCtx.Done():
		logger.GlobalLogger.Warn("Shutdown timeout exceeded, forcing exit")
	}
}
