package error

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUserNotFoundError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	type args struct {
		err error
	}
	type want struct {
		status  int
		message string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "User not found",
			args: args{err: NewUserNotFoundError("User not found", nil)},
			want: want{status: http.StatusNotFound, message: "User not found"},
		},
		{
			name: "Underlying error present",
			args: args{err: NewUserNotFoundError("User not found", errors.New("database error"))},
			want: want{status: http.StatusNotFound, message: "User not found"},
		},
		{
			name: "No error",
			args: args{err: nil},
			want: want{status: http.StatusOK, message: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(HandleUserNotFoundError())
			r.GET("/:any", func(c *gin.Context) {
				if tt.args.err != nil {
					_ = c.Error(tt.args.err)
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.want.status, w.Code)
			if tt.want.status == http.StatusNotFound {
				var resp ErrorMessage
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.want.status, resp.Status)
				assert.Equal(t, tt.want.message, resp.Message)
			}
		})
	}
}
