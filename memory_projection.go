package main

import (
	"sort"
	"strings"
)

const (
	memoryLifecycleActive            = "active"
	memoryLifecycleSuperseded        = "superseded"
	memoryLifecycleRetracted         = "retracted"
	memoryLifecycleBranching         = "branching"
	memoryLifecycleDependencyMissing = "dependency-missing"

	memoryApplicabilityApplicable    = "applicable"
	memoryApplicabilityInapplicable  = "inapplicable"
	memoryApplicabilityAnchorMissing = "anchor-missing"
	memoryApplicabilityAnchorInvalid = "anchor-invalid"

	memoryEvidenceResolved = "resolved"
	memoryEvidenceMissing  = "missing"
	memoryEvidenceInvalid  = "invalid"

	memoryTrustQualified      = "qualified"
	memoryTrustActorUntrusted = "actor-untrusted"
	memoryTrustKindUntrusted  = "kind-untrusted"
	memoryTrustPolicyMissing  = "policy-missing"
)

// MemoryProjectionResolver is the read-only boundary between deterministic
// projection and exact Git object inspection. Implementations must never fetch
// missing objects or rewrite refs while answering these queries.
type MemoryProjectionResolver interface {
	Probe(object string) (gitObjectProbe, error)
	IsAncestor(ancestor, descendant string) (isAncestor bool, missing string, err error)
	TreeEntry(commit, path string) (object string, exists bool, err error)
}

type gitMemoryProjectionResolver struct{}

func (gitMemoryProjectionResolver) Probe(object string) (gitObjectProbe, error) {
	return probeExactGitObject(object)
}

func (gitMemoryProjectionResolver) IsAncestor(ancestor, descendant string) (bool, string, error) {
	return exactCommitAncestor(ancestor, descendant)
}

func (gitMemoryProjectionResolver) TreeEntry(commit, path string) (string, bool, error) {
	return exactTreeEntry(commit, path)
}

// MemoryProjectionContext contains only explicit query inputs. PolicyCommit is
// informational; callers load its exact bytes with LoadMemoryProjectionPolicy.
type MemoryProjectionContext struct {
	AtCommit     string                   `json:"atCommit,omitempty"`
	Subject      string                   `json:"subject,omitempty"`
	Path         string                   `json:"path,omitempty"`
	Events       []StoredEvent            `json:"-"`
	MemoryPolicy *MemoryPolicy            `json:"-"`
	PolicyCommit string                   `json:"policyCommit,omitempty"`
	PolicyDigest string                   `json:"policyDigest,omitempty"`
	Resolver     MemoryProjectionResolver `json:"-"`
}

// LoadMemoryProjectionPolicy reads policy only from the exact supplied commit.
func LoadMemoryProjectionPolicy(commit string, context MemoryProjectionContext) (MemoryProjectionContext, error) {
	policy, _, digest, err := loadPolicy(commit)
	if err != nil {
		return MemoryProjectionContext{}, err
	}
	context.PolicyCommit = commit
	context.PolicyDigest = digest
	context.MemoryPolicy = policy.Memory
	return context, nil
}

type MemoryRelationship struct {
	MemoryID  string `json:"memoryId"`
	Stream    string `json:"stream"`
	Actor     string `json:"actor"`
	Operation string `json:"operation"`
	TargetID  string `json:"targetId"`
}

// MemoryDependency is a full, typed, recoverable absence. Invalid data is
// reported separately as a MemoryProjectionDiagnostic.
type MemoryDependency struct {
	Kind      string `json:"kind"`
	OwnerID   string `json:"ownerId"`
	Stream    string `json:"stream"`
	Operation string `json:"operation,omitempty"`
	MissingID string `json:"missingId"`
	Reason    string `json:"reason"`
}

type MemoryProjectionDiagnostic struct {
	MemoryID  string `json:"memoryId"`
	Stream    string `json:"stream"`
	Operation string `json:"operation"`
	TargetID  string `json:"targetId,omitempty"`
	Code      string `json:"code"`
}

