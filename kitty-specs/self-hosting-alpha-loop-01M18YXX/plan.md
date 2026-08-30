# Implementation Plan: Operational Self-Hosting Alpha

**Branch**: `feat/self-hosting-alpha-loop` | **Date**: 2026-08-30 | **Spec**: [spec.md](spec.md)  
**Input**: Expanded operational-alpha specification in `kitty-specs/self-hosting-alpha-loop-01M18YXX/spec.md`

## Summary

Extend the native `nh` CLI across four existing boundaries rather than adding
a service: policy inspection over the canonical policy validator; mutually
signed continuity among distinct single-writer actors backed by a private local
keyring; per-remote selected replication through an isolated Git quarantine
and atomic promotion; and exact-object shallow dependency recovery through the
same selected fetch path. Prove the combined behavior with a subprocess-level
multi-repository acceptance scenario and a staged public self-hosting workflow
in which an old policy governs its amendment and a later candidate requires a
second actor's evidence.

## Technical Context

**Language/Version**: Go 1.26  
**Primary Dependencies**: Go standard library; Git 2.x command-line plumbing as the storage and transport engine; Bubblewrap remains an optional Linux runtime dependency for the default live CI backend; no new Go module dependency  
**Storage**: Git objects and `refs/nh/*` for public signed state; owner-only files under `.git/nh/` for the identity keyring, active actor, replication selections, and transient transaction state; generated bare Git repositories for fetch quarantine  
**Testing**: Go `testing` package with test-first unit tests, real temporary Git repositories for integration, helper-process CLI black-box scenarios, `go test ./...`, `go test -race ./...`, `go vet ./...`, and a three-run operational acceptance check  
**Target Platform**: Git-supported developer hosts for identity, policy, replication, and shallow behavior; Linux for the live default Bubblewrap runner proof; unsafe host execution remains explicit and is used only with controlled synthetic test input  
**Project Type**: Single native command-line executable  
**Performance Goals**: Promotion work stays within configured positive event/object/byte budgets; the full offline operational scenario completes within 120 seconds per run; selected sync requests no unselected ref graph  
**Constraints**: No mandatory service, database, Docker daemon, hosting-provider API, shared private actor key, silent policy authority transfer, silent unshallow, or mutation of existing event payload bytes and IDs  
**Scale/Scope**: Version-0 operational alpha for repositories with multiple actor histories and candidate refs; configurable budgets define accepted scale rather than an unbounded global fetch; hard pre-download transport quotas are explicitly deferred

## Branch Contract

- Current branch at plan start: `feat/self-hosting-alpha-loop`.
- Planning/base branch: `feat/self-hosting-alpha-loop`.
- Spec Kitty merge target: `feat/self-hosting-alpha-loop`.
- `branch_matches_target` is `true`.
- After Spec Kitty assembles and accepts the feature candidate, Nichthub itself
  governs the staged public landing into `main`.

## Charter Check

*GATE: Passed before Phase 0 research; re-evaluated after Phase 1 design.*

| Charter rule | Design response | Gate |
| --- | --- | --- |
| NH-001 Repository-native protocol | New public facts are signed events in Git; private operational state remains repository-local; no service or provider API is added. | Pass |
| NH-002 Exact evidence integrity | Identity claims reference exact actor keys/events; policy amendments remain base-bound candidates; selected facts are never reinterpreted across IDs. | Pass |
| NH-003 Hostile input and least authority | Selected refs fetch into a separate quarantine, face positive budgets and validation before atomic promotion, and grant no policy authority. | Pass |
| NH-004 Immutable recovery | Rotation, identity conflicts, missing dependencies, and candidate conflicts add or preserve facts; no published event is rewritten. | Pass |
| NH-005 Protocol and docs together | Event contracts, protocol docs, threat model, compatibility record, operator guide, and executable examples ship with behavior. | Pass |
| DIRECTIVE_001 Deep boundaries | Keyring, continuity projection, policy inspection, replication transaction, shallow resolver, and existing governance have explicit ownership and interfaces. | Pass |
| DIRECTIVE_003 Decisions recorded | Material design choices and rejected alternatives are recorded in `research.md` and contracts. | Pass |
| DIRECTIVE_010 Spec fidelity | Every concern and contract maps to stable FR IDs; deviations require a recorded decision before implementation. | Pass |
| DIRECTIVE_024 Locality | Existing files remain authoritative; new deep modules are added only where current mixed command/storage concerns would otherwise spread. | Pass |
| DIRECTIVE_030 Quality gates | Unit, integration, race, vet, compatibility, black-box, and live evidence gates are explicit. | Pass |
| DIRECTIVE_033 Targeted staging | Each future work package names exact deliverables and uses Spec Kitty safe commits. | Pass |
| DIRECTIVE_034 Test first | Every behavior starts from a failing public-seam or focused domain test before production code. | Pass |
| DIRECTIVE_036 Black-box integration | The complete operational flow runs through CLI subprocesses against real Git repositories. | Pass |
| DIRECTIVE_037 Living docs | Protocol, security, host compatibility, README, quickstart, and verification record evolve in the same mission. | Pass |
| DIRECTIVE_041 Tests as scaffold | Tests assert observable contracts and signed identities, not incidental internal line/layout details. | Pass |
| DIRECTIVE_044 Canonical unification | Policy commands reuse `validatePolicy`; shallow recovery reuses the replication transaction; identity callers use one keyring interface. | Pass |

