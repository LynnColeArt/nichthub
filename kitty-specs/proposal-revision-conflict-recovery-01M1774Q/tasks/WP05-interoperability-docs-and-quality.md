---
work_package_id: WP05
title: Git Interoperability, Documentation, and Acceptance Proof
dependencies:
- WP03
- WP04
requirement_refs:
- FR-003
- FR-007
- FR-008
- FR-015
- FR-016
planning_base_branch: chore/spec-kitty-bootstrap
merge_target_branch: chore/spec-kitty-bootstrap
branch_strategy: Planning artifacts for this mission were generated on chore/spec-kitty-bootstrap. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into chore/spec-kitty-bootstrap unless the human explicitly redirects the landing branch.
subtasks:
- T020
- T021
- T022
- T023
- T024
phase: Phase 5 - Integration and Documentation
history:
- at: '2026-08-29T17:18:13Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: implementer-ivan
agent: codex
authoritative_surface: revision_sync_test.go
create_intent:
- revision_sync_test.go
execution_mode: code_change
model: ''
owned_files:
- revision_sync_test.go
- commands.go
- README.md
- docs/protocol-v0.md
- docs/governance-v0.md
role: implementer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP05 – Git Interoperability, Documentation, and Acceptance Proof

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `agent_profile` from frontmatter
- **Role**: `role` from frontmatter
- **Agent/tool**: `agent` from frontmatter

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## ⚠️ IMPORTANT: Review Feedback

Check Spec Kitty status and any reviewer feedback before beginning. Resolve all
findings and append chronological Activity Log entries.

## Objectives & Success Criteria

Prove the feature works through Nichthub's actual boundary: ordinary Git refs,
objects, and a bare remote with no hosted process. Revision events and exact code
must synchronize through unchanged wildcard refspecs; peers with the same valid
facts derive identical lineage regardless of observation order. Histories with
no revisions retain their observable behavior.

Update user, wire-protocol, and governance documentation to match shipped code,
then pass the complete charter quality gate. The final result must require no
Docker daemon or Nichthub server.

## Context & Constraints

Read charter and all mission artifacts, especially the contract and quickstart.
WP03 and WP04 must be integrated before this lane starts. The expected design is
that `commands.go` needs no refspec change because revision code already fits
`refs/nh/proposals/*`; edit it only if a red real-remote test proves a localized
change is necessary.

Do not simulate unsupported same-private-key concurrent publication. Sibling
facts can be serialized on the author chain, then delivered/observed in distinct
orders for projection tests.

## Branch Strategy

- **Strategy**: Work in the WP05 execution lane from `lanes.json` after WP03 and
  WP04 are integrated.
- **Planning base branch**: `chore/spec-kitty-bootstrap`
- **Merge target branch**: `chore/spec-kitty-bootstrap`
- **Implementation command**: `spec-kitty agent action implement WP05 --agent <name>`

## Subtasks & Detailed Guidance

### T020 – Prove revision sync through a bare Git remote

- **Purpose**: Verify repository-native transport rather than assuming wildcard refs work.
- **Steps**:
  1. Create a temporary origin repository and at least two working clones using
     the existing integration-test helpers.
  2. Open a proposal, publish/sync it, create a resolved revision as the author,
     and sync again.
  3. In the peer, verify the revision event signature/relationship, predecessor
     visibility, and exact proposal code ref/head.
  4. Review or request a run against the fetched exact revision to prove it is
     usable, not merely listed.
  5. Assert existing actor and proposal remote-tracking ref namespaces; no new
     server behavior or ref namespace may appear.
- **Files**: `revision_sync_test.go`; `commands.go` only if the red test requires it.
- **Validation**: A real Git fetch/push transports both signed event and code.

### T021 – Prove delivery-order convergence and legacy synchronization

