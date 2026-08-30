package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
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
	paths, err := identityKeyringPaths()
	if err != nil {
		return "", err
	}
	return paths.legacy, nil
}

func createIdentity(name string) (*Identity, string, error) {
	existing, err := loadActiveIdentity()
	if err == nil {
		paths, pathErr := identityKeyringPaths()
		if pathErr != nil {
			return nil, "", pathErr
		}
		return nil, "", fmt.Errorf("identity already exists at %s", filepathForIdentityRecord(paths, existing.Actor))
	}
	if !errors.Is(err, errNoIdentity) {
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
	path, err := storeIdentityRecord(identity, identityLifecycleAvailable)
	if err != nil {
		return nil, "", err
	}
	if err := initializeActiveIdentity(identity.Actor); err != nil {
		return nil, "", err
	}
	return identity, path, nil
}

func loadIdentity() (*Identity, error) {
	identity, err := loadActiveIdentity()
	if errors.Is(err, errNoIdentity) {
		return nil, fmt.Errorf("no identity; run 'nh init'")
	}
	return identity, err
}

func validateIdentity(identity *Identity) error {
	if identity == nil {
		return fmt.Errorf("identity is missing")
	}
	publicKey, err := base64.RawStdEncoding.DecodeString(identity.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("identity has an invalid public key")
	}
	privateKey, err := base64.RawStdEncoding.DecodeString(identity.PrivateKey)
	if err != nil || len(privateKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("identity has an invalid private key")
	}
	if actorForPublicKey(publicKey) != identity.Actor {
		return fmt.Errorf("identity actor does not match its public key")
	}
	if !ed25519.PublicKey(privateKey[32:]).Equal(ed25519.PublicKey(publicKey)) {
		return fmt.Errorf("identity key pair does not match")
	}
	return nil
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
