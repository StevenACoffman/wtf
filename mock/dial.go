package mock

import (
	"context"
	"time"

	"github.com/benbjohnson/wtf"
)

var _ wtf.DialService = (*DialService)(nil)

// DialService represents a mock of wtf.DialService.
type DialService struct {
	FindDialByIDFn      func(ctx context.Context, id int) (*wtf.Dial, error)
	FindDialByIDInvoked bool

	FindDialsFn      func(ctx context.Context, filter wtf.DialFilter) ([]*wtf.Dial, int, error)
	FindDialsInvoked bool

	CreateDialFn      func(ctx context.Context, dial *wtf.Dial) error
	CreateDialInvoked bool

	UpdateDialFn      func(ctx context.Context, id int, upd wtf.DialUpdate) (*wtf.Dial, error)
	UpdateDialInvoked bool

	DeleteDialFn      func(ctx context.Context, id int) error
	DeleteDialInvoked bool

	SetDialMembershipValueFn      func(ctx context.Context, dialID, value int) error
	SetDialMembershipValueInvoked bool

	AverageDialValueReportFn      func(ctx context.Context, start, end time.Time, interval time.Duration) (*wtf.DialValueReport, error)
	AverageDialValueReportInvoked bool
}

func (s *DialService) FindDialByID(ctx context.Context, id int) (*wtf.Dial, error) {
	s.FindDialByIDInvoked = true
	return s.FindDialByIDFn(ctx, id)
}

func (s *DialService) FindDials(
	ctx context.Context,
	filter wtf.DialFilter,
) ([]*wtf.Dial, int, error) {
	s.FindDialsInvoked = true
	return s.FindDialsFn(ctx, filter)
}

func (s *DialService) CreateDial(ctx context.Context, dial *wtf.Dial) error {
	s.CreateDialInvoked = true
	return s.CreateDialFn(ctx, dial)
}

func (s *DialService) UpdateDial(
	ctx context.Context,
	id int,
	upd wtf.DialUpdate,
) (*wtf.Dial, error) {
	s.UpdateDialInvoked = true
	return s.UpdateDialFn(ctx, id, upd)
}

func (s *DialService) DeleteDial(ctx context.Context, id int) error {
	s.DeleteDialInvoked = true
	return s.DeleteDialFn(ctx, id)
}

func (s *DialService) SetDialMembershipValue(ctx context.Context, dialID, value int) error {
	s.SetDialMembershipValueInvoked = true
	return s.SetDialMembershipValueFn(ctx, dialID, value)
}

func (s *DialService) AverageDialValueReport(
	ctx context.Context,
	start, end time.Time,
	interval time.Duration,
) (*wtf.DialValueReport, error) {
	s.AverageDialValueReportInvoked = true
	return s.AverageDialValueReportFn(ctx, start, end, interval)
}
