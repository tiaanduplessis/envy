package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func TestUpdateCmd_Basic(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("DB=localhost\nPORT=5432\n"), 0o644)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "update", "--project", "foo", "--file", envFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Updated 2 variable") {
		t.Errorf("output = %q", out)
	}

	p, _ = store.Load("foo")
	if got := p.Environments["dev"]["DB"]; got != "localhost" {
		t.Errorf("DB = %q, want %q", got, "localhost")
	}
	if got := p.Environments["dev"]["PORT"]; got != "5432" {
		t.Errorf("PORT = %q, want %q", got, "5432")
	}
}

func TestUpdateCmd_PreservesDisabledVars(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("DB=localhost\n# API_KEY=disabled\n"), 0o644)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "update", "--project", "foo", "--file", envFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Updated 2 variable") {
		t.Errorf("output = %q", out)
	}

	p, _ = store.Load("foo")
	if got := p.Environments["dev"]["DB"]; got != "localhost" {
		t.Errorf("DB = %q, want %q", got, "localhost")
	}
	if got := p.DisabledEnvironments["dev"]["API_KEY"]; got != "disabled" {
		t.Errorf("disabled API_KEY = %q, want %q", got, "disabled")
	}
}

func TestUpdateCmd_MergeMode(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "DB", "original")
	store.Save(p)

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("DB=new-value\nPORT=5432\n"), 0o644)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "update", "--project", "foo",
		"--file", envFile, "--merge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Merged 1 variable") {
		t.Errorf("output = %q", out)
	}

	p, _ = store.Load("foo")
	if got := p.Environments["dev"]["DB"]; got != "original" {
		t.Errorf("DB should be unchanged: got %q, want %q", got, "original")
	}
	if got := p.Environments["dev"]["PORT"]; got != "5432" {
		t.Errorf("PORT = %q, want %q", got, "5432")
	}
}

func TestUpdateCmd_WithEnvAndPath(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	store.Save(p)

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("PORT=3000\n"), 0o644)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "update", "--project", "foo",
		"--file", envFile, "--env", "staging", "--path", "services/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ = store.Load("foo")
	if got := p.Paths["services/api"]["staging"]["PORT"]; got != "3000" {
		t.Errorf("PORT = %q, want %q", got, "3000")
	}
}

func TestUpdateCmd_NonexistentProject(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	_, err := executeCommand(root, "update", "--project", "nope")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestUpdateCmd_RoundTrip(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("dev", "PORT", "5432")
	p.SetVar("dev", "URL", "postgres://host:5432/db?opt=val")
	store.Save(p)

	dir := t.TempDir()
	outFile := filepath.Join(dir, ".env")

	root := NewRootCmd(store)
	_, err := executeCommand(root, "load", "--project", "foo", "--output", outFile)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	p2, _ := config.NewProject("bar", []string{"dev"}, "dev")
	store.Save(p2)

	root = NewRootCmd(store)
	_, err = executeCommand(root, "update", "--project", "bar", "--file", outFile)
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	p2, _ = store.Load("bar")
	original, _ := store.Load("foo")

	for key, want := range original.Environments["dev"] {
		got := p2.Environments["dev"][key]
		if got != want {
			t.Errorf("key %q: got %q, want %q", key, got, want)
		}
	}
}

func TestUpdateCmd_Encrypted(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv(crypto.EnvPassphrase, "test-passphrase")

	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "EXISTING", "keep-me")
	store.Save(p)

	cmd := NewRootCmd(store)
	if _, err := executeCommand(cmd, "encrypt", "foo"); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("NEW=from-file\nEXISTING=updated\n# OLD=disabled\n"), 0o644)

	cmd = NewRootCmd(store)
	_, err := executeCommand(cmd, "update", "--project", "foo", "--file", envFile)
	if err != nil {
		t.Fatalf("update on encrypted project: %v", err)
	}

	raw, _ := store.LoadRaw("foo")
	if !crypto.IsEncrypted(raw.Environments["dev"]["NEW"]) {
		t.Error("new value should be encrypted on disk")
	}
	if !crypto.IsEncrypted(raw.DisabledEnvironments["dev"]["OLD"]) {
		t.Error("disabled value should be encrypted on disk")
	}

	loaded, _ := store.Load("foo")
	if got := loaded.Environments["dev"]["NEW"]; got != "from-file" {
		t.Errorf("NEW = %q, want %q", got, "from-file")
	}
	if got := loaded.Environments["dev"]["EXISTING"]; got != "updated" {
		t.Errorf("EXISTING = %q, want %q", got, "updated")
	}
	if got := loaded.DisabledEnvironments["dev"]["OLD"]; got != "disabled" {
		t.Errorf("OLD = %q, want %q", got, "disabled")
	}
}
