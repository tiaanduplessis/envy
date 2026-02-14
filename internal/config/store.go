package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tiaanduplessis/envy/internal/crypto"
	"gopkg.in/yaml.v3"
)

// Store manages project YAML files on disk.
type Store struct {
	basePath         string
	passphraseFunc   func(string) (string, error)
	cachedPassphrase string
}

func NewStore(basePath string) *Store {
	return &Store{basePath: basePath}
}

// SetPassphraseFunc sets the callback used to obtain a passphrase for
// encrypted projects. In production this is crypto.GetPassphrase;
// in tests it can return a fixed string.
func (s *Store) SetPassphraseFunc(fn func(string) (string, error)) {
	s.passphraseFunc = fn
}

// Save writes a project to disk as a YAML file. Creates directories as needed.
// If the project is encrypted, values are encrypted before writing; the
// caller's in-memory Project stays plaintext.
func (s *Store) Save(p *Project) error {
	if err := p.Validate(); err != nil {
		return err
	}

	toWrite := p
	if p.IsEncrypted() {
		clone, err := s.encryptProject(p)
		if err != nil {
			return fmt.Errorf("encrypting project %q: %w", p.Name, err)
		}
		toWrite = clone
	}

	return s.writeProject(toWrite)
}

// Load reads a project from disk by name. If the project is encrypted,
// values are decrypted transparently.
func (s *Store) Load(name string) (*Project, error) {
	p, err := s.loadFromDisk(name)
	if err != nil {
		return nil, err
	}

	if p.IsEncrypted() {
		if err := s.decryptProject(p); err != nil {
			return nil, fmt.Errorf("decrypting project %q: %w", name, err)
		}
	}

	return p, nil
}

// LoadRaw reads a project from disk without decrypting values.
// Used by encrypt/decrypt/rekey commands that handle encryption themselves.
func (s *Store) LoadRaw(name string) (*Project, error) {
	return s.loadFromDisk(name)
}

// SaveRaw writes a project to disk without encrypting values.
// Used by encrypt/decrypt/rekey commands that handle encryption themselves.
func (s *Store) SaveRaw(p *Project) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return s.writeProject(p)
}

// ResetPassphraseCache clears any cached passphrase. Useful between
// operations that need different passphrases (e.g. rekey).
func (s *Store) ResetPassphraseCache() {
	s.cachedPassphrase = ""
}

func (s *Store) loadFromDisk(name string) (*Project, error) {
	data, err := os.ReadFile(s.ProjectPath(name))
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

func (s *Store) writeProject(p *Project) error {
	if err := os.MkdirAll(s.basePath, 0o755); err != nil {
		return fmt.Errorf("creating projects directory: %w", err)
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshalling project %q: %w", p.Name, err)
	}

	path := s.ProjectPath(p.Name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing project %q: %w", p.Name, err)
	}
	return nil
}

func (s *Store) getPassphrase(prompt string) (string, error) {
	if s.cachedPassphrase != "" {
		return s.cachedPassphrase, nil
	}
	if s.passphraseFunc == nil {
		return "", fmt.Errorf("project is encrypted but no passphrase provider configured")
	}
	pass, err := s.passphraseFunc(prompt)
	if err != nil {
		return "", err
	}
	s.cachedPassphrase = pass
	return pass, nil
}

func (s *Store) deriveKey(enc *EncryptionConfig) ([]byte, error) {
	passphrase, err := s.getPassphrase("Enter passphrase: ")
	if err != nil {
		return nil, err
	}
	salt, err := base64.StdEncoding.DecodeString(enc.Salt)
	if err != nil {
		return nil, fmt.Errorf("decoding salt: %w", err)
	}
	return crypto.DeriveKey(passphrase, salt, enc.Params), nil
}

func (s *Store) decryptProject(p *Project) error {
	key, err := s.deriveKey(p.Encryption)
	if err != nil {
		return err
	}

	for env, vars := range p.Environments {
		decrypted, err := crypto.DecryptMap(key, vars)
		if err != nil {
			return fmt.Errorf("environment %q: %w", env, err)
		}
		p.Environments[env] = decrypted
	}

	for path, envMap := range p.Paths {
		for env, vars := range envMap {
			decrypted, err := crypto.DecryptMap(key, vars)
			if err != nil {
				return fmt.Errorf("path %q env %q: %w", path, env, err)
			}
			p.Paths[path][env] = decrypted
		}
	}

	return nil
}

func (s *Store) encryptProject(p *Project) (*Project, error) {
	key, err := s.deriveKey(p.Encryption)
	if err != nil {
		return nil, err
	}

	clone := *p
	if p.Encryption != nil {
		enc := *p.Encryption
		clone.Encryption = &enc
	}
	if p.EnvFiles != nil {
		clone.EnvFiles = make(map[string]string, len(p.EnvFiles))
		for k, v := range p.EnvFiles {
			clone.EnvFiles[k] = v
		}
	}
	clone.Environments = make(map[string]map[string]string, len(p.Environments))
	for env, vars := range p.Environments {
		encrypted, err := crypto.EncryptMap(key, vars)
		if err != nil {
			return nil, fmt.Errorf("environment %q: %w", env, err)
		}
		clone.Environments[env] = encrypted
	}

	clone.Paths = make(map[string]map[string]map[string]string, len(p.Paths))
	for path, envMap := range p.Paths {
		clone.Paths[path] = make(map[string]map[string]string, len(envMap))
		for env, vars := range envMap {
			encrypted, err := crypto.EncryptMap(key, vars)
			if err != nil {
				return nil, fmt.Errorf("path %q env %q: %w", path, env, err)
			}
			clone.Paths[path][env] = encrypted
		}
	}

	return &clone, nil
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

func (s *Store) Delete(name string) error {
	path := s.ProjectPath(name)
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("project %q not found", name)
		}
		return fmt.Errorf("deleting project %q: %w", name, err)
	}
	return nil
}

func (s *Store) Exists(name string) bool {
	_, err := os.Stat(s.ProjectPath(name))
	return err == nil
}

func (s *Store) ProjectPath(name string) string {
	return filepath.Join(s.basePath, name+".yaml")
}