No charter violation requires a complexity exception.

## Architecture and Data Flow

### Identity continuity

1. Existing commands obtain a signer through the keyring's active-identity
   interface; legacy repositories migrate their exact key bytes once.
2. An actor publishes `identity.authorize` with relationship, target actor,
   and target public key.
3. The target actor publishes `identity.accept` referencing that exact event.
4. The relationship projector validates both signatures and derives pending,
   accepted, retired, or ambiguous state without touching policy roles.
5. A local planned rotation switches the active pointer only after both actor
   refs and transaction state are durable.

### Policy amendment

1. Policy inspection loads the canonical policy bytes from a commit or an
   explicit draft file and runs the existing validator.
2. A deterministic comparison reports digests, roles, thresholds, pipeline
   changes, and lockout risks.
3. The Git policy diff is committed and proposed through the ordinary candidate
   flow.
4. Evaluation continues to load only the signed base policy; the head policy
   becomes relevant only as the base of later candidates.

### Selected replication

1. Local configuration resolves a remote and full actor/candidate selections.
2. Exact refspecs fetch into a generated bare quarantine repository.
3. The transaction measures event, attachment, object, and reachable-byte
   budgets before domain projection.
4. Actor chains validate independently; cross-event relationships validate
   over accepted facts plus the selected transaction set.
5. Missing dependencies are classified separately from malformed facts.
6. Promotable refs and objects are copied into the main repository and accepted
   remote-tracking refs update atomically.
7. Invalid and over-budget selections never become accepted projection roots.

### Shallow recovery

1. Integrity-sensitive operations resolve every exact required event, commit,
   tree/blob, ref, and ancestor.
2. Missing-object errors are classified against Git's shallow boundary.
3. Diagnostics name the exact missing ID and selected ref that can supply it.
4. Explicit recovery routes that ref through the same quarantine, budgets,
   validation, and atomic promotion; it never issues a global unshallow.
5. The original operation restarts from accepted state rather than reusing a
   partial trust decision.

## Transaction and Failure Boundaries

| Boundary | Durable-before-next invariant | Recovery |
| --- | --- | --- |
| Legacy identity migration | New keyring record recomputes the same actor and private/public pair before active pointer changes. | Keep legacy file until atomic pointer and record writes succeed; retry idempotently. |
| Planned rotation | Predecessor authorization and successor acceptance are durable before switching the active signer. | Persist in-progress IDs; retry missing event/ref publication; old signer remains active until completion. |
| Quarantine fetch | Accepted refs are unchanged while objects and refs are measured and validated separately. | Reject/discard transaction; report each selection result. |
| Ref promotion | All promotable objects exist in the main object database before one atomic accepted-ref transaction. | A failed transaction leaves old accepted refs intact and can be retried. |
| Shallow recovery | No trust-sensitive command advances until all exact dependencies validate after promotion. | Report and retry explicit selected recovery; never assume absent history. |
| Policy amendment | Candidate evidence signs the base digest; proposed head policy has no authority. | Reject readiness/decision with exact base policy and missing evidence. |
| Git merge/publication | Local merge fact and primary branch publication remain separate observable actions. | Preserve merge fact/commit IDs and retry ordinary branch or collaboration-ref publication explicitly. |

## Project Structure

### Documentation (this mission)

```text
kitty-specs/self-hosting-alpha-loop-01M18YXX/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── identity-continuity-v0.md
│   ├── policy-amendment-cli-v0.md
│   └── selected-replication-v0.md
├── decisions/
└── tasks/                       # created by the tasks phase
```

### Source Code (repository root)

