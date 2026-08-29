# Nichthub Project Charter

This is the human-readable companion to the authoritative runtime charter in
`.kittify/charter/charter.yaml`.

## Purpose

Nichthub supplies the collaboration layer Git deliberately does not: portable
issues, proposals, reviews, policy decisions, CI requests and attestations. Its
central promise is that this state travels with a repository, remains signed
and independently verifiable, and does not depend on a particular forge.

## Engineering constraints

- Implement the native `nh` CLI in Go, preferring the standard library and
  narrow, explicit interfaces.
- Keep Git as the content store and transport. Do not introduce a mandatory
  service, service database, GitHub workflow, or Docker dependency.
- Treat all fetched collaboration data and repository-defined execution as
  hostile input. Trust is explicit policy, never ambient authority.
- Signed history is immutable. Corrections and conflict resolutions create
  linked successor events rather than rewriting what participants signed.
- Observable protocol behavior is specified with examples and tested through
  public CLI and repository boundaries, including real temporary remotes.

## Quality gates

Before review, Go changes pass formatting checks, `go test -race ./...`,
`go vet ./...`, `go build ./...`, and `git diff --check`. Wire-format, trust,
runner-isolation, policy, or recovery changes update the relevant documentation
in the same mission.

Every work package is reviewed before approval. A mission uses a focused branch
and Spec Kitty execution lanes. As Nichthub gains the necessary capabilities,
we also dogfood its signed proposal, review, CI, decision, and merge path; that
demonstrates interoperability, but self-approval is not independent trust.

## Current product boundary

Core collaboration requires only Git. Hardened Linux pipeline execution may
require Bubblewrap. Unsafe host execution remains an explicit per-invocation
opt-in. Correctness, deterministic evidence, and bounded handling of untrusted
data take priority over premature performance optimization.

The first governed mission is proposal revision and conflict recovery: preserve
the original proposal, link it to a new resolved head, and reevaluate all
approval and CI evidence against that exact revision.
