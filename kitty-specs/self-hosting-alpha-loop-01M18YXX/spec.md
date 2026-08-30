# Mission Specification: Self-Hosting Alpha Loop

**Mission Branch**: `feat/self-hosting-alpha-loop`  
**Created**: 2026-08-30  
**Status**: Draft  
**Input**: Prove Nichthub can govern one real change to Nichthub from signed issue through public merge.

## Intent

A Nichthub maintainer needs to operate the project through the collaboration
protocol it is building. The mission succeeds when one real, reviewable
Nichthub change travels through a public Git remote as an issue, immutable
proposal, verified CI result, review, acceptance decision, and merge fact, and
when a fresh participant can reconstruct that history without a Nichthub
server or GitHub collaboration feature.

The loop is a product acceptance exercise, not a staged demonstration. Any
blocker discovered in the existing commands must be reproduced, corrected,
and covered before the live loop continues. Signed history remains immutable;
recovery creates new facts rather than rewriting failed attempts.

## User Scenarios & Testing

### User Story 1 - Govern a Real Nichthub Change (Priority: P1)

As a Nichthub maintainer, I want to take a real repository change from issue
through merge using Nichthub so that the project demonstrates its own central
collaboration claim in ordinary use.

**Why this priority**: A complete real loop is the smallest credible proof that
the separate protocol features compose into a usable workflow.

**Independent Test**: Starting with the public repository and the current
maintainer identity, complete the issue, proposal, test request/result, review,
acceptance, merge, and publication flow. Verify every referenced identifier and
the final repository state.

**Acceptance Scenarios**:

1. **Given** a clean primary branch, a real candidate branch, and the committed project policy, **when** the maintainer opens an issue and proposes the candidate, **then** both signed facts and the exact candidate code are available through the public remote.
2. **Given** that exact proposal, **when** its configured test pipeline passes and the required review and acceptance are published, **then** the proposal is reported ready and can be merged under the policy from its signed base.
3. **Given** a successful governed merge, **when** collaboration refs and the primary branch are published, **then** the public remote contains the merged code and a verifiable merge fact bound to the accepted candidate.

---

### User Story 2 - Reconstruct the Public History (Priority: P2)

As an independent participant, I want a fresh clone to reconstruct the
collaboration history and proposed code so that the proof does not depend on
the originating worktree or hidden server state.

**Why this priority**: Repository-native collaboration is only distributed if
another clone can independently obtain and verify the same facts.

**Independent Test**: Create a fresh clone with no copied Nichthub state,
synchronize from the public remote, and inspect the issue, proposal, evidence,
lineage, and merge result using public identifiers.

**Acceptance Scenarios**:

1. **Given** a fresh clone of the public repository, **when** a new participant synchronizes Nichthub refs, **then** the participant sees the same issue, proposal, review, run result, decision, and merge identities as the originating clone.
2. **Given** the reconstructed proposal, **when** the participant inspects its code and evidence, **then** every signed reference resolves to the exact public object it claims and no local-only state is required.

---

### User Story 3 - Recover Without Rewriting History (Priority: P3)

As a maintainer, I want a failed execution, stale candidate, or merge conflict
to produce actionable recovery information so that self-hosting does not
require deleting or editing signed history.

**Why this priority**: The first real use is likely to expose operational gaps;
failure must strengthen the protocol rather than tempt an ad hoc bypass.

**Independent Test**: Exercise at least one safe failure path locally and
verify that the original facts remain visible, no unsupported target mutation
occurs, and the command identifies the exact next action or candidate.

**Acceptance Scenarios**:

1. **Given** a command cannot safely advance the candidate, **when** it fails, **then** it leaves Git and signed collaboration state consistent and identifies the blocking fact by its full identifier.
2. **Given** the candidate conflicts with its target, **when** the maintainer resolves the conflict, **then** recovery publishes an immutable revision whose evidence must be earned again.

### Edge Cases

- The public remote accepts actor refs but rejects or rewrites proposal refs.
- Synchronization succeeds for collaboration refs while publication of the
  primary Git branch fails, or vice versa.
