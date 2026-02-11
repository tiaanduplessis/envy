package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEnvFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanCmd_Basic(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "DB=localhost\nPORT=5432\n")
	writeEnvFile(t, filepath.Join(dir, ".env.prod"), "DB=prod-db\n")

	root := newScanCmd(store, strings.NewReader("y\n"))
	out, err := executeCommand(root, "my-app", dir, "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Created project") {
		t.Errorf("expected success message, got: %q", out)
	}

	p, err := store.Load("my-app")
	if err != nil {
		t.Fatalf("loading project: %v", err)
	}

	if p.DefaultEnv != "dev" {
		t.Errorf("expected default env 'dev', got %q", p.DefaultEnv)
	}
	if p.Environments["dev"]["DB"] != "localhost" {
		t.Errorf("expected DB=localhost in dev, got %q", p.Environments["dev"]["DB"])
	}
	if p.Environments["prod"]["DB"] != "prod-db" {
		t.Errorf("expected DB=prod-db in prod, got %q", p.Environments["prod"]["DB"])
	}
}

func TestScanCmd_DryRun(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "DB=localhost\n")

	root := NewRootCmd(store)
	out, err := executeCommand(root, "scan", "my-app", dir, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "dev") {
		t.Errorf("expected env listing in output: %q", out)
	}

	if store.Exists("my-app") {
		t.Error("project should not be created in dry-run mode")
	}
}

func TestScanCmd_ProjectExists(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "DB=localhost\n")

	root := NewRootCmd(store)
	executeCommand(root, "init", "my-app")

	root = NewRootCmd(store)
	_, err := executeCommand(root, "scan", "my-app", dir)
	if err == nil {
		t.Error("expected error when project exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestScanCmd_ForceOverwrite(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "DB=localhost\n")

	root := NewRootCmd(store)
	executeCommand(root, "init", "my-app")

	root = NewRootCmd(store)
	_, err := executeCommand(root, "scan", "my-app", dir, "--force")
	if err != nil {
		t.Fatalf("unexpected error with --force: %v", err)
	}

	p, err := store.Load("my-app")
	if err != nil {
		t.Fatalf("loading project: %v", err)
	}
	if p.Environments["dev"]["DB"] != "localhost" {
		t.Errorf("expected DB=localhost after force overwrite, got %q", p.Environments["dev"]["DB"])
	}
}

func TestScanCmd_NoFilesFound(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()

	root := NewRootCmd(store)
	_, err := executeCommand(root, "scan", "my-app", dir, "--force")
	if err == nil {
		t.Error("expected error when no .env files found")
	}
	if !strings.Contains(err.Error(), "no .env files found") {
		t.Errorf("expected 'no .env files' error, got: %v", err)
	}
}

func TestScanCmd_DefaultEnvFlag(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env.staging"), "DB=staging\n")
	writeEnvFile(t, filepath.Join(dir, ".env.prod"), "DB=prod\n")

	root := NewRootCmd(store)
	_, err := executeCommand(root, "scan", "my-app", dir, "--force", "--default-env", "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := store.Load("my-app")
	if err != nil {
		t.Fatalf("loading project: %v", err)
	}
	if p.DefaultEnv != "prod" {
		t.Errorf("expected default env 'prod', got %q", p.DefaultEnv)
	}
}

func TestScanCmd_Monorepo(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "SHARED=true\n")
	writeEnvFile(t, filepath.Join(dir, "services", "api", ".env"), "PORT=3000\n")
	writeEnvFile(t, filepath.Join(dir, "services", "api", ".env.prod"), "PORT=8080\n")
	writeEnvFile(t, filepath.Join(dir, "services", "worker", ".env"), "CONCURRENCY=4\n")

	root := NewRootCmd(store)
	_, err := executeCommand(root, "scan", "mono", dir, "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := store.Load("mono")
	if err != nil {
		t.Fatalf("loading project: %v", err)
	}

	if p.Environments["dev"]["SHARED"] != "true" {
		t.Error("missing SHARED in root dev env")
	}

	apiPath := filepath.Join("services", "api")
	if p.Paths[apiPath] == nil {
		t.Fatalf("missing path %q", apiPath)
	}
	if p.Paths[apiPath]["dev"]["PORT"] != "3000" {
		t.Errorf("expected PORT=3000 in api dev, got %q", p.Paths[apiPath]["dev"]["PORT"])
	}
	if p.Paths[apiPath]["prod"]["PORT"] != "8080" {
		t.Errorf("expected PORT=8080 in api prod, got %q", p.Paths[apiPath]["prod"]["PORT"])
	}

	workerPath := filepath.Join("services", "worker")
	if p.Paths[workerPath] == nil {
		t.Fatalf("missing path %q", workerPath)
	}
	if p.Paths[workerPath]["dev"]["CONCURRENCY"] != "4" {
		t.Errorf("expected CONCURRENCY=4 in worker dev, got %q", p.Paths[workerPath]["dev"]["CONCURRENCY"])
	}
}

func TestScanCmd_Aborted(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "DB=localhost\n")

	cmd := newScanCmd(store, strings.NewReader("n\n"))
	out, err := executeCommand(cmd, "my-app", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Aborted") {
		t.Errorf("expected abort message, got: %q", out)
	}

	if store.Exists("my-app") {
		t.Error("project should not be created after abort")
	}
}

func TestScanCmd_SortedOutput(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "A=1\n")
	writeEnvFile(t, filepath.Join(dir, ".env.staging"), "A=2\n")
	writeEnvFile(t, filepath.Join(dir, ".env.prod"), "A=3\n")

	root := NewRootCmd(store)
	out, err := executeCommand(root, "scan", "my-app", dir, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	devIdx := strings.Index(out, "dev")
	prodIdx := strings.Index(out, "prod")
	stagingIdx := strings.Index(out, "staging")

	if devIdx > prodIdx || prodIdx > stagingIdx {
		t.Errorf("expected sorted env output (dev, prod, staging), got:\n%s", out)
	}
}

func TestScanCmd_InvalidProjectName(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "DB=localhost\n")

	root := NewRootCmd(store)
	_, err := executeCommand(root, "scan", "invalid name!", dir, "--force")
	if err == nil {
		t.Error("expected error for invalid project name")
	}
}

