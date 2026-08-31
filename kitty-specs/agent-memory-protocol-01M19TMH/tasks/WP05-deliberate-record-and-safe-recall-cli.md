---
work_package_id: "WP05"
title: "Deliberate Record and Safe Recall CLI"
dependencies: ["WP03", "WP04"]
requirement_refs:
  - "FR-001"
  - "FR-002"
  - "FR-003"
  - "FR-004"
  - "FR-005"
  - "FR-008"
  - "FR-009"
  - "FR-010"
  - "FR-011"
  - "FR-012"
  - "FR-013"
  - "FR-014"
  - "FR-015"
  - "FR-016"
  - "FR-018"
  - "FR-020"
  - "FR-021"
  - "FR-022"
  - "NFR-001"
  - "NFR-002"
  - "NFR-003"
  - "NFR-004"
  - "NFR-009"
  - "NFR-011"
  - "C-005"
  - "C-007"
  - "C-012"
subtasks: ["T021", "T022", "T023", "T024", "T025"]
owned_files:
  - "memory_commands.go"
  - "memory_commands_test.go"
  - "main.go"
authoritative_surface: "memory_commands.go"
create_intent:
  - "memory_commands.go"
  - "memory_commands_test.go"
execution_mode: "code_change"
task_type: "implement"
agent_profile: "implementer-ivan"
role: "implementer"
agent: "codex"
model: ""
---

# Work Package Prompt: WP05 – Deliberate Record and Safe Recall CLI

## ⚡ Do This First: Load Agent Profile

Use the `/ad-hoc-profile-load` skill to load the agent profile specified in the frontmatter, and behave according to its guidance before parsing the rest of this prompt.

- **Profile**: `implementer-ivan`
- **Role**: `implementer`
- **Agent/tool**: `codex`

If no profile is specified, run `spec-kitty agent profile list` and select the best match for this work package's `task_type` and `authoritative_surface`.

---

## Objective

Expose the Agent Memory Protocol through deliberate human commands and strict,
versioned, vendor-neutral JSON interfaces. Deliver bounded exact/lexical recall
whose author-controlled text remains inert data and whose every item preserves
complete provenance and independent classifications.

## Context

WP01 supplies strict signed memory envelopes, WP02 supplies independent local
and accepted streams, WP03 supplies lifecycle/applicability/evidence/trust
projection, and WP04 supplies the deterministic disposable index. This work
package composes those APIs; it must not duplicate their canonical validation,
silently repair their results, read quarantine refs, or make the index canonical.

Read `spec.md`, `plan.md`, `data-model.md`, `research.md`, all three contracts,
and the landed dependency APIs before implementation. The stable surface is
`contracts/memory-cli-v0.md`: human flags and strict JSON normalize into the
same internal requests, all trust-bearing inputs use full IDs, and recalled
author prose appears only beneath `memories[].data.content`.

Keep parsing, normalization, envelope composition, presentation, and bounded
errors in `memory_commands.go`. Modify existing `main.go` only to add the
top-level `memory` route and accurate usage text. Never capture prompts,
responses, terminal scrollback, clipboard contents, environment variables, or
arbitrary working-tree files. No recall path may execute, fetch, append, update
refs, invoke adapters, amend policy, or grant authority.

Run implementation through:

```sh
spec-kitty agent action implement WP05 --agent <name>
```

### Subtask T021: Implement strict versioned record input and command routing

**Purpose**

Provide one deliberate record path for human flags and versioned JSON while
adding the complete `nh memory` command family without changing legacy routing.

**Steps**

1. Start with public command tests for `nh memory record` using both human flags
   and `--input FILE|- --json`; make tests fail before production routing exists.
2. Define a `RecordRequestV0` or equivalent machine DTO with `version` exactly
   `0` and only vendor-neutral record, anchor, applicability, topic, evidence,
   attempt-outcome, handoff, actor, and optional full stream fields.
3. Decode JSON strictly with unknown-field rejection, one top-level value,
   trailing-data rejection, valid UTF-8, bounded reads, and field-oriented errors.
4. Do not accept omitted, negative, future, or string-coerced versions; version
   negotiation is explicit and fail-closed in version 0.
