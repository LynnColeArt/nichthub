# Mission Specification: `hn` Hard Cutover

**Mission Branch**: `feat/hn-hard-cutover`  
**Created**: 2026-09-01  
**Status**: Draft  
**Input**: Rename Hubnot's complete active `nh` command, storage, ref, protocol,
and environment namespace to `hn` without backward compatibility because the
project has no users.

## Mission Intent

This mission renames the old active namespace `nh` to `hn` throughout Hubnot.
The exact per-category rules are recorded in `occurrence_map.yaml` before
implementation. This is a deliberate protocol reset, not a compatibility
feature: the resulting program exposes only `hn` surfaces and neither reads nor
writes legacy state.

The historical `refs/nh/*` facts, the checked-in `.nh/` policy and pipeline that
govern the transition, old self-hosting reports, completed mission artifacts,
and immutable event journals remain byte-for-byte historical evidence. They are
not compatibility inputs. Active policy and pipeline configuration is added
under `.hn/`, and new local identities and public facts start independently in
`.git/hn/` and `refs/hn/*`.

## User Scenarios & Testing

### User Story 1 - One unambiguous command (Priority: P1)

As a Hubnot operator, I invoke `hn` everywhere so the product name and command
namespace no longer disagree.

**Why this priority**: A command rename that leaves aliases, diagnostics, build
instructions, or generated binaries behind is not a clean cutover.

**Independent Test**: Build `./hn`, execute representative commands and error
paths, and scan every active source and document for obsolete `nh` command
forms.

**Acceptance Scenarios**:

1. **Given** a fresh Git repository and the built program, **when** an operator
   runs `hn init`, opens an issue, inspects the log, and synchronizes, **then**
   each operation succeeds using only `hn` commands.
2. **Given** invalid arguments, **when** a command reports usage or recovery
   guidance, **then** every suggested command begins with `hn`.
3. **Given** the source tree, **when** it is built according to the README,
   **then** the generated executable is named `hn`, and no `nh` alias or
   compatibility executable is produced.

---

### User Story 2 - Fresh protocol and storage namespace (Priority: P1)

As a Hubnot participant, I want all newly created collaboration state isolated
under the `hn` namespace so legacy facts cannot silently enter new trust
decisions.

**Why this priority**: Reusing old refs or private state would turn a cosmetic
rename into an ambiguous protocol migration.

**Independent Test**: Exercise signing, governance, CI, memory, selective
replication, shallow recovery, and identity operations in fresh repositories,
then inspect Git refs and private paths directly.

**Acceptance Scenarios**:

1. **Given** a fresh repository, **when** Hubnot creates collaboration and
   memory facts, **then** their wire versions are `hn/0`, `hn-memory/0`,
   `hn.pipeline/0`, and `hn.policy/0`, and their refs are below `refs/hn/*`.
2. **Given** local identity, runner, memory-index, and replication state,
   **when** Hubnot persists it, **then** it is below `.git/hn/`, temporary runner
   paths use `hn`, and supported environment variables begin with `HN_`.
3. **Given** a repository containing only valid legacy `refs/nh/*` facts and
   `.git/nh/` state, **when** the new `hn` program reads the repository, **then**
   it does not discover, import, migrate, or modify that state.
4. **Given** a remote advertising both namespaces, **when** `hn sync` runs,
   **then** it fetches and publishes only `refs/hn/*`.

---

### User Story 3 - Honest active documentation (Priority: P2)

As a contributor, I want current documentation and project governance to
describe the actual `hn` implementation while retaining clearly marked records
of what happened before the cutover.

**Why this priority**: The protocol is only useful when users can reproduce it
from its written contract.

**Independent Test**: Compare the README, current protocol documents, active
charter, examples, and tracked configuration against executable behavior, then
classify every surviving old-token occurrence as historical evidence.

**Acceptance Scenarios**:

1. **Given** an active guide or protocol contract, **when** it names commands,
   refs, private directories, versions, runner labels, or environment variables,
   **then** it uses the `hn` spelling implemented by the program.
