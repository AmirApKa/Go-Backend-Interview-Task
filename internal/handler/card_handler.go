package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

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
	Code    string `json:"code,omitempty"`
}

type CardHandler struct {
	zarinClient *zarinhub.Client
	repo        repository.Repository
	logger      *slog.Logger
}

func NewCardHandler(zarinClient *zarinhub.Client, repo repository.Repository, logger *slog.Logger) *CardHandler {
	return &CardHandler{zarinClient: zarinClient, repo: repo, logger: logger}
}

func generateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type errorMapping struct {
	httpStatus int
	appCode    string
	message    string
}

var categoryMappings = map[zarinhub.ErrorCategory]errorMapping{
	zarinhub.CategoryAuth:        {http.StatusBadGateway, "EXTERNAL_AUTH_ERROR", "خطا در احراز هویت با سرویس بیرونی"},
	zarinhub.CategoryRejected:    {http.StatusUnprocessableEntity, "EXTERNAL_REJECTED", "درخواست توسط سرویس بیرونی رد شد"},
	zarinhub.CategoryTimeout:     {http.StatusGatewayTimeout, "EXTERNAL_TIMEOUT", "زمان پاسخ‌گویی سرویس بیرونی به پایان رسید"},
	zarinhub.CategoryUnavailable: {http.StatusServiceUnavailable, "EXTERNAL_UNAVAILABLE", "سرویس بیرونی در حال حاضر در دسترس نیست"},
	zarinhub.CategoryServerError: {http.StatusBadGateway, "EXTERNAL_SERVER_ERROR", "خطای داخلی در سرویس بیرونی"},
}

func mapZarinhubError(err error) errorMapping {
	var zErr *zarinhub.Error
	if errors.As(err, &zErr) {
		if m, ok := categoryMappings[zErr.Category]; ok {
			return m
		}
	}
	return errorMapping{http.StatusBadGateway, "EXTERNAL_SERVER_ERROR", "خطا در استعلام شبا"}
}

func (h *CardHandler) CardToIban(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	requestID := generateRequestID()
	ip := clientIP(r)
	start := time.Now()
	log := h.logger.With("request_id", requestID, "client_ip", ip)

	if r.Method != http.MethodPost {
		log.Warn("method not allowed", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "روش درخواست نامعتبر است", Code: "METHOD_NOT_ALLOWED"})
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("invalid request body")
		h.audit(r.Context(), log, requestID, "", "VALIDATION_ERROR", ip, start, nil, nil, "فرمت ورودی نامعتبر است")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "فرمت ورودی نامعتبر است", Code: "VALIDATION_ERROR"})
		return
	}

	cleanCard := service.NormalizeCardNumber(req.CardNumber)
	maskedCard := service.MaskCardNumber(cleanCard)
	log = log.With("masked_card", maskedCard)

	if err := service.ValidateCard(cleanCard); err != nil {
		log.Warn("card validation failed")
		h.audit(r.Context(), log, requestID, maskedCard, "VALIDATION_ERROR", ip, start, nil, nil, "شماره کارت وارد شده نامعتبر است")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Error: "شماره کارت وارد شده نامعتبر است", Code: "VALIDATION_ERROR"})
		return
	}

	iban, extStatus, rawBody, err := h.zarinClient.FetchIban(r.Context(), cleanCard)
	if err != nil {
		m := mapZarinhubError(err)
		log.Error("zarinhub call failed", "reason", err.Error(), "external_http_status", extStatus)
		h.audit(r.Context(), log, requestID, maskedCard, m.appCode, ip, start, &extStatus, &rawBody, m.message)
		w.WriteHeader(m.httpStatus)
		json.NewEncoder(w).Encode(Response{Success: false, Error: m.message, Code: m.appCode})
		return
	}

	log.Info("card-to-iban succeeded", "external_http_status", extStatus)
	h.audit(r.Context(), log, requestID, maskedCard, "SUCCESS", ip, start, &extStatus, &rawBody, "")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Success: true, Iban: iban})
}

func (h *CardHandler) audit(ctx context.Context, log *slog.Logger, requestID, maskedCard, status, ip string, start time.Time, extStatus *int, rawBody *string, errMsg string) {
	entry := &repository.AuditLog{
		RequestID:  requestID,
		MaskedCard: maskedCard,
		Status:     status,
		ClientIP:   ip,
		DurationMs: time.Since(start).Milliseconds(),
	}
	if extStatus != nil {
		entry.ExternalHTTPStatus = sql.NullInt64{Int64: int64(*extStatus), Valid: true}
	}
	if rawBody != nil {
		entry.ExternalResponse = sql.NullString{String: *rawBody, Valid: true}
	}
	if errMsg != "" {
		entry.ErrorMessage = sql.NullString{String: errMsg, Valid: true}
	}
	if err := h.repo.SaveAuditLog(ctx, entry); err != nil {
		log.Error("failed to save audit log", "error", err)
	}
}
