package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func testIdentity(t *testing.T, name string) *Identity {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return &Identity{
		Actor:      actorForPublicKey(publicKey),
		Name:       name,
		PublicKey:  base64.RawStdEncoding.EncodeToString(publicKey),
		PrivateKey: base64.RawStdEncoding.EncodeToString(privateKey),
	}
}

func TestSignedEventRoundTrip(t *testing.T) {
	identity := testIdentity(t, "Alice")
	event := newEvent(identity, "issue.open", 1, "")
	event.Title = "Distributed issues"

	payload, signature, err := encodeAndSign(event, identity)
	if err != nil {
		t.Fatal(err)
	}
	got, id, err := verifyEvent(payload, signature)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != event.Title {
		t.Fatalf("title = %q, want %q", got.Title, event.Title)
	}
	if len(id) != len("sha256:")+64 {
		t.Fatalf("unexpected event ID %q", id)
	}
}

func TestModifiedEventFailsVerification(t *testing.T) {
	identity := testIdentity(t, "Alice")
	event := newEvent(identity, "issue.open", 1, "")
	event.Title = "Original"
	payload, signature, err := encodeAndSign(event, identity)
	if err != nil {
		t.Fatal(err)
	}
	payload = bytes.Replace(payload, []byte("Original"), []byte("Modified"), 1)
	if _, _, err := verifyEvent(payload, signature); err == nil {
		t.Fatal("modified event unexpectedly verified")
	}
}
