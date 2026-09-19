# go-crud-killer

An in-process Go library for defining simple entities and performing Create, Read, Edit, and Terminate at scale.

See **[docs/vision.md](docs/vision.md)** for the v1 product vision.

```go
clock := crudkiller.Frozen(time.Date(2024, 4, 15, 12, 0, 0, 0, time.UTC))
store, err := crudkiller.New(clock, memory.New[Employee]())
version, err := store.Create(ctx, "alice", Employee{Salary: 50000}, crudkiller.Actor{ID: "hr"}, crudkiller.Period{})
```

`Create` records a bitemporal Version: caller-minted Identity, Payload, Actor, Valid time, and Transaction time from the injected Clock (UTC). A zero Period is `[now, unbounded)`. Persistence is a Recording port; this module ships a memory adapter.

