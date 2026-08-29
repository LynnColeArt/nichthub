# Nichthub

Nichthub is an experiment in distributing collaboration with a Git repository.
Git distributes the work; Nichthub distributes the intent, discussion, and
evidence around it.

The current prototype proves one narrow claim: two people can exchange signed
issues, proposals, reviews, and CI attestations—including proposed Git commits
and verified run logs—through an ordinary Git remote without a Nichthub server.
It is not yet a stable or secure protocol.

## Build

Go 1.26 or newer is currently used for development.

```sh
go build -o nh .
go test ./...
```

## Try it

Inside a Git repository:

```sh
nh init --name Alice
nh issue open --body "The description" "An issue title"
nh sync
```

In another clone:

```sh
nh init --name Bob
nh sync
nh issue list
nh issue comment <issue-id> "I can see it."
nh sync
```

Proposals bind signed collaboration events to Git commits and publish a
content-bearing proposal ref so the proposed code is fetched with the event:

```sh
nh proposal open --base main --head feature \
  --body "Please review this change." "Add the feature"
nh sync

# In a reviewer's clone:
nh sync
nh proposal list
nh proposal show <proposal-id>
nh review <proposal-id> --approve --body "The fetched code looks good."
nh sync
```

Pipelines are JSON files stored with the proposed code. A step can invoke an
installed tool or a custom executable tracked in the repository:

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

Requesting and executing a run are separate signed actions:

```sh
# Requester
nh run request <proposal-id> test
nh sync

# Runner
nh sync
nh run list
nh run execute <request-id>
nh sync

# Requester
nh sync
nh run show <request-id>
nh run logs <result-id>
```

On Linux, the default executor uses Bubblewrap with separate user, PID,
network, IPC, UTS, and cgroup namespaces; dropped capabilities; no host home;
read-only system tools; and a writable generated checkout. `bwrap` must be
installed. The sandbox has no external network access.

The unsafe host backend remains available for debugging and unsupported
platforms, but requires both flags on every invocation:

```sh
nh run execute <request-id> \
  --backend host \
  --allow-unsafe-host-execution
```

A runner can discover work automatically while applying a narrow local
acceptance policy. Both the pipeline and the full request-signer fingerprint
are mandatory:

```sh
nh runner once \
  --accept-pipeline test \
  --accept-actor <full-actor-fingerprint>

nh runner watch \
  --accept-pipeline test \
  --accept-actor <full-actor-fingerprint> \
  --interval 30s
```

`runner once` executes at most one pending matching request. `runner watch`
continues synchronizing and processing matching requests until interrupted.
Requests from every other signer and for every other pipeline are ignored.

Project governance lives in `.nh/policy.json`. A proposal is always evaluated
against the exact policy bytes in its signed base commit, so a proposal cannot
weaken the rules used to accept itself:

```sh
nh proposal status <proposal-id>
nh decide <proposal-id> --accept

# On the target branch, with a clean worktree:
nh merge <proposal-id>
```

Accept decisions sign the policy digest and the exact review and CI evidence
that satisfied it. Merge events sign the accepted proposal head, resulting Git
commit, policy digest, and acceptance decisions.

Back in the first clone:

```sh
nh sync
nh issue show <issue-id>
```

`nh init` creates a local Ed25519 identity under `.git/nh/`. The private key is
not part of the repository and is not cloned. Events are exact-byte signed,
stored as Git objects, and connected through one append-only ref per actor:

```text
refs/nh/actors/<actor-key-fingerprint>
```

`nh sync` fetches these refs into a remote-tracking namespace and publishes the
current actor's ref. Git is the storage and transport layer; the remote does not
need Nichthub-specific software.

## Prototype commands

```text
nh init [--name NAME]
nh identity show
nh issue open [--body TEXT] TITLE
nh issue comment ISSUE [--body TEXT] [TEXT]
nh issue list
nh issue show ISSUE
nh proposal open --base REV --head REV [--body TEXT] TITLE
nh proposal list
nh proposal show PROPOSAL
nh proposal status PROPOSAL
nh review PROPOSAL <--approve|--request-changes> [--body TEXT]
nh run request PROPOSAL PIPELINE
nh run list
nh run show REQUEST
nh run execute REQUEST [--backend sandbox|host] [--rerun]
nh run logs RESULT
nh runner once --accept-pipeline NAME --accept-actor ACTOR
nh runner watch --accept-pipeline NAME --accept-actor ACTOR
nh decide PROPOSAL <--accept|--reject> [--body TEXT]
nh merge PROPOSAL
nh sync [REMOTE]
nh log
```

## Scope of this experiment

The prototype deliberately omits policy amendment tooling, merge queues,
portable/container backends, secrets, configurable network permissions, strong
CPU/memory/disk quotas, key rotation, moderation, selective replication,
redaction, shallow-clone handling, and multiple writers using the same
identity. Those should only be added after the repository-native event model
survives testing.

The draft wire/storage format is described in
[`docs/protocol-v0.md`](docs/protocol-v0.md).

Live Git transport results are recorded in
[`docs/host-compatibility.md`](docs/host-compatibility.md).

The current CI and runner threat model is documented in
[`docs/ci-v0.md`](docs/ci-v0.md).

Policy evaluation and decision semantics are documented in
[`docs/governance-v0.md`](docs/governance-v0.md).
