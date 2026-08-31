---
affected_files: []
cycle_number: 1
mission_slug: agent-memory-protocol-01M19TMH
reproduction_command:
reviewed_at: '2026-08-31T04:37:35Z'
reviewer_agent: user
wp_id: WP03
---

# WP03 Review Feedback — Cycle 1

Verdict: changes requested.

## Issue 1 — Same-author challenges violate the locked cross-author rule

`ProjectMemories` applies its actor mismatch check only to `supersede` and
`retract` (`memory_projection.go:228-230`). A signed challenge authored by the
same actor as its target is therefore accepted into `Relationships` and added
to the target's `Challengers` list. That contradicts the WP03 requirement to
enforce the version-0 cross-author challenge rule, the plan's “cross-author
challenges” decision, and the specification's cross-actor lifecycle entity.

An independent review probe constructed both records through the real WP01
signing helpers and observed the defect:

```text
--- FAIL: TestReviewProbeSameAuthorChallengeRejected
same-author challenge entered the canonical graph:
operation="challenge" targetId="sha256:e0e837cb..."
```

Fix the relationship validator so a challenge whose signed actor equals the
target record's signed actor is rejected with a stable, scoped actor-rule
diagnostic. Preserve the target and all unrelated records. Add focused tests
that prove a same-author challenge is rejected, a cross-author challenge is
accepted, and challenge evidence never changes trust, evidence resolution,
applicability, lifecycle activity, truth, or authorization semantics.

## Issue 2 — T015's required convergence and separation proof is incomplete

The current tests cover a useful core, but they do not implement the acceptance
matrix required by T011–T015 and therefore do not establish several referenced
FRs/NFRs. In particular, add production-path assertions for:

- target-before-edge and edge-before-target sets, cross-stream targets, and
  same-author records in separate streams;
- a linear supersession chain and lifecycle precedence with a missing edge,
  retraction, branching, and challenges all present without erasing edge IDs;
- original, reversed, and deterministic shuffled complete-corpus projections,
  including deliberately misleading timestamps;
- `exact`, `descendants`, and `subject` applicability at the anchor, a child,
  an unrelated commit, an unavailable commit, and a wrong-type object;
- correct blob and explicit-absent anchors plus later change, removal, and
  rename, proving anchor-only validation and no rename inference;
- resolved, missing, and invalid evidence across all three namespaces, with
  full typed sorted dependency details;
- two exact policy commits after moving `HEAD`, asserting both the selected
  policy digest and the resulting actor/kind trust class;
- one-dimension-at-a-time separation of signature, lifecycle, applicability,
  evidence, actor trust, kind trust, and policy presence; and
- complete provenance/full-ID output plus an unchanged collaboration-only
  projection/public-event-ID baseline.

The existing `TestMemoryProjectionCompletePermutation` corpus contains only a
target, one successor, and one missing challenge; it does not satisfy T015's
representative-corpus contract. `TestMemoryPolicyExactCommitLoad` checks that a
policy loads but does not project under both commits or demonstrate independence
from moved `HEAD`.

## Gate evidence

- Focused relationship, lifecycle, applicability/evidence, and policy tests:
  pass.
- `go test ./... -run 'TestMemoryProjection'` three times: pass.
- `go test ./...`: pass.
- `go test -race ./...`: pass.
- `go vet ./...`: pass.
- `go build ./...`: pass.
- `git diff --check`: pass.
- WP03 implementation commit changes only the four owned files and does not
  modify frozen mission contracts or existing collaboration source files.

## WP anti-pattern checklist

1. Dead code: N/A for this intermediate package; its exported projection seam
   is explicitly handed to dependent WP04/WP05/WP06 and is not wired by WP03.
2. Synthetic-fixture tests: PASS for the tests present; fixtures invoke signed
   WP01 envelopes and the production projection path.
3. Silent empty return: PASS; no new silent empty failure branch found.
4. FR coverage: FAIL; Issue 2 lists required behavior without production-path
   assertions.
5. Frozen surface: PASS.
6. Locked decision: FAIL; same-author challenge acceptance contradicts the
   cross-author/cross-actor lifecycle decision.
7. Shared-file ownership: PASS; the implementation commit is confined to the
   declared WP03 ownership map.
8. Production fragility: PASS; no new bare raise/panic path was introduced.
