package service

import (
	"context"
	"errors"
	"testing"

	apperr "migrated-app/pkg/error"
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

func (m *mockUserDAO) Save(ctx context.Context, u *user.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	u.SetUserID(1)
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
	return nil, apperr.NewUserNotFoundError("User not found", nil)
}

func (m *mockUserDAO) Delete(ctx context.Context, u *user.User) error {
	return m.deleteErr
}

func (m *mockUserDAO) Update(ctx context.Context, u *user.User) error {
	return m.updateErr
}

func (m *mockUserDAO) FindByName(ctx context.Context, name string) (*user.User, error) {
	if m.findByNameErr != nil {
		return nil, m.findByNameErr
	}
	if name == "Alice" {
		return m.findByNameUser, nil
	}
	return nil, apperr.NewUserNotFoundError("User not found", nil)
}

func TestUserService(t *testing.T) {
	ctx := context.Background()
	errSave := errors.New("Failed to save user")
	errFetch := errors.New("No users found")
	errNotFound := apperr.NewUserNotFoundError("User not found", nil)
	errDelete := errors.New("User not found")
	errUpdate := errors.New("Failed to update user")

	alice := user.NewUser(1, "Alice", "alice@example.com", "password", "admin", "About Alice")

	tests := []struct {
		name           string
		dao            *mockUserDAO
		user           *user.User
		expectedOutput []*user.User
		expectedError  error
	}{
		{"SaveUser - Success", &mockUserDAO{}, user.NewUser(0, "Charlie", "charlie@example.com", "password", "guest", "About Charlie"),
			[]*user.User{user.NewUser(1, "Charlie", "charlie@example.com", "password", "guest", "About Charlie")}, nil},
		{"SaveUser - Failure", &mockUserDAO{saveErr: errSave}, user.NewUser(0, "Diana", "diana@example.com", "password", "user", "About Diana"), nil, errSave},
		{"FetchUserList - Empty", &mockUserDAO{fetchAllErr: errFetch}, nil, nil, errFetch},
		{"FetchUserList - Multiple Users", &mockUserDAO{}, nil, []*user.User{
			user.NewUser(1, "Alice", "alice@example.com", "password", "admin", "About Alice"),
			user.NewUser(2, "Bob", "bob@example.com", "password", "user", "About Bob"),
		}, nil},
		{"FetchUserByID - Valid ID", &mockUserDAO{}, user.NewUser(1, "", "", "", "", ""), []*user.User{alice}, nil},
		{"FetchUserByID - Invalid ID", &mockUserDAO{findByIDErr: errNotFound}, user.NewUser(3, "", "", "", "", ""), nil, errNotFound},
		{"DeleteUser - Valid ID", &mockUserDAO{}, user.NewUser(1, "", "", "", "", ""), nil, nil},
		{"DeleteUser - Invalid ID", &mockUserDAO{deleteErr: errDelete}, user.NewUser(3, "", "", "", "", ""), nil, errDelete},
		{"UpdateUser - Success", &mockUserDAO{}, user.NewUser(1, "Alice", "alice@example.com", "new_password", "admin", "Updated About Alice"), nil, nil},
		{"UpdateUser - Failure", &mockUserDAO{updateErr: errUpdate}, user.NewUser(1, "Alice", "alice@example.com", "new_password", "admin", "Updated About Alice"), nil, errUpdate},
		{"GetUserByName - Valid Name", &mockUserDAO{findByNameUser: alice}, user.NewUser(1, "Alice", "", "", "", ""), []*user.User{alice}, nil},
		{"GetUserByName - Invalid Name", &mockUserDAO{findByNameErr: errNotFound}, user.NewUser(1, "Eve", "", "", "", ""), nil, errNotFound},
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
				if len(tt.expectedOutput) == 1 && tt.user.UserID() != tt.expectedOutput[0].UserID() {
					t.Errorf("Expected user ID %d, got %d", tt.expectedOutput[0].UserID(), tt.user.UserID())
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
				got, err := svc.FetchUserByID(ctx, tt.user.UserID())
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				if !equalOne(got, tt.expectedOutput) {
					t.Errorf("Expected user %v, got %v", tt.expectedOutput, got)
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
				got, err := svc.GetUserByName(ctx, tt.user.UserName())
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				if !equalOne(got, tt.expectedOutput) {
					t.Errorf("Expected user %v, got %v", tt.expectedOutput, got)
				}
			}
		})
	}
}

func equalOne(got *user.User, expected []*user.User) bool {
	if got == nil {
		return len(expected) == 0
	}
	return equalUsers([]*user.User{got}, expected)
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
