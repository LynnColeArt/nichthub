# Mission Specification: Operational Self-Hosting Alpha

**Mission Branch**: `feat/self-hosting-alpha-loop`  
**Created**: 2026-08-30  
**Status**: Draft  
**Input**: Make Nichthub capable of sustaining its own public development, not merely demonstrating one prearranged loop.

## Intent

Nichthub's repository-native event model has survived unit, integration,
conflict-recovery, governance, and live public-transport testing. The project
now needs to cross from protocol experiment to operational alpha: maintainers
must be able to evolve trust policy, replace or add device keys without sharing
private material, select and bound what they replicate, work safely from
shallow clones, and then use those capabilities to govern Nichthub itself.

The mission succeeds when the project performs a real public workflow with
role-distinct actor identities and a policy amendment governed by the policy it
replaces. A fresh participant must be able to reconstruct the exact signed
facts without a Nichthub server, GitHub collaboration feature, copied private
identity, or Docker daemon.

The immutable-history and explicit-authority boundaries remain unchanged.
Identity continuity never silently transfers project authority; policies must
name every actor whose claims count. Corrections, rotations, and conflict
recovery add signed facts rather than rewriting history.

## User Scenarios & Testing

### User Story 1 - Evolve Project Trust Safely (Priority: P1)

As a maintainer, I want to inspect, validate, and propose a policy amendment so
that Nichthub can add or remove trusted actors and requirements without
hand-editing opaque trust state or letting a new policy authorize itself.

**Why this priority**: Sustainable self-hosting requires the project to evolve
who may review, run code, decide, and merge while preserving the immutable
base-policy invariant.

**Independent Test**: Starting from a committed policy with one maintainer,
prepare an amendment that adds a second trusted reviewer and runner, inspect
the effective before/after rules, and merge it only after satisfying the old
policy.

**Acceptance Scenarios**:

1. **Given** a valid committed policy and a proposed replacement, **when** the maintainer validates the amendment, **then** the result identifies both policy digests, every changed trust role and threshold, and any lockout or unsatisfiable rule before a candidate is opened.
2. **Given** an amendment candidate whose head contains more permissive rules, **when** readiness and acceptance are evaluated, **then** only the exact policy bytes from the signed base authorize its evidence and decisions.
3. **Given** an invalid amendment that removes every maintainer or makes a threshold unsatisfiable, **when** it is validated or proposed, **then** it is rejected without publishing a policy candidate.

---

### User Story 2 - Rotate and Separate Device Identities (Priority: P1)

As a maintainer, I want each device to use a distinct private key and to create
verifiable continuity when a key is deliberately replaced so that one copied
secret is not shared by concurrent writers.

**Why this priority**: A project cannot operate indefinitely from one key on
one workstation, while shared-key concurrent append would violate the actor
chain's single-writer invariant.

**Independent Test**: Establish a second actor on a separate clone, publish a
mutually verifiable device or successor relationship, rotate a disposable
actor to a new key, and confirm that project authority changes only after an
explicit policy amendment.

**Acceptance Scenarios**:

1. **Given** a newly initialized device identity, **when** an existing actor authorizes it and the new actor accepts, **then** every clone derives the same relationship from signatures made by both distinct private keys.
2. **Given** a planned key rotation, **when** the old actor names a successor and the successor accepts, **then** the old history remains verifiable, the successor begins its own single-writer history, and local commands stop using the retired key by default.
3. **Given** a valid identity relationship but a policy that names only the old actor, **when** the new actor reviews, runs, or decides, **then** those claims remain valid signed facts but do not satisfy project policy.
4. **Given** contradictory, cyclic, incomplete, or replayed identity claims, **when** histories are projected, **then** the conflict is visible and no implicit authority or unambiguous rotation is inferred.

---

### User Story 3 - Replicate Only Accepted Collaboration State (Priority: P1)

As a participant, I want to choose which remote actors and candidates I import
under explicit resource limits so that an unrelated or hostile history cannot
force unbounded work or disable collaboration with actors I trust.

**Why this priority**: Open collaboration turns fetched refs and objects into
hostile input. Fetching every advertised actor before deciding whose claims
matter is not a sustainable public default.

**Independent Test**: Offer valid, malformed, oversized, and unselected actor
histories from one disposable remote; synchronize an explicit selection under
small test budgets; and verify that only selected, valid facts become visible
while failures identify their exact actor or dependency.

