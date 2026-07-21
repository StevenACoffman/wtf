package mock

import (
	"context"

	"github.com/benbjohnson/wtf"
)

var _ wtf.AuthService = (*AuthService)(nil)

type AuthService struct {
	FindAuthByIDFn      func(ctx context.Context, id int) (*wtf.Auth, error)
	FindAuthByIDInvoked bool

	FindAuthsFn      func(ctx context.Context, filter wtf.AuthFilter) ([]*wtf.Auth, int, error)
	FindAuthsInvoked bool

	CreateAuthFn      func(ctx context.Context, auth *wtf.Auth) error
	CreateAuthInvoked bool

	DeleteAuthFn      func(ctx context.Context, id int) error
	DeleteAuthInvoked bool
}

func (s *AuthService) FindAuthByID(ctx context.Context, id int) (*wtf.Auth, error) {
	s.FindAuthByIDInvoked = true
	return s.FindAuthByIDFn(ctx, id)
}

func (s *AuthService) FindAuths(
	ctx context.Context,
	filter wtf.AuthFilter,
) ([]*wtf.Auth, int, error) {
	s.FindAuthsInvoked = true
	return s.FindAuthsFn(ctx, filter)
}

func (s *AuthService) CreateAuth(ctx context.Context, auth *wtf.Auth) error {
	s.CreateAuthInvoked = true
	return s.CreateAuthFn(ctx, auth)
}

func (s *AuthService) DeleteAuth(ctx context.Context, id int) error {
	s.DeleteAuthInvoked = true
	return s.DeleteAuthFn(ctx, id)
}
