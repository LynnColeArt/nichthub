---
work_package_id: WP03
title: Policy Amendment Inspection
dependencies: []
requirement_refs:
- FR-001
- FR-002
- FR-003
- FR-004
- FR-008
- FR-016
- FR-018
planning_base_branch: feat/self-hosting-alpha-loop
merge_target_branch: feat/self-hosting-alpha-loop
branch_strategy: Planning artifacts for this mission were generated on feat/self-hosting-alpha-loop. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/self-hosting-alpha-loop unless the human explicitly redirects the landing branch.
subtasks:
- T011
- T012
- T013
- T014
- T015
phase: Phase 1 - Independent Foundations
history:
- at: '2026-08-30T17:26:50Z'
  actor: system
  action: Prompt generated via /spec-kitty.tasks
agent_profile: implementer-ivan
agent: codex
authoritative_surface: policy.go
create_intent:
- policy_commands.go
- policy_commands_test.go
execution_mode: code_change
model: ''
owned_files:
- policy.go
- policy_commands.go
- policy_commands_test.go
role: implementer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP03 – Policy Amendment Inspection

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter (or any user-defined profile), and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objectives & Success Criteria

Give maintainers a deterministic way to inspect the exact policy governing a
commit and to validate/compare a proposed replacement before publishing a
candidate. Policy amendment remains an ordinary Git change and ordinary
proposal; the proposed head policy must never authorize its own evidence.

Success means `nh policy show` and `nh policy check` expose exact resolved
commits, exact policy-byte digests, full actor IDs, thresholds, sorted changes,
and actionable structural failures through one canonical parser/validator.
Invalid drafts must fail before creating an event or proposal ref.

## Context & Constraints

Read the [policy CLI contract](../contracts/policy-amendment-cli-v0.md),
[research decision D-001](../research.md), and the existing policy evaluation
in `policy.go`. `loadPolicy` and `validatePolicy` are the canonical authority;
do not fork their rules into the CLI layer.

The digest is over exact bytes, not normalized JSON. A policy candidate is
governed by the policy at its signed base commit. A valid head policy matters
only to later candidates whose base contains it. Identity continuity may be
shown informationally later, but cannot affect validity or role membership.

Top-level routing in `main.go` is intentionally deferred to WP06 so this
independent lane owns no shared router file.

## Branch Strategy

- **Strategy**: independent execution lane from `feat/self-hosting-alpha-loop`; merge through Spec Kitty into `feat/self-hosting-alpha-loop`
- **Planning base branch**: `feat/self-hosting-alpha-loop`
- **Merge target branch**: `feat/self-hosting-alpha-loop`

Use the lane workspace allocated by Spec Kitty. Keep all changes inside the
declared policy surfaces and commit with targeted staging.

## Subtasks & Detailed Guidance

### T011 – Write black-box policy show/check contract tests

- Start with tests in `policy_commands_test.go` using real temporary Git
  repositories and commits containing `.nh/policy.json`.
- Cover `show` defaulting to `HEAD`, an explicit revision, and an absent or
  malformed policy.
- Cover `check --base REV --head REV` and `check --base REV --file PATH`, plus
  missing/both proposed source flags and unexpected arguments.
- Assert resolved full commit IDs, exact full policy digests, full actor IDs,
  thresholds, pipeline details, and the base-governs statement.
- Compare output from differently ordered input maps/lists where semantics
  permit and require deterministic lexical rendering.
- Assert a rejected check creates no actor event, proposal event, candidate
  code ref, working-tree write, or commit.

### T012 – Unify exact policy loading from commits and draft files

- Refactor `policy.go` around one byte-oriented parse-and-validate function
  used by both existing commit loading and the new explicit draft loader.
- Preserve size limits, unknown-field rejection, trailing-JSON rejection,
  version checks, actor validation, and threshold validation.
- Return source context separately from the canonical `PolicyDocument`, exact
  bytes, and digest so errors can name `base` or `proposed` without duplicating
  validation rules.
- Resolve revision arguments to full commit object IDs before reading policy.
  Reject non-commit revisions and name the missing `.nh/policy.json` source.
- Read `--file` exactly once with the same maximum size and parser. Do not
  rewrite, format, stage, or commit the working-tree draft.
