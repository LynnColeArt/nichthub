package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemoryRefGrammarAndDefaultStream(t *testing.T) {
	identity := deterministicMemoryIdentity()
	actor := identity.Actor
	stream := "sha256:" + strings.Repeat("0", 63) + "1"

	local, err := memoryRef(actor, stream)
	if err != nil {
		t.Fatal(err)
	}
	if want := "refs/hn/memory/" + actor + "/" + strings.TrimPrefix(stream, "sha256:"); local != want {
		t.Fatalf("local ref = %q, want %q", local, want)
	}
	gotActor, gotStream, ok := parseMemoryRef(local)
	if !ok || gotActor != actor || gotStream != stream {
		t.Fatalf("local parse = %q, %q, %v", gotActor, gotStream, ok)
	}

	accepted, err := acceptedMemoryRef("origin.v1", actor, stream)
	if err != nil {
		t.Fatal(err)
	}
	gotRemote, gotActor, gotStream, ok := parseAcceptedMemoryRef(accepted)
	if !ok || gotRemote != "origin.v1" || gotActor != actor || gotStream != stream {
		t.Fatalf("accepted parse = %q, %q, %q, %v", gotRemote, gotActor, gotStream, ok)
	}
	if _, _, ok := parseMemoryRef(accepted); ok {
		t.Fatal("accepted ref parsed as a local append ref")
	}

	goldenInput := []byte("hn-memory-stream-v0\x00" + actor + "\x00default")
	goldenSum := sha256.Sum256(goldenInput)
	wantDefault := "sha256:" + hex.EncodeToString(goldenSum[:])
	if got := defaultMemoryStream(actor); got != wantDefault {
		t.Fatalf("default stream = %q, want %q", got, wantDefault)
	}
	other := testIdentity(t, "Other")
	if defaultMemoryStream(other.Actor) == wantDefault {
		t.Fatal("different actors derived the same representative default stream")
	}
}

func TestMemoryRefRejectsHostileComponents(t *testing.T) {
	actor := deterministicMemoryIdentity().Actor
	stream := fullMemoryID("a")
	badActors := []string{"", actor[:63], strings.ToUpper(actor), actor + "/x", "../" + actor, actor + ".lock", actor + "@{1}", actor + " ", actor + "\n"}
	for _, candidate := range badActors {
		if _, err := memoryRef(candidate, stream); err == nil {
			t.Errorf("accepted hostile actor %q", candidate)
		}
	}
	badStreams := []string{"", strings.TrimPrefix(stream, "sha256:"), "sha256:" + strings.Repeat("a", 63), "sha256:" + strings.Repeat("A", 64), "sha256:sha256:" + strings.Repeat("a", 64), stream + "/x", stream + ".lock", stream + "@{1}", stream + " ", stream + "\n"}
	for _, candidate := range badStreams {
		if _, err := memoryRef(actor, candidate); err == nil {
			t.Errorf("accepted hostile stream %q", candidate)
		}
	}
	badRemotes := []string{"", ".", "..", "../origin", "origin/x", "origin.lock", "origin@{1}", "origin ", "origin\n"}
	for _, candidate := range badRemotes {
		if _, err := acceptedMemoryRef(candidate, actor, stream); err == nil {
			t.Errorf("accepted hostile remote %q", candidate)
		}
	}
	privateMaterial := strings.Repeat("private-token-", 32)
	if _, err := memoryRef(privateMaterial, stream); err == nil || strings.Contains(err.Error(), privateMaterial) {
		t.Fatalf("invalid actor diagnostic exposed private material: %v", err)
	}

	badRefs := []string{
		"refs/hn/memory/" + actor,
		"refs/hn/memory/" + actor + "/" + strings.Repeat("a", 64) + "/extra",
		"refs/hn/memory/" + strings.ToUpper(actor) + "/" + strings.Repeat("a", 64),
		"refs/hn/remotes/origin/memory/" + actor,
		"refs/hn/remotes/origin/memory/" + actor + "/" + strings.Repeat("A", 64),
	}
	for _, ref := range badRefs {
		if _, _, ok := parseMemoryRef(ref); ok {
			t.Errorf("parsed malformed local ref %q", ref)
		}
		if _, _, _, ok := parseAcceptedMemoryRef(ref); ok {
			t.Errorf("parsed malformed accepted ref %q", ref)
		}
	}
}

func TestMemoryStoreAppendLoadAndCAS(t *testing.T) {
	withMemoryRepository(t, func() {
		identity := deterministicMemoryIdentity()
		first := validMemoryEnvelopeFixture(memoryOperationRecord)
		storedFirst, err := appendMemory(first, identity)
		if err != nil {
			t.Fatal(err)
		}
		if storedFirst.ID != memoryID(storedFirst.Payload) || storedFirst.Envelope.Sequence != 1 {
			t.Fatalf("unexpected first stored memory: %#v", storedFirst)
		}
		ref, _ := memoryRef(identity.Actor, first.Stream)
		if got := mustGitText(t, "rev-parse", ref); got != storedFirst.Commit {
			t.Fatalf("ref = %s, want %s", got, storedFirst.Commit)
		}
		entries := strings.Fields(mustGitText(t, "ls-tree", "--name-only", storedFirst.Commit))
		if fmt.Sprint(entries) != "[memory.json signature]" {
			t.Fatalf("tree entries = %v", entries)
		}
		if parents := mustGitText(t, "show", "-s", "--format=%P", storedFirst.Commit); parents != "" {
			t.Fatalf("root parents = %q", parents)
		}

		winner := nextMemoryEnvelope(t, identity, first.Stream, storedFirst, "winner")
		loser := nextMemoryEnvelope(t, identity, first.Stream, storedFirst, "loser")
		storedWinner, err := appendMemoryAtHead(winner, identity, storedFirst.Commit)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := appendMemoryAtHead(loser, identity, storedFirst.Commit); err == nil || !strings.Contains(err.Error(), "reload and retry") {
			t.Fatalf("stale append error = %v", err)
		}
		if got := mustGitText(t, "rev-parse", ref); got != storedWinner.Commit {
			t.Fatalf("stale append changed winner: got %s want %s", got, storedWinner.Commit)
		}
		if parents := mustGitText(t, "show", "-s", "--format=%P", storedWinner.Commit); parents != storedFirst.Commit {
			t.Fatalf("successor parents = %q, want %q", parents, storedFirst.Commit)
		}

		retry := nextMemoryEnvelope(t, identity, first.Stream, storedWinner, "retry")
		storedRetry, err := appendMemory(retry, identity)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := loadStoredMemory(storedRetry.Commit)
		if err != nil || loaded.ID != storedRetry.ID || loaded.Envelope.Previous != storedWinner.ID {
			t.Fatalf("loaded retry = %#v, %v", loaded, err)
		}
		chain, err := loadMemoryStreamAt("", memoryStreamSource{Ref: ref, Actor: identity.Actor, Stream: first.Stream, Head: storedRetry.Commit})
		if err != nil || len(chain) != 3 || chain[0].ID != storedFirst.ID || chain[2].ID != storedRetry.ID {
			t.Fatalf("loaded chain = %#v, %v", chain, err)
		}
	})
}

