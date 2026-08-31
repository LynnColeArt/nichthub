package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func projectionMemory(t *testing.T, identity *Identity, operation, target string, record *MemoryRecord) StoredMemory {
	t.Helper()
	envelope := newMemoryEnvelope(identity, operation, defaultMemoryStream(identity.Actor), 1, "")
	envelope.Timestamp = "2026-08-30T12:34:56Z"
	envelope.Target = target
	envelope.Record = record
	switch operation {
	case memoryOperationRetract:
		envelope.Reason = "incorrect"
	case memoryOperationChallenge:
		envelope.Reason = "evidence-mismatch"
		envelope.Evidence = []string{}
	}
	payload, signature, err := encodeAndSignMemory(envelope, identity)
	if err != nil {
		t.Fatal(err)
	}
	return StoredMemory{ID: memoryID(payload), Envelope: envelope, Payload: payload, Signature: signature}
}

func projectionRecord(t *testing.T, identity *Identity, kind, content string) StoredMemory {
	t.Helper()
	record := validMemoryRecordFixture(kind)
	record.Content = content
	record.Anchor.Paths = nil
	record.Applicability = Applicability{Mode: memoryApplicabilityExact}
	return projectionMemory(t, identity, memoryOperationRecord, "", &record)
}

func projectionLifecycle(t *testing.T, identity *Identity, operation, target, content string) StoredMemory {
	t.Helper()
	var record *MemoryRecord
	if operation == memoryOperationSupersede {
		value := validMemoryRecordFixture(memoryKindDecision)
		value.Content = content
		value.Anchor.Paths = nil
		value.Applicability = Applicability{Mode: memoryApplicabilityExact}
		record = &value
	}
	return projectionMemory(t, identity, operation, target, record)
}

func TestMemoryProjectionRelationshipValidation(t *testing.T) {
	alice := deterministicMemoryIdentity()
	bob := testIdentity(t, "Bob")
	target := projectionRecord(t, alice, memoryKindDecision, "original")
	successor := projectionLifecycle(t, alice, memoryOperationSupersede, target.ID, "replacement")
	challenge := projectionLifecycle(t, bob, memoryOperationChallenge, target.ID, "")
	wrongAuthor := projectionLifecycle(t, bob, memoryOperationRetract, target.ID, "")
	missingID := fullMemoryID("f")
	missing := projectionLifecycle(t, alice, memoryOperationRetract, missingID, "")

	got := ProjectMemories([]StoredMemory{missing, wrongAuthor, challenge, successor, target}, MemoryProjectionContext{})
	if len(got.Relationships) != 2 {
		t.Fatalf("relationships = %#v, want supersede and challenge", got.Relationships)
	}
	gap, ok := findMemoryDependency(got.MissingDependencies, "lifecycle-target", missingID)
	if !ok || gap.OwnerID != missing.ID {
		t.Fatalf("missing dependencies = %#v", got.MissingDependencies)
	}
	if len(got.Diagnostics) != 1 || got.Diagnostics[0].MemoryID != wrongAuthor.ID || got.Diagnostics[0].Code != "actor-mismatch" {
		t.Fatalf("diagnostics = %#v", got.Diagnostics)
	}
	if len(got.Rows) != 2 {
		t.Fatalf("rows = %d, unrelated record was lost", len(got.Rows))
	}
}

func TestMemoryProjectionRelationshipWrongKindAndSelfAreInvalid(t *testing.T) {
	alice := deterministicMemoryIdentity()
	target := projectionRecord(t, alice, memoryKindDecision, "target")
	lifecycle := projectionLifecycle(t, alice, memoryOperationRetract, target.ID, "")
	self := projectionLifecycle(t, alice, memoryOperationRetract, fullMemoryID("a"), "")
	self.ID = self.Envelope.Target
	bad := projectionLifecycle(t, alice, memoryOperationChallenge, lifecycle.ID, "")

	got := ProjectMemories([]StoredMemory{target, lifecycle, self, bad}, MemoryProjectionContext{})
	if _, ok := findMemoryDependency(got.MissingDependencies, "lifecycle-target", lifecycle.ID); ok {
		t.Fatalf("invalid relationship became lifecycle gap: %#v", got.MissingDependencies)
	}
	codes := []string{got.Diagnostics[0].Code, got.Diagnostics[1].Code}
	slices.Sort(codes)
	if strings.Join(codes, ",") != "self-target,wrong-target-kind" {
		t.Fatalf("diagnostic codes = %v", codes)
	}
}

