---
work_package_id: WP07
title: Operational Agent Handoff and Living Protocol
dependencies:
- WP05
- WP06
requirement_refs:
- FR-001
- FR-005
- FR-006
- FR-007
- FR-008
- FR-009
- FR-010
- FR-011
- FR-012
- FR-013
- FR-014
- FR-015
- FR-016
- FR-017
- FR-018
- FR-019
- FR-020
- FR-021
- FR-022
- FR-023
- FR-024
- NFR-001
- NFR-002
- NFR-003
- NFR-005
- NFR-006
- NFR-007
- NFR-008
- NFR-009
- NFR-010
- NFR-011
- NFR-012
- C-001
- C-009
- C-011
- C-012
planning_base_branch: feat/agent-memory-protocol
merge_target_branch: feat/agent-memory-protocol
branch_strategy: Planning artifacts for this mission were generated on feat/agent-memory-protocol. During /spec-kitty.implement this WP may branch from a dependency-specific base, but completed changes must merge back into feat/agent-memory-protocol unless the human explicitly redirects the landing branch.
subtasks:
- T031
- T032
- T033
- T034
- T035
history: []
agent_profile: reviewer-renata
authoritative_surface: memory_acceptance_test.go
create_intent:
- memory_acceptance_test.go
- docs/memory-v0.md
- docs/memory-safety.md
execution_mode: code_change
model: ''
owned_files:
- memory_acceptance_test.go
- .nh/policy.json
- README.md
- docs/protocol-v0.md
- docs/replication-v0.md
- docs/host-compatibility.md
- docs/memory-v0.md
- docs/memory-safety.md
role: reviewer
tags: []
task_type: implement
tracker_refs: []
---

