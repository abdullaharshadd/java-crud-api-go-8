package repository

import (
	"context"
	"database/sql"
	"fmt"

	apperr "migrated-app/pkg/error"
	"migrated-app/pkg/user"
)

// UserDAO defines the interface for managing user entities.
type UserDAO interface {
	Save(ctx context.Context, user *user.User) error
	FetchAll(ctx context.Context) ([]*user.User, error)
	FindByID(ctx context.Context, id int) (*user.User, error)
	FindByName(ctx context.Context, name string) (*user.User, error)
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
func (u *userDAOImpl) Save(ctx context.Context, usr *user.User) error {
	stmt := `INSERT INTO users (name, email, password, role, about) VALUES (?, ?, ?, ?, ?)`
	res, err := u.db.ExecContext(ctx, stmt, usr.UserName(), usr.UserEmail(), usr.UserPassword(), usr.UserRole(), usr.UserAbout())
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	usr.SetUserID(int(id))
	return nil
}

// FetchAll retrieves all users.
func (u *userDAOImpl) FetchAll(ctx context.Context) ([]*user.User, error) {
	rows, err := u.db.QueryContext(ctx, `SELECT id, name, email, password, role, COALESCE(about, '') FROM users`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer rows.Close()
	users := []*user.User{}
	for rows.Next() {
		var id int
		var name, email, password, role, about string
		if err := rows.Scan(&id, &name, &email, &password, &role, &about); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user.NewUser(id, name, email, password, role, about))
	}
	return users, rows.Err()
}

// FindByID retrieves a user by ID.
func (u *userDAOImpl) FindByID(ctx context.Context, id int) (*user.User, error) {
	stmt := `SELECT id, name, email, password, role, COALESCE(about, '') FROM users WHERE id = ?`
	var uid int
	var name, email, password, role, about string
	err := u.db.QueryRowContext(ctx, stmt, id).Scan(&uid, &name, &email, &password, &role, &about)
	if err == sql.ErrNoRows {
		return nil, apperr.NewUserNotFoundError(fmt.Sprintf("User with id %d not found", id), nil)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return user.NewUser(uid, name, email, password, role, about), nil
}

// FindByName retrieves a user by their name.
func (u *userDAOImpl) FindByName(ctx context.Context, name string) (*user.User, error) {
	stmt := `SELECT id, name, email, password, role, COALESCE(about, '') FROM users WHERE name = ?`
	var id int
	var userName, userEmail, userPassword, userRole, userAbout string
	err := u.db.QueryRowContext(ctx, stmt, name).Scan(&id, &userName, &userEmail, &userPassword, &userRole, &userAbout)
	if err == sql.ErrNoRows {
		return nil, apperr.NewUserNotFoundError(fmt.Sprintf("User with name %s not found", name), nil)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find user by name: %w", err)
	}
	return user.NewUser(id, userName, userEmail, userPassword, userRole, userAbout), nil
}

// Delete removes a user from the database.
func (u *userDAOImpl) Delete(ctx context.Context, usr *user.User) error {
	_, err := u.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, usr.UserID())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Update modifies an existing user in the database.
func (u *userDAOImpl) Update(ctx context.Context, usr *user.User) error {
	stmt := `UPDATE users SET name = ?, email = ?, password = ?, role = ?, about = ? WHERE id = ?`
	_, err := u.db.ExecContext(ctx, stmt, usr.UserName(), usr.UserEmail(), usr.UserPassword(), usr.UserRole(), usr.UserAbout(), usr.UserID())
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
