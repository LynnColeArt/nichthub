package main

import (
	"crypto/ed25519"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNamespaceCutoverWireIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"collaboration", protocolVersion, "hn/0"},
		{"memory", memoryProtocolVersion, "hn-memory/0"},
		{"pipeline", pipelineVersion, "hn.pipeline/0"},
		{"policy", policyVersion, "hn.policy/0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("wire identifier = %q, want %q", test.got, test.want)
			}
		})
	}
}

func TestNamespaceCutoverRefConstructors(t *testing.T) {
	actor := strings.Repeat("a", 64)
	event := "sha256:" + strings.Repeat("b", 64)
	stream := "sha256:" + strings.Repeat("c", 64)

	refs := []string{
		actorRef(actor),
		proposalRef(event),
		acceptedActorRef("origin", actor),
		acceptedProposalRef("origin", event),
	}
	memory, err := memoryRef(actor, stream)
	if err != nil {
		t.Fatal(err)
	}
	refs = append(refs, memory)
	acceptedMemory, err := acceptedMemoryRef("origin", actor, stream)
	if err != nil {
		t.Fatal(err)
	}
	refs = append(refs, acceptedMemory)

	for _, ref := range refs {
		if !strings.HasPrefix(ref, "refs/hn/") {
			t.Errorf("ref %q is outside refs/hn", ref)
		}
		if strings.HasPrefix(ref, "refs/nh/") {
			t.Errorf("ref %q uses legacy namespace", ref)
		}
	}
}

func TestNamespaceCutoverPrivatePaths(t *testing.T) {
	root := inIdentityTestRepository(t)
	gitDir := filepath.Join(root, ".git")

	paths, err := identityKeyringPaths()
	if err != nil {
		t.Fatal(err)
	}
	selection, err := replicationSelectionPath("origin")
	if err != nil {
		t.Fatal(err)
	}
	index, err := memoryIndexPath()
	if err != nil {
		t.Fatal(err)
	}
	gap, err := shallowGapPath()
	if err != nil {
		t.Fatal(err)
	}

	wantRoot := filepath.Join(gitDir, "hn")
	for _, path := range []string{paths.root, paths.legacy, paths.active, paths.rotation, selection, index, gap} {
		relative, err := filepath.Rel(wantRoot, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			t.Errorf("private path %q is outside %q", path, wantRoot)
		}
	}
}

