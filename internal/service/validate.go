package service

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidCard = errors.New("invalid card number")

// NormalizeCardNumber تبدیل ارقام فارسی/عربی به انگلیسی و حذف فاصله
func NormalizeCardNumber(input string) string {
	var builder strings.Builder
	for _, r := range input {
		switch {
		case unicode.IsDigit(r):
			if r >= '۰' && r <= '۹' {
				builder.WriteRune(r - '۰' + '0')
			} else if r >= '٠' && r <= '٩' {
				builder.WriteRune(r - '٠' + '0')
			} else {
				builder.WriteRune(r)
			}
		}
	}
	return builder.String()
}

func ValidateCard(card string) error {
	card = NormalizeCardNumber(card)

	if card == "" {
		return ErrInvalidCard
	}

	if len(card) != 16 {
		return ErrInvalidCard
	}

	for _, r := range card {
		if r < '0' || r > '9' {
			return ErrInvalidCard
		}
	}

	if !luhnValid(card) {
		return ErrInvalidCard
	}

	return nil
}

func luhnValid(card string) bool {
	sum := 0
	double := false

	for i := len(card) - 1; i >= 0; i-- {
		d := int(card[i] - '0')

		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
		double = !double
	}

	return sum%10 == 0
}

func MaskCardNumber(card string) string {
	card = NormalizeCardNumber(card)
	if len(card) != 16 {
		return card
	}
	return card[:6] + "******" + card[12:]
}
