package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	memoryProtocolVersion = "nh-memory/0"

	memoryOperationRecord    = "record"
	memoryOperationSupersede = "supersede"
	memoryOperationRetract   = "retract"
	memoryOperationChallenge = "challenge"

	memoryKindObservation  = "observation"
	memoryKindDecision     = "decision"
	memoryKindAssumption   = "assumption"
	memoryKindAttempt      = "attempt"
	memoryKindVerification = "verification"
	memoryKindHandoff      = "handoff"

	memoryApplicabilityExact       = "exact"
	memoryApplicabilityDescendants = "descendants"
	memoryApplicabilitySubject     = "subject"

	maxMemoryContentBytes        = 65_536
	maxMemoryTopics              = 32
	maxMemoryTopicBytes          = 128
	maxMemoryEvidence            = 64
	maxMemoryPaths               = 128
	maxMemoryPathBytes           = 4_096
	maxMemoryTotalPathBytes      = 65_536
	maxMemoryHandoffEntries      = 64
	maxMemoryHandoffEntryBytes   = 4_096
	maxMemoryHandoffBytes        = 65_536
	maxMemoryReasonBytes         = 128
	maxMemoryAttemptOutcomeBytes = 64
	maxMemoryActorNameBytes      = 256
)

// MemoryEnvelope is the immutable, signed nh-memory/0 wire object. Field order
// is part of its canonical encoding; do not reorder fields without a protocol
// version change.
type MemoryEnvelope struct {
	Protocol  string        `json:"protocol"`
	Operation string        `json:"operation"`
	Actor     string        `json:"actor"`
	ActorName string        `json:"actorName"`
	PublicKey string        `json:"publicKey"`
	Stream    string        `json:"stream"`
	Sequence  uint64        `json:"sequence"`
	Timestamp string        `json:"timestamp"`
	Previous  string        `json:"previous,omitempty"`
	Record    *MemoryRecord `json:"record,omitempty"`
	Target    string        `json:"target,omitempty"`
	Reason    string        `json:"reason,omitempty"`
	Evidence  []string      `json:"evidence,omitempty"`
}

type MemoryRecord struct {
	Kind           string         `json:"kind"`
	Content        string         `json:"content"`
	Anchor         MemoryAnchor   `json:"anchor"`
	Applicability  Applicability  `json:"applicability"`
	Topics         []string       `json:"topics"`
	Evidence       []string       `json:"evidence"`
	AttemptOutcome string         `json:"attemptOutcome,omitempty"`
	Handoff        *HandoffFields `json:"handoff,omitempty"`
}

type MemoryAnchor struct {
	Commit  string       `json:"commit"`
	Paths   []PathAnchor `json:"paths,omitempty"`
	Subject string       `json:"subject,omitempty"`
}

type PathAnchor struct {
	Path string `json:"path"`
	Blob string `json:"blob"`
}

type Applicability struct {
	Mode    string `json:"mode"`
	Subject string `json:"subject,omitempty"`
}

type HandoffFields struct {
	Completed   []string `json:"completed"`
	Assumptions []string `json:"assumptions"`
	Blockers    []string `json:"blockers"`
	NextActions []string `json:"nextActions"`
}

func memoryID(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func defaultMemoryStream(actor string) string {
	return memoryID([]byte("nh-memory-stream-v0\x00" + actor + "\x00default"))
}

func validMemoryID(id string) bool {
	return validMemoryDigestID(id)
}

func validMemoryStreamID(id string) bool {
	return validMemoryDigestID(id)
}

func validMemoryDigestID(id string) bool {
	if len(id) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(id, "sha256:") {
		return false
	}
	encoded := strings.TrimPrefix(id, "sha256:")
	if encoded != strings.ToLower(encoded) {
		return false
	}
	decoded, err := hex.DecodeString(encoded)
	return err == nil && len(decoded) == sha256.Size
}

func newMemoryEnvelope(identity *Identity, operation, stream string, sequence uint64, previous string) MemoryEnvelope {
	return MemoryEnvelope{
		Protocol:  memoryProtocolVersion,
		Operation: operation,
		Actor:     identity.Actor,
		ActorName: identity.Name,
		PublicKey: identity.PublicKey,
		Stream:    stream,
		Sequence:  sequence,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Previous:  previous,
	}
}

func encodeMemoryEnvelope(envelope MemoryEnvelope) ([]byte, error) {
	if err := validateMemoryEnvelope(envelope); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("encode memory envelope: %w", err)
	}
	return payload, nil
}

