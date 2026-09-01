package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	memoryIndexVersion  = 0
	maxMemoryIndexBytes = 256 << 20

	memoryIndexMissing      = "missing"
	memoryIndexCorrupt      = "corrupt"
	memoryIndexIncompatible = "incompatible"
	memoryIndexStale        = "stale"
)

// MemoryIndexV0 is private, derived state. Its deliberately small top-level
// schema makes it impossible to confuse index metadata with canonical memory.
type MemoryIndexV0 struct {
	Version           int                    `json:"version"`
	SourceFingerprint string                 `json:"sourceFingerprint"`
	Records           []MemoryIndexRecordV0  `json:"records"`
	Tokens            []MemoryTokenPostingV0 `json:"tokens"`
	// projectionBinding is deliberately not persisted: raw cache bytes are not
	// queryable until rebuild or verification binds every applicability input.
	projectionBinding memoryIndexProjectionBinding
}

type memoryIndexProjectionBinding struct {
	AtCommit string
	Subject  string
	Path     string
}

// MemoryIndexRecordV0 preserves a verified projection and the signed timestamp
// needed for deterministic recall ordering. Data remains inert author data.
type MemoryIndexRecordV0 struct {
	ID              string                 `json:"id"`
	Stream          string                 `json:"stream"`
	Actor           string                 `json:"actor"`
	SignedTimestamp string                 `json:"signedTimestamp"`
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
	Dependencies    []MemoryDependency     `json:"dependencies"`
	Trust           string                 `json:"trust"`
	Data            MemoryRecord           `json:"data"`
}

type MemoryTokenPostingV0 struct {
	Token     string   `json:"token"`
	MemoryIDs []string `json:"memoryIds"`
}

// MemoryIndexQuery is an internal candidate filter. Values inside one category
// are a union; topics and lexical terms are intersections across all values.
type MemoryIndexQuery struct {
	AtCommit        string
	Subject         string
	Path            string
	Topics          []string
	Kinds           []string
	Actors          []string
	Lifecycles      []string
	Applicabilities []string
	Trust           []string
	Query           string
}

// memoryIndexSource is emitted only by the verified WP02 collection seam.
// Ref and Head remain separate from deduplicated records for the fingerprint.
type memoryIndexSource struct {
	Ref      string
	Head     string
	Accepted bool
	Memory   []StoredMemory
}

type memoryIndexRebuildOptions struct {
	GitDir  string
	Context MemoryProjectionContext
	Collect func(string) ([]memoryIndexSource, error)
	Project func([]StoredMemory, MemoryProjectionContext) MemoryProjection
	Write   func(string, []byte) error
}

type MemoryIndexError struct {
	Kind  string
	Cause error
}

func (e *MemoryIndexError) Error() string {
	if e.Cause == nil {
		return "memory index " + e.Kind
	}
	return "memory index " + e.Kind + ": " + e.Cause.Error()
}

func (e *MemoryIndexError) Unwrap() error { return e.Cause }

func isMemoryIndexError(err error, kind string) bool {
	var indexError *MemoryIndexError
	return errors.As(err, &indexError) && indexError.Kind == kind
}

func memoryIndexPath() (string, error) {
	gitDir, err := requireGitRepository()
	if err != nil {
		return "", err
	}
	return memoryIndexPathAtGitDir(gitDir)
}

func memoryIndexPathAtGitDir(gitDir string) (string, error) {
	if gitDir == "" || !filepath.IsAbs(gitDir) || filepath.Clean(gitDir) != gitDir {
		return "", fmt.Errorf("memory index requires an absolute resolved Git directory")
	}
	info, err := os.Lstat(gitDir)
	if err != nil {
		return "", fmt.Errorf("memory index Git directory is unavailable")
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", fmt.Errorf("memory index Git directory is unsafe")
	}
	return filepath.Join(gitDir, "hn", "memory", "index-v0.json"), nil
}

func ensureMemoryIndexDirectory(gitDir string) error {
	for _, directory := range []string{filepath.Join(gitDir, "hn"), filepath.Join(gitDir, "hn", "memory")} {
		if err := ensurePrivateDirectory(directory); err != nil {
			return fmt.Errorf("prepare private memory index directory: %w", err)
		}
	}
	return nil
}

func validateMemoryIndexDirectory(gitDir string) error {
	for _, directory := range []string{filepath.Join(gitDir, "hn"), filepath.Join(gitDir, "hn", "memory")} {
		info, err := os.Lstat(directory)
		if err != nil {
			return err
		}
		if err := validatePrivateDirectory(directory, info); err != nil {
			return err
		}
	}
	return nil
}