- **Purpose**: Cover distributed determinism and compatibility at the public boundary.
- **Steps**:
  1. Publish two same-author siblings serially, both naming the same predecessor.
  2. Construct peer views that observe/present the same verified facts in
     different orders without forking the author chain.
  3. Compare exact sibling, superseded, and merged-lineage outputs.
  4. Repeat the existing proposal/review sync scenario with no revisions and
     assert unchanged list/show/review behavior.
  5. Include remote ref conflict/mismatch checks already enforced by
     `proposalHead`; never trust a fetched ref merely because its name matches.
- **Files**: `revision_sync_test.go`.
- **Validation**: SC-002, SC-005, FR-015, and FR-016 have integration evidence.

### T022 – Document the wire and synchronization contract

- **Purpose**: Keep protocol documentation synchronized with the new signed semantics.
- **Steps**:
  1. Add `proposal.revise` to the implemented event-kind list and document its
     exact subject/base/head/body fields.
  2. Document author/predecessor/cycle validation, immutable per-revision code
     refs, and exact evidence isolation.
  3. Explain that old clients fail closed on the unknown kind under experimental
     `nh/0` and that histories without revisions remain readable by new clients.
  4. State the single-writer actor boundary and why delivery order/timestamps do
     not prove revision-versus-merge creation order.
  5. Confirm synchronization refspec text is unchanged and revisions use the
     existing proposal wildcard.
- **Files**: `docs/protocol-v0.md`.
- **Parallel?**: Yes, after behavior and tests stabilize.
- **Validation**: Every documented field/rule has an executable test reference.

### T023 – Document recovery and governance behavior

- **Purpose**: Give users an actionable workflow and maintainers a precise safety model.
- **Steps**:
  1. Add `nh proposal revise` syntax to README examples and command inventory.
  2. Explain conflict recovery: Git resolves code, Nichthub publishes a new
     immutable signed candidate, and evidence restarts.
  3. Explain siblings, explicit IDs, superseded/closed/conflict states, and the
     absence of a global latest winner.
  4. Update governance docs with local acceptance/merge gates, fresh base policy
     evaluation, competing merge-fact preservation, and actionable errors.
  5. State plainly that no Docker requirement is introduced and same-identity
     disconnected concurrent publication remains out of scope.
- **Files**: `README.md`, `docs/governance-v0.md`.
- **Parallel?**: Yes, file-disjoint from T022.
- **Validation**: Follow documented commands against the integration fixture.

### T024 – Run and record the charter quality gate

- **Purpose**: Produce acceptance-grade evidence over the integrated mission.
- **Steps**:
  1. Format all touched Go files with `gofmt`.
  2. Run race-enabled tests, vet, build, and whitespace validation exactly as
     chartered.
  3. Run focused revision tests and any lineage benchmark, recording commands
     and outcomes in the Activity Log/handoff.
  4. Inspect `git status` and diff to ensure only owned files changed in this lane.
  5. Confirm no generated binary, temporary repository, identity key, log, or
     Docker artifact is tracked.
- **Files**: No new product file; only owned files may receive required fixes.
- **Validation**: All commands below exit zero.

## Test Strategy

The integration tests must invoke real Git repositories and refs. They may call
Go command functions directly but cannot replace fetch/push with mocks. Keep
fixtures bounded and deterministic so `-race` remains reliable.

```sh
gofmt -w *.go
go test -race ./...
go vet ./...
go build ./...
git diff --check
```

## Risks & Mitigations

- **False transport confidence**: inspect fetched event and proposal refs plus
  signed head equality in the peer.
- **Unsupported concurrency claim**: serialize author events and vary only view
  presentation/observation order.
- **Documentation drift**: cross-check CLI usage, event validation, and status
  strings against tests after code settles.
- **Flaky integration**: use local bare remotes, explicit branches, and no sleeps.
- **Scope creep**: no new backend, container, service, or dependency.

## Review Guidance

Require evidence that the existing wildcard refspec carries revisions and exact
code. Compare documented commands/fields against source. Run every charter gate,
verify legacy sync, and reject any language implying global latest or same-key
multiwriter support.

## Activity Log

- 2026-08-29T17:18:13Z – system – Prompt created.

Append later entries chronologically. Use Spec Kitty status events for WP and
subtask transitions.
