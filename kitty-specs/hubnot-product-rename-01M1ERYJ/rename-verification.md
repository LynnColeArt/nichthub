# Hubnot Rename Verification

## Verification snapshot

- Evidence captured: 2026-09-01T16:15:53Z
- Baseline: 7dcfbc1264d3e08db6b4cbbe5ccd29eb73219205
- Integrated candidate: ee1e4876b7c953ce6551f565e0ec2bd30d9f143b
- Governed proposal head: ededa84683caa1dfeed028e5938e61c00e2b588b
- Governed public merge: ec2ded35804d781d0d8338fc593d191e9b8fe8eb
- Shared renamed content tree: 89e2e4eef21bf22193a6f2b8984eb851880335c5
- Toolchain: go1.26.4 linux/amd64; Git 2.43.0
- Public URL: https://github.com/LynnColeArt/hubnot.git

WP01 and WP02 were independently reviewed. Because Spec Kitty retains approved
lanes until final mission merge, aggregate checks used a disposable clone based
on the public baseline with both approved commits applied. The governed proposal
and public merge have the identical Git tree above, so all gates evaluated the
same product content even where commit ancestry differed.

## Occurrence classification

The authoritative occurrence_map.yaml renames active product, module, and
runtime identity while preserving nh compatibility identifiers and immutable
historical records.

