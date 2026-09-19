# Mutations require an Actor; the library never invents one

Create, Edit, and Terminate take an Actor. Actor ID is required. Reads do not. There is no ambient "system" default. A nullable `updated_by` is how Audit goes missing; requiring Actor on the interface makes that impossible to forget at a call site.

## Status

accepted
