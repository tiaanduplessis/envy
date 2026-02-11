package cli

import (
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
)

func TestUnsetCmd_SingleVar(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	executeCommand(root, "set", "foo", "DB=localhost", "PORT=5432")

	root = NewRootCmd(store)
	out, err := executeCommand(root, "unset", "foo", "DB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Removed 1 variable") {
		t.Errorf("output = %q", out)
	}

	p, _ := store.Load("foo")
	if _, ok := p.Environments["dev"]["DB"]; ok {
		t.Error("expected DB to be deleted")
	}
	if got := p.Environments["dev"]["PORT"]; got != "5432" {
		t.Errorf("PORT = %q, want %q", got, "5432")
	}
}

func TestUnsetCmd_MultipleVars(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	executeCommand(root, "set", "foo", "A=1", "B=2", "C=3")

	root = NewRootCmd(store)
	out, err := executeCommand(root, "unset", "foo", "A", "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Removed 2 variable") {
		t.Errorf("output = %q", out)
	}

	p, _ := store.Load("foo")
	if _, ok := p.Environments["dev"]["A"]; ok {
		t.Error("expected A to be deleted")
	}
	if _, ok := p.Environments["dev"]["C"]; ok {
		t.Error("expected C to be deleted")
	}
	if _, ok := p.Environments["dev"]["B"]; !ok {
		t.Error("expected B to still exist")
	}
}

func TestUnsetCmd_WithEnv(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo", "--env", "dev", "--env", "staging")

	root = NewRootCmd(store)
	executeCommand(root, "set", "foo", "DB=devdb", "--env", "dev")
	root = NewRootCmd(store)
	executeCommand(root, "set", "foo", "DB=stagingdb", "--env", "staging")

	root = NewRootCmd(store)
	_, err := executeCommand(root, "unset", "foo", "DB", "--env", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ := store.Load("foo")
	if _, ok := p.Environments["staging"]["DB"]; ok {
		t.Error("expected DB deleted from staging")
	}
	if got := p.Environments["dev"]["DB"]; got != "devdb" {
		t.Errorf("dev DB = %q, want %q", got, "devdb")
	}
}

func TestUnsetCmd_WithPath(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	executeCommand(root, "set", "foo", "PORT=3000", "--path", "services/api")

	root = NewRootCmd(store)
	out, err := executeCommand(root, "unset", "foo", "PORT", "--path", "services/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Removed 1 variable") {
		t.Errorf("output = %q", out)
	}

	p, _ := store.Load("foo")
	if _, ok := p.Paths["services/api"]["dev"]["PORT"]; ok {
		t.Error("expected PORT deleted from path")
	}
}

func TestUnsetCmd_NonexistentKey(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	executeCommand(root, "init", "foo")

	root = NewRootCmd(store)
	out, err := executeCommand(root, "unset", "foo", "NOPE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Removed 0 variable") {
		t.Errorf("output = %q", out)
	}
}

func TestUnsetCmd_Encrypted(t *testing.T) {
	store := setupEncryptedTestStore(t)
	t.Setenv("ENVY_PASSPHRASE", "test-passphrase")

	p, _ := config.NewProject("foo", nil, "")
	p.SetVar("dev", "SECRET", "hunter2")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "unset", "foo", "SECRET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ = store.Load("foo")
	if _, ok := p.Environments["dev"]["SECRET"]; ok {
		t.Error("expected SECRET to be deleted")
	}
}
