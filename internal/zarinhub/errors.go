package zarinhub

import (
	"context"
	"errors"
	"fmt"
)

type ErrorCategory string

const (
	CategoryAuth        ErrorCategory = "AUTH"
	CategoryRejected    ErrorCategory = "REJECTED"
	CategoryTimeout     ErrorCategory = "TIMEOUT"
	CategoryUnavailable ErrorCategory = "UNAVAILABLE"
	CategoryServerError ErrorCategory = "SERVER_ERROR"
)

type Error struct {
	Category   ErrorCategory
	RemoteCode string
	Message    string
}

func (e *Error) Error() string {
	if e.RemoteCode != "" {
		return fmt.Sprintf("[%s] %s (code: %s)", e.Category, e.Message, e.RemoteCode)
	}
	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

func remoteErrorTypeToCategory(errType string) ErrorCategory {
	switch errType {
	case "Unauthorized", "AuthenticationFailed":
		return CategoryAuth
	case "InvalidInput", "Rejected", "NotFound":
		return CategoryRejected
	case "Timeout":
		return CategoryTimeout
	case "ServiceUnavailable":
		return CategoryUnavailable
	default:
		return CategoryServerError
	}
}

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
