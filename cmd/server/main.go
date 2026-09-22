package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"card-to-iban/internal/handler"
	"card-to-iban/internal/repository"
	"card-to-iban/internal/zarinhub"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, relying on OS environment variables")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	username := os.Getenv("ZARINHUB_USERNAME")
	password := os.Getenv("ZARINHUB_PASSWORD")
	token := os.Getenv("ZARINHUB_TOKEN")

	if token == "" && (username == "" || password == "") {
		logger.Warn("Neither ZARINHUB_TOKEN nor (ZARINHUB_USERNAME & ZARINHUB_PASSWORD) is set")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}

	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	zarinClient := zarinhub.NewService(username, password)
	if token != "" {
		zarinClient.SetStaticToken(token)
	}

	cardHandler := handler.NewCardHandler(zarinClient, repo, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/card-to-iban", cardHandler.CardToIban)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	} else {
		logger.Info("server stopped gracefully")
	}
}