2. **Given** a historical mission record, signed journal, self-hosting report,
   or the frozen `.nh/` transition inputs, **when** the cutover is applied,
   **then** its old spellings remain unchanged and its historical purpose is
   documented.
3. **Given** the project charter, **when** it describes the native CLI and its
   project directive identifiers, **then** the active forms use `hn` and
   `HN-###`.

---

### User Story 4 - Governed and recoverable transition (Priority: P2)

As the project maintainer, I want the exact cutover commit accepted through the
existing public governance and a fresh `hn` trust root established afterward,
so the break is intentional and independently auditable.

**Why this priority**: The old protocol must be able to attest to its final
transition without becoming a compatibility dependency of the new protocol.

**Independent Test**: Use the pre-cutover executable and frozen `.nh/` pipeline
to propose, run, review, decide, merge, and publish the exact candidate; then
initialize fresh maintainer and reviewer identities and verify new `hn` facts on
the public remote.

**Acceptance Scenarios**:

1. **Given** the exact final candidate, **when** the old governance loop reviews
   it, **then** its signed acceptance and merge facts remain under
   `refs/nh/*`, and the candidate contains no code that reads those refs.
2. **Given** fresh maintainer and reviewer `hn` identities, **when** active
   `.hn/policy.json` is evaluated, **then** it names only the new actors and uses
   only new protocol identifiers.
3. **Given** the merged public tree, **when** a fresh clone performs the documented
   smoke test, **then** it can discover and validate `refs/hn/*` using ordinary
   Git transport without Docker or a Hubnot service.

### Edge Cases

- A checkout contains both `.git/nh/` and `.git/hn/`; only the latter is active.
- A remote advertises colliding actor or proposal identifiers in both ref
  namespaces; selection and validation remain confined to `refs/hn/*`.
- A user has an unrelated file or prose word containing the letters `nh`; the
  occurrence map prevents blind substring replacement.
- Historical documents contain commands that were correct when recorded; they
  remain unchanged instead of being rewritten as if the new CLI produced them.
- The legacy `.nh/` policy and pipeline are present after cutover; `hn` ignores
  them and uses `.hn/` exclusively.
- A legacy environment variable is set; it has no effect unless a non-Hubnot
  dependency independently interprets it.

## Requirements

### Functional Requirements

| ID | Title | Requirement | Priority | Status |
|----|-------|-------------|----------|--------|
| FR-001 | CLI cutover | The shipped command, every usage string, recovery instruction, example, and generated binary MUST use `hn`; the project MUST NOT ship an `nh` alias. | High | Open |
| FR-002 | Ref cutover | Active local, remote-tracking, proposal, actor, memory, quarantine, and replication refs MUST use `refs/hn/*`; runtime code MUST NOT enumerate, fetch, push, validate, or update `refs/nh/*`. | High | Open |
| FR-003 | Private-path cutover | Active identity, keyring, memory-index, replication, runner, and temporary state MUST use `.git/hn/` or an `hn`-named temporary path; runtime code MUST NOT probe or migrate `.git/nh/`. | High | Open |
| FR-004 | Repository-config cutover | Active policy, pipeline, and action paths MUST use `.hn/`; `hn` MUST ignore `.nh/`. | High | Open |
| FR-005 | Wire cutover | Collaboration, memory, pipeline, and policy versions MUST be exactly `hn/0`, `hn-memory/0`, `hn.pipeline/0`, and `hn.policy/0`; old versions MUST fail closed as unsupported. | High | Open |
| FR-006 | Environment cutover | Hubnot-owned environment variables and runner labels MUST use `HN_*` and `hn/<version>`; no `NH_*` compatibility reads or duplicate exports are allowed. | High | Open |
| FR-007 | Test cutover | Unit, integration, race, acceptance, and real-Git tests MUST assert only the new active namespace except for explicit isolation tests proving old state is ignored. | High | Open |
| FR-008 | Documentation cutover | README and current protocol, governance, CI, identity, memory, safety, replication, and host-compatibility guidance MUST match the new executable contract. | High | Open |
| FR-009 | Historical preservation | Existing signed Git facts, immutable journals, completed mission artifacts, historical self-hosting evidence, and the frozen `.nh/` transition inputs MUST remain unchanged and be classified as exceptions. | High | Open |
| FR-010 | Fresh trust roots | Active `.hn/policy.json` MUST name newly generated `hn` maintainer/reviewer identities, not actors inherited from legacy state. | High | Open |
| FR-011 | Governed cutover | The exact candidate MUST be landed using the final old governance loop before new public `hn` facts are established. | High | Open |
| FR-012 | Active charter | The active project charter MUST describe `hn` and use `HN-###` directive identifiers without rewriting historical mission references. | Medium | Open |
| FR-013 | QA findings | The cutover MUST retain the resolved behavior for QA findings F-1 through F-5 and MUST NOT regress the validated signing, governance, CI, replication, or keyring properties. | High | Open |