func TestNamespaceCutoverIgnoresLegacyIdentityState(t *testing.T) {
	root := inIdentityTestRepository(t)
	legacyRoot := filepath.Join(root, ".git", "nh")
	if err := os.MkdirAll(legacyRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	legacyIdentity := testIdentity(t, "Legacy Alice")
	legacyBytes, err := json.Marshal(legacyIdentity)
	if err != nil {
		t.Fatal(err)
	}
	legacyPath := filepath.Join(legacyRoot, "identity.json")
	if err := os.WriteFile(legacyPath, legacyBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := loadIdentity(); err == nil || !strings.Contains(err.Error(), "run 'hn init'") {
		t.Fatalf("legacy-only identity lookup = %v, want fresh hn identity guidance", err)
	}
	active, _, err := createIdentity("Active Alice")
	if err != nil {
		t.Fatal(err)
	}
	if active.Actor == legacyIdentity.Actor {
		t.Fatal("active identity was imported from legacy private state")
	}
	after, err := os.ReadFile(legacyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(legacyBytes) {
		t.Fatal("legacy identity state was modified")
	}
	if _, err := os.Stat(filepath.Join(root, ".git", "hn", "active")); err != nil {
		t.Fatalf("active hn identity state was not created: %v", err)
	}
}

func TestNamespaceCutoverIgnoresLegacyInterruptionEnvironment(t *testing.T) {
	t.Setenv("NH_INTERNAL_TESTING", "1")
	t.Setenv("NH_TEST_ROTATION_INTERRUPT_AFTER", "before-rotation-cleanup")
	t.Setenv("NH_TEST_REPLICATION_INTERRUPT_AFTER", "after-fetch")
	if err := runIdentityStorageHook("before-rotation-cleanup"); err != nil {
		t.Fatalf("legacy rotation environment affected hn operation: %v", err)
	}
	if err := runReplicationProcessHook("after-fetch"); err != nil {
		t.Fatalf("legacy replication environment affected hn operation: %v", err)
	}
}

func TestNamespaceCutoverRejectsLegacyWireVersions(t *testing.T) {
	identity := testIdentity(t, "Alice")
	event := newEvent(identity, "issue.open", 1, "")
	event.Protocol = "nh/0" // Intentional hostile legacy input.
	event.Title = "legacy"
	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, err := identity.privateKey()
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = verifyEvent(payload, ed25519.Sign(privateKey, payload))
	if err == nil || !strings.Contains(err.Error(), `unsupported protocol "nh/0"`) {
		t.Fatalf("legacy collaboration protocol validation = %v", err)
	}

	memory := validMemoryEnvelopeFixture(memoryOperationRecord)
	memory.Protocol = "nh-memory/0" // Intentional hostile legacy input.
	if err := validateMemoryEnvelope(memory); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("legacy memory protocol validation = %v", err)
	}
	if err := validatePolicy(PolicyDocument{Version: "nh.policy/0"}); err == nil || !strings.Contains(err.Error(), "unsupported policy version") {
		t.Fatalf("legacy policy protocol validation = %v", err)
	}
	if err := validatePipeline(PipelineDefinition{Version: "nh.pipeline/0"}); err == nil || !strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("legacy pipeline protocol validation = %v", err)
	}
}

func TestNamespaceCutoverIgnoresLegacyRefsAndRepositoryConfig(t *testing.T) {
	root := inIdentityTestRepository(t)
	actor := strings.Repeat("a", 64)
	event := strings.Repeat("b", 64)
	stream := strings.Repeat("c", 64)
	legacyRefs := []string{
		"refs/nh/actors/" + actor, // Intentional hostile legacy refs.
		"refs/nh/proposals/" + event,
		"refs/nh/memory/" + actor + "/" + stream,
	}

	legacyRoot := filepath.Join(root, ".nh") // Intentional frozen legacy config.
	if err := os.MkdirAll(filepath.Join(legacyRoot, "pipelines"), 0o755); err != nil {
		t.Fatal(err)
	}
	legacyPolicy := []byte(`{"version":"nh.policy/0"}` + "\n")
	legacyPipeline := []byte(`{"version":"nh.pipeline/0","steps":[{"name":"test","command":"go"}]}` + "\n")
	policyPath := filepath.Join(legacyRoot, "policy.json")
	pipelinePath := filepath.Join(legacyRoot, "pipelines", "test.json")
	if err := os.WriteFile(policyPath, legacyPolicy, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pipelinePath, legacyPipeline, 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", ".nh")
	mustGit(t, "commit", "-q", "-m", "legacy-only state")
	for _, ref := range legacyRefs {
		mustGit(t, "update-ref", ref, "HEAD")
	}

	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("legacy actor/proposal refs became active events: %#v", events)
	}
	gitDir := filepath.Join(root, ".git")
	memorySources, err := collectMemoryIndexSourcesAt(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(memorySources) != 0 {
		t.Fatalf("legacy memory refs became active sources: %#v", memorySources)
	}
	if _, _, _, err := loadPolicy("HEAD"); err == nil || !strings.Contains(err.Error(), "no .hn/policy.json") {
		t.Fatalf("legacy-only policy lookup = %v", err)
	}
	if _, _, _, err := loadPipeline("HEAD", "test"); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("legacy-only pipeline lookup = %v", err)
	}

	for path, want := range map[string][]byte{policyPath: legacyPolicy, pipelinePath: legacyPipeline} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Fatalf("legacy config %s was modified", path)
		}
	}
	for _, ref := range legacyRefs {
		if got := mustGitText(t, "rev-parse", ref); got != mustGitText(t, "rev-parse", "HEAD") {
			t.Fatalf("legacy ref %s changed to %s", ref, got)
		}
	}
}

