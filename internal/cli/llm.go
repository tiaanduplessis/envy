package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func generateLLMOutput(root *cobra.Command, w io.Writer) {
	writeOverview(w)
	writeEnvironmentVariables(w)
	writeKeyConcepts(w)
	writeCommands(w, root)
	writeWorkflows(w)
}

func writeOverview(w io.Writer) {
	fmt.Fprintln(w, `# envy

Envy is a local-first CLI for managing .env files across projects from a centralised YAML config store.

All project configuration is stored in YAML files under ~/.config/envy/projects/ (configurable via ENVY_CONFIG_DIR). Each project has its own YAML file containing environment variables organised by environment name (e.g. dev, staging, production) and optional monorepo subpath overrides.

Envy does not require a server, database, or network access. It is a single binary that reads and writes local YAML files and generates .env files from them.`)
	fmt.Fprintln(w)
}

func writeEnvironmentVariables(w io.Writer) {
	fmt.Fprintln(w, `## Environment Variables

- ENVY_CONFIG_DIR: Override the default config directory (~/.config/envy). When set, envy stores and reads all project YAML files from this directory instead.
- ENVY_ENV: Set the default environment for commands that accept --env. This is overridden by the --env flag and by a project's default_env setting. Resolution order: --env flag > ENVY_ENV > project default_env > "dev".
- ENVY_PASSPHRASE: Provide the encryption passphrase non-interactively. Used by encrypt, decrypt, rekey, and any command that loads an encrypted project. When unset, envy prompts on the terminal.`)
	fmt.Fprintln(w)
}

func writeKeyConcepts(w io.Writer) {
	fmt.Fprintln(w, `## Key Concepts

### Projects
A project is a named collection of environment variables stored as a YAML file. Each project maps to one application or service. Projects are created with "envy init" or "envy scan" and stored at ~/.config/envy/projects/<name>.yaml.

### Environments
Each project contains one or more environments (e.g. dev, staging, production). An environment is a named set of key-value pairs. The default environment is "dev" unless overridden by --default-env during init or by ENVY_ENV.

### Paths (Monorepo Support)
For monorepos, a project can have subpath overrides. A path represents a subdirectory that needs its own .env file. Path-level variables are merged on top of root-level environment variables at resolution time, so shared variables live at the root and path-specific overrides live under the path.

### Variable Resolution
When resolving variables for a given environment and path:
1. Start with the root-level variables for that environment.
2. Merge path-level overrides on top (path values win on conflict).
This merge happens at resolution time (e.g. during "envy load" or "envy get"), not at storage time. The store keeps root and path variables separate.

### Encryption
Projects can optionally be encrypted with AES-256-GCM. When enabled, all variable values are encrypted at rest in the YAML file (stored as ENC:<base64>). Encryption is transparent: "envy load", "envy show", and other read commands decrypt automatically when given the passphrase. Key derivation uses Argon2id with a per-project salt.

### Env Files
Each environment can optionally have a custom output filename (instead of the default .env). This is useful when environments map to files like .env.local or .env.production. Set with "envy env file set" and used by "envy load" to determine the output path.`)
	fmt.Fprintln(w)
}

func writeCommands(w io.Writer, root *cobra.Command) {
	fmt.Fprintln(w, "## Commands")
	fmt.Fprintln(w)
	writeCommandTree(w, root, nil)
}

func writeCommandTree(w io.Writer, cmd *cobra.Command, parentPath []string) {
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" || child.Name() == "completion" {
			continue
		}

		path := append(parentPath, child.Name())
		fullPath := strings.Join(append([]string{cmd.Root().Name()}, path...), " ")

		fmt.Fprintf(w, "### %s\n\n", fullPath)

		desc := child.Long
		if desc == "" {
			desc = child.Short
		}
		if desc != "" {
			fmt.Fprintln(w, desc)
			fmt.Fprintln(w)
		}

		fmt.Fprintf(w, "Usage: %s\n", child.UseLine())
		fmt.Fprintln(w)

		if child.HasSubCommands() {
			writeCommandTree(w, child, path)
			continue
		}

		flags := collectFlags(child)
		if len(flags) > 0 {
			fmt.Fprintln(w, "Flags:")
			for _, f := range flags {
				fmt.Fprintf(w, "  --%s", f.name)
				if f.shorthand != "" {
					fmt.Fprintf(w, ", -%s", f.shorthand)
				}
				fmt.Fprintf(w, " (%s)", f.typeName)
				if f.defValue != "" && f.defValue != "false" && f.defValue != "[]" {
					fmt.Fprintf(w, " [default: %s]", f.defValue)
				}
				if f.required {
					fmt.Fprint(w, " [required]")
				}
				fmt.Fprintf(w, ": %s\n", f.usage)
			}
			fmt.Fprintln(w)
		}
	}
}

type flagInfo struct {
	name      string
	shorthand string
	typeName  string
	defValue  string
	usage     string
	required  bool
}

func collectFlags(cmd *cobra.Command) []flagInfo {
	var flags []flagInfo
	required := make(map[string]bool)
	if ann := cmd.Annotations; ann != nil {
		// Cobra stores required flags in annotations on the command
	}
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		flags = append(flags, flagInfo{
			name:      f.Name,
			shorthand: f.Shorthand,
			typeName:  f.Value.Type(),
			defValue:  f.DefValue,
			usage:     f.Usage,
			required:  required[f.Name],
		})
	})

	// Check required annotations from Cobra
	for i := range flags {
		annotations := cmd.Flags().Lookup(flags[i].name).Annotations
		if annotations != nil {
			if _, ok := annotations[cobra.BashCompOneRequiredFlag]; ok {
				flags[i].required = true
			}
		}
	}

	return flags
}

func writeWorkflows(w io.Writer) {
	fmt.Fprintln(w, `## Common Workflows

### Create a project and load a .env file
  envy init myapp --env dev,staging
  envy set myapp DB_HOST=localhost DB_PORT=5432 --env dev
  envy set myapp DB_HOST=db.staging.example.com DB_PORT=5432 --env staging
  envy load --project myapp --env dev

### Import an existing .env file
  envy init myapp
  envy update --project myapp --file .env --env dev

### Scan a directory to auto-discover .env files
  envy scan myapp ./path/to/project
  envy load --project myapp

### Compare environments
  envy diff myapp --env dev --env staging

### Compare local .env against stored config
  envy diff myapp --env dev

### Set up monorepo paths
  envy init myapp --env dev --path services/api --path services/web
  envy set myapp API_KEY=xxx --env dev --path services/api
  envy load --project myapp --env dev --path services/api
  envy load --project myapp --all-paths

### Enable encryption
  envy encrypt myapp
  envy load --project myapp

### Change encryption passphrase
  envy rekey myapp

### Set custom output filename for an environment
  envy env file set myapp production .env.production
  envy load --project myapp --env production`)
	fmt.Fprintln(w)
}
