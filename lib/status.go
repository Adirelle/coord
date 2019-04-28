package lib

import (
	"errors"
	"fmt"
)

// Status represents the status of a process
type Status int

const (
	UNDEFINED Status = iota
	STARTED
	SUCCEEDED
	FAILED
)

func (s Status) String() string {
	switch s {
	case UNDEFINED:
		return "undefined"
	case STARTED:
		return "started"
	case SUCCEEDED:
		return "succeeded"
	case FAILED:
		return "failed"
	default:
		return fmt.Sprintf("unknown-status-%d", s)
	}
}

// StatusUpdate changes a status or returns an error explaining why it is no possible
type StatusUpdate interface {
	Execute(Status) (Status, error)
}

// StatusUpdateFunc is a StatusUpdate implementation using a single function
type StatusUpdateFunc func(Status) (Status, error)

func (f StatusUpdateFunc) Execute(s Status) (Status, error) {
	return f(s)
}

var (
	ErrInvalidTransition = errors.New("invalid transition")
)

var predefinedStatusUpdates = map[string]StatusUpdate{
	"start": StatusUpdateFunc(func(s Status) (Status, error) {
		if s == UNDEFINED || s == STARTED {
			return STARTED, nil
		}
		return s, ErrInvalidTransition
	}),
	"finish": StatusUpdateFunc(func(s Status) (Status, error) {
		if s == STARTED {
			return SUCCEEDED, nil
		}
		if s == SUCCEEDED || s == FAILED {
			return s, nil
		}
		return s, ErrInvalidTransition
	}),
	"fail": StatusUpdateFunc(func(s Status) (Status, error) {
		if s != SUCCEEDED {
			return FAILED, nil
		}
		return s, ErrInvalidTransition
	}),
}

type PredefinedStatusUpdate struct {
	StatusUpdate
}

func (u *PredefinedStatusUpdate) UnmarshalText(text []byte) error {
	if f, ok := predefinedStatusUpdates[string(text)]; ok {
		u.StatusUpdate = f
		return nil
	}
	return fmt.Errorf("unknown update: '%s'", text)
}

// StatusPredicate is used to test a status
type StatusPredicate interface {
	IsFulfilled(Status) bool
	IsPossible(Status) bool
}

// StatusPredicateFunc is StatusPredicate implementation using a single function
type StatusPredicateFunc func(Status) (fulfilled bool, possible bool)

func (f StatusPredicateFunc) IsFulfilled(s Status) (fulfilled bool) {
	fulfilled, _ = f(s)
	return
}

func (f StatusPredicateFunc) IsPossible(s Status) (possible bool) {
	_, possible = f(s)
	return
}

var predefinedStatusPredicates = map[string]StatusPredicate{
	// Base predicates
	"started": StatusPredicateFunc(func(s Status) (bool, bool) {
		return s == STARTED, s <= STARTED
	}),
	"succeeded": StatusPredicateFunc(func(s Status) (bool, bool) {
		return s == SUCCEEDED, s <= SUCCEEDED
	}),
	"failed": StatusPredicateFunc(func(s Status) (bool, bool) {
		return s == FAILED, s < SUCCEEDED
	}),

	// Simple predicates
	"running": StatusPredicateFunc(func(s Status) (bool, bool) {
		return s == STARTED, s < SUCCEEDED
	}),
	"finished": StatusPredicateFunc(func(s Status) (bool, bool) {
		return s >= SUCCEEDED, true
	}),
}

type PredefinedStatusPredicate struct {
	StatusPredicate
}

func (p *PredefinedStatusPredicate) UnmarshalText(text []byte) error {
	if pred, ok := predefinedStatusPredicates[string(text)]; ok {
		p.StatusPredicate = pred
		return nil
	}
	return fmt.Errorf("unknown predicate: '%s'", text)
}
