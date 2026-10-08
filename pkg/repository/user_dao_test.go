package repository

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"migrated-app/pkg/user"
	"migrated-app/pkg/error"
)

type mockDB struct {
	mockQueryRow func(query string, args ...interface{}) (*sql.Row, error)
	mockExec     func(query string, args ...interface{}) (driver.Result, error)
}

func (m *mockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	row, _ := m.mockQueryRow(query, args...)
	return row
}

func (m *mockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (driver.Result, error) {
	return m.mockExec(query, args...)
}

func TestUserDAO_FindByName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name        string
		args        args
		mockDB      func() *mockDB
		wantUser    *user.User
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name: "Valid user name",
			args: args{
				ctx:  context.Background(),
				name: "JohnDoe",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					mockQueryRow: func(query string, args ...interface{}) (*sql.Row, error) {
						return sql.Rows{
							Scan: func(dest ...interface{}) error {
								dest[0].(*int) = 1
								dest[1].(*string) = "JohnDoe"
								dest[2].(*string) = "john.doe@example.com"
								dest[3].(*string) = "password123"
								dest[4].(*string) = "admin"
								dest[5].(*string) = "About John Doe"
								return nil
							},
						}.NextRow(), nil
					},
					mockExec: func(query string, args ...interface{}) (driver.Result, error) {
						return nil, nil
					},
				}
			},
			wantUser: user.NewUser(1, "JohnDoe", "john.doe@example.com", "password123", "admin", "About John Doe"),
			wantErr:  false,
		},
		{
			name: "User name does not exist",
			args: args{
				ctx:  context.Background(),
				name: "NonExistentUser",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					mockQueryRow: func(query string, args ...interface{}) (*sql.Row, error) {
						return sql.Rows{}, sql.ErrNoRows
					},
					mockExec: func(query string, args ...interface{}) (driver.Result, error) {
						return nil, nil
					},
				}
			},
			wantUser: nil,
			wantErr:  true,
			wantErrMsg: "User with name NonExistentUser not found",
		},
		{
			name: "Empty user name",
			args: args{
				ctx:  context.Background(),
				name: "",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					mockQueryRow: func(query string, args ...interface{}) (*sql.Row, error) {
						return sql.Rows{}, sql.ErrNoRows
					},
					mockExec: func(query string, args ...interface{}) (driver.Result, error) {
						return nil, nil
					},
				}
			},
			wantUser: nil,
			wantErr:  true,
			wantErrMsg: "User with name  not found",
		},
		{
			name: "Null user name",
			args: args{
				ctx:  context.Background(),
				name: "",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					mockQueryRow: func(query string, args ...interface{}) (*sql.Row, error) {
						return sql.Rows{}, sql.ErrNoRows
					},
					mockExec: func(query string, args ...interface{}) (driver.Result, error) {
						return nil, nil
					},
				}
			},
			wantUser: nil,
			wantErr:  true,
			wantErrMsg: "User with name  not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao := NewUserDAO(tt.mockDB().mockDB)
			gotUser, gotErr := dao.FindByName(tt.args.ctx, tt.args.name)

			if !tt.wantErr && gotErr != nil {
				t.Errorf("%s: unexpected error: %v", tt.name, gotErr)
			}
			if tt.wantErr && gotErr == nil {
				t.Errorf("%s: expected error but got none", tt.name)
			}
			if tt.wantErr && gotErr != nil && gotErr.Error() != tt.wantErrMsg {
				t.Errorf("%s: expected error message '%s', got '%s'", tt.name, tt.wantErrMsg, gotErr.Error())
			}
			if !tt.wantErr && gotUser != nil && gotUser.UserID() != tt.wantUser.UserID() {
				t.Errorf("%s: expected user ID %d, got %d", tt.name, tt.wantUser.UserID(), gotUser.UserID())
			}
			if !tt.wantErr && gotUser != nil && gotUser.UserName() != tt.wantUser.UserName() {
				t.Errorf("%s: expected user name '%s', got '%s'", tt.name, tt.wantUser.UserName(), gotUser.UserName())
			}
			if !tt.wantErr && gotUser != nil && gotUser.UserEmail() != tt.wantUser.UserEmail() {
				t.Errorf("%s: expected user email '%s', got '%s'", tt.name, tt.wantUser.UserEmail(), gotUser.UserEmail())
			}
			if !tt.wantErr && gotUser != nil && gotUser.UserPassword() != tt.wantUser.UserPassword() {
				t.Errorf("%s: expected user password '%s', got '%s'", tt.name, tt.wantUser.UserPassword(), gotUser.UserPassword())
			}
			if !tt.wantErr && gotUser != nil && gotUser.UserRole() != tt.wantUser.UserRole() {
				t.Errorf("%s: expected user role '%s', got '%s'", tt.name, tt.wantUser.UserRole(), gotUser.UserRole())
			}
			if !tt.wantErr && gotUser != nil && gotUser.UserAbout() != tt.wantUser.UserAbout() {
				t.Errorf("%s: expected user about '%s', got '%s'", tt.name, tt.wantUser.UserAbout(), gotUser.UserAbout())
			}
		})
	}
}