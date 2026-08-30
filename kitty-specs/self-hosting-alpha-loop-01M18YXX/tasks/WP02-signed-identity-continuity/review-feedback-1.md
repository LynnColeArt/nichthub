## WP02 review feedback — cycle 1

### 1. A failure after the target record write creates a sibling key on retry

`prepareIdentityRotation` calls `storeIdentityRecord(target, ...)` before it
calls `storeIdentityRotation(state)`. If the record write succeeds but the
rotation-state write fails, the target key is durable while no transaction is
discoverable. The next `nh identity rotate` sees `rotation.json` absent,
generates a different key, and leaves the first target record orphaned.

This violates T004/T009's explicit `target key durable` recovery state and the
requirement that retries reuse recorded keys and event IDs rather than create
siblings. Make this boundary recoverable and add a narrow failure seam/test
that proves: the predecessor remains active; retry selects the original
target actor; exactly two continuity facts result; and no extra keyring record
or actor history is created.

### 2. Command-level retry after durable rotation cleanup starts another rotation

WP01 makes `switchActiveIdentity` converge after both cleanup boundaries when
called with the original transaction state. `cmdIdentityRotate` loses that
ability: after the target pointer is switched and `rotation.json` is removed,
an error from directory synchronization (or the existing
`after-rotation-cleanup` seam) is returned to the operator. Retrying the CLI
calls `prepareIdentityRotation`, finds no transaction, treats the already
active successor as a new predecessor, and creates another target and another
authorization/acceptance pair.

Extend the CLI-level recovery design so a reported cleanup-boundary failure
cannot turn a retry into a second rotation. Add command-level tests for
`before-rotation-cleanup` and the durable-removal/post-cleanup boundary,
asserting stable actor IDs, record count, event IDs, and active signer across
repeated retries.

## Verification and review evidence

The following gates pass:

- `go test ./... -run 'Test.*(IdentityContinuity|IdentityAuthorize|IdentityAccept|Rotation)'`
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `gofmt -d` on all WP02-touched Go files is empty.

The remaining reviewed behavior passes inspection:

- authorization and acceptance use real Ed25519 signing and exact target-key
  validation;
- projection verifies signed payloads, exposes pending/accepted/missing/
  competing/cyclic states, and is stable under shuffled input;
- existing event JSON and IDs remain unchanged through omitted new fields and
  golden checks for every prior kind;
- continuity does not modify policy bytes or grant reviewer, runner,
  maintainer, decision, or merge authority;
- the `commands.go` and `store.go` ownership exceptions are narrow and
  explicitly justified in commit `0935e97`; WP04 must inherit/rebase them.

## WP anti-pattern checklist

1. Dead code: **PASS** — every command handler is routed through production
   `cmdIdentity`, and the projector is used by production identity listing.
2. Synthetic-fixture test: **PASS** — tests sign through `encodeAndSign`, append
   through the real Git store where integration matters, and exercise real
   policy evaluation.
3. Silent empty return: **PASS** — no swallowed-error or unexplained empty
   production return was found.
4. FR coverage: **FAIL** — FR-018/recoverable rotation does not cover or
   satisfy the target-record/state-write and post-cleanup CLI boundaries.
5. Frozen surface: **PASS** — no frozen/untouchable surface was modified.
6. Locked decision: **PASS** — no private-key copying, name-based identity,
   order-based winner, or implicit policy authority was introduced.
7. Shared-file ownership: **PASS** — the `commands.go`/`store.go` exceptions
   are scoped and justified; downstream WP04 owns their later evolution.
8. Production fragility: **N/A** — no exception-style/bare-raise path exists
   in this Go change.