# Work Package Prompt: WP07 – Operational Agent Handoff and Living Protocol

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `reviewer-renata`
- **Role**: `reviewer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Prove the Agent Memory Protocol through one public-command, two-actor handoff
that survives correction, selective replication, index loss, and a fresh clone,
then make the public documentation and mission contracts describe exactly the
shipped behavior and its security limits. Produce repeatable ordinary-Git
evidence without publishing private state or relying on a vendor service.

## Context

WP01–WP06 supply the signed envelope, independent streams, deterministic
projection, disposable index, deliberate record/recall CLI, and selected
quarantine replication. This final package must consume those public surfaces;
it must not add test-only bypasses, duplicate core validators, weaken a failed
dependency, or rewrite implementation merely to match an obsolete draft.

Read all mission design documents and contracts, the landed dependency APIs,
`README.md`, every current public document, and `operational_acceptance_test.go`.
Follow its real-binary, temporary-repository, isolated-environment, ordinary-Git
style. New black-box coverage belongs in `memory_acceptance_test.go`; reuse
helpers where compatible, but do not edit the prior operational acceptance file.

The living protocol must keep five properties visibly separate: signature
validity, policy qualification, evidence resolution, semantic truth, and prompt
authority. Retraction is an auditable fact, not an erasure guarantee. Selection
is transport authorization, not trust. A handoff's proposed next actions are
inert data and never permission to act.

Run implementation through:

```sh
spec-kitty agent action implement WP07 --agent <name>
```

### Subtask T031: Prove a two-actor correction and handoff in a fresh clone

**Purpose**

Create the black-box operational scenario that records project cognition with
two independent actors, corrects it immutably, hands work off, selects exact
streams, and reconstructs bounded recall from an identity-free fresh clone.

**Steps**

1. Create `memory_acceptance_test.go` and a top-level
   `TestOperationalAgentMemory` that builds the real `nh` binary once.
2. Create separate author and successor-agent repositories plus an ordinary
   local bare remote; initialize each actor independently through public CLI.
3. Isolate home, global Git config, credential helpers, askpass, tokens, and
   Nichthub test variables using the established operational helper pattern.
4. Fail if the actors or public keys match, or if either clone receives the
   other actor's private identity/keyring state.
5. Commit a policy whose optional memory section qualifies the exact actors and
   all intended kinds while preserving existing governance fields and digest rules.
6. Through public commands only, record a corpus covering all six record kinds,
   exact commit/path or subject anchors, topics, applicability, and typed evidence.
7. Include an unsuccessful attempt with outcome, a verification with resolvable
   evidence, and an explicit handoff with completed, assumptions, blockers, and next actions.
8. Have the original author supersede one stale record and retract another;
   have the second actor challenge an exact memory with evidence.
9. Prove the original, replacement, retraction, and challenge remain immutable
   and the projected active/branch/dispute state uses full exact IDs.
10. Publish actor, proposal when needed, and memory refs through `nh sync`; save
    exact memory selectors rather than using an implicit broad fetch.
11. Create a new credential-disabled clone from the bare remote with no `.git/nh`
    keyring, copied index, embeddings, adapter cache, or vendor state.
12. Select the exact supplying memory streams, synchronize through quarantine,
    rebuild the index, and recall at the exact current commit with public CLI.
13. Decode the recall JSON and compare every expected full memory/actor/stream
    ID, anchor, lifecycle edge, evidence status, trust class, digest, and inert data.
14. Assert the bounded handoff identifies completed work, active assumptions,
    blockers, and proposed next actions without executing or authorizing them.
15. Delete the fresh clone's private index, rebuild twice, and require identical
    membership, classifications, ordering, and exact-filter results.

**Files**

- Create `memory_acceptance_test.go` for the public two-actor flow and narrowly
  shared acceptance helpers; do not modify `operational_acceptance_test.go`.
- Modify `.nh/policy.json` only to add the validated optional memory policy used
  by the repository's living example; preserve existing governance authority.

**Validation**

- `go test ./... -run TestOperationalAgentMemory -count=3 -v` passes offline.
- The verifier has no signer/index before sync, reconstructs exact results after
  sync, and cannot author as either originating actor.
- Record completion with `spec-kitty agent tasks mark-status T031 --status done`.

### Subtask T032: Lock adversarial inertness and legacy collaboration isolation

**Purpose**

Exercise the integrated safety boundary with hostile memory while preserving
every established collaboration payload, public ID, projection, and command.

**Steps**

1. Extend `memory_acceptance_test.go` with black-box hostile recall cases; reuse
   the real binary, temporary repositories, and environment isolation from T031.
2. Record instruction-like prose, shell fragments, tool-call-shaped JSON,
   system-prompt language, ANSI escapes, controls, newlines, markup, Unicode,
   and highly repetitive text only through explicit record inputs.
3. Place unique sentinel secrets in environment, credential-helper config,
   keyring files, transcript-like files, and unrelated working-tree files.
4. Instrument a marker executable/file and other practical effect boundaries;
   recall and adapter consumption must create no marker or subprocess effect.
5. Require hostile prose to appear only in JSON `memories[].data.content` or
   nested handoff data, with the constant warning outside author-controlled data.
6. Parse JSON to prove controls remain encoded and cannot replace warnings,
   field names, provenance, classifications, or continuation metadata.
7. Exercise default and explicit bounds, truncation, pagination, and cursor
   invalidation after source, policy, filter, query, commit, or bound changes.
8. Include untrusted, inactive, challenged, and dependency-missing records;
   explicit inspection may reveal them but never upgrade their classification.
9. Synchronize a mixed remote with valid collaboration plus invalid, missing,
   over-budget, and unselected memory, then inspect independent outcomes.
10. Snapshot collaboration refs, exact event payload bytes, public event IDs,
    `collectEvents`-observable output, policy bytes, and branch HEAD before memory failure.
11. Require every snapshot to remain byte-identical and every independently
    valid issue/proposal/review/CI/governance inspection to remain usable.
12. Run the existing all-kind legacy ID fixture without changing any golden;
    add an explicit regression assertion only if current coverage lacks one.
13. Search canonical payloads, index bytes, stdout/stderr, errors, and published
    refs for all ambient sentinels; only deliberately supplied memory may appear.
14. Keep fixtures offline and free of real credentials, provider APIs, services,
    model APIs, Docker, nondeterministic timing, and test-only production bypasses.

**Files**

- Extend only `memory_acceptance_test.go`; fixes exposed in dependency-owned
  implementation files must return to their owning WP rather than leaking here.

**Validation**

- Run `go test ./... -run 'TestOperationalAgentMemory|TestIdentityFieldsPreserveEveryExistingEventKindID' -count=3 -v`.
- Run `go test ./...`, `go test -race ./...`, `go vet ./...`, `go build ./...`,
  and `git diff --check`; all sentinel/effect and collaboration snapshots pass.
- Record completion with `spec-kitty agent tasks mark-status T032 --status done`.

### Subtask T033: Publish the memory protocol, safety policy, and operator surface

**Purpose**

Turn the implemented protocol into public, navigable operator and adapter
documentation without overstating truth, authority, deletion, or host support.

**Steps**

1. Create `docs/memory-v0.md` as the public protocol/operator reference linked
   from the README and the existing collaboration protocol document.
2. Document `nh-memory/0`, exact two-file commit trees, full IDs, actor-owned
   stream refs, sequence/previous continuity, operations, kinds, and bounds.
3. Describe exact anchors, path/blob pairs, applicability, typed evidence, and
   the deterministic lifecycle rules for supersede, retract, challenge, and branching.
4. Document the optional `.nh/policy.json` memory section, exact-commit policy
   evaluation, default qualification, and explicit untrusted inspection.
5. Explain all recall filters, stable ordering, default 20-record/65,536-byte
   limits, continuation invalidation, and complete provenance requirements.
6. Document the private `index-v0.json` location, source fingerprint, strict
   verification, owner-only handling, deletion/rebuild, and noncanonical status.
7. Include versioned machine record/recall examples and state that vendor-neutral
   adapters preserve content beneath the inert data boundary.
8. Create `docs/memory-safety.md` as the threat model for hostile content,
   deliberate capture, secrets, quarantine, bounds, policy, and adapter duties.
9. State explicitly that signatures prove attribution/integrity only; policy
   qualification is not truth, evidence availability is not correctness, and
   no memory or handoff grants prompt priority or operational authority.
10. State that retraction and local index deletion do not erase replicated Git
    objects; redaction, legal deletion, moderation, and retention remain deferred.
11. Extend `README.md` with the memory value proposition, concise record/handoff/
    recall/select/index examples, command surface, documentation links, and alpha limits.
12. Update `docs/protocol-v0.md` to preserve the frozen `nh/0` boundary and link
    memory as a separate additive protocol rather than inserting memory event kinds.
13. Update `docs/replication-v0.md` with memory selectors, exact namespaces,
    quarantine validation, independent outcomes, publication, and shallow recovery.
14. Update `docs/host-compatibility.md` only with directly observed memory-ref
    transport results, dates, exact verification method, and honest untested limits.
15. Add the sorted memory actor/kind section to `.nh/policy.json` without
    changing maintainers, proposal thresholds, pipeline roles, or policy version.

**Files**

- Create `docs/memory-v0.md` and `docs/memory-safety.md`.
- Modify `README.md`, `.nh/policy.json`, `docs/protocol-v0.md`,
  `docs/replication-v0.md`, and `docs/host-compatibility.md`.

**Validation**

- Run every documented local command against a disposable repository and check
  links, ref spellings, defaults, full-ID examples, and JSON fields against code.
- Search docs for claims that collapse trust dimensions or promise execution/
  erasure; record completion with `spec-kitty agent tasks mark-status T033 --status done`.

### Subtask T034: Prove shipped behavior conforms to frozen mission contracts

**Purpose**

Treat the mission contracts and quickstart as read-only acceptance authorities,
and prove the shipped code and public documentation conform without editing
coordination artifacts from an implementation lane.

**Steps**

1. Diff each frozen contract and `quickstart.md` against the landed public CLI,
   wire structs/tags, validation constants, ref parsers, projection, and outcomes.
2. Verify encoded field names, omission rules, operation/kind shapes, bounds,
   signatures, and ref grammar exactly; correct owned code when it diverges.
3. Preserve the distinction between payload memory ID, Git commit ID, actor
   fingerprint, stream ID, collaboration event ID, and typed evidence ID.
4. Verify subcommands, flags, strict request/response versions, defaults, error
   classes, output fields, cursor rules, and encoded-content accounting.
5. Verify saved-selection compatibility, memory discovery, budgets, independent
   outcomes, transaction behavior, publication, and recovery against the contract.
6. Exercise every quickstart phase in disposable repositories, including record,
   handoff, correction, policy, recall, index, selection, sync, and fresh clone.
7. Keep signer-required commands distinct from read-only fresh-clone commands;
   never imply that a clone receives or recovers another actor's private key.
8. Reconcile public docs with full-ID, missing-versus-invalid, selected-supplier,
   no-global-unshallow, and no-network-recall behavior.
9. Use one vocabulary for canonical memory, accepted refs, quarantine,
   qualification, evidence resolution, lifecycle, and inert data.
10. Retain explicit limits: repository-local v0, no federation, semantic truth,
    autonomous action, automatic capture, redaction, or erasure guarantee.
11. If shipped behavior violates a contract or security invariant, return the
    defect to its owning WP; do not weaken or edit `kitty-specs/` from this lane.
12. Record a concise contract-conformance table in the WP review evidence so
    mission acceptance can audit every frozen surface.

**Files**

- Read the three mission contracts and `quickstart.md` without modifying them.
- Correct only WP07-owned source/public-documentation surfaces when they drift.

**Validation**

- Run quickstart smoke commands with the built binary and strict-decode every
  JSON example; compare all documented constants/ref forms against production definitions.
- Confirm `git diff` contains no `kitty-specs/` changes from the WP07 lane.
- `git diff --check` passes; record completion with
  `spec-kitty agent tasks mark-status T034 --status done`.

### Subtask T035: Produce repeatable ordinary-Git public-repository evidence

**Purpose**

Demonstrate that the operational handoff survives a public Git endpoint and a
credential-disabled fresh verifier, while publishing no secret or local cache
and recording exact, independently checkable transport and recall evidence.

**Steps**

1. Run the exact proposed public sequence first against a new local bare remote;
   require T031/T032 and all quality gates to pass before any public mutation.
2. Use only a user-authorized public repository already in mission scope and
   ordinary `git`/`nh` commands; do not use hosting-provider APIs or UI claims.
3. Record the endpoint, UTC observation time, current advertised branch and
   `refs/nh/*` OIDs with `git ls-remote` before publication.
4. Review explicit source and destination refs before pushing; use no wildcard,
   deletion, force update, tag rewrite, or unrelated primary-branch mutation.
5. Publish only the exact two actor histories and selected memory streams needed
   for the corpus, plus separately governed code/policy refs if the proof requires them.
6. Keep private keys, `.git/nh` keyring/active state, replication selections,
   indexes, transaction journals, credentials, and environment state local.
7. Verify post-publication advertisement with `git ls-remote`, recording every
   full actor/stream ref and exact Git OID rather than trusting push output.
8. Clone the advertised public branch into a new verifier with system/global
   Git config, credential helpers, terminal prompts, askpass, tokens, and SSH agent disabled.
9. Prove the clone initially has no private identity, local memory index,
   embeddings, adapter state, Docker state, or local/accepted memory refs.
10. Save exact stream selectors and positive budgets, run `nh sync`, rebuild the
    index, and issue the same exact bounded handoff recall used by the publisher.
11. Compare full memory IDs, actor/stream ownership, anchors, lifecycle edges,
    evidence/dependency status, trust classes, query digest, ordering, and handoff data.
12. Exercise one deliberate index deletion/rebuild and, when the fixture exposes
    a shallow gap, only the documented exact selected recovery flow.
13. Record actual commands, exit results, advertised OIDs, policy digest, stream
    and memory IDs, recall counts/digest, absence checks, and observed timings.
14. Add those exact observations to `docs/host-compatibility.md` and the public
    memory documentation; never prefill IDs, redact them to prefixes, or fabricate a pass.
15. State what the proof does not establish: distinct humans, truth, authority,
    permanent host retention, portable pre-download quotas, deletion, or provider generality.
16. If authorization or public transport is unavailable, leave the evidence
    explicitly incomplete and report the gate; a disposable-remote pass cannot
    be mislabeled as the required public-repository proof.

**Files**

- Extend `docs/host-compatibility.md` with the exact public observation and
  verification recipe; update `docs/memory-v0.md`/`README.md` links if needed.
- Do not add keys, indexes, captured credentials, raw private state, generated
  vendor metadata, Docker artifacts, or unbounded transcripts to the repository.

**Validation**

- A reader can repeat the documented `git ls-remote`, clone, exact selection,
  sync, index rebuild, and recall checks using only public data and the `nh` binary.
- Run secret/index/private-state scans and all mission gates; record completion
  with `spec-kitty agent tasks mark-status T035 --status done` only after the
  exact public evidence exists.

## Definition of Done

- T031 has event-sourced `done` evidence for the real-binary two-actor record,
  correction, handoff, exact selection, fresh-clone recall, and index rebuild loop.
- T032 has event-sourced `done` evidence for hostile inert content, zero effects,
  secret isolation, memory/collaboration failure isolation, and unchanged legacy IDs.
- T033 has event-sourced `done` evidence for navigable public protocol, policy,
  index, safety, replication, host, README, and command documentation.
- T034 has event-sourced `done` evidence that every contract and quickstart
  command/field/limit matches shipped behavior and preserves security language.
- T035 has event-sourced `done` evidence only after the public ordinary-Git
  advertisement and credential-disabled fresh-clone verification are recorded.
- The acceptance suite uses public commands and real temporary Git repositories;
  no test-only core bypass is accepted as operational proof.
- All recall items retain complete full-ID provenance and separate signature,
  applicability, lifecycle, evidence, and policy classifications.
- Memory content and handoff next actions remain inert data with zero execution,
  callback, network, append, ref, policy, or authority side effect.
- Existing collaboration payload bytes, public IDs, refs, projections, command
  behavior, and collaboration-only clone results remain unchanged.
- Only the owned files change; create intent is limited to
  `memory_acceptance_test.go`, `docs/memory-v0.md`, and `docs/memory-safety.md`.
- `gofmt -w memory_acceptance_test.go`, `git diff --check`, `go test ./...`,
  `go test -race ./...`, `go vet ./...`, and `go build ./...` all pass.

## Risks

- A black-box test can accidentally call package internals; invoke the built
  binary and inspect Git/JSON artifacts through public, read-only boundaries.
- Fresh-clone proof can inherit credentials or private state; sanitize process
  environment and inspect the clone before and after selected synchronization.
- Hostile prose can escape through human labels/errors even when JSON is safe;
  test both renderers and effect boundaries with controls and sentinel commands.
- A broad memory failure can regress collaboration; retain exact byte/ref/ID
  snapshots around invalid, missing, over-budget, and unselected streams.
- Documentation can turn attribution into truth or handoff into authority;
  use the five-way classification language consistently across every artifact.
- Retraction language can imply deletion; state immutable replication and the
  lack of erasure/redaction guarantees at each operator-facing correction flow.
- Public proof can leak keys, indexes, credentials, or local paths; publish only
  reviewed exact refs and scan both tracked changes and advertised objects.
- Public host behavior can change; record dated direct Git observations and
  avoid extrapolating beyond the exact provider/configuration tested.
- A local-bare pass may be presented as public evidence; keep the two evidence
  layers separate and leave T035 incomplete until public verification exists.
- Contract reconciliation can hide a security regression; send invariant
  conflicts back to the owning WP instead of documenting unsafe behavior.

## Reviewer Guidance

Run `TestOperationalAgentMemory` from a clean checkout and trace every external
effect. Verify both actors were independently initialized, all memory was
deliberately recorded, correction remained append-only, selectors were exact,
and the fresh clone received neither signer nor index. Delete the index and
recompute the same handoff result. Inspect hostile JSON and human output to
confirm author prose never escapes nested inert data or causes an effect.

Compare representative collaboration payloads and every frozen legacy ID
before and after memory failures. Review policy trust, evidence availability,
applicability, lifecycle, signature validity, semantic truth, and prompt
authority as separate concepts. Reject docs that imply selection is trust,
verification is truth, challenge is automatic contradiction, or retraction is
erasure.

For the public proof, independently run the recorded `git ls-remote` and
fresh-clone recipe. Confirm all published IDs/OIDs are full, the endpoint and
date are explicit, private state is absent, and no provider API, mandatory
service, model API, Docker operation, copied index, or copied key participates.
Do not approve T035 from a local-only substitute or an unverified transcript.

## Implementation Command

```bash
spec-kitty agent action implement WP07 --agent <name>
```

## Activity Log

- 2026-08-31T18:05:21Z – codex – shell_pid=2156508 – Contract conformance: wire PASS (nh-memory/0 two-file signed streams, full IDs, lifecycle and bounds); CLI PASS (strict v0 record/handoff/recall, frozen handoff composition, filters/cursors/index); replication PASS (exact memory selectors, quarantine, independent outcomes, accepted refs, recovery); quickstart PASS (real binary, two actors, correction/challenge, fresh clone, deterministic rebuild); public Git PASS (GitHub HTTPS proof branch plus two actor/two memory refs, credential-disabled verifier, full OIDs/digests recorded). Gates: focused x3, full, race, vet, build, diff, links, secret/private-state scan.
- 2026-08-31T18:05:55Z – codex – shell_pid=2156508 – for_review transition blocked after successful implementation: lane contains 20 inherited committed kitty-specs planning files relative to feat/agent-memory-protocol; pre-review coverage authority tests.architectural._gate_coverage is unavailable in this Go repository. No force or kitty-specs mutation attempted.
