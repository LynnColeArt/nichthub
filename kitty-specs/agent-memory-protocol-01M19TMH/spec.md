# Mission Specification: Agent Memory Protocol

**Mission Branch**: `feat/agent-memory-protocol`  
**Created**: 2026-08-30  
**Status**: Draft  
**Input**: Give agents working in Git projects a durable, distributed, attributable memory layer without turning retrieved text into trusted instructions.

## Intent

Git preserves what changed. Nichthub should also preserve what project
participants knew, intended, assumed, attempted, decided, verified, and handed
off at the time of those changes. An agent arriving in a later session or fresh
clone must be able to reconstruct relevant project cognition without relying on
a vendor account, centralized memory service, mutable summary file, or raw chat
archive.

Memories are signed claims, not truth. A valid signature proves who recorded a
memory and protects its exact contents; it does not prove correctness, safety,
current applicability, or permission to act. Recall must preserve provenance,
trust classification, repository-state anchors, lifecycle relationships, and
supporting evidence. Memory content is always returned as inert data and never
promoted automatically into instructions or executable actions.

Canonical memory lives as immutable repository-native protocol data and
travels through Git. Search indexes, summaries, and embeddings are optional,
local, rebuildable projections. The memory protocol depends on the operational
self-hosting alpha's durable actor identities, explicit policy, selected
replication, quarantine, budgets, and shallow-history handling.

## User Scenarios & Testing

### User Story 1 - Record Durable Project Cognition (Priority: P1)

As a coding agent or human collaborator, I want to record a concise structured
memory bound to exact project state so that later participants can distinguish
an observation, decision, assumption, attempt, verification, or handoff and
inspect what supports it.

**Why this priority**: Durable recall is only valuable when the recorded unit
has explicit meaning, attribution, scope, and evidence rather than being an
unstructured transcript fragment.

**Independent Test**: Record one memory of every supported kind against exact
commits and Nichthub subjects, synchronize them to another clone, and verify
identical IDs, content, anchors, authors, applicability, and evidence.

**Acceptance Scenarios**:

1. **Given** an agent at an exact commit, **when** it records a decision with rationale and evidence, **then** the resulting immutable memory identifies its actor, kind, repository anchor, applicability scope, exact evidence, and content digest.
2. **Given** an attempted approach that failed, **when** the agent records the attempt and outcome, **then** a later agent can see both what was tried and the exact project state in which it failed without replaying raw conversation.
3. **Given** an explicit session handoff, **when** it is recorded, **then** it identifies completed work, current assumptions, blockers, and proposed next actions while preserving each statement as attributable memory data.

---

### User Story 2 - Recall Relevant Memory Safely (Priority: P1)

As an agent beginning or resuming work, I want bounded recall for the current
repository state, topic, file, issue, proposal, or event so that I regain useful
context without ingesting an entire history or silently following stored
instructions.

**Why this priority**: Recall is the product boundary where stale, hostile, or
irrelevant text can become prompt injection. Relevance and safety must be
explicit rather than delegated to blind concatenation.

**Independent Test**: Query a mixed set of trusted, untrusted, active,
superseded, challenged, and state-inapplicable memories under a strict result
budget; verify deterministic filtering, full provenance, lifecycle labels, and
an inert structured output envelope.

**Acceptance Scenarios**:

1. **Given** memories anchored to different commits and subjects, **when** an agent recalls at the current commit with explicit filters, **then** only memories whose declared applicability matches are returned by default.
2. **Given** memory content that says to ignore policy or execute a command, **when** it is recalled, **then** the text remains quoted data with actor, trust, and warning metadata and causes no action or instruction-priority change.
3. **Given** more matching memories than the result budget allows, **when** recall runs, **then** it returns a deterministic bounded subset plus an explicit truncation count and continuation mechanism.
4. **Given** memories from actors outside the applicable memory policy, **when** default recall runs, **then** they are excluded or clearly separated; explicit inspection may include them without upgrading their trust classification.

---

### User Story 3 - Correct Memory Without Rewriting It (Priority: P1)

As a participant, I want to supersede or retract my own memory and challenge
another actor's memory so that obsolete or disputed cognition remains
auditable without continuing to appear as uncontested current context.

