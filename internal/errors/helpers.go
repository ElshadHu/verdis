package errors

import (
	"errors"
)

// Is reports whether any error in err's chain matches target.
// This is a convenience wrapper around errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
// This is a convenience wrapper around errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// IsTransient checks if an error (or any error in its chain) is transient.
// A transient error indicates a temporary failure that can be retried.
func IsTransient(err error) bool {
	if err == nil {
		return false
	}

	// Check if error implements TransientError interface
	var te TransientError
	if errors.As(err, &te) {
		return te.IsTransient()
	}

	return false
}

// IsNotFound checks if the error indicates a key was not found
func IsNotFound(err error) bool {
	var ve *VerdisError
	if errors.As(err, &ve) {
		return ve.Code == NOTFOUND
	}
	return false
}

// IsDeleted checks if the error indicates a key was deleted
func IsDeleted(err error) bool {
	var ve *VerdisError
	if errors.As(err, &ve) {
		return ve.Code == DELETED
	}
	return false
}

// IsVersionNotFound checks if the error indicates a version was not found
func IsVersionNotFound(err error) bool {
	var ve *VerdisError
	if errors.As(err, &ve) {
		return ve.Code == VERSIONNOTFOUND
	}
	return false
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	var ve *VerdisError
	if errors.As(err, &ve) {
		return ve.Code == TIMEOUT
	}
	return false
}

// IsBusy checks if the error indicates the server is busy
func IsBusy(err error) bool {
	var ve *VerdisError
	if errors.As(err, &ve) {
		return ve.Code == BUSY
	}
	return false
}

// Code extracts the error code from a VerdisError, or returns ERR if not a VerdisError
func GetCode(err error) Code {
	var ve *VerdisError
	if errors.As(err, &ve) {
		return ve.Code
	}
	return ERR
}

// ToVerdisError converts any error to a VerdisError.
// If it's already a VerdisError, returns it as-is.
// Otherwise, wraps it in a generic ERR VerdisError.
func ToVerdisError(err error) *VerdisError {
	if err == nil {
		return nil
	}

	var ve *VerdisError
	if errors.As(err, &ve) {
		return ve
	}

	return &VerdisError{
		Code:      ERR,
		Message:   err.Error(),
		Cause:     err,
		Transient: false,
	}
}

// Must panics if err is not nil. Useful for initialization code.
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// Join combines multiple errors into one.
// This is a convenience wrapper around errors.Join.
func Join(errs ...error) error {
	return errors.Join(errs...)
}
