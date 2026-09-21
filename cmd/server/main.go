package main

import (
	"log"
	"net/http"
	"os"

	"card-to-iban/internal/handler"
	"card-to-iban/internal/repository"
	"card-to-iban/internal/zarinhub"
)

func main() {
	// خواندن توکن از متغیر محیطی
	token := os.Getenv("ZARINHUB_TOKEN")
	if token == "" {
		log.Println("Warning: ZARINHUB_TOKEN is not set")
	}

	zarinClient := zarinhub.NewService(token)
	repo := repository.NewMemoryRepository()
	cardHandler := handler.NewCardHandler(zarinClient, repo)

	http.HandleFunc("/api/v1/card-to-iban", cardHandler.CardToIban)

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
