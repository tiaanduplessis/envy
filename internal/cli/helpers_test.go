package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func setupTestStore(t *testing.T) *config.Store {
	t.Helper()
	return config.NewStore(filepath.Join(t.TempDir(), "projects"))
}

func setupEncryptedTestStore(t *testing.T) *config.Store {
	t.Helper()
	store := config.NewStore(filepath.Join(t.TempDir(), "projects"))
	store.SetPassphraseFunc(func(string) (string, error) {
		return "test-passphrase", nil
	})
	return store
}

func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}
