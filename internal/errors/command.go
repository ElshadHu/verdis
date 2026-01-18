package errors

import "fmt"

// Command-layer errors (all non-transient)

// ErrUnknownCommand creates an error for unknown commands
func ErrUnknownCommand(cmd string) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   fmt.Sprintf("unknown command '%s'", cmd),
		Transient: false,
		Context:   []KV{{Key: "command", Value: cmd}},
	}
}

// ErrWrongArity creates an error for wrong number of arguments
func ErrWrongArity(cmd string, got, min, max int) *VerdisError {
	var msg string
	if max < 0 {
		msg = fmt.Sprintf("wrong number of arguments for '%s' command", cmd)
	} else if min == max {
		msg = fmt.Sprintf("wrong number of arguments for '%s' command (expected %d, got %d)", cmd, min, got)
	} else {
		msg = fmt.Sprintf("wrong number of arguments for '%s' command (expected %d-%d, got %d)", cmd, min, max, got)
	}

	return &VerdisError{
		Code:      ERR,
		Message:   msg,
		Transient: false,
		Context: []KV{
			{Key: "command", Value: cmd},
			{Key: "got", Value: got},
			{Key: "min", Value: min},
			{Key: "max", Value: max},
		},
	}
}

// ErrWrongType creates an error for type mismatches
func ErrWrongType(operation string) *VerdisError {
	return &VerdisError{
		Code:      WRONGTYPE,
		Message:   "Operation against a key holding the wrong kind of value",
		Transient: false,
		Context:   []KV{{Key: "operation", Value: operation}},
	}
}

// ErrInvalidArgument creates an error for invalid command arguments
func ErrInvalidArgument(arg, reason string) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   fmt.Sprintf("invalid argument '%s': %s", arg, reason),
		Transient: false,
		Context: []KV{
			{Key: "argument", Value: arg},
			{Key: "reason", Value: reason},
		},
	}
}

// ErrInvalidInteger creates an error for invalid integer values
func ErrInvalidInteger(value string) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   "value is not an integer or out of range",
		Transient: false,
		Context:   []KV{{Key: "value", Value: value}},
	}
}

// ErrInvalidVersion creates an error for invalid version numbers
func ErrInvalidVersion(version string) *VerdisError {
	return &VerdisError{
		Code:      ERR,
		Message:   fmt.Sprintf("invalid version number: %s", version),
		Transient: false,
		Context:   []KV{{Key: "version", Value: version}},
	}
}
