---
work_package_id: WP03
title: Revision Creation and Lineage Inspection
dependencies:
- WP01
- WP02
requirement_refs:
- FR-001
- FR-003
- FR-007
- FR-008
- FR-009
- FR-010
- FR-013
planning_base_branch: chore/spec-kitty-bootstrap
merge_target_branch: chore/spec-kitty-bootstrap
branch_strategy: Planning artifacts were generated on chore/spec-kitty-bootstrap; the execution lane is allocated from lanes.json and completed changes merge back into chore/spec-kitty-bootstrap.
subtasks:
- T010
- T011
- T012
- T013
phase: Phase 3 - Conflict Recovery UX
history:
- at: '2026-08-29T17:18:13Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: ''
authoritative_surface: proposal.go
create_intent: []
execution_mode: code_change
model: ''
owned_files:
- proposal.go
- proposal_test.go
- main.go
role: ''
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP03 – Revision Creation and Lineage Inspection

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `agent_profile` from frontmatter
- **Role**: `role` from frontmatter
- **Agent/tool**: `agent` from frontmatter

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## ⚠️ IMPORTANT: Review Feedback

Check Spec Kitty status and any review feedback before editing. Address all
feedback and append Activity Log entries oldest to newest.

## Objectives & Success Criteria

Ship the public conflict-recovery path:

```text
nh proposal revise PREDECESSOR --base REV --head REV [--body TEXT]
```

Only the predecessor author can create it; local creation refuses a predecessor
already known merged. The operation signs a new event and creates a new
immutable proposal ref without changing predecessor bytes or refs. Proposal
list/show and review accept exact revision candidates and display enough exact
lineage IDs to continue safely, without an implicit winner.

## Context & Constraints

Read charter, spec, plan, data model, contract, research, and quickstart. WP01
defines trusted candidates; WP02 owns all graph queries. Reuse those interfaces
instead of rescanning events in each command.

The proposal-ref transaction boundary already exists: an event may append before
its code ref is created. Preserve the event and report its exact ID if that
happens. Never rewrite the predecessor or retry by forging the same timestamp.

## Branch Strategy

- **Strategy**: Implement in the WP03 execution worktree resolved from
  `lanes.json`, with WP01 and WP02 integrated.
- **Planning base branch**: `chore/spec-kitty-bootstrap`
- **Merge target branch**: `chore/spec-kitty-bootstrap`
- **Implementation command**: `spec-kitty agent action implement WP03 --agent <name>`

## Subtasks & Detailed Guidance

### T010 – Add red CLI tests for author-created immutable revisions

- **Purpose**: Express the P1 recovery story through public command behavior.
- **Steps**:
  1. Extend temporary-repository proposal tests to open a predecessor, create a
     resolved base/head, and invoke the revision command path.
  2. Assert a different identity is refused before event/ref mutation.
  3. Assert same-author creation succeeds for open, rejected, and
     accepted-but-unmerged predecessors.
  4. Assert locally known merged predecessors are refused with an exact ID and
     advice to create an independent proposal.
  5. Snapshot predecessor payload/ref before and after and compare exactly.
  6. Add a second successor naming the same predecessor and assert both remain.
- **Files**: `proposal_test.go`.
- **Validation**: Tests are red only because `revise` and revision-aware output
  are not implemented.

### T011 – Implement `nh proposal revise`

- **Purpose**: Publish a fresh signed recovery candidate through existing Git primitives.
- **Steps**:
  1. Add command routing and usage text in `cmdProposal` and `main.go`.
  2. Parse exact predecessor before flags; require `--base` and `--head`; accept
     optional body but no new title.
  3. Resolve commits through `resolveCommit` and reject equal base/head.
  4. Collect/resolve verified events, build lineage, and require loaded identity
     to equal predecessor actor.
  5. Refuse local creation if any valid merge exists on the predecessor.
  6. Append `proposal.revise` with `subject`, base, head, and body, then call the
     existing immutable `createProposalRef` with the new event ID.
  7. On ref failure, return an error naming the created revision ID and leave all
     existing facts intact.
- **Files**: `proposal.go`, `main.go`.
- **Validation**: T010 happy path, authorization, merge guard, and immutability pass.

### T012 – Generalize proposal and review boundaries

- **Purpose**: Make exact revision IDs usable anywhere this file expects a proposal.
- **Steps**:
  1. Make `proposalEvents` return original and revision candidates.
  2. Generalize show/review kind guards via WP01's helper.
  3. Keep review code-ref availability and signed-head match checks exact.
  4. Confirm `currentReviews` stays keyed only by selected proposal ID—do not
     traverse predecessor edges or copy evidence.
  5. Preserve pre-mission output and behavior for standalone opens as closely as
     possible.
- **Files**: `proposal.go`, `proposal_test.go`.
- **Validation**: A review of a revision succeeds only with its matching code ref;
  predecessor reviews do not appear under it.

### T013 – Add lineage-aware list/show output

- **Purpose**: Make distributed ambiguity visible and actionable.
- **Steps**:
  1. Build one lineage index per command invocation and reuse it for every row.
  2. Show a revision's direct predecessor and inherited display title.
  3. Show exact sorted successor and sibling IDs, superseded state, merged
     lineage members, and merge-conflict state when applicable.
  4. Keep body/actor/title output sanitized with existing helpers.
  5. Never label a candidate latest/current or hide a sibling.
  6. Add exact-output assertions for original, linear, sibling, and conflicting
     lineages, including actionable full/short IDs consistent with current UX.
- **Files**: `proposal.go`, `proposal_test.go`.
- **Validation**: SC-001 and the inspection portion of SC-006 are executable.

## Test Strategy

Use public command functions inside real temporary repositories. Inspect Git
refs and stored payload bytes directly only to prove immutability and failure
semantics. Avoid exact timestamps in output assertions.

```sh
gofmt -w proposal.go proposal_test.go main.go
go test -race ./...
go vet ./...
go build ./...
git diff --check
```

## Risks & Mitigations

- **Wrong author comparison**: compare full actor IDs, never display names.
- **Partial publication**: preserve the appended event and make recovery output
  explicit; never move an existing proposal ref.
- **N+1 graph work**: construct lineage once for list output.
- **Evidence leakage**: keep review selection exact by subject.
- **Ambiguous CLI input**: use existing prefix resolution and its ambiguity error.

## Review Guidance

Verify old and new proposal refs are immutable, non-authors cannot supersede,
known merged predecessors are refused locally, siblings remain visible, and no
wording implies a latest winner. Require public CLI tests and clean failure-path
assertions.

## Activity Log

- 2026-08-29T17:18:13Z – system – Prompt created.

Append later entries chronologically and use Spec Kitty event-sourced status
commands rather than editing task state manually.
