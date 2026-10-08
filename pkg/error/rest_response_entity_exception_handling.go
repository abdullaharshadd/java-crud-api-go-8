package error

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// Middleware to handle UserNotFoundError and return appropriate HTTP response.
func HandleUserNotFoundError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if err := c.Errors.Last(); err != nil && IsUserNotFoundError(err.Err) {
			userErr := err.Err.(*UserNotFoundError)
			errorMsg := NewErrorMessage(http.StatusNotFound, userErr.message)
			c.JSON(http.StatusNotFound, errorMsg)
			c.Abort()
		}
	}
}