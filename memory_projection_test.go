package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func projectionMemory(t *testing.T, identity *Identity, operation, target string, record *MemoryRecord) StoredMemory {
	t.Helper()
	return projectionMemoryAt(t, identity, operation, target, record, defaultMemoryStream(identity.Actor), "2026-08-30T12:34:56Z", nil)
}

func projectionMemoryAt(t *testing.T, identity *Identity, operation, target string, record *MemoryRecord, stream, timestamp string, evidence []string) StoredMemory {
	t.Helper()
	envelope := newMemoryEnvelope(identity, operation, stream, 1, "")
	envelope.Timestamp = timestamp
	envelope.Target = target
	envelope.Record = record
	switch operation {
	case memoryOperationRetract:
		envelope.Reason = "incorrect"
	case memoryOperationChallenge:
		envelope.Reason = "evidence-mismatch"
		envelope.Evidence = append([]string(nil), evidence...)
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

func TestMemoryProjectionRelationshipChallengeMustBeCrossAuthor(t *testing.T) {
	alice := deterministicMemoryIdentity()
	bob := testIdentity(t, "Bob")
	target := projectionRecord(t, alice, memoryKindDecision, "target")
	unrelated := projectionRecord(t, bob, memoryKindObservation, "unrelated")
	sameAuthor := projectionMemoryAt(t, alice, memoryOperationChallenge, target.ID, nil, fullMemoryID("1"), "2026-08-30T00:00:01Z", []string{"memory:" + unrelated.ID})
	crossAuthor := projectionMemoryAt(t, bob, memoryOperationChallenge, target.ID, nil, fullMemoryID("2"), "2026-08-30T00:00:00Z", []string{"memory:" + unrelated.ID})

	context := MemoryProjectionContext{MemoryPolicy: &MemoryPolicy{TrustedActors: []string{alice.Actor}, TrustedKinds: []string{memoryKindDecision}}}
	before := findProjectionRow(t, ProjectMemories([]StoredMemory{unrelated, target}, context).Rows, target.ID)
	result := ProjectMemories([]StoredMemory{sameAuthor, unrelated, target, crossAuthor}, context)
	row := findProjectionRow(t, result.Rows, target.ID)
	if len(row.Challengers) != 1 || row.Challengers[0] != crossAuthor.ID {
		t.Fatalf("challengers = %v, want only cross-author challenge %s", row.Challengers, crossAuthor.ID)
	}
	if row.Lifecycle != before.Lifecycle || row.Trust != before.Trust || row.Evidence != before.Evidence || row.Applicability != before.Applicability || row.Signature != before.Signature {
		t.Fatalf("challenge changed an orthogonal classification: %#v", row)
	}
	diagnostic, ok := findProjectionDiagnostic(result.Diagnostics, sameAuthor.ID)
	if !ok || diagnostic.Code != "challenge-actor-rule" || diagnostic.TargetID != target.ID {
		t.Fatalf("same-author challenge diagnostic = %#v", diagnostic)
	}
	if len(result.Rows) != 2 || findProjectionRow(t, result.Rows, unrelated.ID).ID != unrelated.ID {
		t.Fatalf("same-author challenge suppressed target or unrelated row: %#v", result.Rows)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), `"truth"`) || strings.Contains(string(encoded), `"authorization"`) {
		t.Fatalf("projection manufactured truth or authorization semantics: %s", encoded)
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

func TestMemoryProjectionLifecycleDependencyPrecedencePreservesAllEdges(t *testing.T) {
	alice := deterministicMemoryIdentity()
	bob := testIdentity(t, "Bob")
	carol := testIdentity(t, "Carol")
	anchor := strings.Repeat("a", 40)
	source := projectionSupersessionAt(t, alice, fullMemoryID("f"), "replacement with unavailable target", fullMemoryID("1"), "2026-08-30T06:00:00Z", anchor)
	left := projectionSupersessionAt(t, alice, source.ID, "left", fullMemoryID("2"), "2026-08-30T05:00:00Z", anchor)
	right := projectionSupersessionAt(t, alice, source.ID, "right", fullMemoryID("3"), "2026-08-30T04:00:00Z", anchor)
	retract := projectionMemoryAt(t, alice, memoryOperationRetract, source.ID, nil, fullMemoryID("4"), "2026-08-30T03:00:00Z", nil)
	challenge1 := projectionMemoryAt(t, bob, memoryOperationChallenge, source.ID, nil, fullMemoryID("5"), "2026-08-30T02:00:00Z", nil)
	challenge2 := projectionMemoryAt(t, carol, memoryOperationChallenge, source.ID, nil, fullMemoryID("6"), "2026-08-30T01:00:00Z", nil)
	projection := ProjectMemories([]StoredMemory{right, challenge2, source, retract, left, challenge1}, MemoryProjectionContext{})
	row := findProjectionRow(t, projection.Rows, source.ID)
	if row.Lifecycle != memoryLifecycleDependencyMissing || len(row.Successors) != 2 || len(row.Retractions) != 1 || len(row.Challengers) != 2 {
		t.Fatalf("dependency precedence erased graph facts: %#v", row)
	}
	if _, ok := findMemoryDependency(projection.MissingDependencies, "lifecycle-target", fullMemoryID("f")); !ok {
		t.Fatalf("missing target fact absent: %#v", projection.MissingDependencies)
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

func TestMemoryProjectionApplicabilityProductionGitMatrix(t *testing.T) {
	withMemoryRepository(t, func() {
		mustWriteFile(t, "tracked.txt", []byte("anchor\n"))
		mustGit(t, "add", "tracked.txt")
		mustGit(t, "commit", "-q", "-m", "anchor")
		anchor := mustGitText(t, "rev-parse", "HEAD")
		anchorTree := mustGitText(t, "rev-parse", anchor+"^{tree}")
		blob := mustGitText(t, "rev-parse", anchor+":tracked.txt")
		unrelated := mustGitText(t, "commit-tree", anchorTree, "-m", "unrelated")

		mustWriteFile(t, "tracked.txt", []byte("changed\n"))
		mustWriteFile(t, "later.txt", []byte("now present\n"))
		mustGit(t, "add", "tracked.txt", "later.txt")
		mustGit(t, "commit", "-q", "-m", "child changes paths")
		child := mustGitText(t, "rev-parse", "HEAD")
		mustGit(t, "mv", "tracked.txt", "renamed.txt")
		mustGit(t, "rm", "-q", "later.txt")
		mustGit(t, "commit", "-q", "-m", "rename and remove later")
		renamed := mustGitText(t, "rev-parse", "HEAD")

		alice := deterministicMemoryIdentity()
		anchored := MemoryAnchor{Commit: anchor, Paths: []PathAnchor{{Path: "later.txt", Blob: "absent"}, {Path: "tracked.txt", Blob: blob}}}
		exact := projectionRecordAt(t, alice, memoryKindDecision, "exact", fullMemoryID("1"), "2026-08-30T01:00:00Z", anchored, Applicability{Mode: memoryApplicabilityExact}, nil)
		descendants := projectionRecordAt(t, alice, memoryKindDecision, "descendants", fullMemoryID("2"), "2026-08-30T02:00:00Z", anchored, Applicability{Mode: memoryApplicabilityDescendants}, nil)
		subjectID := "proposal:" + fullMemoryID("3")
		subject := projectionRecordAt(t, alice, memoryKindDecision, "subject", fullMemoryID("3"), "2026-08-30T03:00:00Z", MemoryAnchor{Commit: anchor, Subject: subjectID}, Applicability{Mode: memoryApplicabilitySubject, Subject: subjectID}, nil)

		cases := []struct {
			name    string
			stored  StoredMemory
			context MemoryProjectionContext
			want    string
		}{
			{"exact anchor", exact, MemoryProjectionContext{AtCommit: anchor}, memoryApplicabilityApplicable},
			{"exact child", exact, MemoryProjectionContext{AtCommit: child}, memoryApplicabilityInapplicable},
			{"descendant child", descendants, MemoryProjectionContext{AtCommit: child}, memoryApplicabilityApplicable},
			{"descendant renamed", descendants, MemoryProjectionContext{AtCommit: renamed, Path: "tracked.txt"}, memoryApplicabilityApplicable},
			{"no rename inference", descendants, MemoryProjectionContext{AtCommit: renamed, Path: "renamed.txt"}, memoryApplicabilityInapplicable},
			{"unrelated", descendants, MemoryProjectionContext{AtCommit: unrelated}, memoryApplicabilityInapplicable},
			{"matching subject", subject, MemoryProjectionContext{AtCommit: renamed, Subject: subjectID}, memoryApplicabilityApplicable},
			{"missing subject", subject, MemoryProjectionContext{AtCommit: renamed}, memoryApplicabilityInapplicable},
			{"wrong subject", subject, MemoryProjectionContext{AtCommit: renamed, Subject: "issue:" + fullMemoryID("4")}, memoryApplicabilityInapplicable},
			{"unavailable query", descendants, MemoryProjectionContext{AtCommit: strings.Repeat("f", 40)}, memoryApplicabilityAnchorMissing},
			{"wrong-type query", exact, MemoryProjectionContext{AtCommit: blob}, memoryApplicabilityAnchorInvalid},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				row := findProjectionRow(t, ProjectMemories([]StoredMemory{test.stored}, test.context).Rows, test.stored.ID)
				if row.Applicability != test.want {
					t.Fatalf("applicability = %q, want %q; row=%#v", row.Applicability, test.want, row)
				}
			})
		}

		wrongBlobAnchor := anchored
		wrongBlobAnchor.Paths = []PathAnchor{{Path: "tracked.txt", Blob: strings.Repeat("d", 40)}}
		wrongBlob := projectionRecordAt(t, alice, memoryKindDecision, "wrong blob", fullMemoryID("5"), "2026-08-30T04:00:00Z", wrongBlobAnchor, Applicability{Mode: memoryApplicabilityExact}, nil)
		if got := findProjectionRow(t, ProjectMemories([]StoredMemory{wrongBlob}, MemoryProjectionContext{AtCommit: anchor}).Rows, wrongBlob.ID).Applicability; got != memoryApplicabilityAnchorInvalid {
			t.Fatalf("wrong blob applicability = %q", got)
		}
		wrongTypeAnchor := projectionRecordAt(t, alice, memoryKindDecision, "wrong anchor type", fullMemoryID("6"), "2026-08-30T05:00:00Z", MemoryAnchor{Commit: blob}, Applicability{Mode: memoryApplicabilityExact}, nil)
		if got := findProjectionRow(t, ProjectMemories([]StoredMemory{wrongTypeAnchor}, MemoryProjectionContext{AtCommit: blob}).Rows, wrongTypeAnchor.ID).Applicability; got != memoryApplicabilityAnchorInvalid {
			t.Fatalf("wrong-type anchor applicability = %q", got)
		}
		missingAnchor := projectionRecordAt(t, alice, memoryKindDecision, "missing anchor", fullMemoryID("7"), "2026-08-30T06:00:00Z", MemoryAnchor{Commit: strings.Repeat("e", 40)}, Applicability{Mode: memoryApplicabilityExact}, nil)
		missingProjection := ProjectMemories([]StoredMemory{missingAnchor}, MemoryProjectionContext{AtCommit: strings.Repeat("e", 40)})
		if got := findProjectionRow(t, missingProjection.Rows, missingAnchor.ID).Applicability; got != memoryApplicabilityAnchorMissing {
			t.Fatalf("missing anchor applicability = %q", got)
		}
		if dependency, ok := findMemoryDependency(missingProjection.MissingDependencies, "anchor-commit", strings.Repeat("e", 40)); !ok || dependency.OwnerID != missingAnchor.ID {
			t.Fatalf("missing anchor dependency = %#v", missingProjection.MissingDependencies)
		}
	})
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

func TestMemoryProjectionEvidenceProductionMatrix(t *testing.T) {
	alice := deterministicMemoryIdentity()
	anchor := strings.Repeat("a", 40)
	gitResolved, gitMissing, gitInvalid := strings.Repeat("b", 40), strings.Repeat("c", 40), strings.Repeat("d", 40)
	eventResolved, eventMissing := fullMemoryID("1"), fullMemoryID("2")
	memoryResolved := projectionRecordAt(t, alice, memoryKindObservation, "support", fullMemoryID("3"), "2026-08-30T01:00:00Z", MemoryAnchor{Commit: anchor}, Applicability{Mode: memoryApplicabilityExact}, nil)
	memoryMissing := fullMemoryID("4")
	record := projectionRecordAt(t, alice, memoryKindDecision, "mixed evidence", fullMemoryID("5"), "2026-08-30T02:00:00Z", MemoryAnchor{Commit: anchor}, Applicability{Mode: memoryApplicabilityExact}, []string{
		"memory:" + memoryMissing,
		"git:" + gitResolved,
		"event:" + eventMissing,
		"memory:" + memoryResolved.ID,
		"git:" + gitInvalid,
		"event:" + eventResolved,
		"git:" + gitMissing,
	})
	resolver := &projectionResolverStub{
		probes: map[string]gitObjectProbe{
			anchor: {Exists: true, Type: "commit"}, gitResolved: {Exists: true, Type: "blob"}, gitMissing: {Exists: false},
		},
		probeErrors: map[string]error{gitInvalid: errors.New("invalid object database response")},
	}
	projection := ProjectMemories([]StoredMemory{record, memoryResolved}, MemoryProjectionContext{AtCommit: anchor, Events: []StoredEvent{{ID: eventResolved}}, Resolver: resolver})
	row := findProjectionRow(t, projection.Rows, record.ID)
	if row.Evidence != memoryEvidenceInvalid || len(row.EvidenceDetails) != 7 {
		t.Fatalf("mixed evidence row = %#v", row)
	}
	for index := 1; index < len(row.EvidenceDetails); index++ {
		left, right := row.EvidenceDetails[index-1], row.EvidenceDetails[index]
		if left.Type > right.Type || (left.Type == right.Type && left.Requested > right.Requested) {
			t.Fatalf("evidence details are not sorted: %#v", row.EvidenceDetails)
		}
	}
	wants := map[string]string{
		"git:" + gitResolved:          memoryEvidenceResolved,
		"git:" + gitMissing:           memoryEvidenceMissing,
		"git:" + gitInvalid:           memoryEvidenceInvalid,
		"event:" + eventResolved:      memoryEvidenceResolved,
		"event:" + eventMissing:       memoryEvidenceMissing,
		"memory:" + memoryResolved.ID: memoryEvidenceResolved,
		"memory:" + memoryMissing:     memoryEvidenceMissing,
	}
	for _, detail := range row.EvidenceDetails {
		if got, want := detail.Status, wants[detail.Type+":"+detail.Requested]; got != want || detail.OwnerID != record.ID || detail.Reason == "" {
			t.Errorf("evidence detail = %#v, want status %q and full owner", detail, want)
		}
	}
	for _, missing := range []struct{ kind, id string }{{"evidence-git", gitMissing}, {"evidence-event", eventMissing}, {"evidence-memory", memoryMissing}} {
		dependency, ok := findMemoryDependency(projection.MissingDependencies, missing.kind, missing.id)
		if !ok || dependency.OwnerID != record.ID || dependency.Stream != record.Envelope.Stream {
			t.Errorf("missing typed dependency %s:%s in %#v", missing.kind, missing.id, projection.MissingDependencies)
		}
	}
	if _, ok := findMemoryDependency(projection.MissingDependencies, "evidence-git", gitInvalid); ok {
		t.Fatal("invalid evidence was mislabeled as recoverable missing data")
	}
}

func TestMemoryProjectionCompletePermutation(t *testing.T) {
	alice := deterministicMemoryIdentity()
	bob := testIdentity(t, "Bob")
	carol := testIdentity(t, "Carol")
	anchor, child, blob := strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("c", 40)
	eventEvidence := fullMemoryID("e")
	evidenceRecord := projectionRecordAt(t, bob, memoryKindObservation, "supporting memory", fullMemoryID("1"), "2026-08-30T00:00:00Z", MemoryAnchor{Commit: anchor}, Applicability{Mode: memoryApplicabilityExact}, nil)
	target := projectionRecordAt(t, alice, memoryKindDecision, "target", fullMemoryID("2"), "2026-08-30T23:59:59Z",
		MemoryAnchor{Commit: anchor, Paths: []PathAnchor{{Path: "gone.txt", Blob: "absent"}, {Path: "src/main.go", Blob: blob}}},
		Applicability{Mode: memoryApplicabilityDescendants}, []string{"memory:" + evidenceRecord.ID, "git:" + blob, "event:" + eventEvidence})
	left := projectionSupersessionAt(t, alice, target.ID, "left branch", fullMemoryID("3"), "2026-08-30T00:00:01Z", anchor)
	right := projectionSupersessionAt(t, alice, target.ID, "right branch", fullMemoryID("4"), "2026-08-30T00:00:01Z", anchor)
	linear := projectionSupersessionAt(t, alice, left.ID, "linear successor", fullMemoryID("5"), "2026-08-29T00:00:00Z", anchor)
	retract := projectionMemoryAt(t, alice, memoryOperationRetract, target.ID, nil, fullMemoryID("6"), "2026-08-28T00:00:00Z", nil)
	challenge1 := projectionMemoryAt(t, bob, memoryOperationChallenge, target.ID, nil, fullMemoryID("7"), "2026-08-27T00:00:00Z", []string{"event:" + eventEvidence})
	challenge2 := projectionMemoryAt(t, carol, memoryOperationChallenge, target.ID, nil, fullMemoryID("8"), "2026-08-26T00:00:00Z", nil)
	missingID := fullMemoryID("f")
	missing := projectionMemoryAt(t, bob, memoryOperationChallenge, missingID, nil, fullMemoryID("9"), "2026-08-25T00:00:00Z", nil)
	input := []StoredMemory{evidenceRecord, target, left, right, linear, retract, challenge1, challenge2, missing}
	actors := []string{alice.Actor, bob.Actor, carol.Actor}
	slices.Sort(actors)
	context := MemoryProjectionContext{
		AtCommit: child, Events: []StoredEvent{{ID: eventEvidence}},
		MemoryPolicy: &MemoryPolicy{TrustedActors: actors, TrustedKinds: []string{memoryKindDecision, memoryKindObservation}},
		PolicyCommit: anchor, PolicyDigest: fullMemoryID("d"),
		Resolver: &projectionResolverStub{
			probes:      map[string]gitObjectProbe{anchor: {Exists: true, Type: "commit"}, child: {Exists: true, Type: "commit"}, blob: {Exists: true, Type: "blob"}},
			ancestors:   map[string]bool{anchor + "\x00" + child: true},
			treeEntries: map[string]string{anchor + "\x00src/main.go": blob},
		},
	}
	want, _ := json.Marshal(ProjectMemories(input, context))
	reversed := slices.Clone(input)
	slices.Reverse(reversed)
	shuffled := []StoredMemory{challenge2, left, missing, evidenceRecord, retract, linear, target, challenge1, right}
	for _, permutation := range [][]StoredMemory{reversed, shuffled, append(slices.Clone(input), right)} {
		got, _ := json.Marshal(ProjectMemories(permutation, context))
		if string(got) != string(want) {
			t.Fatalf("complete projection differs by permutation:\n%s\n%s", want, got)
		}
	}
	projection := ProjectMemories(input, context)
	targetRow := findProjectionRow(t, projection.Rows, target.ID)
	if targetRow.Lifecycle != memoryLifecycleRetracted || len(targetRow.Successors) != 2 || len(targetRow.Retractions) != 1 || len(targetRow.Challengers) != 2 {
		t.Fatalf("precedence erased lifecycle edges: %#v", targetRow)
	}
	if findProjectionRow(t, projection.Rows, left.ID).Lifecycle != memoryLifecycleSuperseded || findProjectionRow(t, projection.Rows, linear.ID).Lifecycle != memoryLifecycleActive {
		t.Fatalf("linear supersession chain was not retained: %#v", projection.Rows)
	}
	if _, ok := findMemoryDependency(projection.MissingDependencies, "lifecycle-target", missingID); !ok {
		t.Fatalf("missing lifecycle edge was not retained: %#v", projection.MissingDependencies)
	}
	if target.Envelope.Stream == left.Envelope.Stream || left.Envelope.Stream == right.Envelope.Stream || left.Envelope.Actor != target.Envelope.Actor {
		t.Fatal("fixture did not prove same-author cross-stream lifecycle edges")
	}
	if left.Envelope.Timestamp != right.Envelope.Timestamp || left.Envelope.Timestamp >= target.Envelope.Timestamp || retract.Envelope.Timestamp >= target.Envelope.Timestamp {
		t.Fatal("fixture timestamps are not equal and deliberately misleading")
	}
}

func projectionRecordAt(t *testing.T, identity *Identity, kind, content, stream, timestamp string, anchor MemoryAnchor, applicability Applicability, evidence []string) StoredMemory {
	t.Helper()
	record := validMemoryRecordFixture(kind)
	record.Content = content
	record.Anchor = anchor
	record.Applicability = applicability
	record.Evidence = append([]string{}, evidence...)
	return projectionMemoryAt(t, identity, memoryOperationRecord, "", &record, stream, timestamp, nil)
}

func projectionSupersessionAt(t *testing.T, identity *Identity, target, content, stream, timestamp, anchor string) StoredMemory {
	t.Helper()
	record := validMemoryRecordFixture(memoryKindDecision)
	record.Content = content
	record.Anchor = MemoryAnchor{Commit: anchor}
	record.Applicability = Applicability{Mode: memoryApplicabilityExact}
	record.Evidence = []string{}
	return projectionMemoryAt(t, identity, memoryOperationSupersede, target, &record, stream, timestamp, nil)
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

func findProjectionDiagnostic(diagnostics []MemoryProjectionDiagnostic, id string) (MemoryProjectionDiagnostic, bool) {
	for _, diagnostic := range diagnostics {
		if diagnostic.MemoryID == id {
			return diagnostic, true
		}
	}
	return MemoryProjectionDiagnostic{}, false
}

type projectionResolverStub struct {
	probes          map[string]gitObjectProbe
	probeErrors     map[string]error
	ancestors       map[string]bool
	ancestorMissing string
	ancestorError   error
	treeEntries     map[string]string
	treeErrors      map[string]error
}

func (stub *projectionResolverStub) Probe(object string) (gitObjectProbe, error) {
	if err := stub.probeErrors[object]; err != nil {
		return gitObjectProbe{}, err
	}
	return stub.probes[object], nil
}

func (stub *projectionResolverStub) IsAncestor(ancestor, descendant string) (bool, string, error) {
	if stub.ancestorError != nil {
		return false, "", stub.ancestorError
	}
	if stub.ancestorMissing != "" {
		return false, stub.ancestorMissing, nil
	}
	if ancestor == descendant {
		return true, "", nil
	}
	return stub.ancestors[ancestor+"\x00"+descendant], "", nil
}

func (stub *projectionResolverStub) TreeEntry(commit, path string) (string, bool, error) {
	key := commit + "\x00" + path
	if err := stub.treeErrors[key]; err != nil {
		return "", false, err
	}
	object, exists := stub.treeEntries[key]
	return object, exists, nil
}

func TestMemoryPolicyExactCommitLoad(t *testing.T) {
	withMemoryRepository(t, func() {
		alice := deterministicMemoryIdentity()
		bob := testIdentity(t, "Bob")
		writeProjectionPolicy(t, alice, []string{alice.Actor}, []string{memoryKindDecision})
		mustGit(t, "add", ".hn/policy.json")
		mustGit(t, "commit", "-q", "-m", "alice memory policy")
		alicePolicyCommit := mustGitText(t, "rev-parse", "HEAD")
		writeProjectionPolicy(t, alice, []string{bob.Actor}, []string{memoryKindObservation})
		mustGit(t, "add", ".hn/policy.json")
		mustGit(t, "commit", "-q", "-m", "bob memory policy")
		bobPolicyCommit := mustGitText(t, "rev-parse", "HEAD")

		aliceMemory := projectionRecordAt(t, alice, memoryKindDecision, "alice decision", fullMemoryID("1"), "2026-08-30T01:00:00Z", MemoryAnchor{Commit: alicePolicyCommit}, Applicability{Mode: memoryApplicabilityExact}, nil)
		bobMemory := projectionRecordAt(t, bob, memoryKindDecision, "bob decision", fullMemoryID("2"), "2026-08-30T02:00:00Z", MemoryAnchor{Commit: alicePolicyCommit}, Applicability{Mode: memoryApplicabilityExact}, nil)

		// HEAD is at Bob's policy while the selected exact policy is Alice's.
		aliceContext, err := LoadMemoryProjectionPolicy(alicePolicyCommit, MemoryProjectionContext{AtCommit: alicePolicyCommit})
		if err != nil {
			t.Fatal(err)
		}
		aliceBytes, err := gitOutput("show", alicePolicyCommit+":.hn/policy.json")
		if err != nil {
			t.Fatal(err)
		}
		aliceProjection := ProjectMemories([]StoredMemory{bobMemory, aliceMemory}, aliceContext)
		if aliceProjection.PolicyCommit != alicePolicyCommit || aliceProjection.PolicyDigest != eventID(aliceBytes) ||
			findProjectionRow(t, aliceProjection.Rows, aliceMemory.ID).Trust != memoryTrustQualified ||
			findProjectionRow(t, aliceProjection.Rows, bobMemory.ID).Trust != memoryTrustActorUntrusted {
			t.Fatalf("Alice exact-policy projection = %#v", aliceProjection)
		}

		// Move HEAD away from Bob's policy and still select Bob's exact bytes.
		mustGit(t, "checkout", "-q", "--detach", alicePolicyCommit)
		bobContext, err := LoadMemoryProjectionPolicy(bobPolicyCommit, MemoryProjectionContext{AtCommit: alicePolicyCommit})
		if err != nil {
			t.Fatal(err)
		}
		bobBytes, err := gitOutput("show", bobPolicyCommit+":.hn/policy.json")
		if err != nil {
			t.Fatal(err)
		}
		bobProjection := ProjectMemories([]StoredMemory{aliceMemory, bobMemory}, bobContext)
		if bobProjection.PolicyCommit != bobPolicyCommit || bobProjection.PolicyDigest != eventID(bobBytes) ||
			findProjectionRow(t, bobProjection.Rows, aliceMemory.ID).Trust != memoryTrustActorUntrusted ||
			findProjectionRow(t, bobProjection.Rows, bobMemory.ID).Trust != memoryTrustKindUntrusted {
			t.Fatalf("Bob exact-policy projection = %#v", bobProjection)
		}
		if findProjectionRow(t, aliceProjection.Rows, aliceMemory.ID).Lifecycle != findProjectionRow(t, bobProjection.Rows, aliceMemory.ID).Lifecycle || aliceProjection.PolicyDigest == bobProjection.PolicyDigest {
			t.Fatal("policy selection changed lifecycle or failed to expose distinct exact digests")
		}
	})
}

func TestMemoryProjectionClassificationSeparationAndProvenance(t *testing.T) {
	alice := deterministicMemoryIdentity()
	bob := testIdentity(t, "Bob")
	anchor, child := strings.Repeat("a", 40), strings.Repeat("b", 40)
	resolver := &projectionResolverStub{probes: map[string]gitObjectProbe{
		anchor: {Exists: true, Type: "commit"}, child: {Exists: true, Type: "commit"},
	}}
	qualified := &MemoryPolicy{TrustedActors: []string{alice.Actor}, TrustedKinds: []string{memoryKindDecision}}
	baseMemory := projectionRecordAt(t, alice, memoryKindDecision, "base", fullMemoryID("1"), "2026-08-30T01:00:00Z", MemoryAnchor{Commit: anchor}, Applicability{Mode: memoryApplicabilityExact}, nil)
	baseContext := MemoryProjectionContext{AtCommit: anchor, MemoryPolicy: qualified, Resolver: resolver}
	base := findProjectionRow(t, ProjectMemories([]StoredMemory{baseMemory}, baseContext).Rows, baseMemory.ID)
	if base.Signature != "valid" || base.Lifecycle != memoryLifecycleActive || base.Applicability != memoryApplicabilityApplicable || base.Evidence != memoryEvidenceResolved || base.Trust != memoryTrustQualified {
		t.Fatalf("base classifications = %#v", base)
	}

	successor := projectionSupersessionAt(t, alice, baseMemory.ID, "replacement", fullMemoryID("2"), "2026-08-29T00:00:00Z", anchor)
	lifecycleChanged := findProjectionRow(t, ProjectMemories([]StoredMemory{successor, baseMemory}, baseContext).Rows, baseMemory.ID)
	assertProjectionDimensions(t, base, lifecycleChanged, "lifecycle")
	if lifecycleChanged.Lifecycle != memoryLifecycleSuperseded {
		t.Fatalf("lifecycle did not change independently: %#v", lifecycleChanged)
	}

	applicabilityChanged := findProjectionRow(t, ProjectMemories([]StoredMemory{baseMemory}, MemoryProjectionContext{AtCommit: child, MemoryPolicy: qualified, Resolver: resolver}).Rows, baseMemory.ID)
	assertProjectionDimensions(t, base, applicabilityChanged, "applicability")
	if applicabilityChanged.Applicability != memoryApplicabilityInapplicable {
		t.Fatalf("applicability did not change independently: %#v", applicabilityChanged)
	}

	evidenceMemory := projectionRecordAt(t, alice, memoryKindDecision, "base", fullMemoryID("3"), "2026-08-30T01:00:00Z", MemoryAnchor{Commit: anchor}, Applicability{Mode: memoryApplicabilityExact}, []string{"memory:" + fullMemoryID("f")})
	evidenceChanged := findProjectionRow(t, ProjectMemories([]StoredMemory{evidenceMemory}, baseContext).Rows, evidenceMemory.ID)
	assertProjectionDimensions(t, base, evidenceChanged, "evidence")
	if evidenceChanged.Evidence != memoryEvidenceMissing {
		t.Fatalf("evidence did not change independently: %#v", evidenceChanged)
	}

	actorTrust := findProjectionRow(t, ProjectMemories([]StoredMemory{baseMemory}, MemoryProjectionContext{AtCommit: anchor, MemoryPolicy: &MemoryPolicy{TrustedActors: []string{bob.Actor}, TrustedKinds: []string{memoryKindDecision}}, Resolver: resolver}).Rows, baseMemory.ID)
	assertProjectionDimensions(t, base, actorTrust, "trust")
	if actorTrust.Trust != memoryTrustActorUntrusted {
		t.Fatalf("actor trust did not change independently: %#v", actorTrust)
	}
	kindTrust := findProjectionRow(t, ProjectMemories([]StoredMemory{baseMemory}, MemoryProjectionContext{AtCommit: anchor, MemoryPolicy: &MemoryPolicy{TrustedActors: []string{alice.Actor}, TrustedKinds: []string{memoryKindObservation}}, Resolver: resolver}).Rows, baseMemory.ID)
	assertProjectionDimensions(t, base, kindTrust, "trust")
	if kindTrust.Trust != memoryTrustKindUntrusted {
		t.Fatalf("kind trust did not change independently: %#v", kindTrust)
	}
	policyMissing := findProjectionRow(t, ProjectMemories([]StoredMemory{baseMemory}, MemoryProjectionContext{AtCommit: anchor, Resolver: resolver}).Rows, baseMemory.ID)
	assertProjectionDimensions(t, base, policyMissing, "trust")
	if policyMissing.Trust != memoryTrustPolicyMissing {
		t.Fatalf("policy presence did not change independently: %#v", policyMissing)
	}

	retraction := projectionMemoryAt(t, alice, memoryOperationRetract, baseMemory.ID, nil, fullMemoryID("4"), "2026-08-28T00:00:00Z", nil)
	retracted := findProjectionRow(t, ProjectMemories([]StoredMemory{baseMemory, retraction}, baseContext).Rows, baseMemory.ID)
	if retracted.Lifecycle != memoryLifecycleRetracted || retracted.Applicability != base.Applicability || retracted.Evidence != base.Evidence || retracted.Trust != base.Trust {
		t.Fatalf("retraction crossed classification boundaries: %#v", retracted)
	}

	for _, value := range []string{base.ID, base.Stream, base.ContentDigest} {
		if !validMemoryID(value) {
			t.Errorf("provenance ID is not full and canonical: %q", value)
		}
	}
	if !validActorFingerprint(base.Actor) || base.Anchor.Commit != anchor || base.Data.Content != "base" {
		t.Fatalf("projection lost full provenance: %#v", base)
	}
}

func assertProjectionDimensions(t *testing.T, base, changed MemoryProjectionRow, dimension string) {
	t.Helper()
	if base.Signature != changed.Signature {
		t.Fatalf("%s change altered signature: %q -> %q", dimension, base.Signature, changed.Signature)
	}
	if dimension != "lifecycle" && base.Lifecycle != changed.Lifecycle {
		t.Fatalf("%s change altered lifecycle: %q -> %q", dimension, base.Lifecycle, changed.Lifecycle)
	}
	if dimension != "applicability" && base.Applicability != changed.Applicability {
		t.Fatalf("%s change altered applicability: %q -> %q", dimension, base.Applicability, changed.Applicability)
	}
	if dimension != "evidence" && base.Evidence != changed.Evidence {
		t.Fatalf("%s change altered evidence: %q -> %q", dimension, base.Evidence, changed.Evidence)
	}
	if dimension != "trust" && base.Trust != changed.Trust {
		t.Fatalf("%s change altered trust: %q -> %q", dimension, base.Trust, changed.Trust)
	}
}

func TestMemoryProjectionCollaborationBaselineUnchanged(t *testing.T) {
	alice := deterministicMemoryIdentity()
	event := newEvent(alice, "issue.open", 1, "")
	event.Timestamp = "2026-08-30T12:34:56Z"
	event.Title = "Memory must not rewrite collaboration"
	beforePayload, beforeSignature, err := encodeAndSign(event, alice)
	if err != nil {
		t.Fatal(err)
	}
	beforeID := eventID(beforePayload)
	projection := ProjectMemories(nil, MemoryProjectionContext{})
	if len(projection.Rows) != 0 || len(projection.Relationships) != 0 || len(projection.MissingDependencies) != 0 || len(projection.Diagnostics) != 0 {
		t.Fatalf("empty memory input changed collaboration-only state: %#v", projection)
	}
	afterPayload, afterSignature, err := encodeAndSign(event, alice)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforePayload, afterPayload) || !bytes.Equal(beforeSignature, afterSignature) || eventID(afterPayload) != beforeID || !validEventID(beforeID) {
		t.Fatalf("memory projection changed public collaboration bytes or ID: before=%s after=%s", beforeID, eventID(afterPayload))
	}
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
