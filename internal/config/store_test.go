package config

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/crypto"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(filepath.Join(t.TempDir(), "projects"))
}

func TestStore_SaveAndLoad(t *testing.T) {
	store := setupTestStore(t)
	p, _ := NewProject("test", []string{"dev", "staging"}, "dev")
	p.SetVar("dev", "DB_HOST", "localhost")
	p.SetVar("dev", "DB_PORT", "5432")

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load("test")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Name != "test" {
		t.Errorf("Name = %q, want %q", loaded.Name, "test")
	}
	if loaded.DefaultEnv != "dev" {
		t.Errorf("DefaultEnv = %q, want %q", loaded.DefaultEnv, "dev")
	}
	if got := loaded.Environments["dev"]["DB_HOST"]; got != "localhost" {
		t.Errorf("DB_HOST = %q, want %q", got, "localhost")
	}
	if got := loaded.Environments["dev"]["DB_PORT"]; got != "5432" {
		t.Errorf("DB_PORT = %q, want %q", got, "5432")
	}
}

func TestStore_RoundTrip_WithPaths(t *testing.T) {
	store := setupTestStore(t)
	p, _ := NewProject("mono", []string{"dev"}, "dev")
	p.SetVar("dev", "SHARED", "true")
	p.SetPathVar("services/api", "dev", "PORT", "3000")

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load("mono")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := loaded.Environments["dev"]["SHARED"]; got != "true" {
		t.Errorf("SHARED = %q, want %q", got, "true")
	}
	if got := loaded.Paths["services/api"]["dev"]["PORT"]; got != "3000" {
		t.Errorf("PORT = %q, want %q", got, "3000")
	}
}

func TestStore_LoadNonexistent(t *testing.T) {
	store := setupTestStore(t)
	_, err := store.Load("nonexistent")
	if err == nil {
		t.Error("expected error loading nonexistent project")
	}
}

func TestStore_ListEmpty(t *testing.T) {
	store := setupTestStore(t)
	names, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected empty list, got %v", names)
	}
}

func TestStore_ListProjects(t *testing.T) {
	store := setupTestStore(t)

	for _, name := range []string{"charlie", "alpha", "bravo"} {
		p, _ := NewProject(name, nil, "")
		if err := store.Save(p); err != nil {
			t.Fatalf("Save(%q): %v", name, err)
		}
	}

	names, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	want := []string{"alpha", "bravo", "charlie"}
	if len(names) != len(want) {
		t.Fatalf("got %d projects, want %d", len(names), len(want))
	}
	for i, name := range names {
		if name != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, name, want[i])
		}
	}
}

func TestStore_Delete(t *testing.T) {
	store := setupTestStore(t)
	p, _ := NewProject("todelete", nil, "")
	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if !store.Exists("todelete") {
		t.Fatal("expected project to exist")
	}

	if err := store.Delete("todelete"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if store.Exists("todelete") {
		t.Error("expected project to be deleted")
	}
}

func TestStore_DeleteNonexistent(t *testing.T) {
	store := setupTestStore(t)
	err := store.Delete("nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent project")
	}
}

func TestStore_Exists(t *testing.T) {
	store := setupTestStore(t)

	if store.Exists("nope") {
		t.Error("expected Exists to return false for nonexistent project")
	}

	p, _ := NewProject("exists", nil, "")
	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if !store.Exists("exists") {
		t.Error("expected Exists to return true for saved project")
	}
}

func TestStore_LoadInvalidYAML(t *testing.T) {
	store := setupTestStore(t)
	os.MkdirAll(store.basePath, 0o755)
	path := filepath.Join(store.basePath, "bad.yaml")
	os.WriteFile(path, []byte("{{{{not yaml"), 0o644)

	_, err := store.Load("bad")
	if err == nil {
		t.Error("expected error loading invalid YAML")
	}
}

func TestStore_RoundTrip_WithEnvFiles(t *testing.T) {
	store := setupTestStore(t)
	p, _ := NewProject("test", []string{"dev", "local", "staging"}, "dev")
	p.SetEnvFile("local", ".env.local")
	p.SetEnvFile("staging", ".env.staging")

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load("test")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := loaded.EnvFiles["local"]; got != ".env.local" {
		t.Errorf("EnvFiles[local] = %q, want %q", got, ".env.local")
	}
	if got := loaded.EnvFiles["staging"]; got != ".env.staging" {
		t.Errorf("EnvFiles[staging] = %q, want %q", got, ".env.staging")
	}
}