func TestMemoryStoreCollectionIsCanonicalDeterministicAndIndependent(t *testing.T) {
	withMemoryRepository(t, func() {
		collaborator := testIdentity(t, "Collaborator")
		event := newEvent(collaborator, "issue.open", 1, "")
		event.Title = "Collaboration remains independent"
		storedEvent, err := appendEvent(event, collaborator)
		if err != nil {
			t.Fatal(err)
		}
		beforeEvents, err := collectEvents()
		if err != nil {
			t.Fatal(err)
		}
		if len(beforeEvents) != 1 || beforeEvents[0].ID != storedEvent.ID {
			t.Fatalf("collaboration fixture = %#v", beforeEvents)
		}
		beforeCollaborationRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/actors", "refs/hn/proposals")
		if empty, err := collectMemories(); err != nil || len(empty) != 0 {
			t.Fatalf("collaboration-only memory collection = %#v, %v", empty, err)
		}
		identity := deterministicMemoryIdentity()
		first, err := appendMemory(validMemoryEnvelopeFixture(memoryOperationRecord), identity)
		if err != nil {
			t.Fatal(err)
		}
		ref, _ := memoryRef(identity.Actor, first.Envelope.Stream)
		accepted, _ := acceptedMemoryRef("origin", identity.Actor, first.Envelope.Stream)
		mustGit(t, "update-ref", accepted, first.Commit)

		gitDir := mustGitText(t, "rev-parse", "--absolute-git-dir")
		privateDir := filepath.Join(gitDir, "hn", "memory")
		if err := os.MkdirAll(privateDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(privateDir, "index-v0.json"), first.Payload, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("plausible-memory.json", first.Payload, 0o600); err != nil {
			t.Fatal(err)
		}

		memories, err := collectMemories()
		if err != nil || len(memories) != 1 || memories[0].ID != first.ID {
			t.Fatalf("collected memories = %#v, %v", memories, err)
		}
		mustGit(t, "update-ref", "-d", ref)
		acceptedOnly, err := collectMemories()
		if err != nil || len(acceptedOnly) != 1 || acceptedOnly[0].ID != first.ID {
			t.Fatalf("accepted-only memories = %#v, %v", acceptedOnly, err)
		}
		mustGit(t, "update-ref", ref, first.Commit)
		if got := mustGitText(t, "rev-parse", ref); got != first.Commit {
			t.Fatalf("local ref changed during collection: %s", got)
		}
		afterEvents, err := collectEvents()
		if err != nil || len(afterEvents) != 1 || afterEvents[0].ID != beforeEvents[0].ID || string(afterEvents[0].Payload) != string(beforeEvents[0].Payload) {
			t.Fatalf("memory changed collaboration projection: before=%#v after=%#v err=%v", beforeEvents, afterEvents, err)
		}
		if refs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/actors", "refs/hn/proposals"); refs != beforeCollaborationRefs {
			t.Fatalf("memory changed collaboration refs:\nbefore=%s\nafter=%s", beforeCollaborationRefs, refs)
		}
	})
}

func TestMemoryStoreSupportsSHA256GitObjectsWhenAvailable(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("restore directory: %v", err)
		}
	}()
	if _, err := gitOutput("init", "-q", "--object-format=sha256", "-b", "main"); err != nil {
		t.Skip("installed Git does not support SHA-256 repositories")
	}
	mustGit(t, "config", "user.name", "Test")
	mustGit(t, "config", "user.email", "test@hn.invalid")
	identity := deterministicMemoryIdentity()
	stored, err := appendMemory(validMemoryEnvelopeFixture(memoryOperationRecord), identity)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Commit) != 64 {
		t.Fatalf("SHA-256 Git commit length = %d, want 64", len(stored.Commit))
	}
	loaded, err := collectMemories()
	if err != nil || len(loaded) != 1 || loaded[0].ID != stored.ID {
		t.Fatalf("SHA-256 memory collection = %#v, %v", loaded, err)
	}
}

