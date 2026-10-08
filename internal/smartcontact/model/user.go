// Package user holds the User domain model for the Smart Contact Manager and
// the DDL that creates its storage at boot. It is the Go counterpart of the
// JPA entity com.smartContact.model.User (@Entity @Table(name = "USER")).
//
// MIGRATION_NOTE: the package name stays "user", even though the directory is
// model/, because the already-migrated callers (application.go, user_dao.go,
// api.go) refer to this package as user.User and user.CreateUserTable.
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"migrated-app/internal/config"
	"migrated-app/pkg/db"
)

// Physical schema names taken verbatim from the JPA @Table / @Column mappings.
const (
	// TableName is the physical table from @Table(name = "USER"). USER is a
	// reserved word in MySQL, so it must always be backtick-quoted in SQL.
	TableName = "USER"
	// ColumnID maps User.id (@Id @GeneratedValue(strategy = AUTO)).
	ColumnID = "User_id"
	// ColumnName maps User.name.
	ColumnName = "User_name"
	// ColumnEmail maps User.email (unique = true).
	ColumnEmail = "User_Email"
	// ColumnPassword maps User.password.
	ColumnPassword = "User_Password"
	// ColumnRole maps User.role.
	ColumnRole = "User_Role"
	// ColumnAbout maps User.about (length = 500).
	ColumnAbout = "User_About"

	// EmailUniqueConstraint is the name of the unique constraint on User_Email.
	// MySQL includes it in duplicate-key errors, which IsDuplicateEmailError
	// relies on.
	EmailUniqueConstraint = "uk_user_email"

	// CompatViewName is the updatable view that exposes USER under the table
	// and column names used by pkg/repository/user_dao.go
	// (users: id, name, email, password, role, about).
	CompatViewName = "users"

	// AboutMaxLength mirrors @Column(name = "User_About", length = 500).
	AboutMaxLength = 500
)

// Validation errors. The message of ErrNameBlank is copied verbatim from the
// source @NotBlank annotation.
var (
	// ErrNameBlank mirrors @NotBlank(message = "please Add the department Name").
	ErrNameBlank = errors.New("please Add the department Name")
	// ErrAboutTooLong is returned when About exceeds the 500-character column.
	ErrAboutTooLong = fmt.Errorf("user about must be at most %d characters", AboutMaxLength)
	// ErrDuplicateEmail can be returned by repositories when the unique
	// constraint on User_Email is violated (see IsDuplicateEmailError).
	ErrDuplicateEmail = errors.New("a user with this email already exists")
)

// User is a registered user of the Smart Contact Manager.
// The zero value replaces Lombok's @NoArgsConstructor.
type User struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
	Role     string `json:"role" db:"role"`
	About    string `json:"about" db:"about"`
}

// NewUser builds a User from all of its fields. It replaces Lombok's
// @AllArgsConstructor and @Builder. It performs no validation; call Validate
// before persisting.
func NewUser(id int, name, email, password, role, about string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		About:    about,
	}
}

// Validate enforces the constraints declared on the JPA entity: name must not
// be blank (@NotBlank), and about must fit in its 500-character column.
// All violations are joined into a single error.
func (u *User) Validate() error {
	var errs []error
	if strings.TrimSpace(u.Name) == "" {
		errs = append(errs, ErrNameBlank)
	}
	if utf8.RuneCountInString(u.About) > AboutMaxLength {
		errs = append(errs, ErrAboutTooLong)
	}
	return errors.Join(errs...)
}

// String returns a debug representation of the user, replacing Lombok's
// generated toString().
//
// MIGRATION_NOTE: Lombok's toString() printed the password. It is redacted
// here so credentials never reach logs.
func (u *User) String() string {
	return fmt.Sprintf("User(id=%d, name=%s, email=%s, password=[REDACTED], role=%s, about=%s)",
		u.ID, u.Name, u.Email, u.Role, u.About)
}

// The accessors below keep the method set of the previous migration so that
// existing callers still compile. New code should use the exported fields
// directly.

// UserID returns the user ID.
func (u *User) UserID() int { return u.ID }

// SetUserID sets the user ID.
func (u *User) SetUserID(id int) { u.ID = id }

// UserName returns the user name.
func (u *User) UserName() string { return u.Name }

// SetUserName sets the user name. It returns ErrNameBlank if the name is blank.
func (u *User) SetUserName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameBlank
	}
	u.Name = name
	return nil
}

// UserEmail returns the user email.
func (u *User) UserEmail() string { return u.Email }

// SetUserEmail sets the user email. The source declares no validation on
// email beyond database uniqueness, so this never fails. The error return is
// kept for signature compatibility.
func (u *User) SetUserEmail(email string) error {
	u.Email = email
	return nil
}

// UserPassword returns the user password.
func (u *User) UserPassword() string { return u.Password }

