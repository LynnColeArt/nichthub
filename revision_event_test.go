package main

import (
	"os"
	"strings"
	"testing"
)

func revisionTestOID(digit string) string {
	return strings.Repeat(digit, 40)
}

func revisionTestStored(label string, event Event) StoredEvent {
	return StoredEvent{ID: eventID([]byte(label)), Event: event}
}

func TestProposalRevisionSignedRoundTrip(t *testing.T) {
	identity := testIdentity(t, "Alice")
	event := newEvent(identity, "proposal.revise", 1, "")
	event.Subject = eventID([]byte("predecessor"))
	event.Body = "Resolved against the new base."
	event.Base = revisionTestOID("a")
	event.Head = revisionTestOID("b")

	payload, signature, err := encodeAndSign(event, identity)
	if err != nil {
		t.Fatal(err)
	}
	got, id, err := verifyEvent(payload, signature)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != event.Kind || got.Subject != event.Subject || got.Base != event.Base || got.Head != event.Head || got.Body != event.Body {
		t.Fatalf("verified revision = %#v, want %#v", got, event)
	}
	if id != eventID(payload) {
		t.Fatalf("revision ID = %q, want %q", id, eventID(payload))
	}
}

func TestProposalRevisionContentValidation(t *testing.T) {
	valid := Event{
		Kind:    "proposal.revise",
		Subject: eventID([]byte("predecessor")),
		Base:    revisionTestOID("a"),
		Head:    revisionTestOID("b"),
	}
	tests := []struct {
		name    string
		mutate  func(*Event)
		wantErr bool
	}{
		{name: "valid"},
		{name: "missing predecessor", mutate: func(event *Event) { event.Subject = "" }, wantErr: true},
		{name: "invalid predecessor", mutate: func(event *Event) { event.Subject = "sha256:not-a-digest" }, wantErr: true},
		{name: "missing base", mutate: func(event *Event) { event.Base = "" }, wantErr: true},
		{name: "missing head", mutate: func(event *Event) { event.Head = "" }, wantErr: true},
		{name: "same base and head", mutate: func(event *Event) { event.Head = event.Base }, wantErr: true},
		{name: "independent title authority", mutate: func(event *Event) { event.Title = "Renamed proposal" }, wantErr: true},
		{name: "whitespace title authority", mutate: func(event *Event) { event.Title = "   " }, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := valid
			if test.mutate != nil {
				test.mutate(&event)
			}
			err := validateEventContent(event)
			if test.wantErr && err == nil {
				t.Fatal("validation unexpectedly succeeded")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("validation failed: %v", err)
			}
		})
	}
}

func TestProposalRevisionRelationships(t *testing.T) {
	root := revisionTestStored("root", Event{
		Kind:  "proposal.open",
		Actor: "alice",
		Base:  revisionTestOID("a"),
		Head:  revisionTestOID("b"),
	})
	revisionA := revisionTestStored("revision-a", Event{
		Kind:    "proposal.revise",
		Actor:   "alice",
		Subject: root.ID,
		Base:    revisionTestOID("c"),
		Head:    revisionTestOID("d"),
	})
	revisionB := revisionTestStored("revision-b", Event{
		Kind:    "proposal.revise",
		Actor:   "alice",
		Subject: root.ID,
		Base:    revisionTestOID("e"),
		Head:    revisionTestOID("f"),
	})
	revisionA2 := revisionTestStored("revision-a2", Event{
		Kind:    "proposal.revise",
		Actor:   "alice",
		Subject: revisionA.ID,
		Base:    revisionTestOID("1"),
		Head:    revisionTestOID("2"),
	})
	review := revisionTestStored("review", Event{
		Kind:    "review.submit",
		Actor:   "bob",
		Subject: revisionA.ID,
		Verdict: "approve",
	})
	runRequest := revisionTestStored("run-request", Event{
		Kind:    "run.request",
		Actor:   "alice",
		Subject: revisionA.ID,
		Commit:  revisionA.Event.Head,
	})

	orders := [][]StoredEvent{
		{root, revisionA, revisionB, revisionA2, review, runRequest},
		{runRequest, review, revisionA2, revisionB, revisionA, root},
	}
	for index, events := range orders {
		if err := validateEventRelationships(events); err != nil {
			t.Fatalf("order %d failed: %v", index, err)
		}
	}
	badRunRequest := runRequest
	badRunRequest.ID = eventID([]byte("bad-run-request"))
	badRunRequest.Event.Commit = revisionB.Event.Head
	if err := validateEventRelationships([]StoredEvent{root, revisionA, badRunRequest}); err == nil || !strings.Contains(err.Error(), shortID(badRunRequest.ID)) {
		t.Fatalf("mismatched revision run request error = %v, want request ID %s", err, shortID(badRunRequest.ID))
	}
}

