---
affected_files: []
cycle_number: 2
mission_slug: agent-memory-protocol-01M19TMH
reproduction_command:
reviewed_at: '2026-08-31T05:49:33Z'
reviewer_agent: user
wp_id: WP06
---

# WP06 review feedback — cycle 2

Verdict: changes requested.

Cycle two resolves the synthetic anchor-gap path, the basic structured outcome,
the named interruption/receipt hooks, independent pending anchors, and the
single-clone projection comparison. The remaining findings are limited to
typed supplier recovery and the explicit memory transaction evidence required
by T029.

## 1. Lifecycle and typed-evidence recovery still selects the owning stream, not the supplier

The new production anchor scenario works because a selected proposal supplies
the missing Git commit. The general recovery path does not work for the other
typed dependencies required by T030:

- `memoryReplicationRequiredSelectors` only derives proposal selectors whose
  event head equals a missing Git object. It never derives an actor supplier
  for `evidence-event` or a memory-stream supplier for `lifecycle-target` /
  `evidence-memory`.
- `memoryReplicationRecovery` is passed `missing.Stream`, which is the owning
  memory's stream. Since that stream is already selected, it recommends
  `nh sync <remote> --recover-shallow` even when the missing lifecycle/evidence
  fact lives in another, unselected stream.
- `recoverySelectionSubset` consequently re-fetches only the owning stream and
  the optional proposal selectors. Even if the operator adds the actual actor
  or memory supplier to the saved selection, the recovery subset excludes it
  unless it happened to be recorded in `RequiredSelectors`.

This makes lifecycle-target, event-evidence, and memory-evidence gaps
deterministically retry the same unresolved transaction instead of importing
their exact supplier. It does not meet FR-022 or T030 steps 1–7.

Carry the actual known supplier kind/full ID into `RequiredSelectors` and the
recovery command, and keep an honest non-actionable gap when no exact supplier
can be derived without unauthorized discovery. Add production-path shallow
tests for lifecycle target, `event:` evidence, `memory:` evidence, and `git:`
evidence; prove wrong-type/malformed cases remain ordinary invalid data.

## 2. `RequiredRef` identifies the owner memory ref, not the required supplier ref

Both dependency paths assign `outcome.RequiredRef = outcome.request.SourceRef`,
and `recordReplicationMemoryShallowGap` copies that value. In the new anchor
recovery fixture the known supplier is `refs/nh/proposals/<proposal-id>`, but
the recorded `RequiredRef` is the memory stream ref. The test asserts the
derived proposal selector but omits `RequiredRef`, so this protocol-visible
misbinding passes.

Populate `RequiredRef` from the exact supplying proposal/actor/memory ref when
known and leave it empty when unknown; never label the owning stream ref as the
missing dependency's required ref. Assert it in every typed-gap fixture.

## 3. Memory-specific accepted/shared-object budget semantics and mixed atomicity remain unproven

`TestMemoryReplicationBudgetCountsCommits` measures one complete graph against
an empty accepted repository and then calls `enforceReplicationBudgets`
directly for the four numeric boundaries. It does not test the WP's required
incremental semantics: objects reachable from the same stream's prior accepted
head are discounted, while objects shared with another stream or merely present
unreferenced in the main object database are still charged. No test passes a
real previous accepted root to
`measureQuarantinedSelectionUnderBudgets` for memory.

The crash tests now cover every named boundary, but each uses one memory
promotion. The ref-transaction conflict test therefore cannot prove the
required all-old/all-new atomicity across a mixed actor + proposal + memory
promotion set or that collaboration bytes remain unchanged at each failed
boundary. Generic pre-memory transaction tests do not exercise accepted-memory
ref parsing/receipts inside that mixed atomic set.

Add the same-stream prior-head discount and cross-stream shared/unreferenced
object cases at exact rejecting/admitting limits. Run the crash/ref/receipt
matrix with a mixed promotable actor, proposal, and memory set, snapshotting all
accepted refs and collaboration payload bytes before and after the injected
failure.

## Anti-pattern checklist

1. Dead code: **PASS** — the new shallow dependency mapping and verification
   helpers now have production callers.
2. Synthetic-fixture test: **FAIL** — anchor recovery is now real, but the
   remaining lifecycle/event/memory evidence requirements have no production
   recovery fixture.
3. Silent empty return: **PASS** — no new unexplained silent-empty return found.
4. FR coverage: **FAIL** — FR-022 and the T029/T030 portions of FR-007/FR-023
   remain incomplete for the cases above.
5. Frozen surface: **PASS** — remediation commits `1c56eb8` and `260fc11`
   remain within the five WP06-owned files and do not modify frozen contracts.
6. Locked decision: **PASS** — no service, Docker, provider API, transcript
   capture, trust conflation, or quarantine recall path was introduced.
7. Shared-file ownership: **PASS** — all modified product paths belong to
   lane-f/WP06 in `lanes.json`.
8. Production fragility: **PASS** — new errors remain scoped and fail loud;
   no panic or bare transient rethrow was added.

## Validation evidence

- Focused WP06 and legacy selection suites: pass.
- Focused mixed/budget/transaction/shallow/fresh-clone suite repeated five
  times: pass.
- `git diff --check kitty/mission-agent-memory-protocol-01M19TMH..HEAD`: pass.
- `go test ./... -count=1`: pass (`42.466s`).
- `go test -race ./... -count=1`: pass (`47.215s`).
- `go vet ./...`: pass.
- `go build ./...`: pass.

The green gates confirm the remediation is stable; this rejection is for the
remaining contract gaps, not regression failures.
