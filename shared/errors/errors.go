package errors

import (
	"errors"
	"fmt"
)

// ErrorCode is a machine-readable error identifier.
type ErrorCode string

const (
	ErrNotFound             ErrorCode = "NOT_FOUND"
	ErrAlreadyExists        ErrorCode = "ALREADY_EXISTS"
	ErrValidation           ErrorCode = "VALIDATION_ERROR"
	ErrAuthentication       ErrorCode = "AUTHENTICATION_FAILED"
	ErrAuthorization        ErrorCode = "AUTHORIZATION_DENIED"
	ErrRateLimited          ErrorCode = "RATE_LIMITED"
	ErrCircularChain        ErrorCode = "CIRCULAR_CHAIN"
	ErrMaxDepthExceeded     ErrorCode = "MAX_DEPTH_EXCEEDED"
	ErrGuardrailBlocked     ErrorCode = "GUARDRAIL_BLOCKED"
	ErrCapabilityUnhealthy  ErrorCode = "CAPABILITY_UNHEALTHY"
	ErrCapabilityNotFound   ErrorCode = "CAPABILITY_NOT_FOUND"
	ErrManifestInvalid      ErrorCode = "MANIFEST_INVALID"
	ErrManifestParseFailed  ErrorCode = "MANIFEST_PARSE_FAILED"
	ErrTransportError       ErrorCode = "TRANSPORT_ERROR"
	ErrInvocationFailed     ErrorCode = "INVOCATION_FAILED"
	ErrTimeout              ErrorCode = "TIMEOUT"
	ErrInternal             ErrorCode = "INTERNAL_ERROR"
)

// MapToHTTPStatus converts an ErrorCode to its canonical HTTP status code.
// Single source of truth — used by the Wrap middleware in both CP and proxy.
func MapToHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrNotFound, ErrCapabilityNotFound:
		return 404
	case ErrAlreadyExists, ErrCircularChain:
		return 409
	case ErrValidation, ErrManifestInvalid, ErrManifestParseFailed:
		return 400
	case ErrAuthentication:
		return 401
	case ErrAuthorization:
		return 403
	case ErrRateLimited, ErrMaxDepthExceeded:
		return 429
	case ErrGuardrailBlocked:
		return 422
	case ErrCapabilityUnhealthy:
		return 502
	case ErrTimeout:
		return 504
	default:
		return 500
	}
}

// AppError is a structured domain error carrying a machine-readable code,
// a human-readable message, optional detail, and an optional cause.
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Detail  string    `json:"detail,omitempty"`
	Cause   error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

// New creates an AppError with the given code and message.
func New(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap creates an AppError wrapping an existing error.
func Wrap(code ErrorCode, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

// WithDetail creates an AppError with an additional detail string.
func WithDetail(code ErrorCode, message, detail string) *AppError {
	return &AppError{Code: code, Message: message, Detail: detail}
}

// NotFound returns a NOT_FOUND error for a named entity.
func NotFound(entity, id string) *AppError {
	return New(ErrCapabilityNotFound, fmt.Sprintf("%s '%s' not found", entity, id))
}

// Forbidden returns an AUTHORIZATION_DENIED error.
func Forbidden(reason string) *AppError {
	return New(ErrAuthorization, reason)
}

// Internal returns an INTERNAL_ERROR wrapping an unexpected error.
func Internal(msg string, cause error) *AppError {
	return Wrap(ErrInternal, msg, cause)
}

// CircularChain returns a CIRCULAR_CHAIN error with the chain detail.
func CircularChain(chain []string) *AppError {
	detail := ""
	for i, id := range chain {
		if i > 0 {
			detail += " -> "
		}
		detail += id
	}
	return WithDetail(ErrCircularChain, "circular agent chain detected", detail)
}

// IsCode returns true if err is an AppError with the given code.
func IsCode(err error, code ErrorCode) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code == code
	}
	return false
}

// AsAppError unwraps any error to AppError, wrapping unknown errors as Internal.
func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return Internal("unexpected error", err)
}