func TestMemoryProjectionLifecycleConvergesWithBranchingRetractionAndChallenge(t *testing.T) {
	alice := deterministicMemoryIdentity()
	bob := testIdentity(t, "Bob")
	carol := testIdentity(t, "Carol")
	target := projectionRecord(t, alice, memoryKindDecision, "target")
	first := projectionLifecycle(t, alice, memoryOperationSupersede, target.ID, "first")
	second := projectionLifecycle(t, alice, memoryOperationSupersede, target.ID, "second")
	retract := projectionLifecycle(t, alice, memoryOperationRetract, target.ID, "")
	challenge1 := projectionLifecycle(t, bob, memoryOperationChallenge, target.ID, "")
	challenge2 := projectionLifecycle(t, carol, memoryOperationChallenge, target.ID, "")
	input := []StoredMemory{target, first, second, retract, challenge1, challenge2}

	want := ProjectMemories(input, MemoryProjectionContext{})
	reversed := slices.Clone(input)
	slices.Reverse(reversed)
	got := ProjectMemories(reversed, MemoryProjectionContext{})
	wantJSON, _ := json.Marshal(want)
	gotJSON, _ := json.Marshal(got)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("projection depends on arrival order:\n%s\n%s", wantJSON, gotJSON)
	}
	row := findProjectionRow(t, want.Rows, target.ID)
	if row.Lifecycle != memoryLifecycleRetracted || len(row.Successors) != 2 || len(row.Retractions) != 1 || len(row.Challengers) != 2 {
		t.Fatalf("target row = %#v", row)
	}
	if findProjectionRow(t, want.Rows, first.ID).Lifecycle != memoryLifecycleActive || findProjectionRow(t, want.Rows, second.ID).Lifecycle != memoryLifecycleActive {
		t.Fatal("successors were not independently active")
	}
}

func TestMemoryProjectionLifecycleMissingSupersessionIsExplicit(t *testing.T) {
	alice := deterministicMemoryIdentity()
	successor := projectionLifecycle(t, alice, memoryOperationSupersede, fullMemoryID("f"), "replacement")
	got := ProjectMemories([]StoredMemory{successor}, MemoryProjectionContext{})
	row := findProjectionRow(t, got.Rows, successor.ID)
	if _, ok := findMemoryDependency(got.MissingDependencies, "lifecycle-target", fullMemoryID("f")); row.Lifecycle != memoryLifecycleDependencyMissing || !ok {
		t.Fatalf("row=%#v missing=%#v", row, got.MissingDependencies)
	}
}

func TestMemoryProjectionLifecycleBranchingSummary(t *testing.T) {
	alice := deterministicMemoryIdentity()
	target := projectionRecord(t, alice, memoryKindDecision, "target")
	left := projectionLifecycle(t, alice, memoryOperationSupersede, target.ID, "left")
	right := projectionLifecycle(t, alice, memoryOperationSupersede, target.ID, "right")
	row := findProjectionRow(t, ProjectMemories([]StoredMemory{right, target, left}, MemoryProjectionContext{}).Rows, target.ID)
	if row.Lifecycle != memoryLifecycleBranching || len(row.Successors) != 2 {
		t.Fatalf("branching row = %#v", row)
	}
}

