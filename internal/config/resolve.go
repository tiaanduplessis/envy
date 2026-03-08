package config

import (
	"fmt"
	"os"
)

// ResolveEnv determines which environment to use based on priority:
// 1. Explicit flag value
// 2. ENVY_ENV environment variable
// 3. Project default_env
// 4. "dev" as final fallback
func ResolveEnv(flagValue, defaultEnv string) string {
	if flagValue != "" {
		return flagValue
	}
	if envVar := os.Getenv("ENVY_ENV"); envVar != "" {
		return envVar
	}
	if defaultEnv != "" {
		return defaultEnv
	}
	return "dev"
}

// ResolveOutputFile determines the output filename for a load operation.
// Priority: explicit --output flag > per-environment mapping > ".env" default.
func ResolveOutputFile(p *Project, env, flagValue string, flagChanged bool) string {
	if flagChanged {
		return flagValue
	}
	if p.EnvFiles != nil {
		if filename, ok := p.EnvFiles[env]; ok {
			return filename
		}
	}
	return ".env"
}

// ResolveVars returns the merged variable map for a given environment and optional path.
// If path is empty, returns only root-level environment variables.
// If the path has no overrides, returns root-level variables without error.
// Returns an error if the environment does not exist.
func ResolveVars(p *Project, env, path string) (map[string]string, error) {
	base, ok := p.Environments[env]
	if !ok {
		return nil, fmt.Errorf("environment %q not found in project %q", env, p.Name)
	}

	result := make(map[string]string, len(base))
	for k, v := range base {
		result[k] = v
	}

	if path == "" {
		return result, nil
	}

	overrides := p.GetPathVars(path, env)
	for k, v := range overrides {
		result[k] = v
	}

	return result, nil
}

// ResolveDisabledVars returns the merged disabled variable map for a given environment and optional path.
// Root disabled variables are inherited into path-level outputs and path-level disabled variables override
// root disabled values when the same key is present.
func ResolveDisabledVars(p *Project, env, path string) (map[string]string, error) {
	if _, ok := p.Environments[env]; !ok {
		return nil, fmt.Errorf("environment %q not found in project %q", env, p.Name)
	}

	base := p.GetDisabledVars(env)
	result := make(map[string]string, len(base))
	for k, v := range base {
		result[k] = v
	}

	if path == "" {
		return result, nil
	}

	overrides := p.GetDisabledPathVars(path, env)
	for k, v := range overrides {
		result[k] = v
	}

	return result, nil
}