5. Support `FILE` and stdin (`-`) as exact explicit inputs. Never scan a directory,
   shell history, environment, clipboard, transcript, prompt, or response.
6. Reject ambiguous combinations of `--input`, positional content, and record
   flags instead of silently choosing one source.
7. Parse human `--kind`, `--at`, `--applies`, content, repeated topic/evidence,
   path/blob, subject, outcome, and stream inputs into the same internal request.
8. Resolve `--at REV` to one exact commit and every path/blob pair at that commit
   before signing; default only documented human conveniences, never machine JSON.
9. Load the existing local actor identity, derive the default stream when absent,
   inspect the stream head, then use WP01/WP02 APIs to build, sign, and CAS-append.
10. Let WP01 enforce record shape and public bounds; do not introduce a second
    validator whose normalization could produce different signed bytes or IDs.
11. Return full memory and stream IDs in stable JSON. Human output may add safe
    labels but must not substitute shortened IDs where subsequent action is expected.
12. Add `case "memory"` to `run` in `main.go`, dispatch all documented subcommands
    through `cmdMemory`, and add the exact command family to `printUsage`.
13. Route `memory index rebuild|verify` to WP04 without copying index behavior or
    allowing an index error to mutate canonical memory.
14. Preserve all existing command names, usage, exit behavior, stdout/stderr
    separation, and no-argument/help behavior outside the additive memory surface.

**Files**

- Create `memory_commands.go` for DTOs, strict input, parsing, and command dispatch.
- Create `memory_commands_test.go` for human/JSON equivalence and routing fixtures.
- Modify existing `main.go` narrowly for the `memory` switch arm and usage text.

**Validation**

- `go test ./... -run 'TestMemoryCommand.*(Record|Routing|Input)'` passes.
- Equivalent normalized human/JSON requests yield the same record shape when
  timestamp and stream-head inputs are fixed; unknown/future JSON fails closed.

### Subtask T022: Add handoff, lifecycle, and show commands

**Purpose**

Expose structured handoffs, immutable correction/dispute operations, and exact
inspection without erasing history or turning any recorded statement into action.

**Steps**

1. Add red public tests for `nh memory handoff`, `supersede`, `retract`,
   `challenge`, and `show`, covering human and JSON output where contracted.
2. Implement `handoff --at REV --applies MODE --input FILE|- --json` as an
   explicit handoff-record adapter, never as automatic session-state capture.
3. Require the handoff input to preserve separate bounded `completed`,
   `assumptions`, `blockers`, and `nextActions` lists plus ordinary record content.
4. Treat every next action as inert proposed data; do not invoke it, authorize it,
   schedule it, translate it to shell, or call an agent/tool callback.
5. Implement `supersede MEMORY` with a complete replacement record and the same
   normalized record-input pipeline used by `record`.
6. Require a full target memory ID on supersede, retract, and challenge inputs;
   short IDs are permitted only for unambiguous read-only human lookup in `show`.
7. Resolve the target from accepted/local verified memory and let WP03 enforce
   same-author supersession/retraction and cross-author challenge constraints.
8. Implement `retract MEMORY --reason REASON` as an appended signed fact; never
   rewrite the target commit, delete a ref, alter an index as canonical state,
   or imply replicated Git objects were erased.
9. Implement `challenge MEMORY --reason REASON` plus repeatable typed
   `--evidence`; preserve target attribution and never label a challenge as truth.
10. Append lifecycle operations through the actor's selected/default WP02 stream
    with normal sequence, previous-ID, signing, and CAS conflict behavior.
11. Implement `show MEMORY [--json]` over verified canonical sources, returning
    the exact envelope/projection and full IDs without fetching missing data.
12. Include lifecycle edges, applicability, evidence, trust, signature status,
    anchor, digest, and exact missing-dependency details in `show` output.
13. Put author content only in nested JSON `data`; render human text through
    existing `safeText`/`oneLine`-equivalent control-safe helpers and a warning.
14. Reject incompatible flags and operation-specific fields with bounded errors
    that name safe public IDs but never echo unbounded content or secrets.