// memoryIndexSourceFingerprint hashes a length-delimited v0 domain, the exact
// policy digest, and sorted canonical (ref, head) pairs.
func memoryIndexSourceFingerprint(sources []memoryIndexSource, policyDigest string) (string, error) {
	if !validMemoryDigestID(policyDigest) {
		return "", fmt.Errorf("invalid memory index policy digest")
	}
	ordered := append([]memoryIndexSource(nil), sources...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Ref != ordered[j].Ref {
			return ordered[i].Ref < ordered[j].Ref
		}
		return ordered[i].Head < ordered[j].Head
	})
	var preimage bytes.Buffer
	writeMemoryIndexFingerprintPart(&preimage, "hn-memory-index-source-v0")
	writeMemoryIndexFingerprintPart(&preimage, policyDigest)
	seen := make(map[string]bool, len(ordered))
	for _, source := range ordered {
		if err := validateMemoryIndexSourceIdentity(source); err != nil {
			return "", err
		}
		if seen[source.Ref] {
			return "", fmt.Errorf("duplicate memory index source ref")
		}
		seen[source.Ref] = true
		writeMemoryIndexFingerprintPart(&preimage, source.Ref)
		writeMemoryIndexFingerprintPart(&preimage, source.Head)
	}
	sum := sha256.Sum256(preimage.Bytes())
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func writeMemoryIndexFingerprintPart(buffer *bytes.Buffer, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	buffer.Write(length[:])
	buffer.WriteString(value)
}

func validateMemoryIndexSourceIdentity(source memoryIndexSource) error {
	if !validFullGitObjectID(source.Head) {
		return fmt.Errorf("invalid memory index source head")
	}
	if _, _, ok := parseMemoryRef(source.Ref); ok {
		if source.Accepted {
			return fmt.Errorf("local memory index source marked accepted")
		}
		return nil
	}
	if _, _, _, ok := parseAcceptedMemoryRef(source.Ref); !ok || !source.Accepted {
		return fmt.Errorf("memory index source is not canonical")
	}
	return nil
}

func collectMemoryIndexSourcesAt(gitDir string) ([]memoryIndexSource, error) {
	text, err := gitTextAt(gitDir, "for-each-ref", "--format=%(refname) %(objectname)", "refs/hn/memory", "refs/hn/remotes")
	if err != nil {
		return nil, fmt.Errorf("collect verified memory sources: %w", err)
	}
	if text == "" {
		return []memoryIndexSource{}, nil
	}
	sources := make([]memoryIndexSource, 0)
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid memory ref advertisement")
		}
		ref, head := fields[0], fields[1]
		source := memoryIndexSource{Ref: ref, Head: head}
		var streamSource memoryStreamSource
		if strings.HasPrefix(ref, memoryRefPrefix) {
			actor, stream, ok := parseMemoryRef(ref)
			if !ok {
				return nil, fmt.Errorf("malformed memory ref %s", safeDiagnostic(ref))
			}
			streamSource = memoryStreamSource{Ref: ref, Actor: actor, Stream: stream, Head: head}
		} else if strings.HasPrefix(ref, acceptedMemoryRefPrefix) {
			parts := strings.Split(strings.TrimPrefix(ref, acceptedMemoryRefPrefix), "/")
			if len(parts) < 2 || parts[1] != "memory" {
				continue
			}
			remote, actor, stream, ok := parseAcceptedMemoryRef(ref)
			if !ok {
				return nil, fmt.Errorf("malformed accepted memory ref %s", safeDiagnostic(ref))
			}
			source.Accepted = true
			streamSource = memoryStreamSource{Ref: ref, Remote: remote, Actor: actor, Stream: stream, Head: head, Accepted: true}
		} else {
			// Only refs in the exact accepted-memory shape are eligible. Other
			// collaboration or remote namespaces are intentionally ignored.
			continue
		}
		if err := validateMemoryIndexSourceIdentity(source); err != nil {
			return nil, err
		}
		memories, err := loadMemoryStreamAt(gitDir, streamSource)
		if err != nil {
			return nil, fmt.Errorf("verify memory index source %s: %w", safeDiagnostic(ref), err)
		}
		source.Memory = memories
		sources = append(sources, source)
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Ref < sources[j].Ref })
	return sources, nil
}

func rebuildMemoryIndex(context MemoryProjectionContext) (MemoryIndexV0, error) {
	gitDir, err := requireGitRepository()
	if err != nil {
		return MemoryIndexV0{}, err
	}
	return rebuildMemoryIndexV0(memoryIndexRebuildOptions{GitDir: gitDir, Context: context})
}

