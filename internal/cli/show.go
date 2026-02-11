package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/util"
)

type showOutput struct {
	Name         string                       `json:"name"`
	DefaultEnv   string                       `json:"default_env"`
	Encrypted    bool                         `json:"encrypted"`
	EnvFiles     map[string]string            `json:"env_files,omitempty"`
	Environments map[string]map[string]string `json:"environments,omitempty"`
	Paths        map[string]map[string]string `json:"paths,omitempty"`
}

func NewShowCmd(store *config.Store) *cobra.Command {
	var env string
	var path string
	var reveal bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <project>",
		Short: "Show a project's configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			out := cmd.OutOrStdout()

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			// Specific env + path: show resolved vars
			if env != "" || path != "" {
				targetEnv := config.ResolveEnv(env, p.DefaultEnv)
				vars, err := config.ResolveVars(p, targetEnv, path)
				if err != nil {
					return err
				}

				if jsonOutput {
					enc := json.NewEncoder(out)
					enc.SetIndent("", "  ")
					return enc.Encode(vars)
				}

				keys := sortedKeys(vars)
				for _, k := range keys {
					fmt.Fprintln(out, util.FormatKeyValue(k, vars[k], reveal))
				}
				return nil
			}

			// Full project overview
			if jsonOutput {
				output := showOutput{
					Name:         p.Name,
					DefaultEnv:   p.DefaultEnv,
					Encrypted:    p.IsEncrypted(),
					EnvFiles:     p.EnvFiles,
					Environments: p.Environments,
				}
				if path == "" && len(p.Paths) > 0 {
					output.Paths = make(map[string]map[string]string)
					for pth, envMap := range p.Paths {
						for e, vars := range envMap {
							key := pth + "/" + e
							output.Paths[key] = vars
						}
					}
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			fmt.Fprintf(out, "Project: %s\n", p.Name)
			fmt.Fprintf(out, "Default env: %s\n", p.DefaultEnv)

			if p.IsEncrypted() {
				fmt.Fprintln(out, "Encryption: enabled")
			}

			if len(p.EnvFiles) > 0 {
				fmt.Fprintln(out)
				fmt.Fprintln(out, "Env files:")
				for _, e := range sortedKeys(p.EnvFiles) {
					fmt.Fprintf(out, "  %-12s -> %s\n", e, p.EnvFiles[e])
				}
			}

			fmt.Fprintln(out)
			envNames := sortedKeys(p.Environments)
			for _, e := range envNames {
				vars := p.Environments[e]
				fmt.Fprintf(out, "[%s] (%d variables)\n", e, len(vars))
				keys := sortedKeys(vars)
				for _, k := range keys {
					fmt.Fprintf(out, "  %s\n", util.FormatKeyValue(k, vars[k], reveal))
				}
			}

			if len(p.Paths) > 0 {
				fmt.Fprintln(out)
				pathNames := sortedKeys(p.Paths)
				for _, pth := range pathNames {
					envMap := p.Paths[pth]
					fmt.Fprintf(out, "Path: %s\n", pth)
					pathEnvNames := sortedKeys(envMap)
					for _, e := range pathEnvNames {
						vars := envMap[e]
						fmt.Fprintf(out, "  [%s] (%d variables)\n", e, len(vars))
						keys := sortedKeys(vars)
						for _, k := range keys {
							fmt.Fprintf(out, "    %s\n", util.FormatKeyValue(k, vars[k], reveal))
						}
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&env, "env", "", "show specific environment")
	cmd.Flags().StringVar(&path, "path", "", "show resolved vars for a subpath")
	cmd.Flags().BoolVar(&reveal, "reveal", false, "show actual values instead of masked")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