func TestMemoryStreamRejectsTreeAndChainCorruption(t *testing.T) {
	withMemoryRepository(t, func() {
		identity := deterministicMemoryIdentity()
		first, err := appendMemory(validMemoryEnvelopeFixture(memoryOperationRecord), identity)
		if err != nil {
			t.Fatal(err)
		}
		ref, _ := memoryRef(identity.Actor, first.Envelope.Stream)

		wrongOwner := testIdentity(t, "Wrong owner")
		wrongRef, _ := memoryRef(wrongOwner.Actor, first.Envelope.Stream)
		mustGit(t, "update-ref", wrongRef, first.Commit)
		if _, err := loadMemoryStreamAt("", memoryStreamSource{Ref: wrongRef, Actor: wrongOwner.Actor, Stream: first.Envelope.Stream, Head: first.Commit}); err == nil || !strings.Contains(err.Error(), "owner") {
			t.Fatalf("wrong owner error = %v", err)
		}
		mustGit(t, "update-ref", "-d", wrongRef)

		wrongStream := fullMemoryID("7")
		wrongStreamRef, _ := memoryRef(identity.Actor, wrongStream)
		mustGit(t, "update-ref", wrongStreamRef, first.Commit)
		if _, err := loadMemoryStreamAt("", memoryStreamSource{Ref: wrongStreamRef, Actor: identity.Actor, Stream: wrongStream, Head: first.Commit}); err == nil || !strings.Contains(err.Error(), "stream") {
			t.Fatalf("wrong stream error = %v", err)
		}
		mustGit(t, "update-ref", "-d", wrongStreamRef)

		extraBlob := mustGitTextFromInput(t, []byte("extra"), "hash-object", "-w", "--stdin")
		memoryBlob := mustGitText(t, "rev-parse", first.Commit+":memory.json")
		signatureBlob := mustGitText(t, "rev-parse", first.Commit+":signature")
		treeInput := fmt.Sprintf("100644 blob %s\tmemory.json\n100644 blob %s\tsignature\n100644 blob %s\textra\n", memoryBlob, signatureBlob, extraBlob)
		badTree := mustGitTextFromInput(t, []byte(treeInput), "mktree")
		badCommit := mustGitTextFromInput(t, nil, "commit-tree", badTree, "-m", "bad tree")
		mustGit(t, "update-ref", ref, badCommit, first.Commit)
		if _, err := collectMemories(); err == nil || !strings.Contains(err.Error(), "tree") {
			t.Fatalf("extra tree error = %v", err)
		}
	})
}

func TestMemoryStreamRejectsSignedPreviousAndGitMergeIndependently(t *testing.T) {
	withMemoryRepository(t, func() {
		identity := deterministicMemoryIdentity()
		first, err := appendMemory(validMemoryEnvelopeFixture(memoryOperationRecord), identity)
		if err != nil {
			t.Fatal(err)
		}
		ref, _ := memoryRef(identity.Actor, first.Envelope.Stream)

		badPrevious := nextMemoryEnvelope(t, identity, first.Envelope.Stream, first, "bad previous")
		badPrevious.Previous = fullMemoryID("8")
		payload, signature, err := encodeAndSignMemory(badPrevious, identity)
		if err != nil {
			t.Fatal(err)
		}
		badPreviousCommit := writeRawMemoryCommit(t, payload, signature, []string{first.Commit})
		mustGit(t, "update-ref", ref, badPreviousCommit, first.Commit)
		assertMemoryCollectionRejectedWithoutRefMutation(t, "previous")

		mustGit(t, "update-ref", ref, first.Commit, badPreviousCommit)
		otherRoot := writeRawMemoryCommit(t, first.Payload, first.Signature, nil)
		validNext := nextMemoryEnvelope(t, identity, first.Envelope.Stream, first, "merge")
		payload, signature, err = encodeAndSignMemory(validNext, identity)
		if err != nil {
			t.Fatal(err)
		}
		merge := writeRawMemoryCommit(t, payload, signature, []string{first.Commit, otherRoot})
		mustGit(t, "update-ref", ref, merge, first.Commit)
		assertMemoryCollectionRejectedWithoutRefMutation(t, "parent")
	})
}

func TestMemoryStoreRejectsInvalidSignatureWithoutRefMutation(t *testing.T) {
	withMemoryRepository(t, func() {
		identity := deterministicMemoryIdentity()
		envelope := validMemoryEnvelopeFixture(memoryOperationRecord)
		payload, _, err := encodeAndSignMemory(envelope, identity)
		if err != nil {
			t.Fatal(err)
		}
		badSignature := make([]byte, 64)
		commit := writeRawMemoryCommit(t, payload, badSignature, nil)
		ref, _ := memoryRef(identity.Actor, envelope.Stream)
		mustGit(t, "update-ref", ref, commit)
		before := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
		if _, err := collectMemories(); err == nil || !strings.Contains(err.Error(), "signature") {
			t.Fatalf("invalid signature error = %v", err)
		}
		if after := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); after != before {
			t.Fatalf("failed collection mutated refs:\nbefore=%s\nafter=%s", before, after)
		}
	})
}

func TestMemoryStoreRoundTripsEveryLifecycleOperationWithoutResolvingTargets(t *testing.T) {
	for _, operation := range []string{memoryOperationSupersede, memoryOperationRetract, memoryOperationChallenge} {
		t.Run(operation, func(t *testing.T) {
			withMemoryRepository(t, func() {
				identity := deterministicMemoryIdentity()
				envelope := validMemoryEnvelopeFixture(operation)
				stored, err := appendMemory(envelope, identity)
				if err != nil {
					t.Fatal(err)
				}
				memories, err := collectMemories()
				if err != nil {
					t.Fatal(err)
				}
				if len(memories) != 1 || memories[0].ID != stored.ID || memories[0].Envelope.Operation != operation {
					t.Fatalf("stored %s memories = %#v", operation, memories)
				}
				if memories[0].Envelope.Target != envelope.Target || envelope.Target == stored.ID {
					t.Fatalf("%s target was resolved or changed at storage boundary", operation)
				}
			})
		})
	}
}