**Acceptance Scenarios**:

1. **Given** a remote advertising multiple actor and proposal refs, **when** a participant synchronizes an explicit selection, **then** unselected histories and candidate refs do not enter that remote's visible Nichthub projection.
2. **Given** one selected valid actor and one selected invalid actor, **when** synchronization runs, **then** the valid actor can be promoted independently and the invalid actor is quarantined with an exact validation error.
3. **Given** a selected history that exceeds an event, attachment, object, or byte budget, **when** synchronization reaches the configured limit, **then** it aborts that import before promotion and reports the exceeded budget.
4. **Given** a selected event whose required proposal, request, decision, or actor history is absent, **when** projection is attempted, **then** the client identifies the missing full ID and the additional selection required; it does not reinterpret the event.

---

### User Story 4 - Operate from a Shallow Clone (Priority: P2)

As a contributor using a shallow clone, I want Nichthub to detect missing Git
history and obtain or request only the exact history needed so that shallow
state never weakens policy or evidence verification.

**Why this priority**: Shallow clones are ordinary Git usage. Silent missing
history would turn a transport optimization into an integrity failure.

**Independent Test**: Clone the repository with limited depth, select a public
candidate whose base or policy is outside that depth, and verify either bounded
recovery or precise fail-closed instructions followed by successful retry.

**Acceptance Scenarios**:

1. **Given** a shallow clone that already contains every object needed for an operation, **when** Nichthub inspects or verifies it, **then** the operation succeeds without requiring an unshallow clone.
2. **Given** a missing candidate base, policy blob, code object, or actor-chain predecessor, **when** an integrity-sensitive command runs, **then** it fails before accepting evidence and names the missing object or history boundary.
3. **Given** operator consent and sufficient resource budget, **when** the missing history is recoverable from the selected remote, **then** Nichthub obtains only the necessary history and retries verification without broadening trust selection.

---

### User Story 5 - Govern Nichthub with Nichthub (Priority: P1)

As a Nichthub maintainer, I want the project to land real changes through these
capabilities so that self-hosting becomes the normal operating model rather
than a one-time demonstration.

**Why this priority**: The individual features matter only if they compose into
a public, inspectable workflow with meaningful role separation.

**Independent Test**: Use the public Git remote to amend policy under its old
rules, establish a second actor, and land a subsequent real candidate with a
non-author actor supplying required review or CI evidence. Verify all facts
from a fresh selected and shallow clone.

**Acceptance Scenarios**:

1. **Given** the current one-actor policy, **when** an amendment adds a second trusted actor, **then** the old policy alone governs the amendment candidate and the new actor gains no qualifying authority before the amended policy becomes a signed base for a later candidate.
2. **Given** the amended policy forbids author approval and trusts a separate reviewer or runner, **when** a later real candidate is evaluated, **then** it cannot become ready until the required second actor publishes exact qualifying evidence.
3. **Given** a successful governed merge, **when** collaboration refs and the primary branch are published separately, **then** the remote advertises the exact actor, candidate, and branch refs and a fresh clone reconstructs the same merged state.

### Edge Cases

- The policy file is syntactically valid but removes the last effective path to
  satisfy its own future acceptance threshold.
- A successor key accepts a different rotation than the old key published, or
  one actor appears in competing rotation lineages.
- A retired local key is still present for historical verification but must not
  sign ordinary new events accidentally.
- A remote advertises thousands of actors, a very long actor chain, oversized
  attachments, invalid signatures, ref/object mismatches, or relationships to
  unselected facts.
- A valid actor depends on a candidate authored by an unselected actor.
- One selected actor is hostile while another selected actor is valid.
- A shallow boundary cuts through an actor chain, proposal base, policy blob,
  pipeline definition, or merge ancestry.
- Collaboration refs publish successfully while the primary branch push fails,
  or vice versa.
- A live candidate conflicts after policy amendment; the repair must use an
  immutable revision and re-earn evidence under the exact revised base.

## Domain Language

- **Operational self-hosting**: Nichthub development routinely governed by
  signed Nichthub facts rather than a provider's collaboration database.
- **Actor**: One Ed25519 signing key and its single-writer append-only history.
  Do not use “user” or “person” as an exact synonym.
- **Device identity**: A distinct actor controlled from one device and linked
  to another actor through mutually signed continuity facts. It does not reuse
  the other actor's private key.
