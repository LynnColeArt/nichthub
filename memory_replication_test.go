package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMemoryReplicationSelectionCompatibilityAndValidation(t *testing.T) {
	inReplicationTestRepository(t)
	actor := deterministicMemoryIdentity().Actor
	streamA := fullMemoryID("1")
	streamB := fullMemoryID("2")
	proposal := "sha256:" + strings.Repeat("3", 64)

	if err := cmdReplication([]string{
		"select", "origin", "--actor", actor, "--proposal", proposal,
		"--memory", streamB, "--memory", streamA,
	}); err != nil {
		t.Fatal(err)
	}
	selection, explicit, err := loadReplicationSelection("origin")
	if err != nil {
		t.Fatal(err)
	}
	if !explicit || !reflect.DeepEqual(selection.Memories, []string{streamA, streamB}) {
		t.Fatalf("memory selection = %#v, explicit=%t", selection.Memories, explicit)
	}
	shown, err := captureTestOutput(t, func() error { return cmdReplication([]string{"show", "origin"}) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(shown, "memory: "+streamA) || strings.Index(shown, streamA) > strings.Index(shown, streamB) {
		t.Fatalf("memory selection display is not deterministic:\n%s", shown)
	}

	cases := [][]string{
		{"select", "--memory", streamA[:20]},
		{"select", "--memory", strings.ToUpper(streamA)},
		{"select", "--memory", streamA + " "},
		{"select", "--memory", streamA, "--memory", streamA},
		{"select", "--all", "--memory", streamA},
	}
	for _, args := range cases {
		if err := cmdReplication(args); err == nil {
			t.Fatalf("cmdReplication(%q) succeeded", args)
		}
	}

	legacy := ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "legacy", Actors: []string{actor},
		Budgets: defaultReplicationBudgets(),
	}
	path, err := replicationSelectionPath("legacy")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, explicit, err := loadReplicationSelection("legacy")
	if err != nil || !explicit || len(loaded.Memories) != 0 || !reflect.DeepEqual(loaded.Actors, legacy.Actors) {
		t.Fatalf("legacy selection = %#v, explicit=%t, err=%v", loaded, explicit, err)
	}
}

func TestMemoryReplicationDiscoveryRequestAndPromotion(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "project.git")
	receiver := filepath.Join(root, "receiver")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@nh.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
	if err := os.Chdir(publisher); err != nil {
		t.Fatal(err)
	}
	identity := deterministicMemoryIdentity()
	envelope := validMemoryEnvelopeFixture(memoryOperationRecord)
	envelope.Record.Anchor = MemoryAnchor{Commit: mustGitText(t, "rev-parse", "HEAD")}
	envelope.Record.Applicability = Applicability{Mode: memoryApplicabilityExact}
	envelope.Record.Evidence = []string{}
	stored, err := appendMemory(envelope, identity)
	if err != nil {
		t.Fatal(err)
	}
	localRef, err := memoryRef(identity.Actor, stored.Envelope.Stream)
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	mustGit(t, "push", "-q", "origin", localRef+":"+localRef)
	mustGit(t, "clone", "-q", remote, receiver)
	if err := os.Chdir(receiver); err != nil {
		t.Fatal(err)
	}

	selection := ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin", Memories: []string{stored.Envelope.Stream},
		Budgets: defaultReplicationBudgets(),
	}
	result, err := runReplicationTransaction(selection)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outcomes) != 1 || result.Outcomes[0].Kind != replicationMemory || result.Outcomes[0].ID != stored.Envelope.Stream || result.Outcomes[0].Status != replicationPromoted {
		t.Fatalf("memory outcomes = %#v", result.Outcomes)
	}
	accepted, err := acceptedMemoryRef("origin", identity.Actor, stored.Envelope.Stream)
	if err != nil {
		t.Fatal(err)
	}
	assertRefValue(t, accepted, stored.Commit)
	memories, err := collectMemories()
	if err != nil || len(memories) != 1 || memories[0].ID != stored.ID {
		t.Fatalf("accepted memories = %#v, err=%v", memories, err)
	}
}