func TestMemoryStreamRejectsCompleteTreeAndWireCorruptionMatrix(t *testing.T) {
	type corruptionFixture struct {
		name string
		want string
		make func(t *testing.T, identity *Identity, envelope MemoryEnvelope, payload, signature []byte) string
	}
	fixtures := []corruptionFixture{
		{
			name: "missing memory json", want: "exactly memory.json and signature",
			make: func(t *testing.T, _ *Identity, _ MemoryEnvelope, _ []byte, signature []byte) string {
				signatureBlob := writeMemoryBlob(t, []byte(base64.RawStdEncoding.EncodeToString(signature)))
				return writeMemoryCommitFromEntries(t, []memoryTreeFixture{{Mode: "100644", Kind: "blob", OID: signatureBlob, Name: "signature"}}, nil)
			},
		},
		{
			name: "missing signature", want: "exactly memory.json and signature",
			make: func(t *testing.T, _ *Identity, _ MemoryEnvelope, payload, _ []byte) string {
				memoryBlob := writeMemoryBlob(t, payload)
				return writeMemoryCommitFromEntries(t, []memoryTreeFixture{{Mode: "100644", Kind: "blob", OID: memoryBlob, Name: "memory.json"}}, nil)
			},
		},
		{
			name: "nested path", want: "invalid tree entry",
			make: func(t *testing.T, _ *Identity, _ MemoryEnvelope, payload, signature []byte) string {
				memoryBlob, signatureBlob := writeMemoryBlobs(t, payload, signature)
				subtree := writeMemoryTree(t, []memoryTreeFixture{{Mode: "100644", Kind: "blob", OID: memoryBlob, Name: "memory.json"}}, false)
				return writeMemoryCommitFromEntries(t, []memoryTreeFixture{
					{Mode: "040000", Kind: "tree", OID: subtree, Name: "nested"},
					{Mode: "100644", Kind: "blob", OID: signatureBlob, Name: "signature"},
				}, nil)
			},
		},
		{
			name: "wrong mode", want: "invalid tree entry",
			make: func(t *testing.T, _ *Identity, _ MemoryEnvelope, payload, signature []byte) string {
				memoryBlob, signatureBlob := writeMemoryBlobs(t, payload, signature)
				return writeMemoryCommitFromEntries(t, []memoryTreeFixture{
					{Mode: "100755", Kind: "blob", OID: memoryBlob, Name: "memory.json"},
					{Mode: "100644", Kind: "blob", OID: signatureBlob, Name: "signature"},
				}, nil)
			},
		},
		{
			name: "wrong type", want: "invalid tree entry",
			make: func(t *testing.T, _ *Identity, _ MemoryEnvelope, payload, signature []byte) string {
				memoryBlob, signatureBlob := writeMemoryBlobs(t, payload, signature)
				subtree := writeMemoryTree(t, []memoryTreeFixture{{Mode: "100644", Kind: "blob", OID: memoryBlob, Name: "value"}}, false)
				return writeMemoryCommitFromEntries(t, []memoryTreeFixture{
					{Mode: "040000", Kind: "tree", OID: subtree, Name: "memory.json"},
					{Mode: "100644", Kind: "blob", OID: signatureBlob, Name: "signature"},
				}, nil)
			},
		},
		{
			name: "duplicate names", want: "duplicate entry",
			make: func(t *testing.T, _ *Identity, _ MemoryEnvelope, payload, signature []byte) string {
				memoryBlob, signatureBlob := writeMemoryBlobs(t, payload, signature)
				return writeMemoryCommitFromLiteralTree(t, []memoryTreeFixture{
					{Mode: "100644", Kind: "blob", OID: memoryBlob, Name: "memory.json"},
					{Mode: "100644", Kind: "blob", OID: signatureBlob, Name: "signature"},
					{Mode: "100644", Kind: "blob", OID: signatureBlob, Name: "signature"},
				}, nil)
			},
		},
		{
			name: "invalid raw base64", want: "invalid signature encoding",
			make: func(t *testing.T, _ *Identity, _ MemoryEnvelope, payload, _ []byte) string {
				memoryBlob := writeMemoryBlob(t, payload)
				invalid := writeMemoryBlob(t, []byte("***"))
				return writeMemoryCommitFromEntries(t, []memoryTreeFixture{
					{Mode: "100644", Kind: "blob", OID: memoryBlob, Name: "memory.json"},
					{Mode: "100644", Kind: "blob", OID: invalid, Name: "signature"},
				}, nil)
			},
		},
		{
			name: "noncanonical padded base64", want: "invalid signature encoding",
			make: func(t *testing.T, _ *Identity, _ MemoryEnvelope, payload, signature []byte) string {
				memoryBlob := writeMemoryBlob(t, payload)
				padded := writeMemoryBlob(t, []byte(base64.StdEncoding.EncodeToString(signature)))
				return writeMemoryCommitFromEntries(t, []memoryTreeFixture{
					{Mode: "100644", Kind: "blob", OID: memoryBlob, Name: "memory.json"},
					{Mode: "100644", Kind: "blob", OID: padded, Name: "signature"},
				}, nil)
			},
		},
		{
			name: "unknown envelope", want: "invalid memory JSON",
			make: func(t *testing.T, identity *Identity, _ MemoryEnvelope, _ []byte, _ []byte) string {
				payload := []byte(`{"protocol":"not-memory","private":"must-not-echo"}`)
				return writeRawMemoryCommit(t, payload, signMemoryPayload(t, identity, payload), nil)
			},
		},
		{
			name: "invalid envelope bytes", want: "invalid memory JSON",
			make: func(t *testing.T, identity *Identity, _ MemoryEnvelope, _ []byte, _ []byte) string {
				payload := []byte("not-json-memory-envelope")
				return writeRawMemoryCommit(t, payload, signMemoryPayload(t, identity, payload), nil)
			},
		},
		{
			name: "actor public key mismatch", want: "does not match publicKey",
			make: func(t *testing.T, identity *Identity, _ MemoryEnvelope, payload, _ []byte) string {
				other := testIdentity(t, "Other")
				payload = bytes.Replace(payload, []byte(identity.Actor), []byte(other.Actor), 1)
				return writeRawMemoryCommit(t, payload, signMemoryPayload(t, identity, payload), nil)
			},
		},
		{
			name: "zero sequence", want: "sequence must be positive",
			make: func(t *testing.T, identity *Identity, _ MemoryEnvelope, payload, _ []byte) string {
				payload = bytes.Replace(payload, []byte(`"sequence":1`), []byte(`"sequence":0`), 1)
				return writeRawMemoryCommit(t, payload, signMemoryPayload(t, identity, payload), nil)
			},
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			withMemoryRepository(t, func() {
				identity := deterministicMemoryIdentity()
				envelope := validMemoryEnvelopeFixture(memoryOperationRecord)
				payload, signature, err := encodeAndSignMemory(envelope, identity)
				if err != nil {
					t.Fatal(err)
				}
				commit := fixture.make(t, identity, envelope, payload, signature)
				ref, _ := memoryRef(identity.Actor, envelope.Stream)
				mustGit(t, "update-ref", ref, commit)
				assertMemoryCollectionRejectedWithoutRefMutation(t, fixture.want)
			})
		})
	}
}

