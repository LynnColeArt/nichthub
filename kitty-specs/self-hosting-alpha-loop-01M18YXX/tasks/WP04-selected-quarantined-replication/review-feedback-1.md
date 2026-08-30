# WP04 review feedback — cycle 1

## Verdict

**REJECTED.** The main isolation architecture is present and the required Go
quality gates pass, but quarantine currently accepts one malformed governance
relationship, fatal transport diagnostics disclose private remote arguments,
and the required hostile-input/budget acceptance matrix is not exercised
through the real transaction path.

## Blocking issues

### 1. A merge fact with the wrong candidate head can cross quarantine

`replicationEventDependency` routes `proposal.merged` directly to
`validateReplicationMergeAt` (`quarantine.go:769-783`). That validator checks
policy and acceptance evidence, but never checks that the merge event's `Head`
equals the referenced proposal's immutable `Head` (`quarantine.go:900-920`).
The canonical accepted-event validator does enforce this invariant
(`store.go:328-334`). The two validation surfaces have therefore drifted.

An adversarial test using a real temporary Git repository, real policy commit,
signed proposal, signed acceptance decision, and signed merge event reproduced
the defect: `replicationEventDependency` returned no dependency and no error
when `merge.Event.Head` was deliberately set to the base commit rather than the
proposal head. Such an actor selection can be marked promotable; after
promotion, normal `collectEvents` rejects the now-accepted history and the
projection is poisoned.

Fix this by making quarantine enforce the canonical complete relationship
contract (preferably through one reusable authority rather than another
partial copy). Add a real-repository selected-sync regression proving the bad
actor remains absent or at its old accepted ref while an unrelated valid actor
still promotes.

### 2. Fatal transport errors disclose credentials and host-private paths

`gitInputWithDirectory` includes every raw Git argument in returned errors
(`git.go:44-60`). `resolveReplicationRequests` passes the credential-bearing or
host-private `remoteURL` to that wrapper and returns its error without
redaction (`quarantine.go:407-410`). Other setup/recording failures can likewise
surface absolute quarantine or `.git/nh` paths before the outcome-level
redaction helper is reached.

The adversarial reproduction passed
`file:///definitely-missing/review-user:REVIEW_SECRET/repo.git` as the remote.
The returned user-facing error contained both the full URL and
`REVIEW_SECRET`, including the absolute path echoed by Git.

Introduce a safe Git-error boundary for replication so user-facing failures
never contain a remote URL, credentials, quarantine path, Git directory, or
transaction-record path. Preserve useful phase and exact selection IDs. Add
tests for advertisement/fetch/setup failures with credential-like URLs and
host-private paths.

### 3. The required budget and hostile-input acceptance coverage is incomplete

`TestReplicationBudgetBoundaries` measures one actor once, then calls
`enforceReplicationBudgets` directly with a preconstructed
`ReplicationMeasurements` value (`replication_test.go:112-190`). It never runs
`runReplicationTransaction` or `cmdSync`, never proves that measurement is
performed per selection, and never inspects accepted refs or object
availability at one-below/equal/one-above for any of the five budgets. Deleting
the transaction's budget-enforcement calls would leave this test green.

The WP prompt also requires hostile real-repository cases for exact signature,
actor chain, event-tree/attachment, governance, and identity-continuity
relationships, plus cleanup-failure non-promotion. The only mixed transaction
case covers a non-event commit, a candidate ref/head mismatch, and a missing
issue dependency; it does not exercise those other required validators.

Add transaction-level boundary tests for event count, reachable object count,
largest object bytes, largest attachment bytes, and total reachable bytes at
all three boundaries, asserting accepted refs and destination object
availability after every result. Add non-vacuous hostile cases for the listed
validation layers and a cleanup failure; each rejected selection must preserve
its prior accepted ref while an unrelated valid selection remains usable.

### 4. Completion-record failure is silently discarded after promotion

After atomic ref promotion, the implementation ignores the result of writing
the final `complete` transaction record (`quarantine.go:376-383`). A full disk,
permission change, or state-file failure can therefore leave a durable record
at `validated` while the command reports success and accepted refs have moved.
This contradicts the required local transaction-state/outcome record and makes
recovery diagnostics ambiguous.

Handle this failure explicitly and report the truthful state (promotion
succeeded, completion recording failed); do not imply rollback. Add a focused
failure-injection test for this post-promotion boundary.

## Subtask disposition

- T016: **REJECTED** — required hostile-input and transaction-level budget
  boundary matrix is incomplete.
- T017: **APPROVED** — full-ID selection, validation, deterministic storage,
  permissions, safe encoded filenames, and bounded compatibility defaults are
  implemented.
- T018: **REJECTED** — separate bare quarantine and exact refspecs are present,
  but fatal diagnostics leak private transport/path material and cleanup
  failure lacks the required regression proof.
- T019: **REJECTED** — production measurements exist, but the mandated real
  below/equal/above promotion tests are absent.
- T020: **REJECTED** — merge-head relationship validation is incomplete and
  can admit an actor history rejected by canonical projection.
- T021: **APPROVED WITH FOLLOW-UP TESTS REQUIRED** — object import precedes one
  expected-old atomic ref transaction and stale refs are preserved; the
  cleanup and final-record failure boundaries still need tests/fixes above.
- T022: **APPROVED** — sync routes import through quarantine, publication is a
  separate phase, unsupported shallow recovery is explicit, and identity-free
  import is supported.

## Verification evidence

- `gofmt -l` on all Go files changed by `fc6a42d`: clean.
- `go test ./... -run 'Test.*(Replication|Selection|Quarantine|Budget|Sync)'`:
  pass.
- `go test ./...`: pass.
- `go test -race ./...`: pass.
- `go vet ./...`: pass.
- `go build ./...`: pass.
- `git diff --check fc6a42d^..fc6a42d`: pass.
- Source audit found no direct fetch destination under `refs/nh/remotes`, no
  quarantine alternate, and no shell-string command execution.
- Two temporary adversarial review tests failed exactly on issues 1 and 2;
  both were removed after reproduction and no product code was modified.

## Anti-pattern checklist

1. Dead code: **PASS** — production calls exist for the transaction surfaces;
   the prompt explicitly hands the `cmdReplication` top-level route to WP06.
2. Synthetic-fixture test: **FAIL** — the budget boundary test asserts a
   prebuilt measurement rather than exercising the production sync path.
3. Silent empty return: **PASS** — no new unexplained empty-return pattern was
   found. The separately ignored completion-record error is issue 4.
4. FR coverage: **FAIL** — FR-010/FR-012 validation and budget boundaries lack
   the required non-vacuous acceptance coverage.
5. Frozen surface: **PASS** — no frozen/untouchable file was changed.
6. Locked decision: **FAIL** — the selected-replication contract requires all
   existing relationships and candidate heads to validate before promotion;
   the merge-head mismatch violates that MUST-level boundary.
7. Shared-file ownership: **PASS** — the feature commit stays within WP04's
   ownership map, and WP02 review feedback explicitly hands its narrow
   `commands.go`/`store.go` crossings forward to WP04.
8. Production fragility: **PASS** — no bare raise/panic/exit path was added;
   errors generally fail loud, subject to issue 4's explicit silent discard.

## Governance note

The generated review prompt still resolved `implementer-ivan` with role
`implementer`. I loaded and applied that specified profile as required while
maintaining independent reviewer/implementer separation. The next cycle should
restore a reviewer profile before review handoff.
