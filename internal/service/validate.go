package service

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrInvalidCardFormat = errors.New("شماره کارت باید دقیقاً ۱۶ رقم باشد")
	ErrInvalidCardDigits = errors.New("شماره کارت حاوی کاراکترهای نامعتبر است")
	ErrInvalidCardLuhn   = errors.New("شماره کارت وارد شده معتبر نیست")
)

// NormalizeCardNumber تبدیل ارقام فارسی/عربی به انگلیسی و حذف فاصله و خط تیره
func NormalizeCardNumber(input string) string {
	var builder strings.Builder
	for _, r := range input {
		if r == ' ' || r == '-' {
			continue
		}
		switch {
		case r >= '۰' && r <= '۹':
			builder.WriteRune(r - '۰' + '0')
		case r >= '٠' && r <= '٩':
			builder.WriteRune(r - '٠' + '0')
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// ValidateCard اعتبارسنجی دقیق شماره کارت
func ValidateCard(card string) error {
	normalized := NormalizeCardNumber(card)

	if len(normalized) != 16 {
		return ErrInvalidCardFormat
	}

	for _, r := range normalized {
		if !unicode.IsDigit(r) {
			return ErrInvalidCardDigits
		}
	}

	if !luhnValid(normalized) {
		return ErrInvalidCardLuhn
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

// MaskCardNumber ماسک کردن ایمن شماره کارت.
// برای طول ۱۶: ۶ رقم اول + ۶ ستاره + ۴ رقم آخر.
// برای سایر طول‌ها: همه ارقام به‌جز ۴ رقم آخر ستاره می‌شوند تا از افشای
// تصادفی شماره کارت در audit log جلوگیری شود.
func MaskCardNumber(card string) string {
	normalized := NormalizeCardNumber(card)
	n := len(normalized)

	if n == 0 {
		return ""
	}
	if n <= 4 {
		return strings.Repeat("*", n)
	}
	if n == 16 {
		return normalized[:6] + "******" + normalized[12:]
	}
	// طول غیراستاندارد: فقط ۴ رقم آخر نمایش داده شود
	return strings.Repeat("*", n-4) + normalized[n-4:]
}