func rebuildMemoryIndexV0(options memoryIndexRebuildOptions) (MemoryIndexV0, error) {
	options = defaultMemoryIndexOptions(options)
	path, err := memoryIndexPathAtGitDir(options.GitDir)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	sources, memories, projection, err := deriveMemoryIndex(options)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	index, err := buildMemoryIndexV0WithMemories(sources, memories, projection)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	encoded, err := encodeMemoryIndexV0(index)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	if err := ensureMemoryIndexDirectory(options.GitDir); err != nil {
		return MemoryIndexV0{}, err
	}
	if err := options.Write(path, encoded); err != nil {
		return MemoryIndexV0{}, fmt.Errorf("write private memory index: %w", err)
	}
	index.projectionBinding = memoryIndexBinding(options.Context)
	return index, nil
}

func defaultMemoryIndexOptions(options memoryIndexRebuildOptions) memoryIndexRebuildOptions {
	if options.Collect == nil {
		options.Collect = collectMemoryIndexSourcesAt
	}
	if options.Project == nil {
		options.Project = ProjectMemories
	}
	if options.Write == nil {
		options.Write = replacePrivateFileAtomic
	}
	return options
}

func deriveMemoryIndex(options memoryIndexRebuildOptions) ([]memoryIndexSource, []StoredMemory, MemoryProjection, error) {
	if !validMemoryGitOID(options.Context.AtCommit) {
		return nil, nil, MemoryProjection{}, fmt.Errorf("memory index requires an exact projection commit")
	}
	if !validMemoryDigestID(options.Context.PolicyDigest) {
		return nil, nil, MemoryProjection{}, fmt.Errorf("memory index requires an exact policy digest")
	}
	sources, err := options.Collect(options.GitDir)
	if err != nil {
		return nil, nil, MemoryProjection{}, err
	}
	if _, err := memoryIndexSourceFingerprint(sources, options.Context.PolicyDigest); err != nil {
		return nil, nil, MemoryProjection{}, err
	}
	byID := make(map[string]StoredMemory)
	for _, source := range sources {
		for _, stored := range source.Memory {
			if !validMemoryID(stored.ID) || stored.Envelope.Timestamp == "" {
				return nil, nil, MemoryProjection{}, fmt.Errorf("verified memory source contains an invalid record")
			}
			if previous, exists := byID[stored.ID]; exists {
				if !bytes.Equal(previous.Payload, stored.Payload) || previous.Envelope.Timestamp != stored.Envelope.Timestamp {
					return nil, nil, MemoryProjection{}, fmt.Errorf("verified memory sources disagree for one memory ID")
				}
				continue
			}
			byID[stored.ID] = stored
		}
	}
	memories := make([]StoredMemory, 0, len(byID))
	for _, stored := range byID {
		memories = append(memories, stored)
	}
	sort.Slice(memories, func(i, j int) bool { return memories[i].ID < memories[j].ID })
	projection := options.Project(memories, options.Context)
	if projection.PolicyDigest != options.Context.PolicyDigest {
		return nil, nil, MemoryProjection{}, fmt.Errorf("memory projection used a different policy digest")
	}
	return sources, memories, projection, nil
}

// buildMemoryIndexV0 is the deterministic core used by fixtures. Production
// rebuilding additionally supplies verified memories for timestamp binding.
func buildMemoryIndexV0(sources []memoryIndexSource, projection MemoryProjection) (MemoryIndexV0, error) {
	memories := make([]StoredMemory, 0)
	for _, source := range sources {
		memories = append(memories, source.Memory...)
	}
	return buildMemoryIndexV0WithMemories(sources, memories, projection)
}