func encodeAndSignMemory(envelope MemoryEnvelope, identity *Identity) ([]byte, []byte, error) {
	if identity == nil || envelope.Actor != identity.Actor || envelope.PublicKey != identity.PublicKey {
		return nil, nil, fmt.Errorf("memory identity does not match its signer")
	}
	if err := validateIdentity(identity); err != nil {
		return nil, nil, fmt.Errorf("invalid memory signer")
	}
	payload, err := encodeMemoryEnvelope(envelope)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := identity.privateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("invalid memory signer")
	}
	return payload, ed25519.Sign(privateKey, payload), nil
}

func decodeMemoryEnvelope(payload []byte) (MemoryEnvelope, error) {
	if !utf8.Valid(payload) {
		return MemoryEnvelope{}, fmt.Errorf("invalid memory JSON encoding")
	}
	var envelope MemoryEnvelope
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return MemoryEnvelope{}, fmt.Errorf("invalid memory JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return MemoryEnvelope{}, fmt.Errorf("invalid memory JSON trailing data")
	}
	if err := validateMemoryEnvelope(envelope); err != nil {
		return MemoryEnvelope{}, err
	}
	canonical, err := json.Marshal(envelope)
	if err != nil || !bytes.Equal(canonical, payload) {
		return MemoryEnvelope{}, fmt.Errorf("memory JSON is not canonical")
	}
	return envelope, nil
}

func verifyMemory(payload, signature []byte) (MemoryEnvelope, string, error) {
	envelope, err := decodeMemoryEnvelope(payload)
	if err != nil {
		return MemoryEnvelope{}, "", err
	}
	publicBytes, err := base64.RawStdEncoding.DecodeString(envelope.PublicKey)
	if err != nil || len(publicBytes) != ed25519.PublicKeySize {
		return MemoryEnvelope{}, "", fmt.Errorf("invalid memory public key")
	}
	publicKey := ed25519.PublicKey(publicBytes)
	if actorForPublicKey(publicKey) != envelope.Actor {
		return MemoryEnvelope{}, "", fmt.Errorf("memory actor does not match public key")
	}
	if len(signature) != ed25519.SignatureSize || !ed25519.Verify(publicKey, payload, signature) {
		return MemoryEnvelope{}, "", fmt.Errorf("invalid memory signature")
	}
	return envelope, memoryID(payload), nil
}

func validateMemoryEnvelope(envelope MemoryEnvelope) error {
	if envelope.Protocol != memoryProtocolVersion {
		return fmt.Errorf("memory protocol is unsupported")
	}
	if !validActorFingerprint(envelope.Actor) {
		return fmt.Errorf("memory actor is invalid")
	}
	if !validBoundedDisplayName(envelope.ActorName) {
		return fmt.Errorf("memory actorName is invalid")
	}
	publicBytes, err := base64.RawStdEncoding.DecodeString(envelope.PublicKey)
	if err != nil || len(publicBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("memory publicKey is invalid")
	}
	if base64.RawStdEncoding.EncodeToString(publicBytes) != envelope.PublicKey {
		return fmt.Errorf("memory publicKey is not canonical")
	}
	if actorForPublicKey(ed25519.PublicKey(publicBytes)) != envelope.Actor {
		return fmt.Errorf("memory actor does not match publicKey")
	}
	if !validMemoryStreamID(envelope.Stream) {
		return fmt.Errorf("memory stream is invalid")
	}
	if envelope.Sequence == 0 {
		return fmt.Errorf("memory sequence must be positive")
	}
	if envelope.Sequence == 1 && envelope.Previous != "" {
		return fmt.Errorf("memory previous must be absent at sequence 1")
	}
	if envelope.Sequence > 1 && !validMemoryID(envelope.Previous) {
		return fmt.Errorf("memory previous is required after sequence 1")
	}
	parsed, err := time.Parse(time.RFC3339Nano, envelope.Timestamp)
	if err != nil || parsed.Format(time.RFC3339Nano) != envelope.Timestamp {
		return fmt.Errorf("memory timestamp is invalid")
	}

	switch envelope.Operation {
	case memoryOperationRecord:
		if envelope.Record == nil || envelope.Target != "" || envelope.Reason != "" || len(envelope.Evidence) != 0 {
			return fmt.Errorf("memory record operation has invalid fields")
		}
	case memoryOperationSupersede:
		if envelope.Record == nil || !validMemoryID(envelope.Target) || envelope.Reason != "" || len(envelope.Evidence) != 0 {
			return fmt.Errorf("memory supersede operation has invalid fields")
		}
	case memoryOperationRetract:
		if envelope.Record != nil || !validMemoryID(envelope.Target) || !validMemoryToken(envelope.Reason, maxMemoryReasonBytes) || len(envelope.Evidence) != 0 {
			return fmt.Errorf("memory retract operation has invalid fields")
		}
	case memoryOperationChallenge:
		if envelope.Record != nil || !validMemoryID(envelope.Target) || !validMemoryToken(envelope.Reason, maxMemoryReasonBytes) {
			return fmt.Errorf("memory challenge operation has invalid fields")
		}
		if err := validateMemoryEvidence(envelope.Evidence, true); err != nil {
			return fmt.Errorf("memory challenge evidence is invalid")
		}
	default:
		return fmt.Errorf("memory operation is unsupported")
	}
	if envelope.Record != nil {
		if err := validateMemoryRecord(*envelope.Record); err != nil {
			return err
		}
	}
	return nil
}

