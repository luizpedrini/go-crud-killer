package crudkiller

// Actor is who caused a Version to be recorded. ID is required on Create,
// Edit, and Terminate. Reason is optional. The library never invents one.
type Actor struct {
	ID     string
	Reason string
}

func (a Actor) missingID() bool {
	return a.ID == ""
}

// Version is one recording of a Payload over a Valid-time period, believed
// during a Transaction-time period.
type Version[T any] struct {
	Identity        string
	Payload         T
	Actor           Actor
	ValidTime       Period
	TransactionTime Period
}

func (v Version[T]) currentlyBelieved() bool {
	return v.TransactionTime.Unbounded()
}

func (v Version[T]) currentBeliefOverlaps(valid Period) bool {
	return v.currentlyBelieved() && v.ValidTime.overlaps(valid)
}
