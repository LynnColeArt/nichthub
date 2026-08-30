# Selected Replication Contract v0

## Purpose

Replace unconditional direct wildcard import with an explicit per-remote
selection and a validate-before-promotion transaction while preserving an
opt-in compatibility-all mode.

## Local selection

```text
nh replication select [REMOTE] [--actor ACTOR]... [--proposal ID]...
  [--all]
  [--max-events N]
  [--max-objects N]
  [--max-object-bytes N]
  [--max-attachment-bytes N]
  [--max-total-bytes N]

nh replication show [REMOTE]
```

The remote defaults to `origin`. Actor and proposal selectors must be full IDs.
Every budget is a positive integer. `--all` is mutually exclusive with actor
and proposal selectors. Configuration is stored below `.git/nh` and is never
committed or published.

`show` reports the exact selected IDs, budgets, and whether compatibility-all
mode is active.

## Synchronization

```text
nh sync [REMOTE] [--recover-shallow]
```

1. Resolve the saved selection; when none exists, use documented bounded
   compatibility-all behavior for version-0 repositories.
2. Request only exact selected actor and proposal refs, or advertised wildcard
   refs only when compatibility-all was explicit/defaulted.
3. Fetch into a generated separate bare quarantine repository.
4. Measure each selected reachable graph and reject a selection exceeding any
   configured budget before promotion.
5. Validate Git object types, event trees, exact signatures, actor chains,
   attachments, candidate ref/head bindings, identity relations, and existing
   event relationships.
6. Classify absent cross-selection facts as missing dependencies with full IDs.
7. Copy promotable objects to the main object database.
8. Atomically update accepted refs under:

```text
refs/nh/remotes/<remote>/actors/<actor>
refs/nh/remotes/<remote>/proposals/<candidate-hash>
```

9. Publish only the active local actor ref and locally authored proposal refs
   allowed by the existing synchronization rules.

## Isolation

- An invalid, over-budget, or dependency-missing selection never advances its
  accepted remote-tracking ref.
- A valid selection whose relationship closure does not depend on a failed
  selection may promote independently.
- Promotion of one consistent set uses one atomic ref transaction; failure
  leaves the previous accepted ref values intact.
- Objects from a rejected quarantine are not accepted projection roots. The
  quarantine is recoverable or disposable local state, not protocol history.
- `collectEvents` reads only local actor refs and accepted remote-tracking refs.

## Budgets

Budgets are measured over objects newly reachable for each selected ref and
over its parsed event chain:

- actor event count;
- reachable object count;
- largest individual Git object bytes;
- largest event attachment bytes;
- total reachable bytes.

Tests cover one below, exactly at, and one above every configured value.
Standard Git may download a selected pack before Nichthub can measure it; the
budget is a hard promotion/retention boundary, not a portable hard network byte
quota. Direct refspecs ensure unselected histories are not requested.

## Missing dependencies

Missing dependency is distinct from invalid data. Diagnostics include:

- failing selected actor/candidate full ID;
- referencing event full ID and kind;
- missing event/object/candidate/actor full ID when derivable;
- exact additional actor or proposal selection, or selected shallow-recovery
  action needed to retry.

No command substitutes a similarly titled proposal, related identity actor,
predecessor, successor, or shortened ID.

## Shallow repositories

Shallow status is detected, but operations succeed when all exact dependencies
already exist. When a required object is absent beyond the shallow boundary,
trust-sensitive commands fail with its exact ID. `--recover-shallow` explicitly
fetches only the selected supplying refs through quarantine under the same
budgets and restarts verification. It never issues a global unshallow or adds
an actor/candidate selection silently.

## Compatibility

- Existing valid event payloads and IDs do not change.
- Existing `nh sync [REMOTE]` syntax remains accepted.
- Once a remote has an explicit saved selection, that selection is
  authoritative until explicitly replaced.
- Compatibility-all mode faces the same quarantine, validation, and budgets;
  it does not restore direct unvalidated wildcard fetch.