- The configured pipeline passes locally but cannot run in the default
  isolated executor.
- A candidate becomes stale because the primary branch advances before merge.
- The same maintainer acts as author, reviewer, runner, and decision maker
  under a policy that explicitly permits those roles.
- A fresh clone has no private identity and must still verify and inspect all
  published facts.
- A live-loop blocker requires a behavior change after a proposal has already
  been published; the repair must use a new candidate or revision.

## Domain Language

- **Self-hosting alpha loop**: One real Nichthub project change governed from
  signed issue through published merge using Nichthub collaboration facts.
- **Candidate**: One immutable proposal or proposal revision bound to exact
  base and head commits. Avoid using “PR” as a synonym.
- **Public remote**: An ordinary publicly readable Git remote. Its hosting
  provider is transport, not a collaboration authority.
- **Fresh participant**: A clone that did not inherit the originating clone's
  private `.git/nh` state.
- **Live evidence**: Signed collaboration events and reachable Git objects
  published during the real public workflow, not hand-authored screenshots or
  prose claims.

## Requirements

### Functional Requirements

| ID | Title | User Story | Priority | Status |
|----|-------|------------|----------|--------|
| FR-001 | Publish the governing issue | As a maintainer, I can publish a signed issue describing the real candidate change and retrieve the same issue from the public remote. | High | Open |
| FR-002 | Publish the exact candidate | As a maintainer, I can publish an immutable proposal whose signed base and head identify the candidate and whose code ref makes that head fetchable. | High | Open |
| FR-003 | Produce bound CI evidence | As a maintainer, I can request and execute the candidate's configured test pipeline and publish a signed result bound to the exact proposal, commit, and pipeline definition. | High | Open |
| FR-004 | Review and decide under policy | As a maintainer, I can publish the required review and acceptance decision only when the exact candidate satisfies the policy from its signed base. | High | Open |
| FR-005 | Merge and publish the accepted candidate | As a maintainer, I can merge the accepted candidate, publish its signed merge fact, and publish the resulting primary branch without using a GitHub pull request or collaboration API. | High | Open |
| FR-006 | Reconstruct from a fresh clone | As an independent participant, I can synchronize the public collaboration refs and verify the issue, proposal code, CI, review, decision, and merge chain without copied private state. | High | Open |
| FR-007 | Preserve recoverable failures | As a maintainer, I receive exact blocking identifiers and recovery guidance when the loop cannot advance, while existing signed facts and the previously clean target remain intact. | High | Open |
| FR-008 | Make the scenario repeatable | As a contributor, I can run an automated offline acceptance scenario over an ordinary temporary Git remote that exercises the same externally observable workflow without changing the public repository. | Medium | Open |
| FR-009 | Record verifiable operating evidence | As a contributor, I can follow project documentation for the self-hosting workflow and inspect a durable record of the public event and commit identifiers produced by the inaugural loop. | Medium | Open |

### Non-Functional Requirements

| ID | Title | Requirement | Category | Priority | Status |
|----|-------|-------------|----------|----------|--------|
| NFR-001 | Completion time | Once a candidate branch is ready and required local tools are installed, the happy-path live loop completes within 15 minutes of operator time, excluding network outages. | Usability | Medium | Open |
| NFR-002 | Fresh-clone convergence | After public synchronization completes, 100% of live evidence identifiers recorded by the originating clone are present and verify identically in the fresh clone. | Reliability | High | Open |
| NFR-003 | Secret isolation | Zero private identity keys, temporary credentials, access tokens, or host-private paths appear in tracked files, signed logs, or published evidence. | Security | High | Open |
| NFR-004 | Repeatable acceptance | The offline acceptance scenario passes in three consecutive runs and completes within 90 seconds per run on the development host. | Reliability | High | Open |
| NFR-005 | Bounded failure | Every tested failure path returns a non-zero result within its configured timeout and leaves the previously clean target worktree clean with no unrecorded commit. | Safety | High | Open |
| NFR-006 | Provider independence | All collaboration operations in the live loop use documented Git and Nichthub commands; zero hosting-provider collaboration API calls are required. | Portability | High | Open |

