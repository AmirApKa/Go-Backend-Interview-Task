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
	zarinClient *zarinhub.Client
	repo        repository.Repository
}

func NewCardHandler(zarinClient *zarinhub.Client, repo repository.Repository) *CardHandler {
	return &CardHandler{
		zarinClient: zarinClient,
		repo:        repo,
	}
}

func (h *CardHandler) CardToIban(w http.ResponseWriter, r *http.Request) {
	// ۱. تنظیم هدرهای CORS جهت اجازه ارتباط فرانت‌اند
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// ۲. پاسخ فوری به درخواست Preflight مرورگر
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// ۳. بررسی متد HTTP
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

	// نرمال‌سازی شماره کارت (تبدیل ارقام فارسی/عربی و حذف فاصله)
	cleanCard := service.NormalizeCardNumber(req.CardNumber)

	if err := service.ValidateCard(cleanCard); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "شماره کارت وارد شده نامعتبر است"})
		return
	}

	// ۴. فراخوانی متد استعلام شبا با شماره کارت تمیز شده
	iban, err := h.zarinClient.FetchIban(cleanCard)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "خطا در استعلام شبا"})
		return
	}

	maskedCard := service.MaskCardNumber(cleanCard)
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
