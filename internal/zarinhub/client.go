package zarinhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	token      string
	httpClient *http.Client
}

// مدل درخواست ارسال شماره کارت
type CardToIbanRequest struct {
	Card string `json:"card"`
}

// مدل پاسخ دریافت شده از زرین‌هاب
type CardToIbanResponse struct {
	Data struct {
		Iban      string `json:"iban"`
		OwnerName string `json:"ownerName"`
		BankName  string `json:"bankName"`
	} `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewService(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) FetchIban(cardNumber string) (string, error) {
	url := "https://zarin-hub.com/api/v5/Kyc/CardToIban"

	reqBody, err := json.Marshal(CardToIbanRequest{Card: cardNumber})
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	// هدرها دقیقاً طبق مستندات زرین‌هاب
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var res CardToIbanResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %w", err)
	}

	// در صورتی که پاسخ خطای داخلی زرین‌هاب داشته باشد
	if res.Error.Code != "" {
		return "", fmt.Errorf("zarinhub error [%s]: %s", res.Error.Code, res.Error.Message)
	}

	return res.Data.Iban, nil
}
