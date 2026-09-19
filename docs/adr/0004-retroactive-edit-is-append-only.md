# Edit may be retroactive: split Valid time, never rewrite Transaction time

An Edit may supply a Valid-time period that overlaps the past. Overlap is resolved by splitting or closing Valid-time portions on *new* Transaction-time recordings. Old Transaction-time recordings stay immutable, which is what lets Read answer "what did we believe in April that March was."

## Status

accepted

## Considered Options

- **Current-only Edit.** Simpler CRUD. Makes Valid time decoration: a later correction silently rewrites the only timeline.
- **Retroactive Edit with in-place rewrite of old recordings.** Looks like a temporal table, but forges prior belief.
- **Retroactive Edit that only appends (chosen).** The costly option, and the one that makes bitemporal Audit real.