func validateMemoryRecord(record MemoryRecord) error {
	if !utf8.ValidString(record.Content) || strings.TrimSpace(record.Content) == "" || len(record.Content) > maxMemoryContentBytes {
		return fmt.Errorf("memory record content is invalid")
	}
	if record.Topics == nil {
		return fmt.Errorf("memory record topics are required")
	}
	if record.Evidence == nil {
		return fmt.Errorf("memory record evidence is required")
	}
	if err := validateMemoryAnchor(record.Anchor); err != nil {
		return err
	}
	if err := validateMemoryApplicability(record.Applicability, record.Anchor.Subject); err != nil {
		return err
	}
	if err := validateMemoryTopics(record.Topics); err != nil {
		return err
	}
	if err := validateMemoryEvidence(record.Evidence, true); err != nil {
		return fmt.Errorf("memory record evidence is invalid")
	}

	switch record.Kind {
	case memoryKindObservation, memoryKindDecision, memoryKindAssumption:
		if record.AttemptOutcome != "" || record.Handoff != nil {
			return fmt.Errorf("memory record kind has invalid fields")
		}
	case memoryKindAttempt:
		if !validMemoryToken(record.AttemptOutcome, maxMemoryAttemptOutcomeBytes) || record.Handoff != nil {
			return fmt.Errorf("memory attempt fields are invalid")
		}
	case memoryKindVerification:
		if len(record.Evidence) == 0 || record.AttemptOutcome != "" || record.Handoff != nil {
			return fmt.Errorf("memory verification fields are invalid")
		}
	case memoryKindHandoff:
		if record.AttemptOutcome != "" || record.Handoff == nil {
			return fmt.Errorf("memory handoff fields are invalid")
		}
		if err := validateMemoryHandoff(*record.Handoff); err != nil {
			return err
		}
	default:
		return fmt.Errorf("memory record kind is unsupported")
	}
	return nil
}

func validateMemoryAnchor(anchor MemoryAnchor) error {
	if !validMemoryGitOID(anchor.Commit) {
		return fmt.Errorf("memory anchor commit is invalid")
	}
	if anchor.Subject != "" && !validMemorySubject(anchor.Subject) {
		return fmt.Errorf("memory anchor subject is invalid")
	}
	if len(anchor.Paths) > maxMemoryPaths {
		return fmt.Errorf("memory anchor paths exceed limit")
	}
	seen := make(map[string]struct{}, len(anchor.Paths))
	last := ""
	total := 0
	for index, item := range anchor.Paths {
		if !validMemoryPath(item.Path) {
			return fmt.Errorf("memory anchor path is invalid")
		}
		if _, exists := seen[item.Path]; exists || (index > 0 && item.Path <= last) {
			return fmt.Errorf("memory anchor paths are duplicate or unsorted")
		}
		if item.Blob != "absent" && !validMemoryGitOID(item.Blob) {
			return fmt.Errorf("memory anchor blob is invalid")
		}
		seen[item.Path] = struct{}{}
		last = item.Path
		total += len(item.Path)
	}
	if total > maxMemoryTotalPathBytes {
		return fmt.Errorf("memory anchor path bytes exceed limit")
	}
	return nil
}

