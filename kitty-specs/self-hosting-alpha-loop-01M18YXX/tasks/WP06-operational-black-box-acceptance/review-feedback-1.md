# WP06 Review Cycle 1 — REJECTED

The operational scenario, repeatability run, and repository-wide quality gates
pass. The WP05 arbiter repair is nevertheless not complete: loss of an entire
validated v2 transaction receipt erases its denial state, and the receipt
directory is not protected by the same no-symlink/private-permission contract
as the receipt files. The claimed crash fixture also creates the partial state
in the test process rather than through the compiled `nh` process.

## Blocking issue 1 — a missing validated v2 receipt fails open

`loadReplicationAcceptanceState` derives the pending-denial union solely by
scanning whichever `transactions/*.json` files still exist. If a validated
receipt disappears after object copy, accepted-ref promotion, and shallow
marker release, `loadReplicationTransactionRecords` sees no record and returns
an empty pending set. The accepted projection then reads the promoted event.
There is no durable envelope, index, journal head, or other independent state
from which the missing receipt can be detected.

This is not merely uncovered. `TestOperationalSelfHostingAlphaAcceptanceLedgerUnion`
deletes `txn-union-a.json` and explicitly asserts that the transaction's unique
object is no longer denied. That expected behavior contradicts the mandatory
WP05 cycle-3 handoff: disappearance of a pending record MUST NOT be interpreted
as acceptance and one operation must not erase denial without a durable
completion.

A removed review-only probe reproduced the real consequence:

1. create the existing depth-one selected-recovery fixture;
2. inject failure of the `complete` receipt after ref promotion and successful
   marker release;
3. delete the sole validated v2 receipt;
4. start a fresh compiled `nh log` process.

The fresh process exited zero and printed both the recovered predecessor and
successor events. Expected behavior was a non-zero `replication acceptance
pending` result. This violates the selected-replication acceptance boundary,
FR-010, FR-018, FR-019, NFR-005, and NFR-009.

### Required repair

- Add an independently durable transaction-envelope/index/journal invariant
  written before promotion, so loss of a referenced validated receipt fails
  closed until an atomic durable completion proves acceptance. A directory
  listing of surviving receipts is not sufficient authority.
- Never treat receipt deletion as denial clearance. Replace the current union
  test's fail-open assertion with a fail-closed missing-receipt assertion while
  retaining the valid concurrent-union/shared-object/pre-existing-object
  cases.
- Test an entirely absent v2 receipt, not only an existing v2 JSON document
  with missing fields or malformed syntax.
- Reproduce the post-promotion + post-marker-release + pre-completion boundary
  by running the compiled binary with a narrowly guarded test seam. Kill or
  fail that process at the boundary, then use separate fresh processes for
  `log`, policy, proposal, review, run, decision, merge, and recovery retry.
  The current test substitutes the package-global
  `replicationRecordTransaction` and calls `recoverSelectedShallow` in the Go
  test process; only its later probes and retry are compiled-process calls.
- Prove the retry reconciles the durable envelope and receipt atomically and
  that no accepted denial can be erased by another incomplete transaction.

## Blocking issue 2 — receipt directory safety is not enforced

`loadReplicationTransactionRecords` calls `os.ReadDir` directly on
`.git/nh/replication/transactions`. It does not `Lstat` and validate that path
as a real owner-only directory. Consequently it follows a directory symlink
and accepts a directory with mode `0755`. Individual receipt files do receive
regular-file, symlink, size, permission, and strict-JSON checks through
`readPrivateFile`/`decodePrivateJSON`, but those checks do not protect the
directory lookup boundary.

Removed review-only subtests confirmed both failures:

- replacing `transactions` with a symlink to another directory returned no
  error;
- changing `transactions` to mode `0755` returned no error.

### Required repair

- Validate the receipt directory (and the relevant private parent chain) with
  the established `Lstat`/real-directory/owner-only rules before listing it.
- Add production-path tests for directory symlink, non-directory, unsafe mode,
  oversized receipt, receipt-file symlink, unsafe file mode, unknown/trailing
  JSON, malformed JSON, missing v2 fields, and entirely missing referenced v2
  receipt. All unsafe states must return the stable fail-closed acceptance
  error without exposing pending facts.

## Accepted evidence

- The offline black-box scenario uses compiled `nh` subprocesses and ordinary
  Git repositories/bare remotes for policy amendment, distinct actors,
  rotation retry, remote cycle ambiguity, hostile selected replication,
  shallow recovery, role-distinct governance, and identity-free
  reconstruction.
- The old base policy exclusively authorizes its amendment; the later policy
  requires non-author reviewer/runner evidence.
- Valid, invalid, mismatched, unselected, over-budget, and missing-dependency
  paths are externally asserted. Lower-level production transaction tests
  retain the one-below/exact/one-above matrix for all five budget dimensions.
- The rotation interruption hook requires both `NH_INTERNAL_TESTING=1` and the
  exact step variable, does not emit private key material, and leaves ordinary
  semantics unchanged when absent.
- Approved dependency commits are present in the integration lane. The
  out-of-map `store.go`, `shallow.go`, and `quarantine.go` changes are within
  the recorded WP05 arbiter handoff; the narrow command/policy/identity/CI
  integration has a review-visible rationale.

## Verification evidence

- `go test ./... -run TestOperationalSelfHostingAlpha -count=1`: PASS (13.0s)
- `go test ./... -run TestOperationalSelfHostingAlpha -count=3`: PASS (40.1s)
- `go test ./...`: PASS (35.0s)
- `go test -race ./...`: PASS (38.4s)
- `go vet ./...`: PASS
- `go build ./...`: PASS
- `gofmt -l` over the WP commit's Go files: PASS (no output)
- `git diff --check 8eca178^..8eca178`: PASS
- Review-only missing-receipt fresh-process probe: FAIL as described above;
  removed before verdict capture.
- Review-only symlink/unsafe-directory probes: FAIL as described above;
  removed before verdict capture.

## Anti-pattern checklist

1. **Dead code — PASS.** New production helpers have live production callers.
2. **Synthetic-fixture test — FAIL.** The mandatory crash state is injected
   in-process while the WP claims compiled-process crash reproduction, and the
   entirely missing receipt case is absent.
3. **Silent empty return — FAIL.** An absent transaction directory/receipt is
   reduced to an empty pending set even after a transaction may have reached
   the validated/promotion boundary.
4. **FR coverage — FAIL.** FR-010/018/019 are not satisfied across complete
   validated-receipt loss; the fresh process admits the pending facts.
5. **Frozen surface — N/A.** No frozen file was identified for WP06.
6. **Locked decision — FAIL.** Pending imported state can become accepted
   merely because its validated receipt disappeared, contrary to the explicit
   fail-closed transaction contract and arbiter handoff.
7. **Shared-file ownership — PASS.** The arbiter and integration handoffs are
   recorded in the WP activity log and commit rationale.
8. **Production fragility — FAIL.** Correctness currently depends on the sole
   validated receipt and its containing directory never disappearing or being
   replaced after promotion.
