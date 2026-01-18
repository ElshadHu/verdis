package errors

// Code represents a Redis-compatible error code prefix
type Code string

// Redis-compatible error codes
const (
	// ERR is the generic error code for most errors
	ERR Code = "ERR"

	// WRONGTYPE indicates an operation against a key holding the wrong type
	WRONGTYPE Code = "WRONGTYPE"

	// SYNTAX indicates a syntax/parsing error in command or protocol
	SYNTAX Code = "SYNTAX"

	// BUSY indicates the server is temporarily busy (transient)
	BUSY Code = "BUSY"

	// TIMEOUT indicates an operation timed out (transient)
	TIMEOUT Code = "TIMEOUT"

	// INTERNAL indicates an internal server error
	INTERNAL Code = "INTERNAL"

	// NOTFOUND indicates a key was not found (internal use, maps to nil for clients)
	NOTFOUND Code = "NOTFOUND"

	// DELETED indicates a key was deleted at the requested version
	DELETED Code = "DELETED"

	// VERSIONNOTFOUND indicates the requested version doesn't exist
	VERSIONNOTFOUND Code = "VERSIONNOTFOUND"
)

// String returns the code as a string
func (c Code) String() string {
	return string(c)
}

// IsTransientCode returns true if this error code typically represents a transient error
func (c Code) IsTransientCode() bool {
	switch c {
	case BUSY, TIMEOUT:
		return true
	default:
		return false
	}
}