**Why this priority**: Immutable false or stale memory is more dangerous than
no memory unless correction and disagreement are first-class, deterministic,
and visible.

**Independent Test**: Publish supersession, retraction, and cross-actor
challenge facts in different delivery orders and verify identical active,
superseded, retracted, challenged, and conflicting projections.

**Acceptance Scenarios**:

1. **Given** an actor's earlier assumption, **when** the same actor publishes a replacement that explicitly supersedes it, **then** default recall shows the replacement while preserving a navigable immutable lineage to the original.
2. **Given** an incorrect memory, **when** its author retracts it, **then** it is excluded from active default recall but remains available for audit with the retraction reason.
3. **Given** one actor disputes another actor's memory, **when** a challenge is published, **then** neither original nor challenge is silently erased and policy determines how the dispute is presented or filtered.
4. **Given** two memories that may conflict only in natural-language meaning, **when** no explicit relationship names the conflict, **then** the protocol does not pretend to have inferred semantic contradiction automatically.

---

### User Story 4 - Exchange Memory Without Coupling Collaboration (Priority: P1)

As a project participant, I want to select which agents' memory streams I
replicate independently from issues, proposals, reviews, and CI so that a large
or hostile memory corpus cannot block ordinary collaboration state.

**Why this priority**: Agent memory can grow rapidly and contains untrusted
text. Reusing one indivisible collaboration chain would make memory volume and
prompt risk prerequisites for reviewing code.

**Independent Test**: Offer selected and unselected valid memory streams plus
an invalid and oversized stream alongside valid collaboration refs; verify
quarantine, budgets, independent promotion, and uninterrupted collaboration
projection.

**Acceptance Scenarios**:

1. **Given** an actor with both collaboration events and memory, **when** a participant selects only collaboration state, **then** no memory stream is required or imported.
2. **Given** selected memory streams within budget, **when** synchronization completes, **then** the receiving clone reconstructs their signed memory and lifecycle relationships exactly.
3. **Given** an invalid, over-budget, or dependency-incomplete memory stream, **when** it is synchronized, **then** it remains quarantined and does not disable independently valid collaboration or memory streams.
4. **Given** a shallow clone, **when** selected memory references unavailable evidence or anchors, **then** recall identifies the missing full IDs and requires explicit bounded recovery before presenting the memory as evidence-resolved.

---

### User Story 5 - Rebuild and Query Local Memory Views (Priority: P2)

As an agent operator, I want every local search index and derived summary to be
rebuildable from signed accepted memory so that convenience layers cannot
become a hidden canonical database.

**Why this priority**: Fast recall may need indexes or embeddings, but project
truth must remain portable and independently verifiable.

**Independent Test**: Delete the local memory index, rebuild it twice from the
same accepted refs, and verify identical membership and lifecycle state while
queries return the same deterministically ordered records.

**Acceptance Scenarios**:

1. **Given** accepted memory refs and no local index, **when** an index is rebuilt, **then** all verified records, anchors, lifecycle edges, evidence status, and policy classifications are reconstructed without a network service.
2. **Given** optional semantic-search data is unavailable or generated by a different model, **when** exact filters and lexical recall run, **then** canonical memory identity and lifecycle results remain unchanged.
3. **Given** an index that disagrees with the signed source data, **when** verification runs, **then** the index is discarded or rebuilt rather than changing canonical memory.

---

### User Story 6 - Integrate Agents Without Vendor Lock-In (Priority: P2)

As an agent-tool author, I want stable structured record and recall interfaces
so that different agents can participate without making any model vendor,
orchestrator, vector database, or prompt format part of the protocol.

**Why this priority**: A distributed memory layer loses its value if only one
agent harness can read or write it safely.

**Independent Test**: Drive record, recall, handoff, and lifecycle operations
through machine-readable interfaces from two independent test adapters and
verify identical protocol events and recall envelopes.

**Acceptance Scenarios**:

1. **Given** two agent adapters using the same repository and actor permissions, **when** each records equivalent structured input, **then** both produce valid protocol memories without vendor-specific fields becoming canonical.
2. **Given** a bounded recall envelope, **when** an adapter consumes it, **then** every memory retains its full ID, provenance, trust classification, applicability, evidence resolution, and inert content boundary.
3. **Given** an agent lacks permission to publish a memory kind under project policy, **when** it attempts to record it, **then** the claim is either refused locally or remains non-qualifying and cannot masquerade as an authorized project decision.

