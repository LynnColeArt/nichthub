package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Identity struct {
	Actor      string `json:"actor"`
	Name       string `json:"name"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

func actorForPublicKey(publicKey ed25519.PublicKey) string {
	sum := sha256.Sum256(publicKey)
	return hex.EncodeToString(sum[:])
}

func identityPath() (string, error) {
	gitDir, err := requireGitRepository()
	if err != nil {
		return "", err
	}
	return filepath.Join(gitDir, "nh", "identity.json"), nil
}

func createIdentity(name string) (*Identity, string, error) {
	path, err := identityPath()
	if err != nil {
		return nil, "", err
	}
	if _, err := os.Stat(path); err == nil {
		return nil, "", fmt.Errorf("identity already exists at %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, "", err
	}

	if strings.TrimSpace(name) == "" {
		name, _ = gitText("config", "--get", "user.name")
	}
	if strings.TrimSpace(name) == "" {
		name = "anonymous"
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("generate identity: %w", err)
	}
	identity := &Identity{
		Actor:      actorForPublicKey(publicKey),
		Name:       name,
		PublicKey:  base64.RawStdEncoding.EncodeToString(publicKey),
		PrivateKey: base64.RawStdEncoding.EncodeToString(privateKey),
	}
	encoded, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return nil, "", err
	}
	encoded = append(encoded, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, "", err
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return nil, "", err
	}
	return identity, path, nil
}

func loadIdentity() (*Identity, error) {
	path, err := identityPath()
	if err != nil {
		return nil, err
	}
	encoded, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no identity; run 'nh init'")
	}
	if err != nil {
		return nil, err
	}
	var identity Identity
	if err := json.Unmarshal(encoded, &identity); err != nil {
		return nil, fmt.Errorf("read identity: %w", err)
	}
	publicKey, err := base64.RawStdEncoding.DecodeString(identity.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("identity has an invalid public key")
	}
	privateKey, err := base64.RawStdEncoding.DecodeString(identity.PrivateKey)
	if err != nil || len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("identity has an invalid private key")
	}
	if actorForPublicKey(publicKey) != identity.Actor {
		return nil, fmt.Errorf("identity actor does not match its public key")
	}
	if !ed25519.PublicKey(privateKey[32:]).Equal(ed25519.PublicKey(publicKey)) {
		return nil, fmt.Errorf("identity key pair does not match")
	}
	return &identity, nil
}

func (i *Identity) publicKey() (ed25519.PublicKey, error) {
	key, err := base64.RawStdEncoding.DecodeString(i.PublicKey)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key")
	}
	return ed25519.PublicKey(key), nil
}

func (i *Identity) privateKey() (ed25519.PrivateKey, error) {
	key, err := base64.RawStdEncoding.DecodeString(i.PrivateKey)
	if err != nil || len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key")
	}
	return ed25519.PrivateKey(key), nil
}