func TestMemoryStreamRejectsCompleteContinuityMatrix(t *testing.T) {
	type continuityFixture struct {
		name string
		want string
		make func(t *testing.T, identity *Identity) (string, string)
	}
	fixtures := []continuityFixture{
		{
			name: "root with previous", want: "previous must be absent",
			make: func(t *testing.T, identity *Identity) (string, string) {
				envelope := validMemoryEnvelopeFixture(memoryOperationRecord)
				envelope.Sequence = 2
				envelope.Previous = fullMemoryID("4")
				payload, _ := mustSignMemoryEnvelope(t, envelope, identity)
				payload = bytes.Replace(payload, []byte(`"sequence":2`), []byte(`"sequence":1`), 1)
				signature := signMemoryPayload(t, identity, payload)
				return envelope.Stream, writeRawMemoryCommit(t, payload, signature, nil)
			},
		},
		{
			name: "gap sequence", want: "sequence 3, want 2",
			make: func(t *testing.T, identity *Identity) (string, string) {
				first := writeFirstMemory(t, identity, defaultMemoryStream(identity.Actor))
				third := nextMemoryEnvelope(t, identity, first.Envelope.Stream, first, "gap")
				third.Sequence = 3
				payload, signature := mustSignMemoryEnvelope(t, third, identity)
				return first.Envelope.Stream, writeRawMemoryCommit(t, payload, signature, []string{first.Commit})
			},
		},
		{
			name: "duplicate sequence", want: "sequence 1, want 2",
			make: func(t *testing.T, identity *Identity) (string, string) {
				first := writeFirstMemory(t, identity, defaultMemoryStream(identity.Actor))
				duplicate := validMemoryEnvelopeFixture(memoryOperationRecord)
				payload, signature := mustSignMemoryEnvelope(t, duplicate, identity)
				return first.Envelope.Stream, writeRawMemoryCommit(t, payload, signature, []string{first.Commit})
			},
		},
		{
			name: "decreasing sequence", want: "sequence 1, want 3",
			make: func(t *testing.T, identity *Identity) (string, string) {
				first := writeFirstMemory(t, identity, defaultMemoryStream(identity.Actor))
				secondEnvelope := nextMemoryEnvelope(t, identity, first.Envelope.Stream, first, "second")
				secondPayload, secondSignature := mustSignMemoryEnvelope(t, secondEnvelope, identity)
				secondCommit := writeRawMemoryCommit(t, secondPayload, secondSignature, []string{first.Commit})
				decreasing := validMemoryEnvelopeFixture(memoryOperationRecord)
				payload, signature := mustSignMemoryEnvelope(t, decreasing, identity)
				return first.Envelope.Stream, writeRawMemoryCommit(t, payload, signature, []string{secondCommit})
			},
		},
		{
			name: "skipped predecessor", want: "sequence 3, want 2",
			make: func(t *testing.T, identity *Identity) (string, string) {
				first := writeFirstMemory(t, identity, defaultMemoryStream(identity.Actor))
				secondEnvelope := nextMemoryEnvelope(t, identity, first.Envelope.Stream, first, "second")
				secondPayload, secondSignature := mustSignMemoryEnvelope(t, secondEnvelope, identity)
				secondCommit := writeRawMemoryCommit(t, secondPayload, secondSignature, []string{first.Commit})
				second := &StoredMemory{ID: memoryID(secondPayload), Commit: secondCommit, Envelope: secondEnvelope}
				third := nextMemoryEnvelope(t, identity, first.Envelope.Stream, second, "third")
				payload, signature := mustSignMemoryEnvelope(t, third, identity)
				return first.Envelope.Stream, writeRawMemoryCommit(t, payload, signature, []string{first.Commit})
			},
		},
		{
			name: "cross stream graft", want: "has stream",
			make: func(t *testing.T, identity *Identity) (string, string) {
				wantedStream := fullMemoryID("1")
				otherStream := fullMemoryID("2")
				other := writeFirstMemory(t, identity, otherStream)
				return wantedStream, other.Commit
			},
		},
		{
			name: "wrong ref owner", want: "has owner",
			make: func(t *testing.T, identity *Identity) (string, string) {
				first := writeFirstMemory(t, identity, defaultMemoryStream(identity.Actor))
				return first.Envelope.Stream, first.Commit
			},
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			withMemoryRepository(t, func() {
				identity := deterministicMemoryIdentity()
				stream, head := fixture.make(t, identity)
				owner := identity.Actor
				if fixture.name == "wrong ref owner" {
					owner = testIdentity(t, "Wrong owner").Actor
				}
				ref, _ := memoryRef(owner, stream)
				mustGit(t, "update-ref", ref, head)
				assertMemoryCollectionRejectedWithoutRefMutation(t, fixture.want)
			})
		})
	}

	t.Run("unavailable predecessor", func(t *testing.T) {
		withMemoryRepository(t, func() {
			identity := deterministicMemoryIdentity()
			first := writeFirstMemory(t, identity, defaultMemoryStream(identity.Actor))
			secondEnvelope := nextMemoryEnvelope(t, identity, first.Envelope.Stream, first, "second")
			payload, signature := mustSignMemoryEnvelope(t, secondEnvelope, identity)
			second := writeRawMemoryCommit(t, payload, signature, []string{first.Commit})
			ref, _ := memoryRef(identity.Actor, first.Envelope.Stream)
			mustGit(t, "update-ref", ref, second)
			gitDir := mustGitText(t, "rev-parse", "--absolute-git-dir")
			objectPath := filepath.Join(gitDir, "objects", first.Commit[:2], first.Commit[2:])
			if err := os.Remove(objectPath); err != nil {
				t.Fatal(err)
			}
			assertMemoryCollectionRejectedWithoutRefMutation(t, "unavailable")
		})
	})
}