**Files**

- Extend `memory_commands.go` with handoff, lifecycle append, and show handlers.
- Extend `memory_commands_test.go` with immutable-history and safe-output cases.
- Touch `main.go` only if the usage block needs these subcommands enumerated.

**Validation**

- `go test ./... -run 'TestMemoryCommand.*(Handoff|Supersede|Retract|Challenge|Show)'` passes.
- Git refs only advance by new signed facts; originals remain inspectable and no
  command executes or authorizes handoff/challenge content.

### Subtask T023: Implement exact-filtered recall with complete provenance

**Purpose**

Compose accepted projections and the WP04 lexical index into deterministic
recall while retaining every required identity and orthogonal classification.

**Steps**

1. Define strict `RecallRequestV0` machine input and `RecallEnvelopeV0` output
   matching `data-model.md`; require response version `0` and a constant warning.
2. Accept human filters for current/exact commit, exact subject, path, repeated
   topic, kind, actor, lifecycle, trust, lexical query, and explicit untrusted view.
3. Default human `atCommit` to the exact current `HEAD`; do not let JSON omission
   ambiguously bind to mutable state unless the frozen DTO explicitly documents it.
4. Normalize topics/query and validate every enum, full trust-bearing ID, path,
   and commit once before index lookup or cursor calculation.
5. Load policy bytes at the exact requested commit and use WP03 trust classes;
   never read mutable working-tree policy as a substitute.
6. Collect only verified local and accepted memory streams through WP02, then
   verify/rebuild the WP04 private index without network access when required.
7. Apply exact filters before deterministic lexical token intersection and do
   not add embeddings, model inference, fuzzy matching, or prose-based relevance.
8. Default to active, applicable, evidence-resolved, policy-qualified records;
   keep each reason for exclusion independently inspectable.
9. Make `--include-untrusted` include valid non-qualifying claims with their real
   `actor-untrusted`, `kind-untrusted`, or `policy-missing` class unchanged.
10. Preserve full memory ID, actor, stream, signature status, exact anchor and
    path/blob pairs, applicability, lifecycle summary and every edge ID.
11. Preserve evidence status and details, trust class, exact content digest,
    record kind/topics, and sorted missing dependencies with recovery guidance.
12. Return author prose only as `memories[].data.content`; structured handoff
    fields exist only beneath `memories[].data.handoff`.
13. Never synthesize `instruction`, `command`, `tool`, `authorization`, truth,
    confidence, or inferred-conflict fields from content or metadata.
14. Use stable ordering defined by the plan/index API: anchor relevance,
    lifecycle class, signed timestamp, then full memory ID.

**Files**

- Extend `memory_commands.go` with recall DTOs, filter parsing, and composition.
- Extend `memory_commands_test.go` with every filter and provenance field case.
- Reuse WP03/WP04 data rather than copying projection or tokenization logic.

**Validation**

- `go test ./... -run 'TestMemoryRecall.*(Filter|Provenance|Classification)'` passes.
- Every returned item contains all NFR-001 fields and explicit untrusted recall
  changes inclusion only, never policy or classification.

### Subtask T024: Enforce bounds, query-bound cursors, and inert output

**Purpose**

Make every recall page resource-bounded and continuation-safe while proving
hostile author bytes cannot escape the JSON/human data boundary or cause effects.

**Steps**

1. Add count-bound tests at one below, exactly, and one above the default of 20,
   plus explicit positive overrides and rejection of zero/negative values.
2. Add encoded-content tests at one below, exactly, and one above the default
   65,536-byte budget, including multibyte UTF-8, escapes, and repetitive text.
3. Define precisely whether the byte budget counts JSON-encoded `data.content`
   values or their containing data objects, then freeze that contract in tests.
4. Sort the full deterministic match set first; apply cursor, record count, and
   content-byte limits only afterward so input/ref enumeration cannot change pages.
5. Populate `matched`, `returned`, `truncated`, and `nextCursor` consistently;
   omit the cursor only when no deterministic match remains.
6. Bind each opaque cursor cryptographically to the normalized request including
   filters and bounds, accepted-source fingerprint, policy digest, and last sort key.