```text
main.go                          # top-level command routing and usage
event.go                         # canonical signed event payload and validation
store.go                         # event object storage and accepted projection loading
identity.go                      # active-identity facade and legacy migration entry
identity_keyring.go              # private keyring records, pointer, atomic local state
identity_continuity.go           # public relationship commands and deterministic projection
policy.go                        # canonical policy parser/validator/evaluator
policy_commands.go               # inspect/check/diff CLI over canonical policy
commands.go                      # issue/log and sync entry routing retained during extraction
replication.go                   # selection configuration and transaction orchestration
quarantine.go                    # isolated Git fetch, measurement, validation, promotion
shallow.go                       # exact dependency classification and recovery routing
proposal.go                      # candidate creation/inspection/revision
ci.go / runner.go / backend.go   # existing exact CI execution and attestations
governance.go                    # existing decision and merge gates

identity_keyring_test.go
identity_continuity_test.go
policy_commands_test.go
replication_test.go
shallow_test.go
operational_acceptance_test.go
existing *_test.go               # compatibility and regression suite

docs/protocol-v0.md
docs/governance-v0.md
docs/ci-v0.md
docs/host-compatibility.md
docs/identity-v0.md
docs/replication-v0.md
docs/self-hosting-alpha.md
README.md
```

**Structure Decision**: Keep the flat single-package Go executable while
introducing deep files around cohesive stateful boundaries. Callers continue
to use package-private interfaces, avoiding speculative internal packages and
dependency cycles. `policy.go` remains the only policy validator;
`replication.go` is the only accepted-remote mutation path;
`identity_keyring.go` is the only private-key persistence path.

## CLI Surface Plan

```text
nh identity show
nh identity list
nh identity public
nh identity authorize --relationship device|successor --actor ACTOR --public-key KEY
nh identity accept AUTHORIZATION
nh identity rotate [--name NAME]

nh policy show [REV]
nh policy check --base REV (--head REV | --file PATH)

nh replication select [REMOTE] [--actor ACTOR]... [--proposal ID]...
  [--all] [--max-events N] [--max-objects N]
  [--max-object-bytes N] [--max-attachment-bytes N]
  [--max-total-bytes N]
nh replication show [REMOTE]
nh sync [REMOTE] [--recover-shallow]
```

- Existing `nh sync [REMOTE]` remains valid. If a saved selection exists, it
  is authoritative. Repositories without one retain explicitly documented
  bounded compatibility-all behavior until configured.
- All actor and event arguments accepted by selection/identity commands require
  full IDs; short prefixes are inspection convenience, not trust configuration.
- Machine-stable JSON output is not introduced across the whole CLI in this
  mission; contracts define exact human-readable success/error facts and tests
  use stable identifiers rather than column spacing.

## Test-First Verification Matrix

| Behavior | First failing seam | Required proof |
| --- | --- | --- |
| Legacy key migration | Existing repository loads signer after legacy file is converted | Same actor/key bytes, permissions, signature verification, idempotent retry |
| Device/successor claims | Signed event validation and relationship projection | Both signatures required; cycles/competing successors deterministic; policy unchanged |
| Rotation transaction | CLI helper process interrupted at each durable boundary | Old signer remains active before completion; retry completes once; no shared key |
| Policy check | CLI over base/head and base/draft fixtures | Exact digests/diff, structural failures, no self-authorization |
| Selected sync | Real remote with selected/unselected actor and proposal refs | Only selected valid refs promote; atomic old-ref preservation on failure |
| Budget enforcement | One-below/exact/one-above fixtures for every budget | Exact boundary behavior and rejected refs absent from accepted projection |
| Invalid isolation | Valid and hostile selected histories in one run | Valid selection usable; hostile selection quarantined with full ID |
| Missing dependency | Review/result/decision referencing an unselected chain | Dependency classification and exact next selection, not corruption |
| Shallow operation | Depth-limited clone missing base/policy/actor predecessor | Trust action blocked; explicit selected recovery; no global unshallow |
| Existing protocol compatibility | Golden public seven-event baseline plus current tests | Unchanged IDs and projections for all existing kinds |
| Operational acceptance | CLI subprocesses, two actor clones, disposable bare remote | Amendment under old policy; later role-distinct readiness and merge; fresh shallow reconstruction |
| Live public proof | Exact public refs and event/commit IDs | Default sandbox result, separate branch/ref publication, identity-free verification |

## Implementation Concern Map

### IC-01 — Canonical Identity Keyring

- **Purpose**: Preserve existing actors while making private signer selection,
  migration, and multi-key local transaction state explicit and recoverable.
- **Relevant requirements**: FR-005, FR-006, FR-018; NFR-002, NFR-006, NFR-009.
- **Affected surfaces**: `identity.go`, new `identity_keyring.go`, identity tests,
  command callers that load the active signer.
- **Sequencing/depends-on**: none.
- **Risks**: Secret leakage, permission regression, non-atomic migration, or
  changing an existing actor fingerprint.

