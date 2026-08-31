---
affected_files: []
cycle_number: 3
mission_slug: self-hosting-alpha-loop-01M18YXX
reproduction_command:
reviewed_at: '2026-08-30T21:52:20Z'
reviewer_agent: user
wp_id: WP05
---

# WP05 Review Cycle 3 — REJECTED

Cycle 3 fixes the prior marker-independent ancestry, production replay,
same-head/advanced diagnostics, and pre-promotion object-residue issues. The
required gates pass. One crash-boundary blocker remains: the new durable
unaccepted-object record is consulted by exact object probes, but it is not an
authoritative boundary for accepted event loading after ref promotion and
shallow-marker release.

## Blocking issue

**An acceptance-record failure or crash after ref promotion and marker release
exposes objects that the transaction still records as unaccepted.**

The transaction correctly records absent required objects before copy
(`quarantine.go:391`), copies them, promotes refs (`quarantine.go:416`), and
releases selected shallow boundaries (`quarantine.go:425`). It clears the deny
record only afterward (`quarantine.go:428`). If that final clear fails, the
returned error says that trust operations remain fail-closed. They do not.

A review-only real `file://` depth-one probe used the existing selected actor
recovery fixture and wrapped `replicationReleaseShallow` to:

1. call the real `releaseRecoveredShallowBoundaries` successfully;
2. make `unaccepted-objects.json` unreadable, simulating failure/crash at the
   immediately following acceptance-record step;
3. run production `recoverSelectedShallow("origin")`; and
4. call the production accepted projection loader, `collectEvents()`.

Recovery returned the expected non-zero error:

```text
replication refs and shallow boundaries committed for remote origin, but
imported-object acceptance recording failed; trust operations remain
fail-closed; retry nh sync origin --recover-shallow
```

Nevertheless, `collectEvents()` returned success and included the recovered
predecessor event whose commit remained in the unaccepted set. The failing
assertion was:

```text
unaccepted recovered event sha256:... entered collectEvents after
acceptance-record failure
```

The cause is direct and production-reachable:

- `probeExactGitObject` consults the deny record (`shallow.go:91-104`), but
  `collectEvents` traverses accepted refs and calls `loadStoredEvent` without
  any deny check (`store.go:230-267`).
- `shallowAcceptedFacts` likewise runs `rev-list` and loads each event commit
  directly (`shallow.go:521-573`). In the reproduced repository this admitted
  the event even though another branch shallow marker kept the repository
  globally shallow.
- If release removes the repository's final marker, both
  `prepareShallowVerification` and `guardShallowEventClosure` skip their guards
  solely because Git reports a complete repository (`shallow.go:454-458` and
  `shallow.go:1092-1098`). Pending denied objects therefore become even less
  visible to the guard layer.
- `loadReplicationUnacceptedObjects` treats an absent record as an empty set
  (`shallow.go:134-140`). Deleting the record (or losing it independently of
  the already-written `validated` transaction receipt) silently converts all
  copied residue to trusted objects, so the persisted boundary is not
  tamper-safe.

This violates the selected-replication contract's isolation rule that objects
outside successful acceptance are not accepted projection roots and its rule
that `collectEvents` is the accepted projection loader. It also violates the
WP05 objective and the plan's Shallow Recovery transaction boundary: no
trust-sensitive command may advance until the complete dependency set is
accepted and freshly verified. A process may crash between any two of these
ordinary filesystem operations; a test-only hook is not required for the
window to exist.

The current marker-failure tests do not cover this window. They inject an error
*instead of* releasing the marker, so the old shallow boundary happens to keep
the projection closed. The interruption matrix stops before ref promotion.
There is no test for successful ref promotion + successful marker release +
failed acceptance-record clear, nor a restart test proving the deny record is
honored by every accepted fact/object loader.

## Required success semantics

The minimal compliant repair is a central accepted-event/object denial check,
not another command-specific guard:

1. Every main-repository trust loader must reject a reachable commit/object in
   the pending-unaccepted set. At minimum this includes the accepted event
   projection path (`collectEvents`/main-repository `loadStoredEvent`) as well as
   exact commit, policy, pipeline, and ancestry probes.
