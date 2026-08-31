---
affected_files: []
cycle_number: 1
mission_slug: agent-memory-protocol-01M19TMH
reproduction_command:
reviewed_at: '2026-08-31T05:25:05Z'
reviewer_agent: user
wp_id: WP06
---

# WP06 review feedback — cycle 1

Verdict: changes requested.

## 1. Real memory shallow recovery is not connected

`classifyMemoryShallowDependency` and the `shallowMemoryPredecessor` /
`shallowMemoryStream` kinds have no production callers. The only WP06 shallow
test constructs a `ShallowDependencyGap` literal and calls
`recoverySelectionSubset`; it never obtains a gap from an unavailable anchor,
Git/event/memory evidence item, lifecycle target, or predecessor through the
real replication/projection path. Consequently FR-007, FR-022, and T030 are
not met: missing memory dependencies become ordinary replication outcomes,
but no durable typed gap with owning memory ID, stream ID, selected remote,
required ref, and exact recovery action is produced or replayed.

Connect the WP03 dependency classifications to the existing durable shallow
gap seam, only when the repository is actually shallow. Cover each memory
dependency kind through production code, prove malformed/wrong-type data is
not reclassified as shallow, and exercise `nh sync <remote> --recover-shallow`
end to end with unchanged saved selection bytes and full post-recovery
verification.

## 2. Required hostile mixed-transaction and budget boundaries are unproven

`memory_replication_test.go` has no mixed actor + proposal + valid memory +
invalid/over-budget/dependency-missing memory transaction. It never snapshots
collaboration event IDs or encoded payloads, and it tests only `max-events`
above the limit. There are no one-below/exact/one-above cases for event count,
object count, largest object, attachment size, or total bytes; no attempted
memory attachment; no same-stream accepted-object discount versus
shared/unreferenced objects; and no shared-transaction pending-state test.
This leaves FR-023, NFR-006, NFR-010, T028, and T029 without executable
evidence.

Add the real-remote mixed hostile corpus and assert independent outcomes keyed
by kind plus full ID. Prove byte-identical collaboration projection for every
memory failure and cover every configured budget boundary, including zero
valid memory attachments and rejection of an attempted attachment.

## 3. Durable interruption coverage is incomplete

The WP adds only one interruption test, at `replicationBeforePromoteHook`.
There is no WP06 memory coverage for after-fetch, after-measure, after-copy,
pending-anchor/validated receipt boundaries, ref-transaction failure, or
completion-record failure. The required all-old/all-new accepted-ref behavior,
residue denial, receipt reconciliation, idempotent retry, and preservation of
another transaction's pending state are therefore not demonstrated.

Exercise the existing production hooks at every T029 boundary with mixed
actor/proposal/memory promotions. Assert exact old refs before promotion,
truthful partial-success diagnostics after promotion, fail-closed projection,
and deterministic reconciliation/retry.

## 4. Fresh-clone convergence test does not test the required projection

`TestMemoryReplicationFreshCloneConvergesWithoutPrivateState` imports one
record and checks one memory ID plus three absent private paths. It does not
compare publisher and clone anchors, lifecycle edges, evidence status,
applicability, or exact-filtered projection, and it has no initial shallow gap
or explicit recovery. It also does not prove collaboration-only compatibility
alongside the memory corpus. Thus FR-024, NFR-010, NFR-012, and T030 remain
unproven.

Build the specified multi-stream corpus with lifecycle links, anchors, and
typed evidence; synchronize it into a credential-disabled fresh clone; compare
the full deterministic projection and exact filters; then induce and recover
one exact shallow dependency. Include a collaboration-only clone that imports
zero memory refs and reproduces its pre-memory projection.

## 5. Exact recovery guidance is not actionable

`classifyReplicationMemoryDependencies` currently emits prose such as
`select the full memory stream supplying <missing-id>`, and Git dependency
revalidation emits `select the exact collaboration or memory ref supplying
<object>`. These do not identify an exact selector update or command, and the
outcome does not retain the dependency kind as structured recovery state.
That contradicts FR-022 and the memory replication contract's exact recovery
requirements.

Preserve the typed dependency and owner/stream information through the
replication outcome/shallow-gap path. When the supplier is known, emit the
full `--memory <stream-id>` update while preserving existing selectors and
budgets; only offer `--recover-shallow` when the saved exact selection already
authorizes it.

## Anti-pattern checklist

1. Dead code: **FAIL** — the memory shallow classifier and two new dependency
   kinds have no production callers.
2. Synthetic-fixture test: **FAIL** — the only shallow test fabricates the gap
   that production code is required to create.
3. Silent empty return: **PASS** — no new unexplained silent-empty failure path
   found.
4. FR coverage: **FAIL** — FR-007/022/023/024 and NFR-006/010/012 lack the
   required behavior-level assertions described above.
5. Frozen surface: **PASS** — commit `4e9b7a4` changes only the five WP06-owned
   files and no frozen mission contract.
6. Locked decision: **PASS** — no service, Docker, provider API, automatic
   transcript capture, or selection-as-trust path was introduced.
7. Shared-file ownership: **PASS** — the five changed paths are assigned only
   to lane-f/WP06 in `lanes.json`.
8. Production fragility: **PASS** — new fail-loud paths are scoped and return
   diagnostics rather than panicking or silently swallowing failures.

## Validation evidence

- Focused WP06 and legacy selection tests: pass.
- `git diff --check kitty/mission-agent-memory-protocol-01M19TMH..HEAD`: pass.
- `go test ./... -count=1`: pass (`43.381s`).
- `go test -race ./... -count=1`: pass (`46.437s`).
- `go vet ./...`: pass.
- `go build ./...`: pass.

The green gates establish regression health, but they do not compensate for
the missing contract behaviors and synthetic shallow coverage above.
