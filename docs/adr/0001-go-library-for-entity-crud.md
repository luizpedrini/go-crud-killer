# This is a Go library for entity CRUD

The module is `github.com/luizpedrini/go-crud-killer` (Go 1.23+). Callers define entities and perform Create, Read, Edit, and Terminate in-process. v1 is library-only: no CLI and no HTTP tool. A tool can sit on the library later without becoming the product.

## Status

accepted

## Considered Options

- **Go library only in v1 (chosen).** Matches the product brief. The GitHub description still says "an opiniated tool"; that description should follow this decision.
- **CLI / HTTP tool as the product, or shipped alongside in v1.** Useful later as an adapter. Shipping it now would split the interface: we would design a process tool instead of a deep in-process module.
