# Research: Proposal Revision and Conflict Recovery

## Repository Findings

Nichthub already has the primitives needed for immutable recovery:

- every actor publishes one exact-byte-signed, append-only event chain;
- `proposal.open` binds an exact base and head;
- an immutable proposal ref keeps each signed head reachable through Git;
- reviews, run requests/results, decisions, and merge events bind exact event
  IDs and policy or code inputs; and
- a failed Git merge is automatically aborted, restoring the required clean
  worktree boundary.

The missing piece is not storage or transport. It is an explicit successor fact
plus a deterministic projection and governance gates over that fact.

## Decision 1: Add `proposal.revise` as a distinct event kind

**Decision**: A revision is a new event kind. It reuses `subject` for the exact
predecessor, and carries fresh `base`, `head`, and optional `body`. It inherits
the predecessor title for display.

**Rationale**: A distinct kind makes compatibility fail closed. A client that
does not understand revisions rejects the unknown signed event rather than
silently projecting it as an unrelated open proposal and potentially accepting
a predecessor that a newer client knows is superseded. Reusing existing fields
keeps the wire change small and preserves exact-byte signing.

**Alternatives rejected**:

- Add optional `predecessor` to `proposal.open`: old clients would ignore the
  semantic relationship, which is unsafe for acceptance and merge.
- Store lineage only in refs: refs are mutable transport pointers and are not
  signed protocol facts.
- Mutate the original proposal: violates content addressing and signed-history
  immutability.

## Decision 2: Preserve sibling graph semantics within the actor-chain boundary

**Decision**: Multiple valid revisions may name the same predecessor and remain
sibling candidates. Their publication still follows the proposal author's one
linear actor chain.

**Rationale**: The actor ref and signed `previous`/`sequence` chain currently
enforce a single writer per identity. Two disconnected clones using the same
private key would create conflicting sequence numbers and non-fast-forward ref
updates. Solving that is a different protocol problem involving multi-device
identity append semantics. This mission needs sibling graph semantics, not an
actor-history redesign: an author may publish revision A and later revision B,
both naming the same predecessor, and peers derive the same sibling set
regardless of presentation order.

**Alternatives rejected**:

- Permit forked actor chains: weakens a foundational validation invariant and
  changes synchronization conflict semantics throughout the protocol.
- Assign another signer as revision author: violates author ownership and lets
  an unrelated actor supersede someone else's proposal.
- Pick the first or newest sibling: invents central ordering from local arrival
  time or untrusted timestamps.

## Decision 3: Derive lineage; do not persist a current/latest pointer

**Decision**: Build a lineage index from the full locally verified event set.
The index maps proposals, predecessor edges, sorted successor sets, roots,
lineage merge facts, and derived states.

**Rationale**: A derived projection converges for the same verified inputs and
cannot overwrite history. IDs, not timestamps or arrival order, provide stable
display ordering. A single module can own cycle-safe traversal and prevent
slightly different lineage rules from appearing in CLI, policy, CI, and merge
code.

**Alternatives rejected**:

- Mutable `latest` ref: loses siblings and creates a contested writer.
- Store a materialized index event: duplicates derivable state and creates new
  invalidation/consensus problems.
- Recompute ad hoc in each command: spreads security-sensitive invariants and
  increases drift risk.

## Decision 4: Keep evidence exact and fresh by construction

**Decision**: Treat both `proposal.open` and `proposal.revise` as proposal
candidates, while leaving evidence selection keyed by exact proposal ID. Load
policy and pipeline definitions from the revision's own base/head.

**Rationale**: Existing review, run, decision, and merge events already bind an
exact subject. Generalizing proposal-kind checks is sufficient; no inheritance
or evidence-copy mechanism should exist. A revision begins with zero qualifying
reviews, results, or decisions even if its predecessor was ready or accepted.

**Alternatives rejected**:

- Carry approvals forward: the signed code and possibly the governing policy
  changed, so old evidence answers a different question.
- Copy evidence events with a new subject: reinterprets signatures and would be
  cryptographically false.

## Decision 5: Block only unsafe new terminal actions

**Decision**: A known successor blocks a new acceptance or merge of its
predecessor. Local revision creation refuses a predecessor already known merged.
A known merge anywhere in a lineage blocks new acceptance or merge of every
other unmerged member. Rejections and historical evidence remain visible. If a
remote revision and merge, or multiple merge facts, are later observed together,
preserve them and report the derived closed/conflicted lineage state.

**Rationale**: These gates prevent a locally stale terminal action while
preserving immutable facts created under a prior view. They also respect the
distributed boundary: a peer cannot react to an event it has not received.

**Alternatives rejected**:

- Reject the whole event history when two siblings were independently merged:
  both signed facts may be valid products of disconnected local knowledge.
- Retroactively invalidate or remove a merge: signed history is immutable and
  Nichthub cannot undo Git history.
- Reject an otherwise valid received revision merely because a merge is also
  present: actor-local chains, untrusted timestamps, and delivery order cannot
  prove which fact was created first.
- Block all reviews/runs on superseded candidates: the specification requires
  blocking acceptance and merge; historical investigation need not be erased.

## Decision 6: Reuse proposal refs and synchronization unchanged

**Decision**: A revision receives the existing
`refs/nh/proposals/<revision-event-id>` code ref, and `nh sync` continues to
fetch/push the existing actor and proposal wildcard refspecs.

**Rationale**: The ref namespace is already per immutable proposal identity.
No mutable lineage ref or server support is required, and ordinary Git remotes
continue to transport both events and code.

## Security and Performance Notes

- Relationship validation must run after all locally reachable events have been
  verified, so result is independent of event iteration order.
- A revision is trusted only if its predecessor exists, is a proposal candidate,
  has the same actor, and introduces no cycle. Known merge state governs local
  creation and future acceptance/merge, not retrospective fact deletion.
- Lineage traversal must track visited IDs, return actionable exact IDs, and
  operate in O(events + lineage edges) time.
- Unknown event kinds continue to fail closed under protocol `nh/0`.
- No supply-chain research is needed: the design adds no dependency.

## Evidence Trail and Open Questions

Repository evidence is indexed in `research/source-register.csv` and individual
findings are mapped to decisions in `research/evidence-log.csv`. No external
source was required because the mission extends only local experimental
protocol behavior and adds no dependency.

There are no unresolved implementation questions. Multi-device append for one
private identity remains an explicit future protocol problem, not a hidden
dependency of sibling revision semantics.