### Edge Cases

- A memory anchors a commit that exists but is not an ancestor of the queried
  project state.
- A memory names a path whose blob changed, moved, or disappeared after its
  anchor.
- Evidence is valid but unselected, shallow-missing, later challenged, or bound
  to a different candidate revision.
- An agent records plausible-looking shell commands, system-prompt language,
  encoded control characters, or extremely repetitive text.
- The same author publishes branching supersessors or retracts a memory after
  another actor challenges it.
- Two trusted actors explicitly challenge one another and policy has no
  configured dispute preference.
- Memory events arrive before their anchors, evidence, or lifecycle targets.
- A selected memory stream is huge while the same actor's collaboration chain
  is small and valid.
- An index was produced under older policy or from a different accepted ref set.
- A memory contains information that later should not have been published;
  retraction cannot guarantee erasure from replicated Git object databases.
- A compromised or rotated agent key authored earlier memory; historical
  signature validity and current policy trust must remain distinct.

## Domain Language

- **Project memory**: A concise structured signed claim about project cognition
  bound to explicit repository or Nichthub state. It is not raw conversation.
- **Memory record**: One immutable project-memory fact with kind, actor,
  content, anchors, applicability, and evidence references.
- **Memory stream**: A separately replicated append-only sequence of memories
  signed by one actor. It is not the actor's collaboration event chain.
- **Anchor**: Exact Git or Nichthub identity locating the state or subject a
  memory concerns.
- **Applicability**: The author's explicit statement that a memory concerns an
  exact state, descendants of a commit, or a named Nichthub subject. It is not
  inferred from prose.
- **Evidence**: Exact content-addressed references supporting a memory. Evidence
  availability and validity do not make the memory true automatically.
- **Supersession**: A same-author immutable replacement edge that removes the
  predecessor from active default recall while preserving its lineage.
- **Retraction**: A same-author signed statement that an earlier memory should
  no longer be treated as active.
- **Challenge**: Another actor's signed dispute of a memory. It does not edit or
  erase the target.
- **Trust classification**: Derived policy result describing whether an actor
  and memory kind qualify for a recall context. It is separate from signature
  validity.
- **Recall envelope**: Bounded structured output that preserves memory content
  as inert data alongside provenance, lifecycle, applicability, evidence, and
  trust metadata.
- **Canonical memory**: Signed repository data. Indexes, embeddings, rankings,
  summaries, and prompt renderings are derived and disposable.

## Requirements

### Functional Requirements

