package errors

import "fmt"

// Storage/MVCC layer errors (all non-transient)

// Sentinel errors for common storage conditions
var (
	// ErrKeyNotFound indicates the requested key does not exist
	ErrKeyNotFound = &VerdisError{
		Code:      NOTFOUND,
		Message:   "key not found",
		Transient: false,
	}

	// ErrKeyDeleted indicates the key was deleted at the requested version
	ErrKeyDeleted = &VerdisError{
		Code:      DELETED,
		Message:   "key was deleted at this version",
		Transient: false,
	}

	// ErrVersionNotFound indicates the requested version doesn't exist for the key
	ErrVersionNotFound = &VerdisError{
		Code:      VERSIONNOTFOUND,
		Message:   "version not found",
		Transient: false,
	}
)

// KeyNotFound creates a key not found error with the key in context
func KeyNotFound(key string) *VerdisError {
	return &VerdisError{
		Code:      NOTFOUND,
		Message:   "key not found",
		Transient: false,
		Context:   []KV{{Key: "key", Value: key}},
	}
}

// KeyDeleted creates a key deleted error with context
func KeyDeleted(key string, version uint64) *VerdisError {
	return &VerdisError{
		Code:      DELETED,
		Message:   "key was deleted at this version",
		Transient: false,
		Context: []KV{
			{Key: "key", Value: key},
			{Key: "version", Value: version},
		},
	}
}

// VersionNotFound creates a version not found error with context
func VersionNotFound(key string, version uint64) *VerdisError {
	return &VerdisError{
		Code:      VERSIONNOTFOUND,
		Message:   fmt.Sprintf("version %d not found for key", version),
		Transient: false,
		Context: []KV{
			{Key: "key", Value: key},
			{Key: "version", Value: version},
		},
	}
}
