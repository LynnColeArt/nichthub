package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

func signedIdentityEvent(t *testing.T, identity *Identity, event Event) StoredEvent {
	t.Helper()
	payload, signature, err := encodeAndSign(event, identity)
	if err != nil {
		t.Fatal(err)
	}
	return StoredEvent{ID: eventID(payload), Event: event, Payload: payload, Signature: signature}
}

func identityAuthorization(t *testing.T, from, target *Identity, relationship string, sequence uint64, previous string) StoredEvent {
	t.Helper()
	event := newEvent(from, "identity.authorize", sequence, previous)
	event.Relationship = relationship
	event.TargetActor = target.Actor
	event.TargetKey = target.PublicKey
	return signedIdentityEvent(t, from, event)
}

func identityAcceptance(t *testing.T, target *Identity, authorization string, sequence uint64, previous string) StoredEvent {
	t.Helper()
	event := newEvent(target, "identity.accept", sequence, previous)
	event.Subject = authorization
	return signedIdentityEvent(t, target, event)
}

func TestIdentityContinuityEventValidation(t *testing.T) {
	alice := testIdentity(t, "Alice")
	bob := testIdentity(t, "Bob")
	authorization := newEvent(alice, "identity.authorize", 1, "")
	authorization.Relationship = identityRelationshipDevice
	authorization.TargetActor = bob.Actor
	authorization.TargetKey = bob.PublicKey
	if err := validateEventContent(authorization); err != nil {
		t.Fatalf("valid authorization: %v", err)
	}

	for name, mutate := range map[string]func(*Event){
		"relationship": func(event *Event) { event.Relationship = "owner" },
		"short actor":  func(event *Event) { event.TargetActor = event.TargetActor[:12] },
		"bad key":      func(event *Event) { event.TargetKey = "not-a-key" },
		"key mismatch": func(event *Event) { event.TargetKey = alice.PublicKey },
		"self target": func(event *Event) {
			event.TargetActor = alice.Actor
			event.TargetKey = alice.PublicKey
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := authorization
			mutate(&candidate)
			if err := validateEventContent(candidate); err == nil {
				t.Fatal("invalid authorization unexpectedly passed validation")
			}
		})
	}

	acceptance := newEvent(bob, "identity.accept", 1, "")
	acceptance.Subject = strings.Repeat("a", 64)
	if err := validateEventContent(acceptance); err == nil {
		t.Fatal("acceptance with shortened/noncanonical subject unexpectedly passed")
	}
	acceptance.Subject = "sha256:" + strings.Repeat("a", 64)
	if err := validateEventContent(acceptance); err != nil {
		t.Fatalf("valid acceptance: %v", err)
	}
}

func TestIdentityFieldsPreserveExistingPayloadBytesAndID(t *testing.T) {
	event := Event{
		Protocol:  protocolVersion,
		Kind:      "issue.open",
		Actor:     "actor",
		ActorName: "Alice",
		PublicKey: "key",
		Sequence:  1,
		Timestamp: "2026-08-30T00:00:00Z",
		Title:     "Stable",
	}
	got, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte(`{"protocol":"hn/0","kind":"issue.open","actor":"actor","actorName":"Alice","publicKey":"key","sequence":1,"timestamp":"2026-08-30T00:00:00Z","title":"Stable"}`)
	if !bytes.Equal(got, want) {
		t.Fatalf("legacy payload changed:\n got %s\nwant %s", got, want)
	}
	if gotID := eventID(got); gotID != "sha256:017faf5d4cd870f251691ae66833ec1b85a90de9bf002e7777084e635702412d" {
		t.Fatalf("legacy event ID = %s", gotID)
	}
}

