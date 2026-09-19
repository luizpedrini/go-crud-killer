package memory_test

import (
	"testing"

	crudkiller "github.com/luizpedrini/go-crud-killer"
	"github.com/luizpedrini/go-crud-killer/memory"
	"github.com/luizpedrini/go-crud-killer/storetest"
)

func TestCreate(t *testing.T) {
	storetest.Create(t, func(clock crudkiller.Clock, opts ...crudkiller.Option[storetest.Employee]) (crudkiller.Store[storetest.Employee], error) {
		return crudkiller.New(clock, memory.New[storetest.Employee](), opts...)
	})
}