func TestScanCmd_PopulatesEnvFiles(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env"), "DB=localhost\n")
	writeEnvFile(t, filepath.Join(dir, ".env.local"), "DB=local\n")
	writeEnvFile(t, filepath.Join(dir, ".env.staging"), "DB=staging\n")

	root := NewRootCmd(store)
	_, err := executeCommand(root, "scan", "my-app", dir, "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := store.Load("my-app")
	if err != nil {
		t.Fatalf("loading project: %v", err)
	}

	// .env.local and .env.staging should have env file mappings
	if got := p.EnvFiles["local"]; got != ".env.local" {
		t.Errorf("EnvFiles[local] = %q, want %q", got, ".env.local")
	}
	if got := p.EnvFiles["staging"]; got != ".env.staging" {
		t.Errorf("EnvFiles[staging] = %q, want %q", got, ".env.staging")
	}

	// Bare .env (mapped to dev) should not have a mapping
	if _, ok := p.EnvFiles["dev"]; ok {
		t.Error("bare .env should not create an env file mapping for dev")
	}
}

func TestScanCmd_DefaultEnvFallback(t *testing.T) {
	store := setupTestStore(t)
	dir := t.TempDir()
	writeEnvFile(t, filepath.Join(dir, ".env.staging"), "DB=staging\n")

	root := NewRootCmd(store)
	_, err := executeCommand(root, "scan", "my-app", dir, "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := store.Load("my-app")
	if err != nil {
		t.Fatalf("loading project: %v", err)
	}
	if p.DefaultEnv != "staging" {
		t.Errorf("expected default env 'staging' (first sorted), got %q", p.DefaultEnv)
	}
}
