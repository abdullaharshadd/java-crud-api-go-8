package service

import (
	"context"
	"testing"

	"migrated-app/pkg/error"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/user"
)

type mockUserDAO struct {
	repository.UserDAO
	saveErr        error
	fetchAllErr    error
	findByIDErr    error
	deleteErr      error
	updateErr      error
	findByNameUser *user.User
	findByNameErr  error
}

func (m *mockUserDAO) Save(ctx context.Context, user *user.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	user.SetUserID(1)
	return nil
}

func (m *mockUserDAO) FetchAll(ctx context.Context) ([]*user.User, error) {
	if m.fetchAllErr != nil {
		return nil, m.fetchAllErr
	}
	return []*user.User{
		user.NewUser(1, "Alice", "alice@example.com", "password", "admin", "About Alice"),
		user.NewUser(2, "Bob", "bob@example.com", "password", "user", "About Bob"),
	}, nil
}

func (m *mockUserDAO) FindByID(ctx context.Context, id int) (*user.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if id == 1 {
		return user.NewUser(1, "Alice", "alice@example.com", "password", "admin", "About Alice"), nil
	}
	return nil, error.NewUserNotFoundError("User not found", nil)
}

func (m *mockUserDAO) Delete(ctx context.Context, user *user.User) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return nil
}

func (m *mockUserDAO) Update(ctx context.Context, user *user.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return nil
}

func (m *mockUserDAO) FindByName(ctx context.Context, name string) (*user.User, *error.UserNotFoundError) {
	if m.findByNameErr != nil {
		return nil, m.findByNameErr
	}
	if name == "Alice" {
		return m.findByNameUser, nil
	}
	return nil, error.NewUserNotFoundError("User not found", nil)
}

func TestUserService(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		dao            *mockUserDAO
		user           *user.User
		expectedOutput *user.User
		expectedError  error
	}{
		{
			name: "SaveUser - Success",
			dao: &mockUserDAO{},
			user: user.NewUser(0, "Charlie", "charlie@example.com", "password", "guest", "About Charlie"),
			expectedOutput: user.NewUser(1, "Charlie", "charlie@example.com", "password", "guest", "About Charlie"),
			expectedError:  nil,
		},
		{
			name: "SaveUser - Failure",
			dao: &mockUserDAO{
				saveErr: errors.New("Failed to save user"),
			},
			user:           user.NewUser(0, "Diana", "diana@example.com", "password", "user", "About Diana"),
			expectedOutput: nil,
			expectedError:  errors.New("Failed to save user"),
		},
		{
			name: "FetchUserList - Empty",
			dao: &mockUserDAO{
				fetchAllErr: errors.New("No users found"),
			},
			expectedOutput: nil,
			expectedError:  errors.New("No users found"),
		},
		{
			name: "FetchUserList - Multiple Users",
			dao: &mockUserDAO{},
			expectedOutput: []*user.User{
				user.NewUser(1, "Alice", "alice@example.com", "password", "admin", "About Alice"),
				user.NewUser(2, "Bob", "bob@example.com", "password", "user", "About Bob"),
			},
			expectedError: nil,
		},
		{
			name: "FetchUserByID - Valid ID",
			dao: &mockUserDAO{},
			user: user.NewUser(1, "", "", "", "", ""),
			expectedOutput: user.NewUser(1, "Alice", "alice@example.com", "password", "admin", "About Alice"),
			expectedError:  nil,
		},
		{
			name: "FetchUserByID - Invalid ID",
			dao: &mockUserDAO{
				findByIDErr: error.NewUserNotFoundError("User not found", nil),
			},
			user:           user.NewUser(3, "", "", "", "", ""),
			expectedOutput: nil,
			expectedError:  error.NewUserNotFoundError("User not found", nil),
		},
		{
			name: "DeleteUser - Valid ID",
			dao: &mockUserDAO{},
			user: user.NewUser(1, "", "", "", "", ""),
			expectedOutput: nil,
			expectedError:  nil,
		},
		{
			name: "DeleteUser - Invalid ID",
			dao: &mockUserDAO{
				deleteErr: errors.New("User not found"),
			},
			user:           user.NewUser(3, "", "", "", "", ""),
			expectedOutput: nil,
			expectedError:  errors.New("User not found"),
		},
		{
			name: "UpdateUser - Success",
			dao: &mockUserDAO{},
			user: user.NewUser(1, "Alice", "alice@example.com", "new_password", "admin", "Updated About Alice"),
			expectedOutput: nil,
			expectedError:  nil,
		},
		{
			name: "UpdateUser - Failure",
			dao: &mockUserDAO{
				updateErr: errors.New("Failed to update user"),
			},
			user:           user.NewUser(1, "Alice", "alice@example.com", "new_password", "admin", "Updated About Alice"),
			expectedOutput: nil,
			expectedError:  errors.New("Failed to update user"),
		},
		{
			name: "GetUserByName - Valid Name",
			dao: &mockUserDAO{
				findByNameUser: user.NewUser(1, "Alice", "alice@example.com", "password", "admin", "About Alice"),
			},
			user:           user.NewUser(1, "Alice", "", "", "", ""),
			expectedOutput: user.NewUser(1, "Alice", "alice@example.com", "password", "admin", "About Alice"),
			expectedError:  nil,
		},
		{
			name: "GetUserByName - Invalid Name",
			dao: &mockUserDAO{
				findByNameErr: error.NewUserNotFoundError("User not found", nil),
			},
			user:           user.NewUser(1, "Eve", "", "", "", ""),
			expectedOutput: nil,
			expectedError:  error.NewUserNotFoundError("User not found", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewUserService(tt.dao)
			switch tt.name {
			case "SaveUser - Success", "SaveUser - Failure":
				err := svc.SaveUser(ctx, tt.user)
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				if tt.expectedOutput != nil && tt.user.UserID() != tt.expectedOutput.UserID() {
					t.Errorf("Expected user ID %d, got %d", tt.expectedOutput.UserID(), tt.user.UserID())
				}
			case "FetchUserList - Empty", "FetchUserList - Multiple Users":
				users, err := svc.FetchUserList(ctx)
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				if !equalUsers(users, tt.expectedOutput) {
					t.Errorf("Expected users %v, got %v", tt.expectedOutput, users)
				}
			case "FetchUserByID - Valid ID", "FetchUserByID - Invalid ID":
				user, err := svc.FetchUserByID(ctx, tt.user.UserID())
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				if !equalUsers([]*user.User{user}, []*user.User{tt.expectedOutput}) {
					t.Errorf("Expected user %v, got %v", tt.expectedOutput, user)
				}
			case "DeleteUser - Valid ID", "DeleteUser - Invalid ID":
				err := svc.DeleteUser(ctx, tt.user.UserID())
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			case "UpdateUser - Success", "UpdateUser - Failure":
				err := svc.UpdateUser(ctx, tt.user.UserID(), tt.user)
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			case "GetUserByName - Valid Name", "GetUserByName - Invalid Name":
				user, err := svc.GetUserByName(ctx, tt.user.UserName())
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				if !equalUsers([]*user.User{user}, []*user.User{tt.expectedOutput}) {
					t.Errorf("Expected user %v, got %v", tt.expectedOutput, user)
				}
			}
		})
	}
}

func equalUsers(a, b []*user.User) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].UserID() != b[i].UserID() || a[i].UserName() != b[i].UserName() || a[i].UserEmail() != b[i].UserEmail() || a[i].UserPassword() != b[i].UserPassword() || a[i].UserRole() != b[i].UserRole() || a[i].UserAbout() != b[i].UserAbout() {
			return false
		}
	}
	return true
}