// SetUserPassword sets the user password. The source declares no validation
// on it, so this never fails.
func (u *User) SetUserPassword(password string) error {
	u.Password = password
	return nil
}

// UserRole returns the user role.
func (u *User) UserRole() string { return u.Role }

// SetUserRole sets the user role. The source declares no validation on it,
// so this never fails.
func (u *User) SetUserRole(role string) error {
	u.Role = role
	return nil
}

// UserAbout returns the user's 'about' text.
func (u *User) UserAbout() string { return u.About }

// SetUserAbout sets the user's 'about' text. It returns ErrAboutTooLong if
// the text exceeds 500 characters.
func (u *User) SetUserAbout(about string) error {
	if utf8.RuneCountInString(about) > AboutMaxLength {
		return ErrAboutTooLong
	}
	u.About = about
	return nil
}

// IsDuplicateEmailError reports whether err is a MySQL duplicate-key error
// raised by the unique constraint on User_Email. Repositories can use it to
// translate driver errors into ErrDuplicateEmail.
func IsDuplicateEmailError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrDuplicateEmail) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") && strings.Contains(msg, EmailUniqueConstraint)
}

// Execer is the minimal database capability needed to run DDL.
// *sql.DB, *sql.Conn and *sql.Tx all satisfy it.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// createTableSQL is the DDL Hibernate would derive from the entity:
//   - The table name is backtick-quoted because USER is reserved in MySQL.
//   - int with GenerationType.AUTO becomes INT AUTO_INCREMENT PRIMARY KEY.
//   - Plain String columns become VARCHAR(255), Hibernate's default length.
//   - User_About becomes VARCHAR(500), from length = 500.
//   - User_Email gets a named UNIQUE constraint, from unique = true.
//
// @NotBlank is a bean-validation rule, not a DDL rule, so like Hibernate no
// NOT NULL is emitted. Validate enforces it in the application instead.
const createTableSQL = "CREATE TABLE IF NOT EXISTS `" + TableName + "` (" +
	"`" + ColumnID + "` INT NOT NULL AUTO_INCREMENT, " +
	"`" + ColumnName + "` VARCHAR(255), " +
	"`" + ColumnEmail + "` VARCHAR(255), " +
	"`" + ColumnPassword + "` VARCHAR(255), " +
	"`" + ColumnRole + "` VARCHAR(255), " +
	"`" + ColumnAbout + "` VARCHAR(500), " +
	"PRIMARY KEY (`" + ColumnID + "`), " +
	"CONSTRAINT `" + EmailUniqueConstraint + "` UNIQUE (`" + ColumnEmail + "`)" +
	")"

// createCompatViewSQL exposes USER under the names that
// pkg/repository/user_dao.go queries (table `users`, columns id, name, email,
// password, role, about).
//
// It is a single-table view with a 1:1 column mapping and no aggregates, so
// MySQL treats it as updatable and insertable. INSERT, UPDATE, DELETE and
// SELECT all reach the real USER table. INSERT honours AUTO_INCREMENT and
// LAST_INSERT_ID(), and inserts still enforce uk_user_email.
const createCompatViewSQL = "CREATE OR REPLACE ALGORITHM = MERGE VIEW `" + CompatViewName + "` AS SELECT " +
	"`" + ColumnID + "` AS `id`, " +
	"`" + ColumnName + "` AS `name`, " +
	"`" + ColumnEmail + "` AS `email`, " +
	"`" + ColumnPassword + "` AS `password`, " +
	"`" + ColumnRole + "` AS `role`, " +
	"`" + ColumnAbout + "` AS `about` " +
	"FROM `" + TableName + "`"

// EnsureSchema creates the USER table if it does not already exist, then
// creates or replaces the `users` compatibility view used by the repository.
// The two statements run separately because the MySQL driver rejects
// multi-statement Exec calls by default. It is idempotent and safe to call on
// every boot.
func EnsureSchema(ctx context.Context, exec Execer) error {
	if exec == nil {
		return errors.New("ensure user schema: nil database handle")
	}
	if _, err := exec.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("create table %s: %w", TableName, err)
	}
	if _, err := exec.ExecContext(ctx, createCompatViewSQL); err != nil {
		return fmt.Errorf("create view %s over %s: %w", CompatViewName, TableName, err)
	}
	return nil
}

// CreateUserTable creates the user schema at boot. It keeps the signature
// already called from internal/smartcontact/application.go, opens a
// short-lived connection from configuration, and delegates to EnsureSchema.
//
// MIGRATION_NOTE: prefer calling EnsureSchema(ctx, db) with the application's
// shared *sql.DB. That avoids opening a second pool just for DDL.
func CreateUserTable(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database connection: %w", err)
	}
	defer conn.Close()

	return EnsureSchema(ctx, conn)
}