func TestIdentityFieldsPreserveEveryExistingEventKindID(t *testing.T) {
	fullA := "sha256:" + strings.Repeat("a", 64)
	fullB := "sha256:" + strings.Repeat("b", 64)
	gitA := strings.Repeat("a", 40)
	gitB := strings.Repeat("b", 40)
	base := Event{
		Protocol: protocolVersion, Actor: "actor", ActorName: "Alice", PublicKey: "key",
		Sequence: 1, Timestamp: "2026-08-30T00:00:00Z",
	}
	cases := map[string]Event{}
	issue := base
	issue.Kind, issue.Title, issue.Body = "issue.open", "Issue", "Body"
	cases[issue.Kind] = issue
	comment := base
	comment.Kind, comment.Subject, comment.Body = "issue.comment", fullA, "Comment"
	cases[comment.Kind] = comment
	proposal := base
	proposal.Kind, proposal.Title, proposal.Body, proposal.Base, proposal.Head = "proposal.open", "Proposal", "Body", gitA, gitB
	cases[proposal.Kind] = proposal
	revision := base
	revision.Kind, revision.Subject, revision.Body, revision.Base, revision.Head = "proposal.revise", fullA, "Body", gitA, gitB
	cases[revision.Kind] = revision
	review := base
	review.Kind, review.Subject, review.Body, review.Verdict = "review.submit", fullA, "Body", "approve"
	cases[review.Kind] = review
	request := base
	request.Kind, request.Subject, request.Pipeline, request.Definition, request.Commit = "run.request", fullA, "test", fullB, gitA
	cases[request.Kind] = request
	result := request
	result.Kind, result.Outcome, result.ExitCode, result.DurationMS = "run.result", "passed", 2, 3
	result.Log, result.Backend, result.Platform, result.Runner = fullA, "sandbox", "test/test", "hn/test"
	cases[result.Kind] = result
	decision := base
	decision.Kind, decision.Subject, decision.Body, decision.Verdict, decision.Policy = "proposal.decision", fullA, "Body", "accept", fullB
	decision.Evidence = []string{fullA}
	cases[decision.Kind] = decision
	merged := base
	merged.Kind, merged.Subject, merged.Head, merged.Commit, merged.Policy = "proposal.merged", fullA, gitA, gitB, fullB
	merged.Evidence = []string{fullA}
	cases[merged.Kind] = merged

	wantIDs := map[string]string{
		"issue.open":        "sha256:d976a3f3c6930d342ee42a2e254a0ab20ff47ae2ac3baa0a0d40fd9ed203554a",
		"issue.comment":     "sha256:400db40f30cabebdc8a09a81c5da4657fc5a2d7e1c613c27361e4a13b6b51715",
		"proposal.open":     "sha256:7fa1c9fa208653fad960fee5f325753ada2106e9d16d3a09b9723600aa572cd6",
		"proposal.revise":   "sha256:24edfc3ed0393e76608721e81a88d5dd8c1f54a38af04ef71a6673259de32611",
		"review.submit":     "sha256:87d65394a62ee0d8ee1de8b6d010014c0fc48fce6cfc41cabd517ea91778e71e",
		"run.request":       "sha256:8e1f77e01bf41530d363fa5f7197857e32a5f78042f306c161e644722e5a8159",
		"run.result":        "sha256:b9d287d1950b0e9ecfc771cdba97d2686ac2f343d64bec0ee48d270565da4256",
		"proposal.decision": "sha256:e61f5034ccb7ad5eab3b7bd0f3af02e14f82303438ffa66f57c41f3cbff3b76e",
		"proposal.merged":   "sha256:a7584d7f568283a468cf867dad71ed0096fac6fccfbaa6638af8580422f19283",
	}
	for kind, event := range cases {
		payload, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(payload, []byte(`"relationship"`)) || bytes.Contains(payload, []byte(`"targetActor"`)) || bytes.Contains(payload, []byte(`"targetKey"`)) {
			t.Fatalf("legacy %s payload contains identity-continuity fields: %s", kind, payload)
		}
		if got := eventID(payload); got != wantIDs[kind] {
			t.Errorf("%s ID = %q", kind, got)
		}
	}
}

func TestIdentityContinuityProjectionIsDeterministicAndExplicit(t *testing.T) {
	alice := testIdentity(t, "Alice")
	bob := testIdentity(t, "Bob")
	carol := testIdentity(t, "Carol")
	dave := testIdentity(t, "Dave")

	device := identityAuthorization(t, alice, bob, identityRelationshipDevice, 1, "")
	deviceAccept := identityAcceptance(t, bob, device.ID, 1, "")
	successor := identityAuthorization(t, alice, carol, identityRelationshipSuccessor, 2, device.ID)
	successorAccept := identityAcceptance(t, carol, successor.ID, 1, "")
	competing := identityAuthorization(t, alice, dave, identityRelationshipSuccessor, 3, successor.ID)
	competingAccept := identityAcceptance(t, dave, competing.ID, 1, "")
	missing := identityAcceptance(t, dave, "sha256:"+strings.Repeat("f", 64), 2, competingAccept.ID)

	events := []StoredEvent{device, deviceAccept, successor, successorAccept, competing, competingAccept, missing}
	want, err := ProjectIdentityContinuity(events)
	if err != nil {
		t.Fatal(err)
	}
	if len(want.Edges) != 3 || want.Edges[0].AuthorizationID > want.Edges[1].AuthorizationID {
		t.Fatalf("edges are not complete and sorted: %#v", want.Edges)
	}
	if len(want.Missing) != 1 || want.Missing[0].SubjectID != missing.Event.Subject {
		t.Fatalf("missing dependencies = %#v", want.Missing)
	}
	if got := want.Actor(alice.Actor); got == nil || got.State != IdentityActorAmbiguous {
		t.Fatalf("Alice state = %#v, want ambiguous", got)
	}
	for seed := int64(0); seed < 32; seed++ {
		shuffled := append([]StoredEvent(nil), events...)
		rand.New(rand.NewSource(seed)).Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		got, err := ProjectIdentityContinuity(shuffled)
		if err != nil {
			t.Fatal(err)
		}
		gotJSON, _ := json.Marshal(got)
		wantJSON, _ := json.Marshal(want)
		if !bytes.Equal(gotJSON, wantJSON) {
			t.Fatalf("projection changed for seed %d:\n got %s\nwant %s", seed, gotJSON, wantJSON)
		}
	}
}