7. Reject malformed cursors and cursors reused after a source, policy, filter,
   query, commit, or bound change; never silently start or continue another query.
8. Ensure cursor material contains no author prose, key material, credentials,
   environment values, filesystem paths, or unsigned mutable server state.
9. JSON-encode controls, newlines, terminal escapes, quotes, backslashes, markup,
   and plausible shell/prompt/tool instructions only inside nested data strings.
10. Emit the constant inert-data warning outside author-controlled data, with no
    author value able to replace, prefix, suppress, or interpolate into it.
11. Render human output with the same warning and complete classification labels,
    passing every author-controlled value through control-safe display helpers.
12. Assert recall performs no command execution, process spawn, network fetch,
    event/memory append, ref/index mutation beyond allowed local rebuild, or callback.
13. Keep errors bounded and safe: name invalid fields/full public IDs without
    echoing hostile content, private keys, tokens, environment, or credentials.
14. Compare repeated runs and paginated traversal byte-for-byte; concatenated
    pages must equal the unbounded deterministic match order without gaps/duplicates.

**Files**

- Extend `memory_commands.go` with page budgeting, cursor encoding, and renderers.
- Extend `memory_commands_test.go` with hostile JSON/human and cursor matrices.
- Do not add a cursor database, cache authority, adapter callback, or network seam.

**Validation**

- `go test ./... -run 'TestMemoryRecall.*(Bound|Cursor|Inert|Human)' -count=3` passes.
- Hostile fixtures produce zero effects, valid output, deterministic continuation,
  and exact default count/content-byte enforcement.

### Subtask T025: Prove neutral adapters, deliberate capture, and CLI compatibility

**Purpose**

Close the public CLI boundary with two vendor-neutral adapter fixtures, privacy
and inertness evidence, and a complete regression proof for pre-memory commands.

**Steps**

1. Build two independently shaped test adapters that communicate only through
   the frozen version-0 record request and recall envelope JSON byte interfaces.
2. Keep adapters inside `memory_commands_test.go` as neutral fixture code; do not
   name a model vendor, hosting provider, agent harness, or proprietary field.
3. Have both adapters submit equivalent normalized records with fixed identity,
   timestamp, stream/head context, and exact anchors; assert identical canonical
   payload bytes, full memory IDs, and stored record fields.
4. Have both consume a mixed bounded recall envelope and assert preservation of
   every full provenance, lifecycle, applicability, evidence, trust, and digest field.
5. Include all six record kinds across adapter/public CLI fixtures and exercise
   record, handoff, supersede, retract, challenge, show, recall, and index routing.
6. Seed prompt-like overrides, shell commands, tool-call-shaped JSON, ANSI/control
   bytes, newlines, markup, Unicode, and high repetition as author content.
7. Instrument process, callback, file/ref, policy, and network boundaries where
   practical; assert the hostile corpus causes no execution or authority change.
8. Place sentinel secrets in environment variables, actor private-key storage,
   terminal/transcript-like files, clipboard-like fixtures, and unrelated worktree files.
9. Record only explicit request data, then inspect canonical payloads, index data,
   stdout/stderr, errors, and recall output for zero sentinel-secret leakage.
10. Prove failed/ambiguous/oversized input appends nothing and leaves memory,
    collaboration, policy, and proposal refs unchanged.
11. Snapshot representative legacy command behavior and run every existing CLI
    test; `nh help`, no-arg output, errors, and all non-memory routes remain valid.
12. Assert collaboration event payload bytes, public IDs, collection results,
    and collaboration-only operation remain unchanged with no memory refs/index.
13. Run the memory command suite repeatedly and with the race detector to expose
    global output hooks, shared cursors, mutable request DTOs, or ordering races.
14. Keep all test repositories temporary and offline; add no service, model API,
    provider API, Docker step, external fixture, new dependency, or real secret.

**Files**

- Complete `memory_commands_test.go` with both adapter and regression fixtures.
- Adjust `memory_commands.go` and the narrow `main.go` routing diff only for
  behavior exposed by public tests; do not edit dependency modules to bypass APIs.

