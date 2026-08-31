# WP03 Review Feedback — Cycle 2

## Blocking issues

**Issue 1 — The continuity-shaped fixture is synthetic and does not prove FR-008/T015.**

`makeVerifiedContinuityFixture` creates test-only structs, signs their JSON bytes, and manually rechecks those same structs. The resulting value is only inspected for field equality at `policy_commands_test.go:300-302`; it is never passed to `appendEvent`, `verifyEvent`, `collectEvents`, a continuity projector, or `evaluateProposal`. Deleting the fixture construction and assertion would leave every policy qualification assertion unchanged. Current production makes the gap explicit: `validateEventContent` rejects both `identity.authorize` and `identity.accept` as unsupported event kinds, so these bytes are not yet valid Nichthub facts in this lane.

This is exactly the synthetic-fixture anti-pattern: the test recreates a future shape instead of invoking the production path that could grant or withhold authority. Cryptographic self-consistency does not establish that WP02's eventual verified projection cannot affect governance.

Remediation: use one of the two valid paths from cycle 1. Either make WP03 depend on landed WP02 and test real signed, verified, projected continuity facts through production code, or formally reassign the continuity-specific T015/FR-008 assertion to a dependent WP after WP02 and remove it from WP03's completion claim. Do not substitute another test-local parser/projector or hand-built authority model.

**Issue 2 — The correction self-assigned `main.go` and `proposal.go` despite explicit ownership boundaries.**

The live calls themselves behave correctly: `main.go` routes `nh policy`, and `cmdProposalOpen` runs `policyAmendmentDiagnostic` before identity loading, event append, or proposal-ref creation. However, WP03's prompt says top-level routing is intentionally deferred to WP06 and `proposal.go` is outside WP03's owned surface. `lanes.json` assigns `main.go` to WP06 and `proposal.go` to WP05. Cycle-1 feedback explicitly requested a downstream ownership handoff and said not to broaden WP03's owned-file edits. A commit-message “ownership exception” is not an approved ownership transfer or coordination agreement.

Remediation: resolve the planning/ownership conflict explicitly before retaining these edits. Either update the governed lane/task ownership and dependent handoffs with approval, or keep WP03 within its declared files and assign the two minimal call sites to the owning WPs. Preserve the already-correct mutation ordering and exact diagnostic behavior whichever owner lands them.

## Verified behavior

- Live caller search finds production calls from `main.go` to `cmdPolicy` and from `proposal.go` to `policyAmendmentDiagnostic`.
- `go run . policy show HEAD` succeeds and prints the full commit, exact-byte digest, full actor IDs, thresholds, and pipeline information.
- The proposal diagnostic resolves immutable base/head commits, detects an endpoint policy-path change, loads both policies through the canonical parser, and reports both exact digests.
- Invalid changed head policy is rejected before `loadIdentity`, `nextEvent`, `appendEvent`, or `createProposalRef`.
- `go test ./... -run 'Test.*Policy|TestProposalOpenUsesPolicyAmendmentDiagnostic|TestProposalOpenRejectsInvalidPolicyBeforeEventsOrRefs'` — PASS
- `go test ./...` — PASS
- `go test -race ./...` — PASS
- `go vet ./...` — PASS
- `git diff --check` — PASS

## Anti-pattern checklist

1. **Dead code — PASS.** Both new command/diagnostic entry points now have live production callers.
2. **Synthetic-fixture test — FAIL.** The continuity fixture never enters any production verification, storage, projection, or policy-evaluation path.
3. **Silent empty return — PASS.** The unchanged policy-path/byte cases are documented, intentional no-diagnostic results.
4. **FR coverage — FAIL.** Exact base-policy behavior is covered, but the continuity-specific FR-008/T015 assertion remains untested through production behavior.
5. **Frozen surface — FAIL.** The correction modifies `main.go` and `proposal.go` despite the WP prompt explicitly deferring/excluding those files.
6. **Locked decision — PASS.** The implementation introduces no new policy event, policy rewrite, or continuity-derived role grant.
7. **Shared-file ownership — FAIL.** `main.go` belongs to WP06 and `proposal.go` belongs to WP05; no governed ownership amendment or owner handoff authorizes WP03's edits.
8. **Production fragility — PASS.** The new read-only diagnostic errors fail loudly before mutation and add no panic/bare-raise equivalent.

## Dependent-lane warning

WP04 and WP06 depend on WP03 and remain planned/unclaimed. If either starts before remediation is merged, rebase it afterward with:

```bash
git rebase kitty/mission-self-hosting-alpha-loop-01M18YXX
```