### Non-Functional Requirements

| ID | Title | Requirement | Category | Priority | Status |
|----|-------|-------------|----------|----------|--------|
| NFR-001 | Race safety | `go test -race -count=1 ./...` MUST pass with zero failing packages. | Reliability | High | Open |
| NFR-002 | Static quality | `go build ./...`, `go vet ./...`, `gofmt -l .`, and `git diff --check` MUST be clean. | Quality | High | Open |
| NFR-003 | Namespace completeness | A scoped active-tree scan MUST report zero unclassified legacy namespace occurrences. | Correctness | High | Open |
| NFR-004 | Dependency boundary | `go.mod` MUST retain zero third-party dependencies and the runtime MUST add no Docker or service requirement. | Portability | High | Open |
| NFR-005 | Security invariants | Exact-byte Ed25519 verification, append CAS, quarantine-before-promote, fail-closed policy evaluation, keyring permissions, and sandbox isolation MUST retain their existing test coverage and behavior. | Security | High | Open |
| NFR-006 | Public smoke | A fresh public clone MUST complete the representative `hn` smoke flow and expose only expected `refs/hn/*` facts. | Interoperability | High | Open |

### Constraints

| ID | Title | Constraint | Category | Priority | Status |
|----|-------|------------|----------|----------|--------|
| C-001 | No compatibility | Do not add aliases, deprecation shims, migration code, fallback reads, dual writes, or legacy protocol acceptance. | Product | High | Open |
| C-002 | No history rewrite | Do not rewrite Git history, signed event bytes, completed mission records, immutable journals, or historical reports. | Integrity | High | Open |
| C-003 | Preserve product identity | The product and Go module remain `Hubnot` and `hubnot`; only the active short namespace changes from `nh` to `hn`. | Naming | High | Open |
| C-004 | Ordinary Git | Distribution continues through standard Git refs and transports, with no provider-specific or container dependency. | Architecture | High | Open |
| C-005 | Exact scope | Replacement MUST be token- and meaning-aware; unrelated substrings and external names are outside scope. | Quality | High | Open |

### Key Entities

- **Active namespace**: The `hn` CLI token, Git ref root, private storage root,
  wire identifiers, environment prefix, runner label, and repository config root.
- **Legacy evidence**: Old `nh` facts and documents retained solely to preserve
  provenance; they are never active inputs to the new implementation.
- **Cutover identity**: A fresh Ed25519 actor whose private state exists only
  below `.git/hn/` and whose fingerprint appears in active `.hn/policy.json`.
- **Transition candidate**: The exact commit accepted by the final legacy
  governance loop and then published as the basis for new `hn` facts.

## Success Criteria

### Measurable Outcomes

- **SC-001**: All active CLI, ref, private-path, config-path, wire, environment,
  runner, test, and current-documentation occurrences use `hn`, with zero
  unclassified exceptions.
- **SC-002**: A repository seeded only with legacy state produces no discovered
  events or migrated files when inspected by the new program.
- **SC-003**: The full build, vet, formatting, race, acceptance, and live smoke
  suite passes with zero failures.
- **SC-004**: The public remote contains the final legacy merge evidence and
  independently rooted new `refs/hn/*` facts, and ordinary Git can enumerate
  both without namespace overlap.
- **SC-005**: README and current protocol documents can be followed verbatim
  with an `hn` binary, no Docker daemon, no Hubnot server, and no hidden
  compatibility step.
