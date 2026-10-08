package service

import (
	"context"
	"migrated-app/pkg/repository"
	"migrated-app/pkg/user"
	"migrated-app/pkg/error"
)

// UserServiceImp implements the UserService interface.
type UserServiceImp struct {
	UserDAO repository.UserDAO
}

// NewUserServiceImp creates a new instance of UserServiceImp.
func NewUserServiceImp(userDAO repository.UserDAO) UserService {
	return &UserServiceImp{
		UserDAO: userDAO,
	}
}

// SaveUser saves a new user to the database.
func (usi *UserServiceImp) SaveUser(ctx context.Context, user *user.User) error {
	return usi.UserDAO.Save(ctx, user)
}

// FetchUserList fetches a list of all users from the database.
func (usi *UserServiceImp) FetchUserList(ctx context.Context) ([]*user.User, error) {
	users, err := usi.UserDAO.FetchAll(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FetchUserByID fetches a specific user by their ID from the database.
func (usi *UserServiceImp) FetchUserByID(ctx context.Context, id int) (*user.User, error) {
	user, err := usi.UserDAO.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser deletes a specific user by their ID from the database.
func (usi *UserServiceImp) DeleteUser(ctx context.Context, id int) error {
	user := user.NewUser()
	user.SetUserID(id)
	return usi.UserDAO.Delete(ctx, user)
}

// UpdateUser updates a specific user by their ID with the provided user object.
func (usi *UserServiceImp) UpdateUser(ctx context.Context, id int, user *user.User) error {
	user.SetUserID(id)
	return usi.UserDAO.Update(ctx, user)
}

// GetUserByName fetches a specific user by their name from the database.
func (usi *UserServiceImp) GetUserByName(ctx context.Context, name string) (*user.User, error) {
	user, err := usi.UserDAO.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return user, nil
}
