package crypto

import (
	"bytes"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	salt := []byte("fixed-test-salt!")
	return DeriveKey("test-passphrase", salt, DefaultParams())
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := testKey(t)

	cases := []string{
		"hello world",
		"",
		"a",
		"secret-api-key-12345",
		"value with spaces and special chars: !@#$%^&*()",
		"multi\nline\nvalue",
	}

	for _, plaintext := range cases {
		encrypted, err := EncryptValue(key, plaintext)
		if err != nil {
			t.Fatalf("EncryptValue(%q): %v", plaintext, err)
		}
		if !IsEncrypted(encrypted) {
			t.Fatalf("encrypted value should have ENC: prefix, got %q", encrypted)
		}

		decrypted, err := DecryptValue(key, encrypted)
		if err != nil {
			t.Fatalf("DecryptValue(%q): %v", encrypted, err)
		}
		if decrypted != plaintext {
			t.Fatalf("round-trip failed: got %q, want %q", decrypted, plaintext)
		}
	}
}

func TestEncryptValueUniqueNonces(t *testing.T) {
	key := testKey(t)

	enc1, err := EncryptValue(key, "same-value")
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := EncryptValue(key, "same-value")
	if err != nil {
		t.Fatal(err)
	}
	if enc1 == enc2 {
		t.Fatal("encrypting the same value twice should produce different ciphertexts")
	}
}

func TestDecryptValueWrongKey(t *testing.T) {
	key := testKey(t)
	encrypted, err := EncryptValue(key, "secret")
	if err != nil {
		t.Fatal(err)
	}

	wrongKey := DeriveKey("wrong-passphrase", []byte("fixed-test-salt!"), DefaultParams())
	_, err = DecryptValue(wrongKey, encrypted)
	if err != ErrDecryptionFailed {
		t.Fatalf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestDecryptValueNotEncrypted(t *testing.T) {
	key := testKey(t)
	_, err := DecryptValue(key, "plaintext-value")
	if err != ErrNotEncrypted {
		t.Fatalf("expected ErrNotEncrypted, got %v", err)
	}
}

func TestIsEncrypted(t *testing.T) {
	if !IsEncrypted("ENC:abc123") {
		t.Fatal("expected true for ENC: prefixed value")
	}
	if IsEncrypted("plaintext") {
		t.Fatal("expected false for non-prefixed value")
	}
	if IsEncrypted("") {
		t.Fatal("expected false for empty string")
	}
}

func TestEncryptDecryptMap(t *testing.T) {
	key := testKey(t)

	vars := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PASSWORD": "secret123",
		"EMPTY":       "",
	}

	encrypted, err := EncryptMap(key, vars)
	if err != nil {
		t.Fatalf("EncryptMap: %v", err)
	}

	for k, v := range encrypted {
		if !IsEncrypted(v) {
			t.Fatalf("value for %q should be encrypted, got %q", k, v)
		}
	}

	decrypted, err := DecryptMap(key, encrypted)
	if err != nil {
		t.Fatalf("DecryptMap: %v", err)
	}

	for k, want := range vars {
		got, ok := decrypted[k]
		if !ok {
			t.Fatalf("missing key %q after decrypt", k)
		}
		if got != want {
			t.Fatalf("key %q: got %q, want %q", k, got, want)
		}
	}
}

func TestDecryptMapMixedValues(t *testing.T) {
	key := testKey(t)

	encrypted, err := EncryptValue(key, "secret")
	if err != nil {
		t.Fatal(err)
	}

	vars := map[string]string{
		"ENCRYPTED": encrypted,
		"PLAIN":     "not-encrypted",
	}

	decrypted, err := DecryptMap(key, vars)
	if err != nil {
		t.Fatal(err)
	}

	if decrypted["ENCRYPTED"] != "secret" {
		t.Fatalf("ENCRYPTED: got %q, want %q", decrypted["ENCRYPTED"], "secret")
	}
	if decrypted["PLAIN"] != "not-encrypted" {
		t.Fatalf("PLAIN: got %q, want %q", decrypted["PLAIN"], "not-encrypted")
	}
}

func TestGenerateSaltRandomness(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Fatal("two generated salts should differ")
	}
	if len(salt1) != 16 {
		t.Fatalf("salt length: got %d, want 16", len(salt1))
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	salt := []byte("deterministic!!!")
	params := DefaultParams()

	key1 := DeriveKey("passphrase", salt, params)
	key2 := DeriveKey("passphrase", salt, params)

	if !bytes.Equal(key1, key2) {
		t.Fatal("same inputs should produce same key")
	}
	if len(key1) != 32 {
		t.Fatalf("key length: got %d, want 32", len(key1))
	}
}

func TestDeriveKeyDifferentSalts(t *testing.T) {
	params := DefaultParams()
	key1 := DeriveKey("passphrase", []byte("salt-one-16bytes"), params)
	key2 := DeriveKey("passphrase", []byte("salt-two-16bytes"), params)

	if bytes.Equal(key1, key2) {
		t.Fatal("different salts should produce different keys")
	}
}
