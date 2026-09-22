package zarinhub

import "fmt"

// ErrorCategory یک دسته‌بندی پایدار و در سطح Application برای خطاهای فراخوانی زرین‌هاب است.
// این دسته‌ها برای تفکیک خطای احراز هویت، رد شدن درخواست، Timeout/عدم دسترسی
// و خطای داخلی سرویس بیرونی استفاده می‌شوند.
type ErrorCategory string

const (
	CategoryAuth        ErrorCategory = "EXTERNAL_AUTH_ERROR"
	CategoryRejected    ErrorCategory = "EXTERNAL_REJECTED"
	CategoryTimeout     ErrorCategory = "EXTERNAL_TIMEOUT"
	CategoryUnavailable ErrorCategory = "EXTERNAL_UNAVAILABLE"
	CategoryServerError ErrorCategory = "EXTERNAL_SERVER_ERROR"
)

// Error خطای ساختاریافته‌ای است که کلاینت زرین‌هاب برمی‌گرداند.
// Message فقط برای Log داخلی است و هرگز نباید مستقیماً به کاربر نمایش داده شود.
type Error struct {
	Category   ErrorCategory
	RemoteCode string
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("zarinhub error [%s/%s]: %s", e.Category, e.RemoteCode, e.Message)
}

// remoteErrorTypeToCategory errorType مستندشده‌ی زرین‌هاب را
// به دسته‌بندی داخلی ما نگاشت می‌کند.
func remoteErrorTypeToCategory(errorType string) ErrorCategory {
	switch errorType {
	case "UnAuthorized":
		return CategoryAuth
	case "RequestTimeOut":
		return CategoryTimeout
	case "Unavailable":
		return CategoryUnavailable
	case "ServerError", "ProviderError":
		return CategoryServerError
	case "BadRequest", "NotFound", "LogicError":
		return CategoryRejected
	default:
		return CategoryServerError
	}
}
