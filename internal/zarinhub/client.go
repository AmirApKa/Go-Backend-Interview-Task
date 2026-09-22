package zarinhub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const authURL = "https://zarin-hub.com/api/v5/Authentication/GetToken"
const cardToIbanURL = "https://zarin-hub.com/api/v5/Kyc/CardToIban"

type Client struct {
	username   string
	password   string
	httpClient *http.Client

	mu           sync.Mutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

type apiMeta struct {
	TrackID      string   `json:"trackId"`
	Status       string   `json:"status"`
	IsSuccess    bool     `json:"isSuccess"`
	Code         int      `json:"code"`
	Message      string   `json:"message"`
	ErrorMessage *string  `json:"errorMessage"`
	ErrorType    *string  `json:"errorType"`
	Errors       []string `json:"errors"`
}

type getTokenRequest struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type getTokenResponse struct {
	Meta apiMeta `json:"meta"`
	Data struct {
		AccessToken  string `json:"accessToken"`
		TokenType    string `json:"tokenType"`
		ExpiresTn    int    `json:"expiresTn"`
		RefreshToken string `json:"refreshToken"`
	} `json:"data"`
}

type CardToIbanRequest struct {
	Card string `json:"card"`
}

type CardToIbanResponse struct {
	Meta apiMeta `json:"meta"`
	Data struct {
		Iban      string `json:"iban"`
		OwnerName string `json:"ownerName"`
		BankName  string `json:"bankName"`
	} `json:"data"`
}

func NewService(username, password string) *Client {
	return &Client{
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt.Add(-1*time.Minute)) {
		return nil
	}

	var reqBody getTokenRequest
	if c.refreshToken != "" {
		reqBody = getTokenRequest{RefreshToken: c.refreshToken}
	} else {
		reqBody = getTokenRequest{Username: c.username, Password: c.password}
	}

	resp, err := c.callAuth(ctx, reqBody)
	if err != nil && c.refreshToken != "" {
		c.refreshToken = ""
		reqBody = getTokenRequest{Username: c.username, Password: c.password}
		resp, err = c.callAuth(ctx, reqBody)
	}
	if err != nil {
		return err
	}

	c.accessToken = resp.Data.AccessToken
	c.refreshToken = resp.Data.RefreshToken
	c.expiresAt = time.Now().Add(time.Duration(resp.Data.ExpiresTn) * time.Second)
	return nil
}

func (c *Client) callAuth(ctx context.Context, body getTokenRequest) (*getTokenResponse, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal auth request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create auth request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, classifyTransportError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{Category: CategoryServerError, Message: fmt.Sprintf("read auth response failed: %v", err)}
	}

	var res getTokenResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, &Error{Category: CategoryServerError, Message: fmt.Sprintf("unmarshal auth response failed (status %d): %v", resp.StatusCode, err)}
	}

	if !res.Meta.IsSuccess {
		category := CategoryAuth
		if res.Meta.ErrorType != nil {
			category = remoteErrorTypeToCategory(*res.Meta.ErrorType)
		}
		return nil, &Error{
			Category:   category,
			RemoteCode: fmt.Sprintf("%d", res.Meta.Code),
			Message:    res.Meta.Message,
		}
	}

	return &res, nil
}

// classifyTransportError خطاهای سطح شبکه (قبل از دریافت هرگونه پاسخ HTTP) را
// به Timeout یا Unavailable تفکیک می‌کند.
func classifyTransportError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return &Error{Category: CategoryTimeout, Message: err.Error()}
	}
	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) && netErr.Timeout() {
		return &Error{Category: CategoryTimeout, Message: err.Error()}
	}
	return &Error{Category: CategoryUnavailable, Message: err.Error()}
}

// FetchIban شماره کارت را به زرین‌هاب می‌فرستد.
// httpStatus و rawBody در صورت دریافت پاسخ برگردانده می‌شوند تا در Audit Log ثبت شوند.
// خطای برگشتی همیشه از نوع *Error است تا Handler بتواند دسته‌بندی آن را بخواند.
func (c *Client) FetchIban(ctx context.Context, cardNumber string) (iban string, httpStatus int, rawBody string, err error) {
	if tokenErr := c.ensureToken(ctx); tokenErr != nil {
		return "", 0, "", tokenErr
	}

	reqBody, err := json.Marshal(CardToIbanRequest{Card: cardNumber})
	if err != nil {
		return "", 0, "", &Error{Category: CategoryServerError, Message: fmt.Sprintf("marshal request failed: %v", err)}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cardToIbanURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", 0, "", &Error{Category: CategoryServerError, Message: fmt.Sprintf("create request failed: %v", err)}
	}

	c.mu.Lock()
	token := c.accessToken
	c.mu.Unlock()

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, "", classifyTransportError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, "", &Error{Category: CategoryServerError, Message: fmt.Sprintf("read response body failed: %v", err)}
	}
	rawBody = string(bodyBytes)
	httpStatus = resp.StatusCode

	var res CardToIbanResponse
	if unmarshalErr := json.Unmarshal(bodyBytes, &res); unmarshalErr != nil {
		category := CategoryServerError
		if httpStatus == http.StatusNotFound {
			category = CategoryUnavailable
		}
		return "", httpStatus, rawBody, &Error{Category: category, Message: fmt.Sprintf("unexpected non-JSON response (status %d)", httpStatus)}
	}

	if !res.Meta.IsSuccess {
		category := CategoryRejected
		if res.Meta.ErrorType != nil {
			category = remoteErrorTypeToCategory(*res.Meta.ErrorType)
		}
		return "", httpStatus, rawBody, &Error{
			Category:   category,
			RemoteCode: fmt.Sprintf("%d", res.Meta.Code),
			Message:    res.Meta.Message,
		}
	}

	return res.Data.Iban, httpStatus, rawBody, nil
}
