package mock

import (
	"context"

	"github.com/benbjohnson/wtf"
)

var _ wtf.UserService = (*UserService)(nil)

type UserService struct {
	FindUserByIDFn      func(ctx context.Context, id int) (*wtf.User, error)
	FindUserByIDInvoked bool

	FindUsersFn      func(ctx context.Context, filter wtf.UserFilter) ([]*wtf.User, int, error)
	FindUsersInvoked bool

	CreateUserFn      func(ctx context.Context, user *wtf.User) error
	CreateUserInvoked bool

	UpdateUserFn      func(ctx context.Context, id int, upd wtf.UserUpdate) (*wtf.User, error)
	UpdateUserInvoked bool

	DeleteUserFn      func(ctx context.Context, id int) error
	DeleteUserInvoked bool
}

func (s *UserService) FindUserByID(ctx context.Context, id int) (*wtf.User, error) {
	s.FindUserByIDInvoked = true
	return s.FindUserByIDFn(ctx, id)
}

func (s *UserService) FindUsers(
	ctx context.Context,
	filter wtf.UserFilter,
) ([]*wtf.User, int, error) {
	s.FindUsersInvoked = true
	return s.FindUsersFn(ctx, filter)
}

func (s *UserService) CreateUser(ctx context.Context, user *wtf.User) error {
	s.CreateUserInvoked = true
	return s.CreateUserFn(ctx, user)
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	id int,
	upd wtf.UserUpdate,
) (*wtf.User, error) {
	s.UpdateUserInvoked = true
	return s.UpdateUserFn(ctx, id, upd)
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	s.DeleteUserInvoked = true
	return s.DeleteUserFn(ctx, id)
}
