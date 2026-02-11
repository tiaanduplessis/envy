package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
)

func TestComputeDiff(t *testing.T) {
	left := map[string]string{"A": "1", "B": "2", "C": "same"}
	right := map[string]string{"B": "changed", "C": "same", "D": "4"}

	diff := ComputeDiff(left, right)

	if len(diff.Removed) != 1 || diff.Removed["A"] != "1" {
		t.Errorf("Removed = %v, want {A: 1}", diff.Removed)
	}
	if len(diff.Added) != 1 || diff.Added["D"] != "4" {
		t.Errorf("Added = %v, want {D: 4}", diff.Added)
	}
	if len(diff.Changed) != 1 {
		t.Errorf("Changed = %v, want 1 entry", diff.Changed)
	}
	if pair := diff.Changed["B"]; pair[0] != "2" || pair[1] != "changed" {
		t.Errorf("Changed[B] = %v, want [2, changed]", pair)
	}
}

func TestComputeDiff_Identical(t *testing.T) {
	vars := map[string]string{"A": "1", "B": "2"}
	diff := ComputeDiff(vars, vars)
	if !diff.IsEmpty() {
		t.Errorf("expected empty diff: %+v", diff)
	}
}

func TestDiffCmd_TwoEnvs(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("dev", "DEBUG", "true")
	p.SetVar("staging", "DB", "staging-db")
	p.SetVar("staging", "EXTRA", "yes")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "diff", "foo",
		"--env", "dev", "--env", "staging", "--reveal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "- DEBUG=true") {
		t.Errorf("missing removed DEBUG: %q", out)
	}
	if !strings.Contains(out, "+ EXTRA=yes") {
		t.Errorf("missing added EXTRA: %q", out)
	}
	if !strings.Contains(out, "~ DB:") {
		t.Errorf("missing changed DB: %q", out)
	}
}

func TestDiffCmd_IdenticalEnvs(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("staging", "DB", "localhost")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "diff", "foo",
		"--env", "dev", "--env", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No differences") {
		t.Errorf("expected 'No differences': %q", out)
	}
}

func TestDiffCmd_AgainstLocalFile(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("dev", "PORT", "5432")
	store.Save(p)

	dir := t.TempDir()
	envFile := dir + "/.env"
	os.WriteFile(envFile, []byte("DB=different\nEXTRA=local\n"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "diff", "foo", "--env", "dev", "--reveal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "- EXTRA=local") {
		t.Errorf("missing removed EXTRA: %q", out)
	}
	if !strings.Contains(out, "+ PORT=5432") {
		t.Errorf("missing added PORT: %q", out)
	}
	if !strings.Contains(out, "~ DB:") {
		t.Errorf("missing changed DB: %q", out)
	}
}

func TestDiffCmd_NonexistentEnv(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "diff", "foo",
		"--env", "dev", "--env", "nope")
	if err == nil {
		t.Error("expected error for nonexistent env")
	}
}

func TestDiffCmd_MaskedByDefault(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	p.SetVar("dev", "SECRET", "super-secret")
	p.SetVar("staging", "SECRET", "other-secret")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "diff", "foo",
		"--env", "dev", "--env", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "super-secret") || strings.Contains(out, "other-secret") {
		t.Errorf("values should be masked: %q", out)
	}
	if !strings.Contains(out, "****") {
		t.Errorf("expected masked values: %q", out)
	}
}

func TestDiffCmd_NoEnvFlags(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "diff", "foo")
	if err == nil {
		t.Error("expected error when no --env flags provided")
	}
}