- **Planned rotation**: A mutually signed transition from an old actor to a new
  successor while the old key is still available. Lost-key recovery is not
  implied.
- **Policy amendment**: A normal immutable proposal whose head changes project
  policy and whose authorization is evaluated exclusively under its signed
  base policy.
- **Trust selection**: Local, untracked instruction naming the remote actors or
  candidates a participant chooses to import. Selection controls replication;
  project policy separately controls whose claims count.
- **Quarantine**: Fetched candidate state that has not passed validation and
  therefore is not visible in the accepted local projection.
- **Candidate**: One immutable proposal or proposal revision bound to exact
  base and head commits. Avoid using “PR” as a synonym.
- **Public remote**: An ordinary publicly readable Git remote. Its provider is
  transport, not a collaboration authority.

## Requirements

### Functional Requirements

| ID | Title | User Story | Priority | Status |
|----|-------|------------|----------|--------|
| FR-001 | Inspect effective policy | As a maintainer, I can inspect the exact policy bytes, digest, roles, thresholds, and source commit that govern a candidate or revision. | High | Open |
| FR-002 | Validate policy amendments | As a maintainer, I can validate a proposed policy against structural rules and see a deterministic before/after summary of role and threshold changes before publishing it. | High | Open |
| FR-003 | Preserve old-policy authorization | As a participant, I can verify that a policy-amendment candidate is reviewed, accepted, and merged only under the exact valid policy from its signed base. | High | Open |
| FR-004 | Reject governance lockout | As a maintainer, I am prevented from proposing a policy that has no maintainer or whose configured thresholds cannot be satisfied by its listed actors. | High | Open |
| FR-005 | Establish distinct device actors | As a maintainer, I can link an existing actor and a separately generated device actor through claims signed by both keys without copying either private key. | High | Open |
| FR-006 | Perform planned key rotation | As an actor owner, I can name and mutually confirm a successor actor, preserve the predecessor history, and make the successor the default local signer without appending concurrently to one actor chain. | High | Open |
| FR-007 | Project identity continuity safely | As a participant, I can inspect device and rotation relationships, including incomplete, conflicting, replayed, or cyclic claims, without clients silently choosing an authority winner. | High | Open |
| FR-008 | Keep project authority explicit | As a maintainer, I can verify that device or rotation relationships never make an actor trusted for governance unless the applicable project policy explicitly lists that actor. | High | Open |
| FR-009 | Select replicated actors and candidates | As a participant, I can persist a local per-remote selection of full actor and candidate IDs and synchronize only that selection. | High | Open |
| FR-010 | Quarantine before projection | As a participant, I see newly fetched selected state only after its refs, objects, signatures, chains, relationships, attachments, and candidate heads pass validation. | High | Open |
| FR-011 | Isolate invalid selections | As a participant, I can retain and use independently valid selected histories when another selected history fails validation, with the failing actor or candidate identified exactly. | High | Open |
| FR-012 | Enforce replication budgets | As a participant, I can set event, attachment, object, and byte budgets; an import that exceeds one is rejected before its state is promoted into the accepted projection. | High | Open |
| FR-013 | Explain missing dependencies | As a participant, I receive full missing event, actor, candidate, request, or object IDs and an exact selection or history-recovery action when a selected fact lacks required context. | High | Open |
| FR-014 | Detect shallow boundaries | As a contributor, I am told when a shallow boundary prevents exact policy, candidate, actor-chain, pipeline, or merge verification before any trust-sensitive state change occurs. | High | Open |
| FR-015 | Recover bounded shallow history | As a contributor, I can explicitly request bounded retrieval of the missing selected history without implicitly selecting unrelated actors or candidates. | Medium | Open |
| FR-016 | Complete a role-distinct public loop | As a maintainer, I can amend policy under its old rules and land a later real candidate requiring qualifying evidence from a second actor through the public remote. | High | Open |
| FR-017 | Reconstruct from a fresh clone | As an independent participant, I can selectively synchronize a fresh shallow clone and verify the public policy, identities, candidate code, CI, review, decision, and merge chain without copied private state. | High | Open |
| FR-018 | Preserve recoverable failures | As a maintainer, I receive exact blocking identifiers and recovery guidance while existing signed facts and previously clean targets remain intact. | High | Open |
| FR-019 | Automate the operational scenario | As a contributor, I can run a deterministic offline acceptance scenario over a disposable Git remote covering amendment, identity continuity, selected replication, shallow recovery, and role-distinct governance. | Medium | Open |
| FR-020 | Record public operating evidence | As a contributor, I can follow project documentation and inspect a durable record of the full public event, ref, policy, and commit identifiers produced by the inaugural operational loop. | Medium | Open |