func TestStore_CRUDCycle(t *testing.T) {
	store := setupTestStore(t)

	p, _ := NewProject("lifecycle", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "v1")
	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load("lifecycle")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loaded.Environments["dev"]["KEY"]; got != "v1" {
		t.Errorf("KEY = %q, want %q", got, "v1")
	}

	loaded.SetVar("dev", "KEY", "v2")
	if err := store.Save(loaded); err != nil {
		t.Fatalf("Save update: %v", err)
	}

	reloaded, err := store.Load("lifecycle")
	if err != nil {
		t.Fatalf("Load after update: %v", err)
	}
	if got := reloaded.Environments["dev"]["KEY"]; got != "v2" {
		t.Errorf("KEY = %q after update, want %q", got, "v2")
	}

	if err := store.Delete("lifecycle"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if store.Exists("lifecycle") {
		t.Error("expected project to be deleted")
	}
}

func setupEncryptedTestStore(t *testing.T) *Store {
	t.Helper()
	store := NewStore(filepath.Join(t.TempDir(), "projects"))
	store.SetPassphraseFunc(func(string) (string, error) {
		return "test-passphrase", nil
	})
	return store
}

func enableEncryption(t *testing.T, p *Project) {
	t.Helper()
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	p.Encryption = &EncryptionConfig{
		Enabled: true,
		Salt:    base64.StdEncoding.EncodeToString(salt),
		Params:  crypto.DefaultParams(),
	}
}

func TestStore_EncryptedRoundTrip(t *testing.T) {
	store := setupEncryptedTestStore(t)
	p, _ := NewProject("secret", []string{"dev", "prod"}, "dev")
	p.SetVar("dev", "DB_PASSWORD", "hunter2")
	p.SetVar("dev", "API_KEY", "sk-12345")
	p.SetVar("prod", "DB_PASSWORD", "p4ssw0rd")
	p.SetPathVar("services/api", "dev", "PORT", "3000")
	enableEncryption(t, p)

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load("secret")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := loaded.Environments["dev"]["DB_PASSWORD"]; got != "hunter2" {
		t.Errorf("DB_PASSWORD = %q, want %q", got, "hunter2")
	}
	if got := loaded.Environments["dev"]["API_KEY"]; got != "sk-12345" {
		t.Errorf("API_KEY = %q, want %q", got, "sk-12345")
	}
	if got := loaded.Environments["prod"]["DB_PASSWORD"]; got != "p4ssw0rd" {
		t.Errorf("prod DB_PASSWORD = %q, want %q", got, "p4ssw0rd")
	}
	if got := loaded.Paths["services/api"]["dev"]["PORT"]; got != "3000" {
		t.Errorf("PORT = %q, want %q", got, "3000")
	}
}

func TestStore_EncryptedValuesOnDisk(t *testing.T) {
	store := setupEncryptedTestStore(t)
	p, _ := NewProject("ondisk", []string{"dev"}, "dev")
	p.SetVar("dev", "SECRET", "plaintext-value")
	enableEncryption(t, p)

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, err := store.LoadRaw("ondisk")
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	val := raw.Environments["dev"]["SECRET"]
	if !strings.HasPrefix(val, crypto.EncryptedPrefix) {
		t.Errorf("value on disk should be encrypted, got %q", val)
	}
}

func TestStore_EncryptedWrongPassphrase(t *testing.T) {
	store := setupEncryptedTestStore(t)
	p, _ := NewProject("locked", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "value")
	enableEncryption(t, p)

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	wrongStore := NewStore(store.basePath)
	wrongStore.SetPassphraseFunc(func(string) (string, error) {
		return "wrong-passphrase", nil
	})

	_, err := wrongStore.Load("locked")
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
	if !strings.Contains(err.Error(), "decryption failed") {
		t.Errorf("expected decryption error, got: %v", err)
	}
}

func TestStore_EncryptedNoPassphraseFunc(t *testing.T) {
	store := setupEncryptedTestStore(t)
	p, _ := NewProject("nofunc", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "value")
	enableEncryption(t, p)

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	bareStore := NewStore(store.basePath)
	_, err := bareStore.Load("nofunc")
	if err == nil {
		t.Fatal("expected error with no passphrase func")
	}
	if !strings.Contains(err.Error(), "no passphrase provider") {
		t.Errorf("expected passphrase provider error, got: %v", err)
	}
}

func TestStore_EncryptedLoadSaveCycle(t *testing.T) {
	store := setupEncryptedTestStore(t)
	p, _ := NewProject("cycle", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "original")
	enableEncryption(t, p)

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load("cycle")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	loaded.SetVar("dev", "KEY", "updated")
	if err := store.Save(loaded); err != nil {
		t.Fatalf("Save after update: %v", err)
	}

	reloaded, err := store.Load("cycle")
	if err != nil {
		t.Fatalf("Load after update: %v", err)
	}
	if got := reloaded.Environments["dev"]["KEY"]; got != "updated" {
		t.Errorf("KEY = %q, want %q", got, "updated")
	}
}

func TestStore_UnencryptedProjectUnchanged(t *testing.T) {
	store := setupEncryptedTestStore(t)
	p, _ := NewProject("plain", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "value")

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, err := store.LoadRaw("plain")
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	val := raw.Environments["dev"]["KEY"]
	if val != "value" {
		t.Errorf("unencrypted value should be plaintext, got %q", val)
	}
	if raw.IsEncrypted() {
		t.Error("unencrypted project should not report as encrypted")
	}
}
