package service

import "testing"

func TestValidateCard(t *testing.T) {
	tests := []struct {
		name    string
		card    string
		wantErr bool
	}{
		{
			name:    "Valid Card Number",
			card:    "6219861497401835",
			wantErr: false,
		},
		{
			name:    "Empty Card Number",
			card:    "",
			wantErr: true,
		},
		{
			name:    "Short Card Number",
			card:    "123456",
			wantErr: true,
		},
		{
			name:    "Invalid Luhn",
			card:    "6037997381112222",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCard(tt.card)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
