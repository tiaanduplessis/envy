package crypto

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const EnvPassphrase = "ENVY_PASSPHRASE"

// GetPassphrase resolves a passphrase with priority:
//  1. ENVY_PASSPHRASE env var
//  2. Interactive terminal prompt
func GetPassphrase(prompt string) (string, error) {
	if p := os.Getenv(EnvPassphrase); p != "" {
		return p, nil
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("project is encrypted: set %s or run interactively", EnvPassphrase)
	}

	fmt.Fprint(os.Stderr, prompt)
	passBytes, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}

	passphrase := string(passBytes)
	if passphrase == "" {
		return "", fmt.Errorf("passphrase cannot be empty")
	}

	return passphrase, nil
}

// GetPassphraseWithConfirm prompts twice and checks they match.
// Used when setting a new passphrase.
func GetPassphraseWithConfirm() (string, error) {
	pass1, err := GetPassphrase("Enter passphrase: ")
	if err != nil {
		return "", err
	}

	pass2, err := GetPassphrase("Confirm passphrase: ")
	if err != nil {
		return "", err
	}

	if pass1 != pass2 {
		return "", fmt.Errorf("passphrases do not match")
	}

	return pass1, nil
}