### Constraints

| ID | Title | Constraint | Category | Priority | Status |
|----|-------|------------|----------|----------|--------|
| C-001 | Repository-native boundary | Collaboration state remains signed, content-addressed repository data transported through Git; no mandatory Nichthub service or service database may be introduced. | Product | High | Open |
| C-002 | Existing identity model | The mission operates within the current single-writer-per-identity model; multi-device append, key rotation, and recovery remain outside scope. | Protocol | High | Open |
| C-003 | Immutable evidence | Published events, candidates, logs, and decisions are never edited or replaced; corrections and conflict recovery create new signed facts. | Governance | High | Open |
| C-004 | Existing execution safety | The default isolated runner remains the live execution path, with unsafe host execution allowed only through its existing explicit opt-in. No Docker dependency is added. | Security | High | Open |
| C-005 | Compatibility before expansion | A blocker may justify a protocol or command change only when reproduced by the self-hosting scenario; unrelated roadmap features are excluded from this mission. | Scope | High | Open |
| C-006 | Public proof, local secrets | The remote repository and collaboration facts are public while every actor's private key remains local and untracked. | Security | High | Open |

### Key Entities

- **Self-Hosting Run**: The bounded inaugural workflow, identified by its
  public issue, candidate, evidence, decision, merge event, and resulting Git
  commit identifiers.
- **Candidate**: The immutable proposed project state with exact signed base
  and head commits plus a reachable code ref.
- **Evidence Set**: Reviews and CI results that satisfy the candidate's base
  policy and are signed into its acceptance decision.
- **Verification Record**: A durable, non-secret mapping of the inaugural
  run's public identifiers and the observations made from a fresh clone.

## Scope Boundaries

### In Scope

- One real public self-hosting cycle for this repository.
- A repeatable black-box acceptance scenario using an ordinary disposable Git
  remote.
- Minimal command, validation, documentation, or protocol corrections that the
  real loop proves necessary.
- Fresh-clone verification and a durable public evidence record.

### Out of Scope

- Multiple writers sharing one identity, key rotation, or identity recovery.
- General notification, search, web UI, moderation, or project discovery.
- Portable runner backends, secrets, controlled network access, or stronger
  resource quotas.
- Hosting-provider automation, pull-request mirroring, or a Nichthub server.
- Unrelated protocol redesign or compatibility stabilization.

## Assumptions and Dependencies

- The public remote continues to preserve the documented custom ref
  namespaces and ordinary Git objects.
- The current maintainer identity is the actor named in the committed policy;
  author review is therefore allowed for this inaugural loop.
- The configured `test` pipeline and default isolated runner are available on
  the development host.
- Publishing the primary Git branch remains an explicit ordinary Git action;
  publishing collaboration facts remains an explicit Nichthub synchronization
  action.
- If the live loop exposes no product blocker, the mission may ship acceptance
  coverage, documentation, and evidence without inventing new protocol
  behavior.

## Success Criteria

### Measurable Outcomes

- **SC-001**: One real change reaches the public primary branch with a complete,
  cryptographically verifiable chain containing one issue, candidate, passing
  CI result, qualifying review, acceptance decision, and merge fact.
- **SC-002**: A fresh clone reconstructs and verifies 100% of the identifiers
  in the inaugural verification record without access to the originating
  clone's private state.
- **SC-003**: The repeatable offline acceptance scenario passes three
  consecutive times within 90 seconds per run.
- **SC-004**: The live collaboration loop uses zero GitHub collaboration API
  calls, zero Nichthub services, and zero Docker dependencies.
- **SC-005**: Repository documentation enables a contributor familiar with Git
  to reproduce the happy-path workflow using no more than 12 explicit operator
  commands after identity initialization and candidate preparation.
- **SC-006**: All existing automated tests and static checks pass after any
  blocker corrections, with no private material present in the tracked diff or
  published logs.
