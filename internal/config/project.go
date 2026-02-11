package config

import (
	"fmt"
	"regexp"
	"time"

	"github.com/tiaanduplessis/envy/internal/crypto"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// EncryptionConfig holds per-project encryption metadata.
type EncryptionConfig struct {
	Enabled bool          `yaml:"enabled"`
	Salt    string        `yaml:"salt"`
	Params  crypto.Params `yaml:"params"`
}

type Project struct {
	Name         string                                   `yaml:"name"`
	CreatedAt    time.Time                                `yaml:"created_at"`
	UpdatedAt    time.Time                                `yaml:"updated_at"`
	DefaultEnv   string                                   `yaml:"default_env"`
	Encryption   *EncryptionConfig                        `yaml:"encryption,omitempty"`
	EnvFiles     map[string]string                        `yaml:"env_files,omitempty"`
	Environments map[string]map[string]string             `yaml:"environments,omitempty"`
	Paths        map[string]map[string]map[string]string  `yaml:"paths,omitempty"`
}

func (p *Project) IsEncrypted() bool {
	return p.Encryption != nil && p.Encryption.Enabled
}

// NewProject creates a new project with the given name, environments, and default env.
// If no environments are provided, "dev" is used.
func NewProject(name string, envs []string, defaultEnv string) (*Project, error) {
	if !validName.MatchString(name) {
		return nil, fmt.Errorf("invalid project name %q: must start with alphanumeric and contain only alphanumeric, hyphens, or underscores", name)
	}

	if len(envs) == 0 {
		envs = []string{"dev"}
	}
	if defaultEnv == "" {
		defaultEnv = envs[0]
	}

	now := time.Now().UTC()
	p := &Project{
		Name:         name,
		CreatedAt:    now,
		UpdatedAt:    now,
		DefaultEnv:   defaultEnv,
		Environments: make(map[string]map[string]string),
	}

	for _, env := range envs {
		p.Environments[env] = make(map[string]string)
	}

	return p, nil
}

// Validate checks that the project has a valid name and default environment.
func (p *Project) Validate() error {
	if !validName.MatchString(p.Name) {
		return fmt.Errorf("invalid project name %q", p.Name)
	}
	if p.DefaultEnv == "" {
		return fmt.Errorf("project %q has no default_env", p.Name)
	}
	return nil
}

// SetVar sets a variable in the given environment at root level.
func (p *Project) SetVar(env, key, value string) {
	if p.Environments == nil {
		p.Environments = make(map[string]map[string]string)
	}
	if p.Environments[env] == nil {
		p.Environments[env] = make(map[string]string)
	}
	p.Environments[env][key] = value
	p.UpdatedAt = time.Now().UTC()
}

func (p *Project) SetPathVar(path, env, key, value string) {
	if p.Paths == nil {
		p.Paths = make(map[string]map[string]map[string]string)
	}
	if p.Paths[path] == nil {
		p.Paths[path] = make(map[string]map[string]string)
	}
	if p.Paths[path][env] == nil {
		p.Paths[path][env] = make(map[string]string)
	}
	p.Paths[path][env][key] = value
	p.UpdatedAt = time.Now().UTC()
}

// GetPathVars returns the raw (non-inherited) variables for a path and environment.
func (p *Project) GetPathVars(path, env string) map[string]string {
	if p.Paths == nil {
		return nil
	}
	if p.Paths[path] == nil {
		return nil
	}
	return p.Paths[path][env]
}

func (p *Project) SetEnvFile(env, filename string) {
	if p.EnvFiles == nil {
		p.EnvFiles = make(map[string]string)
	}
	p.EnvFiles[env] = filename
	p.UpdatedAt = time.Now().UTC()
}

func (p *Project) ClearEnvFile(env string) {
	if p.EnvFiles == nil {
		return
	}
	delete(p.EnvFiles, env)
	p.UpdatedAt = time.Now().UTC()
}
