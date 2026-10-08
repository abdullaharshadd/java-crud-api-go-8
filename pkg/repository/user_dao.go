package repository

import (
	"context"
	"database/sql"
	"fmt"
	"migrated-app/pkg/user"
	err "migrated-app/pkg/error"
)

// UserDAO defines the interface for managing user entities.
type UserDAO interface {
	Save(ctx context.Context, user *user.User) error
	FindByName(ctx context.Context, name string) (*user.User, error.Error)
	Delete(ctx context.Context, user *user.User) error
	Update(ctx context.Context, user *user.User) error
}

// NewUserDAO creates a new instance of UserDAO.
func NewUserDAO(db *sql.DB) UserDAO {
	return &userDAOImpl{db: db}
}

type userDAOImpl struct {
	db *sql.DB
}

// Save inserts a new user into the database.
func (u *userDAOImpl) Save(ctx context.Context, user *user.User) error {
	stmt := `INSERT INTO users (name, email, password, role, about) VALUES (?, ?, ?, ?, ?)`
	res, err := u.db.ExecContext(ctx, stmt, user.UserName(), user.UserEmail(), user.UserPassword(), user.UserRole(), user.UserAbout())
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	user.SetUserID(int(id))
	return nil
}

// FindByName retrieves a user by their name.
func (u *userDAOImpl) FindByName(ctx context.Context, name string) (*user.User, *err.UserNotFoundError) {
	stmt := `SELECT id, name, email, password, role, about FROM users WHERE name = ?`
	var id int
	var userName string
	var userEmail string
	var userPassword string
	var userRole string
	var userAbout string
	err := u.db.QueryRowContext(ctx, stmt, name).Scan(&id, &userName, &userEmail, &userPassword, &userRole, &userAbout)
	if err == sql.ErrNoRows {
		return nil, err.NewUserNotFoundError(fmt.Sprintf("User with name %s not found", name), nil)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find user by name: %w", err)
	}
	return user.NewUser(id, userName, userEmail, userPassword, userRole, userAbout), nil
}

// Delete removes a user from the database.
func (u *userDAOImpl) Delete(ctx context.Context, user *user.User) error {
	stmt := `DELETE FROM users WHERE id = ?`
	_, err := u.db.ExecContext(ctx, stmt, user.UserID())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Update modifies an existing user in the database.
func (u *userDAOImpl) Update(ctx context.Context, user *user.User) error {
	stmt := `UPDATE users SET name = ?, email = ?, password = ?, role = ?, about = ? WHERE id = ?`
	_, err := u.db.ExecContext(ctx, stmt, user.UserName(), user.UserEmail(), user.UserPassword(), user.UserRole(), user.UserAbout(), user.UserID())
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