func TestMemoryProjectionApplicabilityAndEvidence(t *testing.T) {
	withMemoryRepository(t, func() {
		mustWriteFile(t, "tracked.txt", []byte("one\n"))
		mustGit(t, "add", "tracked.txt")
		mustGit(t, "commit", "-q", "-m", "anchor")
		anchor := mustGitText(t, "rev-parse", "HEAD")
		blob := mustGitText(t, "rev-parse", "HEAD:tracked.txt")
		mustWriteFile(t, "tracked.txt", []byte("two\n"))
		mustGit(t, "add", "tracked.txt")
		mustGit(t, "commit", "-q", "-m", "child")
		child := mustGitText(t, "rev-parse", "HEAD")

		alice := deterministicMemoryIdentity()
		eventID := fullMemoryID("e")
		memoryEvidence := projectionRecord(t, alice, memoryKindObservation, "evidence memory")
		record := validMemoryRecordFixture(memoryKindVerification)
		record.Anchor = MemoryAnchor{Commit: anchor, Paths: []PathAnchor{{Path: "gone.txt", Blob: "absent"}, {Path: "tracked.txt", Blob: blob}}}
		record.Applicability = Applicability{Mode: memoryApplicabilityDescendants}
		record.Evidence = []string{"event:" + eventID, "git:" + blob, "memory:" + memoryEvidence.ID, "memory:" + fullMemoryID("f")}
		stored := projectionMemory(t, alice, memoryOperationRecord, "", &record)
		result := ProjectMemories([]StoredMemory{stored, memoryEvidence}, MemoryProjectionContext{
			AtCommit: child,
			Events:   []StoredEvent{{ID: eventID}},
		})
		row := findProjectionRow(t, result.Rows, stored.ID)
		if row.Applicability != memoryApplicabilityApplicable || row.Evidence != memoryEvidenceMissing {
			t.Fatalf("row = %#v", row)
		}
		if len(row.EvidenceDetails) != 4 || row.EvidenceDetails[3].Status != memoryEvidenceMissing {
			t.Fatalf("evidence details = %#v", row.EvidenceDetails)
		}

		result = ProjectMemories([]StoredMemory{stored, memoryEvidence}, MemoryProjectionContext{AtCommit: anchor, Path: "other.txt"})
		if got := findProjectionRow(t, result.Rows, stored.ID).Applicability; got != memoryApplicabilityInapplicable {
			t.Fatalf("path filtered applicability = %q", got)
		}
	})
}

func TestMemoryProjectionApplicabilityMissingAndInvalidAnchors(t *testing.T) {
	alice := deterministicMemoryIdentity()
	record := validMemoryRecordFixture(memoryKindDecision)
	record.Anchor.Paths = nil
	record.Applicability = Applicability{Mode: memoryApplicabilityExact}
	missing := projectionMemory(t, alice, memoryOperationRecord, "", &record)
	resolver := &projectionResolverStub{probes: map[string]gitObjectProbe{
		record.Anchor.Commit: {Exists: false},
	}}
	result := ProjectMemories([]StoredMemory{missing}, MemoryProjectionContext{AtCommit: record.Anchor.Commit, Resolver: resolver})
	if got := result.Rows[0].Applicability; got != memoryApplicabilityAnchorMissing {
		t.Fatalf("missing anchor = %q", got)
	}
	resolver.probes[record.Anchor.Commit] = gitObjectProbe{Exists: true, Type: "blob"}
	result = ProjectMemories([]StoredMemory{missing}, MemoryProjectionContext{AtCommit: record.Anchor.Commit, Resolver: resolver})
	if got := result.Rows[0].Applicability; got != memoryApplicabilityAnchorInvalid {
		t.Fatalf("wrong-type anchor = %q", got)
	}
}

func TestMemoryProjectionTrustClassificationIsOrthogonal(t *testing.T) {
	alice := deterministicMemoryIdentity()
	bob := testIdentity(t, "Bob")
	trusted := projectionRecord(t, alice, memoryKindDecision, "trusted")
	untrustedActor := projectionRecord(t, bob, memoryKindDecision, "actor")
	untrustedKind := projectionRecord(t, alice, memoryKindAssumption, "kind")
	policy := &MemoryPolicy{TrustedActors: []string{alice.Actor}, TrustedKinds: []string{memoryKindDecision}}
	result := ProjectMemories([]StoredMemory{trusted, untrustedActor, untrustedKind}, MemoryProjectionContext{MemoryPolicy: policy, PolicyDigest: fullMemoryID("d")})
	if findProjectionRow(t, result.Rows, trusted.ID).Trust != memoryTrustQualified ||
		findProjectionRow(t, result.Rows, untrustedActor.ID).Trust != memoryTrustActorUntrusted ||
		findProjectionRow(t, result.Rows, untrustedKind.ID).Trust != memoryTrustKindUntrusted {
		t.Fatalf("unexpected trust projection: %#v", result.Rows)
	}
	legacy := ProjectMemories([]StoredMemory{trusted}, MemoryProjectionContext{})
	if legacy.Rows[0].Trust != memoryTrustPolicyMissing || legacy.Rows[0].Signature != "valid" {
		t.Fatalf("legacy row = %#v", legacy.Rows[0])
	}
}

