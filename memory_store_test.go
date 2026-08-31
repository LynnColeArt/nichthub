package main

import (
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
	if want := "refs/nh/memory/" + actor + "/" + strings.TrimPrefix(stream, "sha256:"); local != want {
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

	goldenInput := []byte("nh-memory-stream-v0\x00" + actor + "\x00default")
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
		"refs/nh/memory/" + actor,
		"refs/nh/memory/" + actor + "/" + strings.Repeat("a", 64) + "/extra",
		"refs/nh/memory/" + strings.ToUpper(actor) + "/" + strings.Repeat("a", 64),
		"refs/nh/remotes/origin/memory/" + actor,
		"refs/nh/remotes/origin/memory/" + actor + "/" + strings.Repeat("A", 64),
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
		beforeCollaborationRefs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh/actors", "refs/nh/proposals")
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
		privateDir := filepath.Join(gitDir, "nh", "memory")
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
		if refs := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh/actors", "refs/nh/proposals"); refs != beforeCollaborationRefs {
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
	mustGit(t, "config", "user.email", "test@nh.invalid")
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
		if _, err := collectMemories(); err == nil || !strings.Contains(err.Error(), "previous") {
			t.Fatalf("signed previous error = %v", err)
		}

		mustGit(t, "update-ref", ref, first.Commit, badPreviousCommit)
		otherRoot := writeRawMemoryCommit(t, first.Payload, first.Signature, nil)
		validNext := nextMemoryEnvelope(t, identity, first.Envelope.Stream, first, "merge")
		payload, signature, err = encodeAndSignMemory(validNext, identity)
		if err != nil {
			t.Fatal(err)
		}
		merge := writeRawMemoryCommit(t, payload, signature, []string{first.Commit, otherRoot})
		mustGit(t, "update-ref", ref, merge, first.Commit)
		if _, err := collectMemories(); err == nil || !strings.Contains(err.Error(), "parent") {
			t.Fatalf("merge parent error = %v", err)
		}
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
		before := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh")
		if _, err := collectMemories(); err == nil || !strings.Contains(err.Error(), "signature") {
			t.Fatalf("invalid signature error = %v", err)
		}
		if after := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh"); after != before {
			t.Fatalf("failed collection mutated refs:\nbefore=%s\nafter=%s", before, after)
		}
	})
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
	mustGit(t, "config", "user.email", "test@nh.invalid")
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