| ID | Title | User Story | Priority | Status |
|----|-------|------------|----------|--------|
| FR-001 | Record structured memory kinds | As an agent or human, I can record observations, decisions, assumptions, attempts with outcomes, verifications, and handoffs as distinct attributable memory kinds. | High | Open |
| FR-002 | Bind exact anchors | As a recorder, I must bind each memory to an exact Git commit and may additionally bind exact blobs, repository paths at that commit, issues, proposals, events, policies, pipeline definitions, or run evidence. | High | Open |
| FR-003 | Declare applicability | As a recorder, I explicitly declare whether a memory applies only to its exact state, to descendant states, or to a named Nichthub subject; recall does not infer applicability from prose. | High | Open |
| FR-004 | Attach exact evidence | As a recorder, I can attach ordered full content-addressed evidence IDs, while verification memories require at least one resolvable supporting reference. | High | Open |
| FR-005 | Require deliberate capture | As an operator, I know memories are created only by an explicit record or approved adapter action; the system does not automatically publish raw prompts, responses, terminal history, environment data, or chat transcripts. | High | Open |
| FR-006 | Separate memory streams | As a participant, I can replicate an actor's collaboration history without replicating that actor's memory streams, and vice versa. | High | Open |
| FR-007 | Select and synchronize memory | As a participant, I can select full memory stream IDs per remote and synchronize them through the same quarantine, budget, validation, and shallow-recovery boundaries as other hostile repository data. | High | Open |
| FR-008 | Recall by explicit context | As an agent, I can recall by current commit, exact subject, path-at-commit, topic, memory kind, actor, trust class, and lifecycle state. | High | Open |
| FR-009 | Bound every recall | As an operator, I can bound recall by record count and encoded output bytes and receive deterministic truncation and continuation information. | High | Open |
| FR-010 | Preserve provenance in recall | As a consumer, every recalled item includes its full memory ID, actor, signature validity, anchor, applicability, lifecycle state, evidence status, trust classification, and exact content digest. | High | Open |
| FR-011 | Keep content inert | As an agent integrator, I receive memory text in a structurally separate data field with controls safely encoded and no instruction priority, command execution, tool invocation, or authorization side effect. | High | Open |
| FR-012 | Apply policy-filtered defaults | As a maintainer, I can define which actor identities and memory kinds qualify for default project recall without preventing explicit inspection of other valid signed claims. | High | Open |
| FR-013 | Inspect untrusted memory explicitly | As an investigator, I can request non-qualifying memory with a visible untrusted classification and warning, without changing project policy or default recall. | Medium | Open |
| FR-014 | Supersede own memory | As a recorder, I can publish a replacement for my own memory that creates a deterministic immutable lineage and removes the predecessor from active default recall. | High | Open |
| FR-015 | Retract own memory | As a recorder, I can retract my own memory with a reason while preserving the original and retraction for audit. | High | Open |
| FR-016 | Challenge another memory | As a participant, I can explicitly challenge another actor's memory with a typed reason and evidence; the challenge cannot erase or impersonate the target author. | High | Open |
| FR-017 | Project lifecycle deterministically | As a participant, I derive identical active, superseded, retracted, challenged, branching, and dependency-missing memory state regardless of event delivery order. | High | Open |
| FR-018 | Produce and recall handoffs | As an agent, I can record and retrieve a handoff containing completed work, open assumptions, blockers, and proposed next actions, each bound to exact current project state. | Medium | Open |
| FR-019 | Rebuild local indexes | As an operator, I can discard, verify, and rebuild local memory indexes from accepted signed streams without changing canonical memory or requiring network access. | High | Open |
| FR-020 | Support exact and lexical retrieval | As an agent, I can use exact filters and deterministic lexical search without embeddings; optional semantic indexes may improve ranking but never change memory identity, lifecycle, or trust. | Medium | Open |
| FR-021 | Expose stable machine interfaces | As an adapter author, I can record memories and consume recall envelopes through versioned machine-readable input and output without vendor-specific canonical fields. | High | Open |
| FR-022 | Explain unresolved dependencies | As a consumer, I see full missing anchor, evidence, lifecycle-target, actor, or stream IDs and an exact selected recovery action before unresolved memory is represented as evidence-complete. | High | Open |
| FR-023 | Preserve collaboration independence | As a collaborator, I can continue to inspect issues, proposals, reviews, CI, and governance when any memory stream is absent, invalid, over-budget, or deliberately unselected. | High | Open |
| FR-024 | Verify from a fresh clone | As an independent agent, I can selectively synchronize and recall the same memories and lifecycle projection from a fresh clone without copied private keys, indexes, embeddings, or vendor state. | High | Open |

### Non-Functional Requirements

