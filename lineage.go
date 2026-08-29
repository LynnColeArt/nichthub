package main

import (
	"fmt"
	"sort"
)

type lineageIndex struct {
	proposals       map[string]StoredEvent
	predecessors    map[string]string
	successorsByID  map[string][]string
	mergeEventsByID map[string][]StoredEvent
}

type proposalLineageState struct {
	ProposalID         string
	PredecessorID      string
	SuccessorIDs       []string
	SiblingIDs         []string
	RootID             string
	AncestorIDs        []string
	DescendantIDs      []string
	MemberIDs          []string
	MergeEventIDs      []string
	MergedCandidateIDs []string
	Merged             bool
	Superseded         bool
	LineageClosed      bool
	MergeConflict      bool
}

func buildLineageIndex(events []StoredEvent) (*lineageIndex, error) {
	index := &lineageIndex{
		proposals:       make(map[string]StoredEvent),
		predecessors:    make(map[string]string),
		successorsByID:  make(map[string][]string),
		mergeEventsByID: make(map[string][]StoredEvent),
	}
	for _, stored := range events {
		if isProposalKind(stored.Event.Kind) {
			index.proposals[stored.ID] = stored
		}
	}
	for _, stored := range events {
		switch stored.Event.Kind {
		case "proposal.revise":
			if _, exists := index.proposals[stored.Event.Subject]; !exists {
				return nil, fmt.Errorf("revision %s has unavailable predecessor %s", shortID(stored.ID), stored.Event.Subject)
			}
			index.predecessors[stored.ID] = stored.Event.Subject
			index.successorsByID[stored.Event.Subject] = append(index.successorsByID[stored.Event.Subject], stored.ID)
		case "proposal.merged":
			if _, exists := index.proposals[stored.Event.Subject]; !exists {
				return nil, fmt.Errorf("merge %s has unavailable proposal %s", shortID(stored.ID), stored.Event.Subject)
			}
			index.mergeEventsByID[stored.Event.Subject] = append(index.mergeEventsByID[stored.Event.Subject], stored)
		}
	}
	for id := range index.successorsByID {
		sort.Strings(index.successorsByID[id])
	}
	for id := range index.mergeEventsByID {
		sort.Slice(index.mergeEventsByID[id], func(left, right int) bool {
			return index.mergeEventsByID[id][left].ID < index.mergeEventsByID[id][right].ID
		})
	}
	if err := index.validateAcyclic(); err != nil {
		return nil, err
	}
	return index, nil
}

func (index *lineageIndex) requireProposal(id string) error {
	if _, exists := index.proposals[id]; !exists {
		return fmt.Errorf("proposal %s is not in the lineage index", id)
	}
	return nil
}

func (index *lineageIndex) candidate(id string) (StoredEvent, error) {
	if err := index.requireProposal(id); err != nil {
		return StoredEvent{}, err
	}
	return index.proposals[id], nil
}

func (index *lineageIndex) predecessor(id string) (string, bool, error) {
	if err := index.requireProposal(id); err != nil {
		return "", false, err
	}
	predecessor, exists := index.predecessors[id]
	return predecessor, exists, nil
}

func (index *lineageIndex) successors(id string) ([]string, error) {
	if err := index.requireProposal(id); err != nil {
		return nil, err
	}
	return append([]string(nil), index.successorsByID[id]...), nil
}

func (index *lineageIndex) siblings(id string) ([]string, error) {
	if err := index.requireProposal(id); err != nil {
		return nil, err
	}
	predecessor, exists := index.predecessors[id]
	if !exists {
		return []string{}, nil
	}
	siblings := make([]string, 0, len(index.successorsByID[predecessor])-1)
	for _, candidateID := range index.successorsByID[predecessor] {
		if candidateID != id {
			siblings = append(siblings, candidateID)
		}
	}
	return siblings, nil
}

func (index *lineageIndex) root(id string) (string, error) {
	if err := index.requireProposal(id); err != nil {
		return "", err
	}
	visited := make(map[string]bool)
	for {
		if visited[id] {
			return "", fmt.Errorf("revision lineage contains a cycle at %s", shortID(id))
		}
		visited[id] = true
		predecessor, exists := index.predecessors[id]
		if !exists {
			return id, nil
		}
		id = predecessor
	}
}

