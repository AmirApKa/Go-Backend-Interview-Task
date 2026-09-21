package zarinhub

import (
	"context"
	"time"
)

type Client interface {
	GetIbanByCard(ctx context.Context, cardNumber string) (string, error)
}

type Service struct {
	apiKey  string
	baseURL string
}

func NewService(apiKey, baseURL string) *Service {
	return &Service{apiKey: apiKey, baseURL: baseURL}
}

func (s *Service) GetIbanByCard(ctx context.Context, cardNumber string) (string, error) {
	_ = ctx
	time.Sleep(100 * time.Millisecond)
	return "IR120000000000000000000000", nil
}