func TestReplicationNamespaceIgnoresLegacyRemoteAdvertisements(t *testing.T) {
	inIdentityTestRepository(t)
	actor := strings.Repeat("a", 64)
	activeRef := "refs/hn/actors/" + actor
	legacyRef := "refs/nh/actors/" + actor // Intentional colliding legacy advertisement.
	remote := testAdvertisedRemote(t, map[string]string{
		activeRef: "HEAD",
		legacyRef: "HEAD",
	})
	selection := ReplicationSelection{
		Version: replicationSelectionVersion,
		Remote:  "origin",
		All:     true,
		Budgets: defaultReplicationBudgets(),
	}
	requests, outcomes, err := resolveReplicationRequests(selection, remote)
	if err != nil {
		t.Fatal(err)
	}
	if len(outcomes) != 0 {
		t.Fatalf("mixed advertisement outcomes = %#v", outcomes)
	}
	if len(requests) != 1 || requests[0].SourceRef != activeRef {
		t.Fatalf("mixed advertisement requests = %#v, want only %s", requests, activeRef)
	}
	advertised := mustGitText(t, "ls-remote", "--refs", remote, legacyRef)
	if !strings.Contains(advertised, legacyRef) {
		t.Fatalf("legacy advertisement was unexpectedly mutated: %q", advertised)
	}
}

func TestNamespaceCutoverMergeCommitUsesHNLabel(t *testing.T) {
	root := inIdentityTestRepository(t)
	maintainer, _, err := createIdentity("Maintainer")
	if err != nil {
		t.Fatal(err)
	}
	writeTestPolicy(t, root, PolicyDocument{
		Version:     policyVersion,
		Maintainers: []string{maintainer.Actor},
		Proposals: ProposalPolicy{
			RequiredApprovals: 0,
			RequiredAccepts:   1,
		},
		Pipelines: map[string]PipelinePolicy{},
	})
	mustGit(t, "add", ".hn/policy.json")
	mustGit(t, "commit", "-q", "-m", "active policy")
	base := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "switch", "-q", "-c", "candidate")
	if err := os.WriteFile(filepath.Join(root, "candidate.txt"), []byte("candidate\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", "candidate.txt")
	mustGit(t, "commit", "-q", "-m", "candidate")
	head := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "switch", "-q", "main")

	proposalEvent, err := nextEvent(maintainer, "proposal.open")
	if err != nil {
		t.Fatal(err)
	}
	proposalEvent.Title = "Namespace label"
	proposalEvent.Base = base
	proposalEvent.Head = head
	proposal, err := appendEvent(proposalEvent, maintainer)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(proposal.ID, head); err != nil {
		t.Fatal(err)
	}
	_, _, policyDigest, err := loadPolicy(base)
	if err != nil {
		t.Fatal(err)
	}
	appendRevisionDecision(t, maintainer, proposal.ID, policyDigest, nil)
	if _, err := captureTestOutput(t, func() error { return cmdMerge([]string{proposal.ID}) }); err != nil {
		t.Fatal(err)
	}
	subject := mustGitText(t, "show", "-s", "--format=%s", "HEAD")
	want := "Merge HN proposal " + shortID(proposal.ID) + ": Namespace label"
	if subject != want {
		t.Fatalf("merge subject = %q, want %q", subject, want)
	}
	if strings.Contains(subject, " NH ") {
		t.Fatalf("merge subject retained legacy label: %q", subject)
	}
}