### IC-02 — Signed Identity Continuity

- **Purpose**: Define, validate, store, inspect, and project mutually signed
  device and planned-successor relationships without granting policy authority.
- **Relevant requirements**: FR-005–FR-008; NFR-001, NFR-006; C-002–C-005.
- **Affected surfaces**: `event.go`, `store.go`, new
  `identity_continuity.go`, `main.go`, `docs/protocol-v0.md`, new
  `docs/identity-v0.md`.
- **Sequencing/depends-on**: IC-01.
- **Risks**: Ambiguous successor projection, two-event crash boundary, cycles,
  or accidental authority inheritance.

### IC-03 — Policy Amendment Inspection

- **Purpose**: Expose the canonical policy bytes, digest, validity, and
  deterministic base/head change set before an ordinary proposal is opened.
- **Relevant requirements**: FR-001–FR-004, FR-008, FR-016.
- **Affected surfaces**: `policy.go`, new `policy_commands.go`, `main.go`,
  policy tests, `docs/governance-v0.md`.
- **Sequencing/depends-on**: none; consumes actor IDs from IC-02 only in the
  live proof, not in implementation.
- **Risks**: Duplicated validation logic or language suggesting the head policy
  governs itself.

### IC-04 — Selected Quarantined Replication

- **Purpose**: Replace direct wildcard import with explicit local selection,
  isolated fetch, bounded validation, failure isolation, and atomic accepted-ref
  promotion.
- **Relevant requirements**: FR-009–FR-013, FR-017–FR-019; NFR-001–NFR-004,
  NFR-006–NFR-009.
- **Affected surfaces**: `commands.go`, `store.go`, `git.go`, new
  `replication.go`, new `quarantine.go`, `main.go`, integration tests, new
  `docs/replication-v0.md`.
- **Sequencing/depends-on**: IC-02 supplies new relationship validation; its
  storage and budget primitives can be built independently.
- **Risks**: Accepted refs pointing to unavailable objects, invalid actor
  poisoning, non-atomic partial promotion, temporary pack exhaustion, or
  breaking legacy sync.

### IC-05 — Shallow Dependency Recovery

- **Purpose**: Distinguish missing selected history from invalid data and route
  explicit exact-ref recovery through the replication transaction.
- **Relevant requirements**: FR-013–FR-015, FR-017–FR-019; NFR-005, NFR-009.
- **Affected surfaces**: `git.go`, `store.go`, `proposal.go`, `ci.go`,
  `governance.go`, new `shallow.go`, shallow integration tests.
- **Sequencing/depends-on**: IC-04.
- **Risks**: Silent broad fetch, accidental trust expansion, or operation
  proceeding after partial dependency resolution.

### IC-06 — Operational Black-Box Acceptance

- **Purpose**: Prove the externally visible CLI workflow across policy,
  identity, selection, shallow recovery, evidence, and merge using real Git
  repositories and isolated processes.
- **Relevant requirements**: FR-003, FR-005–FR-019; all NFRs.
- **Affected surfaces**: new `operational_acceptance_test.go`, focused fixtures
  generated inside test temporary directories.
- **Sequencing/depends-on**: IC-01–IC-05.
- **Risks**: Recursive test execution, global working-directory interference,
  brittle output matching, or accidental network/public mutation.

### IC-07 — Living Protocol and Public Proof

- **Purpose**: Keep user guidance, wire contracts, threat boundaries,
  compatibility observations, and inaugural public IDs aligned with behavior.
- **Relevant requirements**: FR-016, FR-017, FR-020; NFR-008, NFR-010.
- **Affected surfaces**: `README.md`, all `docs/*.md` listed above, mission
  contracts, quickstart, verification record.
- **Sequencing/depends-on**: Contract drafts precede IC-02–IC-05; final IDs
  follow IC-06 and the staged public landing.
- **Risks**: Overstating human independence, pre-download quota guarantees, or
  global deletion/recovery properties.

## Complexity Tracking

No charter violation is accepted. The additional modules are justified deep
boundaries around private key transactions and hostile remote imports; keeping
either inside the existing broad command file would spread mutable state,
failure recovery, and security invariants across unrelated CLI handlers.

## Post-Design Charter Re-check

- Git remains the only public storage and transport.
- New identity events are exact signed immutable facts and existing event bytes
  remain unchanged.
- Authority stays in the exact base policy; continuity and selection do not
  imply trust.
- Remote data is selected, measured, verified, and atomically promoted before
  accepted projection.
- Unsafe execution remains explicit and no Docker dependency is introduced.
- Tests and documentation cover every observable behavior and known residual
  risk in the same mission.

**Result**: Pass. The plan is ready for work-package decomposition.
