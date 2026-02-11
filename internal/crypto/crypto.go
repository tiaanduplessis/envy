package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// EncryptedPrefix marks a value as encrypted in the YAML file.
	EncryptedPrefix = "ENC:"

	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MiB
	argonThreads = 4
	argonKeyLen  = 32 // AES-256

	saltLength = 16
)

// Params holds Argon2id parameters stored alongside the project so they
// can be tuned per-project without breaking older projects.
type Params struct {
	Time    uint32 `yaml:"time"`
	Memory  uint32 `yaml:"memory"`
	Threads uint8  `yaml:"threads"`
}

var (
	ErrNotEncrypted    = errors.New("value is not encrypted")
	ErrDecryptionFailed = errors.New("decryption failed: wrong passphrase or corrupted data")
)

// DefaultParams returns the recommended Argon2id parameters.
func DefaultParams() Params {
	return Params{
		Time:    argonTime,
		Memory:  argonMemory,
		Threads: argonThreads,
	}
}

func DeriveKey(passphrase string, salt []byte, params Params) []byte {
	return argon2.IDKey(
		[]byte(passphrase), salt, params.Time, params.Memory, params.Threads, argonKeyLen,
	)
}

// GenerateSalt returns a cryptographically random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generating salt: %w", err)
	}
	return salt, nil
}

// EncryptValue encrypts a plaintext value and returns "ENC:<base64(nonce+ciphertext)>".
func EncryptValue(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return EncryptedPrefix + encoded, nil
}

// DecryptValue decrypts a value previously produced by EncryptValue.
func DecryptValue(key []byte, encrypted string) (string, error) {
	if !IsEncrypted(encrypted) {
		return "", ErrNotEncrypted
	}

	encoded := strings.TrimPrefix(encrypted, EncryptedPrefix)
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decoding base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, EncryptedPrefix)
}

// EncryptMap encrypts every value in a map, returning a new map.
// Returns an error if any value is already encrypted (ENC: prefix).
func EncryptMap(key []byte, vars map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(vars))
	for k, v := range vars {
		if IsEncrypted(v) {
			return nil, fmt.Errorf("value for %q is already encrypted", k)
		}
		enc, err := EncryptValue(key, v)
		if err != nil {
			return nil, fmt.Errorf("encrypting %q: %w", k, err)
		}
		result[k] = enc
	}
	return result, nil
}

// DecryptMap decrypts every ENC:-prefixed value in a map, returning a new map.
func DecryptMap(key []byte, vars map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(vars))
	for k, v := range vars {
		if IsEncrypted(v) {
			dec, err := DecryptValue(key, v)
			if err != nil {
				return nil, fmt.Errorf("decrypting %q: %w", k, err)
			}
			result[k] = dec
		} else {
			result[k] = v
		}
	}
	return result, nil
}
