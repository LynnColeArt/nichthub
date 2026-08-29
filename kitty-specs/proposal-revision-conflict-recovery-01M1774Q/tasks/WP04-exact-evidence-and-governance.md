---
work_package_id: WP04
title: Exact Evidence and Lineage-Safe Governance
dependencies:
- WP01
- WP02
- WP03
requirement_refs:
- FR-004
- FR-005
- FR-006
- FR-009
- FR-011
- FR-012
- FR-014
- FR-015
planning_base_branch: chore/spec-kitty-bootstrap
merge_target_branch: chore/spec-kitty-bootstrap
branch_strategy: Planning artifacts for this mission were generated on chore/spec-kitty-bootstrap. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into chore/spec-kitty-bootstrap unless the human explicitly redirects the landing branch.
subtasks:
- T014
- T015
- T016
- T017
- T018
- T019
phase: Phase 4 - Trust and Governance
history:
- at: '2026-08-29T17:18:13Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: implementer-ivan
agent: codex
authoritative_surface: policy.go
create_intent:
- governance_revision_test.go
execution_mode: code_change
model: ''
owned_files:
- policy.go
- policy_test.go
- ci.go
- ci_test.go
- governance.go
- governance_revision_test.go
role: implementer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP04 – Exact Evidence and Lineage-Safe Governance

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `agent_profile` from frontmatter
- **Role**: `role` from frontmatter
- **Agent/tool**: `agent` from frontmatter

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## ⚠️ IMPORTANT: Review Feedback

Inspect the status event log and any `review_ref` first. Every reviewer finding
is a required implementation item; append resolutions to the Activity Log.

## Objectives & Success Criteria

Complete the trust boundary for revisions. Policy comes from the selected
revision's exact base; required pipeline definitions come from its exact head;
reviews, run requests/results, decisions, and merges count only when bound to
that exact proposal identity and inputs. A revision inherits zero evidence.

Known successors block new acceptance/merge of their predecessor. A known merge
anywhere in a lineage blocks new acceptance/merge of other candidates and names
the blocking IDs. If separate repositories already produced competing merges,
both facts remain visible and status reports a conflict rather than rejecting
history or choosing a winner.

## Context & Constraints

Read charter and all mission artifacts. WP01 generalizes stored relationships;
WP02 supplies lineage state; WP03 supplies exact revision selection and review.
Do not create an evidence inheritance path. Existing exact `Subject`, `Commit`,
`Definition`, `Policy`, `Head`, and `Evidence` checks are the model to preserve.

The specification blocks new acceptance and merge, not historical inspection.
Rejection remains an allowed signed exact-proposal fact. A local view can only
gate on facts it has received.

## Branch Strategy

- **Strategy**: Use the WP04 execution lane after WP01–WP03 integration.
- **Planning base branch**: `chore/spec-kitty-bootstrap`
- **Merge target branch**: `chore/spec-kitty-bootstrap`
- **Implementation command**: `spec-kitty agent action implement WP04 --agent <name>`

## Subtasks & Detailed Guidance

### T014 – Prove predecessor evidence never satisfies a revision

- **Purpose**: Turn the central trust claim into an executable regression barrier.
- **Steps**:
  1. Build a predecessor with qualifying approval, passing CI, acceptance, and
     policy digest under its base.
  2. Add a valid revision with a new ID/base/head and matching code ref.
  3. Evaluate the revision and assert zero approvals, zero passes, zero accept
     decisions, and not ready/accepted.
  4. Mix predecessor evidence IDs into a crafted revision decision/merge and
     assert relationship validation rejects them.
  5. Add fresh exact revision evidence and assert it alone qualifies.
- **Files**: `policy_test.go`, `governance_revision_test.go`.
- **Validation**: SC-003 is proven with positive and hostile cases.

### T015 – Generalize policy evaluation and status

- **Purpose**: Evaluate revisions as full proposal candidates in their own context.
- **Steps**:
  1. Replace the `proposal.open`-only guard in `evaluateProposal` with the shared
     candidate predicate.
  2. Continue loading policy from the selected candidate's exact base and code
     from its exact head/ref.
  3. Add lineage state to `ProposalEvaluation` or a composed status view without
     duplicating WP02 graph logic.
  4. Make `proposalStatus` and `cmdProposalStatus` display superseded,
     lineage-closed, and merge-conflict states plus exact relevant IDs.
  5. Preserve legacy standalone status wording/results where no revision edge exists.