func TestMemoryStoreCollectsMultipleSourcesInStableProtocolOrder(t *testing.T) {
	withMemoryRepository(t, func() {
		alice := deterministicMemoryIdentity()
		bob := testIdentity(t, "Bob")
		streams := []struct {
			identity *Identity
			stream   string
			count    int
		}{
			{bob, fullMemoryID("f"), 2},
			{alice, fullMemoryID("2"), 1},
			{alice, fullMemoryID("1"), 2},
		}
		type streamHead struct {
			actor, stream, head string
		}
		var heads []streamHead
		for _, fixture := range streams {
			var prior *StoredMemory
			for sequence := 1; sequence <= fixture.count; sequence++ {
				envelope := memoryEnvelopeForStream(t, fixture.identity, fixture.stream, prior, fmt.Sprintf("record-%d", sequence))
				stored, err := appendMemory(envelope, fixture.identity)
				if err != nil {
					t.Fatal(err)
				}
				prior = stored
			}
			heads = append(heads, streamHead{fixture.identity.Actor, fixture.stream, prior.Commit})
		}
		local, err := collectMemories()
		if err != nil {
			t.Fatal(err)
		}
		want := memoryIdentitySequence(local)
		if len(want) != 5 {
			t.Fatalf("local identity sequence = %v", want)
		}

		for index := len(heads) - 1; index >= 0; index-- {
			head := heads[index]
			accepted, _ := acceptedMemoryRef("origin", head.actor, head.stream)
			mustGit(t, "update-ref", accepted, head.head)
			localRef, _ := memoryRef(head.actor, head.stream)
			mustGit(t, "update-ref", "-d", localRef)
		}
		acceptedOnly, err := collectMemories()
		if err != nil || fmt.Sprint(memoryIdentitySequence(acceptedOnly)) != fmt.Sprint(want) {
			t.Fatalf("accepted order = %v, want %v, err=%v", memoryIdentitySequence(acceptedOnly), want, err)
		}

		for index, head := range heads {
			if index%2 == 0 {
				localRef, _ := memoryRef(head.actor, head.stream)
				mustGit(t, "update-ref", localRef, head.head)
			}
		}
		mixed, err := collectMemories()
		if err != nil || fmt.Sprint(memoryIdentitySequence(mixed)) != fmt.Sprint(want) {
			t.Fatalf("mixed order = %v, want %v, err=%v", memoryIdentitySequence(mixed), want, err)
		}
	})
}

func TestMemoryStoreNamespaceExplicitGitDirPendingAndNoncanonicalSources(t *testing.T) {
	t.Run("malformed local namespace", func(t *testing.T) {
		withMemoryRepository(t, func() {
			identity := deterministicMemoryIdentity()
			stored := writeFirstMemory(t, identity, defaultMemoryStream(identity.Actor))
			badRef := memoryRefPrefix + strings.ToUpper(identity.Actor) + "/" + strings.TrimPrefix(stored.Envelope.Stream, "sha256:")
			mustGit(t, "update-ref", badRef, stored.Commit)
			assertMemoryCollectionRejectedWithoutRefMutation(t, "malformed memory ref")
		})
	})

	t.Run("malformed accepted namespace", func(t *testing.T) {
		withMemoryRepository(t, func() {
			identity := deterministicMemoryIdentity()
			stored := writeFirstMemory(t, identity, defaultMemoryStream(identity.Actor))
			badRef := acceptedMemoryRefPrefix + "origin/memory/" + identity.Actor + "/" + strings.ToUpper(strings.TrimPrefix(stored.Envelope.Stream, "sha256:"))
			mustGit(t, "update-ref", badRef, stored.Commit)
			assertMemoryCollectionRejectedWithoutRefMutation(t, "malformed accepted memory ref")
		})
	})

	t.Run("unrelated accepted refs ignored and explicit git dir reusable", func(t *testing.T) {
		withMemoryRepository(t, func() {
			identity := deterministicMemoryIdentity()
			stored, err := appendMemory(validMemoryEnvelopeFixture(memoryOperationRecord), identity)
			if err != nil {
				t.Fatal(err)
			}
			mustGit(t, "update-ref", acceptedActorRef("origin", identity.Actor), stored.Commit)
			mustGit(t, "update-ref", acceptedProposalRef("origin", fullMemoryID("3")), stored.Commit)
			gitDir := mustGitText(t, "rev-parse", "--absolute-git-dir")
			ref, _ := memoryRef(identity.Actor, stored.Envelope.Stream)
			loaded, err := loadMemoryStreamAt(gitDir, memoryStreamSource{Ref: ref, Actor: identity.Actor, Stream: stored.Envelope.Stream, Head: stored.Commit})
			if err != nil || len(loaded) != 1 || loaded[0].ID != stored.ID {
				t.Fatalf("explicit gitDir load = %#v, %v", loaded, err)
			}
			unreferencedEnvelope := validMemoryEnvelopeFixture(memoryOperationRecord)
			unreferencedEnvelope.Timestamp = "2026-08-30T13:00:00Z"
			unreferencedEnvelope.Record.Content = "unreferenced"
			payload, signature := mustSignMemoryEnvelope(t, unreferencedEnvelope, identity)
			unreferenced := writeRawMemoryCommit(t, payload, signature, nil)
			mustGit(t, "update-ref", "refs/hn/quarantine/txn-memory/stream", unreferenced)
			trulyUnreferencedEnvelope := unreferencedEnvelope
			trulyUnreferencedEnvelope.Timestamp = "2026-08-30T13:01:00Z"
			trulyUnreferencedEnvelope.Record.Content = "truly-unreferenced"
			orphanPayload, orphanSignature := mustSignMemoryEnvelope(t, trulyUnreferencedEnvelope, identity)
			_ = writeRawMemoryCommit(t, orphanPayload, orphanSignature, nil)
			memories, err := collectMemories()
			if err != nil || len(memories) != 1 || memories[0].ID != stored.ID {
				t.Fatalf("noncanonical source collection = %#v, %v", memories, err)
			}
		})
	})

	t.Run("replication pending accepted head", func(t *testing.T) {
		withMemoryRepository(t, func() {
			identity := deterministicMemoryIdentity()
			stored, err := appendMemory(validMemoryEnvelopeFixture(memoryOperationRecord), identity)
			if err != nil {
				t.Fatal(err)
			}
			local, _ := memoryRef(identity.Actor, stored.Envelope.Stream)
			accepted, _ := acceptedMemoryRef("origin", identity.Actor, stored.Envelope.Stream)
			mustGit(t, "update-ref", accepted, stored.Commit)
			mustGit(t, "update-ref", "-d", local)
			gitDir := mustGitText(t, "rev-parse", "--absolute-git-dir")
			recordMemoryPendingFixture(t, gitDir, stored.Commit)
			assertMemoryCollectionRejectedWithoutRefMutation(t, "replication acceptance pending")
		})
	})
}

