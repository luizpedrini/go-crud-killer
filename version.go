package crudkiller

// Actor is who caused a Version to be recorded. ID is required on Create,
// Edit, and Terminate. Reason is optional. The library never invents one.
type Actor struct {
	ID     string
	Reason string
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
