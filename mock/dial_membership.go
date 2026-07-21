package mock

import (
	"context"

	"github.com/benbjohnson/wtf"
)

var _ wtf.DialMembershipService = (*DialMembershipService)(nil)

type DialMembershipService struct {
	FindDialMembershipByIDFn      func(ctx context.Context, id int) (*wtf.DialMembership, error)
	FindDialMembershipByIDInvoked bool

	FindDialMembershipsFn      func(ctx context.Context, filter wtf.DialMembershipFilter) ([]*wtf.DialMembership, int, error)
	FindDialMembershipsInvoked bool

	CreateDialMembershipFn      func(ctx context.Context, membership *wtf.DialMembership) error
	CreateDialMembershipInvoked bool

	UpdateDialMembershipFn      func(ctx context.Context, id int, upd wtf.DialMembershipUpdate) (*wtf.DialMembership, error)
	UpdateDialMembershipInvoked bool

	DeleteDialMembershipFn      func(ctx context.Context, id int) error
	DeleteDialMembershipInvoked bool
}

func (s *DialMembershipService) FindDialMembershipByID(
	ctx context.Context,
	id int,
) (*wtf.DialMembership, error) {
	s.FindDialMembershipByIDInvoked = true
	return s.FindDialMembershipByIDFn(ctx, id)
}

func (s *DialMembershipService) FindDialMemberships(
	ctx context.Context,
	filter wtf.DialMembershipFilter,
) ([]*wtf.DialMembership, int, error) {
	s.FindDialMembershipsInvoked = true
	return s.FindDialMembershipsFn(ctx, filter)
}

func (s *DialMembershipService) CreateDialMembership(
	ctx context.Context,
	membership *wtf.DialMembership,
) error {
	s.CreateDialMembershipInvoked = true
	return s.CreateDialMembershipFn(ctx, membership)
}

func (s *DialMembershipService) UpdateDialMembership(
	ctx context.Context,
	id int,
	upd wtf.DialMembershipUpdate,
) (*wtf.DialMembership, error) {
	s.UpdateDialMembershipInvoked = true
	return s.UpdateDialMembershipFn(ctx, id, upd)
}

func (s *DialMembershipService) DeleteDialMembership(ctx context.Context, id int) error {
	s.DeleteDialMembershipInvoked = true
	return s.DeleteDialMembershipFn(ctx, id)
}