func TestMemoryStoreCorruptionCannotAffectCollaborationOrExposePrivateSentinels(t *testing.T) {
	withMemoryRepository(t, func() {
		collaborator := testIdentity(t, "Collaborator")
		event := newEvent(collaborator, "issue.open", 1, "")
		event.Title = "Independent collaboration"
		storedEvent, err := appendEvent(event, collaborator)
		if err != nil {
			t.Fatal(err)
		}
		beforeEvents, err := collectEvents()
		if err != nil {
			t.Fatal(err)
		}
		beforeRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/actors", "refs/hn/proposals")

		memoryIdentity := deterministicMemoryIdentity()
		if _, exists, err := refValue(actorRef(memoryIdentity.Actor)); err != nil || exists {
			t.Fatalf("memory actor unexpectedly has collaboration ref: exists=%v err=%v", exists, err)
		}
		storedMemory, err := appendMemory(validMemoryEnvelopeFixture(memoryOperationRecord), memoryIdentity)
		if err != nil {
			t.Fatal(err)
		}
		if memories, err := collectMemories(); err != nil || len(memories) != 1 || memories[0].ID != storedMemory.ID {
			t.Fatalf("memory actor without collaboration ref = %#v, %v", memories, err)
		}
		if _, exists, err := refValue(actorRef(memoryIdentity.Actor)); err != nil || exists {
			t.Fatalf("memory operations created collaboration actor ref: exists=%v err=%v", exists, err)
		}
		secret := "HN_PRIVATE_SENTINEL_84a2d992"
		t.Setenv("HN_MEMORY_PRIVATE_SENTINEL", secret)
		gitDir := mustGitText(t, "rev-parse", "--absolute-git-dir")
		privateFiles := []string{
			filepath.Join(gitDir, "hn", "identities", "sentinel-keyring.json"),
			filepath.Join(gitDir, "hn", "memory", "index-v0.json"),
			"private-working-memory.txt",
		}
		for _, name := range privateFiles {
			if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(name, []byte(secret), 0o600); err != nil {
				t.Fatal(err)
			}
		}

		payload := []byte(`{"protocol":"invalid","content":"` + secret + `"}`)
		badCommit := writeRawMemoryCommit(t, payload, signMemoryPayload(t, memoryIdentity, payload), nil)
		ref, _ := memoryRef(memoryIdentity.Actor, storedMemory.Envelope.Stream)
		mustGit(t, "update-ref", ref, badCommit, storedMemory.Commit)
		allRefsBeforeFailure := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
		_, memoryErr := collectMemories()
		if memoryErr == nil || strings.Contains(memoryErr.Error(), secret) {
			t.Fatalf("corrupt memory diagnostic = %v", memoryErr)
		}
		if refs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); refs != allRefsBeforeFailure {
			t.Fatalf("corrupt memory load changed refs:\nbefore=%s\nafter=%s", allRefsBeforeFailure, refs)
		}
		afterEvents, err := collectEvents()
		if err != nil || len(afterEvents) != len(beforeEvents) || afterEvents[0].ID != storedEvent.ID || !bytes.Equal(afterEvents[0].Payload, beforeEvents[0].Payload) {
			t.Fatalf("corrupt memory affected collaboration: %#v, %v", afterEvents, err)
		}
		if refs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/actors", "refs/hn/proposals"); refs != beforeRefs {
			t.Fatalf("corrupt memory changed collaboration refs:\nbefore=%s\nafter=%s", beforeRefs, refs)
		}
		mustGit(t, "update-ref", "-d", ref)
		if memories, err := collectMemories(); err != nil || len(memories) != 0 {
			t.Fatalf("removed memory ref collection = %#v, %v", memories, err)
		}
		if afterRemoval, err := collectEvents(); err != nil || len(afterRemoval) != 1 || afterRemoval[0].ID != storedEvent.ID {
			t.Fatalf("memory ref removal affected collaboration: %#v, %v", afterRemoval, err)
		}
	})
}

type memoryTreeFixture struct {
	Mode string
	Kind string
	OID  string
	Name string
}

func assertMemoryCollectionRejectedWithoutRefMutation(t *testing.T, want string) {
	t.Helper()
	before := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn")
	_, err := collectMemories()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("memory collection error = %v, want substring %q", err, want)
	}
	if after := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn"); after != before {
		t.Fatalf("rejected memory changed refs:\nbefore=%s\nafter=%s", before, after)
	}
}

func writeFirstMemory(t *testing.T, identity *Identity, stream string) *StoredMemory {
	t.Helper()
	envelope := memoryEnvelopeForStream(t, identity, stream, nil, "first")
	payload, signature := mustSignMemoryEnvelope(t, envelope, identity)
	commit := writeRawMemoryCommit(t, payload, signature, nil)
	return &StoredMemory{ID: memoryID(payload), Commit: commit, Envelope: envelope, Payload: payload, Signature: signature}
}

