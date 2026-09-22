package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"card-to-iban/internal/handler"
	"card-to-iban/internal/repository"
)

type mockZarinHubClient struct {
	shouldFail bool
}

func (m *mockZarinHubClient) FetchIban(ctx context.Context, cardNumber string) (string, int, string, error) {
	if m.shouldFail {
		return "", http.StatusBadRequest, `{"meta":{"isSuccess":false}}`, nil
	}
	return "IR120170000000123456789001", http.StatusOK, `{"meta":{"isSuccess":true}}`, nil
}

func TestCardToIbanHandler_Success(t *testing.T) {
	repo, err := repository.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory db: %v", err)
	}
	defer repo.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockClient := &mockZarinHubClient{shouldFail: false}

	cardHandler := handler.NewCardHandler(mockClient, repo, logger)

	// ارسال درخواست با ساختار استاندارد JSON
	reqBody := []byte(`{"card": "6219861497401835", "card_number": "6219861497401835", "cardNumber": "6219861497401835"}`)
	req := httptest.NewRequest("POST", "/api/v1/card-to-iban", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	cardHandler.CardToIban(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d. Response Body: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("expected response success to be true")
	}
}

func TestCardToIbanHandler_InvalidCard(t *testing.T) {
	repo, _ := repository.NewSQLiteRepository(":memory:")
	defer repo.Close()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mockClient := &mockZarinHubClient{shouldFail: false}

	cardHandler := handler.NewCardHandler(mockClient, repo, logger)

	reqBody := []byte(`{"card": "1234"}`)
	req := httptest.NewRequest("POST", "/api/v1/card-to-iban", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	cardHandler.CardToIban(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d for invalid card, got %d", http.StatusBadRequest, recorder.Code)
	}
}
