package crudkiller

import "time"

// Clock reports "now" for Valid-time defaults and Transaction time.
// Callers never set Transaction time; adapters may read a datastore timestamp
// only as this clock's reading.
type Clock interface {
	Now() time.Time
}

// Frozen returns a Clock that always reports t in UTC.
func Frozen(t time.Time) Clock {
	return frozenClock{at: t.UTC()}
}

// Wall returns a Clock that reports time.Now in UTC.
func Wall() Clock {
	return wallClock{}
}

type frozenClock struct{ at time.Time }

func (c frozenClock) Now() time.Time { return c.at }

type wallClock struct{}

func (wallClock) Now() time.Time { return time.Now().UTC() }