| Surface | Query/result | Classification | Verdict |
| --- | --- | --- | --- |
| Entire public merge | git grep -n -i 'nichthub' ec2ded35804d -- .: 29 lines in 10 files | Historical exceptions only | Pass |
| Canonical initialization journal | 1 line | Append-only initialization event | Preserved |
| Completed proposal-revision mission | 28 lines | Historical evidence and literal paths/slugs | Preserved |
| Active Go and go.mod | 0 lines | Current product identity | Renamed |
| README and docs/** | 0 lines | Current user-facing prose | Renamed |
| Active config/charter | 0 lines | Current metadata/prose | Renamed |
| Former hosted URL in active surfaces | 0 lines | Current repository location | Renamed |
| Module and CLI | go list -m → hubnot; help usage → nh | Brand vs compatibility | Pass |

All 29 remaining occurrences are in .kittify/canonical-events.jsonl or
kitty-specs/proposal-revision-conflict-recovery-01M1774Q/**, both explicit
do_not_change exceptions. There are zero unclassified active occurrences.
The existing review actor's historical signed display label is likewise not
rewritten merely to normalize old branding.

## Frozen namespace inventory

Counts use git grep -I -F -o LITERAL REVISION -- . piped to wc -l.

| Literal | Baseline | Candidate | Verdict |
| --- | ---: | ---: | --- |
| refs/nh/ | 187 | 187 | Unchanged |
| .git/nh/ | 12 | 12 | Unchanged |
| .nh | 104 | 104 | Unchanged |
| nh/0 | 13 | 13 | Unchanged |
| nh.pipeline/0 | 11 | 11 | Unchanged |
| nh.policy/0 | 28 | 28 | Unchanged |
| nh-memory/0 | 12 | 12 | Unchanged |

The CLI remains nh; existing NH_* controls and golden signed payloads were not
renamed. Documentation now uses runner nh/VERSION; a real sandbox result
emitted runner nh/0.0.1-dev.

## Immutable evidence

Protected baseline journals were enumerated with git ls-tree -r --name-only and
compared by exact blob ID.

| Protected path | Baseline blob | Public blob | Verdict |
| --- | --- | --- | --- |
| .kittify/canonical-events.jsonl | e63bdf8a7444fcbe203290a13c4a3245fc79cab5 | same | Byte-identical |
| kitty-specs/proposal-revision-conflict-recovery-01M1774Q/status.events.jsonl | 1e1ec8a50fcd45101ae89995620cbdc88b37e502 | same | Byte-identical |

git diff --name-only 7dcfbc1264d3..ec2ded35804d over the protected paths
emitted nothing.

Signed-event manifests were built by walking accepted actor tips, hashing exact
event.json bytes, sorting full IDs, and deduplicating. The pre-cutover manifest
contains 43 events with digest
e3731a09a86e6d67c1a127d959d0ff073e7e686d7b794f4bf30644e440dff958.
The post-cutover manifest contains 49 events with digest
a57a3527b97025e578658011fc78dfdf48f0c6923e4d9eb2f11a5871fabdbdc6.
comm -23 baseline post found zero missing IDs. Supported nh commands
revalidated signatures and chains before trusting the facts.

| Actor history | Before | After | merge-base --is-ancestor |
| --- | --- | --- | --- |
| Reviewer | 3188776773f9a4607c3d18e8e0bc686b48b122e1 | aa52d04331cbb838d74b0e0ef280f90b0afb4959 | Pass |
| Maintainer | 43e26dae4354451b7c471d36338bacae9d0f9877 | aab6b0d93cbda921952bbc0f434aa5fe04d29460 | Pass |
| Memory author | 3f1257bd84ef354390b5f82d25c056663976bc3e | same | Pass |
| Memory successor | c7700147f216d7edd474b33364a6ea5415c965b5 | same | Pass |

No accepted actor history was deleted, rewound, force-updated, forked, or
replaced.

## Release gates

All local gates ran at integrated candidate
ee1e4876b7c953ce6551f565e0ec2bd30d9f143b with generated binaries outside the
repository.

| Gate | Isolation | Exit/result |
| --- | --- | --- |
| gofmt -l . | Candidate checkout | 0; no output |
| git diff --check | Candidate checkout | 0 |
| go vet ./... | Go 1.26.4 | 0 |
| go build ./... | Go 1.26.4 | 0 |
| go test -count=1 ./... | GIT_CONFIG_GLOBAL=/dev/null and GIT_CONFIG_NOSYSTEM=1 | 0; ok hubnot 67.128s |
| go test -race -count=1 ./... | Same Git isolation | 0; ok hubnot 147.671s |
| git status --short --untracked-files=all | After gates | Empty |

The full suites include compiled black-box TestOperationalSelfHostingAlpha and
TestOperationalAgentMemory; both build the binary and use real Git repositories
through public CLI boundaries.

Bubblewrap was present at /usr/bin/bwrap. An isolated verifier opened proposal
sha256:6d1666629b23d2ff51aea1ff026ad63d9a491208bb752ef1700d88e6a6ac8e23
for the exact candidate and requested test as
sha256:d3fa1fbad10bab9dc6b12b5b2f98b738a49d1517952b2faa5631d62948170500.
nh run execute recorded passed sandbox result
sha256:058f7b2dd170a0ffca37ee67154b34bae8a321cfc6a86a024e2593c04c01828e:
exit 0, 66,813 ms, linux/amd64, runner nh/0.0.1-dev, isolated network
namespace. The unsafe host backend was not used.

## Governed public cutover

The clean proposal branch was rebuilt from public main using only the two
approved product commits. Its tree exactly matched the tested candidate.

| Fact | Full ID / commit |
| --- | --- |
| Proposal | sha256:f242029579a76c9a5d9b8407b4410a047bc724f0431aaad7bb4bbed5e72d96a4 |
| Run request | sha256:21f874334998d63d0abbf37b7d02f6916e4aa6d5d01d296a187fcc9a7e1928ed |
| Trusted sandbox result | sha256:3a155da4feed7cab6ede5a479ab5048368e7eb988253f945bb16e9e910f238d5 (exit 0; 67,173 ms) |
| Trusted approval | sha256:72ba03fcd31bfdc36f4c2c0dba4e720bd49f303b9470b16f7bb08fe2249ca0f3 |
| Maintainer acceptance | sha256:414a655e172b2bf6f65d90fed233f8d1a499c4a07633e5d36366a2bc305b4c26 |
| Merge fact | sha256:cc7b861f8d77b145f1e3c712756446af7c880af118e90a718f7d455ceef86719 |
| Public merge | ec2ded35804d781d0d8338fc593d191e9b8fe8eb |

The credential-isolated reviewer could not publish directly. Its verified
fast-forward actor commit was relayed through maintainer FETCH_HEAD without
creating a local foreign actor ref. Before merge, nh proposal status reported
1/1 trusted approvals, 1/1 trusted passes, and 1/1 accepts.

## Credential-free fresh clone

The final clone disabled global/system Git config, prompts, and credential
helpers and used an empty owner-only HOME with the explicit URL
https://github.com/LynnColeArt/hubnot.git.

It checked out ec2ded35804d781d0d8338fc593d191e9b8fe8eb; origin remained the
Hubnot URL. go list -m returned hubnot; build passed; help began “Hubnot
distributes collaboration with a Git repository.” while usage remained nh; and
the two compiled operational scenarios returned ok hubnot 13.245s.
Credential-free ls-remote advertised public main, all four accepted actor refs,
and the exact proposal code ref. No private identity or credential entered the
clone.

## Traceability

| Requirement | Evidence |
| --- | --- |
| FR-004 | Canonical local/public URL and direct credential-free clone |
| FR-007 | Frozen namespace counts and compatibility suites |
| FR-008 | Exact journal blobs, event-manifest inclusion, fast-forward refs |
| NFR-001 | Format, diff, vet, build, uncached, race, and sandbox gates |
| NFR-002 | Equal protocol inventory plus signing/governance/replication/memory tests |
| NFR-003 | Complete historical classification; zero active hits |
| NFR-004 | Anonymous clone, build, branding smoke, compiled scenarios |
| C-001 | No rewrite; old facts remain byte-identical and reachable |
| C-002 | nh command/storage/ref/environment/protocol preserved |
| C-003 | Rename only; QA hardening remains separate |

## Verdict

PASS. Hubnot is the active public product and module identity. Existing
repositories, scripts, signed facts, append-only journals, and protocol/storage
namespaces remain compatible and auditable.
