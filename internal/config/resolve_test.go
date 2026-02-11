package config

import "testing"

func TestResolveEnv(t *testing.T) {
	tests := []struct {
		name       string
		flag       string
		envVar     string
		defaultEnv string
		want       string
	}{
		{"flag takes priority", "staging", "prod", "dev", "staging"},
		{"env var second", "", "prod", "dev", "prod"},
		{"default third", "", "", "staging", "staging"},
		{"fallback to dev", "", "", "", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENVY_ENV", tt.envVar)
			got := ResolveEnv(tt.flag, tt.defaultEnv)
			if got != tt.want {
				t.Errorf("ResolveEnv(%q, %q) = %q, want %q",
					tt.flag, tt.defaultEnv, got, tt.want)
			}
		})
	}
}

func TestResolveOutputFile(t *testing.T) {
	tests := []struct {
		name        string
		envFiles    map[string]string
		env         string
		flagValue   string
		flagChanged bool
		want        string
	}{
		{
			"explicit flag overrides everything",
			map[string]string{"local": ".env.local"},
			"local", "custom.env", true,
			"custom.env",
		},
		{
			"env file mapping used when flag not set",
			map[string]string{"local": ".env.local"},
			"local", ".env", false,
			".env.local",
		},
		{
			"default when no mapping and no flag",
			nil,
			"dev", ".env", false,
			".env",
		},
		{
			"default when env has no mapping",
			map[string]string{"local": ".env.local"},
			"dev", ".env", false,
			".env",
		},
		{
			"flag explicitly set to .env overrides mapping",
			map[string]string{"local": ".env.local"},
			"local", ".env", true,
			".env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := NewProject("test", []string{"dev", "local"}, "dev")
			p.EnvFiles = tt.envFiles

			got := ResolveOutputFile(p, tt.env, tt.flagValue, tt.flagChanged)
			if got != tt.want {
				t.Errorf("ResolveOutputFile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveVars(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Project
		env     string
		path    string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "root only",
			setup: func() *Project {
				p, _ := NewProject("test", []string{"dev"}, "dev")
				p.SetVar("dev", "DB", "localhost")
				p.SetVar("dev", "PORT", "5432")
				return p
			},
			env:  "dev",
			path: "",
			want: map[string]string{"DB": "localhost", "PORT": "5432"},
		},
		{
			name: "path overrides root",
			setup: func() *Project {
				p, _ := NewProject("test", []string{"dev"}, "dev")
				p.SetVar("dev", "DB", "localhost")
				p.SetVar("dev", "PORT", "5432")
				p.SetPathVar("services/api", "dev", "DB", "api-db")
				return p
			},
			env:  "dev",
			path: "services/api",
			want: map[string]string{"DB": "api-db", "PORT": "5432"},
		},
		{
			name: "path adds new keys",
			setup: func() *Project {
				p, _ := NewProject("test", []string{"dev"}, "dev")
				p.SetVar("dev", "DB", "localhost")
				p.SetPathVar("services/api", "dev", "SERVICE", "api")
				return p
			},
			env:  "dev",
			path: "services/api",
			want: map[string]string{"DB": "localhost", "SERVICE": "api"},
		},
		{
			name: "missing env returns error",
			setup: func() *Project {
				p, _ := NewProject("test", []string{"dev"}, "dev")
				return p
			},
			env:     "staging",
			path:    "",
			wantErr: true,
		},
		{
			name: "missing path falls back to root",
			setup: func() *Project {
				p, _ := NewProject("test", []string{"dev"}, "dev")
				p.SetVar("dev", "DB", "localhost")
				return p
			},
			env:  "dev",
			path: "nonexistent",
			want: map[string]string{"DB": "localhost"},
		},
		{
			name: "empty env map",
			setup: func() *Project {
				p, _ := NewProject("test", []string{"dev"}, "dev")
				return p
			},
			env:  "dev",
			path: "",
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			got, err := ResolveVars(p, tt.env, tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveVars() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d vars, want %d: %v", len(got), len(tt.want), got)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %q = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
