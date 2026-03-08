package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/dotenv"
	"github.com/tiaanduplessis/envy/internal/scan"
)

func NewScanCmd(store *config.Store) *cobra.Command {
	return newScanCmd(store, nil)
}

func newScanCmd(store *config.Store, stdin io.Reader) *cobra.Command {
	var dryRun bool
	var force bool
	var defaultEnv string

	cmd := &cobra.Command{
		Use:   "scan <project> [directory]",
		Short: "Create a project by scanning a directory for .env files",
		Example: `  envy scan my-app .
  envy scan my-app ~/code/my-monorepo
  envy scan my-app . --dry-run`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			scanRoot := "."
			if len(args) == 2 {
				scanRoot = args[1]
			}
			absRoot, err := filepath.Abs(scanRoot)
			if err != nil {
				return fmt.Errorf("resolving path: %w", err)
			}

			info, err := os.Stat(absRoot)
			if err != nil {
				return fmt.Errorf("cannot access %q: %w", absRoot, err)
			}
			if !info.IsDir() {
				return fmt.Errorf("%q is not a directory", absRoot)
			}

			if !force && store.Exists(name) {
				return fmt.Errorf("project %q already exists (use --force to overwrite)", name)
			}

			result, err := scan.Dir(absRoot)
			if err != nil {
				return err
			}

			if len(result.Root) == 0 && len(result.Paths) == 0 {
				return fmt.Errorf("no .env files found in %s", absRoot)
			}

			for _, w := range result.Warnings {
				fmt.Fprintf(cmd.OutOrStdout(), "Warning: %s\n", w)
			}

			printScanSummary(cmd.OutOrStdout(), absRoot, result)

			if dryRun {
				return nil
			}

			if !force {
				reader := stdin
				if reader == nil {
					reader = os.Stdin
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\nCreate project %q? [y/N] ", name)
				scanner := bufio.NewScanner(reader)
				scanner.Scan()
				answer := scanner.Text()
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			envNames := collectEnvNames(result)
			if defaultEnv == "" {
				if slices.Contains(envNames, "dev") {
					defaultEnv = "dev"
				} else {
					defaultEnv = envNames[0]
				}
			}

			project, err := config.NewProject(name, envNames, defaultEnv)
			if err != nil {
				return err
			}

			varCount, err := populateProject(project, result)
			if err != nil {
				return err
			}

			for env, filePath := range result.Root {
				filename := filepath.Base(filePath)
				if filename != ".env" {
					project.SetEnvFile(env, filename)
				}
			}

			if err := store.Save(project); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created project %q (%d variable(s) across %d environment(s))\n",
				name, varCount, len(envNames))
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview discovered files without creating project")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing project, skip confirmation")
	cmd.Flags().StringVar(&defaultEnv, "default-env", "", "default environment")

	return cmd
}

func printScanSummary(w io.Writer, root string, result *scan.Result) {
	fmt.Fprintf(w, "Scanning %s\n\n", root)

	if len(result.Root) > 0 {
		envs := sortedKeys(result.Root)
		fmt.Fprintf(w, "Root:\n")
		for _, env := range envs {
			rel, _ := filepath.Rel(root, result.Root[env])
			fmt.Fprintf(w, "  %-12s <- %s\n", env, rel)
		}
	}

	if len(result.Paths) > 0 {
		paths := sortedKeys(result.Paths)
		for _, path := range paths {
			envFiles := result.Paths[path]
			envs := sortedKeys(envFiles)
			fmt.Fprintf(w, "\n%s:\n", path)
			for _, env := range envs {
				rel, _ := filepath.Rel(root, envFiles[env])
				fmt.Fprintf(w, "  %-12s <- %s\n", env, rel)
			}
		}
	}
}

func collectEnvNames(result *scan.Result) []string {
	seen := make(map[string]bool)
	for env := range result.Root {
		seen[env] = true
	}
	for _, envFiles := range result.Paths {
		for env := range envFiles {
			seen[env] = true
		}
	}

	names := make([]string, 0, len(seen))
	for env := range seen {
		names = append(names, env)
	}
	slices.Sort(names)
	return names
}

func populateProject(project *config.Project, result *scan.Result) (int, error) {
	varCount := 0

	for env, filePath := range result.Root {
		parsed, err := parseEnvFile(filePath)
		if err != nil {
			return 0, err
		}
		for key, value := range parsed.Vars {
			project.SetVar(env, key, value)
			varCount++
		}
		for key, value := range parsed.DisabledVars {
			project.SetDisabledVar(env, key, value)
			varCount++
		}
	}

	for path, envFiles := range result.Paths {
		for env, filePath := range envFiles {
			parsed, err := parseEnvFile(filePath)
			if err != nil {
				return 0, err
			}
			for key, value := range parsed.Vars {
				project.SetPathVar(path, env, key, value)
				varCount++
			}
			for key, value := range parsed.DisabledVars {
				project.SetDisabledPathVar(path, env, key, value)
				varCount++
			}
		}
	}

	return varCount, nil
}

func parseEnvFile(path string) (*dotenv.ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %q: %w", path, err)
	}
	defer f.Close()

	vars, err := dotenv.ParseWithDisabled(f)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %w", path, err)
	}
	return vars, nil
}
