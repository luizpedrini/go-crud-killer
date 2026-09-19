// Package storetest is the Store[T] conformance suite.
//
// Behavior is defined at Store[T]. The same Create scenarios run for each
// Recording adapter; this ticket wires the memory adapter.
package storetest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	crudkiller "github.com/luizpedrini/go-crud-killer"
)

// Employee is the Feature Payload.
type Employee struct {
	Salary int
}

// NewStore constructs a Store[Employee] on an injected Clock. Adapters pass their
// Recording implementation here; tests never reach inside it.
type NewStore func(clock crudkiller.Clock, opts ...crudkiller.Option[Employee]) (crudkiller.Store[Employee], error)

// ClockNow is the Feature Background instant: 2024-04-15T12:00:00Z.
var ClockNow = time.Date(2024, 4, 15, 12, 0, 0, 0, time.UTC)

var (
	pastFrom = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	hr       = crudkiller.Actor{ID: "hr"}
)

func openFrozen(t *testing.T, newStore NewStore, opts ...crudkiller.Option[Employee]) crudkiller.Store[Employee] {
	t.Helper()
	store, err := newStore(crudkiller.Frozen(ClockNow), opts...)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func assertUTC(t *testing.T, ts time.Time) {
	t.Helper()
	if ts.IsZero() {
		return
	}
	if ts.Location() != time.UTC {
		t.Fatalf("timestamp location %s, want UTC", ts.Location())
	}
}

func assertValidNowUnbounded(t *testing.T, got crudkiller.Version[Employee]) {
	t.Helper()
	if !got.ValidTime.From.Equal(ClockNow) || !got.ValidTime.Unbounded() {
		t.Fatalf("Valid time = [%v, %v), want [%v, unbounded)", got.ValidTime.From, got.ValidTime.To, ClockNow)
	}
}

// Create runs the Create Feature scenarios and issue #4 criteria through Store[T].
func Create(t *testing.T, newStore NewStore) {
	t.Helper()

	t.Run("Create records a Version with both time axes", func(t *testing.T) {
		store := openFrozen(t, newStore)
		got, err := store.Create(context.Background(), "alice", Employee{Salary: 50000}, hr, crudkiller.Period{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Identity != "alice" {
			t.Fatalf("Identity = %q, want alice", got.Identity)
		}
		if got.Payload.Salary != 50000 {
			t.Fatalf("Payload.Salary = %d, want 50000", got.Payload.Salary)
		}
		if got.Actor.ID != "hr" {
			t.Fatalf("Actor.ID = %q, want hr", got.Actor.ID)
		}
		assertValidNowUnbounded(t, got)
		if !got.TransactionTime.From.Equal(ClockNow) || !got.TransactionTime.Unbounded() {
			t.Fatalf("Transaction time = [%v, %v), want [%v, open)", got.TransactionTime.From, got.TransactionTime.To, ClockNow)
		}
		assertUTC(t, got.ValidTime.From)
		assertUTC(t, got.TransactionTime.From)
	})

	t.Run("Default Period is now to unbounded", func(t *testing.T) {
		store := openFrozen(t, newStore)
		got, err := store.Create(context.Background(), "bob", Employee{Salary: 50000}, hr, crudkiller.Period{})
		if err != nil {
			t.Fatal(err)
		}
		assertValidNowUnbounded(t, got)
	})

	t.Run("Explicit Valid time may lie in the past", func(t *testing.T) {
		store := openFrozen(t, newStore)
		got, err := store.Create(context.Background(), "cara", Employee{Salary: 50000}, hr, crudkiller.Period{From: pastFrom})
		if err != nil {
			t.Fatal(err)
		}
		if !got.ValidTime.From.Equal(pastFrom) || !got.ValidTime.Unbounded() {
			t.Fatalf("Valid time = [%v, %v), want [%v, unbounded)", got.ValidTime.From, got.ValidTime.To, pastFrom)
		}
		if !got.TransactionTime.From.Equal(ClockNow) || !got.TransactionTime.Unbounded() {
			t.Fatalf("Transaction time start = %v, want Clock Now %v and open", got.TransactionTime.From, ClockNow)
		}
	})

	t.Run("Actor is required on Create", func(t *testing.T) {
		store := openFrozen(t, newStore)
		_, err := store.Create(context.Background(), "alice", Employee{Salary: 50000}, crudkiller.Actor{}, crudkiller.Period{})
		if !errors.Is(err, crudkiller.ErrMissingActor) {
			t.Fatalf("err = %v, want ErrMissingActor", err)
		}
		got, err := store.Create(context.Background(), "alice", Employee{Salary: 50000}, hr, crudkiller.Period{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Actor.ID != "hr" {
			t.Fatalf("Actor.ID = %q, library must not supply a default Actor", got.Actor.ID)
		}
	})

	t.Run("Create with a zero Period uses now to unbounded", func(t *testing.T) {
		store := openFrozen(t, newStore)
		for _, id := range []string{"alice", "bob"} {
			got, err := store.Create(context.Background(), id, Employee{Salary: 50000}, hr, crudkiller.Period{})
			if err != nil {
				t.Fatalf("%s: %v", id, err)
			}
			assertValidNowUnbounded(t, got)
		}
	})

	t.Run("Clock is injected and Transaction time is not caller-set", func(t *testing.T) {
		store := openFrozen(t, newStore)
		offset := time.FixedZone("offset", 3600)
		got, err := store.Create(context.Background(), "alice", Employee{Salary: 50000}, hr, crudkiller.Period{From: time.Date(2024, 1, 1, 0, 0, 0, 0, offset)})
		if err != nil {
			t.Fatal(err)
		}
		assertUTC(t, got.ValidTime.From)
		assertUTC(t, got.TransactionTime.From)
		if !got.TransactionTime.From.Equal(ClockNow) {
			t.Fatalf("Transaction time start = %v, want Clock Now", got.TransactionTime.From)
		}
	})

	t.Run("Identity is a non-empty caller-minted string", func(t *testing.T) {
		store := openFrozen(t, newStore)
		got, err := store.Create(context.Background(), "", Employee{Salary: 50000}, hr, crudkiller.Period{})
		if !errors.Is(err, crudkiller.ErrEmptyIdentity) {
			t.Fatalf("err = %v, want ErrEmptyIdentity", err)
		}
		if got.Identity != "" {
			t.Fatalf("Identity = %q, library must not mint one", got.Identity)
		}
		got, err = store.Create(context.Background(), "alice", Employee{Salary: 50000}, hr, crudkiller.Period{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Identity != "alice" {
			t.Fatalf("Identity = %q, want caller-minted alice", got.Identity)
		}
	})

	t.Run("Create of an Identity whose current belief still covers the new Valid time is ErrAlreadyExists", func(t *testing.T) {
		store := openFrozen(t, newStore)
		if _, err := store.Create(context.Background(), "alice", Employee{Salary: 50000}, hr, crudkiller.Period{}); err != nil {
			t.Fatal(err)
		}
		_, err := store.Create(context.Background(), "alice", Employee{Salary: 55000}, hr, crudkiller.Period{})
		if !errors.Is(err, crudkiller.ErrAlreadyExists) {
			t.Fatalf("err = %v, want ErrAlreadyExists", err)
		}

		bounded := openFrozen(t, newStore)
		first := crudkiller.Period{
			From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		next := crudkiller.Period{From: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}
		if _, err := bounded.Create(context.Background(), "alice", Employee{Salary: 1}, hr, first); err != nil {
			t.Fatal(err)
		}
		if _, err := bounded.Create(context.Background(), "alice", Employee{Salary: 2}, hr, next); err != nil {
			t.Fatalf("non-overlapping Valid time: %v", err)
		}
	})

	t.Run("Malformed Period is ErrInvalidPeriod", func(t *testing.T) {
		store := openFrozen(t, newStore)
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		for _, valid := range []crudkiller.Period{{From: from, To: to}, {From: from, To: from}, {To: from}} {
			_, err := store.Create(context.Background(), "alice", Employee{Salary: 50000}, hr, valid)
			if !errors.Is(err, crudkiller.ErrInvalidPeriod) {
				t.Fatalf("period %+v: err = %v, want ErrInvalidPeriod", valid, err)
			}
		}
	})

	t.Run("optional Payload validation may be registered at Store construction", func(t *testing.T) {
		store := openFrozen(t, newStore, crudkiller.WithValidate(func(e Employee) error {
			if e.Salary < 0 {
				return errors.New("salary must be non-negative")
			}
			return nil
		}))
		_, err := store.Create(context.Background(), "alice", Employee{Salary: -1}, hr, crudkiller.Period{})
		if err == nil || !strings.Contains(err.Error(), "salary must be non-negative") {
			t.Fatalf("err = %v, want payload validation error", err)
		}
		if _, err := store.Create(context.Background(), "alice", Employee{Salary: 50000}, hr, crudkiller.Period{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Create records optional Actor reason", func(t *testing.T) {
		store := openFrozen(t, newStore)
		got, err := store.Create(context.Background(), "alice", Employee{Salary: 50000}, crudkiller.Actor{ID: "hr", Reason: "new hire"}, crudkiller.Period{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Actor.Reason != "new hire" {
			t.Fatalf("Actor.Reason = %q, want new hire", got.Actor.Reason)
		}
	})
}
