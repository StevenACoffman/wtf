package mock

import (
	"context"

	"github.com/benbjohnson/wtf"
)

var (
	_ wtf.EventService = (*EventService)(nil)
	_ wtf.Subscription = (*Subscription)(nil)
)

type EventService struct {
	PublishEventFn      func(userID int, event wtf.Event)
	PublishEventInvoked bool

	SubscribeFn      func(ctx context.Context) (wtf.Subscription, error)
	SubscribeInvoked bool
}

func (s *EventService) PublishEvent(userID int, event wtf.Event) {
	s.PublishEventInvoked = true
	s.PublishEventFn(userID, event)
}

func (s *EventService) Subscribe(ctx context.Context) (wtf.Subscription, error) {
	s.SubscribeInvoked = true
	return s.SubscribeFn(ctx)
}

type Subscription struct {
	CloseFn      func() error
	CloseInvoked bool

	CFn      func() <-chan wtf.Event
	CInvoked bool
}

func (s *Subscription) Close() error {
	s.CloseInvoked = true
	return s.CloseFn()
}

func (s *Subscription) C() <-chan wtf.Event {
	s.CInvoked = true
	return s.CFn()
}