func TestReplicationNamespaceMixedRemoteSyncIsolatesLegacyState(t *testing.T) {
	root := t.TempDir()
	publisher := filepath.Join(root, "publisher")
	remote := filepath.Join(root, "project.git")
	receiver := filepath.Join(root, "receiver")
	mustGit(t, "init", "-q", "-b", "main", publisher)
	mustGit(t, "-C", publisher, "config", "user.name", "Publisher")
	mustGit(t, "-C", publisher, "config", "user.email", "publisher@hn.invalid")
	mustGit(t, "-C", publisher, "commit", "--allow-empty", "-q", "-m", "base")
	initBareMainRemote(t, remote)

	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		replicationAfterFetchHook = nil
		_ = os.Chdir(original)
	})
	if err := os.Chdir(publisher); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "remote", "add", "origin", remote)
	mustGit(t, "push", "-q", "origin", "main:main")
	remoteIdentity := testIdentity(t, "Remote Actor")
	remoteEvent := newEvent(remoteIdentity, "issue.open", 1, "")
	remoteEvent.Title = "active remote event"
	remoteStored, err := appendEvent(remoteEvent, remoteIdentity)
	if err != nil {
		t.Fatal(err)
	}
	activeRemoteRef := actorRef(remoteIdentity.Actor)
	legacyRemoteRefs := []string{
		"refs/nh/actors/" + remoteIdentity.Actor,
		"refs/nh/proposals/" + strings.TrimPrefix(remoteStored.ID, "sha256:"),
		"refs/nh/memory/" + remoteIdentity.Actor + "/" + strings.TrimPrefix(defaultMemoryStream(remoteIdentity.Actor), "sha256:"),
	}
	mustGit(t, "push", "-q", "origin", activeRemoteRef+":"+activeRemoteRef)
	for _, ref := range legacyRemoteRefs {
		mustGit(t, "push", "-q", "origin", "HEAD:"+ref)
	}
	mustGit(t, "clone", "-q", remote, receiver)

	if err := os.Chdir(receiver); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "config", "user.name", "Receiver")
	mustGit(t, "config", "user.email", "receiver@hn.invalid")
	legacyLocalRefs := []string{
		"refs/nh/actors/" + remoteIdentity.Actor,
		"refs/nh/proposals/" + strings.TrimPrefix(remoteStored.ID, "sha256:"),
		"refs/nh/memory/" + remoteIdentity.Actor + "/" + strings.TrimPrefix(defaultMemoryStream(remoteIdentity.Actor), "sha256:"),
	}
	for _, ref := range legacyLocalRefs {
		mustGit(t, "update-ref", ref, "HEAD")
	}
	legacyRoot := filepath.Join(receiver, ".git", "nh")
	seedLegacyPrivateTree(t, legacyRoot, remoteIdentity)
	legacyTreeBefore := snapshotLegacyPrivateTree(t, legacyRoot)
	localLegacyRefsBefore := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh")
	remoteLegacyRefsBefore := mustGitText(t, "ls-remote", "--refs", remote, "refs/nh/*")

	if err := cmdIdentity([]string{"show"}); err == nil || !strings.Contains(err.Error(), "run 'hn init'") {
		t.Fatalf("legacy-only identity command = %v", err)
	}
	if err := cmdMemory([]string{"index", "verify"}); err == nil {
		t.Fatal("legacy memory index was accepted as the active index")
	}
	recallOutput, err := captureTestOutput(t, func() error { return cmdMemory([]string{"recall"}) })
	if err != nil {
		t.Fatalf("legacy-only memory recall: %v", err)
	}
	if !strings.Contains(recallOutput, "Matched: 0; returned: 0") {
		t.Fatalf("legacy memory index affected recall:\n%s", recallOutput)
	}

	localIdentity, _, err := createIdentity("Local Actor")
	if err != nil {
		t.Fatal(err)
	}
	localEvent, err := nextEvent(localIdentity, "issue.open")
	if err != nil {
		t.Fatal(err)
	}
	localEvent.Title = "publish active local event"
	localStored, err := appendEvent(localEvent, localIdentity)
	if err != nil {
		t.Fatal(err)
	}
	if err := cmdReplication([]string{"select", "origin", "--actor", remoteIdentity.Actor}); err != nil {
		t.Fatal(err)
	}
	if _, err := captureTestOutput(t, func() error { return cmdReplication([]string{"show", "origin"}) }); err != nil {
		t.Fatal(err)
	}

	reachedQuarantineBoundary := false
	replicationAfterFetchHook = func() error {
		reachedQuarantineBoundary = true
		if got := snapshotLegacyPrivateTree(t, legacyRoot); !reflect.DeepEqual(got, legacyTreeBefore) {
			t.Fatalf("legacy private tree changed during quarantine: before=%#v after=%#v", legacyTreeBefore, got)
		}
		return nil
	}
	output, err := captureTestOutput(t, func() error { return cmdSync([]string{"origin"}) })
	if err != nil {
		t.Fatalf("mixed-remote sync: %v\n%s", err, output)
	}
	if !reachedQuarantineBoundary {
		t.Fatal("sync did not exercise the quarantine transaction")
	}
	assertRefValue(t, acceptedActorRef("origin", remoteIdentity.Actor), remoteStored.Commit)
	if got := mustGitText(t, "ls-remote", "--refs", remote, actorRef(localIdentity.Actor)); !strings.Contains(got, localStored.Commit+"\t"+actorRef(localIdentity.Actor)) {
		t.Fatalf("active local actor was not published: %q", got)
	}
	if got := mustGitText(t, "for-each-ref", "--format=%(refname) %(objectname)", "refs/nh"); got != localLegacyRefsBefore {
		t.Fatalf("legacy local refs changed:\nbefore %s\nafter  %s", localLegacyRefsBefore, got)
	}
	if got := mustGitText(t, "ls-remote", "--refs", remote, "refs/nh/*"); got != remoteLegacyRefsBefore {
		t.Fatalf("legacy remote refs changed:\nbefore %s\nafter  %s", remoteLegacyRefsBefore, got)
	}
	if got := snapshotLegacyPrivateTree(t, legacyRoot); !reflect.DeepEqual(got, legacyTreeBefore) {
		t.Fatalf("legacy private tree changed after sync: before=%#v after=%#v", legacyTreeBefore, got)
	}
	if refs := mustGitText(t, "for-each-ref", "--format=%(refname)", "refs/nh/remotes"); refs != "" {
		t.Fatalf("sync projected active facts into legacy refs: %s", refs)
	}
	transactions, err := filepath.Glob(filepath.Join(receiver, ".git", "hn", "replication", "transactions", "*.json"))
	if err != nil || len(transactions) != 1 {
		t.Fatalf("active transaction records = %v, err=%v", transactions, err)
	}
	quarantines, err := filepath.Glob(filepath.Join(receiver, ".git", "hn", "replication", "quarantine", "txn-*"))
	if err != nil || len(quarantines) != 0 {
		t.Fatalf("active quarantine cleanup = %v, err=%v", quarantines, err)
	}
}