func TestIdentityContinuityProjectionCyclesAndInvalidAcceptances(t *testing.T) {
	alice := testIdentity(t, "Alice")
	bob := testIdentity(t, "Bob")
	carol := testIdentity(t, "Carol")
	ab := identityAuthorization(t, alice, bob, identityRelationshipSuccessor, 1, "")
	ba := identityAuthorization(t, bob, alice, identityRelationshipSuccessor, 1, "")
	abAccept := identityAcceptance(t, bob, ab.ID, 2, ba.ID)
	baAccept := identityAcceptance(t, alice, ba.ID, 2, ab.ID)
	wrong := identityAcceptance(t, carol, ab.ID, 1, "")
	replayTarget := identityAuthorization(t, alice, carol, identityRelationshipSuccessor, 3, baAccept.ID)
	replayAccept := identityAcceptance(t, carol, replayTarget.ID, 2, wrong.ID)

	projection, err := ProjectIdentityContinuity([]StoredEvent{ab, ba, abAccept, baAccept, wrong, replayTarget, replayAccept})
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Conflicts) < 2 {
		t.Fatalf("conflicts = %#v, want cycle and signer mismatch", projection.Conflicts)
	}
	if got := projection.Actor(alice.Actor); got == nil || got.State != IdentityActorAmbiguous {
		t.Fatalf("Alice state = %#v, want ambiguous", got)
	}
	if got := projection.Actor(bob.Actor); got == nil || got.State != IdentityActorAmbiguous {
		t.Fatalf("Bob state = %#v, want ambiguous", got)
	}
}

func TestIdentityContinuityDuplicateAcceptanceIsOneLogicalEdge(t *testing.T) {
	alice := testIdentity(t, "Alice")
	bob := testIdentity(t, "Bob")
	authorization := identityAuthorization(t, alice, bob, identityRelationshipDevice, 1, "")
	first := identityAcceptance(t, bob, authorization.ID, 1, "")
	second := identityAcceptance(t, bob, authorization.ID, 2, first.ID)
	replayedAgainst := identityAuthorization(t, alice, bob, identityRelationshipDevice, 2, authorization.ID)

	projection, err := ProjectIdentityContinuity([]StoredEvent{second, replayedAgainst, authorization, first})
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Edges) != 2 {
		t.Fatalf("duplicate acceptance projection = %#v", projection.Edges)
	}
	states := map[string]IdentityContinuityEdge{}
	for _, edge := range projection.Edges {
		states[edge.AuthorizationID] = edge
	}
	if states[authorization.ID].State != IdentityEdgeAccepted || len(states[authorization.ID].AcceptanceIDs) != 2 || states[replayedAgainst.ID].State != IdentityEdgePending {
		t.Fatalf("acceptance replayed across authorization: %#v", projection.Edges)
	}
	if len(projection.Actor(alice.Actor).AcceptedDeviceIDs) != 1 {
		t.Fatalf("duplicate acceptance multiplied logical edge: %#v", projection.Actor(alice.Actor))
	}
}

func TestIdentityContinuityLongerSuccessorChainAndPendingEdge(t *testing.T) {
	alice := testIdentity(t, "Alice")
	bob := testIdentity(t, "Bob")
	carol := testIdentity(t, "Carol")
	dave := testIdentity(t, "Dave")
	ab := identityAuthorization(t, alice, bob, identityRelationshipSuccessor, 1, "")
	abAccept := identityAcceptance(t, bob, ab.ID, 1, "")
	bc := identityAuthorization(t, bob, carol, identityRelationshipSuccessor, 2, abAccept.ID)
	bcAccept := identityAcceptance(t, carol, bc.ID, 1, "")
	pending := identityAuthorization(t, carol, dave, identityRelationshipDevice, 2, bcAccept.ID)

	projection, err := ProjectIdentityContinuity([]StoredEvent{pending, bcAccept, ab, bc, abAccept})
	if err != nil {
		t.Fatal(err)
	}
	if projection.Actor(alice.Actor).State != IdentityActorRetired || projection.Actor(bob.Actor).State != IdentityActorRetired || projection.Actor(carol.Actor).State != IdentityActorActive {
		t.Fatalf("long successor chain states = %#v", projection.Actors)
	}
	for _, edge := range projection.Edges {
		if edge.AuthorizationID == pending.ID && edge.State != IdentityEdgePending {
			t.Fatalf("pending edge state = %s", edge.State)
		}
	}
}

func TestIdentityContinuityProjectionVerifiesCallerInput(t *testing.T) {
	alice := testIdentity(t, "Alice")
	bob := testIdentity(t, "Bob")
	authorization := identityAuthorization(t, alice, bob, identityRelationshipDevice, 1, "")
	authorization.Payload = append([]byte(nil), authorization.Payload...)
	authorization.Payload[len(authorization.Payload)-1] ^= 1
	if _, err := ProjectIdentityContinuity([]StoredEvent{authorization}); err == nil {
		t.Fatal("projector accepted payload that no longer matches its signature")
	}
}

