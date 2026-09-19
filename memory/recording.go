// Package memory is the in-process Recording adapter for tests.
package memory

import (
	"context"
	"fmt"
	"sync"

	crudkiller "github.com/luizpedrini/go-crud-killer"
)

// New returns a Recording adapter that stores Versions in memory.
func New[T any]() crudkiller.Recording[T] {
	return &adapter[T]{}
}

type adapter[T any] struct {
	mu      sync.Mutex
	records []crudkiller.Version[T]
}

func (a *adapter[T]) Begin(ctx context.Context) (crudkiller.UnitOfWork[T], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	a.mu.Lock()
	return &uow[T]{adapter: a}, nil
}

type uow[T any] struct {
	adapter *adapter[T]
	pending []crudkiller.Version[T]
	done    bool
}

func (u *uow[T]) Versions(ctx context.Context, identity string) ([]crudkiller.Version[T], error) {
	if err := u.guard(ctx); err != nil {
		return nil, err
	}
	var out []crudkiller.Version[T]
	for _, v := range u.adapter.records {
		if v.Identity == identity {
			out = append(out, v)
		}
	}
	for _, v := range u.pending {
		if v.Identity == identity {
			out = append(out, v)
		}
	}
	return out, nil
}

func (u *uow[T]) Insert(ctx context.Context, v crudkiller.Version[T]) error {
	if err := u.guard(ctx); err != nil {
		return err
	}
	u.pending = append(u.pending, v)
	return nil
}

func (u *uow[T]) Commit(ctx context.Context) error {
	if err := u.guard(ctx); err != nil {
		return err
	}
	u.adapter.records = append(u.adapter.records, u.pending...)
	u.finish()
	return nil
}

func (u *uow[T]) Rollback(ctx context.Context) error {
	if u.done {
		return nil
	}
	u.finish()
	return nil
}

func (u *uow[T]) guard(ctx context.Context) error {
	if u.done {
		return fmt.Errorf("unit of work is finished")
	}
	return ctx.Err()
}

func (u *uow[T]) finish() {
	u.done = true
	u.pending = nil
	u.adapter.mu.Unlock()
}
