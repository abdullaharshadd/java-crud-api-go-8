package service

import (
	"context"
	"database/sql"
	"testing"

	"migrated-app/pkg/error"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/user"
)

type mockUserDAO struct {
	repository.UserDAO
	saveErr      error
	fetchAllErr  error
	findByIDErr  error
	findByNameErr error
	deleteErr    error
	updateErr    error
}

func (m *mockUserDAO) Save(ctx context.Context, user *user.User) error {
	return m.saveErr
}

func (m *mockUserDAO) FetchAll(ctx context.Context) ([]*user.User, error) {
	return []*user.User{}, m.fetchAllErr
}

func (m *mockUserDAO) FindByID(ctx context.Context, id int) (*user.User, error) {
	return &user.User{}, m.findByIDErr
}

func (m *mockUserDAO) FindByName(ctx context.Context, name string) (*user.User, *error.UserNotFoundError) {
	return &user.User{}, m.findByNameErr
}

func (m *mockUserDAO) Delete(ctx context.Context, user *user.User) error {
	return m.deleteErr
}

func (m *mockUserDAO) Update(ctx context.Context, user *user.User) error {
	return m.updateErr
}

func TestSaveUser(t *testing.T) {
	tests := []struct {
		name        string
		user        *user.User
		expectedErr error
	}{
		{"ValidUser", user.NewUser(0, "John Doe", "john@example.com", "password123", "admin", "About John"), nil},
		{"InvalidUser", user.NewUser(0, "", "john@example.com", "password123", "admin", "About John"), errors.New("please add the user name")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &mockUserDAO{saveErr: tt.expectedErr}
			userService := NewUserServiceImp(mockDAO)
			err := userService.SaveUser(context.Background(), tt.user)
			if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestFetchUserList(t *testing.T) {
	tests := []struct {
		name        string
		fetchAllErr error
	}{
		{"Success", nil},
		{"Failure", errors.New("database error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &mockUserDAO{fetchAllErr: tt.fetchAllErr}
			userService := NewUserServiceImp(mockDAO)
			_, err := userService.FetchUserList(context.Background())
			if err != tt.fetchAllErr {
				t.Errorf("Expected error %v, got %v", tt.fetchAllErr, err)
			}
		})
	}
}

func TestFetchUserByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		findByIDErr error
	}{
		{"ValidID", 1, nil},
		{"InvalidID", 2, &error.UserNotFoundError{"User not found", nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &mockUserDAO{findByIDErr: tt.findByIDErr}
			userService := NewUserServiceImp(mockDAO)
			_, err := userService.FetchUserByID(context.Background(), tt.id)
			if !error.IsUserNotFoundError(err) && tt.findByIDErr != nil {
				t.Errorf("Expected error %v, got %v", tt.findByIDErr, err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		deleteErr error
	}{
		{"ValidID", 1, nil},
		{"InvalidID", 2, errors.New("database error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &mockUserDAO{deleteErr: tt.deleteErr}
			userService := NewUserServiceImp(mockDAO)
			err := userService.DeleteUser(context.Background(), tt.id)
			if err != tt.deleteErr {
				t.Errorf("Expected error %v, got %v", tt.deleteErr, err)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		user      *user.User
		updateErr error
	}{
		{"ValidUser", 1, user.NewUser(0, "Jane Doe", "jane@example.com", "password456", "user", "About Jane"), nil},
		{"InvalidUser", 2, user.NewUser(0, "", "jane@example.com", "password456", "user", "About Jane"), errors.New("please add the user name")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &mockUserDAO{updateErr: tt.updateErr}
			userService := NewUserServiceImp(mockDAO)
			err := userService.UpdateUser(context.Background(), tt.id, tt.user)
			if err != tt.updateErr {
				t.Errorf("Expected error %v, got %v", tt.updateErr, err)
			}
		})
	}
}

func TestGetUserByName(t *testing.T) {
	tests := []struct {
		name            string
		name            string
		findByNameErr   error
		expectedUser    *user.User
		expectedUserErr error
	}{
		{"ValidName", "hemraj", nil, user.NewUser(1, "hemraj", "hemraj@example.com", "password789", "user", "About Hemraj"), nil},
		{"InvalidName", "nonexistent", &error.UserNotFoundError{"User not found", nil}, nil, &error.UserNotFoundError{"User not found", nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &mockUserDAO{findByNameErr: tt.findByNameErr}
			userService := NewUserServiceImp(mockDAO)
			user, err := userService.GetUserByName(context.Background(), tt.name)
			if !error.IsUserNotFoundError(err) && tt.expectedUserErr != nil {
				t.Errorf("Expected error %v, got %v", tt.expectedUserErr, err)
			}
			if user != nil && user.Name != tt.expectedUser.Name {
				t.Errorf("Expected user name %s, got %s", tt.expectedUser.Name, user.Name)
			}
		})
	}
}