2. Pending denial must be enforced independently of `.git/shallow`; a completed
   marker release must not disable it. The original operation must remain
   blocked until ref promotion, marker release, and acceptance-state commit all
   complete, then rerun fresh.
3. Persisted denial state and the validated transaction receipt must be
   versioned and reconciled so missing, malformed, unsafe, or incomplete state
   fails closed on restart. Do not interpret disappearance of a pending record
   as acceptance.
4. Add a real crash/retry fixture for the exact window above. Assert the
   promoted-ref/marker partial success truthfully, verify every affected
   proposal/CI/decision/merge trust probe remains blocked, then prove a retry
   clears denial only after complete verification.

Concurrency matters if the implementation remains a global ID set. Two syncs
can currently perform unsynchronized load/modify/atomic-replace operations;
one can lose another transaction's denied IDs, and one successful transaction
can clear an ID still pending in another transaction. Serialize replication or
use per-transaction/ref-counted pending records whose trusted view is their
union. Continue marking only objects absent before copy so a pre-existing
accepted/shared object is not falsely denied. When the same immutable object is
shared, clear it only when some successfully committed accepted root makes it
trusted or no incomplete transaction still requires denial.

The review-only probe was removed. The lane has no tracked review edits.

## Prior-defect disposition

- Pre-promotion copy residue: corrected to the WP04 contract. Refs, HEAD, raw
  shallow bytes, saved selection, and accepted projection stay unchanged;
  unavoidable unreachable ODB residue is reported honestly and denied by exact
  probes.
- Wrong-type, present-malformed, and unrelated probe errors: fixed.
- Policy/proposal/request/result/decision/merge entry-point gap classification:
  fixed for the exercised unresolved-shallow paths.
- Narrow selected-supplier recovery and unchanged budgets/selection: fixed.
- Full production operation replay and repeated next-gap surfacing: fixed.
- Marker-independent exact ancestry, including hidden, wrong-type, malformed,
  and disjoint cases: fixed.
- Same-head versus advanced marker-release reporting: fixed.
- Post-release acceptance failure/crash denial: **not fixed**.

## Subtask disposition

- T023: rejected — the post-release acceptance failure/restart case is absent.
- T024: partially accepted — exact gap classification is sound, but pending
  unaccepted IDs are not a global integrity boundary.
- T025: rejected — accepted event loading can admit a denied signed fact after
  the partial-success window, bypassing every higher command guard that relies
  on that projection.
- T026: rejected — recovery can report fail-closed while its promoted
  projection is readable before acceptance state commits.
- T027: rejected — crash/retry denial metadata and concurrent pending-set
  safety are not proved.

## Anti-pattern checklist

1. Dead code: PASS — the correction helpers have production callers.
2. Synthetic-fixture test: FAIL — real depth-one coverage is strong, but the
   critical acceptance-record/crash interval is represented by no fixture.
3. Silent empty return: FAIL — an absent unaccepted-object record is silently
   treated as an empty set even when validated/incomplete transaction state may
   remain.
4. FR coverage: FAIL — FR-014/015/017/018 remain unproved across the final
   acceptance-state crash boundary.
5. Frozen surface: N/A — no frozen file was identified for WP05.
6. Locked decision: FAIL — denied residue becomes readable from a promoted
   accepted ref before the transaction's own acceptance record succeeds.
7. Shared-file ownership: PASS — integration changes remain narrow and the
   prior WP03 behavior is preserved.
8. Production fragility: FAIL — correctness depends on no failure occurring
   between marker release and acceptance-record clearing.

## Verification evidence

- `go test ./... -run 'Test.*(Shallow|DependencyGap|Recover)' -count=1`: pass
- `go test ./... -count=1`: pass
- `go test -race ./... -count=1`: pass
- `go vet ./...`: pass
- `go build ./...`: pass
- `gofmt -l` on all files changed by `0bc996b`: clean
- `git diff --check 0bc996b^..0bc996b`: pass
- Mission-wide `git diff --check` still reports only pre-existing trailing
  whitespace in planning artifacts; the WP05 correction commit is clean.
- One review-only real depth-one probe reproduced denied-event visibility after
  acceptance-record failure. It was removed before verdict capture.
