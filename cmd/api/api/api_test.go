package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"migrated-app/internal/config"
	"migrated-app/pkg/error"
	"migrated-app/pkg/service"
	"migrated-app/pkg/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var mockUserService service.UserService

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	mockUserService = &mockUserServiceImp{}
	m.Run()
}

type mockUserServiceImp struct{}

func (m *mockUserServiceImp) SaveUser(ctx context.Context, user *user.User) error {
	if user.Name == "error" {
		return errors.New("internal server error")
	}
	return nil
}

func (m *mockUserServiceImp) FetchUserList(ctx context.Context) ([]*user.User, error) {
	return []*user.User{}, nil
}

func (m *mockUserServiceImp) FetchUserByID(ctx context.Context, id int) (*user.User, error) {
	if id == 404 {
		return nil, error.NewUserNotFoundError("User not found", nil)
	}
	return &user.User{ID: id, Name: "Test User"}, nil
}

func (m *mockUserServiceImp) DeleteUser(ctx context.Context, id int) error {
	return nil
}

func (m *mockUserServiceImp) UpdateUser(ctx context.Context, id int, user *user.User) error {
	return nil
}

func (m *mockUserServiceImp) GetUserNameByName(ctx context.Context, name string) (*user.User, error) {
	if name == "notfound" {
		return nil, error.NewUserNotFoundError("User not found", nil)
	}
	return &user.User{Name: name}, nil
}

func TestSaveUser(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name     string
		args     args
		expected int
	}{
		{"valid user", args{body: `{"name":"John Doe","email":"john@example.com","password":"securepassword","role":"admin","about":"I am John."}`}, http.StatusOK},
		{"invalid user", args{body: `{"name":"","email":"john@example.com","password":"securepassword","role":"admin","about":"I am John."}`}, http.StatusBadRequest},
		{"server error", args{body: `{"name":"error","email":"john@example.com","password":"securepassword","role":"admin","about":"I am John."}`}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/save_user_data", bytes.NewBufferString(tt.args.body))

			userController := NewUserController(mockUserService)
			userController.saveUser(c)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestFetchUserList(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/get_user_data", nil)

	userController := NewUserController(mockUserService)
	userController.fetchUserList(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFetchUserById(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected int
	}{
		{"valid user", "1", http.StatusOK},
		{"invalid id", "notanumber", http.StatusBadRequest},
		{"user not found", "404", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/get_user_data/"+tt.id, nil)
			c.Params = append(c.Params, gin.Param{"id", tt.id})

			userController := NewUserController(mockUserService)
			userController.fetchUserById(c)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected int
	}{
		{"valid id", "1", http.StatusOK},
		{"invalid id", "notanumber", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("DELETE", "/delete_user_data/"+tt.id, nil)
			c.Params = append(c.Params, gin.Param{"id", tt.id})

			userController := NewUserController(mockUserService)
			userController.deleteUser(c)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		body     string
		expected int
	}{
		{"valid user", "1", `{"name":"John Doe","email":"john@example.com","password":"securepassword","role":"admin","about":"I am John."}`, http.StatusOK},
		{"invalid id", "notanumber", `{"name":"John Doe","email":"john@example.com","password":"securepassword","role":"admin","about":"I am John."}`, http.StatusBadRequest},
		{"invalid user", "1", `{"name":"","email":"john@example.com","password":"securepassword","role":"admin","about":"I am John."}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("PUT", "/update_user_data/"+tt.id, bytes.NewBufferString(tt.body))
			c.Params = append(c.Params, gin.Param{"id", tt.id})

			userController := NewUserController(mockUserService)
			userController.updateUser(c)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestGetUserNameByName(t *testing.T) {
	tests := []struct {
		name     string
		name     string
		expected int
	}{
		{"valid name", "John Doe", http.StatusOK},
		{"user not found", "notfound", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/get_user_name/name/"+tt.name, nil)
			c.Params = append(c.Params, gin.Param{"name", tt.name})

			userController := NewUserController(mockUserService)
			userController.getUserNameByName(c)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}