| ID | Title | Requirement | Category | Priority | Status |
|----|-------|-------------|----------|----------|--------|
| NFR-001 | Provenance completeness | 100% of records returned by every recall mode include full memory ID, actor ID, anchor, applicability, lifecycle state, evidence status, trust classification, and content digest. | Integrity | High | Open |
| NFR-002 | Inert-content safety | Across adversarial acceptance fixtures, 0 recalled memory strings execute commands, invoke tools, alter instruction priority, or enter an output field designated as trusted instructions. | Security | High | Open |
| NFR-003 | Default recall bound | Without explicit overrides, one recall returns at most 20 records and at most 65,536 encoded content bytes, with deterministic truncation metadata when more matches exist. | Resource safety | High | Open |
| NFR-004 | Record bound | One memory contains at most 65,536 UTF-8 content bytes, 64 evidence references, and 32 normalized topic labels; one-below, exact, and one-above tests cover every limit. | Resource safety | High | Open |
| NFR-005 | Projection convergence | Three projections receiving the same valid memories in different orders derive byte-identical lifecycle and trust-classification fixtures. | Reliability | High | Open |
| NFR-006 | Collaboration isolation | In every invalid, missing, or over-budget memory-stream test, 100% of independently valid collaboration events remain readable and unchanged. | Reliability | High | Open |
| NFR-007 | Index reproducibility | Two clean index rebuilds over the same accepted refs produce identical record membership, anchors, lifecycle edges, evidence resolution, and trust classes. | Reliability | High | Open |
| NFR-008 | Local recall performance | With 10,000 accepted memory records, exact-filter and lexical recall complete within 1 second at p95 after indexing on the development host, and a clean index rebuild completes within 30 seconds. | Performance | Medium | Open |
| NFR-009 | Secret isolation | Zero actor private keys, access tokens, environment values, or automatically captured prompt/terminal transcripts appear in protocol fixtures, published memories, indexes intended for export, or recall logs. | Security | High | Open |
| NFR-010 | Backward compatibility | All existing collaboration event payloads and public event IDs remain unchanged, and collaboration-only clones require zero memory refs to reproduce their prior projections. | Compatibility | High | Open |
| NFR-011 | Provider independence | Record, synchronization, indexing, lifecycle projection, and recall require zero model-vendor APIs, hosting-provider collaboration APIs, mandatory memory services, or Docker operations. | Portability | High | Open |
| NFR-012 | Fresh-clone convergence | A fresh selected clone reconstructs 100% of a reference corpus's memory IDs, anchors, lifecycle relationships, and exact-filter results without receiving the originating clone's local index. | Reliability | High | Open |

### Constraints

| ID | Title | Constraint | Category | Priority | Status |
|----|-------|------------|----------|----------|--------|
| C-001 | Operational-alpha dependency | Implementation begins only after the operational self-hosting alpha provides durable identities, explicit policy amendment, selected quarantined replication, budgets, and shallow recovery. | Dependency | High | Open |
| C-002 | Repository-native canon | Canonical memory consists only of signed content-addressed repository data transported by Git; no service database or vendor account is authoritative. | Architecture | High | Open |
| C-003 | Separate replication roots | Memory streams must not make collaboration event availability depend on fetching an actor's memory corpus. | Architecture | High | Open |
| C-004 | Claims are not truth | Signature validity, policy qualification, evidence availability, and semantic correctness remain separate properties; no layer may collapse them into one trusted flag. | Integrity | High | Open |
| C-005 | No prompt authority | Memory content never becomes a system, developer, user, or tool instruction by protocol semantics; an adapter must preserve the inert-data boundary. | Security | High | Open |
| C-006 | Immutable correction | Published memory is never edited or overwritten; supersession, retraction, challenge, and correction add signed facts. | Protocol | High | Open |
| C-007 | No automatic transcript capture | The protocol does not automatically ingest chats, prompts, responses, terminal scrollback, environment variables, or arbitrary working-tree contents. | Privacy | High | Open |
| C-008 | Derived indexes only | Lexical indexes, embeddings, summaries, rankings, and tokenized prompt views are disposable local projections and never canonical protocol state. | Architecture | High | Open |
| C-009 | No erasure guarantee | Retraction and local exclusion do not promise removal from already replicated Git objects; redaction and legal deletion semantics remain separate work. | Product | High | Open |
| C-010 | Explicit semantic limits | Natural-language contradiction, factual truth, confidence calibration, and relevance cannot be declared solved by protocol validation; only explicit relationships and deterministic filters are canonical. | Product | High | Open |
| C-011 | Repository scope | Version 0 memory is scoped to one Git repository; cross-repository identity and memory federation remain outside this mission. | Scope | Medium | Open |
| C-012 | No autonomous authority escalation | Recording or recalling a decision, handoff, or verified memory grants no permission to change code, run tools, publish refs, approve proposals, or amend policy. | Governance | High | Open |

### Key Entities

- **Memory Record**: Immutable signed cognition with type, actor, content,
  anchors, applicability, topics, evidence, and digest.
- **Memory Stream**: Separately replicated append-only sequence signed by one
  actor and selected independently from collaboration history.
- **Memory Anchor**: Exact commit plus optional path/blob or full Nichthub
  subject identity locating the state a memory concerns.
- **Applicability Rule**: Explicit exact-state, descendant-state, or
  subject-lifecycle scope used by recall.
- **Memory Lifecycle Fact**: Same-author supersession/retraction or cross-actor
  challenge referencing exact predecessor memory IDs.
