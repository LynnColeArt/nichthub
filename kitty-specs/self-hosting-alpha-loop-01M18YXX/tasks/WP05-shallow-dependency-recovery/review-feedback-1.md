# WP05 Review — REJECTED

The implementation compiles and its current tests pass, but it does not yet
meet the fail-closed recovery contract. The review used real `file://`
depth-one repositories and found observable state corruption plus several
untested or unimplemented boundaries.

## Blocking issues

1. **A failed recovery mutates the accepted projection before ref promotion.**
   `runReplicationTransaction` calls `releaseRecoveredShallowBoundaries` before
   `replicationBeforePromoteHook` and `promoteReplicationRefs`. A review-only
   depth-one probe injected the before-promotion interruption: the accepted
   actor ref remained unchanged, but the previously reported predecessor gap
   disappeared because the old ref's shallow marker had already been removed.
   Thus a failed transaction changes trusted reachability despite reporting
   failure. The existing interruption test checks refs only and misses the
   marker/projection mutation. Make shallow-boundary changes transactional with
   accepted-ref promotion (including rollback/failure behavior), and assert the
   shallow file, projection, refs, selection, and HEAD are identical after each
   injected failure.

2. **The classifier confuses present-invalid data and unrelated Git failures
   with missing shallow history.** `classifyShallowDependency` turns any failed
   exact read into a gap whenever the repository is shallow. A review probe
   stored a real blob and required it as a commit; the present wrong-type object
   was reported as a recoverable shallow gap. `shallowAcceptedFacts` similarly
   converts every `loadStoredEvent` error (including malformed/signature-invalid
   present events) into an unavailable predecessor, while `guardBasePolicy`,
   `guardPipelineDefinition`, and `guardMergeAncestry` classify all command
   failures without proving omission at a shallow boundary. Preserve malformed,
   wrong-type, unauthorized, ref-mismatch, and genuinely unrelated-history
   errors as invalid/ordinary failures; emit `ShallowDependencyGap` only after
   proving that the exact required object or fact is absent because traversal
   stops at a shallow boundary.

3. **Trust-sensitive and policy-inspection command coverage is incomplete and
   some recovery guidance names the wrong supplier.** `nh policy show/check`
   still use `resolveCommit`/`loadPolicyRevision` directly. A real depth-one
   probe of a missing exact policy commit returned only `Git revision ... is not
   a commit`, not the required gap and recovery action. `nh review`, `nh run
   request/execute`, `nh decide`, and `nh merge` also resolve a missing full
   proposal/request query through the ordinary resolver, so they do not report
   candidate/request gaps. Even `guardProposalQuery` labels a missing signed
   candidate event as supplied by the proposal code ref, although signed events
   come from actor history; fetching that proposal ref cannot recover the event.
   An accepted remote-tracking ref is also treated as currently selected merely
   because its remote name is non-empty, so the diagnostic can recommend
   `nh sync ... --recover-shallow` when no saved selection exists. Inventory and
   guard every specified boundary with accurate, executable supplier guidance.

4. **`--recover-shallow` is not the narrow missing-dependency retry required by
   T026.** `recoverSelectedShallow` accepts only a remote, passes the entire
   saved selection to `runReplicationTransaction`, and therefore has neither
   the missing full IDs nor the supplying-ref subset. With multiple selected
   actors/candidates it re-fetches all selected refs instead of only the refs
   supplying the gap. After promotion it runs only global event-closure checking;
   it cannot discard partial domain results and rerun the originally blocked
   policy/proposal/CI/decision/merge operation from fresh accepted state. Carry
   exact gap context into recovery, restrict the transaction to already-selected
   suppliers, preserve the saved selection and budgets, and perform the
   contracted fresh retry (or provide the explicit command-level orchestration
   that does so).

5. **T023/T027 evidence is materially incomplete and currently masks the above
   defects.** The only real command-path gap fixture is an actor predecessor.
   Candidate event/code ref, base, policy, pipeline, run request, decision, and
   merge-ancestry “coverage” merely calls `classifyShallowDependency` with a
   constructed enum value. Every guarded command in the table short-circuits on
   the same actor-predecessor gap, so it cannot prove its own dependency guard.
   Recovery tests exercise only the event-count budget, the unselected guidance
   test calls the classifier directly, and interruption tests compare refs only.
   Add independent real `file://` depth-one fixtures for every required gap,
   complete-vs-shallow-vs-malformed behavior, all applicable recovery budgets,
   selected and unselected suppliers, proposal-ref shallow markers, merge
   ancestry, narrow multi-selection fetches, fresh operation retry, and full
   state preservation across all interruption points. Each test must fail when
   its production guard/transaction fix is deleted.

## Trust-boundary failure matrix