**Validation**

- `go test ./... -run 'TestMemory(Command|Recall|Adapter)'` passes repeatedly.
- `go test -race ./...`, `go vet ./...`, `go build ./...`, and `git diff --check`
  pass with zero secret sentinel occurrence and unchanged legacy CLI behavior.

## Definition of Done

- T021 has an event-sourced `done` record and strict version-0 JSON plus human
  record input share one normalized signing/append path with additive routing.
- T022 has an event-sourced `done` record and handoff, supersede, retract,
  challenge, and show append or inspect immutable signed facts correctly.
- T023 has an event-sourced `done` record and every exact/lexical recall filter
  returns complete full-ID provenance and separate classifications.
- T024 has an event-sourced `done` record and count/content bounds, deterministic
  truncation, query/source/policy-bound cursors, and inert JSON/human output pass.
- T025 has an event-sourced `done` record and two neutral adapters, deliberate
  capture, secret isolation, hostile content, and legacy CLI regressions pass.
- JSON machine input is strict and versioned; no vendor-specific canonical field,
  unknown field, trailing value, automatic capture source, or ambiguous input wins.
- Recall defaults to at most 20 records and 65,536 encoded content bytes, and
  pagination reconstructs stable ordered results without gaps or duplicates.
- All author prose is nested inert data; no command grants authorization,
  executes content, invokes tools/adapters, fetches data, or mutates policy.
- Only `memory_commands.go`, `memory_commands_test.go`, and the narrow additive
  `main.go` routing/usage surface change, with create intent limited to new files.
- Focused tests, `go test ./...`, `go test -race ./...`, `go vet ./...`,
  `go build ./...`, and `git diff --check` pass.
- After each subtask's evidence exists, record completion separately:

```sh
spec-kitty agent tasks mark-status T021 --status done
spec-kitty agent tasks mark-status T022 --status done
spec-kitty agent tasks mark-status T023 --status done
spec-kitty agent tasks mark-status T024 --status done
spec-kitty agent tasks mark-status T025 --status done
```

## Risks

- **Parser divergence**: human and JSON inputs could sign different shapes;
  normalize once and compare fixed-time/fixed-head adapter fixtures.
- **Prompt-injection escape**: author text could enter warnings, labels, errors,
  or terminal controls; keep it nested and encode/sanitize every presentation.
- **Cursor confusion**: a token might continue after sources or filters change;
  bind the complete normalized request, bounds, policy, source set, and sort key.
- **Budget ambiguity**: raw versus encoded bytes can drift; document one rule,
  test escaping/multibyte boundaries, and apply bounds after stable ordering.
- **Classification collapse**: convenient trusted/valid flags can hide gaps;
  preserve signature, lifecycle, applicability, evidence, and trust separately.
- **Secret capture**: generic stdin/files or diagnostics may leak ambient data;
  read only explicitly named input, bound it, and scan every observable artifact.
- **Legacy CLI drift**: broad `main.go` edits may alter unrelated behavior;
  keep routing additive and run the complete existing command suite.
- **Index authority creep**: CLI convenience may trust stale derived state;
  verify/rebuild from canonical accepted refs and never mutate records from index data.

## Reviewer Guidance

Review JSON boundaries before happy-path UX. Confirm machine DTOs reject unknown
fields, future/missing versions, trailing JSON, ambiguous input sources, and
unbounded reads. Trace equivalent human and adapter input through exactly one
normalization, WP01 validation/signing, and WP02 CAS append path.

Inspect every recall item for full memory/actor/stream IDs, signature, anchor,
applicability, lifecycle plus edge IDs, evidence details, trust class, content
digest, and nested inert data. Recompute boundary cases for encoded byte counts
and attempt cursor reuse after changing each query/source/policy/bounds component.

Demand effect evidence for hostile content: JSON must remain parseable, human
output control-safe, warnings immutable, and no execution, callback, fetch,
authorization, append, ref, or policy path may consume author prose. Finally,
run all legacy CLI tests and inspect `main.go` for an additive route only.

## Implementation Command

```bash
spec-kitty agent action implement WP05 --agent <name>
```