func buildMemoryIndexV0WithMemories(sources []memoryIndexSource, memories []StoredMemory, projection MemoryProjection) (MemoryIndexV0, error) {
	fingerprint, err := memoryIndexSourceFingerprint(sources, projection.PolicyDigest)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	timestamps := make(map[string]string, len(memories))
	for _, stored := range memories {
		if prior, exists := timestamps[stored.ID]; exists && prior != stored.Envelope.Timestamp {
			return MemoryIndexV0{}, fmt.Errorf("memory index source timestamps disagree")
		}
		timestamps[stored.ID] = stored.Envelope.Timestamp
	}
	dependencies := make(map[string][]MemoryDependency)
	for _, dependency := range projection.MissingDependencies {
		dependencies[dependency.OwnerID] = append(dependencies[dependency.OwnerID], dependency)
	}
	index := MemoryIndexV0{Version: memoryIndexVersion, SourceFingerprint: fingerprint, Records: []MemoryIndexRecordV0{}, Tokens: []MemoryTokenPostingV0{}}
	for _, projected := range projection.Rows {
		timestamp, exists := timestamps[projected.ID]
		if !exists || timestamp == "" {
			return MemoryIndexV0{}, fmt.Errorf("memory projection row has no verified signed timestamp")
		}
		row := memoryIndexRecordFromProjection(projected, timestamp)
		row.Dependencies = append([]MemoryDependency(nil), dependencies[projected.ID]...)
		normalizeMemoryIndexRecord(&row)
		if err := validateMemoryIndexRecord(row); err != nil {
			return MemoryIndexV0{}, err
		}
		index.Records = append(index.Records, row)
	}
	sort.Slice(index.Records, func(i, j int) bool { return index.Records[i].ID < index.Records[j].ID })
	for i := 1; i < len(index.Records); i++ {
		if index.Records[i-1].ID == index.Records[i].ID {
			return MemoryIndexV0{}, fmt.Errorf("memory index contains duplicate record")
		}
	}
	postings := make(map[string][]string)
	for _, row := range index.Records {
		text := row.Data.Content + "\n" + strings.Join(row.Data.Topics, "\n")
		for _, token := range memoryIndexTokens(text) {
			postings[token] = append(postings[token], row.ID)
		}
	}
	tokens := make([]string, 0, len(postings))
	for token := range postings {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)
	for _, token := range tokens {
		ids := sortedUniqueMemoryIndexStrings(postings[token])
		index.Tokens = append(index.Tokens, MemoryTokenPostingV0{Token: token, MemoryIDs: ids})
	}
	return index, nil
}

func memoryIndexRecordFromProjection(row MemoryProjectionRow, timestamp string) MemoryIndexRecordV0 {
	return MemoryIndexRecordV0{
		ID: row.ID, Stream: row.Stream, Actor: row.Actor, SignedTimestamp: timestamp,
		Kind: row.Kind, ContentDigest: row.ContentDigest, Anchor: cloneMemoryAnchor(row.Anchor),
		Signature: row.Signature, Lifecycle: row.Lifecycle,
		Challengers: append([]string(nil), row.Challengers...), Successors: append([]string(nil), row.Successors...),
		Retractions: append([]string(nil), row.Retractions...), Applicability: row.Applicability,
		Evidence: row.Evidence, EvidenceDetails: append([]MemoryEvidenceDetail(nil), row.EvidenceDetails...),
		Dependencies: []MemoryDependency{},
		Trust:        row.Trust, Data: cloneMemoryRecord(row.Data),
	}
}

func cloneMemoryAnchor(anchor MemoryAnchor) MemoryAnchor {
	anchor.Paths = append([]PathAnchor(nil), anchor.Paths...)
	return anchor
}

func cloneMemoryRecord(record MemoryRecord) MemoryRecord {
	record.Anchor = cloneMemoryAnchor(record.Anchor)
	record.Topics = append([]string(nil), record.Topics...)
	record.Evidence = append([]string(nil), record.Evidence...)
	if record.Handoff != nil {
		handoff := *record.Handoff
		handoff.Completed = cloneMemoryIndexStrings(handoff.Completed)
		handoff.Assumptions = cloneMemoryIndexStrings(handoff.Assumptions)
		handoff.Blockers = cloneMemoryIndexStrings(handoff.Blockers)
		handoff.NextActions = cloneMemoryIndexStrings(handoff.NextActions)
		record.Handoff = &handoff
	}
	return record
}

func cloneMemoryIndexStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string{}, values...)
}

func normalizeMemoryIndexRecord(row *MemoryIndexRecordV0) {
	row.Challengers = sortedUniqueMemoryIndexStrings(row.Challengers)
	row.Successors = sortedUniqueMemoryIndexStrings(row.Successors)
	row.Retractions = sortedUniqueMemoryIndexStrings(row.Retractions)
	sort.Slice(row.EvidenceDetails, func(i, j int) bool {
		left, right := row.EvidenceDetails[i], row.EvidenceDetails[j]
		if left.Type != right.Type {
			return left.Type < right.Type
		}
		if left.Requested != right.Requested {
			return left.Requested < right.Requested
		}
		return left.OwnerID < right.OwnerID
	})
	sort.Slice(row.Dependencies, func(i, j int) bool {
		return compareMemoryIndexDependencies(row.Dependencies[i], row.Dependencies[j]) < 0
	})
	row.Data.Topics = sortedUniqueMemoryIndexStrings(row.Data.Topics)
	row.Data.Evidence = sortedUniqueMemoryIndexStrings(row.Data.Evidence)
	sort.Slice(row.Anchor.Paths, func(i, j int) bool { return row.Anchor.Paths[i].Path < row.Anchor.Paths[j].Path })
	sort.Slice(row.Data.Anchor.Paths, func(i, j int) bool { return row.Data.Anchor.Paths[i].Path < row.Data.Anchor.Paths[j].Path })
}

func sortedUniqueMemoryIndexStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	result := append([]string(nil), values...)
	sort.Strings(result)
	write := 1
	for read := 1; read < len(result); read++ {
		if result[read] != result[write-1] {
			result[write] = result[read]
			write++
		}
	}
	return result[:write]
}

func encodeMemoryIndexV0(index MemoryIndexV0) ([]byte, error) {
	if err := validateMemoryIndexV0(index); err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode memory index: %w", err)
	}
	if len(encoded)+1 > maxMemoryIndexBytes {
		return nil, fmt.Errorf("memory index exceeds private cache bound")
	}
	return append(encoded, '\n'), nil
}

func loadMemoryIndexV0At(gitDir string) (MemoryIndexV0, error) {
	path, err := memoryIndexPathAtGitDir(gitDir)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	if err := validateMemoryIndexDirectory(gitDir); errors.Is(err, os.ErrNotExist) {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexMissing}
	} else if err != nil {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexCorrupt, Cause: fmt.Errorf("private cache directory is unsafe")}
	}
	contents, err := readPrivateFileBounded(path, maxMemoryIndexBytes)
	if errors.Is(err, os.ErrNotExist) {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexMissing}
	}
	if err != nil {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexCorrupt, Cause: fmt.Errorf("private cache cannot be read")}
	}
	if !utf8.Valid(contents) {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexCorrupt, Cause: fmt.Errorf("invalid UTF-8")}
	}
	var index MemoryIndexV0
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&index); err != nil {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexCorrupt, Cause: fmt.Errorf("invalid JSON")}
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexCorrupt, Cause: fmt.Errorf("trailing JSON")}
	}
	if index.Version != memoryIndexVersion {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexIncompatible, Cause: fmt.Errorf("unsupported version")}
	}
	if err := validateMemoryIndexV0(index); err != nil {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexCorrupt, Cause: err}
	}
	canonical, err := encodeMemoryIndexV0(index)
	if err != nil || !bytes.Equal(contents, canonical) {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexCorrupt, Cause: fmt.Errorf("noncanonical encoding")}
	}
	return index, nil
}

func verifyMemoryIndex(context MemoryProjectionContext) (MemoryIndexV0, error) {
	gitDir, err := requireGitRepository()
	if err != nil {
		return MemoryIndexV0{}, err
	}
	return verifyMemoryIndexV0(memoryIndexRebuildOptions{GitDir: gitDir, Context: context})
}

func verifyMemoryIndexV0(options memoryIndexRebuildOptions) (MemoryIndexV0, error) {
	options = defaultMemoryIndexOptions(options)
	persisted, err := loadMemoryIndexV0At(options.GitDir)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	sources, memories, projection, err := deriveMemoryIndex(options)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	liveFingerprint, err := memoryIndexSourceFingerprint(sources, options.Context.PolicyDigest)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	if persisted.SourceFingerprint != liveFingerprint {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexStale, Cause: fmt.Errorf("source or policy changed")}
	}
	expected, err := buildMemoryIndexV0WithMemories(sources, memories, projection)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	persistedBytes, err := encodeMemoryIndexV0(persisted)
	if err != nil {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexCorrupt, Cause: err}
	}
	expectedBytes, err := encodeMemoryIndexV0(expected)
	if err != nil {
		return MemoryIndexV0{}, err
	}
	if !bytes.Equal(persistedBytes, expectedBytes) {
		return MemoryIndexV0{}, &MemoryIndexError{Kind: memoryIndexStale, Cause: fmt.Errorf("derived projection changed")}
	}
	persisted.projectionBinding = memoryIndexBinding(options.Context)
	return persisted, nil
}

func memoryIndexBinding(context MemoryProjectionContext) memoryIndexProjectionBinding {
	return memoryIndexProjectionBinding{AtCommit: context.AtCommit, Subject: context.Subject, Path: context.Path}
}

func validateMemoryIndexV0(index MemoryIndexV0) error {
	if index.Version != memoryIndexVersion {
		return fmt.Errorf("unsupported memory index version")
	}
	if !validMemoryDigestID(index.SourceFingerprint) || index.Records == nil || index.Tokens == nil {
		return fmt.Errorf("invalid memory index header")
	}
	for i, row := range index.Records {
		if i > 0 && index.Records[i-1].ID >= row.ID {
			return fmt.Errorf("memory index records are not sorted and unique")
		}
		if err := validateMemoryIndexRecord(row); err != nil {
			return err
		}
	}
	postings := make(map[string][]string)
	for _, row := range index.Records {
		for _, token := range memoryIndexTokens(row.Data.Content + "\n" + strings.Join(row.Data.Topics, "\n")) {
			postings[token] = append(postings[token], row.ID)
		}
	}
	if len(index.Tokens) != len(postings) {
		return fmt.Errorf("memory index token membership is inconsistent")
	}
	for i, posting := range index.Tokens {
		normalized := memoryIndexTokens(posting.Token)
		if len(normalized) != 1 || normalized[0] != posting.Token || (i > 0 && index.Tokens[i-1].Token >= posting.Token) {
			return fmt.Errorf("memory index tokens are not normalized and sorted")
		}
		want := sortedUniqueMemoryIndexStrings(postings[posting.Token])
		if !equalMemoryIndexStrings(posting.MemoryIDs, want) {
			return fmt.Errorf("memory index token posting is inconsistent")
		}
	}
	return nil
}

