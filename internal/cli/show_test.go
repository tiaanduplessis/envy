package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tiaanduplessis/envy/internal/config"
)

func TestShowCmd_FullProject(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "staging"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("staging", "DB", "staging-db")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "show", "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Project: foo") {
		t.Errorf("missing project header in output: %q", out)
	}
	if !strings.Contains(out, "[dev]") {
		t.Errorf("missing [dev] in output: %q", out)
	}
	if !strings.Contains(out, "[staging]") {
		t.Errorf("missing [staging] in output: %q", out)
	}
}

func TestShowCmd_SpecificEnv(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("dev", "PORT", "5432")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "show", "foo", "--env", "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DB=") && !strings.Contains(out, "PORT=") {
		t.Errorf("output = %q, expected key-value pairs", out)
	}
}

func TestShowCmd_Masking(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "SECRET", "super-secret-value")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "show", "foo", "--env", "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "super-secret-value") {
		t.Errorf("value should be masked: %q", out)
	}
	if !strings.Contains(out, "****") {
		t.Errorf("expected masked value in output: %q", out)
	}
}

func TestShowCmd_Reveal(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "SECRET", "super-secret-value")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "show", "foo", "--env", "dev", "--reveal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "super-secret-value") {
		t.Errorf("expected revealed value in output: %q", out)
	}
}

func TestShowCmd_JSON(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "show", "foo", "--env", "dev", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var vars map[string]string
	if err := json.Unmarshal([]byte(out), &vars); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if vars["DB"] != "localhost" {
		t.Errorf("DB = %q, want %q", vars["DB"], "localhost")
	}
}

func TestShowCmd_DisplaysEnvFiles(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "local", "staging"}, "dev")
	p.SetVar("dev", "DB", "localhost")
	p.SetEnvFile("local", ".env.local")
	p.SetEnvFile("staging", ".env.staging")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "show", "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Env files:") {
		t.Errorf("missing 'Env files:' section: %q", out)
	}
	if !strings.Contains(out, ".env.local") {
		t.Errorf("missing .env.local: %q", out)
	}
	if !strings.Contains(out, ".env.staging") {
		t.Errorf("missing .env.staging: %q", out)
	}
}

func TestShowCmd_NoEnvFiles(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev"}, "dev")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "show", "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "Env files:") {
		t.Errorf("should not show env files section when none configured: %q", out)
	}
}

func TestShowCmd_JSON_IncludesEnvFiles(t *testing.T) {
	store := setupTestStore(t)
	p, _ := config.NewProject("foo", []string{"dev", "local"}, "dev")
	p.SetEnvFile("local", ".env.local")
	store.Save(p)

	root := NewRootCmd(store)
	out, err := executeCommand(root, "show", "foo", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result showOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result.EnvFiles["local"] != ".env.local" {
		t.Errorf("EnvFiles[local] = %q, want %q", result.EnvFiles["local"], ".env.local")
	}
}

func TestShowCmd_NonexistentProject(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)
	_, err := executeCommand(root, "show", "nope")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}