### Non-Functional Requirements

| ID | Title | Requirement | Category | Priority | Status |
|----|-------|-------------|----------|----------|--------|
| NFR-001 | Deterministic projection | Every clone given the same accepted refs and objects derives identical identity, policy, candidate, and evidence relationships in 100% of repeated test runs. | Reliability | High | Open |
| NFR-002 | Secret isolation | Zero private keys, access tokens, temporary credentials, or host-private paths appear in tracked files, signed logs, exported public identity facts, or verification records. | Security | High | Open |
| NFR-003 | Bounded promotion | No import promotes more events, attachments, objects, or bytes than the participant's configured positive limits; boundary tests exercise each limit at one below, exactly at, and one above its value. | Security | High | Open |
| NFR-004 | Failure isolation | In a synchronization containing at least one valid and one invalid selected actor, 100% of independently valid selections remain usable and every invalid selection remains outside the accepted projection. | Reliability | High | Open |
| NFR-005 | Shallow integrity | Across all acceptance cases, zero review, run, decision, or merge operations proceed when a required object lies beyond an unresolved shallow boundary. | Security | High | Open |
| NFR-006 | Backward readability | All existing valid issue, proposal, revision, review, run, decision, and merge histories continue to verify with unchanged event IDs after the new behavior is introduced. | Compatibility | High | Open |
| NFR-007 | Repeatable acceptance | The complete offline operational scenario passes in three consecutive runs and finishes within 120 seconds per run on the development host. | Reliability | High | Open |
| NFR-008 | Provider independence | The live loop requires zero hosting-provider collaboration API calls, zero mandatory Nichthub services, and zero Docker operations. | Portability | High | Open |
| NFR-009 | Bounded failure | Every tested unsafe or incomplete transition returns a non-zero result within its configured timeout and leaves the previously clean target with no unrecorded commit. | Safety | High | Open |
| NFR-010 | Public convergence | A fresh clone reconstructs and verifies 100% of the exact identifiers in the inaugural public verification record. | Reliability | High | Open |

### Constraints

| ID | Title | Constraint | Category | Priority | Status |
|----|-------|------------|----------|----------|--------|
| C-001 | Repository-native boundary | Collaboration and identity-continuity state remains signed, content-addressed repository data transported through Git; no mandatory service or service database may be introduced. | Product | High | Open |
| C-002 | Actor remains a key | An actor identifier remains the digest of one public signing key and its history remains single-writer; a person or device group is a projection over distinct actors, not a shared actor secret. | Protocol | High | Open |
| C-003 | Explicit policy authority | Identity continuity never transfers maintainer, reviewer, runner, or decision authority automatically; only the exact applicable project policy grants those roles. | Governance | High | Open |
| C-004 | Immutable history | Published events, candidates, attachments, decisions, rotations, and merge facts are never edited or replaced; corrections add signed facts. | Governance | High | Open |
| C-005 | Planned rotation only | Lost-key recovery, social recovery, and compromise adjudication remain outside this mission; rotation requires the predecessor key and successor key to participate. | Scope | High | Open |
| C-006 | Local selection | Trust-selection configuration remains local and untracked by default; publishing a selection must never be required to read public collaboration data. | Privacy | High | Open |
| C-007 | No silent history expansion | Shallow recovery or dependency completion requires explicit operator intent and may not broaden the selected actor or candidate set silently. | Security | High | Open |
| C-008 | Existing execution safety | The default isolated runner remains the live execution path, unsafe host execution remains explicitly opted in, and no Docker dependency is added. | Security | High | Open |
| C-009 | Experimental compatibility | Existing version-0 event payloads and IDs remain unchanged; any new event kinds and failure behavior are documented as experimental extensions rather than a stable compatibility promise. | Protocol | High | Open |
| C-010 | Focused deferrals | Secrets, configurable runner networking, strong CPU/memory/disk quotas, portable/container backends, shared moderation, redaction semantics, merge queues, and shared-key concurrent writers remain outside this mission. | Scope | High | Open |