func validateMemoryIndexRecord(row MemoryIndexRecordV0) error {
	if _, err := time.Parse(time.RFC3339Nano, row.SignedTimestamp); err != nil {
		return fmt.Errorf("memory index contains invalid signed timestamp")
	}
	if !validMemoryID(row.ID) || !validMemoryStreamID(row.Stream) || !validActorFingerprint(row.Actor) ||
		!validMemoryKind(row.Kind) || !validMemoryDigestID(row.ContentDigest) || row.Signature != "valid" {
		return fmt.Errorf("memory index contains invalid record identity")
	}
	if row.Kind != row.Data.Kind || row.ContentDigest != memoryID([]byte(row.Data.Content)) || !reflectMemoryAnchorEqual(row.Anchor, row.Data.Anchor) {
		return fmt.Errorf("memory index record projection disagrees with inert data")
	}
	if err := validateMemoryRecord(row.Data); err != nil {
		return fmt.Errorf("memory index contains invalid record data: %w", err)
	}
	if !isOneOf(row.Lifecycle, memoryLifecycleActive, memoryLifecycleSuperseded, memoryLifecycleRetracted, memoryLifecycleBranching, memoryLifecycleDependencyMissing) ||
		!isOneOf(row.Applicability, memoryApplicabilityApplicable, memoryApplicabilityInapplicable, memoryApplicabilityAnchorMissing, memoryApplicabilityAnchorInvalid) ||
		!isOneOf(row.Evidence, memoryEvidenceResolved, memoryEvidenceMissing, memoryEvidenceInvalid) ||
		!isOneOf(row.Trust, memoryTrustQualified, memoryTrustActorUntrusted, memoryTrustKindUntrusted, memoryTrustPolicyMissing) {
		return fmt.Errorf("memory index contains invalid projection classification")
	}
	for _, ids := range [][]string{row.Challengers, row.Successors, row.Retractions} {
		if !sortedUniqueValidMemoryIDs(ids) {
			return fmt.Errorf("memory index lifecycle edges are invalid")
		}
	}
	if !sortedUniqueStringsExact(row.Data.Topics) || !sortedUniqueStringsExact(row.Data.Evidence) || !sortedUniquePaths(row.Anchor.Paths) || !sortedUniquePaths(row.Data.Anchor.Paths) {
		return fmt.Errorf("memory index nested collections are not sorted and unique")
	}
	for i, detail := range row.EvidenceDetails {
		if detail.OwnerID != row.ID || !isOneOf(detail.Type, "git", "event", "memory") || !isOneOf(detail.Status, memoryEvidenceResolved, memoryEvidenceMissing, memoryEvidenceInvalid) || detail.Requested == "" || detail.Reason == "" {
			return fmt.Errorf("memory index evidence detail is invalid")
		}
		if i > 0 && compareMemoryEvidenceDetails(row.EvidenceDetails[i-1], detail) >= 0 {
			return fmt.Errorf("memory index evidence details are not sorted and unique")
		}
	}
	for i, dependency := range row.Dependencies {
		if dependency.OwnerID != row.ID || dependency.Stream != row.Stream || dependency.Kind == "" || dependency.MissingID == "" || dependency.Reason == "" {
			return fmt.Errorf("memory index dependency is invalid")
		}
		if i > 0 && compareMemoryIndexDependencies(row.Dependencies[i-1], dependency) >= 0 {
			return fmt.Errorf("memory index dependencies are not sorted and unique")
		}
	}
	return nil
}

func reflectMemoryAnchorEqual(left, right MemoryAnchor) bool {
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	return bytes.Equal(leftBytes, rightBytes)
}

func sortedUniqueValidMemoryIDs(values []string) bool {
	for i, value := range values {
		if !validMemoryID(value) || (i > 0 && values[i-1] >= value) {
			return false
		}
	}
	return values != nil
}

func sortedUniqueStringsExact(values []string) bool {
	if values == nil {
		return false
	}
	for i := 1; i < len(values); i++ {
		if values[i-1] >= values[i] {
			return false
		}
	}
	return true
}

