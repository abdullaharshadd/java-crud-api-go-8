package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"migrated-app/pkg/error"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/service"
	"migrated-app/pkg/user"
)

var testUsers = []*user.User{
	{ID: 1, Name: "Alice", Email: "alice@example.com", Password: "securepassword", Role: "admin", About: "Loves cats"},
	{ID: 2, Name: "Bob", Email: "bob@example.com", Password: "anotherpassword", Role: "user", About: "Enjoys hiking"},
}

type MockUserDAO struct {
	mock.Mock
}

func (m *MockUserDAO) Save(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserDAO) FetchAll(ctx context.Context) ([]*user.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockUserDAO) FindByID(ctx context.Context, id int) (*user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserDAO) Delete(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserDAO) Update(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserDAO) FindByName(ctx context.Context, name string) (*user.User, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*user.User), args.Error(1)
}

func TestSaveUser(t *testing.T) {
	type args struct {
		user *user.User
	}
	tests := []struct {
		name          string
		args          args
		wantStatus    int
		wantBody      string
		sideEffects   func(m *MockUserDAO)
	}{
		{
			name: "Valid User",
			args: args{
				user: &user.User{Name: "Charlie", Email: "charlie@example.com", Password: "password", Role: "guest", About: "Explorer"},
			},
			wantStatus: http.StatusOK,
			wantBody:   "User data saved successfully!",
			sideEffects: func(m *MockUserDAO) {
				m.On("Save", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					return u.Name == "Charlie" && u.Email == "charlie@example.com" && u.Password == "password" && u.Role == "guest" && u.About == "Explorer"
				})).Return(nil)
			},
		},
		{
			name: "Invalid JSON",
			args: args{
				user: &user.User{},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid JSON",
			sideEffects: func(m *MockUserDAO) {},
		},
		{
			name: "Database Error",
			args: args{
				user: &user.User{Name: "Dave", Email: "dave@example.com", Password: "password", Role: "guest", About: "Explorer"},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Internal server error",
			sideEffects: func(m *MockUserDAO) {
				m.On("Save", mock.Anything, mock.Anything).Return(fmt.Errorf("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockUserDAO{}
			tt.sideEffects(m)

			userService := service.NewUserService(m)
			userCtrl := NewUserController(userService)
			r := gin.Default()
			r.POST("/save_user_data", userCtrl.SaveUser)

			jsonData, _ := json.Marshal(tt.args.user)
			req, _ := http.NewRequest(http.MethodPost, "/save_user_data", bytes.NewBuffer(jsonData))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("SaveUser() = %v, want %v", w.Code, tt.wantStatus)
			}

			var respBody map[string]interface{}
			json.Unmarshal([]byte(w.Body.String()), &respBody)
			if respBody["message"] != tt.wantBody {
				t.Errorf("SaveUser() = %v, want %v", respBody["message"], tt.wantBody)
			}
		})
	}
}

func TestFetchUserList(t *testing.T) {
	tests := []struct {
		name          string
		sideEffects   func(m *MockUserDAO)
		wantStatus    int
		wantBody      string
	}{
		{
			name: "Successful Fetch",
			sideEffects: func(m *MockUserDAO) {
				m.On("FetchAll", mock.Anything).Return(testUsers, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `[{"id":1,"name":"Alice","email":"alice@example.com","password":"securepassword","role":"admin","about":"Loves cats"},{"id":2,"name":"Bob","email":"bob@example.com","password":"anotherpassword","role":"user","about":"Enjoys hiking"}]`,
		},
		{
			name: "Database Error",
			sideEffects: func(m *MockUserDAO) {
				m.On("FetchAll", mock.Anything).Return(nil, fmt.Errorf("database error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"database error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockUserDAO{}
			tt.sideEffects(m)

			userService := service.NewUserService(m)
			userCtrl := NewUserController(userService)
			r := gin.Default()
			r.GET("/get_user_data", userCtrl.FetchUserList)

			req, _ := http.NewRequest(http.MethodGet, "/get_user_data", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("FetchUserList() = %v, want %v", w.Code, tt.wantStatus)
			}

			body := w.Body.String()
			if body != tt.wantBody {
				t.Errorf("FetchUserList() = %v, want %v", body, tt.wantBody)
			}
		})
	}
}

func TestFetchUserByID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		sideEffects   func(m *MockUserDAO)
		wantStatus    int
		wantBody      string
	}{
		{
			name: "Valid User ID",
			id:   "1",
			sideEffects: func(m *MockUserDAO) {
				m.On("FindByID", mock.Anything, 1).Return(testUsers[0], nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"id":1,"name":"Alice","email":"alice@example.com","password":"securepassword","role":"admin","about":"Loves cats"}`,
		},
		{
			name: "Invalid User ID",
			id:   "invalid",
			sideEffects: func(m *MockUserDAO) {},
			wantStatus:  http.StatusBadRequest,
			wantBody:    `"error":"Invalid user ID"`,
		},
		{
			name: "User Not Found",
			id:   "3",
			sideEffects: func(m *MockUserDAO) {
				m.On("FindByID", mock.Anything, 3).Return(nil, error.NewUserNotFoundError("User not found", nil))
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"User not found"`,
		},
		{
			name: "Database Error",
			id:   "4",
			sideEffects: func(m *MockUserDAO) {
				m.On("FindByID", mock.Anything, 4).Return(nil, fmt.Errorf("database error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"database error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockUserDAO{}
			tt.sideEffects(m)

			userService := service.NewUserService(m)
			userCtrl := NewUserController(userService)
			r := gin.Default()
			r.GET("/get_user_data/:id", userCtrl.FetchUserByID)

			req, _ := http.NewRequest(http.MethodGet, "/get_user_data/"+tt.id, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("FetchUserByID() = %v, want %v", w.Code, tt.wantStatus)
			}

			body := w.Body.String()
			if body != tt.wantBody {
				t.Errorf("FetchUserByID() = %v, want %v", body, tt.wantBody)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		sideEffects   func(m *MockUserDAO)
		wantStatus    int
		wantBody      string
	}{
		{
			name: "Valid User ID",
			id:   "1",
			sideEffects: func(m *MockUserDAO) {
				m.On("Delete", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					return u.UserID() == 1
				})).Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `"message":"User data deleted successfully!"`,
		},
		{
			name: "Invalid User ID",
			id:   "invalid",
			sideEffects: func(m *MockUserDAO) {},
			wantStatus:  http.StatusBadRequest,
			wantBody:    `"error":"Invalid user ID"`,
		},
		{
			name: "Database Error",
			id:   "2",
			sideEffects: func(m *MockUserDAO) {
				m.On("Delete", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					return u.UserID() == 2
				})).Return(fmt.Errorf("database error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"database error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockUserDAO{}
			tt.sideEffects(m)

			userService := service.NewUserService(m)
			userCtrl := NewUserController(userService)
			r := gin.Default()
			r.DELETE("/delete_user_data/:id", userCtrl.DeleteUser)

			req, _ := http.NewRequest(http.MethodDelete, "/delete_user_data/"+tt.id, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("DeleteUser() = %v, want %v", w.Code, tt.wantStatus)
			}

			body := w.Body.String()
			if body != tt.wantBody {
				t.Errorf("DeleteUser() = %v, want %v", body, tt.wantBody)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		user          *user.User
		sideEffects   func(m *MockUserDAO)
		wantStatus    int
		wantBody      string
	}{
		{
			name: "Valid User ID and Data",
			id:   "1",
			user: &user.User{Name: "Alice", Email: "alice@example.com", Password: "newpassword", Role: "admin", About: "Loves cats"},
			sideEffects: func(m *MockUserDAO) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					return u.UserID() == 1 && u.UserPassword() == "newpassword"
				})).Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"id":1,"name":"Alice","email":"alice@example.com","password":"newpassword","role":"admin","about":"Loves cats"}`,
		},
		{
			name: "Invalid User ID",
			id:   "invalid",
			user: &user.User{},
			sideEffects: func(m *MockUserDAO) {},
			wantStatus:  http.StatusBadRequest,
			wantBody:    `"error":"Invalid user ID"`,
		},
		{
			name: "Invalid User Object",
			id:   "1",
			user: &user.User{},
			sideEffects: func(m *MockUserDAO) {},
			wantStatus:  http.StatusBadRequest,
			wantBody:    `"error":"Key: 'User.Name' Error:Field validation for 'Name' failed on the 'required' tag"`,
		},
		{
			name: "Database Error",
			id:   "2",
			user: &user.User{Name: "Bob", Email: "bob@example.com", Password: "anotherpassword", Role: "user", About: "Enjoys hiking"},
			sideEffects: func(m *MockUserDAO) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
					return u.UserID() == 2
				})).Return(fmt.Errorf("database error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"database error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockUserDAO{}
			tt.sideEffects(m)

			userService := service.NewUserService(m)
			userCtrl := NewUserController(userService)
			r := gin.Default()
			r.PUT("/update_user_data/:id", userCtrl.UpdateUser)

			jsonData, _ := json.Marshal(tt.user)
			req, _ := http.NewRequest(http.MethodPut, "/update_user_data/"+tt.id, bytes.NewBuffer(jsonData))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("UpdateUser() = %v, want %v", w.Code, tt.wantStatus)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.wantBody) {
				t.Errorf("UpdateUser() = %v, want %v", body, tt.wantBody)
			}
		})
	}
}

func TestGetUserByName(t *testing.T) {
	tests := []struct {
		name          string
		name          string
		sideEffects   func(m *MockUserDAO)
		wantStatus    int
		wantBody      string
	}{
		{
			name: "Valid User Name",
			name: "Alice",
			sideEffects: func(m *MockUserDAO) {
				m.On("FindByName", mock.Anything, "Alice").Return(testUsers[0], nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"id":1,"name":"Alice","email":"alice@example.com","password":"securepassword","role":"admin","about":"Loves cats"}`,
		},
		{
			name: "User Not Found",
			name: "Charlie",
			sideEffects: func(m *MockUserDAO) {
				m.On("FindByName", mock.Anything, "Charlie").Return(nil, error.NewUserNotFoundError("User not found", nil))
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"User not found"`,
		},
		{
			name: "Database Error",
			name: "Bob",
			sideEffects: func(m *MockUserDAO) {
				m.On("FindByName", mock.Anything, "Bob").Return(nil, fmt.Errorf("database error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"database error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockUserDAO{}
			tt.sideEffects(m)

			userService := service.NewUserService(m)
			userCtrl := NewUserController(userService)
			r := gin.Default()
			r.GET("/get_user_name/name/:name", userCtrl.GetUserByName)

			req, _ := http.NewRequest(http.MethodGet, "/get_user_name/name/"+tt.name, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetUserByName() = %v, want %v", w.Code, tt.wantStatus)
			}

			body := w.Body.String()
			if body != tt.wantBody {
				t.Errorf("GetUserByName() = %v, want %v", body, tt.wantBody)
			}
		})
	}
}