func TestProposalRevisionRelationshipRejections(t *testing.T) {
	root := revisionTestStored("root", Event{Kind: "proposal.open", Actor: "alice"})
	issue := revisionTestStored("issue", Event{Kind: "issue.open", Actor: "alice"})
	validRevision := Event{Kind: "proposal.revise", Actor: "alice", Subject: root.ID}

	tests := []struct {
		name        string
		events      func() []StoredEvent
		wantErr     string
		wantIDIndex int
	}{
		{
			name: "missing predecessor",
			events: func() []StoredEvent {
				revision := revisionTestStored("missing", validRevision)
				revision.Event.Subject = eventID([]byte("absent"))
				return []StoredEvent{revision}
			},
			wantErr:     "does not reference an available proposal",
			wantIDIndex: 0,
		},
		{
			name: "wrong predecessor kind",
			events: func() []StoredEvent {
				revision := revisionTestStored("wrong-kind", validRevision)
				revision.Event.Subject = issue.ID
				return []StoredEvent{issue, revision}
			},
			wantErr:     "does not reference an available proposal",
			wantIDIndex: 1,
		},
		{
			name: "different author",
			events: func() []StoredEvent {
				revision := revisionTestStored("different-author", validRevision)
				revision.Event.Actor = "mallory"
				return []StoredEvent{root, revision}
			},
			wantErr:     "is not signed by predecessor author",
			wantIDIndex: 1,
		},
		{
			name: "self link",
			events: func() []StoredEvent {
				revision := revisionTestStored("self", validRevision)
				revision.Event.Subject = revision.ID
				return []StoredEvent{revision}
			},
			wantErr:     "revision lineage contains a cycle at",
			wantIDIndex: 0,
		},
		{
			name: "cycle",
			events: func() []StoredEvent {
				first := revisionTestStored("cycle-a", validRevision)
				second := revisionTestStored("cycle-b", validRevision)
				first.Event.Subject = second.ID
				second.Event.Subject = first.ID
				return []StoredEvent{first, second}
			},
			wantErr:     "revision lineage contains a cycle at",
			wantIDIndex: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := test.events()
			revisionID := events[test.wantIDIndex].ID
			err := validateEventRelationships(events)
			if err == nil {
				t.Fatal("relationship validation unexpectedly succeeded")
			}
			if !strings.Contains(err.Error(), test.wantErr) || !strings.Contains(err.Error(), shortID(revisionID)) {
				t.Fatalf("relationship error = %q, want reason %q and revision ID %q", err, test.wantErr, shortID(revisionID))
			}
		})
	}
}