func sortedUniquePaths(values []PathAnchor) bool {
	for i := 1; i < len(values); i++ {
		if values[i-1].Path >= values[i].Path {
			return false
		}
	}
	return true
}

func compareMemoryEvidenceDetails(left, right MemoryEvidenceDetail) int {
	if left.Type != right.Type {
		return strings.Compare(left.Type, right.Type)
	}
	if left.Requested != right.Requested {
		return strings.Compare(left.Requested, right.Requested)
	}
	return strings.Compare(left.OwnerID, right.OwnerID)
}

func compareMemoryIndexDependencies(left, right MemoryDependency) int {
	for _, pair := range [][2]string{
		{left.MissingID, right.MissingID},
		{left.Kind, right.Kind},
		{left.Operation, right.Operation},
		{left.OwnerID, right.OwnerID},
		{left.Stream, right.Stream},
		{left.Reason, right.Reason},
	} {
		if pair[0] != pair[1] {
			return strings.Compare(pair[0], pair[1])
		}
	}
	return 0
}

func isOneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func equalMemoryIndexStrings(left, right []string) bool {
	if len(left) != len(right) || left == nil != (right == nil) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// memoryIndexTokens lowercases Unicode letters and numbers and splits on every
// other rune. Combining marks are boundaries, making behavior locale-free.
func memoryIndexTokens(value string) []string {
	if !utf8.ValidString(value) {
		return []string{}
	}
	var tokens []string
	var word []rune
	flush := func() {
		if len(word) != 0 {
			tokens = append(tokens, string(word))
			word = word[:0]
		}
	}
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			word = append(word, unicode.ToLower(character))
		} else {
			flush()
		}
	}
	flush()
	return sortedUniqueMemoryIndexStrings(tokens)
}

func queryMemoryIndexV0(index MemoryIndexV0, query MemoryIndexQuery) ([]MemoryIndexRecordV0, error) {
	if err := validateMemoryIndexV0(index); err != nil {
		return nil, err
	}
	if err := validateMemoryIndexQuery(query); err != nil {
		return nil, err
	}
	if index.projectionBinding.AtCommit == "" {
		return nil, fmt.Errorf("verify memory index before query")
	}
	queryBinding := memoryIndexProjectionBinding{AtCommit: query.AtCommit, Subject: query.Subject, Path: query.Path}
	if queryBinding != index.projectionBinding {
		return nil, fmt.Errorf("memory index projection context does not match query")
	}
	exact := make([]MemoryIndexRecordV0, 0)
	exactIDs := make(map[string]bool)
	for _, row := range index.Records {
		if !memoryIndexRecordMatches(row, query) {
			continue
		}
		exact = append(exact, row)
		exactIDs[row.ID] = true
	}
	lexical := memoryIndexTokens(query.Query)
	if query.Query != "" && len(lexical) == 0 {
		return nil, fmt.Errorf("memory index query has no lexical terms")
	}
	lexicalIDs := memoryIndexLexicalIntersectionForCandidates(index.Tokens, lexical, exactIDs)
	result := make([]MemoryIndexRecordV0, 0, len(exact))
	for _, row := range exact {
		if len(lexical) != 0 && !lexicalIDs[row.ID] {
			continue
		}
		result = append(result, cloneMemoryIndexRecord(row))
	}
	sort.Slice(result, func(i, j int) bool { return memoryIndexRecordLess(result[i], result[j]) })
	return result, nil
}

func cloneMemoryIndexRecord(row MemoryIndexRecordV0) MemoryIndexRecordV0 {
	row.Anchor = cloneMemoryAnchor(row.Anchor)
	row.Challengers = append([]string(nil), row.Challengers...)
	row.Successors = append([]string(nil), row.Successors...)
	row.Retractions = append([]string(nil), row.Retractions...)
	row.EvidenceDetails = append([]MemoryEvidenceDetail(nil), row.EvidenceDetails...)
	row.Dependencies = append([]MemoryDependency(nil), row.Dependencies...)
	row.Data = cloneMemoryRecord(row.Data)
	return row
}

