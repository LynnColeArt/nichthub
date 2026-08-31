package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	identityRecordVersion   = 1
	identityRotationVersion = 1
	maxIdentityStateBytes   = 64 * 1024

	identityLifecycleAvailable = "available"

	identityRelationshipDevice    = "device"
	identityRelationshipSuccessor = "successor"
)

var (
	errNoIdentity       = errors.New("no identity")
	identityStorageHook func(step string) error
)

type identityKeyringLayout struct {
	root       string
	legacy     string
	identities string
	active     string
	rotation   string
}

type identityRecord struct {
	Version    int    `json:"version"`
	Actor      string `json:"actor"`
	Name       string `json:"name"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
	Lifecycle  string `json:"lifecycle"`
}

type identityRecordMetadata struct {
	Actor     string
	Name      string
	PublicKey string
	Lifecycle string
	Active    bool
}

type identityRotationState struct {
	Version            int    `json:"version"`
	PredecessorActor   string `json:"predecessorActor"`
	TargetActor        string `json:"targetActor"`
	Relationship       string `json:"relationship"`
	AuthorizationEvent string `json:"authorizationEvent,omitempty"`
	AcceptanceEvent    string `json:"acceptanceEvent,omitempty"`
}

func identityKeyringPaths() (identityKeyringLayout, error) {
	gitDir, err := requireGitRepository()
	if err != nil {
		return identityKeyringLayout{}, err
	}
	root := filepath.Join(gitDir, "nh")
	return identityKeyringLayout{
		root:       root,
		legacy:     filepath.Join(root, "identity.json"),
		identities: filepath.Join(root, "identities"),
		active:     filepath.Join(root, "active"),
		rotation:   filepath.Join(root, "rotation.json"),
	}, nil
}

func filepathForIdentityRecord(paths identityKeyringLayout, actor string) string {
	return filepath.Join(paths.identities, actor+".json")
}

func ensureIdentityKeyringDirectories(paths identityKeyringLayout) error {
	if err := ensurePrivateDirectory(paths.root); err != nil {
		return err
	}
	return ensurePrivateDirectory(paths.identities)
}

func ensurePrivateDirectory(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(path, 0o700); err != nil {
			return fmt.Errorf("create private identity directory: %w", err)
		}
		if err := os.Chmod(path, 0o700); err != nil {
			return fmt.Errorf("protect private identity directory: %w", err)
		}
		return nil
	}
	if err != nil {
		return err
	}
	return validatePrivateDirectory(path, info)
}

func validatePrivateDirectory(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("private identity directory %s is a symbolic link", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("private identity path %s is not a directory", path)
	}
	if info.Mode().Perm() != 0o700 {
		return fmt.Errorf("private identity directory %s has unsafe permissions", path)
	}
	return nil
}

func readPrivateFile(path string) ([]byte, error) {
	return readPrivateFileBounded(path, maxIdentityStateBytes)
}

func readPrivateFileBounded(path string, maximum int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("private identity file %s is a symbolic link", path)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("private identity path %s is not a regular file", path)
	}
	if info.Mode().Perm() != 0o600 {
		return nil, fmt.Errorf("private identity file %s has unsafe permissions", path)
	}
	if info.Size() > maximum {
		return nil, fmt.Errorf("private identity file %s is too large", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	contents, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(contents)) > maximum {
		return nil, fmt.Errorf("private identity file %s is too large", path)
	}
	return contents, nil
}

func decodePrivateJSON(contents []byte, target any, description string) error {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("read %s: %w", description, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("read %s: unexpected trailing JSON", description)
		}
		return fmt.Errorf("read %s: %w", description, err)
	}
	return nil
}

func writePrivateFileAtomic(path string, contents []byte) error {
	return writePrivateFileAtomicWithReplacement(path, contents, false)
}

func replacePrivateFileAtomic(path string, contents []byte) error {
	return writePrivateFileAtomicWithReplacement(path, contents, true)
}

func writePrivateFileAtomicWithReplacement(path string, contents []byte, replaceInvalid bool) error {
	directory := filepath.Dir(path)
	if info, err := os.Lstat(path); err == nil && !replaceInvalid {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("private identity file %s is a symbolic link", path)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("private identity path %s is not a regular file", path)
		}
		if info.Mode().Perm() != 0o600 {
			return fmt.Errorf("private identity file %s has unsafe permissions", path)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	temporary, err := os.CreateTemp(directory, ".nh-private-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	closed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		_ = os.Remove(temporaryPath)
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return err
	}
	if _, err := temporary.Write(contents); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	closed = true
	if err := os.Chmod(temporaryPath, 0o600); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	return syncPrivateDirectory(directory)
}

func syncPrivateDirectory(directory string) error {
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer directoryHandle.Close()
	if err := directoryHandle.Sync(); err != nil {
		return err
	}
	return nil
}

func identityFromRecord(record identityRecord) *Identity {
	return &Identity{
		Actor:      record.Actor,
		Name:       record.Name,
		PublicKey:  record.PublicKey,
		PrivateKey: record.PrivateKey,
	}
}

func recordFromIdentity(identity *Identity, lifecycle string) identityRecord {
	return identityRecord{
		Version:    identityRecordVersion,
		Actor:      identity.Actor,
		Name:       identity.Name,
		PublicKey:  identity.PublicKey,
		PrivateKey: identity.PrivateKey,
		Lifecycle:  lifecycle,
	}
}

func storeIdentityRecord(identity *Identity, lifecycle string) (string, error) {
	if err := validateIdentity(identity); err != nil {
		return "", err
	}
	if lifecycle != identityLifecycleAvailable {
		return "", fmt.Errorf("unsupported local identity lifecycle %q", lifecycle)
	}
	paths, err := identityKeyringPaths()
	if err != nil {
		return "", err
	}
	if err := ensureIdentityKeyringDirectories(paths); err != nil {
		return "", err
	}
	path := filepathForIdentityRecord(paths, identity.Actor)
	want := recordFromIdentity(identity, lifecycle)
	existing, err := loadIdentityRecord(identity.Actor)
	if err == nil {
		got := recordFromIdentity(existing, lifecycle)
		if got != want {
			return "", fmt.Errorf("identity record for actor %s already contains different key material or metadata", identity.Actor)
		}
		return path, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	contents, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		return "", err
	}
	if err := writePrivateFileAtomic(path, append(contents, '\n')); err != nil {
		return "", err
	}
	return path, nil
}

func loadIdentityRecord(actor string) (*Identity, error) {
	if !validActorFingerprint(actor) {
		return nil, fmt.Errorf("invalid identity actor fingerprint")
	}
	paths, err := identityKeyringPaths()
	if err != nil {
		return nil, err
	}
	if err := checkPrivateDirectory(paths.root); err != nil {
		return nil, err
	}
	if err := checkPrivateDirectory(paths.identities); err != nil {
		return nil, err
	}
	contents, err := readPrivateFile(filepathForIdentityRecord(paths, actor))
	if err != nil {
		return nil, err
	}
	var record identityRecord
	if err := decodePrivateJSON(contents, &record, "identity record"); err != nil {
		return nil, err
	}
	if record.Version != identityRecordVersion {
		return nil, fmt.Errorf("unsupported identity record version %d", record.Version)
	}
	if record.Lifecycle != identityLifecycleAvailable {
		return nil, fmt.Errorf("unsupported local identity lifecycle %q", record.Lifecycle)
	}
	if record.Actor != actor {
		return nil, fmt.Errorf("identity record actor does not match its path")
	}
	identity := identityFromRecord(record)
	if err := validateIdentity(identity); err != nil {
		return nil, err
	}
	return identity, nil
}

func listIdentityRecords() ([]identityRecordMetadata, error) {
	paths, err := identityKeyringPaths()
	if err != nil {
		return nil, err
	}
	if err := checkPrivateDirectory(paths.root); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if err := checkPrivateDirectory(paths.identities); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	active, activeErr := loadActiveActor()
	if activeErr != nil && !errors.Is(activeErr, os.ErrNotExist) {
		return nil, activeErr
	}
	entries, err := os.ReadDir(paths.identities)
	if err != nil {
		return nil, err
	}
	metadata := make([]identityRecordMetadata, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		actor := strings.TrimSuffix(entry.Name(), ".json")
		identity, err := loadIdentityRecord(actor)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, identityRecordMetadata{
			Actor:     identity.Actor,
			Name:      identity.Name,
			PublicKey: identity.PublicKey,
			Lifecycle: identityLifecycleAvailable,
			Active:    activeErr == nil && active == identity.Actor,
		})
	}
	sort.Slice(metadata, func(i, j int) bool { return metadata[i].Actor < metadata[j].Actor })
	return metadata, nil
}

func loadActiveActor() (string, error) {
	paths, err := identityKeyringPaths()
	if err != nil {
		return "", err
	}
	if err := checkPrivateDirectory(paths.root); err != nil {
		return "", err
	}
	contents, err := readPrivateFile(paths.active)
	if err != nil {
		return "", err
	}
	actor := strings.TrimSpace(string(contents))
	if !validActorFingerprint(actor) || string(contents) != actor+"\n" {
		return "", fmt.Errorf("active identity contains an invalid actor fingerprint")
	}
	return actor, nil
}

func initializeActiveIdentity(actor string) error {
	if _, err := loadIdentityRecord(actor); err != nil {
		return fmt.Errorf("cannot activate identity: %w", err)
	}
	paths, err := identityKeyringPaths()
	if err != nil {
		return err
	}
	if err := ensureIdentityKeyringDirectories(paths); err != nil {
		return err
	}
	current, err := loadActiveActor()
	if err == nil {
		if current == actor {
			return nil
		}
		return fmt.Errorf("active identity already exists for actor %s", current)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return writePrivateFileAtomic(paths.active, []byte(actor+"\n"))
}

func replaceActiveIdentity(actor, expected string) error {
	if _, err := loadIdentityRecord(actor); err != nil {
		return fmt.Errorf("cannot activate identity: %w", err)
	}
	paths, err := identityKeyringPaths()
	if err != nil {
		return err
	}
	current, err := loadActiveActor()
	if err != nil {
		return err
	}
	if current == actor {
		return nil
	}
	if current != expected {
		return fmt.Errorf("active identity changed from expected actor %s", expected)
	}
	return writePrivateFileAtomic(paths.active, []byte(actor+"\n"))
}

func loadLegacyIdentity() (*Identity, error) {
	paths, err := identityKeyringPaths()
	if err != nil {
		return nil, err
	}
	if err := checkPrivateDirectory(paths.root); err != nil {
		return nil, err
	}
	contents, err := readPrivateFile(paths.legacy)
	if err != nil {
		return nil, err
	}
	var identity Identity
	if err := decodePrivateJSON(contents, &identity, "legacy identity"); err != nil {
		return nil, err
	}
	if err := validateIdentity(&identity); err != nil {
		return nil, err
	}
	return &identity, nil
}

func loadActiveIdentity() (*Identity, error) {
	actor, err := loadActiveActor()
	if err == nil {
		return loadIdentityRecord(actor)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	legacy, err := loadLegacyIdentity()
	if errors.Is(err, os.ErrNotExist) {
		return nil, errNoIdentity
	}
	if err != nil {
		return nil, err
	}
	if _, err := storeIdentityRecord(legacy, identityLifecycleAvailable); err != nil {
		return nil, err
	}
	if err := runIdentityStorageHook("migration-record-durable"); err != nil {
		return nil, err
	}
	if err := initializeActiveIdentity(legacy.Actor); err != nil {
		return nil, err
	}
	return loadIdentityRecord(legacy.Actor)
}

func validateRotationState(state identityRotationState, requireComplete bool) error {
	if state.Version != identityRotationVersion {
		return fmt.Errorf("unsupported identity rotation version %d", state.Version)
	}
	if !validActorFingerprint(state.PredecessorActor) || !validActorFingerprint(state.TargetActor) {
		return fmt.Errorf("identity rotation has an invalid actor fingerprint")
	}
	if state.PredecessorActor == state.TargetActor {
		return fmt.Errorf("identity rotation target must differ from its predecessor")
	}
	if state.Relationship != identityRelationshipDevice && state.Relationship != identityRelationshipSuccessor {
		return fmt.Errorf("unsupported identity relationship %q", state.Relationship)
	}
	if state.AuthorizationEvent != "" && !validEventID(state.AuthorizationEvent) {
		return fmt.Errorf("identity rotation has an invalid authorization event ID")
	}
	if state.AcceptanceEvent != "" && !validEventID(state.AcceptanceEvent) {
		return fmt.Errorf("identity rotation has an invalid acceptance event ID")
	}
	if state.AcceptanceEvent != "" && state.AuthorizationEvent == "" {
		return fmt.Errorf("identity rotation acceptance requires an authorization")
	}
	if requireComplete && (state.AuthorizationEvent == "" || state.AcceptanceEvent == "") {
		return fmt.Errorf("identity rotation is incomplete")
	}
	return nil
}

func storeIdentityRotation(state identityRotationState) error {
	if err := validateRotationState(state, false); err != nil {
		return err
	}
	if _, err := loadIdentityRecord(state.PredecessorActor); err != nil {
		return fmt.Errorf("read rotation predecessor: %w", err)
	}
	if _, err := loadIdentityRecord(state.TargetActor); err != nil {
		return fmt.Errorf("read rotation target: %w", err)
	}
	paths, err := identityKeyringPaths()
	if err != nil {
		return err
	}
	if err := ensureIdentityKeyringDirectories(paths); err != nil {
		return err
	}
	existing, err := loadIdentityRotation()
	if err == nil {
		if existing.PredecessorActor != state.PredecessorActor || existing.TargetActor != state.TargetActor || existing.Relationship != state.Relationship {
			return fmt.Errorf("another identity rotation is already in progress")
		}
		if existing.AuthorizationEvent != "" {
			if state.AuthorizationEvent != "" && state.AuthorizationEvent != existing.AuthorizationEvent {
				return fmt.Errorf("identity rotation authorization cannot be replaced")
			}
			state.AuthorizationEvent = existing.AuthorizationEvent
		}
		if existing.AcceptanceEvent != "" {
			if state.AcceptanceEvent != "" && state.AcceptanceEvent != existing.AcceptanceEvent {
				return fmt.Errorf("identity rotation acceptance cannot be replaced")
			}
			state.AcceptanceEvent = existing.AcceptanceEvent
		}
		if state == existing {
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	contents, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return writePrivateFileAtomic(paths.rotation, append(contents, '\n'))
}

func loadIdentityRotation() (identityRotationState, error) {
	paths, err := identityKeyringPaths()
	if err != nil {
		return identityRotationState{}, err
	}
	if err := checkPrivateDirectory(paths.root); err != nil {
		return identityRotationState{}, err
	}
	contents, err := readPrivateFile(paths.rotation)
	if err != nil {
		return identityRotationState{}, err
	}
	var state identityRotationState
	if err := decodePrivateJSON(contents, &state, "identity rotation state"); err != nil {
		return identityRotationState{}, err
	}
	if err := validateRotationState(state, false); err != nil {
		return identityRotationState{}, err
	}
	return state, nil
}

func switchActiveIdentity(state identityRotationState) error {
	if err := validateRotationState(state, true); err != nil {
		return err
	}
	if _, err := loadIdentityRecord(state.TargetActor); err != nil {
		return fmt.Errorf("read rotation target: %w", err)
	}
	active, err := loadActiveActor()
	if err != nil {
		return err
	}
	persisted, err := loadIdentityRotation()
	if errors.Is(err, os.ErrNotExist) && active == state.TargetActor {
		return nil
	}
	if err != nil {
		return err
	}
	if persisted != state {
		return fmt.Errorf("completed identity rotation does not match durable transaction state")
	}
	if active != state.TargetActor {
		if active != state.PredecessorActor {
			return fmt.Errorf("active identity is %s, not rotation predecessor %s", active, state.PredecessorActor)
		}
		if err := runIdentityStorageHook("before-active-switch"); err != nil {
			return err
		}
		if err := replaceActiveIdentity(state.TargetActor, state.PredecessorActor); err != nil {
			return err
		}
	}
	return removeIdentityRotation(state)
}

func removeIdentityRotation(expected identityRotationState) error {
	persisted, err := loadIdentityRotation()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if persisted != expected {
		return fmt.Errorf("completed identity rotation does not match durable transaction state")
	}
	if err := runIdentityStorageHook("before-rotation-cleanup"); err != nil {
		return err
	}
	paths, err := identityKeyringPaths()
	if err != nil {
		return err
	}
	if err := os.Remove(paths.rotation); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := syncPrivateDirectory(paths.root); err != nil {
		return err
	}
	return runIdentityStorageHook("after-rotation-cleanup")
}

func runIdentityStorageHook(step string) error {
	if identityStorageHook != nil {
		return identityStorageHook(step)
	}
	if os.Getenv("NH_INTERNAL_TESTING") == "1" && os.Getenv("NH_TEST_ROTATION_INTERRUPT_AFTER") == step {
		return fmt.Errorf("injected rotation interruption after %s", step)
	}
	return nil
}

func checkPrivateDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	return validatePrivateDirectory(path, info)
}

func validActorFingerprint(actor string) bool {
	if len(actor) != 64 || strings.ToLower(actor) != actor {
		return false
	}
	for _, character := range actor {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