| Entry point | Current result for the named exact shallow gap | Required result before any trust-sensitive advance |
|---|---|---|
| `nh policy show FULL_COMMIT` | Generic `Git revision ... is not a commit` | `base commit`/`policy blob` gap with the full commit or blob identity and an accurate selected supplier/recovery action |
| `nh policy check --base FULL_BASE --head FULL_HEAD` | Both sides use the ordinary revision/policy loader | Side-specific base/head/policy gap; a present malformed policy must remain an invalid-policy error |
| `nh proposal show FULL_CANDIDATE` | Ordinary event-resolution failure after only global closure checking | `candidate event` gap for a full missing event ID, with actor-history supplier guidance when known |
| `nh proposal status FULL_CANDIDATE` | Emits a candidate gap but falsely maps the signed event to `refs/nh/proposals/*` | Candidate-event gap naming the actor-history supplier, or honest exact selection guidance when the actor is not derivable |
| `nh review FULL_CANDIDATE ...` | Ordinary event-resolution failure; when present, guards only base/head/code ref | Candidate event plus the full review-evidence dependency set must be resolved before `review.submit` is appended |
| `nh run request FULL_CANDIDATE PIPELINE` | Ordinary candidate-resolution failure; when present, checks only the requested pipeline path | Candidate/base/head/policy and exact pipeline-definition gaps must stop before `run.request` is appended; present mismatch/malformed definition remains invalid |
| `nh run execute FULL_REQUEST ...` | Ordinary request-resolution failure | `run request` gap with full request ID and supplying actor ref/action before code execution or `run.result` append |
| `nh run show FULL_REQUEST` | Ordinary request-resolution failure | Exact request-event gap and recovery guidance for verifiable CI inspection |
| `nh run logs FULL_RESULT` | Ordinary result-resolution failure | Exact result/event gap and supplier guidance; present malformed attachment/result remains invalid |
| `nh decide FULL_CANDIDATE ...` | Ordinary candidate-resolution failure | Candidate/base/head/base-policy/pipeline/request/result/review gaps resolved before `proposal.decision` append |
| `nh merge FULL_CANDIDATE` | Ordinary candidate-resolution failure | Candidate/evidence/decision/code/ancestry gaps resolved before Git merge, HEAD movement, or merge-event append |
| `nh proposal open/revise` | Missing full base/head objects are classified, but revision omits base-policy guarding and neither path proves the complete evaluation dependency set before append | Exact base/head/base-policy and required definition gaps before proposal/revision event creation; no fallback to `HEAD`, working tree, names, prefixes, or successors |
| Merge ancestry with present disjoint histories | Any `merge-base` failure in a shallow repo becomes a recoverable gap | Only a genuinely omitted ancestor is a shallow gap; a proven unrelated/mismatched ancestry remains the ordinary fail-closed conflict/error |

`proposal list`, `run list`, and other aggregate read-only views may operate
when every fact they actually consume is present, but their global closure
check must not convert malformed present events into missing shallow history.

## Subtask disposition

- T023: rejected — independent real gap fixtures and before-advance assertions are missing.
- T024: rejected — absence is not distinguished from present-invalid/unrelated failure; supplier guidance can be false.
- T025: rejected — policy and exact query boundaries remain unguarded or misclassified.
- T026: rejected — recovery lacks missing-ID/supplier scope and does not rerun the blocked operation.
- T027: rejected — interruption atomicity is broken and budget/selection/no-expansion evidence is incomplete.

## Anti-pattern checklist

1. Dead code: PASS — new production helpers have live production call chains.
2. Synthetic-fixture test: FAIL — most advertised dependency kinds test only a directly constructed classifier input.
3. Silent empty return: FAIL — `selectedRemoteFor` silently skips selection-load errors, permitting inaccurate recovery guidance.
4. FR coverage: FAIL — FR-013/014/015/017/018/019 are not behaviorally covered at the WP boundaries claimed above.
5. Frozen surface: N/A — no frozen file was identified for this WP.
6. Locked decision: FAIL — failed recovery mutates shallow trust state, and recovery is not limited to the missing supplying refs/fresh original-operation retry.
7. Shared-file ownership: PASS — the WP03 `proposal.go` handoff is recorded, and the WP04 integration exceptions are narrowly described in the commit/activity rationale.
8. Production fragility: N/A — no new panic/bare-raise equivalent was introduced.

## Verification evidence

- `go test ./... -run 'Test.*(Shallow|DependencyGap|Recover)'`: pass
- `go test ./...`: pass
- `go test -race ./...`: pass
- `go vet ./...`: pass
- `go build ./...`: pass
- `gofmt -l` on files changed by `def8fa8`: clean
- `git diff --check def8fa8^..def8fa8`: pass
- Mission-wide `git diff --check` reports pre-existing trailing whitespace in
  planning artifacts; the WP05 feature commit itself is clean.

Governance note: the WP entered review with `agent_profile: implementer-ivan`
rather than a reviewer profile. The required resolved implementer profile was
loaded, but the independent review was conducted under reviewer/implementer
separation.