func TestMemoryReplicationExplicitCollaborationSelectionRequestsNoMemory(t *testing.T) {
	inReplicationTestRepository(t)
	actor := deterministicMemoryIdentity().Actor
	stream := fullMemoryID("4")
	ref, err := memoryRef(actor, stream)
	if err != nil {
		t.Fatal(err)
	}
	requests, _, err := resolveReplicationRequests(ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin", Actors: []string{actor},
		Budgets: defaultReplicationBudgets(),
	}, testAdvertisedRemote(t, map[string]string{ref: mustGitText(t, "rev-parse", "HEAD")}))
	if err != nil {
		t.Fatal(err)
	}
	for _, request := range requests {
		if request.Kind == replicationMemory {
			t.Fatalf("actor-only selection requested memory: %#v", request)
		}
	}
}

func TestMemoryReplicationDiscoveryRejectsAmbiguousOwnersAndIgnoresMalformedRefs(t *testing.T) {
	inReplicationTestRepository(t)
	stream := fullMemoryID("5")
	alice := deterministicMemoryIdentity()
	bob := testIdentity(t, "Bob Memory")
	validAlice, _ := memoryRef(alice.Actor, stream)
	validBob, _ := memoryRef(bob.Actor, stream)
	malformed := memoryRefPrefix + alice.Actor + "/" + strings.Repeat("A", 64)
	head := mustGitText(t, "rev-parse", "HEAD")
	remote := testAdvertisedRemote(t, map[string]string{validAlice: head, validBob: head, malformed: head})
	requests, outcomes, err := resolveReplicationRequests(ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin", All: true,
		Budgets: defaultReplicationBudgets(),
	}, remote)
	if err != nil {
		t.Fatal(err)
	}
	for _, request := range requests {
		if request.Kind == replicationMemory {
			t.Fatalf("ambiguous or malformed memory was requested: %#v", request)
		}
	}
	if !replicationOutcomeHasStatus(outcomes, replicationMemory, stream, replicationStructuralInvalid) {
		t.Fatalf("ambiguous memory owners were not a scoped failure: %#v", outcomes)
	}
}

func TestMemoryReplicationPublishUsesExactLocalRefs(t *testing.T) {
	inReplicationTestRepository(t)
	remote := filepath.Join(t.TempDir(), "published.git")
	mustGit(t, "init", "--bare", "-q", remote)
	mustGit(t, "remote", "add", "origin", remote)
	identity := deterministicMemoryIdentity()
	stored := appendReplicableMemory(t, identity, defaultMemoryStream(identity.Actor), nil, "publish")
	ref, _ := memoryRef(identity.Actor, stored.Envelope.Stream)
	if err := publishLocalFacts("origin"); err != nil {
		t.Fatal(err)
	}
	advertised := mustGitText(t, "ls-remote", "--refs", remote, ref)
	if !strings.Contains(advertised, stored.Commit+"\t"+ref) {
		t.Fatalf("published refs = %q, want exact %s", advertised, ref)
	}
}

