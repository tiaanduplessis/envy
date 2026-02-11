package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
)

func TestListCmd_Empty(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)

	out, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No projects") {
		t.Errorf("output = %q, want 'No projects'", out)
	}
}

func TestListCmd_WithProjects(t *testing.T) {
	store := setupTestStore(t)
	for _, name := range []string{"alpha", "bravo"} {
		p, _ := config.NewProject(name, []string{"dev"}, "dev")
		store.Save(p)
	}

	root := NewRootCmd(store)
	out, err := executeCommand(root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "bravo") {
		t.Errorf("output = %q, want both project names", out)
	}
}

func TestListCmd_JSON(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []listEntry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Name != "foo" {
		t.Errorf("name = %q, want %q", entries[0].Name, "foo")
	}
}

func TestListCmd_Quiet(t *testing.T) {
	store := setupTestStore(t)
	for _, name := range []string{"alpha", "bravo"} {
		p, _ := config.NewProject(name, nil, "")
		store.Save(p)
	}

	root := NewRootCmd(store)
	out, err := executeCommand(root, "list", "--quiet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), out)
	}
	if lines[0] != "alpha" || lines[1] != "bravo" {
		t.Errorf("lines = %v, want [alpha bravo]", lines)
	}
}

func TestListCmd_EmptyJSON(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)

	out, err := executeCommand(root, "list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("output = %q, want %q", strings.TrimSpace(out), "[]")
	}
}

func TestListCmd_CorruptProject(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "projects")
	store := config.NewStore(dir)

	p, _ := config.NewProject("good", nil, "")
	store.Save(p)

	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("invalid: [yaml"), 0o600)

	root := NewRootCmd(store)
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "good") {
		t.Errorf("stdout = %q, expected to contain 'good'", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Warning") || !strings.Contains(stderr.String(), "bad") {
		t.Errorf("stderr = %q, expected warning about 'bad'", stderr.String())
	}
}
