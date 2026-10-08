package error

// ErrorMessage represents an error message along with its HTTP status code.
type ErrorMessage struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NewErrorMessage creates a new instance of ErrorMessage with the provided status and message.
func NewErrorMessage(status int, message string) *ErrorMessage {
	return &ErrorMessage{
		Status:  status,
		Message: message,
	}
}