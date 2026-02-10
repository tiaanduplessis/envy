package config

import (
	"os"
	"path/filepath"
	"testing"
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

func TestStore_CRUDCycle(t *testing.T) {
	store := setupTestStore(t)

	// Create
	p, _ := NewProject("lifecycle", []string{"dev"}, "dev")
	p.SetVar("dev", "KEY", "v1")
	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Read
	loaded, err := store.Load("lifecycle")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loaded.Environments["dev"]["KEY"]; got != "v1" {
		t.Errorf("KEY = %q, want %q", got, "v1")
	}

	// Update
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

	// Delete
	if err := store.Delete("lifecycle"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if store.Exists("lifecycle") {
		t.Error("expected project to be deleted")
	}
}
