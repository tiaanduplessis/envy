package cli

import (
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
)

func TestSetCmd_SingleVar(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	out, err := executeCommand(root, "set", "foo", "DB=localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set 1 variable") {
		t.Errorf("output = %q", out)
	}

	p, _ := store.Load("foo")
	if got := p.Environments["dev"]["DB"]; got != "localhost" {
		t.Errorf("DB = %q, want %q", got, "localhost")
	}
}

func TestSetCmd_MultipleVars(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	_, err := executeCommand(root, "set", "foo", "A=1", "B=2", "C=3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ := store.Load("foo")
	for _, k := range []string{"A", "B", "C"} {
		if _, ok := p.Environments["dev"][k]; !ok {
			t.Errorf("expected key %q", k)
		}
	}
}

func TestSetCmd_WithEnv(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo", "--env", "dev", "--env", "staging")

	root = NewRootCmd(store)
	_, err := executeCommand(root, "set", "foo", "DB=staging-db", "--env", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ := store.Load("foo")
	if got := p.Environments["staging"]["DB"]; got != "staging-db" {
		t.Errorf("DB = %q, want %q", got, "staging-db")
	}
}

func TestSetCmd_WithPath(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	_, err := executeCommand(root, "set", "foo", "PORT=3000", "--path", "services/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ := store.Load("foo")
	if got := p.Paths["services/api"]["dev"]["PORT"]; got != "3000" {
		t.Errorf("PORT = %q, want %q", got, "3000")
	}
}

func TestSetCmd_InvalidFormat(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	_, err := executeCommand(root, "set", "foo", "NOEQUALS")
	if err == nil {
		t.Error("expected error for missing =")
	}
}

func TestSetCmd_ValueContainingEquals(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	_, err := executeCommand(root, "set", "foo", "URL=postgres://host:5432/db?opt=val")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ := store.Load("foo")
	want := "postgres://host:5432/db?opt=val"
	if got := p.Environments["dev"]["URL"]; got != want {
		t.Errorf("URL = %q, want %q", got, want)
	}
}

func TestSetCmd_AutoCreatesEnv(t *testing.T) {
	store := setupTestStore(t)

	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "set", "foo", "KEY=val", "--env", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ = store.Load("foo")
	if got := p.Environments["staging"]["KEY"]; got != "val" {
		t.Errorf("KEY = %q, want %q", got, "val")
	}
}
