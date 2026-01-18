package errors

import (
	"fmt"
	"log/slog"
)

// KV represents a key-value pair for structured context
type KV struct {
	Key   string
	Value any
}

// VerdisError is the base error type for all Verdis errors.
// It provides Redis-compatible error codes, structured context for logging,
// error wrapping, and transient/non-transient classification.
type VerdisError struct {
	Code      Code   // Redis-style error code prefix
	Message   string // Human-readable error message
	Cause     error  // Wrapped underlying error (for error chains)
	Context   []KV   // Structured key-value context for logging
	Transient bool   // True if this error is temporary and retryable
}

// Error implements the error interface
func (e *VerdisError) Error() string {
	if e.Code == "" {
		return e.Message
	}
	return fmt.Sprintf("%s %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for error chain traversal
func (e *VerdisError) Unwrap() error {
	return e.Cause
}

// IsTransient returns true if this error is temporary and the operation can be retried
func (e *VerdisError) IsTransient() bool {
	return e.Transient
}

// LogValue implements slog.LogValuer for structured logging
func (e *VerdisError) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 3+len(e.Context))

	attrs = append(attrs,
		slog.String("code", string(e.Code)),
		slog.String("message", e.Message),
		slog.Bool("transient", e.Transient),
	)

	for _, kv := range e.Context {
		attrs = append(attrs, slog.Any(kv.Key, kv.Value))
	}

	if e.Cause != nil {
		attrs = append(attrs, slog.String("cause", e.Cause.Error()))
	}

	return slog.GroupValue(attrs...)
}

// With returns a copy of the error with additional context
func (e *VerdisError) With(key string, value any) *VerdisError {
	newErr := *e
	newErr.Context = append(newErr.Context, KV{Key: key, Value: value})
	return &newErr
}

// Wrap returns a copy of the error with the given cause wrapped
func (e *VerdisError) Wrap(cause error) *VerdisError {
	newErr := *e
	newErr.Cause = cause
	return &newErr
}

// RESPError returns the RESP-formatted error string (without the - prefix)
func (e *VerdisError) RESPError() string {
	return e.Error()
}

// TransientError is the interface for errors that know if they're transient
type TransientError interface {
	error
	IsTransient() bool
}

// New creates a new VerdisError with the given code and message
func New(code Code, message string) *VerdisError {
	return &VerdisError{
		Code:      code,
		Message:   message,
		Transient: code.IsTransientCode(),
	}
}

// Newf creates a new VerdisError with a formatted message
func Newf(code Code, format string, args ...any) *VerdisError {
	return &VerdisError{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Transient: code.IsTransientCode(),
	}
}

// NewTransient creates a new transient (retryable) error
func NewTransient(code Code, message string) *VerdisError {
	return &VerdisError{
		Code:      code,
		Message:   message,
		Transient: true,
	}
}

// Wrapf wraps an error with a new VerdisError
func Wrapf(cause error, code Code, format string, args ...any) *VerdisError {
	return &VerdisError{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Cause:     cause,
		Transient: code.IsTransientCode(),
	}
}
