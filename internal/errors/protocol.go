package errors

import "fmt"

// Protocol/parsing layer errors (mixed transient/non-transient)

// ErrSyntax creates a syntax error (non-transient)
func ErrSyntax(detail string) *VerdisError {
	return &VerdisError{
		Code:      SYNTAX,
		Message:   fmt.Sprintf("syntax error: %s", detail),
		Transient: false,
		Context:   []KV{{Key: "detail", Value: detail}},
	}
}

// ErrProtocol creates a protocol error for malformed RESP data (non-transient)
func ErrProtocol(detail string) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   fmt.Sprintf("protocol error: %s", detail),
		Transient: false,
		Context:   []KV{{Key: "detail", Value: detail}},
	}
}

// ErrReadTimeout creates a read timeout error (transient)
func ErrReadTimeout(cause error) *VerdisError {
	return &VerdisError{
		Code:      TIMEOUT,
		Message:   "read timeout",
		Cause:     cause,
		Transient: true,
	}
}

// ErrWriteTimeout creates a write timeout error (transient)
func ErrWriteTimeout(cause error) *VerdisError {
	return &VerdisError{
		Code:      TIMEOUT,
		Message:   "write timeout",
		Cause:     cause,
		Transient: true,
	}
}

// ErrConnectionReset creates a connection reset error (transient)
func ErrConnectionReset(cause error) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   "connection reset by peer",
		Cause:     cause,
		Transient: true,
	}
}

// ErrParseFailed creates a parse error with context
func ErrParseFailed(what string, cause error) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   fmt.Sprintf("failed to parse %s", what),
		Cause:     cause,
		Transient: false,
		Context:   []KV{{Key: "parsing", Value: what}},
	}
}

// ErrInvalidCommand creates an error for invalid command format
func ErrInvalidCommand(reason string) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   fmt.Sprintf("invalid command: %s", reason),
		Transient: false,
		Context:   []KV{{Key: "reason", Value: reason}},
	}
}

// ErrEmptyCommand creates an error for empty commands
func ErrEmptyCommand() *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   "empty command",
		Transient: false,
	}
}