func (index *lineageIndex) ancestors(id string) ([]string, error) {
	if err := index.requireProposal(id); err != nil {
		return nil, err
	}
	ancestors := make([]string, 0)
	visited := map[string]bool{id: true}
	for {
		predecessor, exists := index.predecessors[id]
		if !exists {
			break
		}
		if visited[predecessor] {
			return nil, fmt.Errorf("revision lineage contains a cycle at %s", shortID(predecessor))
		}
		visited[predecessor] = true
		ancestors = append(ancestors, predecessor)
		id = predecessor
	}
	sort.Strings(ancestors)
	return ancestors, nil
}

func (index *lineageIndex) descendants(id string) ([]string, error) {
	if err := index.requireProposal(id); err != nil {
		return nil, err
	}
	descendants := make([]string, 0)
	visited := map[string]bool{id: true}
	stack := append([]string(nil), index.successorsByID[id]...)
	for len(stack) > 0 {
		last := len(stack) - 1
		candidateID := stack[last]
		stack = stack[:last]
		if visited[candidateID] {
			return nil, fmt.Errorf("revision lineage contains a cycle at %s", shortID(candidateID))
		}
		visited[candidateID] = true
		descendants = append(descendants, candidateID)
		stack = append(stack, index.successorsByID[candidateID]...)
	}
	sort.Strings(descendants)
	return descendants, nil
}

func (index *lineageIndex) members(id string) ([]string, error) {
	rootID, err := index.root(id)
	if err != nil {
		return nil, err
	}
	descendants, err := index.descendants(rootID)
	if err != nil {
		return nil, err
	}
	members := append(descendants, rootID)
	sort.Strings(members)
	return members, nil
}

func (index *lineageIndex) state(id string) (proposalLineageState, error) {
	if err := index.requireProposal(id); err != nil {
		return proposalLineageState{}, err
	}
	predecessorID := index.predecessors[id]
	successorIDs, err := index.successors(id)
	if err != nil {
		return proposalLineageState{}, err
	}
	siblingIDs, err := index.siblings(id)
	if err != nil {
		return proposalLineageState{}, err
	}
	rootID, err := index.root(id)
	if err != nil {
		return proposalLineageState{}, err
	}
	ancestorIDs, err := index.ancestors(id)
	if err != nil {
		return proposalLineageState{}, err
	}
	descendantIDs, err := index.descendants(id)
	if err != nil {
		return proposalLineageState{}, err
	}
	memberIDs, err := index.members(id)
	if err != nil {
		return proposalLineageState{}, err
	}
	mergeEvents := index.mergeEventsByID[id]
	mergeEventIDs := make([]string, len(mergeEvents))
	for position, merge := range mergeEvents {
		mergeEventIDs[position] = merge.ID
	}
	mergedCandidateIDs := make([]string, 0)
	for _, memberID := range memberIDs {
		if len(index.mergeEventsByID[memberID]) > 0 {
			mergedCandidateIDs = append(mergedCandidateIDs, memberID)
		}
	}
	merged := len(mergeEventIDs) > 0
	lineageClosed := false
	for _, mergedID := range mergedCandidateIDs {
		if mergedID != id {
			lineageClosed = true
			break
		}
	}
	return proposalLineageState{
		ProposalID:         id,
		PredecessorID:      predecessorID,
		SuccessorIDs:       successorIDs,
		SiblingIDs:         siblingIDs,
		RootID:             rootID,
		AncestorIDs:        ancestorIDs,
		DescendantIDs:      descendantIDs,
		MemberIDs:          memberIDs,
		MergeEventIDs:      mergeEventIDs,
		MergedCandidateIDs: mergedCandidateIDs,
		Merged:             merged,
		Superseded:         !merged && len(successorIDs) > 0,
		LineageClosed:      lineageClosed,
		MergeConflict:      len(mergedCandidateIDs) > 1,
	}, nil
}

func (index *lineageIndex) validateAcyclic() error {
	const (
		visiting = 1
		visited  = 2
	)
	state := make(map[string]uint8, len(index.proposals))
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case visiting:
			return fmt.Errorf("revision lineage contains a cycle at %s", shortID(id))
		case visited:
			return nil
		}
		state[id] = visiting
		if predecessor, exists := index.predecessors[id]; exists {
			if err := visit(predecessor); err != nil {
				return err
			}
		}
		state[id] = visited
		return nil
	}
	proposalIDs := make([]string, 0, len(index.proposals))
	for id := range index.proposals {
		proposalIDs = append(proposalIDs, id)
	}
	sort.Strings(proposalIDs)
	for _, id := range proposalIDs {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
