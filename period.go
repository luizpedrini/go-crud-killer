package crudkiller

import (
	"fmt"
	"time"
)

// Period is a half-open interval [From, To). A zero To is unbounded.
// A zero Period (both ends zero) on a mutation means [clock.Now(), unbounded).
type Period struct {
	From time.Time
	To   time.Time
}

// Unbounded reports whether To is open-ended.
func (p Period) Unbounded() bool {
	return p.To.IsZero()
}

// IsZero reports whether both ends are the zero time.
func (p Period) IsZero() bool {
	return p.From.IsZero() && p.To.IsZero()
}

// Overlaps reports whether the half-open intervals share any instant.
func (p Period) Overlaps(q Period) bool {
	pBeforeQEnd := q.To.IsZero() || p.From.Before(q.To)
	qBeforePEnd := p.To.IsZero() || q.From.Before(p.To)
	return pBeforeQEnd && qBeforePEnd
}

func (p Period) utc() Period {
	if !p.From.IsZero() {
		p.From = p.From.UTC()
	}
	if !p.To.IsZero() {
		p.To = p.To.UTC()
	}
	return p
}

func (p Period) wellFormed() error {
	if p.From.IsZero() {
		return fmt.Errorf("from is required")
	}
	if !p.To.IsZero() && !p.From.Before(p.To) {
		return fmt.Errorf("from must be before to")
	}
	return nil
}