func seedLegacyPrivateTree(t *testing.T, root string, identity *Identity) {
	t.Helper()
	encodedIdentity, err := json.Marshal(identity)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		"identity.json": encodedIdentity,
		filepath.Join("identities", identity.Actor+".json"): encodedIdentity,
		"active":        []byte(identity.Actor + "\n"),
		"rotation.json": []byte(`{"version":1,"legacy":true}` + "\n"),
		filepath.Join("replication", "selections", "origin.json"):   []byte(`{"version":1,"legacy":true}` + "\n"),
		filepath.Join("replication", "transactions", "legacy.json"): []byte(`{"version":1,"legacy":true}` + "\n"),
		filepath.Join("replication", "anchors", "legacy.json"):      []byte(`{"version":1,"legacy":true}` + "\n"),
		filepath.Join("memory", "index-v0.json"):                    []byte(`{"version":0,"legacy":true}` + "\n"),
	}
	for relative, contents := range files {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func snapshotLegacyPrivateTree(t *testing.T, root string) map[string]string {
	t.Helper()
	snapshot := make(map[string]string)
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		value := info.Mode().String()
		if info.Mode().IsRegular() {
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += "\x00" + string(contents)
		}
		snapshot[relative] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestCLINamespaceUsesHNOnly(t *testing.T) {
	output := captureStdout(t, printUsage)
	if !strings.Contains(output, "Usage:\n  hn init") {
		t.Fatalf("root help does not advertise hn:\n%s", output)
	}
	if strings.Contains(output, "\n  nh ") {
		t.Fatalf("root help advertises legacy nh command:\n%s", output)
	}
	err := run([]string{"not-a-command"})
	if err == nil || !strings.Contains(err.Error(), "run 'hn help'") {
		t.Fatalf("unknown-command guidance = %v, want hn help", err)
	}
	if strings.Contains(err.Error(), "nh help") {
		t.Fatalf("unknown-command guidance contains legacy command: %v", err)
	}
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = writer
	run()
	os.Stdout = previous
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	contents, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(contents)
}
