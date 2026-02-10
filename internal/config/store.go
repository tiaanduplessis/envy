package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Store manages project YAML files on disk.
type Store struct {
	basePath string
}

// NewStore creates a Store backed by the given directory.
func NewStore(basePath string) *Store {
	return &Store{basePath: basePath}
}

// Save writes a project to disk as a YAML file. Creates directories as needed.
func (s *Store) Save(p *Project) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(s.basePath, 0o755); err != nil {
		return fmt.Errorf("creating projects directory: %w", err)
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshalling project %q: %w", p.Name, err)
	}

	path := s.projectPath(p.Name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing project %q: %w", p.Name, err)
	}
	return nil
}

// Load reads a project from disk by name.
func (s *Store) Load(name string) (*Project, error) {
	data, err := os.ReadFile(s.projectPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project %q not found", name)
		}
		return nil, fmt.Errorf("reading project %q: %w", name, err)
	}

	var p Project
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing project %q: %w", name, err)
	}
	return &p, nil
}

// List returns the names of all stored projects, sorted alphabetically.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") {
			names = append(names, strings.TrimSuffix(name, ".yaml"))
		}
	}
	sort.Strings(names)
	return names, nil
}

// Delete removes a project file from disk.
func (s *Store) Delete(name string) error {
	path := s.projectPath(name)
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("project %q not found", name)
		}
		return fmt.Errorf("deleting project %q: %w", name, err)
	}
	return nil
}

// Exists checks whether a project file exists on disk.
func (s *Store) Exists(name string) bool {
	_, err := os.Stat(s.projectPath(name))
	return err == nil
}

func (s *Store) projectPath(name string) string {
	return filepath.Join(s.basePath, name+".yaml")
}