func TestMemoryProjectionCompletePermutation(t *testing.T) {
	alice := deterministicMemoryIdentity()
	target := projectionRecord(t, alice, memoryKindDecision, "target")
	successor := projectionLifecycle(t, alice, memoryOperationSupersede, target.ID, "replacement")
	missing := projectionLifecycle(t, alice, memoryOperationChallenge, fullMemoryID("f"), "")
	input := []StoredMemory{target, successor, missing}
	context := MemoryProjectionContext{MemoryPolicy: &MemoryPolicy{TrustedActors: []string{alice.Actor}, TrustedKinds: []string{memoryKindDecision}}}
	want, _ := json.Marshal(ProjectMemories(input, context))
	for _, permutation := range [][]StoredMemory{{missing, target, successor}, {successor, missing, target}, {target, successor, missing, successor}} {
		got, _ := json.Marshal(ProjectMemories(permutation, context))
		if string(got) != string(want) {
			t.Fatalf("complete projection differs by permutation:\n%s\n%s", want, got)
		}
	}
}

func findProjectionRow(t *testing.T, rows []MemoryProjectionRow, id string) MemoryProjectionRow {
	t.Helper()
	for _, row := range rows {
		if row.ID == id {
			return row
		}
	}
	t.Fatalf("missing projection row %s", id)
	return MemoryProjectionRow{}
}

func findMemoryDependency(dependencies []MemoryDependency, kind, id string) (MemoryDependency, bool) {
	for _, dependency := range dependencies {
		if dependency.Kind == kind && dependency.MissingID == id {
			return dependency, true
		}
	}
	return MemoryDependency{}, false
}

type projectionResolverStub struct {
	probes map[string]gitObjectProbe
}

func (stub *projectionResolverStub) Probe(object string) (gitObjectProbe, error) {
	return stub.probes[object], nil
}

func (stub *projectionResolverStub) IsAncestor(ancestor, descendant string) (bool, string, error) {
	return ancestor == descendant, "", nil
}

func (stub *projectionResolverStub) TreeEntry(commit, path string) (string, bool, error) {
	return "", false, nil
}

func TestMemoryPolicyExactCommitLoad(t *testing.T) {
	withMemoryRepository(t, func() {
		alice := deterministicMemoryIdentity()
		bob := testIdentity(t, "Bob")
		writeProjectionPolicy(t, alice, []string{alice.Actor}, []string{memoryKindDecision})
		mustGit(t, "add", ".nh/policy.json")
		mustGit(t, "commit", "-q", "-m", "alice memory policy")
		alicePolicyCommit := mustGitText(t, "rev-parse", "HEAD")
		writeProjectionPolicy(t, alice, []string{bob.Actor}, []string{memoryKindObservation})
		mustGit(t, "add", ".nh/policy.json")
		mustGit(t, "commit", "-q", "-m", "bob memory policy")

		context, err := LoadMemoryProjectionPolicy(alicePolicyCommit, MemoryProjectionContext{})
		if err != nil {
			t.Fatal(err)
		}
		if context.MemoryPolicy == nil || !actorListed(alice.Actor, context.MemoryPolicy.TrustedActors) || context.PolicyDigest == "" || context.PolicyCommit != alicePolicyCommit {
			t.Fatalf("loaded exact policy context = %#v", context)
		}
	})
}

func writeProjectionPolicy(t *testing.T, maintainer *Identity, actors, kinds []string) {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	writeTestPolicy(t, root, PolicyDocument{
		Version:     policyVersion,
		Maintainers: []string{maintainer.Actor},
		Proposals:   ProposalPolicy{RequiredAccepts: 1},
		Pipelines:   map[string]PipelinePolicy{},
		Memory:      &MemoryPolicy{TrustedActors: actors, TrustedKinds: kinds},
	})
}

func mustWriteFile(t *testing.T, name string, contents []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}