func TestMemoryReplicationValidationIsolatedFromValidStream(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "project.git")
	receiver := filepath.Join(root, "receiver")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@nh.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
	if err := os.Chdir(publisher); err != nil {
		t.Fatal(err)
	}
	alice := deterministicMemoryIdentity()
	valid := appendReplicableMemory(t, alice, defaultMemoryStream(alice.Actor), nil, "valid")
	bob := testIdentity(t, "Bob Memory")
	invalid := appendReplicableMemory(t, bob, defaultMemoryStream(bob.Actor), nil, "invalid signature")
	badSignature := append([]byte(nil), invalid.Signature...)
	badSignature[0] ^= 0xff
	badCommit := writeRawMemoryCommit(t, invalid.Payload, badSignature, nil)
	badRef, _ := memoryRef(bob.Actor, invalid.Envelope.Stream)
	mustGit(t, "update-ref", badRef, badCommit)

	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	validRef, _ := memoryRef(alice.Actor, valid.Envelope.Stream)
	for _, ref := range []string{validRef, badRef} {
		mustGit(t, "push", "-q", "origin", ref+":"+ref)
	}
	mustGit(t, "clone", "-q", remote, receiver)
	if err := os.Chdir(receiver); err != nil {
		t.Fatal(err)
	}
	result, err := runReplicationTransaction(ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin",
		Memories: []string{valid.Envelope.Stream, invalid.Envelope.Stream}, Budgets: defaultReplicationBudgets(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !replicationOutcomeHasStatus(result.Outcomes, replicationMemory, valid.Envelope.Stream, replicationPromoted) ||
		!replicationOutcomeHasStatus(result.Outcomes, replicationMemory, invalid.Envelope.Stream, replicationStructuralInvalid) {
		t.Fatalf("isolated validation outcomes = %#v", result.Outcomes)
	}
	validAccepted, _ := acceptedMemoryRef("origin", alice.Actor, valid.Envelope.Stream)
	badAccepted, _ := acceptedMemoryRef("origin", bob.Actor, invalid.Envelope.Stream)
	assertRefValue(t, validAccepted, valid.Commit)
	assertRefAbsent(t, badAccepted)
}

func TestMemoryReplicationBudgetCountsCommits(t *testing.T) {
	publisher, remote, receiver, identity, first := setupMemoryReplicationFixture(t)
	if err := os.Chdir(publisher); err != nil {
		t.Fatal(err)
	}
	second := appendReplicableMemory(t, identity, first.Envelope.Stream, first, "second")
	ref, _ := memoryRef(identity.Actor, first.Envelope.Stream)
	mustGit(t, "push", "-q", "--force", "origin", ref+":"+ref)
	if err := os.Chdir(receiver); err != nil {
		t.Fatal(err)
	}
	budgets := defaultReplicationBudgets()
	budgets.MaxEvents = 1
	result, err := runReplicationTransaction(ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin", Memories: []string{first.Envelope.Stream}, Budgets: budgets,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !replicationOutcomeHasStatus(result.Outcomes, replicationMemory, first.Envelope.Stream, replicationOverBudget) {
		t.Fatalf("two-commit stream was not over max-events=1: %#v (head %s, remote %s)", result.Outcomes, second.Commit, remote)
	}
}

func TestMemoryReplicationTransactionInterruptionPreservesAcceptedRef(t *testing.T) {
	_, _, receiver, identity, stored := setupMemoryReplicationFixture(t)
	if err := os.Chdir(receiver); err != nil {
		t.Fatal(err)
	}
	selection := ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin", Memories: []string{stored.Envelope.Stream}, Budgets: defaultReplicationBudgets(),
	}
	replicationBeforePromoteHook = func() error { return os.ErrClosed }
	t.Cleanup(func() { replicationBeforePromoteHook = nil })
	_, err := runReplicationTransaction(selection)
	replicationBeforePromoteHook = nil
	if err == nil || !strings.Contains(err.Error(), "accepted refs and trust projection are unchanged") {
		t.Fatalf("pre-promotion interruption = %v", err)
	}
	accepted, _ := acceptedMemoryRef("origin", identity.Actor, stored.Envelope.Stream)
	assertRefAbsent(t, accepted)
	result, err := runReplicationTransaction(selection)
	if err != nil || !replicationOutcomeHasStatus(result.Outcomes, replicationMemory, stored.Envelope.Stream, replicationPromoted) {
		t.Fatalf("retry outcomes = %#v, err=%v", result.Outcomes, err)
	}
	assertRefValue(t, accepted, stored.Commit)
}

func TestMemoryReplicationShallowRecoverySubsetIsExact(t *testing.T) {
	stream := fullMemoryID("6")
	other := fullMemoryID("7")
	selection := ReplicationSelection{
		Version: replicationSelectionVersion, Remote: "origin", Memories: []string{stream, other}, Budgets: defaultReplicationBudgets(),
	}
	subset, err := recoverySelectionSubset(selection, &ShallowDependencyGap{
		OwnerKind: replicationMemory, OwnerID: stream, MissingID: fullMemoryID("8"), Recovery: "retry",
	})
	if err != nil || !reflect.DeepEqual(subset.Memories, []string{stream}) || len(subset.Actors) != 0 || len(subset.Proposals) != 0 {
		t.Fatalf("memory recovery subset = %#v, err=%v", subset, err)
	}
}

