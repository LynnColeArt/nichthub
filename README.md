# Nichthub

Nichthub is an experiment in distributing collaboration with a Git repository.
Git distributes the work; Nichthub distributes the signed intent, discussion,
policy, and evidence around it. There is no Nichthub service or database: an
ordinary Git remote transports both code and `refs/nh/*` collaboration facts.

The current operational alpha supports signed issues, immutable proposal
candidates and revisions, reviews, CI requests/results/logs, governance
decisions and merge facts, distinct device actors, policy amendments, selected
quarantined replication, exact shallow-history recovery, and deliberate signed
agent memory. It remains an experimental protocol, not a stable or hardened
multi-tenant system.

## Project status

The initial operational-alpha roadmap is complete. Nichthub has used its own
published protocol to revise, independently sandbox-test, approve, accept, and
merge changes into public `main`. An ordinary Git remote carried the code and
signed collaboration refs; it did not supply Nichthub-specific governance or
CI services.

"Alpha complete" means the end-to-end protocol boundary documented here is
implemented and exercised. It does not mean production-ready: compatibility is
not yet stable, the security boundary is intentionally narrow, and the
deferred work in [Completed alpha boundary](#completed-alpha-boundary) is the
post-alpha roadmap.

## Build and verify

Go 1.26 or newer and Git 2.x are the baseline requirements. Docker is not
required.

```sh
go build -trimpath -o nh .
go test -count=1 ./...
```

The two end-to-end acceptance scenarios compile the CLI and exercise the
self-hosted governance/CI loop and repository-native agent memory through real
Git repositories:

```sh
go test -count=1 -run 'TestOperational(SelfHostingAlpha|AgentMemory)$' ./...
```

Run the full local release checks with:

```sh
go test -race -count=1 ./...
go vet ./...
git diff --check
```

Bubblewrap is used by the default Linux CI executor when you choose to execute
a pipeline. It is not required for identity, policy, proposal, review,
replication, inspection, or verification commands. The explicit unsafe host
executor remains available for controlled debugging.

## Start a repository

```sh
nh init --name "Alice's device"
nh identity public
nh issue open --body "The description" "An issue title"
nh sync origin
```

`nh init` creates a local Ed25519 identity below `.git/nh/`. Private keys and
local replication selections are never Git objects and are not cloned. Public
events are exact-byte signed and stored in one append-only Git history per
actor:

```text
refs/nh/actors/<full-actor-fingerprint>
```

## Distinct device actors

Initialize another clone independently, then exchange only the output of
`nh identity public`. The existing actor authorizes the new public key and the
new actor accepts that exact authorization:

```sh
nh identity authorize --relationship device \
  --actor <full-new-actor> \
  --public-key <new-public-key>
nh sync origin

# In the independently initialized target clone, after selected sync:
nh identity accept <full-authorization-event-id>
nh sync origin
nh identity list
```

The relationship is descriptive. It does not add the new actor to any policy
role. `nh identity rotate` performs the same two-sided protocol locally with a
`successor` relationship and retryable transaction state; it is planned
rotation while the predecessor key is available, not lost-key recovery.

## Inspect and amend policy

Project governance lives in `.nh/policy.json`. Inspection prints complete
trust-bearing actor IDs. An amendment is an ordinary candidate, and only the
exact policy bytes from its signed base govern it:

```sh
nh policy show main
nh policy check --base main --file .nh/policy.json
nh policy check --base main --head HEAD

nh proposal open --base main --head HEAD \
  --body "The base policy governs this amendment." \
  "Amend collaboration policy"
```

Continuity relationships and replication selections never grant maintainer,
reviewer, runner, or decision authority. Only the explicit full actor lists in
the candidate's base policy do.

## Select and synchronize facts

Save exact full actor/candidate selections and positive local budgets before
synchronizing an unfamiliar remote:

```sh
nh replication select origin \
  --actor <full-maintainer-actor> \
  --actor <full-reviewer-actor> \
  --proposal <full-candidate-event-id> \
  --max-events 10000 \
  --max-objects 100000 \
  --max-object-bytes 16777216 \
  --max-attachment-bytes 1048576 \
  --max-total-bytes 268435456
nh replication show origin
nh sync origin
```

Each selected ref is fetched into a separate bare quarantine repository,
measured and verified, then promoted to `refs/nh/remotes/<remote>/*` in one
atomic ref transaction. Independently valid selections may promote when
another selected history fails. Standard Git can download a selected pack
before Nichthub measures it, so these are hard validation, promotion, and
retention limits—not portable pre-download network quotas.

A trust-sensitive command in a depth-limited clone reports a full missing ID
and exact recovery action. Recovery is explicit and uses the saved selection:

```sh
nh sync origin --recover-shallow
```

This never performs a global unshallow and never silently adds a selector.

## Record and recall agent memory

Nichthub can carry deliberate project cognition without a hosted memory
service. `nh-memory/0` is separate from the collaboration protocol: each actor
owns append-only memory streams under `refs/nh/memory/*`, while private keys,
replication selections, and the rebuildable lexical index stay below
`.git/nh/` and are never cloned.

```sh
nh memory record --kind decision --at HEAD --applies descendants \
  --topic architecture --evidence git:$(git rev-parse HEAD) \
  --content "Keep memory streams independent from actor event chains."

nh memory handoff --at HEAD --applies descendants --input handoff.json --json
nh memory supersede sha256:<full-memory-id> \
  --kind decision --at HEAD --applies descendants \
  --content "Replacement decision with current rationale."

nh replication select origin --memory sha256:<full-stream-id>
nh sync origin
nh memory index rebuild
nh memory recall --at HEAD --topic architecture --json
```

Default recall is policy-qualified, active, local-only, and bounded to 20
records and 65,536 encoded content bytes. Every item retains its full IDs,
anchor, lifecycle edges, evidence, applicability, signature, trust class, and
content digest. Memory content and handoff next actions are untrusted inert
data—not instructions, truth, policy authority, or permission to act.

Add sorted `memory.trustedActors` and `memory.trustedKinds` to the exact commit's
`.nh/policy.json` to qualify default recall. Without it, valid signed memory is
explicitly inspectable but not trusted by default. Retraction preserves an
auditable fact; neither retraction nor deleting the private index erases
replicated Git objects.

## Proposals, CI, decisions, and merge

```sh
nh proposal open --base main --head feature \
  --body "Please review this exact change." "Add the feature"
nh run request <full-candidate-event-id> test
nh sync origin

# In a selected participant clone:
nh run execute <full-run-request-event-id>
nh review <full-candidate-event-id> --approve \
  --body "The fetched candidate is correct."
nh sync origin

# In a maintainer clone:
nh proposal status <full-candidate-event-id>
nh decide <full-candidate-event-id> --accept
nh merge <full-candidate-event-id>
nh sync origin                    # publish collaboration refs
git push origin main:main         # publish the primary branch separately
```

Pipeline definitions are repository JSON files. A step can invoke an installed
tool or an executable tracked in the candidate, without a shell unless the
command itself is a shell:

```json
{
  "version": "nh.pipeline/0",
  "steps": [
    {
      "name": "Tests",
      "command": "./.nh/actions/test",
      "args": ["--all"],
      "timeoutSeconds": 300
    }
  ]
}
```

On Linux, the default executor uses Bubblewrap with separate user, PID,
network, IPC, UTS, and cgroup namespaces; dropped capabilities; no host home;
read-only system tools; and a writable generated checkout. It exposes no
external network interface. The unsafe host backend requires both flags on
every invocation:

```sh
nh run execute <full-run-request-event-id> \
  --backend host \
  --allow-unsafe-host-execution
```

If Git reports a merge conflict, resolve it outside Nichthub and publish a new
immutable revision. The predecessor and all prior evidence remain unchanged;
the revision needs its own exact review, CI, and acceptance evidence:

```sh
nh proposal revise <full-predecessor-candidate-id> \
  --base <full-resolved-base-commit> \
  --head <full-resolved-head-commit> \
  --body "Resolve the merge conflict"
nh sync origin
```

## Command surface

```text
nh init [--name NAME]
nh identity show|list|public|authorize|accept|rotate
nh issue open|comment|list|show
nh proposal open|revise|list|show|status
nh policy show [REV]
nh policy check --base REV <--head REV|--file PATH>
nh review PROPOSAL <--approve|--request-changes> [--body TEXT]
nh run request|list|show|execute|logs
nh runner once|watch --accept-pipeline NAME --accept-actor ACTOR
nh decide PROPOSAL <--accept|--reject> [--body TEXT]
nh merge PROPOSAL
nh memory record|handoff|supersede|retract|challenge|show|recall|index
nh replication select|show [REMOTE] [--memory STREAM]...
nh sync [REMOTE] [--recover-shallow]
nh log
```

Trust-bearing commands require full actor fingerprints and full
`sha256:<64-hex>` event IDs. Short IDs are display conveniences only.

## Completed alpha boundary

The operational-alpha column is implemented. The deferred column describes
the post-alpha roadmap; those capabilities are not implied by alpha completion.

| Area | Operational alpha | Explicitly deferred |
| --- | --- | --- |
| Policy | Show, validate, diff, and govern amendments from exact base bytes | Merge queues and implicit role migration |
| Identity | Distinct actors, mutual device/successor facts, local keyring, retryable planned rotation | Lost-key, compromise, social/organizational recovery, or concurrent writers sharing one actor key |
| Replication | Exact selections, quarantine, positive budgets, validation, atomic accepted refs, compatibility-all | Portable hard pre-download quotas, moderation, selective deletion, and global redaction |
| CI | Repository-defined actions, default Bubblewrap runner, explicit unsafe host fallback | Secrets, configurable network access, strong CPU/memory/disk/process quotas, and portable/container backends |
| Memory | Signed streams, exact anchors/evidence, lifecycle projection, bounded lexical recall, disposable local index, and selected transport | Automatic capture, embeddings, semantic truth, autonomous action, federation, moderation, redaction, retention enforcement, and erasure guarantees |
| Product | CLI and Git-native signed facts | Notifications, general search, web UI, discovery, and a stable protocol compatibility promise |

Published immutable facts may be superseded by new facts, but the alpha cannot
promise global erasure from every replica.

## Documentation

- [Protocol and storage](docs/protocol-v0.md)
- [Agent Memory Protocol and operator guide](docs/memory-v0.md)
- [Agent memory safety model](docs/memory-safety.md)
- [Identity continuity and keyring safety](docs/identity-v0.md)
- [Governance and policy amendments](docs/governance-v0.md)
- [Selected replication and shallow recovery](docs/replication-v0.md)
- [CI and runner threat model](docs/ci-v0.md)
- [Hosted Git compatibility](docs/host-compatibility.md)
- [Operational self-hosting proof](docs/self-hosting-alpha.md)
