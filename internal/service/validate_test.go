package service

import (
	"errors"
	"testing"
)

func TestValidateCard(t *testing.T) {
	tests := []struct {
		name    string
		card    string
		wantErr error
	}{
		{
			name:    "Valid Card Number",
			card:    "6219861497401835",
			wantErr: nil,
		},
		{
			name:    "Valid Card with Persian Digits and Spaces",
			card:    "۶۲۱۹ ۸۶۱۴ ۹۷۴۰ ۱۸۳۵",
			wantErr: nil,
		},
		{
			name:    "Empty Card Number",
			card:    "",
			wantErr: ErrInvalidCardFormat,
		},
		{
			name:    "Short Card Number",
			card:    "123456",
			wantErr: ErrInvalidCardFormat,
		},
		{
			name:    "Long Card Number",
			card:    "6219861497401835123",
			wantErr: ErrInvalidCardFormat,
		},
		{
			name:    "Card with Letters",
			card:    "62198614ABCD1835",
			wantErr: ErrInvalidCardDigits,
		},
		{
			name:    "Invalid Luhn",
			card:    "6037997381112222",
			wantErr: ErrInvalidCardLuhn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCard(tt.card)
			if tt.wantErr == nil && err != nil {
				t.Errorf("ValidateCard() unexpected error = %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateCard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMaskCardNumber(t *testing.T) {
	tests := []struct {
		name     string
		card     string
		expected string
	}{
		{
			name:     "Standard 16-digit Card",
			card:     "6037997599999999",
			expected: "603799******9999",
		},
		{
			name:     "Card with Persian Digits and Spaces",
			card:     "۶۰۳۷ ۹۹۷۵ ۹۹۹۹ ۹۹۹۹",
			expected: "603799******9999",
		},
		{
			name:     "Invalid Short Card Number (6 digits)",
			card:     "123456",
			expected: "**3456",
		},
		{
			name:     "Very Short Card Number (3 digits)",
			card:     "123",
			expected: "***",
		},
		{
			name:     "Empty Card Number",
			card:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskCardNumber(tt.card)
			if got != tt.expected {
				t.Errorf("MaskCardNumber() = %v, want %v", got, tt.expected)
			}
		})
	}
}