func validateMemoryIndexQuery(query MemoryIndexQuery) error {
	if !validMemoryGitOID(query.AtCommit) {
		return fmt.Errorf("invalid atCommit filter")
	}
	if query.Subject != "" && !validMemorySubject(query.Subject) {
		return fmt.Errorf("invalid subject filter")
	}
	if query.Path != "" && !validMemoryPath(query.Path) {
		return fmt.Errorf("invalid path filter")
	}
	for _, topic := range query.Topics {
		if !validNormalizedMemoryTopic(topic) {
			return fmt.Errorf("invalid topic filter")
		}
	}
	for _, kind := range query.Kinds {
		if !validMemoryKind(kind) {
			return fmt.Errorf("invalid kind filter")
		}
	}
	for _, actor := range query.Actors {
		if !validActorFingerprint(actor) {
			return fmt.Errorf("invalid actor filter")
		}
	}
	for _, lifecycle := range query.Lifecycles {
		if !isOneOf(lifecycle, memoryLifecycleActive, memoryLifecycleSuperseded, memoryLifecycleRetracted, memoryLifecycleBranching, memoryLifecycleDependencyMissing) {
			return fmt.Errorf("invalid lifecycle filter")
		}
	}
	for _, applicability := range query.Applicabilities {
		if !isOneOf(applicability, memoryApplicabilityApplicable, memoryApplicabilityInapplicable, memoryApplicabilityAnchorMissing, memoryApplicabilityAnchorInvalid) {
			return fmt.Errorf("invalid applicability filter")
		}
	}
	for _, trust := range query.Trust {
		if !isOneOf(trust, memoryTrustQualified, memoryTrustActorUntrusted, memoryTrustKindUntrusted, memoryTrustPolicyMissing) {
			return fmt.Errorf("invalid trust filter")
		}
	}
	if !utf8.ValidString(query.Query) {
		return fmt.Errorf("invalid UTF-8 lexical query")
	}
	return nil
}

func memoryIndexLexicalIntersectionForCandidates(postings []MemoryTokenPostingV0, tokens []string, candidates map[string]bool) map[string]bool {
	if len(tokens) == 0 {
		return nil
	}
	lookup := make(map[string][]string, len(postings))
	for _, posting := range postings {
		lookup[posting.Token] = posting.MemoryIDs
	}
	counts := make(map[string]int)
	for _, token := range tokens {
		ids, exists := lookup[token]
		if !exists {
			return map[string]bool{}
		}
		for _, id := range ids {
			if candidates != nil && !candidates[id] {
				continue
			}
			counts[id]++
		}
	}
	result := make(map[string]bool)
	for id, count := range counts {
		if count == len(tokens) {
			result[id] = true
		}
	}
	return result
}

func memoryIndexRecordMatches(row MemoryIndexRecordV0, query MemoryIndexQuery) bool {
	if query.Subject != "" && row.Anchor.Subject != query.Subject && row.Data.Applicability.Subject != query.Subject {
		return false
	}
	if query.Path != "" {
		matched := false
		for _, anchored := range row.Anchor.Paths {
			matched = matched || anchored.Path == query.Path
		}
		if !matched {
			return false
		}
	}
	for _, topic := range query.Topics {
		if !sortedMemoryIndexContains(row.Data.Topics, topic) {
			return false
		}
	}
	return memoryIndexCategoryMatches(query.Kinds, row.Kind) &&
		memoryIndexCategoryMatches(query.Actors, row.Actor) &&
		memoryIndexCategoryMatches(query.Lifecycles, row.Lifecycle) &&
		memoryIndexCategoryMatches(query.Applicabilities, row.Applicability) &&
		memoryIndexCategoryMatches(query.Trust, row.Trust)
}

func memoryIndexCategoryMatches(values []string, candidate string) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func sortedMemoryIndexContains(values []string, candidate string) bool {
	position := sort.SearchStrings(values, candidate)
	return position < len(values) && values[position] == candidate
}

func memoryIndexRecordLess(left, right MemoryIndexRecordV0) bool {
	leftApplicability := memoryIndexApplicabilityRank(left.Applicability)
	rightApplicability := memoryIndexApplicabilityRank(right.Applicability)
	if leftApplicability != rightApplicability {
		return leftApplicability < rightApplicability
	}
	leftLifecycle := memoryIndexLifecycleRank(left.Lifecycle)
	rightLifecycle := memoryIndexLifecycleRank(right.Lifecycle)
	if leftLifecycle != rightLifecycle {
		return leftLifecycle < rightLifecycle
	}
	if left.SignedTimestamp != right.SignedTimestamp {
		return left.SignedTimestamp > right.SignedTimestamp
	}
	return left.ID < right.ID
}

func memoryIndexApplicabilityRank(value string) int {
	switch value {
	case memoryApplicabilityApplicable:
		return 0
	case memoryApplicabilityInapplicable:
		return 1
	case memoryApplicabilityAnchorMissing:
		return 2
	default:
		return 3
	}
}

func memoryIndexLifecycleRank(value string) int {
	switch value {
	case memoryLifecycleActive:
		return 0
	case memoryLifecycleSuperseded:
		return 1
	case memoryLifecycleRetracted:
		return 2
	case memoryLifecycleBranching:
		return 3
	default:
		return 4
	}
}
