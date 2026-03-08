package scan

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// Result holds discovered .env files grouped by location and environment.
type Result struct {
	Root     map[string]string            // env name -> absolute file path
	Paths    map[string]map[string]string // relative dir -> env name -> absolute file path
	Warnings []string
}

// Dir walks root, discovers .env files, maps filenames to environments,
// resolves conflicts, and returns a Result.
func Dir(root string) (*Result, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	result := &Result{
		Root:  make(map[string]string),
		Paths: make(map[string]map[string]string),
	}

	// Track conflicts: for each directory, store all candidates per env name.
	// Key is relative dir ("" for root), value is env -> []candidate.
	candidates := make(map[string]map[string][]candidate)

	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path != abs && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isEnvFile(d.Name()) {
			return nil
		}

		rel, err := filepath.Rel(abs, filepath.Dir(path))
		if err != nil {
			return err
		}
		if rel == "." {
			rel = ""
		}

		env := envName(d.Name())

		if candidates[rel] == nil {
			candidates[rel] = make(map[string][]candidate)
		}
		candidates[rel][env] = append(candidates[rel][env], candidate{
			filename: d.Name(),
			absPath:  path,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	for dir, envCandidates := range candidates {
		for env, entries := range envCandidates {
			chosen := resolveConflict(entries)

			if len(entries) > 1 {
				names := make([]string, len(entries))
				for i, e := range entries {
					names[i] = e.filename
				}
				location := "root"
				if dir != "" {
					location = dir
				}
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("multiple files map to %q in %s: %s (using %s)",
						env, location, strings.Join(names, ", "), chosen.filename))
			}

			if dir == "" {
				result.Root[env] = chosen.absPath
			} else {
				if result.Paths[dir] == nil {
					result.Paths[dir] = make(map[string]string)
				}
				result.Paths[dir][env] = chosen.absPath
			}
		}
	}

	return result, nil
}

var ignoredEnvFiles = map[string]bool{
	".env.example":  true,
	".env.sample":   true,
	".env.template": true,
}

func isEnvFile(name string) bool {
	if ignoredEnvFiles[strings.ToLower(name)] {
		return false
	}
	return name == ".env" || strings.HasPrefix(name, ".env.")
}

var envAliases = map[string]string{
	".env":             "dev",
	".env.local":       "local",
	".env.dev":         "dev",
	".env.development": "dev",
	".env.staging":     "staging",
	".env.stage":       "staging",
	".env.prod":        "prod",
	".env.production":  "prod",
	".env.test":        "test",
	".env.testing":     "test",
}

func envName(filename string) string {
	lower := strings.ToLower(filename)
	if mapped, ok := envAliases[lower]; ok {
		return mapped
	}
	if suffix, ok := strings.CutPrefix(lower, ".env."); ok {
		return suffix
	}
	return "dev"
}

var skippedDirs = map[string]bool{
	".git": true, ".svn": true, ".hg": true,
	"node_modules": true, "vendor": true,
	"dist": true, "build": true, ".next": true,
	"target": true, "out": true, "bin": true,
	".venv": true, "venv": true,
	".terraform": true, ".cache": true,
	"tmp": true, "temp": true,
}

func skipDir(name string) bool {
	return skippedDirs[name]
}

type candidate struct {
	filename string
	absPath  string
}

// resolveConflict picks the most specific file when multiple files map to the
// same environment. A bare .env is always least specific; among suffixed files
// the last one alphabetically wins for determinism.
func resolveConflict(entries []candidate) candidate {
	if len(entries) == 1 {
		return entries[0]
	}

	// Prefer any suffixed file over bare .env
	var suffixed []candidate
	for _, e := range entries {
		if e.filename != ".env" {
			suffixed = append(suffixed, e)
		}
	}

	if len(suffixed) == 0 {
		return entries[0]
	}

	// Among suffixed files, pick the last alphabetically for determinism
	best := suffixed[0]
	for _, e := range suffixed[1:] {
		if e.filename > best.filename {
			best = e
		}
	}
	return best
}