- **Files**: `policy.go`, `policy_test.go`.
- **Validation**: Revision readiness starts fresh and status is actionable.

### T016 – Generalize CI request/result handling to revisions

- **Purpose**: Allow customized pipelines to evaluate the exact revised code.
- **Steps**:
  1. Replace open-only proposal checks in run request and execution validation
     with the candidate predicate.
  2. Preserve exact request subject, candidate head, pipeline definition digest,
     maintainer requester, trusted runner, and attached-log checks.
  3. Add a revision run happy path and negative cases where predecessor request,
     result, commit, or definition is reused.
  4. Confirm run list/show/log behavior needs no lineage guessing and continues
     to use explicit request/result IDs.
- **Files**: `ci.go`, `ci_test.go`.
- **Parallel?**: Yes, within this lane it is file-disjoint from policy and
  governance until the final full-suite integration.
- **Validation**: Only fresh exact revision results count.

### T017 – Block stale acceptance

- **Purpose**: Prevent a locally known superseded or closed candidate from gaining
  a new acceptance.
- **Steps**:
  1. Before `--accept`, query lineage state for the explicit candidate.
  2. Refuse if it has known successors or another lineage member is merged.
  3. Include exact successor/merged IDs and the status command in the error.
  4. Continue to permit a well-formed `--reject` fact unless the existing merged
     guard applies; do not erase earlier acceptance.
  5. Revalidate the guard at event creation time from the current collected set.
- **Files**: `governance.go`, `governance_revision_test.go`.
- **Validation**: FR-004 and FR-011 acceptance gates pass for linear and sibling lineages.

### T018 – Block stale merge and preserve conflict recovery context

- **Purpose**: Make terminal Git mutation safe under the locally known lineage.
- **Steps**:
  1. Before any Git merge, refuse superseded candidates, lineages merged
     elsewhere, and lineages already reporting competing merge facts.
  2. Include every blocking exact proposal ID in stable order.
  3. Keep all existing policy, code-ref, clean-worktree, ancestry, author, and
     acceptance checks.
  4. On Git conflict, retain automatic abort and extend the error with the exact
     attempted proposal ID plus `nh proposal revise` recovery guidance.
  5. Do not emit a merge event on conflict; do not attempt conflict resolution.
  6. Do not reject already signed competing merge events during collection;
     they are immutable facts derived into conflict state by WP02.
- **Files**: `governance.go`, `governance_revision_test.go`.
- **Validation**: Worktree cleanliness is asserted before/after injected conflicts.

### T019 – Add governance compatibility and distributed-conflict regressions

- **Purpose**: Prove new gates do not reinterpret old histories.
- **Steps**:
  1. Run the complete pre-mission policy/merge scenario unchanged.
  2. Add accepted-but-unmerged predecessor → revision behavior: acceptance
     remains displayed but predecessor merge is blocked.
  3. Add two siblings with independently valid merge events in a combined event
     view and assert both validate and conflict status lists both proposal IDs.
  4. Permute fact order and compare evaluation/status results.
  5. Cover missing code refs and mismatched heads for revisions.
- **Files**: `policy_test.go`, `governance_revision_test.go`, `ci_test.go`.
- **Validation**: NFR-005 passes and distributed history remains append-only.

## Test Strategy

Use exact event IDs and real temporary Git worktrees for merge effects. Capture
the target branch OID and `git status --porcelain` around failed merges. Avoid
forging impossible same-identity concurrent actor forks; competing merges may be
signed by independently authorized maintainers after both siblings are known.

```sh
gofmt -w policy.go policy_test.go ci.go ci_test.go governance.go governance_revision_test.go
go test -race ./...
go vet ./...
go build ./...
git diff --check
```

## Risks & Mitigations

- **Evidence leakage**: every selector retains exact subject/head/policy/definition checks.
- **Invalid retrospective ordering**: preserve remote facts and gate only future local actions.
- **Git side effects before guard**: all lineage checks run before invoking Git merge.
- **Status precedence**: merged and merge-conflict states must not be hidden by
  ready/accepted flags; test every combination.
- **File coupling**: integrate CI, policy, and governance only after focused tests pass.

## Review Guidance

Trace one predecessor evidence ID through every evaluator and prove it cannot
qualify the revision. Inspect that all terminal guards execute before mutation,
errors name exact candidates, competing merge events remain valid, and legacy
governance tests are unchanged and green.

## Activity Log

- 2026-08-29T17:18:13Z – system – Prompt created.

Append entries chronologically. Record task status through Spec Kitty rather
than modifying reference rows.