func TestIdentityCommandsAndPlannedRotation(t *testing.T) {
	inIdentityTestRepository(t)
	alice, _, err := createIdentity("Alice")
	if err != nil {
		t.Fatal(err)
	}

	publicOutput, err := captureTestOutput(t, func() error { return cmdIdentityPublic(nil) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(publicOutput, alice.Actor) || !strings.Contains(publicOutput, alice.PublicKey) || strings.Contains(publicOutput, alice.PrivateKey) {
		t.Fatalf("unsafe/incomplete public output: %q", publicOutput)
	}

	if err := cmdIdentityRotate([]string{"--name", "Rotated"}); err != nil {
		t.Fatal(err)
	}
	active, err := loadIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if active.Actor == alice.Actor {
		t.Fatal("rotation did not switch the active actor")
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ProjectIdentityContinuity(events)
	if err != nil {
		t.Fatal(err)
	}
	if got := projection.Actor(alice.Actor); got == nil || got.State != IdentityActorRetired {
		t.Fatalf("predecessor state = %#v, want retired", got)
	}
	if len(events) != 2 {
		t.Fatalf("rotation event chains = %#v", events)
	}
	for _, stored := range events {
		if stored.Event.Actor == active.Actor && stored.Event.Sequence != 1 {
			t.Fatalf("successor started at sequence %d, want 1", stored.Event.Sequence)
		}
	}
	listOutput, err := captureTestOutput(t, func() error { return cmdIdentityList(nil) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listOutput, alice.Actor+"  retired") || !strings.Contains(listOutput, active.Actor+"  active") {
		t.Fatalf("identity list did not expose lifecycle states: %q", listOutput)
	}
	if _, err := os.Stat(mustIdentityPaths(t).rotation); !os.IsNotExist(err) {
		t.Fatalf("completed rotation state remains: %v", err)
	}
}

func TestIdentityRotationRetriesWithoutSiblingFacts(t *testing.T) {
	for _, interruptedStep := range []string{
		"rotation-authorization-durable",
		"rotation-acceptance-durable",
		"before-active-switch",
	} {
		t.Run(interruptedStep, func(t *testing.T) {
			inIdentityTestRepository(t)
			predecessor, _, err := createIdentity("Alice")
			if err != nil {
				t.Fatal(err)
			}
			fired := false
			identityStorageHook = func(step string) error {
				if step == interruptedStep && !fired {
					fired = true
					return errors.New("simulated interruption")
				}
				return nil
			}
			t.Cleanup(func() { identityStorageHook = nil })

			if err := cmdIdentityRotate([]string{"--name", "Successor"}); err == nil {
				t.Fatal("interrupted rotation unexpectedly completed")
			}
			active, err := loadIdentity()
			if err != nil {
				t.Fatal(err)
			}
			if active.Actor != predecessor.Actor {
				t.Fatalf("active actor changed early to %s", active.Actor)
			}
			stateBefore, err := loadIdentityRotation()
			if err != nil {
				t.Fatal(err)
			}
			targetBefore := stateBefore.TargetActor

			if err := cmdIdentityRotate([]string{"--name", "Ignored on retry"}); err != nil {
				t.Fatal(err)
			}
			active, err = loadIdentity()
			if err != nil {
				t.Fatal(err)
			}
			if active.Actor != targetBefore {
				t.Fatalf("retry activated %s, want recorded target %s", active.Actor, targetBefore)
			}
			events, err := collectEvents()
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != 2 {
				t.Fatalf("retry left %d events, want one authorization and one acceptance", len(events))
			}
		})
	}
}

func TestIdentityRotationReusesTargetAfterRecordDurabilityFailure(t *testing.T) {
	inIdentityTestRepository(t)
	predecessor, _, err := createIdentity("Alice")
	if err != nil {
		t.Fatal(err)
	}
	fired := false
	identityStorageHook = func(step string) error {
		if step == "rotation-target-record-durable" && !fired {
			fired = true
			return errors.New("simulated state-write boundary failure")
		}
		return nil
	}

	if err := cmdIdentityRotate([]string{"--name", "Successor"}); err == nil {
		t.Fatal("rotation unexpectedly completed after target-record boundary failure")
	}
	active, err := loadIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if active.Actor != predecessor.Actor {
		t.Fatalf("active actor = %s, want predecessor %s", active.Actor, predecessor.Actor)
	}
	commandState, err := loadIdentityRotationCommand()
	if err != nil {
		t.Fatal(err)
	}
	commandPath, err := identityRotationCommandPath()
	if err != nil {
		t.Fatal(err)
	}
	requirePathMode(t, commandPath, 0o600)
	originalTarget := commandState.State.TargetActor
	if _, err := loadIdentityRotation(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rotation state exists before its durable boundary: %v", err)
	}
	records, err := listIdentityRecords()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("record count after interruption = %d, want predecessor and recorded target", len(records))
	}

	identityStorageHook = nil
	if err := cmdIdentityRotate([]string{"--name", "Must not create sibling"}); err != nil {
		t.Fatal(err)
	}
	active, err = loadIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if active.Actor != originalTarget {
		t.Fatalf("retry activated %s, want original target %s", active.Actor, originalTarget)
	}
	records, err = listIdentityRecords()
	if err != nil {
		t.Fatal(err)
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || len(events) != 2 {
		t.Fatalf("retry produced %d records and %d events, want 2 and 2", len(records), len(events))
	}
	if _, err := os.Stat(commandPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("completed command journal still exists: %v", err)
	}
	actorsWithHistory := map[string]bool{}
	for _, event := range events {
		actorsWithHistory[event.Event.Actor] = true
	}
	if len(actorsWithHistory) != 2 || !actorsWithHistory[predecessor.Actor] || !actorsWithHistory[originalTarget] {
		t.Fatalf("retry histories = %#v", actorsWithHistory)
	}
}

func TestIdentityRotationCommandRetriesCleanupBoundariesWithoutSecondRotation(t *testing.T) {
	for _, step := range []string{"before-rotation-cleanup", "after-rotation-cleanup"} {
		t.Run(step, func(t *testing.T) {
			inIdentityTestRepository(t)
			predecessor, _, err := createIdentity("Alice")
			if err != nil {
				t.Fatal(err)
			}
			failures := 0
			identityStorageHook = func(got string) error {
				if got == step {
					failures++
					return errors.New("simulated cleanup boundary failure")
				}
				return nil
			}

			if err := cmdIdentityRotate([]string{"--name", "Successor"}); err == nil {
				t.Fatal("cleanup-boundary rotation unexpectedly reported success")
			}
			active, records, ids := identityRotationSnapshot(t)
			if active == predecessor.Actor || len(records) != 2 || len(ids) != 2 {
				t.Fatalf("first failure active=%s records=%v ids=%v", active, records, ids)
			}
			if _, err := loadIdentityRotationCommand(); err != nil {
				t.Fatalf("cleanup failure lost command journal: %v", err)
			}
			_, rotationErr := loadIdentityRotation()
			if step == "before-rotation-cleanup" && rotationErr != nil {
				t.Fatalf("pre-cleanup failure lost rotation state: %v", rotationErr)
			}
			if step == "after-rotation-cleanup" && !errors.Is(rotationErr, os.ErrNotExist) {
				t.Fatalf("post-cleanup failure retained rotation state: %v", rotationErr)
			}

			if step == "before-rotation-cleanup" {
				if err := cmdIdentityRotate([]string{"--name", "Ignored retry"}); err == nil {
					t.Fatal("repeated pre-cleanup failure unexpectedly reported success")
				}
				retryActive, retryRecords, retryIDs := identityRotationSnapshot(t)
				if retryActive != active || !slices.Equal(retryRecords, records) || !slices.Equal(retryIDs, ids) {
					t.Fatalf("failed retry changed transaction: active=%s records=%v ids=%v", retryActive, retryRecords, retryIDs)
				}
			}

			identityStorageHook = nil
			if err := cmdIdentityRotate([]string{"--name", "Ignored retry"}); err != nil {
				t.Fatal(err)
			}
			finalActive, finalRecords, finalIDs := identityRotationSnapshot(t)
			if finalActive != active || !slices.Equal(finalRecords, records) || !slices.Equal(finalIDs, ids) {
				t.Fatalf("successful retry changed transaction: active=%s records=%v ids=%v", finalActive, finalRecords, finalIDs)
			}
			commandPath, err := identityRotationCommandPath()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(commandPath); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("completed command journal still exists: %v", err)
			}
			if _, err := loadIdentityRotation(); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("completed rotation state still exists: %v", err)
			}
			if failures == 0 {
				t.Fatal("cleanup failure seam was not exercised")
			}
		})
	}
}

func TestIdentityRotationRejectsHostilePersistedEventState(t *testing.T) {
	missingA := "sha256:" + strings.Repeat("a", 64)
	missingB := "sha256:" + strings.Repeat("b", 64)
	tests := []struct {
		name  string
		setup func(*testing.T, *Identity, *Identity) identityRotationState
	}{
		{
			name: "missing event IDs",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, missingA, missingB)
			},
		},
		{
			name: "missing persisted authorization",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, missingA, "")
			},
		},
		{
			name: "signed IDs of wrong kinds",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				issue := appendIdentityTestEvent(t, predecessor, "issue.open", func(event *Event) { event.Title = "not authorization" })
				comment := appendIdentityTestEvent(t, target, "issue.open", func(event *Event) { event.Title = "not acceptance" })
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, issue.ID, comment.ID)
			},
		},
		{
			name: "authorization signed by wrong actor",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				other := testIdentity(t, "Other predecessor")
				authorization := appendIdentityTestAuthorization(t, other, target, identityRelationshipSuccessor)
				acceptance := appendIdentityTestAcceptance(t, target, authorization.ID)
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, authorization.ID, acceptance.ID)
			},
		},
		{
			name: "authorization names wrong target actor and key",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				otherTarget := testIdentity(t, "Other target")
				authorization := appendIdentityTestAuthorization(t, predecessor, otherTarget, identityRelationshipSuccessor)
				acceptance := appendIdentityTestAcceptance(t, otherTarget, authorization.ID)
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, authorization.ID, acceptance.ID)
			},
		},
		{
			name: "device relationship cannot complete rotation",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				authorization := appendIdentityTestAuthorization(t, predecessor, target, identityRelationshipDevice)
				acceptance := appendIdentityTestAcceptance(t, target, authorization.ID)
				return completedRotationState(predecessor, target, identityRelationshipDevice, authorization.ID, acceptance.ID)
			},
		},
		{
			name: "acceptance has unrelated subject",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				authorization := appendIdentityTestAuthorization(t, predecessor, target, identityRelationshipSuccessor)
				unrelated := appendIdentityTestAuthorization(t, predecessor, target, identityRelationshipSuccessor)
				acceptance := appendIdentityTestAcceptance(t, target, unrelated.ID)
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, authorization.ID, acceptance.ID)
			},
		},
		{
			name: "acceptance signed by wrong actor",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				authorization := appendIdentityTestAuthorization(t, predecessor, target, identityRelationshipSuccessor)
				wrongSigner := testIdentity(t, "Wrong signer")
				acceptance := appendIdentityTestAcceptance(t, wrongSigner, authorization.ID)
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, authorization.ID, acceptance.ID)
			},
		},
		{
			name: "corrupt authorization payload",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				authorization, acceptance := appendIdentityTestRotationPair(t, predecessor, target)
				corruptID := appendCorruptIdentityEventCommit(t, predecessor, authorization, true)
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, corruptID, acceptance.ID)
			},
		},
		{
			name: "corrupt acceptance signature",
			setup: func(t *testing.T, predecessor, target *Identity) identityRotationState {
				authorization, acceptance := appendIdentityTestRotationPair(t, predecessor, target)
				corruptID := appendCorruptIdentityEventCommit(t, target, acceptance, false)
				return completedRotationState(predecessor, target, identityRelationshipSuccessor, authorization.ID, corruptID)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inIdentityTestRepository(t)
			predecessor, _, err := createIdentity("Predecessor")
			if err != nil {
				t.Fatal(err)
			}
			target := testIdentity(t, "Target")
			if _, err := storeIdentityRecord(target, identityLifecycleAvailable); err != nil {
				t.Fatal(err)
			}
			state := test.setup(t, predecessor, target)
			if err := storeIdentityRotation(state); err != nil {
				t.Fatal(err)
			}
			if err := updateIdentityRotationCommand(state, target); err != nil {
				t.Fatal(err)
			}

			if err := cmdIdentityRotate(nil); err == nil {
				t.Fatal("hostile persisted state unexpectedly completed rotation")
			}
			active, err := loadIdentity()
			if err != nil {
				t.Fatal(err)
			}
			if active.Actor != predecessor.Actor {
				t.Fatalf("hostile state switched active actor to %s", active.Actor)
			}
			persisted, err := loadIdentityRotation()
			if err != nil {
				t.Fatalf("hostile state deleted rotation recovery record: %v", err)
			}
			if persisted != state {
				t.Fatalf("hostile state modified rotation recovery record: %#v", persisted)
			}
			command, err := loadIdentityRotationCommand()
			if err != nil {
				t.Fatalf("hostile state deleted command journal: %v", err)
			}
			if command.State != state || command.Target != *target {
				t.Fatalf("hostile state modified command journal: %#v", command)
			}
		})
	}
}

