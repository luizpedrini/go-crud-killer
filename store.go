package crudkiller

import (
	"context"
	"fmt"
)

// Store is the inbound port: callers Create (and later Read, Edit, Terminate,
// History, List) through this type. Tests exercise behavior here, not through
// adapter internals.
type Store[T any] struct {
	clock    Clock
	rec      Recording[T]
	validate func(T) error
}

// Option configures a Store at construction.
type Option[T any] func(*Store[T])

// WithValidate registers optional Payload validation. Envelope validation
// stays the library's.
func WithValidate[T any](fn func(T) error) Option[T] {
	return func(s *Store[T]) {
		s.validate = fn
	}
}

// New constructs a Store over an injected Clock and Recording adapter.
func New[T any](clock Clock, rec Recording[T], opts ...Option[T]) (*Store[T], error) {
	if clock == nil {
		return nil, fmt.Errorf("clock is required")
	}
	if rec == nil {
		return nil, fmt.Errorf("recording is required")
	}
	s := &Store[T]{clock: clock, rec: rec}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s, nil
}

// Create records a Version for a caller-minted Identity. valid is the
// Valid-time period; a zero Period means [clock.Now(), unbounded).
// Transaction time is taken from the Clock and is not a caller argument.
func (s *Store[T]) Create(ctx context.Context, identity string, payload T, actor Actor, valid Period) (Version[T], error) {
	var zero Version[T]
	if identity == "" {
		return zero, fmt.Errorf("create: identity must be non-empty")
	}
	if actor.ID == "" {
		return zero, fmt.Errorf("create %q: %w", identity, ErrMissingActor)
	}

	now := s.clock.Now().UTC()
	if valid.IsZero() {
		valid = Period{From: now}
	} else {
		valid = valid.utc()
	}
	if err := valid.wellFormed(); err != nil {
		return zero, fmt.Errorf("create %q: %w: %v", identity, ErrInvalidPeriod, err)
	}
	if s.validate != nil {
		if err := s.validate(payload); err != nil {
			return zero, fmt.Errorf("create %q: %w", identity, err)
		}
	}

	uow, err := s.rec.Begin(ctx)
	if err != nil {
		return zero, fmt.Errorf("create %q: %w", identity, err)
	}
	defer func() { _ = uow.Rollback(ctx) }()

	existing, err := uow.Versions(ctx, identity)
	if err != nil {
		return zero, fmt.Errorf("create %q: %w", identity, err)
	}
	for _, v := range existing {
		if !v.TransactionTime.Unbounded() {
			continue
		}
		if v.ValidTime.Overlaps(valid) {
			return zero, fmt.Errorf("create %q: %w", identity, ErrAlreadyExists)
		}
	}

	ver := Version[T]{
		Identity:        identity,
		Payload:         payload,
		Actor:           actor,
		ValidTime:       valid,
		TransactionTime: Period{From: now},
	}
	if err := uow.Insert(ctx, ver); err != nil {
		return zero, fmt.Errorf("create %q: %w", identity, err)
	}
	if err := uow.Commit(ctx); err != nil {
		return zero, fmt.Errorf("create %q: %w", identity, err)
	}
	return ver, nil
}
