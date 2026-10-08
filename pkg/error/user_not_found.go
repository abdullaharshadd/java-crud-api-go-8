package error

import (
	"errors"
	"fmt"
)

// UserNotFoundError represents an error that occurs when a user is not found.
// It can optionally wrap another error.
//
// Example usage:
//
//     err := errors.New("Some underlying error")
//     userErr := NewUserNotFoundError("User not found", err)
//     if errors.Is(userErr, err) {
//         // Handle specific underlying error
//     }
//
type UserNotFoundError struct {
	message string
	cause   error
}

// NewUserNotFoundError creates a new instance of UserNotFoundError with the provided message and optional cause.
func NewUserNotFoundError(message string, cause error) error {
	return &UserNotFoundError{
		message: message,
		cause:   cause,
	}
}

// Error returns the error message for UserNotFoundError.
func (e *UserNotFoundError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// IsUserNotFoundError checks if the error is a UserNotFoundError.
func IsUserNotFoundError(err error) bool {
	var userNotFoundError *UserNotFoundError
	return errors.As(err, &userNotFoundError)
}
