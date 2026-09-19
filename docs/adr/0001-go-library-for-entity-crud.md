# This is a Go library for entity CRUD

The repo is `github.com/luizpedrini/go-crud-killer`. Callers define simple entities and perform Create, Read, Edit, and Terminate in-process through a Go module, not through a hosted service or a CLI as the primary product.

## Status

accepted

## Considered Options

- **Go library (chosen).** Matches the product brief. The GitHub repo description currently says "an opiniated tool"; that description should follow the library decision, not the other way around.
- **CLI / HTTP tool as the product.** Useful later as an adapter on top of the library. Making the tool the core would hide the entity-definition and CRUD interface that callers need to embed.