type MemoryEvidenceDetail struct {
	Type      string `json:"type"`
	Requested string `json:"requested"`
	OwnerID   string `json:"ownerId"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
}

type MemoryProjectionRow struct {
	ID              string                 `json:"id"`
	Stream          string                 `json:"stream"`
	Actor           string                 `json:"actor"`
	Kind            string                 `json:"kind"`
	ContentDigest   string                 `json:"contentDigest"`
	Anchor          MemoryAnchor           `json:"anchor"`
	Signature       string                 `json:"signature"`
	Lifecycle       string                 `json:"lifecycle"`
	Challengers     []string               `json:"challengers"`
	Successors      []string               `json:"successors"`
	Retractions     []string               `json:"retractions"`
	Applicability   string                 `json:"applicability"`
	Evidence        string                 `json:"evidence"`
	EvidenceDetails []MemoryEvidenceDetail `json:"evidenceDetails"`
	Trust           string                 `json:"trust"`
	Data            MemoryRecord           `json:"data"`
}

type MemoryProjection struct {
	PolicyCommit        string                       `json:"policyCommit,omitempty"`
	PolicyDigest        string                       `json:"policyDigest,omitempty"`
	Rows                []MemoryProjectionRow        `json:"rows"`
	Relationships       []MemoryRelationship         `json:"relationships"`
	MissingDependencies []MemoryDependency           `json:"missingDependencies"`
	Diagnostics         []MemoryProjectionDiagnostic `json:"diagnostics"`
}

// ProjectMemories is a deterministic set projection over already-verified
// memories. It preserves scoped invalid relationships as diagnostics so one
// hostile edge cannot suppress unrelated records.
func ProjectMemories(memories []StoredMemory, context MemoryProjectionContext) MemoryProjection {
	result := MemoryProjection{
		PolicyCommit: context.PolicyCommit,
		PolicyDigest: context.PolicyDigest,
		Rows:         []MemoryProjectionRow{}, Relationships: []MemoryRelationship{},
		MissingDependencies: []MemoryDependency{}, Diagnostics: []MemoryProjectionDiagnostic{},
	}
	resolver := context.Resolver
	if resolver == nil {
		resolver = gitMemoryProjectionResolver{}
	}

	ordered := append([]StoredMemory(nil), memories...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	allByID := make(map[string]StoredMemory, len(ordered))
	acceptedIDs := make(map[string]bool, len(ordered))
	unique := make([]StoredMemory, 0, len(ordered))
	for _, stored := range ordered {
		if _, exists := allByID[stored.ID]; !exists {
			allByID[stored.ID] = stored
			acceptedIDs[stored.ID] = true
			unique = append(unique, stored)
		}
	}
	ordered = unique

	rowIndex := make(map[string]int)
	for _, stored := range ordered {
		if !memoryEnvelopeProducesRecord(stored.Envelope) {
			continue
		}
		if _, duplicate := rowIndex[stored.ID]; duplicate {
			continue
		}
		record := *stored.Envelope.Record
		applicability, anchorDependencies := classifyMemoryApplicability(stored.ID, stored.Envelope.Stream, record, context, resolver)
		evidence, details := classifyMemoryEvidence(stored.ID, record.Evidence, acceptedIDs, context.Events, resolver)
		result.MissingDependencies = append(result.MissingDependencies, anchorDependencies...)
		for _, detail := range details {
			if detail.Status == memoryEvidenceMissing {
				result.MissingDependencies = append(result.MissingDependencies, MemoryDependency{
					Kind: "evidence-" + detail.Type, OwnerID: stored.ID, Stream: stored.Envelope.Stream,
					MissingID: detail.Requested, Reason: detail.Reason,
				})
			}
		}
		rowIndex[stored.ID] = len(result.Rows)
		result.Rows = append(result.Rows, MemoryProjectionRow{
			ID: stored.ID, Stream: stored.Envelope.Stream, Actor: stored.Envelope.Actor,
			Kind: record.Kind, ContentDigest: memoryID([]byte(record.Content)), Anchor: record.Anchor,
			Signature: "valid", Lifecycle: memoryLifecycleActive,
			Challengers: []string{}, Successors: []string{}, Retractions: []string{},
			Applicability: applicability, Evidence: evidence, EvidenceDetails: details,
			Trust: classifyMemoryTrust(context.MemoryPolicy, stored.Envelope.Actor, record.Kind), Data: record,
		})
	}

	missingByOwner := make(map[string]bool)
	for _, source := range ordered {
		envelope := source.Envelope
		if envelope.Operation == memoryOperationRecord {
			continue
		}
		if err := validateMemoryEnvelope(envelope); err != nil || !validMemoryID(source.ID) {
			result.Diagnostics = append(result.Diagnostics, relationshipDiagnostic(source, "malformed-relationship"))
			continue
		}
		if source.ID == envelope.Target {
			result.Diagnostics = append(result.Diagnostics, relationshipDiagnostic(source, "self-target"))
			continue
		}
		target, exists := allByID[envelope.Target]
		if !exists {
			result.MissingDependencies = append(result.MissingDependencies, MemoryDependency{
				Kind: "lifecycle-target", OwnerID: source.ID, Stream: envelope.Stream,
				Operation: envelope.Operation, MissingID: envelope.Target, Reason: "target-unavailable",
			})
			missingByOwner[source.ID] = true
			continue
		}
		if !memoryEnvelopeProducesRecord(target.Envelope) {
			result.Diagnostics = append(result.Diagnostics, relationshipDiagnostic(source, "wrong-target-kind"))
			continue
		}
		if (envelope.Operation == memoryOperationSupersede || envelope.Operation == memoryOperationRetract) && envelope.Actor != target.Envelope.Actor {
			result.Diagnostics = append(result.Diagnostics, relationshipDiagnostic(source, "actor-mismatch"))
			continue
		}
		result.Relationships = append(result.Relationships, MemoryRelationship{
			MemoryID: source.ID, Stream: envelope.Stream, Actor: envelope.Actor,
			Operation: envelope.Operation, TargetID: envelope.Target,
		})
	}

	for _, edge := range result.Relationships {
		index, exists := rowIndex[edge.TargetID]
		if !exists {
			continue
		}
		row := &result.Rows[index]
		switch edge.Operation {
		case memoryOperationSupersede:
			row.Successors = append(row.Successors, edge.MemoryID)
		case memoryOperationRetract:
			row.Retractions = append(row.Retractions, edge.MemoryID)
		case memoryOperationChallenge:
			row.Challengers = append(row.Challengers, edge.MemoryID)
		}
	}
	for index := range result.Rows {
		row := &result.Rows[index]
		sort.Strings(row.Challengers)
		sort.Strings(row.Successors)
		sort.Strings(row.Retractions)
		switch {
		case missingByOwner[row.ID]:
			row.Lifecycle = memoryLifecycleDependencyMissing
		case len(row.Retractions) > 0:
			row.Lifecycle = memoryLifecycleRetracted
		case len(row.Successors) > 1:
			row.Lifecycle = memoryLifecycleBranching
		case len(row.Successors) == 1:
			row.Lifecycle = memoryLifecycleSuperseded
		default:
			row.Lifecycle = memoryLifecycleActive
		}
	}

	sortProjection(result.Rows, result.Relationships, result.MissingDependencies, result.Diagnostics)
	return result
}

func memoryEnvelopeProducesRecord(envelope MemoryEnvelope) bool {
	if envelope.Record == nil || (envelope.Operation != memoryOperationRecord && envelope.Operation != memoryOperationSupersede) {
		return false
	}
	return validateMemoryEnvelope(envelope) == nil
}

func relationshipDiagnostic(stored StoredMemory, code string) MemoryProjectionDiagnostic {
	return MemoryProjectionDiagnostic{
		MemoryID: stored.ID, Stream: stored.Envelope.Stream, Operation: stored.Envelope.Operation,
		TargetID: stored.Envelope.Target, Code: code,
	}
}

func classifyMemoryApplicability(owner, stream string, record MemoryRecord, context MemoryProjectionContext, resolver MemoryProjectionResolver) (string, []MemoryDependency) {
	anchorProbe, err := resolver.Probe(record.Anchor.Commit)
	if err != nil || (anchorProbe.Exists && anchorProbe.Type != "commit") {
		return memoryApplicabilityAnchorInvalid, nil
	}
	if !anchorProbe.Exists {
		return memoryApplicabilityAnchorMissing, []MemoryDependency{{
			Kind: "anchor-commit", OwnerID: owner, Stream: stream,
			MissingID: record.Anchor.Commit, Reason: "anchor-commit-unavailable",
		}}
	}
	for _, anchored := range record.Anchor.Paths {
		object, exists, err := resolver.TreeEntry(record.Anchor.Commit, anchored.Path)
		if err != nil {
			return memoryApplicabilityAnchorInvalid, nil
		}
		if anchored.Blob == "absent" {
			if exists {
				return memoryApplicabilityAnchorInvalid, nil
			}
		} else if !exists || object != anchored.Blob {
			return memoryApplicabilityAnchorInvalid, nil
		}
	}

	requested := context.AtCommit
	if requested == "" {
		requested = record.Anchor.Commit
	}
	if requested != record.Anchor.Commit {
		requestedProbe, err := resolver.Probe(requested)
		if err != nil || (requestedProbe.Exists && requestedProbe.Type != "commit") {
			return memoryApplicabilityAnchorInvalid, nil
		}
		if !requestedProbe.Exists {
			return memoryApplicabilityAnchorMissing, []MemoryDependency{{
				Kind: "query-commit", OwnerID: owner, Stream: stream,
				MissingID: requested, Reason: "query-commit-unavailable",
			}}
		}
	}

	applicable := false
	switch record.Applicability.Mode {
	case memoryApplicabilityExact:
		applicable = requested == record.Anchor.Commit
	case memoryApplicabilityDescendants:
		ancestor, missing, err := resolver.IsAncestor(record.Anchor.Commit, requested)
		if err != nil {
			return memoryApplicabilityAnchorInvalid, nil
		}
		if missing != "" {
			return memoryApplicabilityAnchorMissing, []MemoryDependency{{
				Kind: "anchor-ancestry", OwnerID: owner, Stream: stream,
				MissingID: missing, Reason: "ancestry-commit-unavailable",
			}}
		}
		applicable = ancestor
	case memoryApplicabilitySubject:
		applicable = context.Subject != "" && context.Subject == record.Applicability.Subject && context.Subject == record.Anchor.Subject
	default:
		return memoryApplicabilityAnchorInvalid, nil
	}
	if applicable && context.Subject != "" && record.Applicability.Mode != memoryApplicabilitySubject && context.Subject != record.Anchor.Subject {
		applicable = false
	}
	if applicable && context.Path != "" && !recordAnchorsPath(record, context.Path) {
		applicable = false
	}
	if applicable {
		return memoryApplicabilityApplicable, nil
	}
	return memoryApplicabilityInapplicable, nil
}

func recordAnchorsPath(record MemoryRecord, query string) bool {
	if !validMemoryPath(query) {
		return false
	}
	for _, anchored := range record.Anchor.Paths {
		if anchored.Path == query {
			return true
		}
	}
	return false
}

func classifyMemoryEvidence(owner string, evidence []string, memories map[string]bool, events []StoredEvent, resolver MemoryProjectionResolver) (string, []MemoryEvidenceDetail) {
	eventIDs := make(map[string]bool, len(events))
	for _, event := range events {
		eventIDs[event.ID] = true
	}
	details := make([]MemoryEvidenceDetail, 0, len(evidence))
	overall := memoryEvidenceResolved
	for _, typed := range evidence {
		kind, requested, found := strings.Cut(typed, ":")
		detail := MemoryEvidenceDetail{Type: kind, Requested: requested, OwnerID: owner}
		if !found || !validTypedMemoryEvidence(typed) {
			detail.Status, detail.Reason = memoryEvidenceInvalid, "malformed-id"
		} else {
			switch kind {
			case "git":
				probe, err := resolver.Probe(requested)
				switch {
				case err != nil:
					detail.Status, detail.Reason = memoryEvidenceInvalid, "object-inspection-failed"
				case !probe.Exists:
					detail.Status, detail.Reason = memoryEvidenceMissing, "object-unavailable"
				default:
					detail.Status, detail.Reason = memoryEvidenceResolved, "exact-object"
				}
			case "event":
				if eventIDs[requested] {
					detail.Status, detail.Reason = memoryEvidenceResolved, "accepted-event"
				} else {
					detail.Status, detail.Reason = memoryEvidenceMissing, "event-unavailable"
				}
			case "memory":
				if memories[requested] {
					detail.Status, detail.Reason = memoryEvidenceResolved, "accepted-memory"
				} else {
					detail.Status, detail.Reason = memoryEvidenceMissing, "memory-unavailable"
				}
			default:
				detail.Status, detail.Reason = memoryEvidenceInvalid, "unsupported-type"
			}
		}
		if detail.Status == memoryEvidenceInvalid {
			overall = memoryEvidenceInvalid
		} else if detail.Status == memoryEvidenceMissing && overall != memoryEvidenceInvalid {
			overall = memoryEvidenceMissing
		}
		details = append(details, detail)
	}
	sort.Slice(details, func(i, j int) bool {
		if details[i].Type != details[j].Type {
			return details[i].Type < details[j].Type
		}
		return details[i].Requested < details[j].Requested
	})
	return overall, details
}

func sortProjection(rows []MemoryProjectionRow, relationships []MemoryRelationship, missing []MemoryDependency, diagnostics []MemoryProjectionDiagnostic) {
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	sort.Slice(relationships, func(i, j int) bool {
		if relationships[i].TargetID != relationships[j].TargetID {
			return relationships[i].TargetID < relationships[j].TargetID
		}
		if relationships[i].Operation != relationships[j].Operation {
			return relationships[i].Operation < relationships[j].Operation
		}
		return relationships[i].MemoryID < relationships[j].MemoryID
	})
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].MissingID != missing[j].MissingID {
			return missing[i].MissingID < missing[j].MissingID
		}
		if missing[i].OwnerID != missing[j].OwnerID {
			return missing[i].OwnerID < missing[j].OwnerID
		}
		if missing[i].Kind != missing[j].Kind {
			return missing[i].Kind < missing[j].Kind
		}
		return missing[i].Operation < missing[j].Operation
	})
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].MemoryID != diagnostics[j].MemoryID {
			return diagnostics[i].MemoryID < diagnostics[j].MemoryID
		}
		return diagnostics[i].Code < diagnostics[j].Code
	})
}
