package cli

import (
	"strings"
	"testing"
)

func TestLLMFlag(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)

	output, err := executeCommand(root, "--llm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	requiredSections := []string{
		"# envy",
		"## Environment Variables",
		"## Key Concepts",
		"## Commands",
		"## Common Workflows",
	}
	for _, section := range requiredSections {
		if !strings.Contains(output, section) {
			t.Errorf("output missing section %q", section)
		}
	}
}

func TestLLMFlagIncludesAllCommands(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)

	output, err := executeCommand(root, "--llm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := []string{
		"envy init",
		"envy scan",
		"envy set",
		"envy get",
		"envy list",
		"envy show",
		"envy load",
		"envy update",
		"envy delete",
		"envy diff",
		"envy encrypt",
		"envy decrypt",
		"envy rekey",
		"envy env add",
		"envy env remove",
		"envy env list",
		"envy env copy",
		"envy env file set",
		"envy env file clear",
	}
	for _, cmd := range commands {
		if !strings.Contains(output, cmd) {
			t.Errorf("output missing command %q", cmd)
		}
	}
}

func TestLLMFlagIncludesFlags(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)

	output, err := executeCommand(root, "--llm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	flags := []string{
		"--env",
		"--path",
		"--json",
		"--reveal",
		"--force",
		"--dry-run",
		"--project",
		"--quiet",
		"--merge",
		"--format",
		"--output",
		"--all-paths",
		"--encrypt",
		"--default-env",
	}
	for _, flag := range flags {
		if !strings.Contains(output, flag) {
			t.Errorf("output missing flag %q", flag)
		}
	}
}

func TestLLMFlagIncludesEnvVars(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)

	output, err := executeCommand(root, "--llm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envVars := []string{
		"ENVY_CONFIG_DIR",
		"ENVY_ENV",
		"ENVY_PASSPHRASE",
	}
	for _, v := range envVars {
		if !strings.Contains(output, v) {
			t.Errorf("output missing env var %q", v)
		}
	}
}

func TestWithoutLLMFlagShowsHelp(t *testing.T) {
	store := setupTestStore(t)
	root := NewRootCmd(store)

	output, err := executeCommand(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Usage:") {
		t.Error("expected help output with Usage: section")
	}
	if strings.Contains(output, "## Commands") {
		t.Error("did not expect LLM output without --llm flag")
	}
}
