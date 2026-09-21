package main

import (
	"log"
	"net/http"

	"card-to-iban/internal/handler"
	"card-to-iban/internal/repository"
	"card-to-iban/internal/zarinhub"
)

func main() {
	zarinClient := zarinhub.NewService("fake_api_key", "https://hub-zarin.com")
	repo := repository.NewMemoryRepository()
	cardHandler := handler.NewCardHandler(zarinClient, repo)

	http.HandleFunc("/api/v1/card-to-iban", cardHandler.CardToIban)

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
