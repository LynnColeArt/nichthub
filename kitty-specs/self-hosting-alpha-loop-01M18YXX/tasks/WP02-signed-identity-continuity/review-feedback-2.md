## WP02 review feedback — cycle 2

### Blocking: durable event IDs are trusted on the ordinary resume path

The correction adds `validateCompletedIdentityRotation`, but
`prepareIdentityRotation` calls it only when the command journal exists,
`rotation.json` is absent, and the target is already active. The more common
resume path where both `rotation-command.json` and `rotation.json` exist only
calls `reconcileIdentityRotationStates`. If their authorization and acceptance
fields are non-empty, `cmdIdentityRotate` skips both append/find operations and
passes the syntactically valid IDs directly to `switchActiveIdentity`.

Consequently, corrupted private state can contain two arbitrary full-looking
`sha256:` IDs, a valid predecessor/target pair, and no matching signed events;
the next rotation command will switch the active actor and retire both files.
Neither ID is resolved, signature-verified, nor checked for the exact
authorization/acceptance relationship on that path. This violates the
two-sided-signature contract, C-005, and the WP01 boundary that WP02 must prove
event facts before requesting an active-pointer switch.

Before treating either persisted event ID as durable progress—and always
before `switchActiveIdentity`—resolve it from verified actor histories and
validate the exact predecessor, target actor/key, `successor` relationship,
acceptance signer, and acceptance subject. Missing or mismatched facts must
fail closed without changing the active signer or deleting recovery state.
Add hostile-state tests for:

- two nonexistent but well-formed event IDs in both journal and
  `rotation.json`;
- an ID naming a valid event of the wrong kind or wrong rotation;
- one valid authorization paired with an unrelated acceptance; and
- a valid completed pair, proving the resumed switch still succeeds
  idempotently.

### Prior feedback verification

Both cycle-one crash windows are otherwise fixed:

- the owner-only atomic command journal is written and directory-synced before
  the target record, so failure after target durability reuses the exact key;
- pre-cleanup and post-cleanup retries preserve actor records and signed event
  IDs and do not create sibling rotations;
- `rotation-command.json` uses the canonical `0700` directory, `0600` atomic
  file writer, size bound, strict JSON decode, keypair validation, and
  symlink/permission checks;
- successful completion removes both the command journal and WP01's
  `rotation.json`; the latter remains incomplete-only.

The documented best-effort directory sync after command-receipt deletion does
not create sibling rotations: a surviving receipt is revalidated and retired
on the next invocation.

### Quality gates

All requested mechanical gates pass:

- focused identity/continuity/rotation tests;
- `go test ./...`;
- `go test -race ./...`;
- `go vet ./...`;
- `go build ./...`;
- `gofmt -d` is empty;
- correction diff check is clean.

### WP anti-pattern checklist

1. Dead code: **PASS** — journal and recovery functions are reached by the
   production rotation command.
2. Synthetic-fixture test: **PASS** — correction tests exercise real keyring,
   Git event, signing, and command paths.
3. Silent empty return: **PASS** — the one ignored directory-sync result has
   an explicit fail-safe rationale and cannot turn retry into a sibling.
4. FR coverage: **FAIL** — FR-006/FR-018 do not fail closed when persisted
   event IDs are missing or unrelated.
5. Frozen surface: **PASS** — no frozen/untouchable file changed.
6. Locked decision: **FAIL** — the ordinary resume path can switch actors
   without proving participation by both predecessor and successor keys.
7. Shared-file ownership: **PASS** — cycle-two changes remain inside WP02's
   continuity implementation and tests; prior integration exceptions remain
   documented.
8. Production fragility: **N/A** — no exception-style/bare-raise path exists
   in this Go change.
