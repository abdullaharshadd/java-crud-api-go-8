package user

import (
	"errors"
	"fmt"
	"migrated-app/internal/config"
	"migrated-app/pkg/db"
	"strings"
	"context"
)

// User represents a user entity in the database.
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	About    string `json:"about"`
}

// NewUser creates a new user instance.
func NewUser(id int, name string, email string, password string, role string, about string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// UserID returns the user ID.
func (u *User) UserID() int {
	return u.ID
}

// SetUserID sets the user ID.
func (u *User) SetUserID(id int) {
	u.ID = id
}

// UserName returns the user name.
func (u *User) UserName() string {
	return u.Name
}

// SetUserName sets the user name.
func (u *User) SetUserName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("please add the user name")
	}
	u.Name = name
	return nil
}

// UserEmail returns the user email.
func (u *User) UserEmail() string {
	return u.Email
}

// SetUserEmail sets the user email.
func (u *User) SetUserEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("please add the user email")
	}
	u.Email = email
	return nil
}

// UserPassword returns the user password.
func (u *User) UserPassword() string {
	return u.Password
}

// SetUserPassword sets the user password.
func (u *User) SetUserPassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("please add the user password")
	}
	u.Password = password
	return nil
}

// UserRole returns the user role.
func (u *User) UserRole() string {
	return u.Role
}

// SetUserRole sets the user role.
func (u *User) SetUserRole(role string) error {
	if strings.TrimSpace(role) == "" {
		return errors.New("please add the user role")
	}
	u.Role = role
	return nil
}

// UserAbout returns the user 'about' information.
func (u *User) UserAbout() string {
	return u.About
}

// SetUserAbout sets the user 'about' information.
func (u *User) SetUserAbout(about string) error {
	if len(about) > 500 {
		return errors.New("user about information exceeds maximum length of 500 characters")
	}
	u.About = about
	return nil
}

// CreateUserTable creates the user table if it does not exist.
func CreateUserTable(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	dbConn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer dbConn.Close()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS USER (
		User_id INT AUTO_INCREMENT PRIMARY KEY,
		User_name VARCHAR(255) NOT NULL,
		User_Email VARCHAR(255) UNIQUE NOT NULL,
		User_Password VARCHAR(255) NOT NULL,
		User_Role VARCHAR(255) NOT NULL,
		User_About TEXT
	);
	`

	_, err = dbConn.ExecContext(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create user table: %w", err)
	}

	return nil
}
