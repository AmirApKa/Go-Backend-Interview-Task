package handler

import (
	"encoding/json"
	"net/http"

	"card-to-iban/internal/repository"
	"card-to-iban/internal/service"
	"card-to-iban/internal/zarinhub"
)

type Request struct {
	CardNumber string `json:"card_number"`
}

type Response struct {
	Success bool   `json:"success"`
	Iban    string `json:"iban,omitempty"`
	Error   string `json:"error,omitempty"`
}

type CardHandler struct {
	zarinClient zarinhub.Client
	repo        repository.Repository
}

func NewCardHandler(zarinClient zarinhub.Client, repo repository.Repository) *CardHandler {
	return &CardHandler{
		zarinClient: zarinClient,
		repo:        repo,
	}
}

func (h *CardHandler) CardToIban(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "روش درخواست نامعتبر است"})
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "فرمت ورودی نامعتبر است"})
		return
	}

	if err := service.ValidateCard(req.CardNumber); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "شماره کارت وارد شده نامعتبر است"})
		return
	}

	iban, err := h.zarinClient.GetIbanByCard(r.Context(), req.CardNumber)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "خطا در استعلام شبا"})
		return
	}

	maskedCard := service.MaskCardNumber(req.CardNumber)
	_ = h.repo.SaveAuditLog(r.Context(), &repository.AuditLog{
		MaskedCard:   maskedCard,
		Status:       "SUCCESS",
		ResponseData: iban,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Iban:    iban,
	})
}