func validateMemoryApplicability(applicability Applicability, anchorSubject string) error {
	switch applicability.Mode {
	case memoryApplicabilityExact, memoryApplicabilityDescendants:
		if applicability.Subject != "" {
			return fmt.Errorf("memory applicability subject is prohibited")
		}
	case memoryApplicabilitySubject:
		if !validMemorySubject(applicability.Subject) || applicability.Subject != anchorSubject {
			return fmt.Errorf("memory applicability subject must match anchor")
		}
	default:
		return fmt.Errorf("memory applicability mode is unsupported")
	}
	return nil
}

func validateMemoryTopics(topics []string) error {
	if len(topics) > maxMemoryTopics {
		return fmt.Errorf("memory record topics exceed limit")
	}
	last := ""
	for index, topic := range topics {
		if !validNormalizedMemoryTopic(topic) || (index > 0 && topic <= last) {
			return fmt.Errorf("memory record topics are invalid, duplicate, or unsorted")
		}
		last = topic
	}
	return nil
}

func validateMemoryEvidence(evidence []string, allowEmpty bool) error {
	if (!allowEmpty && len(evidence) == 0) || len(evidence) > maxMemoryEvidence {
		return fmt.Errorf("memory evidence count is invalid")
	}
	seen := make(map[string]struct{}, len(evidence))
	for _, id := range evidence {
		if !validTypedMemoryEvidence(id) {
			return fmt.Errorf("memory evidence ID is invalid")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("memory evidence contains a duplicate")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateMemoryHandoff(handoff HandoffFields) error {
	lists := [][]string{handoff.Completed, handoff.Assumptions, handoff.Blockers, handoff.NextActions}
	total := 0
	for _, list := range lists {
		if list == nil || len(list) > maxMemoryHandoffEntries {
			return fmt.Errorf("memory handoff list is missing or exceeds limit")
		}
		for _, entry := range list {
			if !utf8.ValidString(entry) || strings.TrimSpace(entry) == "" || len(entry) > maxMemoryHandoffEntryBytes {
				return fmt.Errorf("memory handoff entry is invalid")
			}
			total += len(entry)
		}
	}
	if total > maxMemoryHandoffBytes {
		return fmt.Errorf("memory handoff bytes exceed limit")
	}
	return nil
}

func validBoundedDisplayName(value string) bool {
	if !utf8.ValidString(value) || strings.TrimSpace(value) == "" || len(value) > maxMemoryActorNameBytes {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validMemoryToken(value string, maximum int) bool {
	if value == "" || len(value) > maximum || value != strings.ToLower(value) {
		return false
	}
	for index, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r == '-' && index > 0) {
			continue
		}
		return false
	}
	return !strings.HasSuffix(value, "-")
}

func validNormalizedMemoryTopic(topic string) bool {
	if !utf8.ValidString(topic) || topic == "" || len(topic) > maxMemoryTopicBytes || topic != strings.ToLower(strings.TrimSpace(topic)) {
		return false
	}
	for _, r := range topic {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validMemoryPath(value string) bool {
	if !utf8.ValidString(value) || value == "" || len(value) > maxMemoryPathBytes || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || path.Clean(value) != value || value == "." {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validMemoryGitOID(id string) bool {
	if id != strings.ToLower(id) || (len(id) != 40 && len(id) != 64) {
		return false
	}
	decoded, err := hex.DecodeString(id)
	return err == nil && (len(decoded) == 20 || len(decoded) == 32)
}

func validMemorySubject(subject string) bool {
	for _, prefix := range []string{"issue:", "proposal:", "event:", "policy:", "pipeline:", "run:"} {
		if strings.HasPrefix(subject, prefix) {
			return validMemoryID(strings.TrimPrefix(subject, prefix))
		}
	}
	return false
}

func validTypedMemoryEvidence(id string) bool {
	if strings.HasPrefix(id, "git:") {
		return validMemoryGitOID(strings.TrimPrefix(id, "git:"))
	}
	if strings.HasPrefix(id, "event:") {
		return validMemoryID(strings.TrimPrefix(id, "event:"))
	}
	if strings.HasPrefix(id, "memory:") {
		return validMemoryID(strings.TrimPrefix(id, "memory:"))
	}
	return false
}