- **Memory Projection**: Deterministic active and historical state derived from
  accepted streams, explicit lifecycle edges, evidence availability, and policy.
- **Recall Policy**: Project rules classifying actor/memory-kind combinations
  for default recall; it never changes signature validity or instruction safety.
- **Recall Envelope**: Bounded machine-readable collection of inert memory
  records with provenance, trust, lifecycle, evidence, and truncation metadata.
- **Memory Index**: Local rebuildable exact/lexical or optional semantic view
  keyed by accepted memory-ref state and applicable policy digest.

## Scope Boundaries

### In Scope

- Structured observation, decision, assumption, attempt, verification, and
  handoff memories.
- Exact Git/Nichthub anchors, explicit applicability, topics, and evidence.
- Separate signed memory streams with selected, quarantined, bounded
  replication and shallow dependency recovery.
- Same-author supersession/retraction and cross-actor challenge projections.
- Policy-filtered bounded recall with complete provenance and inert content.
- Deterministic exact and lexical query plus rebuildable local indexes.
- Versioned machine-readable record/recall interfaces and two neutral adapter
  fixtures.
- Public documentation, threat model, interoperability fixtures, and a
  multi-clone agent handoff demonstration in this repository.

### Out of Scope

- Automatic capture or long-term storage of complete agent conversations.
- Automatic semantic truth, contradiction, relevance, or confidence judgment.
- Treating memory as authorization or executable instruction.
- Mandatory embeddings, model inference, vector databases, hosted memory
  services, or vendor-specific prompt formats.
- Cross-repository memory federation or personal memory spanning unrelated
  projects.
- Guaranteed erasure, shared redaction, legal retention, or content-moderation
  policy.
- Agent model training, behavioral profiling, private personal notes, web UI,
  or global memory discovery.
- Autonomous execution of handoff next steps.

## Assumptions and Dependencies

- `Operational Self-Hosting Alpha` lands first and its identity, policy,
  selection, quarantine, budget, and shallow-recovery contracts remain stable
  enough to extend.
- Agent processes use distinct actor identities and do not share one private
  key across concurrent sessions or devices.
- Memory replication has its own roots while reusing actor signatures and the
  accepted/quarantine transport boundary.
- Project policy can add memory-specific actor/kind classifications without
  making memory content trusted instructions.
- Every protocol record is potentially public once published; operators keep
  secrets and sensitive transcript content out of explicit memory inputs.
- Exact and lexical retrieval are the interoperable baseline. Semantic search
  is optional and may vary without changing canonical results.
- A human or agent adapter chooses when to record and when to request recall;
  Nichthub does not monitor sessions automatically.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Two independently initialized agent actors exchange a reference
  corpus containing every supported memory kind through an ordinary Git remote
  and reconstruct 100% identical IDs, anchors, content, and evidence.
- **SC-002**: A later agent in a fresh clone retrieves a bounded handoff for the
  current commit and correctly identifies completed work, active assumptions,
  blockers, and proposed next actions with full provenance and no copied local
  index or vendor state.
- **SC-003**: Same-author supersession/retraction and cross-actor challenges
  delivered in three different orders yield byte-identical lifecycle fixtures
  and no erased historical memory.
- **SC-004**: An adversarial corpus containing instruction-like content,
  controls, invalid signatures, oversized records, and untrusted actors causes
  zero tool executions, authority changes, trusted-instruction outputs, or
  collaboration projection failures.
- **SC-005**: A participant can synchronize collaboration refs while importing
  zero memory refs, and all pre-memory public event IDs and projections remain
  unchanged.
- **SC-006**: With 10,000 accepted records, clean index rebuild and p95 local
  recall meet NFR-008 while two rebuilds agree on 100% of canonical membership
  and lifecycle state.
- **SC-007**: Two vendor-neutral test adapters produce and consume the same
  versioned record and recall contracts without vendor fields in canonical
  memory.
- **SC-008**: The full demonstration uses zero mandatory services, model APIs,
  hosting-provider collaboration APIs, shared private keys, and Docker
  operations.
- **SC-009**: Protocol, safety, operator, and adapter documentation explicitly
  distinguishes signature validity, policy trust, evidence resolution,
  semantic truth, and prompt authority with no unresolved ambiguity.
