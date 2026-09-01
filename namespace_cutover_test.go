package main

import (
	"crypto/ed25519"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
