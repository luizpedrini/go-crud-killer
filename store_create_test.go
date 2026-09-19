package crudkiller_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	crudkiller "github.com/luizpedrini/go-crud-killer"
	"github.com/luizpedrini/go-crud-killer/memory"
)

type employee struct {
	Salary int
}

var frozenNow = time.Date(2024, 4, 15, 12, 0, 0, 0, time.UTC)

func newEmployeeStore(t *testing.T, opts ...crudkiller.Option[employee]) *crudkiller.Store[employee] {
	t.Helper()
	store, err := crudkiller.New(crudkiller.Frozen(frozenNow), memory.New[employee](), opts...)
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

func TestCreateRecordsVersionWithBothTimeAxes(t *testing.T) {
	store := newEmployeeStore(t)
	got, err := store.Create(context.Background(), "alice", employee{Salary: 50000}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
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
	if !got.ValidTime.From.Equal(frozenNow) || !got.ValidTime.Unbounded() {
		t.Fatalf("ValidTime = [%v, %v), want [%v, unbounded)", got.ValidTime.From, got.ValidTime.To, frozenNow)
	}
	if !got.TransactionTime.From.Equal(frozenNow) || !got.TransactionTime.Unbounded() {
		t.Fatalf("TransactionTime = [%v, %v), want [%v, open)", got.TransactionTime.From, got.TransactionTime.To, frozenNow)
	}
	assertUTC(t, got.ValidTime.From)
	assertUTC(t, got.TransactionTime.From)
}

func TestCreateZeroPeriodIsNowToUnbounded(t *testing.T) {
	store := newEmployeeStore(t)
	for _, id := range []string{"alice", "bob"} {
		got, err := store.Create(context.Background(), id, employee{Salary: 1}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if !got.ValidTime.From.Equal(frozenNow) {
			t.Fatalf("%s ValidTime.From = %v, want Clock Now %v", id, got.ValidTime.From, frozenNow)
		}
		if !got.ValidTime.Unbounded() {
			t.Fatalf("%s ValidTime.To = %v, want unbounded", id, got.ValidTime.To)
		}
	}
}

func TestCreateExplicitValidTimeMayLieInThePast(t *testing.T) {
	store := newEmployeeStore(t)
	valid := crudkiller.Period{From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	got, err := store.Create(context.Background(), "cara", employee{Salary: 40000}, crudkiller.Actor{ID: "hr"}, valid)
	if err != nil {
		t.Fatal(err)
	}
	if !got.ValidTime.From.Equal(valid.From) || !got.ValidTime.Unbounded() {
		t.Fatalf("ValidTime = [%v, %v), want [%v, unbounded)", got.ValidTime.From, got.ValidTime.To, valid.From)
	}
	if !got.TransactionTime.From.Equal(frozenNow) || !got.TransactionTime.Unbounded() {
		t.Fatalf("TransactionTime.From = %v, want Clock Now %v and open", got.TransactionTime.From, frozenNow)
	}
}

func TestCreateRequiresActorID(t *testing.T) {
	store := newEmployeeStore(t)
	_, err := store.Create(context.Background(), "alice", employee{Salary: 50000}, crudkiller.Actor{}, crudkiller.Period{})
	if !errors.Is(err, crudkiller.ErrMissingActor) {
		t.Fatalf("err = %v, want ErrMissingActor", err)
	}

	_, err = store.Create(context.Background(), "alice", employee{Salary: 50000}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateRejectsEmptyIdentity(t *testing.T) {
	store := newEmployeeStore(t)
	got, err := store.Create(context.Background(), "", employee{Salary: 50000}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
	if err == nil {
		t.Fatalf("Create minted Identity %q, want an error", got.Identity)
	}
	if got.Identity != "" {
		t.Fatalf("Identity = %q, library must not mint one", got.Identity)
	}
}

func TestCreateMalformedPeriodIsErrInvalidPeriod(t *testing.T) {
	store := newEmployeeStore(t)
	from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []crudkiller.Period{
		{From: from, To: to},
		{From: from, To: from},
		{To: from},
	}
	for _, valid := range cases {
		_, err := store.Create(context.Background(), "alice", employee{Salary: 1}, crudkiller.Actor{ID: "hr"}, valid)
		if !errors.Is(err, crudkiller.ErrInvalidPeriod) {
			t.Fatalf("period %+v: err = %v, want ErrInvalidPeriod", valid, err)
		}
	}
}

func TestCreateOverlappingCurrentBeliefIsErrAlreadyExists(t *testing.T) {
	store := newEmployeeStore(t)
	_, err := store.Create(context.Background(), "alice", employee{Salary: 50000}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Create(context.Background(), "alice", employee{Salary: 55000}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
	if !errors.Is(err, crudkiller.ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestCreateNonOverlappingValidLivesAreAllowed(t *testing.T) {
	store := newEmployeeStore(t)
	first := crudkiller.Period{
		From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	second := crudkiller.Period{
		From: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if _, err := store.Create(context.Background(), "alice", employee{Salary: 1}, crudkiller.Actor{ID: "hr"}, first); err != nil {
		t.Fatal(err)
	}
	got, err := store.Create(context.Background(), "alice", employee{Salary: 2}, crudkiller.Actor{ID: "hr"}, second)
	if err != nil {
		t.Fatal(err)
	}
	if got.Payload.Salary != 2 {
		t.Fatalf("Payload.Salary = %d, want 2", got.Payload.Salary)
	}
}

func TestCreateOptionalPayloadValidation(t *testing.T) {
	store := newEmployeeStore(t, crudkiller.WithValidate(func(e employee) error {
		if e.Salary < 0 {
			return errors.New("salary must be non-negative")
		}
		return nil
	}))
	_, err := store.Create(context.Background(), "alice", employee{Salary: -1}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
	if err == nil || !strings.Contains(err.Error(), "salary must be non-negative") {
		t.Fatalf("err = %v, want payload validation error", err)
	}
	if _, err := store.Create(context.Background(), "alice", employee{Salary: 1}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{}); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRecordsOptionalActorReason(t *testing.T) {
	store := newEmployeeStore(t)
	got, err := store.Create(context.Background(), "alice", employee{Salary: 1}, crudkiller.Actor{ID: "hr", Reason: "new hire"}, crudkiller.Period{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Actor.Reason != "new hire" {
		t.Fatalf("Actor.Reason = %q, want new hire", got.Actor.Reason)
	}
}

func TestCreateTimestampsAreUTCAndTxComesFromClock(t *testing.T) {
	store := newEmployeeStore(t)
	valid := crudkiller.Period{From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("offset", 3600))}
	got, err := store.Create(context.Background(), "alice", employee{Salary: 1}, crudkiller.Actor{ID: "hr"}, valid)
	if err != nil {
		t.Fatal(err)
	}
	assertUTC(t, got.ValidTime.From)
	assertUTC(t, got.TransactionTime.From)
	if !got.TransactionTime.From.Equal(frozenNow) {
		t.Fatalf("TransactionTime.From = %v, want Clock Now", got.TransactionTime.From)
	}
}

func TestCreateConcurrentSameIdentityOneAlreadyExists(t *testing.T) {
	store := newEmployeeStore(t)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Create(context.Background(), "alice", employee{Salary: 1}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	var ok, already int
	for err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, crudkiller.ErrAlreadyExists):
			already++
		default:
			t.Fatalf("unexpected err %v", err)
		}
	}
	if ok != 1 || already != 1 {
		t.Fatalf("got %d success and %d ErrAlreadyExists, want 1 and 1", ok, already)
	}
}