func TestIdentityRotationResumesVerifiedCompletedPair(t *testing.T) {
	inIdentityTestRepository(t)
	predecessor, _, err := createIdentity("Predecessor")
	if err != nil {
		t.Fatal(err)
	}
	target := testIdentity(t, "Target")
	if _, err := storeIdentityRecord(target, identityLifecycleAvailable); err != nil {
		t.Fatal(err)
	}
	authorization, acceptance := appendIdentityTestRotationPair(t, predecessor, target)
	state := completedRotationState(predecessor, target, identityRelationshipSuccessor, authorization.ID, acceptance.ID)
	if err := storeIdentityRotation(state); err != nil {
		t.Fatal(err)
	}
	if err := updateIdentityRotationCommand(state, target); err != nil {
		t.Fatal(err)
	}

	if err := cmdIdentityRotate(nil); err != nil {
		t.Fatal(err)
	}
	active, records, ids := identityRotationSnapshot(t)
	wantIDs := []string{authorization.ID, acceptance.ID}
	sort.Strings(wantIDs)
	if active != target.Actor || len(records) != 2 || !slices.Equal(ids, wantIDs) {
		t.Fatalf("verified resume active=%s records=%v ids=%v", active, records, ids)
	}
	if _, err := loadIdentityRotation(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("verified resume retained rotation state: %v", err)
	}
	if _, err := loadIdentityRotationCommand(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("verified resume retained command journal: %v", err)
	}
}

