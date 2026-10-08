package service

import (
	"context"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/user"
	errorpkg "migrated-app/pkg/error"
)

// UserService defines the interface for managing user entities.
type UserService interface {
	SaveUser(ctx context.Context, user *user.User) error
	FetchUserList(ctx context.Context) ([]*user.User, error)
	FetchUserByID(ctx context.Context, id int) (*user.User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUser(ctx context.Context, id int, user *user.User) error
	GetUserByName(ctx context.Context, name string) (*user.User, error)
}

// NewUserService creates a new instance of UserService.
func NewUserService(userDAO repository.UserDAO) UserService {
	return &userService{
		userDAO: userDAO,
	}
}

type userService struct {
	userDAO repository.UserDAO
}

// SaveUser inserts a new user into the database.
func (us *userService) SaveUser(ctx context.Context, user *user.User) error {
	return us.userDAO.Save(ctx, user)
}

// FetchUserList retrieves a list of all users.
func (us *userService) FetchUserList(ctx context.Context) ([]*user.User, error) {
	// Assuming there's a method in UserDao to fetch all users.
	users, err := us.userDAO.FetchAll(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FetchUserByID retrieves a user by their unique identifier.
func (us *userService) FetchUserByID(ctx context.Context, id int) (*user.User, error) {
	user, err := us.userDAO.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser deletes a user by their unique identifier.
func (us *userService) DeleteUser(ctx context.Context, id int) error {
	user := user.NewUser()
	user.SetUserID(id)
	return us.userDAO.Delete(ctx, user)
}

// UpdateUser updates a user by their unique identifier with the provided user data.
func (us *userService) UpdateUser(ctx context.Context, id int, user *user.User) error {
	user.SetUserID(id)
	return us.userDAO.Update(ctx, user)
}

// GetUserByName retrieves a user by their username.
func (us *userService) GetUserByName(ctx context.Context, name string) (*user.User, error) {
	user, err := us.userDAO.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return user, nil
}