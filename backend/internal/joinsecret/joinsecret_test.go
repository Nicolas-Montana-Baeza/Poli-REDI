package joinsecret

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestAESGCMRoundTripAndAAD(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
	if err := Configure("1:"+key, 1); err != nil {
		t.Fatal(err)
	}
	nonce, encrypted, version, err := Encrypt("secret-code", 42)
	if err != nil {
		t.Fatal(err)
	}
	if string(encrypted) == "secret-code" {
		t.Fatal("plaintext leaked")
	}
	got, err := Decrypt(nonce, encrypted, version, 42)
	if err != nil || got != "secret-code" {
		t.Fatalf("round trip = %q, %v", got, err)
	}
	if _, err := Decrypt(nonce, encrypted, version, 43); err == nil {
		t.Fatal("ciphertext must be bound to reservation id")
	}
}

func TestConfigureRejectsInvalidKey(t *testing.T) {
	if err := Configure("not-base64", 1); err == nil {
		t.Fatal("invalid key accepted")
	}
}

func TestKeyringDecryptsPreviousVersion(t *testing.T) {
	oldKey := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("o", 32)))
	newKey := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("n", 32)))
	if err := Configure("1:"+oldKey, 1); err != nil {
		t.Fatal(err)
	}
	nonce, encrypted, version, err := Encrypt("legacy", 8)
	if err != nil {
		t.Fatal(err)
	}
	if err = Configure("1:"+oldKey+",2:"+newKey, 2); err != nil {
		t.Fatal(err)
	}
	if got, err := Decrypt(nonce, encrypted, version, 8); err != nil || got != "legacy" {
		t.Fatalf("old version unavailable: %q %v", got, err)
	}
}