- Preserve existing `loadPolicy(commit)` behavior for proposal evaluation.

### T013 – Implement deterministic structural policy comparison

- Add focused comparison types/functions in `policy_commands.go`; they consume
  two already validated policy values and never decide candidate authority.
- Report added/removed maintainers, trusted reviewers, pipelines, and per-
  pipeline trusted runners using full actor IDs.
- Report changes to required accepts, required approvals, author approval,
  required pipeline results, and pipeline presence.
- Sort pipeline names and every actor collection lexicographically. Use stable
  explicit labels for unchanged versus changed scalar values.
- Treat raw-byte digest changes with structurally equal policy as an exact-byte
  change and display both digests even when the semantic change set is empty.
- Keep warnings factual: identify lockout/unsatisfiable rules through canonical
  validation, not speculative organizational judgments.

### T014 – Implement command handlers and proposal diagnostics

- Implement `cmdPolicy` dispatch plus `show` and `check` handlers in
  `policy_commands.go` with quiet flag sets and exact usage errors.
- `show` prints resolved commit, digest, maintainers/accept threshold,
  reviewers/approval rules, and pipelines/runners/results.
- `check` requires exactly one proposed source, validates both sides, prints
  both provenance records and deterministic changes, and explicitly says the
  base policy governs an amendment candidate.
- Provide a narrow helper that proposal opening can call when base/head policy
  bytes differ, without changing proposal creation semantics or creating a new
  event kind.
- Because `proposal.go` is outside this WP's owned surface, expose the helper
  and record any truly necessary integration request for WP06 rather than
  editing unrelated files here.
- Ensure errors never abbreviate trust-bearing actor/policy identifiers.

### T015 – Prove lockout rejection and exact base-policy authorization

- Test empty maintainers, malformed/duplicate actors, required accepts above
  maintainer count, approvals above trusted reviewers, invalid pipelines, and
  results above trusted runners on both base and proposed sides.
- Construct a candidate whose head adds a new reviewer/runner and relaxes
  author approval. Prove current evaluation ignores those head privileges.
- Prove only a later candidate based on the amended commit observes the new
  policy digest and roles.
- Confirm a valid continuity relation for an actor not listed in policy does
  not change validation or qualification.
- Retain current proposal/governance regression results and existing digest
  behavior.
- Mark T011–T015 complete only after focused and full checks succeed.

## Test Strategy

Required commands:

```bash
go test ./... -run 'Test.*Policy'
go test ./...
go test -race ./...
go vet ./...
```

Use CLI/public-function seams rather than inspecting internal slices. Exact
digest and actor assertions are appropriate; whitespace column alignment is
not. Include semantically equal but byte-different JSON to protect the
exact-bytes rule.

## Risks & Mitigations

- **Split validators**: factor canonical parsing once in `policy.go` and make
  all sources call it.
- **Nondeterministic output**: sort maps and actor sets explicitly.
- **Self-authorization**: repeat the base-governs fact in output and prove it
  with existing evaluator behavior.
- **Unexpected mutation**: `check --file` is read-only and tests compare Git
  status, refs, and file bytes before/after.
- **Digest confusion**: always label base and proposed digest/source together.

## Review Guidance

Reviewers should prepare a more permissive head policy and attempt to satisfy
its own candidate with a newly added actor. The attempt must remain
non-qualifying under the base policy. Independently verify digest computation
over exact bytes and deterministic ordering over multiple pipelines.

Reject duplicated validation logic, automatic policy editing, a new policy
event, abbreviated trust IDs, or any output implying continuity grants roles.

## Activity Log

> Append entries in chronological order. Status changes belong in the mission
> event log.

- 2026-08-30T17:26:50Z – system – Prompt created.
- 2026-08-30T18:32:19Z – codex – Formal cross-WP handoff: WP03 may make the narrow main.go command-route and proposal.go pre-mutation diagnostic wiring required to expose policy amendment inspection. WP06 owns main.go and WP05 owns proposal.go later; both must preserve or deliberately supersede this behavior during rebase. FR008 production continuity evidence is supplied by WP02 T010 and must be exercised, not replaced by synthetic continuity structs.
