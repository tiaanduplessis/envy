package scan

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnvName(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{".env", "dev"},
		{".env.local", "local"},
		{".env.dev", "dev"},
		{".env.development", "dev"},
		{".env.staging", "staging"},
		{".env.stage", "staging"},
		{".env.prod", "prod"},
		{".env.production", "prod"},
		{".env.test", "test"},
		{".env.testing", "test"},
		{".env.custom", "custom"},
		{".env.ci", "ci"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := envName(tt.filename)
			if got != tt.want {
				t.Errorf("envName(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsEnvFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{".env", true},
		{".env.local", true},
		{".env.prod", true},
		{"README.md", false},
		{"env", false},
		{".environment", false},
		{".envrc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEnvFile(tt.name)
			if got != tt.want {
				t.Errorf("isEnvFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestDir_RootOnly(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env"), "DB_URL=localhost\n")
	writeFile(t, filepath.Join(root, ".env.staging"), "DB_URL=staging-db\n")
	writeFile(t, filepath.Join(root, ".env.prod"), "DB_URL=prod-db\n")

	result, err := Dir(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Root) != 3 {
		t.Fatalf("expected 3 root envs, got %d", len(result.Root))
	}

	for _, env := range []string{"dev", "staging", "prod"} {
		if _, ok := result.Root[env]; !ok {
			t.Errorf("missing root env %q", env)
		}
	}

	if len(result.Paths) != 0 {
		t.Errorf("expected no paths, got %d", len(result.Paths))
	}
}

func TestDir_Monorepo(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env"), "SHARED=true\n")
	writeFile(t, filepath.Join(root, "services", "api", ".env"), "PORT=3000\n")
	writeFile(t, filepath.Join(root, "services", "api", ".env.prod"), "PORT=8080\n")
	writeFile(t, filepath.Join(root, "services", "worker", ".env"), "CONCURRENCY=4\n")

	result, err := Dir(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Root) != 1 {
		t.Fatalf("expected 1 root env, got %d", len(result.Root))
	}
	if _, ok := result.Root["dev"]; !ok {
		t.Error("missing root dev env")
	}

	apiPath := filepath.Join("services", "api")
	workerPath := filepath.Join("services", "worker")

	if len(result.Paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(result.Paths))
	}

	apiEnvs := result.Paths[apiPath]
	if len(apiEnvs) != 2 {
		t.Fatalf("expected 2 envs for api, got %d", len(apiEnvs))
	}
	if _, ok := apiEnvs["dev"]; !ok {
		t.Error("missing api dev env")
	}
	if _, ok := apiEnvs["prod"]; !ok {
		t.Error("missing api prod env")
	}

	workerEnvs := result.Paths[workerPath]
	if len(workerEnvs) != 1 {
		t.Fatalf("expected 1 env for worker, got %d", len(workerEnvs))
	}
}

func TestDir_SkipsDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env"), "ROOT=true\n")
	writeFile(t, filepath.Join(root, "node_modules", "pkg", ".env"), "SKIP=true\n")
	writeFile(t, filepath.Join(root, ".git", ".env"), "SKIP=true\n")
	writeFile(t, filepath.Join(root, "vendor", ".env"), "SKIP=true\n")

	result, err := Dir(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Root) != 1 {
		t.Fatalf("expected 1 root env, got %d", len(result.Root))
	}
	if len(result.Paths) != 0 {
		t.Errorf("expected no paths (skipped dirs), got %d", len(result.Paths))
	}
}

func TestDir_Empty(t *testing.T) {
	root := t.TempDir()

	result, err := Dir(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Root) != 0 {
		t.Errorf("expected 0 root envs, got %d", len(result.Root))
	}
	if len(result.Paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(result.Paths))
	}
}

func TestDir_ConflictPrefersSpecific(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env"), "SOURCE=bare\n")
	writeFile(t, filepath.Join(root, ".env.dev"), "SOURCE=specific\n")

	result, err := Dir(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Root) != 1 {
		t.Fatalf("expected 1 root env (conflict resolved), got %d", len(result.Root))
	}

	devPath := result.Root["dev"]
	if filepath.Base(devPath) != ".env.dev" {
		t.Errorf("expected .env.dev to win conflict, got %s", filepath.Base(devPath))
	}

	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestDir_NonEnvFilesIgnored(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env"), "A=1\n")
	writeFile(t, filepath.Join(root, "README.md"), "# Hello\n")
	writeFile(t, filepath.Join(root, ".envrc"), "layout go\n")
	writeFile(t, filepath.Join(root, "config.yaml"), "key: value\n")

	result, err := Dir(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Root) != 1 {
		t.Fatalf("expected 1 root env, got %d", len(result.Root))
	}
}

func TestSkipDir(t *testing.T) {
	skipped := []string{".git", "node_modules", "vendor", "dist", "build", ".next", ".venv"}
	for _, name := range skipped {
		if !skipDir(name) {
			t.Errorf("expected %q to be skipped", name)
		}
	}

	notSkipped := []string{"src", "lib", "services", "packages", "apps"}
	for _, name := range notSkipped {
		if skipDir(name) {
			t.Errorf("expected %q to NOT be skipped", name)
		}
	}
}

func TestResolveConflict(t *testing.T) {
	t.Run("single entry", func(t *testing.T) {
		entries := []candidate{{filename: ".env", absPath: "/a/.env"}}
		got := resolveConflict(entries)
		if got.filename != ".env" {
			t.Errorf("expected .env, got %s", got.filename)
		}
	})

	t.Run("bare vs suffixed", func(t *testing.T) {
		entries := []candidate{
			{filename: ".env", absPath: "/a/.env"},
			{filename: ".env.dev", absPath: "/a/.env.dev"},
		}
		got := resolveConflict(entries)
		if got.filename != ".env.dev" {
			t.Errorf("expected .env.dev, got %s", got.filename)
		}
	})

	t.Run("multiple suffixed picks last alphabetically", func(t *testing.T) {
		entries := []candidate{
			{filename: ".env.development", absPath: "/a/.env.development"},
			{filename: ".env.dev", absPath: "/a/.env.dev"},
		}
		got := resolveConflict(entries)
		if got.filename != ".env.development" {
			t.Errorf("expected .env.development, got %s", got.filename)
		}
	})
}

func TestDir_DeepNesting(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a", "b", "c", ".env"), "DEEP=true\n")

	result, err := Dir(root)
	if err != nil {
		t.Fatal(err)
	}

	deepPath := filepath.Join("a", "b", "c")
	if _, ok := result.Paths[deepPath]; !ok {
		keys := make([]string, 0, len(result.Paths))
		for k := range result.Paths {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		t.Errorf("expected path %q, got paths: %v", deepPath, keys)
	}
}