func completedRotationState(predecessor, target *Identity, relationship, authorizationID, acceptanceID string) identityRotationState {
	return identityRotationState{
		Version: identityRotationVersion, PredecessorActor: predecessor.Actor, TargetActor: target.Actor,
		Relationship: relationship, AuthorizationEvent: authorizationID, AcceptanceEvent: acceptanceID,
	}
}

func appendIdentityTestEvent(t *testing.T, identity *Identity, kind string, fill func(*Event)) *StoredEvent {
	t.Helper()
	event, err := nextEvent(identity, kind)
	if err != nil {
		t.Fatal(err)
	}
	fill(&event)
	stored, err := appendEvent(event, identity)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func appendIdentityTestAuthorization(t *testing.T, predecessor, target *Identity, relationship string) *StoredEvent {
	t.Helper()
	return appendIdentityTestEvent(t, predecessor, "identity.authorize", func(event *Event) {
		event.Relationship = relationship
		event.TargetActor = target.Actor
		event.TargetKey = target.PublicKey
	})
}

func appendIdentityTestAcceptance(t *testing.T, signer *Identity, authorizationID string) *StoredEvent {
	t.Helper()
	return appendIdentityTestEvent(t, signer, "identity.accept", func(event *Event) {
		event.Subject = authorizationID
	})
}

func appendIdentityTestRotationPair(t *testing.T, predecessor, target *Identity) (*StoredEvent, *StoredEvent) {
	t.Helper()
	authorization := appendIdentityTestAuthorization(t, predecessor, target, identityRelationshipSuccessor)
	return authorization, appendIdentityTestAcceptance(t, target, authorization.ID)
}

func appendCorruptIdentityEventCommit(t *testing.T, identity *Identity, predecessor *StoredEvent, corruptPayload bool) string {
	t.Helper()
	payload := append([]byte(nil), predecessor.Payload...)
	signature := append([]byte(nil), predecessor.Signature...)
	if corruptPayload {
		payload = append(payload, ' ')
	} else {
		signature[0] ^= 1
	}
	eventBlob, err := gitInput(payload, nil, "hash-object", "-w", "--stdin")
	if err != nil {
		t.Fatal(err)
	}
	signatureText := []byte(base64.RawStdEncoding.EncodeToString(signature))
	signatureBlob, err := gitInput(signatureText, nil, "hash-object", "-w", "--stdin")
	if err != nil {
		t.Fatal(err)
	}
	treeInput := "100644 blob " + strings.TrimSpace(string(eventBlob)) + "\tevent.json\n" +
		"100644 blob " + strings.TrimSpace(string(signatureBlob)) + "\tsignature\n"
	tree, err := gitInput([]byte(treeInput), nil, "mktree")
	if err != nil {
		t.Fatal(err)
	}
	commit, err := gitInput(nil, nil, "commit-tree", strings.TrimSpace(string(tree)), "-p", predecessor.Commit, "-m", "corrupt identity event")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gitOutput("update-ref", actorRef(identity.Actor), strings.TrimSpace(string(commit)), predecessor.Commit); err != nil {
		t.Fatal(err)
	}
	return eventID(payload)
}

func identityRotationSnapshot(t *testing.T) (string, []string, []string) {
	t.Helper()
	active, err := loadIdentity()
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := listIdentityRecords()
	if err != nil {
		t.Fatal(err)
	}
	records := make([]string, 0, len(metadata))
	for _, record := range metadata {
		records = append(records, record.Actor)
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}
	sort.Strings(ids)
	return active.Actor, records, ids
}

func TestIdentityAuthorizeAndAcceptRequireExactTarget(t *testing.T) {
	inIdentityTestRepository(t)
	alice, _, err := createIdentity("Alice")
	if err != nil {
		t.Fatal(err)
	}
	bob := testIdentity(t, "Bob")
	if err := cmdIdentityAuthorize([]string{"--relationship", "device", "--actor", bob.Actor[:12], "--public-key", bob.PublicKey}); err == nil {
		t.Fatal("short target actor unexpectedly authorized")
	}
	if err := cmdIdentityAuthorize([]string{"--relationship", "device", "--actor", alice.Actor, "--public-key", alice.PublicKey}); err == nil {
		t.Fatal("self-target unexpectedly authorized")
	}
	storedEvent, err := nextEvent(alice, "identity.authorize")
	if err != nil {
		t.Fatal(err)
	}
	storedEvent.Relationship = identityRelationshipDevice
	storedEvent.TargetActor = bob.Actor
	storedEvent.TargetKey = bob.PublicKey
	authorization, err := appendEvent(storedEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := cmdIdentityAccept([]string{authorization.ID[:20]}); err == nil {
		t.Fatal("short authorization ID unexpectedly accepted")
	}
	if err := cmdIdentityAccept([]string{authorization.ID}); err == nil || !strings.Contains(err.Error(), "exact target") {
		t.Fatalf("wrong active actor acceptance error = %v", err)
	}
	if _, err := storeIdentityRecord(bob, identityLifecycleAvailable); err != nil {
		t.Fatal(err)
	}
	if err := replaceActiveIdentity(bob.Actor, alice.Actor); err != nil {
		t.Fatal(err)
	}
	if err := cmdIdentityAccept([]string{authorization.ID}); err != nil {
		t.Fatal(err)
	}
	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	projection, err := ProjectIdentityContinuity(events)
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Edges) != 1 || projection.Edges[0].State != IdentityEdgeAccepted {
		t.Fatalf("accepted command did not complete exact edge: %#v", projection.Edges)
	}
}

func TestIdentityContinuityDoesNotGrantPolicyAuthority(t *testing.T) {
	root := inIdentityTestRepository(t)
	alice := testIdentity(t, "Alice")
	bob := testIdentity(t, "Bob")
	policy := PolicyDocument{
		Version: policyVersion, Maintainers: []string{alice.Actor},
		Proposals: ProposalPolicy{
			RequiredApprovals: 1, RequiredAccepts: 1,
			TrustedReviewers: []string{alice.Actor}, AllowAuthorApproval: true,
		},
		Pipelines: map[string]PipelinePolicy{
			"test": {RequiredResults: 1, TrustedRunners: []string{alice.Actor}},
		},
	}
	writeTestPolicy(t, root, policy)
	if err := os.MkdirAll(filepath.Join(root, ".hn", "pipelines"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".hn", "pipelines", "test.json"), []byte("{\"version\":\"hn.pipeline/0\",\"steps\":[{\"name\":\"test\",\"command\":\"true\"}]}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, "add", ".hn")
	mustGit(t, "commit", "-q", "-m", "policy")
	base := mustGitText(t, "rev-parse", "HEAD")
	_, policyBytesBefore, policyDigestBefore, err := loadPolicy(base)
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, "commit", "--allow-empty", "-q", "-m", "candidate")
	head := mustGitText(t, "rev-parse", "HEAD")

	proposalEvent := newEvent(alice, "proposal.open", 1, "")
	proposalEvent.Title = "Authority stays explicit"
	proposalEvent.Base = base
	proposalEvent.Head = head
	proposal, err := appendEvent(proposalEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	if err := createProposalRef(proposal.ID, head); err != nil {
		t.Fatal(err)
	}
	authorizationEvent, err := nextEvent(alice, "identity.authorize")
	if err != nil {
		t.Fatal(err)
	}
	authorizationEvent.Relationship = identityRelationshipDevice
	authorizationEvent.TargetActor = bob.Actor
	authorizationEvent.TargetKey = bob.PublicKey
	authorization, err := appendEvent(authorizationEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	acceptanceEvent := newEvent(bob, "identity.accept", 1, "")
	acceptanceEvent.Subject = authorization.ID
	acceptance, err := appendEvent(acceptanceEvent, bob)
	if err != nil {
		t.Fatal(err)
	}
	reviewEvent := newEvent(bob, "review.submit", 2, acceptance.ID)
	reviewEvent.Subject = proposal.ID
	reviewEvent.Verdict = "approve"
	review, err := appendEvent(reviewEvent, bob)
	if err != nil {
		t.Fatal(err)
	}
	requestEvent, err := nextEvent(alice, "run.request")
	if err != nil {
		t.Fatal(err)
	}
	_, _, definition, err := loadPipeline(head, "test")
	if err != nil {
		t.Fatal(err)
	}
	requestEvent.Subject = proposal.ID
	requestEvent.Pipeline = "test"
	requestEvent.Definition = definition
	requestEvent.Commit = head
	request, err := appendEvent(requestEvent, alice)
	if err != nil {
		t.Fatal(err)
	}
	log := []byte("passed\n")
	resultEvent := newEvent(bob, "run.result", 3, review.ID)
	resultEvent.Subject = request.ID
	resultEvent.Pipeline = "test"
	resultEvent.Definition = definition
	resultEvent.Commit = head
	resultEvent.Outcome = "passed"
	resultEvent.DurationMS = 1
	resultEvent.Log = eventID(log)
	resultEvent.Backend = "sandbox"
	resultEvent.Platform = "test/test"
	resultEvent.Runner = "hn/test"
	if _, err := appendEventWithAttachments(resultEvent, bob, map[string][]byte{"log.txt": log}); err != nil {
		t.Fatal(err)
	}

	events, err := collectEvents()
	if err != nil {
		t.Fatal(err)
	}
	proposal, err = resolveFullEvent(events, proposal.ID)
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := evaluateProposal(proposal, events)
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Ready || len(evaluation.Approvals) != 0 || len(evaluation.Pipelines) != 1 || len(evaluation.Pipelines[0].Passed) != 0 {
		t.Fatalf("continuity target gained reviewer/runner authority: %#v", evaluation)
	}
	_, policyBytesAfter, policyDigestAfter, err := loadPolicy(base)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(policyBytesBefore, policyBytesAfter) || policyDigestBefore != policyDigestAfter {
		t.Fatal("identity continuity changed policy bytes or digest")
	}

	decisionEvent := newEvent(bob, "proposal.decision", 4, "")
	decisionEvent.Subject = proposal.ID
	decisionEvent.Policy = policyDigestBefore
	decisionEvent.Verdict = "accept"
	decisionEvent.Evidence = []string{review.ID}
	decision := signedIdentityEvent(t, bob, decisionEvent)
	byID := make(map[string]StoredEvent, len(events)+1)
	for _, event := range events {
		byID[event.ID] = event
	}
	byID[decision.ID] = decision
	if err := validateDecisionEvent(decision, *proposal, byID); err == nil || !strings.Contains(err.Error(), "not authorized") {
		t.Fatalf("continuity target gained decision authority: %v", err)
	}
	mergeEvent := newEvent(bob, "proposal.merged", 4, "")
	mergeEvent.Subject = proposal.ID
	mergeEvent.Policy = policyDigestBefore
	mergeEvent.Head = head
	mergeEvent.Commit = head
	mergeEvent.Evidence = []string{decision.ID}
	merge := signedIdentityEvent(t, bob, mergeEvent)
	if err := validateMergeEvent(merge, *proposal, byID); err == nil || !strings.Contains(err.Error(), "not authorized") {
		t.Fatalf("continuity target gained merge authority: %v", err)
	}
}

func mustIdentityPaths(t *testing.T) identityKeyringLayout {
	t.Helper()
	paths, err := identityKeyringPaths()
	if err != nil {
		t.Fatal(err)
	}
	return paths
}