func memoryEnvelopeForStream(t *testing.T, identity *Identity, stream string, previous *StoredMemory, content string) MemoryEnvelope {
	t.Helper()
	envelope := validMemoryEnvelopeFixture(memoryOperationRecord)
	envelope.Actor = identity.Actor
	envelope.ActorName = identity.Name
	envelope.PublicKey = identity.PublicKey
	envelope.Stream = stream
	envelope.Record.Content = content
	if previous != nil {
		envelope.Sequence = previous.Envelope.Sequence + 1
		envelope.Previous = previous.ID
		envelope.Timestamp = fmt.Sprintf("2026-08-30T13:%02d:00Z", envelope.Sequence)
	}
	return envelope
}

func mustSignMemoryEnvelope(t *testing.T, envelope MemoryEnvelope, identity *Identity) ([]byte, []byte) {
	t.Helper()
	payload, signature, err := encodeAndSignMemory(envelope, identity)
	if err != nil {
		t.Fatal(err)
	}
	return payload, signature
}

func signMemoryPayload(t *testing.T, identity *Identity, payload []byte) []byte {
	t.Helper()
	privateKey, err := identity.privateKey()
	if err != nil {
		t.Fatal(err)
	}
	return ed25519.Sign(privateKey, payload)
}

func writeMemoryBlob(t *testing.T, content []byte) string {
	t.Helper()
	return mustGitTextFromInput(t, content, "hash-object", "-w", "--stdin")
}

func writeMemoryBlobs(t *testing.T, payload, signature []byte) (string, string) {
	t.Helper()
	return writeMemoryBlob(t, payload), writeMemoryBlob(t, []byte(base64.RawStdEncoding.EncodeToString(signature)))
}

func writeMemoryTree(t *testing.T, entries []memoryTreeFixture, literal bool) string {
	t.Helper()
	if literal {
		var raw bytes.Buffer
		for _, entry := range entries {
			raw.WriteString(entry.Mode + " " + entry.Name)
			raw.WriteByte(0)
			oid, err := hex.DecodeString(entry.OID)
			if err != nil {
				t.Fatal(err)
			}
			raw.Write(oid)
		}
		return mustGitTextFromInput(t, raw.Bytes(), "hash-object", "-t", "tree", "--literally", "-w", "--stdin")
	}
	var input strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(&input, "%s %s %s\t%s\n", entry.Mode, entry.Kind, entry.OID, entry.Name)
	}
	return mustGitTextFromInput(t, []byte(input.String()), "mktree")
}

func writeMemoryCommitFromEntries(t *testing.T, entries []memoryTreeFixture, parents []string) string {
	t.Helper()
	tree := writeMemoryTree(t, entries, false)
	return writeMemoryCommitFromTree(t, tree, parents)
}

func writeMemoryCommitFromLiteralTree(t *testing.T, entries []memoryTreeFixture, parents []string) string {
	t.Helper()
	tree := writeMemoryTree(t, entries, true)
	return writeMemoryCommitFromTree(t, tree, parents)
}

func writeMemoryCommitFromTree(t *testing.T, tree string, parents []string) string {
	t.Helper()
	args := []string{"commit-tree", tree, "-m", "memory corruption fixture"}
	for _, parent := range parents {
		args = append(args, "-p", parent)
	}
	return mustGitTextFromInput(t, nil, args...)
}

func memoryIdentitySequence(memories []StoredMemory) []string {
	identities := make([]string, len(memories))
	for index, memory := range memories {
		identities[index] = fmt.Sprintf("%s/%s/%d/%s", memory.Envelope.Actor, memory.Envelope.Stream, memory.Envelope.Sequence, memory.ID)
	}
	return identities
}

func recordMemoryPendingFixture(t *testing.T, gitDir, commit string) {
	t.Helper()
	result := replicationTransactionResult{ID: "txn-memory-pending", Remote: "origin", pendingObjects: []string{commit}}
	if err := createReplicationPendingAnchor(gitDir, result); err != nil {
		t.Fatal(err)
	}
	if err := recordReplicationTransaction(gitDir, result, "validated"); err != nil {
		t.Fatal(err)
	}
}

func withMemoryRepository(t *testing.T, run func()) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("restore directory: %v", err)
		}
	}()
	mustGit(t, "init", "-q", "-b", "main")
	mustGit(t, "config", "user.name", "Test")
	mustGit(t, "config", "user.email", "test@hn.invalid")
	run()
}

func nextMemoryEnvelope(t *testing.T, identity *Identity, stream string, previous *StoredMemory, content string) MemoryEnvelope {
	t.Helper()
	envelope := validMemoryEnvelopeFixture(memoryOperationRecord)
	envelope.Actor = identity.Actor
	envelope.ActorName = identity.Name
	envelope.PublicKey = identity.PublicKey
	envelope.Stream = stream
	envelope.Sequence = previous.Envelope.Sequence + 1
	envelope.Previous = previous.ID
	envelope.Timestamp = fmt.Sprintf("2026-08-30T12:34:%02dZ", envelope.Sequence)
	envelope.Record.Content = content
	return envelope
}

func mustGitTextFromInput(t *testing.T, input []byte, args ...string) string {
	t.Helper()
	out, err := gitInput(input, nil, args...)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func writeRawMemoryCommit(t *testing.T, payload, signature []byte, parents []string) string {
	t.Helper()
	memoryBlob := mustGitTextFromInput(t, payload, "hash-object", "-w", "--stdin")
	signatureBlob := mustGitTextFromInput(t, []byte(base64.RawStdEncoding.EncodeToString(signature)), "hash-object", "-w", "--stdin")
	tree := mustGitTextFromInput(t, []byte(fmt.Sprintf("100644 blob %s\tmemory.json\n100644 blob %s\tsignature\n", memoryBlob, signatureBlob)), "mktree")
	args := []string{"commit-tree", tree, "-m", "memory fixture"}
	for _, parent := range parents {
		args = append(args, "-p", parent)
	}
	return mustGitTextFromInput(t, nil, args...)
}