### Key Entities

- **Actor**: One public key fingerprint, one private signer, and one
  append-only event chain.
- **Identity Continuity Claim**: A signed device authorization or planned
  rotation naming another actor, completed only by a corresponding claim from
  that actor.
- **Identity Projection**: Deterministic relationships among actors, including
  active, retired, incomplete, and conflicting continuity state; it grants no
  project role by itself.
- **Policy Amendment**: An immutable candidate whose head changes project
  policy and whose base policy governs its own evidence and acceptance.
- **Replication Selection**: Local per-remote actor and candidate IDs plus
  positive resource budgets.
- **Quarantined Import**: Selected remote refs and objects that are available
  for validation but not yet visible in the accepted collaboration projection.
- **Shallow Boundary**: A recorded Git history cutoff that may hide a required
  predecessor, base, policy, pipeline, or merge ancestor.
- **Operational Verification Record**: Public IDs and observations proving the
  inaugural policy amendment and role-distinct self-hosting loop.

## Scope Boundaries

### In Scope

- Policy inspection, amendment validation, change preview, and safe proposal
  preparation.
- Mutually signed device relationships and planned key rotation across
  distinct single-writer actors.
- Explicit, locally configured actor/candidate replication with validation
  quarantine, failure isolation, dependency guidance, and resource budgets.
- Shallow-boundary detection and explicit bounded history recovery.
- Deterministic black-box acceptance over a disposable ordinary Git remote.
- A real public policy amendment followed by a role-distinct Nichthub-governed
  change and fresh shallow-clone verification.
- Protocol, security, compatibility, operator, and public evidence
  documentation for all introduced behavior.

### Out of Scope

- Lost-key, compromised-key, social, organizational, or threshold recovery.
- Concurrent writers sharing a private actor key.
- Secrets or credentials supplied to repository-defined execution.
- Configurable runner networking or strong CPU, memory, disk, and process
  quotas.
- Portable or container runner backends.
- Shared moderation lists, remote deletion, or redaction guarantees.
- Merge queues, notification delivery, general search, web UI, and project
  discovery.
- A stable protocol-version compatibility promise.

## Assumptions and Dependencies

- The public remote continues to advertise and retain selected custom refs and
  ordinary reachable Git objects.
- The current actor remains available to authorize the inaugural planned
  policy amendment and any disposable rotation proof.
- A second private actor can be created in a separate clone and kept untracked;
  its public actor identity may be committed to policy and signed events.
- Project policy remains a normal versioned repository file, and a policy
  change takes effect only for candidates whose signed base contains it.
- Replication selection determines what facts enter the local projection;
  policy evaluation independently determines which valid facts qualify.
- The configured `test` pipeline and default isolated runner are available for
  the live public proof.
- Collaboration-ref publication and primary-branch publication remain
  separate explicit actions and must both be verified.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A policy amendment adding a second trusted actor reaches the
  public primary branch with acceptance evidence authorized exclusively by the
  policy digest from its signed base.
- **SC-002**: A subsequent real candidate cannot become ready from its author's
  evidence alone and reaches the public primary branch only after a distinct
  actor supplies every role-distinct review or CI claim required by policy.
- **SC-003**: A planned disposable rotation produces mutually signed old/new
  actor continuity that three independent projections derive identically,
  while the successor has no qualifying project authority until policy names
  it.
- **SC-004**: Acceptance tests synchronize a remote containing valid, invalid,
  oversized, and unselected histories and promote 100% of independently valid
  selections, 0% of invalid or over-budget selections, and 0 unselected refs
  into the accepted projection.
- **SC-005**: A fresh depth-limited clone detects every missing integrity
  dependency, performs only explicitly approved bounded recovery, and then
  verifies 100% of the inaugural public identifiers without a private key.
- **SC-006**: The complete offline operational scenario passes three
  consecutive times within 120 seconds per run.
- **SC-007**: The live workflow uses zero GitHub collaboration API calls, zero
  mandatory Nichthub services, zero shared private actor keys, and zero Docker
  operations.
- **SC-008**: All existing event IDs in the public seven-event baseline remain
  unchanged and verifiable after the mission.
- **SC-009**: All automated tests and static checks pass with no private
  material present in tracked changes, signed logs, or public verification
  artifacts.
