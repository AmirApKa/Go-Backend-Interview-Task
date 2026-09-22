package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

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
	if username == "" || password == "" {
		logger.Warn("ZARINHUB_USERNAME or ZARINHUB_PASSWORD is not set")
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
	cardHandler := handler.NewCardHandler(zarinClient, repo, logger)

	http.HandleFunc("/api/v1/card-to-iban", cardHandler.CardToIban)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("server starting", "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
