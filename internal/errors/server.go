package errors

import (
	"fmt"
	"net"
)

// Server-layer errors (mixed transient/non-transient)

// ErrAddressInUse creates an address already in use error (transient during startup)
func ErrAddressInUse(addr string, cause error) *VerdisError {
	port := extractPort(addr)
	return &VerdisError{
		Code: ERR,
		Message: fmt.Sprintf(
			"cannot start server: address %s is already in use\n"+
				"  Possible solutions:\n"+
				"  - Stop the existing process using this port (lsof -i :%s | grep LISTEN)\n"+
				"  - Use a different port with WithPort() configuration\n"+
				"  - Wait for the port to be released (TIME_WAIT state)",
			addr, port,
		),
		Cause:     cause,
		Transient: true, // Can be retried after waiting
		Context: []KV{
			{Key: "address", Value: addr},
			{Key: "port", Value: port},
		},
	}
}

// ErrServerBusy creates a server busy error (transient)
func ErrServerBusy(reason string) *VerdisError {
	return &VerdisError{
		Code:      BUSY,
		Message:   fmt.Sprintf("server is busy: %s", reason),
		Transient: true,
		Context:   []KV{{Key: "reason", Value: reason}},
	}
}

// ErrMaxConnections creates a max connections reached error (transient)
func ErrMaxConnections(current, max int) *VerdisError {
	return &VerdisError{
		Code:      BUSY,
		Message:   "max connections reached",
		Transient: true,
		Context: []KV{
			{Key: "current", Value: current},
			{Key: "max", Value: max},
		},
	}
}

// ErrInternal creates an internal server error (non-transient)
func ErrInternal(detail string) *VerdisError {
	return &VerdisError{
		Code:      INTERNAL,
		Message:   fmt.Sprintf("internal server error: %s", detail),
		Transient: false,
		Context:   []KV{{Key: "detail", Value: detail}},
	}
}

// ErrInternalCause creates an internal server error with a cause (non-transient)
func ErrInternalCause(detail string, cause error) *VerdisError {
	return &VerdisError{
		Code:      INTERNAL,
		Message:   fmt.Sprintf("internal server error: %s", detail),
		Cause:     cause,
		Transient: false,
		Context:   []KV{{Key: "detail", Value: detail}},
	}
}

// ErrConfigValidation creates a configuration validation error (non-transient)
func ErrConfigValidation(field, reason string) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   fmt.Sprintf("invalid configuration: %s - %s", field, reason),
		Transient: false,
		Context: []KV{
			{Key: "field", Value: field},
			{Key: "reason", Value: reason},
		},
	}
}

// ErrListenFailed creates a listen failed error (non-transient unless address in use)
func ErrListenFailed(addr string, cause error) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   fmt.Sprintf("failed to listen on %s", addr),
		Cause:     cause,
		Transient: false,
		Context:   []KV{{Key: "address", Value: addr}},
	}
}

// extractPort extracts the port from an address string
func extractPort(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return port
}
