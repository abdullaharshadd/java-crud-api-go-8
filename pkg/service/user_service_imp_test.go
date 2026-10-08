package service

import (
	"context"
	"errors"
	"testing"

	apperr "migrated-app/pkg/error"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/user"
)

type mockImpUserDAO struct {
	repository.UserDAO
	saveErr        error
	fetchAllErr    error
	findByIDErr    error
	findByNameUser *user.User
	findByNameErr  error
	deleteErr      error
	updateErr      error
}

func (m *mockImpUserDAO) Save(ctx context.Context, u *user.User) error {
	return m.saveErr
}

func (m *mockImpUserDAO) FetchAll(ctx context.Context) ([]*user.User, error) {
	return []*user.User{}, m.fetchAllErr
}

func (m *mockImpUserDAO) FindByID(ctx context.Context, id int) (*user.User, error) {
	return &user.User{}, m.findByIDErr
}

func (m *mockImpUserDAO) FindByName(ctx context.Context, name string) (*user.User, error) {
	if m.findByNameErr != nil {
		return nil, m.findByNameErr
	}
	return m.findByNameUser, nil
}

func (m *mockImpUserDAO) Delete(ctx context.Context, u *user.User) error {
	return m.deleteErr
}

func (m *mockImpUserDAO) Update(ctx context.Context, u *user.User) error {
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
			mockDAO := &mockImpUserDAO{saveErr: tt.expectedErr}
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
			mockDAO := &mockImpUserDAO{fetchAllErr: tt.fetchAllErr}
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
		{"InvalidID", 2, apperr.NewUserNotFoundError("User not found", nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDAO := &mockImpUserDAO{findByIDErr: tt.findByIDErr}
			userService := NewUserServiceImp(mockDAO)
			_, err := userService.FetchUserByID(context.Background(), tt.id)
			if !apperr.IsUserNotFoundError(err) && tt.findByIDErr != nil {
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
			mockDAO := &mockImpUserDAO{deleteErr: tt.deleteErr}
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
			mockDAO := &mockImpUserDAO{updateErr: tt.updateErr}
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
		testName        string
		name            string
		findByNameErr   error
		expectedUser    *user.User
		expectedUserErr error
	}{
		{"ValidName", "hemraj", nil, user.NewUser(1, "hemraj", "hemraj@example.com", "password789", "user", "About Hemraj"), nil},
		{"InvalidName", "nonexistent", apperr.NewUserNotFoundError("User not found", nil), nil, apperr.NewUserNotFoundError("User not found", nil)},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockDAO := &mockImpUserDAO{findByNameUser: tt.expectedUser, findByNameErr: tt.findByNameErr}
			userService := NewUserServiceImp(mockDAO)
			got, err := userService.GetUserByName(context.Background(), tt.name)
			if !apperr.IsUserNotFoundError(err) && tt.expectedUserErr != nil {
				t.Errorf("Expected error %v, got %v", tt.expectedUserErr, err)
			}
			if got != nil && tt.expectedUser != nil && got.Name != tt.expectedUser.Name {
				t.Errorf("Expected user name %s, got %s", tt.expectedUser.Name, got.Name)
			}
		})
	}
}