func TestProposalRevisionGovernanceRelationshipsAndLaterMergeFact(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	alice := testIdentity(t, "Alice")
	mustGit(t, "init", "-q", "-b", "main")
	mustGit(t, "config", "user.name", "Test")
	mustGit(t, "config", "user.email", "test@hn.invalid")
	writeTestPolicy(t, root, PolicyDocument{
		Version:     policyVersion,
		Maintainers: []string{alice.Actor},
		Proposals: ProposalPolicy{
			RequiredApprovals:   0,
			RequiredAccepts:     1,
			AllowAuthorApproval: true,
		},
		Pipelines: map[string]PipelinePolicy{},
	})
	mustGit(t, "add", ".hn/policy.json")
	mustGit(t, "commit", "-q", "-m", "base policy")
	base := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "proposal head")
	head := mustGitText(t, "rev-parse", "HEAD")
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "revision head")
	revisionHead := mustGitText(t, "rev-parse", "HEAD")

	openEvent := newEvent(alice, "proposal.open", 1, "")
	openEvent.Title = "Merge and revision facts"
	openEvent.Base = base
	openEvent.Head = head
	proposal, err := appendEvent(openEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	revisionEvent, err := nextEvent(alice, "proposal.revise")
	if err != nil {
		t.Fatal(err)
	}
	revisionEvent.Subject = proposal.ID
	revisionEvent.Base = base
	revisionEvent.Head = revisionHead
	revision, err := appendEvent(revisionEvent, alice)
	if err != nil {
		t.Fatal(err)
	}

	_, _, policyDigest, err := loadPolicy(base)
	if err != nil {
		t.Fatal(err)
	}
	decisionEvent, err := nextEvent(alice, "proposal.decision")
	if err != nil {
		t.Fatal(err)
	}
	decisionEvent.Subject = proposal.ID
	decisionEvent.Policy = policyDigest
	decisionEvent.Verdict = "accept"
	decision, err := appendEvent(decisionEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	mergeEvent, err := nextEvent(alice, "proposal.merged")
	if err != nil {
		t.Fatal(err)
	}
	mergeEvent.Subject = proposal.ID
	mergeEvent.Policy = policyDigest
	mergeEvent.Head = head
	mergeEvent.Commit = head
	mergeEvent.Evidence = []string{decision.ID}
	if _, err := appendEvent(mergeEvent, alice); err != nil {
		t.Fatal(err)
	}

	events, err := collectEvents()
	if err != nil {
		t.Fatalf("revision plus later merge fact was rejected: %v", err)
	}
	storedRevision, err := resolveEvent(events, revision.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedRevision.Event.Subject != proposal.ID {
		t.Fatalf("revision predecessor = %s, want %s", storedRevision.Event.Subject, proposal.ID)
	}

	revisionDecisionEvent, err := nextEvent(alice, "proposal.decision")
	if err != nil {
		t.Fatal(err)
	}
	revisionDecisionEvent.Subject = revision.ID
	revisionDecisionEvent.Policy = policyDigest
	revisionDecisionEvent.Verdict = "accept"
	revisionDecision, err := appendEvent(revisionDecisionEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	revisionMergeEvent, err := nextEvent(alice, "proposal.merged")
	if err != nil {
		t.Fatal(err)
	}
	revisionMergeEvent.Subject = revision.ID
	revisionMergeEvent.Policy = policyDigest
	revisionMergeEvent.Head = revisionHead
	revisionMergeEvent.Commit = revisionHead
	revisionMergeEvent.Evidence = []string{revisionDecision.ID}
	revisionMerge, err := appendEvent(revisionMergeEvent, alice)
	if err != nil {
		t.Fatal(err)
	}

	revisionGovernance := []StoredEvent{*proposal, *revision, *revisionDecision, *revisionMerge}
	if err := validateEventRelationships(revisionGovernance); err != nil {
		t.Fatalf("revision decision and merge relationships failed: %v", err)
	}
	badDecision := *revisionDecision
	badDecision.Event.Policy = eventID([]byte("wrong-policy"))
	if err := validateEventRelationships([]StoredEvent{*proposal, *revision, badDecision}); err == nil || !strings.Contains(err.Error(), shortID(badDecision.ID)) {
		t.Fatalf("mismatched revision decision error = %v, want decision ID %s", err, shortID(badDecision.ID))
	}
	badMergeHead := *revisionMerge
	badMergeHead.Event.Head = head
	if err := validateEventRelationships([]StoredEvent{*proposal, *revision, *revisionDecision, badMergeHead}); err == nil || !strings.Contains(err.Error(), shortID(badMergeHead.ID)) {
		t.Fatalf("mismatched revision merge head error = %v, want merge ID %s", err, shortID(badMergeHead.ID))
	}
	badMergeEvidence := *revisionMerge
	badMergeEvidence.Event.Evidence = nil
	if err := validateEventRelationships([]StoredEvent{*proposal, *revision, *revisionDecision, badMergeEvidence}); err == nil || !strings.Contains(err.Error(), shortID(badMergeEvidence.ID)) {
		t.Fatalf("mismatched revision merge evidence error = %v, want merge ID %s", err, shortID(badMergeEvidence.ID))
	}
}