func TestMemoryReplicationFreshCloneConvergesWithoutPrivateState(t *testing.T) {
	_, _, receiver, identity, stored := setupMemoryReplicationFixture(t)
	if err := os.Chdir(receiver); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "no-global-config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	if err := cmdReplication([]string{"select", "origin", "--memory", stored.Envelope.Stream}); err != nil {
		t.Fatal(err)
	}
	if output, err := captureTestOutput(t, func() error { return cmdSync(nil) }); err != nil {
		t.Fatalf("fresh-clone sync failed: %v\n%s", err, output)
	}
	accepted, _ := acceptedMemoryRef("origin", identity.Actor, stored.Envelope.Stream)
	assertRefValue(t, accepted, stored.Commit)
	memories, err := collectMemories()
	if err != nil || len(memories) != 1 || memories[0].ID != stored.ID {
		t.Fatalf("fresh clone memories = %#v, err=%v", memories, err)
	}
	gitDir := mustGitText(t, "rev-parse", "--absolute-git-dir")
	for _, private := range []string{
		filepath.Join(gitDir, "nh", "identity.json"),
		filepath.Join(gitDir, "nh", "identities"),
		filepath.Join(gitDir, "nh", "memory", "index-v0.json"),
	} {
		if _, err := os.Stat(private); !os.IsNotExist(err) {
			t.Fatalf("fresh-clone sync created private state %s: %v", filepath.Base(private), err)
		}
	}
}

func appendReplicableMemory(t *testing.T, identity *Identity, stream string, previous *StoredMemory, content string) *StoredMemory {
	t.Helper()
	envelope := validMemoryEnvelopeFixture(memoryOperationRecord)
	envelope.Actor = identity.Actor
	envelope.ActorName = identity.Name
	envelope.PublicKey = identity.PublicKey
	envelope.Stream = stream
	envelope.Record.Anchor = MemoryAnchor{Commit: mustGitText(t, "rev-parse", "HEAD")}
	envelope.Record.Applicability = Applicability{Mode: memoryApplicabilityExact}
	envelope.Record.Evidence = []string{}
	envelope.Record.Content = content
	if previous != nil {
		envelope.Sequence = previous.Envelope.Sequence + 1
		envelope.Previous = previous.ID
		envelope.Timestamp = "2026-08-30T12:35:00Z"
	}
	stored, err := appendMemory(envelope, identity)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func setupMemoryReplicationFixture(t *testing.T) (publisher, remote, receiver string, identity *Identity, stored *StoredMemory) {
	t.Helper()
	root := t.TempDir()
	publisher = filepath.Join(root, "publisher")
	remote = filepath.Join(root, "project.git")
	receiver = filepath.Join(root, "receiver")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@nh.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
	if err := os.Chdir(publisher); err != nil {
		t.Fatal(err)
	}
	identity = deterministicMemoryIdentity()
	stored = appendReplicableMemory(t, identity, defaultMemoryStream(identity.Actor), nil, "replicable")
	mustGit(t, "clone", "-q", "--bare", publisher, remote)
	mustGit(t, "remote", "add", "origin", remote)
	ref, _ := memoryRef(identity.Actor, stored.Envelope.Stream)
	mustGit(t, "push", "-q", "origin", ref+":"+ref)
	mustGit(t, "clone", "-q", remote, receiver)
	return publisher, remote, receiver, identity, stored
}

func testAdvertisedRemote(t *testing.T, refs map[string]string) string {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "advertised.git")
	mustGit(t, "init", "--bare", "-q", remote)
	for ref, oid := range refs {
		mustGit(t, "push", "-q", remote, oid+":"+ref)
	}
	return remote
}
