package config

import "testing"

func TestNewProject_ValidNames(t *testing.T) {
	names := []string{"a", "foo", "my-app", "my_app", "abc123", "A1-b2_c3"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			p, err := NewProject(name, nil, "")
			if err != nil {
				t.Fatalf("NewProject(%q) returned error: %v", name, err)
			}
			if p.Name != name {
				t.Errorf("Name = %q, want %q", p.Name, name)
			}
			if p.DefaultEnv != "dev" {
				t.Errorf("DefaultEnv = %q, want %q", p.DefaultEnv, "dev")
			}
			if _, ok := p.Environments["dev"]; !ok {
				t.Error("expected 'dev' environment to exist")
			}
		})
	}
}

func TestNewProject_InvalidNames(t *testing.T) {
	names := []string{"", "../escape", "has space", ".hidden", "-starts-dash", "_starts-under"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			_, err := NewProject(name, nil, "")
			if err == nil {
				t.Errorf("NewProject(%q) expected error, got nil", name)
			}
		})
	}
}

func TestNewProject_WithEnvs(t *testing.T) {
	p, err := NewProject("test", []string{"dev", "staging", "prod"}, "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DefaultEnv != "staging" {
		t.Errorf("DefaultEnv = %q, want %q", p.DefaultEnv, "staging")
	}
	for _, env := range []string{"dev", "staging", "prod"} {
		if _, ok := p.Environments[env]; !ok {
			t.Errorf("expected %q environment to exist", env)
		}
	}
}

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project Project
		wantErr bool
	}{
		{
			"valid",
			Project{Name: "foo", DefaultEnv: "dev"},
			false,
		},
		{
			"invalid name",
			Project{Name: "../bad", DefaultEnv: "dev"},
			true,
		},
		{
			"empty default env",
			Project{Name: "foo", DefaultEnv: ""},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProject_SetVar(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	p.SetVar("dev", "DB_HOST", "localhost")

	if got := p.Environments["dev"]["DB_HOST"]; got != "localhost" {
		t.Errorf("got %q, want %q", got, "localhost")
	}

	// Auto-creates environment
	p.SetVar("staging", "DB_HOST", "staging-db")
	if got := p.Environments["staging"]["DB_HOST"]; got != "staging-db" {
		t.Errorf("got %q, want %q", got, "staging-db")
	}
}

func TestProject_SetPathVar(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	p.SetPathVar("services/api", "dev", "PORT", "3000")

	if got := p.Paths["services/api"]["dev"]["PORT"]; got != "3000" {
		t.Errorf("got %q, want %q", got, "3000")
	}
}

func TestProject_GetPathVars(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	p.SetPathVar("services/api", "dev", "PORT", "3000")

	vars := p.GetPathVars("services/api", "dev")
	if vars == nil {
		t.Fatal("expected non-nil vars")
	}
	if got := vars["PORT"]; got != "3000" {
		t.Errorf("got %q, want %q", got, "3000")
	}

	if vars := p.GetPathVars("nonexistent", "dev"); vars != nil {
		t.Errorf("expected nil for nonexistent path, got %v", vars)
	}

	if vars := p.GetPathVars("services/api", "prod"); vars != nil {
		t.Errorf("expected nil for nonexistent env, got %v", vars)
	}
}

func TestProject_SetEnvFile(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	p.SetEnvFile("local", ".env.local")

	if got := p.EnvFiles["local"]; got != ".env.local" {
		t.Errorf("got %q, want %q", got, ".env.local")
	}

	p.SetEnvFile("local", ".env.local.override")
	if got := p.EnvFiles["local"]; got != ".env.local.override" {
		t.Errorf("got %q, want %q", got, ".env.local.override")
	}
}

func TestProject_SetEnvFile_NilMap(t *testing.T) {
	p := &Project{Name: "test", DefaultEnv: "dev"}
	p.SetEnvFile("staging", ".env.staging")

	if got := p.EnvFiles["staging"]; got != ".env.staging" {
		t.Errorf("got %q, want %q", got, ".env.staging")
	}
}

func TestProject_ClearEnvFile(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	p.SetEnvFile("local", ".env.local")
	p.ClearEnvFile("local")

	if _, ok := p.EnvFiles["local"]; ok {
		t.Error("expected env file mapping to be cleared")
	}
}

func TestProject_ClearEnvFile_NilMap(t *testing.T) {
	p := &Project{Name: "test", DefaultEnv: "dev"}
	// Should not panic
	p.ClearEnvFile("local")
}

func TestProject_DeleteVar(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	p.SetVar("dev", "DB", "localhost")
	p.SetVar("dev", "PORT", "5432")

	if !p.DeleteVar("dev", "DB") {
		t.Error("expected DeleteVar to return true for existing key")
	}
	if _, ok := p.Environments["dev"]["DB"]; ok {
		t.Error("expected DB to be deleted")
	}
	if got := p.Environments["dev"]["PORT"]; got != "5432" {
		t.Errorf("PORT = %q, want %q", got, "5432")
	}
}

func TestProject_DeleteVar_NonexistentKey(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	if p.DeleteVar("dev", "NOPE") {
		t.Error("expected DeleteVar to return false for nonexistent key")
	}
}

func TestProject_DeleteVar_NilMaps(t *testing.T) {
	p := &Project{Name: "test", DefaultEnv: "dev"}
	if p.DeleteVar("dev", "KEY") {
		t.Error("expected DeleteVar to return false on nil environments")
	}
}

func TestProject_DeletePathVar(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	p.SetPathVar("services/api", "dev", "PORT", "3000")

	if !p.DeletePathVar("services/api", "dev", "PORT") {
		t.Error("expected DeletePathVar to return true for existing key")
	}
	if _, ok := p.Paths["services/api"]["dev"]["PORT"]; ok {
		t.Error("expected PORT to be deleted")
	}
}

func TestProject_DeletePathVar_NonexistentKey(t *testing.T) {
	p, _ := NewProject("test", nil, "")
	if p.DeletePathVar("services/api", "dev", "NOPE") {
		t.Error("expected DeletePathVar to return false for nonexistent key")
	}
}

func TestProject_DeletePathVar_NilMaps(t *testing.T) {
	p := &Project{Name: "test", DefaultEnv: "dev"}
	if p.DeletePathVar("services/api", "dev", "KEY") {
		t.Error("expected DeletePathVar to return false on nil paths")
	}
}
