// Package repository provides the data-access layer for the Smart Contact
// Manager. It replaces the Spring Data JPA repository interfaces with explicit
// database/sql implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	apperr "migrated-app/pkg/error"
	"migrated-app/pkg/user"
)

// The table and column names below mirror the JPA @Table/@Column mappings on
// com.smartContact.model.User. They are also what user.CreateUserTable
// creates at boot. Every identifier is backtick-quoted because USER is a
// MySQL keyword, and table names are case-sensitive on most Linux MySQL
// installs, so the exact upper-case spelling `USER` is required.
const (
	userTable       = "`USER`"
	colUserID       = "`User_id`"
	colUserName     = "`User_name`"
	colUserEmail    = "`User_Email`"
	colUserPassword = "`User_Password`"
	colUserRole     = "`User_Role`"
	colUserAbout    = "`User_About`"
)

// Pre-built statements. Column order in selectUserColumns must match scanUser.
var (
	selectUserColumns = colUserID + ", " + colUserName + ", " + colUserEmail + ", " +
		colUserPassword + ", " + colUserRole + ", " + colUserAbout

	insertUserSQL = "INSERT INTO " + userTable + " (" + colUserName + ", " + colUserEmail + ", " +
		colUserPassword + ", " + colUserRole + ", " + colUserAbout + ") VALUES (?, ?, ?, ?, ?)"

	selectAllUsersSQL = "SELECT " + selectUserColumns + " FROM " + userTable

	selectUserByIDSQL = selectAllUsersSQL + " WHERE " + colUserID + " = ?"

	// LIMIT 1 mirrors Spring Data's single-result derived query findByName,
	// which returns one entity rather than a collection.
	selectUserByNameSQL = selectAllUsersSQL + " WHERE " + colUserName + " = ? LIMIT 1"

	updateUserSQL = "UPDATE " + userTable + " SET " + colUserName + " = ?, " + colUserEmail + " = ?, " +
		colUserPassword + " = ?, " + colUserRole + " = ?, " + colUserAbout + " = ? WHERE " + colUserID + " = ?"

	deleteUserSQL = "DELETE FROM " + userTable + " WHERE " + colUserID + " = ?"
)

// UserDAO defines the persistence operations for user entities. It replaces
// the Spring Data interface UserDao extends JpaRepository<User, Integer>
// together with its derived query findByName.
type UserDAO interface {
	// Save inserts a new user and sets its generated ID.
	Save(ctx context.Context, user *user.User) error
	// FetchAll returns every stored user (JpaRepository.findAll).
	FetchAll(ctx context.Context) ([]*user.User, error)
	// FindByID returns the user with the given ID or a UserNotFound error.
	FindByID(ctx context.Context, id int) (*user.User, error)
	// FindByName returns the user with the given name or a UserNotFound error.
	FindByName(ctx context.Context, name string) (*user.User, error)
	// Delete removes the given user by its ID.
	Delete(ctx context.Context, user *user.User) error
	// Update persists changes to an existing user.
	Update(ctx context.Context, user *user.User) error
}

// NewUserDAO creates a UserDAO backed by the given database handle.
func NewUserDAO(db *sql.DB) UserDAO {
	return &userDAO{db: db}
}

type userDAO struct {
	db *sql.DB
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanUser reads one row in selectUserColumns order. JPA columns are nullable
// by default, so every string column is scanned through sql.NullString.
func scanUser(rs rowScanner) (*user.User, error) {
	var (
		id                                  int
		name, email, password, role, about sql.NullString
	)
	if err := rs.Scan(&id, &name, &email, &password, &role, &about); err != nil {
		return nil, err
	}
	return user.NewUser(id, name.String, email.String, password.String, role.String, about.String), nil
}

// Save inserts a new user into the USER table and assigns the generated ID.
func (d *userDAO) Save(ctx context.Context, usr *user.User) error {
	res, err := d.db.ExecContext(ctx, insertUserSQL,
		usr.UserName(), usr.UserEmail(), usr.UserPassword(), usr.UserRole(), usr.UserAbout())
	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("save user: last insert id: %w", err)
	}
	usr.SetUserID(int(id))
	return nil
}

// FetchAll retrieves all users from the USER table.
func (d *userDAO) FetchAll(ctx context.Context) ([]*user.User, error) {
	rows, err := d.db.QueryContext(ctx, selectAllUsersSQL)
	if err != nil {
		return nil, fmt.Errorf("fetch users: %w", err)
	}
	defer rows.Close()

	users := []*user.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("fetch users: scan: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fetch users: iterate: %w", err)
	}
	return users, nil
}

// FindByID retrieves a user by its User_id.
func (d *userDAO) FindByID(ctx context.Context, id int) (*user.User, error) {
	u, err := scanUser(d.db.QueryRowContext(ctx, selectUserByIDSQL, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.NewUserNotFoundError(fmt.Sprintf("User with id %d not found", id), nil)
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id %d: %w", id, err)
	}
	return u, nil
}

// FindByName retrieves a user by its User_name (Spring Data derived query
// findByName).
func (d *userDAO) FindByName(ctx context.Context, name string) (*user.User, error) {
	u, err := scanUser(d.db.QueryRowContext(ctx, selectUserByNameSQL, name))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperr.NewUserNotFoundError(fmt.Sprintf("User with name %s not found", name), nil)
	}
	if err != nil {
		return nil, fmt.Errorf("find user by name %q: %w", name, err)
	}
	return u, nil
}

// Delete removes the given user from the USER table by its ID.
func (d *userDAO) Delete(ctx context.Context, usr *user.User) error {
	if _, err := d.db.ExecContext(ctx, deleteUserSQL, usr.UserID()); err != nil {
		return fmt.Errorf("delete user %d: %w", usr.UserID(), err)
	}
	return nil
}

// Update writes all mutable fields of an existing user back to the USER table.
func (d *userDAO) Update(ctx context.Context, usr *user.User) error {
	_, err := d.db.ExecContext(ctx, updateUserSQL,
		usr.UserName(), usr.UserEmail(), usr.UserPassword(), usr.UserRole(), usr.UserAbout(), usr.UserID())
	if err != nil {
		return fmt.Errorf("update user %d: %w", usr.UserID(), err)
	}
	return nil
}
