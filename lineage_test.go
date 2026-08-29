package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type lineageTestFixture struct {
	events []StoredEvent
	root   StoredEvent
	a      StoredEvent
	b      StoredEvent
	a2     StoredEvent
}

func lineageTestEvent(label, kind, subject string) StoredEvent {
	return StoredEvent{
		ID: eventID([]byte(label)),
		Event: Event{
			Kind:      kind,
			Actor:     "alice",
			Subject:   subject,
			Timestamp: "2026-08-29T" + label,
		},
	}
}

func newLineageTestFixture() lineageTestFixture {
	root := lineageTestEvent("root", "proposal.open", "")
	a := lineageTestEvent("revision-a", "proposal.revise", root.ID)
	b := lineageTestEvent("revision-b", "proposal.revise", root.ID)
	a2 := lineageTestEvent("revision-a2", "proposal.revise", a.ID)
	return lineageTestFixture{
		events: []StoredEvent{root, a, b, a2},
		root:   root,
		a:      a,
		b:      b,
		a2:     a2,
	}
}

func sortedLineageTestIDs(ids ...string) []string {
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)
	return sorted
}

func requireLineageIDs(t *testing.T, label string, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func lineageTestPermutations(events []StoredEvent) [][]StoredEvent {
	original := append([]StoredEvent(nil), events...)
	reversed := append([]StoredEvent(nil), events...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	shuffled := append([]StoredEvent(nil), events...)
	rand.New(rand.NewSource(42)).Shuffle(len(shuffled), func(left, right int) {
		shuffled[left], shuffled[right] = shuffled[right], shuffled[left]
	})
	return [][]StoredEvent{original, reversed, shuffled}
}

func TestLineageTopologyConvergesAcrossInputOrders(t *testing.T) {
	fixture := newLineageTestFixture()
	wantMembers := sortedLineageTestIDs(fixture.root.ID, fixture.a.ID, fixture.b.ID, fixture.a2.ID)

	for order, events := range lineageTestPermutations(fixture.events) {
		index, err := buildLineageIndex(events)
		if err != nil {
			t.Fatalf("order %d build: %v", order, err)
		}

		predecessor, exists, err := index.predecessor(fixture.a.ID)
		if err != nil || !exists || predecessor != fixture.root.ID {
			t.Fatalf("order %d predecessor = %q, %t, %v; want %q, true, nil", order, predecessor, exists, err, fixture.root.ID)
		}
		if predecessor, exists, err := index.predecessor(fixture.root.ID); err != nil || exists || predecessor != "" {
			t.Fatalf("order %d root predecessor = %q, %t, %v; want empty, false, nil", order, predecessor, exists, err)
		}

		successors, err := index.successors(fixture.root.ID)
		if err != nil {
			t.Fatal(err)
		}
		requireLineageIDs(t, fmt.Sprintf("order %d root successors", order), successors, sortedLineageTestIDs(fixture.a.ID, fixture.b.ID))

		siblings, err := index.siblings(fixture.a.ID)
		if err != nil {
			t.Fatal(err)
		}
		requireLineageIDs(t, fmt.Sprintf("order %d revision-a siblings", order), siblings, []string{fixture.b.ID})

		root, err := index.root(fixture.a2.ID)
		if err != nil || root != fixture.root.ID {
			t.Fatalf("order %d root = %q, %v; want %q, nil", order, root, err, fixture.root.ID)
		}

		ancestors, err := index.ancestors(fixture.a2.ID)
		if err != nil {
			t.Fatal(err)
		}
		requireLineageIDs(t, fmt.Sprintf("order %d ancestors", order), ancestors, sortedLineageTestIDs(fixture.root.ID, fixture.a.ID))

		descendants, err := index.descendants(fixture.root.ID)
		if err != nil {
			t.Fatal(err)
		}
		requireLineageIDs(t, fmt.Sprintf("order %d descendants", order), descendants, sortedLineageTestIDs(fixture.a.ID, fixture.b.ID, fixture.a2.ID))

		members, err := index.members(fixture.b.ID)
		if err != nil {
			t.Fatal(err)
		}
		requireLineageIDs(t, fmt.Sprintf("order %d members", order), members, wantMembers)
	}
}

func TestLineageUnknownProposalFailsLoudly(t *testing.T) {
	index, err := buildLineageIndex(newLineageTestFixture().events)
	if err != nil {
		t.Fatal(err)
	}
	unknown := eventID([]byte("unknown-proposal"))
	queries := []struct {
		name string
		run  func() error
	}{
		{name: "predecessor", run: func() error { _, _, err := index.predecessor(unknown); return err }},
		{name: "successors", run: func() error { _, err := index.successors(unknown); return err }},
		{name: "siblings", run: func() error { _, err := index.siblings(unknown); return err }},
		{name: "root", run: func() error { _, err := index.root(unknown); return err }},
		{name: "ancestors", run: func() error { _, err := index.ancestors(unknown); return err }},
		{name: "descendants", run: func() error { _, err := index.descendants(unknown); return err }},
		{name: "members", run: func() error { _, err := index.members(unknown); return err }},
		{name: "state", run: func() error { _, err := index.state(unknown); return err }},
	}
	for _, query := range queries {
		t.Run(query.name, func(t *testing.T) {
			err := query.run()
			if err == nil || !strings.Contains(err.Error(), unknown) {
				t.Fatalf("error = %v, want exact unknown proposal ID %s", err, unknown)
			}
		})
	}
}

func TestLineageStatesPreserveMergeFacts(t *testing.T) {
	fixture := newLineageTestFixture()
	decision := lineageTestEvent("accepted-a", "proposal.decision", fixture.a.ID)
	mergeA := lineageTestEvent("merge-a", "proposal.merged", fixture.a.ID)
	mergeA2 := lineageTestEvent("merge-a-second-fact", "proposal.merged", fixture.a.ID)
	mergeB := lineageTestEvent("merge-b", "proposal.merged", fixture.b.ID)

	withoutMerges, err := buildLineageIndex(append(fixture.events, decision))
	if err != nil {
		t.Fatal(err)
	}
	rootState, err := withoutMerges.state(fixture.root.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !rootState.Superseded || rootState.Merged || rootState.LineageClosed || rootState.MergeConflict {
		t.Fatalf("unmerged root state = %#v", rootState)
	}
	aState, err := withoutMerges.state(fixture.a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !aState.Superseded || aState.Merged || aState.LineageClosed || aState.MergeConflict {
		t.Fatalf("accepted-but-unmerged revision state = %#v", aState)
	}

	oneMerged, err := buildLineageIndex(append(fixture.events, decision, mergeA, mergeA2))
	if err != nil {
		t.Fatal(err)
	}
	aState, err = oneMerged.state(fixture.a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !aState.Merged || aState.Superseded || aState.LineageClosed || aState.MergeConflict {
		t.Fatalf("merged revision state = %#v", aState)
	}
	requireLineageIDs(t, "own merge events", aState.MergeEventIDs, sortedLineageTestIDs(mergeA.ID, mergeA2.ID))
	bState, err := oneMerged.state(fixture.b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if bState.Merged || bState.Superseded || !bState.LineageClosed || bState.MergeConflict {
		t.Fatalf("open sibling state = %#v", bState)
	}
	requireLineageIDs(t, "one merged candidate", bState.MergedCandidateIDs, []string{fixture.a.ID})

	conflictedEvents := append(append([]StoredEvent(nil), fixture.events...), decision, mergeB, mergeA2, mergeA)
	for order, events := range lineageTestPermutations(conflictedEvents) {
		index, err := buildLineageIndex(events)
		if err != nil {
			t.Fatalf("order %d build: %v", order, err)
		}
		state, err := index.state(fixture.a.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !state.Merged || state.Superseded || !state.LineageClosed || !state.MergeConflict {
			t.Fatalf("order %d conflicting merged state = %#v", order, state)
		}
		requireLineageIDs(t, fmt.Sprintf("order %d merged candidates", order), state.MergedCandidateIDs, sortedLineageTestIDs(fixture.a.ID, fixture.b.ID))
		requireLineageIDs(t, fmt.Sprintf("order %d own merge events", order), state.MergeEventIDs, sortedLineageTestIDs(mergeA.ID, mergeA2.ID))
	}
}

func TestLineageCycleDefense(t *testing.T) {
	first := lineageTestEvent("cycle-a", "proposal.revise", "")
	second := lineageTestEvent("cycle-b", "proposal.revise", first.ID)
	first.Event.Subject = second.ID
	wantID := sortedLineageTestIDs(first.ID, second.ID)[0]
	for order, events := range lineageTestPermutations([]StoredEvent{first, second}) {
		_, err := buildLineageIndex(events)
		if err == nil || !strings.Contains(err.Error(), "cycle") || !strings.Contains(err.Error(), shortID(wantID)) {
			t.Fatalf("order %d cycle error = %v, want cycle and deterministic proposal ID %s", order, err, shortID(wantID))
		}
	}
}

func representativeLineageEvents() []StoredEvent {
	events := make([]StoredEvent, 0, 10000)
	roots := make([]StoredEvent, 900)
	for index := range roots {
		roots[index] = lineageTestEvent(fmt.Sprintf("proposal-%04d", index), "proposal.open", "")
		events = append(events, roots[index])
	}
	for index := 0; index < 100; index++ {
		events = append(events, lineageTestEvent(fmt.Sprintf("revision-%04d", index), "proposal.revise", roots[index].ID))
	}
	for index := 0; index < 9000; index++ {
		events = append(events, lineageTestEvent(fmt.Sprintf("issue-%05d", index), "issue.open", ""))
	}
	return events
}

func TestLineageRepresentativeFixtureConverges(t *testing.T) {
	events := representativeLineageEvents()
	if len(events) != 10000 {
		t.Fatalf("fixture has %d events, want 10000", len(events))
	}
	var want proposalLineageState
	for order, permutation := range lineageTestPermutations(events) {
		index, err := buildLineageIndex(permutation)
		if err != nil {
			t.Fatalf("order %d build: %v", order, err)
		}
		state, err := index.state(events[0].ID)
		if err != nil {
			t.Fatal(err)
		}
		if order == 0 {
			want = state
		} else if !reflect.DeepEqual(state, want) {
			t.Fatalf("order %d state diverged\n got: %#v\nwant: %#v", order, state, want)
		}
	}
}

func BenchmarkLineageRepresentativeIndex(b *testing.B) {
	events := representativeLineageEvents()
	queryID := events[0].ID
	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		index, err := buildLineageIndex(events)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := index.state(queryID); err != nil {
			b.Fatal(err)
		}
	}
}
