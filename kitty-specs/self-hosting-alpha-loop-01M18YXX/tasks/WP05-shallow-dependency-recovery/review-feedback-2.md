# WP05 Review Cycle 2 — REJECTED

Cycle 2 correctly narrows recovery to one saved supplier, preserves wrong-type
and unrelated probe failures, adds real policy/event gap paths, and defers
shallow-marker release until after ref promotion. Four contract blockers remain.

## Blocking issues

1. **The after-object-copy failure path retains recovered objects in the main
   object database.** A review-only real depth-one probe set
   `replicationAfterCopyHook`, ran `recoverSelectedShallow`, received an error,
   and then successfully resolved the previously missing predecessor commit in
   the main repository. The current interruption test asserts object absence
   only for `before-ref-transaction`; however `replicationBeforePromoteHook`
   was moved to *before* object copy, so that assertion no longer covers the
   actual interval immediately before the ref transaction. Required success
   semantics: every failure before accepted-ref promotion—including after copy
   and object verification—must preserve refs, HEAD, raw shallow bytes, saved
   selection bytes, gap/projection result, and absence of newly recovered
   objects. Test object absence at every pre-promotion interruption and keep the
   hook names/order faithful to the phase they claim to simulate.

2. **Merge ancestry still depends on Git's shallow marker instead of exact
   object availability.** A review probe created a real depth-one two-commit
   clone, copied the exact missing parent commit and tree into the clone, and
   called `guardMergeAncestry(base, head)`. All exact ancestry objects were
   present, but `git merge-base` still returned status 1 because it honors the
   shallow marker; the guard returned an ordinary failure. This violates D-007's
   core rule that a shallow marker is irrelevant when the exact objects suffice.
   Conversely, the new `merge-ancestry` fixture removes the base object itself
   and relabels that base-object gap as `merge ancestor`; it does not exercise a
   present base/current pair with an omitted intermediate ancestor. Required
   semantics: prove ancestry from exact commit-parent objects independently of
   marker truncation; succeed when the complete exact path is already present,
   emit a structured `merge ancestor` gap naming the first exact missing parent
   and selected supplier when traversal hits a real omission, and preserve the
   ordinary mismatch for fully present disjoint histories.

3. **The production CLI still does not rerun the original blocked verification.**
   `recoverSelectedShallow` always calls `recoverSelectedShallowWithRetry` with
   `nil`; that substitutes `verifyRecordedShallowGap`, which checks only the
   primitive recorded object/event. The non-nil original-operation callback is
   reachable only from tests. A review-only integration probe recorded a real
   proposal-status base gap, recovered the selected proposal ref, and observed
   `nh sync --recover-shallow` return success and delete the gap record even
   though a fresh `cmdProposalStatus` immediately failed because the base policy
   required a pipeline definition absent from the candidate head. Required
   semantics: the production path must retain enough typed operation context to
   rerun the same complete policy/proposal/CI/decision/merge dependency and
   validation boundary from fresh accepted state. It must return non-zero and
   preserve retry context if any next dependency or validation still fails.
   A test-injected closure does not establish this behavior. Also check whether
   a recorded dependency is already satisfied before fetching; that case must
   be a no-op followed by the same full fresh verification.

4. **Post-promotion marker-release reporting is self-contradictory and test
   coverage still omits required cases.** The marker-failure test proves its
   recovery is same-head (`beforeRefs == afterRefs`) while requiring the error
   text `accepted refs advanced`. For shallow-history recovery, promotion often
   commits the same ref value while importing deeper objects, so that statement
   is false. Report exactly whether ref values changed, whether the ref
   transaction committed, that objects were imported, and that marker release
   failed/facts remain fail-closed; preserve the sanitized underlying cause and
   retry instructions. Coverage also lacks recovery under
   `MaxAttachmentBytes`, a present malformed/signed-event fixture (only malformed
   policy is tested), the existing after-fetch hook, object-absence checks for
   after-copy failure, a real omitted-intermediate merge-ancestor fixture, and a
   production-reachable full-operation retry. Add deletion-sensitive tests for
   each rather than relying on the injected retry callback.

## Prior-blocker disposition

- Wrong-type object classification: fixed and behaviorally tested.
- Unrelated object-probe errors: fixed and behaviorally tested.
- Present malformed policy: preserved as invalid and tested.
- Policy show/check and missing proposal/request/result entry points: materially fixed with real full-ID gaps.
- Saved-selection mutation and unrelated selected supplier fetch: fixed by an ephemeral subset and byte-for-byte selection check.
- Before-ref marker mutation: fixed by deferring marker release, but after-copy object rollback remains open.
- Full original-operation retry: not fixed in the production path.
- Exact merge ancestry under a shallow marker: not fixed.

## Subtask disposition

- T023: rejected — the merge-ancestry fixture does not model a hidden intermediate ancestor, and required malformed-event paths remain absent.
- T024: partially accepted — exact object type/probe classification is corrected, but merge omission classification remains incomplete.
- T025: partially accepted — entry-point guards are substantially improved, but exact ancestry and post-recovery complete verification remain incorrect.
- T026: rejected — supplier subsetting is correct, but after-copy failure retains objects and the production path does not rerun the blocked operation.
- T027: rejected — interruption/object-retention, attachment-budget, truthful partial-success, and production-retry proofs are incomplete.

## Anti-pattern checklist

1. Dead code: PASS — new production helpers have callers, though the non-nil retry behavior is test-only and does not prove the CLI contract.
2. Synthetic-fixture test: FAIL — the claimed original-operation retry is an injected test callback; production always supplies `nil`.
3. Silent empty return: FAIL — successful recovery ignores `os.Remove` failure for the durable gap record, allowing stale retry state while reporting success.
4. FR coverage: FAIL — FR-014/015/017/018/019 still lack the required exact-ancestry, rollback, and production fresh-retry behavior.
5. Frozen surface: N/A — no frozen file was identified for this WP.
6. Locked decision: FAIL — exact present ancestry remains blocked solely by `.git/shallow`, and failed after-copy recovery retains newly imported objects.
7. Shared-file ownership: PASS — WP03 behavior remains intact and correction-cycle WP04/policy ownership exceptions are documented in the commit message/activity record.
8. Production fragility: N/A — no new panic/bare-raise equivalent was introduced.

## Verification evidence

- `go test ./... -run 'Test.*(Shallow|DependencyGap|Recover)'`: pass
- `go test ./...`: pass
- `go test -race ./...`: pass
- `go vet ./...`: pass
- `go build ./...`: pass
- `gofmt -l` on files changed by `fce0d24`: clean
- `git diff --check fce0d24^..fce0d24`: pass
- Three review-only failing probes reproduced after-copy object retention,
  exact-present ancestry overblocking, and false success without full original
  verification. All probes were removed; the lane has no tracked review edits.
