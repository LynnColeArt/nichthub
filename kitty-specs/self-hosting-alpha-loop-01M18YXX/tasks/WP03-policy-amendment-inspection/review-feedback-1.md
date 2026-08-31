# WP03 Review Feedback — Cycle 1

## Blocking issues

**Issue 1 — The policy command and proposal diagnostic have no live production callers, and the required integration handoff was not recorded.**

`policy_commands.go:34` defines `cmdPolicy`, but `main.go` does not route `nh policy`; running `go run . policy show HEAD` exits with `unknown command "policy"`. That routing is deliberately deferred to WP06, which is acceptable only as an explicit downstream handoff. More importantly, `policy_commands.go:138` defines `policyAmendmentDiagnostic`, but repository-wide call-site search finds only its definition and tests. The contract requires `nh proposal open` to print both full policy digests and the base-governs statement when policy bytes differ (`contracts/policy-amendment-cli-v0.md:58-63`). WP03's prompt explicitly requires recording any necessary integration request instead of editing `proposal.go` (`WP03-policy-amendment-inspection.md:151-156`), but no activity-log entry, task note, commit note, or downstream task assigns that call. `proposal.go` is owned by WP05, while WP06 owns only `main.go` and the operational test, so the diagnostic currently risks remaining permanently unreachable.

Remediation: record a review-visible ownership handoff that assigns (a) `cmdPolicy` routing/usage to WP06 and (b) the `policyAmendmentDiagnostic` call from proposal opening to an authorized owner of `proposal.go` (WP05 or an explicitly coordinated shared edit). Ensure the downstream implementation invokes the real helper and preserves the full-digest/base-governs contract. Do not broaden WP03's owned-file edits merely to bypass lane ownership.

**Issue 2 — T015 and FR-008 are not covered by an actual identity-continuity relation.**

`TestPolicyEvaluationUsesExactBaseAcrossAmendment` proves that newly listed raw actor IDs cannot authorize the amendment under the old base policy and can qualify for a later candidate. It never creates, verifies, or projects a device/successor continuity relation. Repository search finds no identity-continuity production call in `policy_commands_test.go`, yet T015 explicitly requires confirming that a valid continuity relation for an actor omitted from policy does not change validation or qualification (`WP03-policy-amendment-inspection.md:168-169`). This leaves the WP's FR-008 claim untested.

Remediation: either make this acceptance check executable after WP02 and add a test using the real verified continuity path, or explicitly move this criterion and its FR-008 coverage to a dependent WP that runs after WP02 (for example the WP06 T029/T030 operational scenario). Update WP03's completion contract so it does not claim the check before its dependency exists.

## Verification evidence

- `go test ./... -run 'Test.*Policy'` — PASS
- `go test ./...` — PASS
- `go test -race ./...` — PASS
- `go vet ./...` — PASS
- `git diff --check 20058dc^..20058dc` — PASS
- WP03 has no declared dependencies; WP04 and WP06 depend on it and are currently planned/unclaimed.
- The implementation commit changes only the declared files: `policy.go`, `policy_commands.go`, and `policy_commands_test.go`.
- No WP03 deliverable code was written into the primary checkout.

## Anti-pattern checklist

1. **Dead code — FAIL.** `cmdPolicy` and `policyAmendmentDiagnostic` have no production call sites. The router deferral is documented, but the proposal diagnostic handoff required by the WP is absent.
2. **Synthetic-fixture test — PASS.** Tests invoke production command/evaluation seams using real temporary Git repositories, signed events, refs, and policy bytes.
3. **Silent empty return — PASS.** The only new empty success value is the documented unchanged-policy result from `policyAmendmentDiagnostic`; failures remain explicit.
4. **FR coverage — FAIL.** T015/FR-008 has no verified continuity-relation assertion. FR-016's public-loop completion is correctly deferred to later WPs, but this WP only proves its base-policy component.
5. **Frozen surface — PASS.** `main.go` and `proposal.go` were not modified.
6. **Locked decision — PASS.** No policy event, self-authorization, continuity-derived role grant, or policy rewrite was introduced.
7. **Shared-file ownership — PASS.** All implementation changes are confined to lane-c's declared files.
8. **Production fragility — PASS.** No panic/bare-raise equivalent or silent transient-race path was added.

## Dependent-lane warning

WP04 and WP06 depend on WP03. They are currently planned, so no active agent needs interruption. If either lane begins before remediation is merged to the mission branch, rebase it afterward with:

```bash
git rebase kitty/mission-self-hosting-alpha-loop-01M18YXX
```
