package cli

import (
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
)

func TestEnvAddCmd(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "env", "add", "foo", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Added environment") {
		t.Errorf("output = %q", out)
	}

	p, _ = store.Load("foo")
	if _, ok := p.Environments["staging"]; !ok {
		t.Error("expected staging environment to exist")
	}
}

func TestEnvAddCmd_Duplicate(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "env", "add", "foo", "dev")
	if err == nil {
		t.Error("expected error for duplicate environment")
	}
}

func TestEnvRemoveCmd_WithForce(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	p.SetVar("staging", "DB", "staging-db")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "env", "remove", "foo", "staging", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ = store.Load("foo")
	if _, ok := p.Environments["staging"]; ok {
		t.Error("staging should be removed")
	}
}

func TestEnvRemoveCmd_WithConfirmation(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	store.Save(p)

	// Confirm with "y"
	envCmd := newEnvCmd(store, strings.NewReader("y\n"))
	envCmd.SetArgs([]string{"remove", "foo", "staging"})
	var buf strings.Builder
	envCmd.SetOut(&buf)
	envCmd.SetErr(&buf)
	envCmd.Execute()

	p, _ = store.Load("foo")
	if _, ok := p.Environments["staging"]; ok {
		t.Error("staging should be removed after confirmation")
	}
}

func TestEnvRemoveCmd_Declined(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	store.Save(p)

	envCmd := newEnvCmd(store, strings.NewReader("n\n"))
	envCmd.SetArgs([]string{"remove", "foo", "staging"})
	var buf strings.Builder
	envCmd.SetOut(&buf)
	envCmd.SetErr(&buf)
	envCmd.Execute()

	p, _ = store.Load("foo")
	if _, ok := p.Environments["staging"]; !ok {
		t.Error("staging should still exist after declining")
	}
}

func TestEnvRemoveCmd_NonexistentEnv(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "env", "remove", "foo", "nope", "--force")
	if err == nil {
		t.Error("expected error for nonexistent environment")
	}
}

func TestEnvListCmd(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging", "prod"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "env", "list", "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	want := []string{"dev", "prod", "staging"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d: %q", len(lines), len(want), out)
	}
	for i, line := range lines {
		if line != want[i] {
			t.Errorf("line %d = %q, want %q", i, line, want[i])
		}
	}
}

func TestEnvCopyCmd(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("dev", "PORT", "5432")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "env", "copy", "foo",
		"--from", "dev", "--to", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Copied 2 variable") {
		t.Errorf("output = %q", out)
	}

	p, _ = store.Load("foo")

	// Verify staging has the copied vars
	if got := p.Environments["staging"]["DB"]; got != "localhost" {
		t.Errorf("staging DB = %q, want %q", got, "localhost")
	}
	if got := p.Environments["staging"]["PORT"]; got != "5432" {
		t.Errorf("staging PORT = %q, want %q", got, "5432")
	}

	// Verify source unchanged
	if got := p.Environments["dev"]["DB"]; got != "localhost" {
		t.Errorf("dev DB should be unchanged: %q", got)
	}
}

func TestEnvCopyCmd_NonexistentSource(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "env", "copy", "foo",
		"--from", "nope", "--to", "staging")
	if err == nil {
		t.Error("expected error for nonexistent source env")
	}
}

func TestEnvFileSetCmd(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "local"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "env", "file", "set", "foo", "local", ".env.local")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, ".env.local") {
		t.Errorf("output = %q", out)
	}

	p, _ = store.Load("foo")
	if got := p.EnvFiles["local"]; got != ".env.local" {
		t.Errorf("EnvFiles[local] = %q, want %q", got, ".env.local")
	}
}

func TestEnvFileSetCmd_NonexistentEnv(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "env", "file", "set", "foo", "nope", ".env.nope")
	if err == nil {
		t.Error("expected error for nonexistent environment")
	}
}

func TestEnvFileClearCmd(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "local"}, "dev")
	p.SetEnvFile("local", ".env.local")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "env", "file", "clear", "foo", "local")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Cleared") {
		t.Errorf("output = %q", out)
	}

	p, _ = store.Load("foo")
	if _, ok := p.EnvFiles["local"]; ok {
		t.Error("expected env file mapping to be cleared")
	}
}

func TestEnvListCmd_ShowsFileMapping(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "local", "staging"}, "dev")
	p.SetEnvFile("local", ".env.local")
	p.SetEnvFile("staging", ".env.staging")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "env", "list", "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "dev") {
		t.Errorf("missing dev in output: %q", out)
	}
	if !strings.Contains(out, "(.env.local)") {
		t.Errorf("missing .env.local annotation: %q", out)
	}
	if !strings.Contains(out, "(.env.staging)") {
		t.Errorf("missing .env.staging annotation: %q", out)
	}
}

func TestEnvRemoveCmd_AlsoRemovesEnvFile(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	p.SetEnvFile("staging", ".env.staging")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "env", "remove", "foo", "staging", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ = store.Load("foo")
	if _, ok := p.EnvFiles["staging"]; ok {
		t.Error("env file mapping for staging should be removed")
	}
}

func TestEnvRemoveCmd_AlsoRemovesFromPaths(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	p.SetPathVar("services/api", "staging", "PORT", "3000")
	store.Save(p)

	root := NewRootCmd(store)
	_, err := executeCommand(root, "env", "remove", "foo", "staging", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ = store.Load("foo")
	if vars := p.GetPathVars("services/api", "staging"); vars != nil {
		t.Errorf("path vars for staging should be removed, got %v", vars)
	}
}
