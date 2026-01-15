package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonx8/chat-service/internal/config"
	"github.com/jonx8/chat-service/internal/database"
)

func main() {
	cfg := config.Load()

	slog.Info(fmt.Sprintf("Starting %s:%s...", cfg.AppName, cfg.AppVersion))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.New(ctx, cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	defer func() {
		slog.Info("Closing database connection...")
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		} else {
			slog.Info("Database connection closed")
		}
	}()

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("Starting HTTP server...", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		slog.Error("Server error", "error", err)

	case sig := <-quit:
		slog.Info("Received shutdown signal", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		slog.Info("Shutting down server gracefully...")
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Server forced to shutdown", "error", err)
		}
	}

	slog.Info("Server stopped")
}
