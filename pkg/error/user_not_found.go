package error

import (
	"errors"
)

// UserNotFoundError represents an error that occurs when a user is not found.
type UserNotFoundError struct {
	Message string
}

// NewUserNotFoundError creates a new instance of UserNotFoundError with the provided message.
func NewUserNotFoundError(message string) error {
	return &UserNotFoundError{Message: message}
}

// Error returns the error message for UserNotFoundError.
func (e *UserNotFoundError) Error() string {
	return e.Message
}

// IsUserNotFoundError checks if the error is a UserNotFoundError.
func IsUserNotFoundError(err error) bool {
	var userNotFoundError *UserNotFoundError
	return errors.As(err, &userNotFoundError)
}