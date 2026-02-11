# envy

A local-first Go CLI for managing `.env` files across projects from a centralised YAML config store.

Envy stores project configurations as YAML files under `~/.config/envy/projects/`. Each project gets its own file. Variables are organised by named environments (e.g. `dev`, `staging`, `prod`) and optional monorepo subpaths. When you run `envy load`, it merges the appropriate variables and writes a `.env` file to your working directory.

## Installation

### Pre-built binaries

Download a pre-built binary for your platform from the [GitHub releases](https://github.com/tiaanduplessis/envy/releases) page. Builds are available for:

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64, arm64 |

After downloading, extract the archive and move the `envy` binary somewhere on your `PATH`:

```bash
# Example for macOS (Apple Silicon)
tar xzf envy_*_darwin_arm64.tar.gz
sudo mv envy /usr/local/bin/
```

### go install

Requires Go 1.25.7 or later.

```bash
go install github.com/tiaanduplessis/envy/cmd/envy@latest
```

This places the binary in your `$GOBIN` directory (usually `$HOME/go/bin`).

### Build from source

```bash
git clone https://github.com/tiaanduplessis/envy.git
cd envy
make build
# binary is at bin/envy
```

## Quick start

### Import existing .env files from a project

If you already have `.env` files in a repo, `scan` discovers them and stores everything in one go:

```bash
cd ~/code/my-monorepo
envy scan my-monorepo .
```

To replay those `.env` files into a fresh clone:

```bash
cd ~/code/my-monorepo-clone
envy load --project my-monorepo --all-paths --force
```

### Create a project from scratch

```bash
# Create a project with dev and staging environments
envy init my-app --env dev --env staging

# Add some variables
envy set my-app DB_HOST=localhost DB_PORT=5432 --env dev
envy set my-app DB_HOST=staging-db.example.com DB_PORT=5432 --env staging

# Write a .env file for the dev environment
envy load --project my-app --env dev

# Preview without writing a file
envy load --project my-app --env staging --dry-run
```

### Import a single .env file

```bash
envy init my-app
envy update --project my-app --file .env --env dev
```

## Commands

### envy init

Create a new project configuration.

```bash
envy init my-app
envy init my-app --env dev --env staging --env prod
envy init my-monorepo --path services/api --path services/worker
envy init my-app --default-env staging
envy init my-app --encrypt
```

| Flag | Description |
|------|-------------|
| `--env <name>` | Environments to create (repeatable). Defaults to `dev` if omitted. |
| `--path <subdir>` | Monorepo subpath stubs to create (repeatable). |
| `--default-env <name>` | Set the default environment. Defaults to the first `--env` value, or `dev`. |
| `--encrypt` | Enable encryption on the new project. Prompts for a passphrase. |

Fails if a project with the same name already exists.

### envy scan

Create a project by scanning a directory tree for `.env` files.

```bash
# Scan the current directory
envy scan my-app

# Scan a specific directory
envy scan my-monorepo ~/code/my-monorepo

# Preview what would be discovered
envy scan my-app . --dry-run

# Overwrite an existing project and skip confirmation
envy scan my-app . --force
```

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview discovered files without creating the project. |
| `--force` | Overwrite an existing project and skip the confirmation prompt. |
| `--default-env <name>` | Set the default environment. Falls back to `dev` if present, otherwise the first discovered environment. |

Scan maps filenames to environment names using these conventions:

| Filename | Environment |
|----------|-------------|
| `.env` | `dev` |
| `.env.local` | `local` |
| `.env.dev`, `.env.development` | `dev` |
| `.env.staging`, `.env.stage` | `staging` |
| `.env.prod`, `.env.production` | `prod` |
| `.env.test`, `.env.testing` | `test` |
| `.env.<other>` | `<other>` |

Directories like `.git`, `node_modules`, `vendor`, `dist`, `build`, `.next`, `target`, `.venv`, `.terraform`, and `.cache` are skipped automatically.

When multiple files in the same directory map to the same environment, the suffixed file takes priority over bare `.env` and a warning is printed.

### envy set

Add or update variables in a project.

```bash
envy set my-app DB_HOST=localhost DB_PORT=5432
envy set my-app DB_HOST=staging-db --env staging
envy set my-app PORT=3000 SERVICE_NAME=api --path services/api
envy set my-app PORT=3000 --path services/api --env prod
```

| Flag | Description |
|------|-------------|
| `--env <name>` | Target environment. Uses the [environment resolution order](#environment-resolution-order) if omitted. |
| `--path <subdir>` | Target a monorepo subdirectory instead of the root level. |

Each `KEY=VALUE` argument is an upsert -- safe to re-run.

### envy get

Retrieve a single variable's resolved value (with inheritance applied).

```bash
envy get my-app DB_HOST
envy get my-app DB_HOST --env staging
envy get my-app PORT --path services/api --env dev
```

| Flag | Description |
|------|-------------|
| `--env <name>` | Environment to query. |
| `--path <subdir>` | Monorepo subpath. Resolution includes inherited root variables. |

Prints the value to stdout with no additional formatting. Returns an error if the key is not found.

### envy load

Write a `.env` file to the current directory from stored config.

```bash
envy load --project my-app
envy load --project my-app --env staging
envy load --project my-app --path services/api
envy load --project my-app --env prod --output .env.prod
envy load --project my-app --format export
envy load --project my-app --force
envy load --project my-app --dry-run
envy load --project my-monorepo --all-paths --force
```

| Flag | Description |
|------|-------------|
| `--project <name>` | Project name (required). |
| `--env <name>` | Environment. |
| `--path <subdir>` | Monorepo subpath. |
| `--output <file>` | Output filename. Default: `.env`. |
| `--format <fmt>` | Output format: `dotenv` (default) or `export`. |
| `--force` | Overwrite an existing file without prompting for confirmation. |
| `--dry-run` | Print the output to stdout instead of writing a file. |
| `--all-paths` | Write `.env` files for all configured paths. Mutually exclusive with `--path` and `--output`. |

The generated file includes a header comment identifying the project, environment, and path.

When using `--all-paths`, root-level variables are written to `.env` in the current directory. Each configured path gets its own file at `<path>/.env` (e.g. `services/api/.env`), containing root variables merged with path-level overrides. Directories are created automatically. A single confirmation prompt lists all files before writing (skipped with `--force`).

If an environment has a [custom output filename](#envy-env-file) configured, `load` uses that filename instead of `.env`.

### envy update

Read an existing `.env` file and update the stored config from it.

```bash
envy update --project my-app
envy update --project my-app --env staging --path services/api
envy update --project my-app --merge
envy update --project my-app --file .env.local
```

| Flag | Description |
|------|-------------|
| `--project <name>` | Project name (required). |
| `--env <name>` | Target environment. |
| `--path <subdir>` | Target monorepo subpath. |
| `--merge` | Only add new keys. Existing keys in the store are not overwritten. |
| `--file <path>` | File to read. Default: `.env`. |

### envy list

List all projects in the config store.

```bash
envy list
envy list --json
envy list --quiet
```

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON. |
| `--quiet` | Print project names only, one per line. |

### envy show

Display a project's configuration.

```bash
envy show my-app
envy show my-app --env dev
envy show my-app --env dev --path services/api
envy show my-app --reveal
envy show my-app --json
```

| Flag | Description |
|------|-------------|
| `--env <name>` | Show a specific environment. |
| `--path <subdir>` | Show resolved variables for a subpath. |
| `--reveal` | Show actual values instead of masked output (`****`). |
| `--json` | Output as JSON. |

When neither `--env` nor `--path` is provided, the full project overview is shown, listing all environments and paths with variable counts.

### envy delete

Remove a project configuration entirely.

```bash
envy delete my-app
envy delete my-app --force
```

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt. |

This deletes the project YAML file from disk. The operation cannot be undone.

### envy env

Manage environments within a project. This command has subcommands for adding, removing, listing, and copying environments, as well as managing custom output filenames.

#### envy env add

```bash
envy env add my-app staging
```

Adds an empty environment to the project. Fails if the environment already exists.

#### envy env remove

```bash
envy env remove my-app staging
envy env remove my-app staging --force
```

Removes the environment from both root-level `environments` and all `paths`.

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt. |

#### envy env list

```bash
envy env list my-app
```

Lists all environment names in the project, sorted alphabetically. If an environment has a custom output filename, it is shown next to the name.

#### envy env copy

```bash
envy env copy my-app --from dev --to staging
```

Copies all root-level variables from the source environment to the destination. If the destination environment does not exist, it is created. Existing keys in the destination are overwritten.

| Flag | Description |
|------|-------------|
| `--from <name>` | Source environment (required). |
| `--to <name>` | Destination environment (required). |

#### envy env file

Manage custom output filenames for environments. By default, `envy load` writes to `.env`. Custom filenames let you map environments to files like `.env.local` or `.env.production`.

##### envy env file set

```bash
envy env file set my-app production .env.production
```

Sets the output filename for an environment. When `envy load` targets this environment, it writes to the specified filename instead of `.env`.

##### envy env file clear

```bash
envy env file clear my-app production
```

Removes the custom output filename, reverting to the default `.env`.

### envy diff

Compare environments or compare the local `.env` file against stored config.

```bash
envy diff my-app --env dev --env staging
envy diff my-app --env dev
envy diff my-app --env dev --env staging --reveal
```

| Flag | Description |
|------|-------------|
| `--env <name>` | Environment(s) to compare (repeatable). One `--env` compares local `.env` against that environment. Two `--env` flags compare those two environments. |
| `--reveal` | Show actual values instead of masked output. |

Output uses `+` for added keys, `-` for removed keys, and `~` for changed values.

### envy encrypt

Enable encryption on an existing project.

```bash
envy encrypt my-app
```

All existing variable values are encrypted in place. Prompts for a passphrase (twice, for confirmation). Fails if the project is already encrypted.

### envy decrypt

Disable encryption and store values as plaintext.

```bash
envy decrypt my-app
```

All values are decrypted and the `encryption` block is removed from the YAML file. Fails if the project is not encrypted.

### envy rekey

Change the encryption passphrase.

```bash
envy rekey my-app
```

Decrypts all values with the old passphrase, generates a new salt, and re-encrypts with the new passphrase. Fails if the project is not encrypted.

### envy version

Print the version string.

```bash
envy version
```

## Concepts

### Config store

All project configurations live as YAML files in a single directory:

```
~/.config/envy/
  projects/
    foo.yaml
    bar.yaml
    my-monorepo.yaml
```

The base config directory defaults to `~/.config/envy` (using `os.UserConfigDir()` on the current platform). Override it by setting `ENVY_CONFIG_DIR`.

### Data model

Each project YAML file has this structure:

```yaml
name: my-monorepo
created_at: 2026-02-10T12:00:00Z
updated_at: 2026-02-10T14:30:00Z
default_env: dev

environments:
  dev:
    DATABASE_URL: "postgres://localhost:5432/myapp_dev"
    API_KEY: "dev-key-123"
    DEBUG: "true"
  staging:
    DATABASE_URL: "postgres://staging-db:5432/myapp"
    API_KEY: "staging-key-456"
    DEBUG: "false"

paths:
  services/api:
    dev:
      PORT: "3000"
      SERVICE_NAME: "api"
      DATABASE_URL: "postgres://localhost:5432/api_dev"
  services/worker:
    dev:
      PORT: "3001"
      SERVICE_NAME: "worker"
      QUEUE_URL: "redis://localhost:6379"
```

Project names must start with an alphanumeric character and contain only alphanumeric characters, hyphens, or underscores. The name maps directly to the filename on disk (`my-app` becomes `my-app.yaml`).

All values are stored as strings. Values can optionally be encrypted at rest -- see [Encryption](#encryption) below.

### Variable inheritance

When loading variables for a path (e.g. `services/api`) in a given environment (e.g. `dev`):

1. Start with all variables from `environments.dev` (the root-level set).
2. Merge `paths.services/api.dev` on top. Path-level values override root-level values for the same key.
3. The merged result is what gets written to `.env`.

This means subdirectories automatically inherit every root variable and only need to declare what differs. If a path has no overrides for the target environment, the root variables are returned without error.

### Environment resolution order

When a command needs to determine which environment to use, the following precedence applies:

1. The `--env` flag (highest priority).
2. The `ENVY_ENV` environment variable.
3. The `default_env` field in the project config.
4. `dev` as a final fallback.

## Encryption

Envy supports optional per-project encryption. When enabled, variable values are encrypted at rest using AES-256-GCM with Argon2id key derivation. The YAML structure (project name, environment names, variable names, paths) remains readable -- only the values are opaque.

Encryption is transparent to all commands. When you run `envy get`, `envy load`, `envy show`, or any other command on an encrypted project, you are prompted for the passphrase automatically.

### Passphrase resolution

The passphrase is resolved in this order:

1. The `ENVY_PASSPHRASE` environment variable (useful for scripting and CI).
2. Interactive terminal prompt (passphrase is hidden).

If neither is available (e.g. piped stdin with no env var), the command fails with a clear error message.

### Encrypted YAML format

An encrypted project's YAML file looks like this:

```yaml
name: my-api
encryption:
  enabled: true
  salt: "base64-encoded-salt"
  params:
    time: 3
    memory: 65536
    threads: 4
environments:
  dev:
    DATABASE_URL: "ENC:base64-encoded-ciphertext..."
    API_KEY: "ENC:base64-encoded-ciphertext..."
```

Each value has a unique random nonce, so encrypting the same plaintext twice produces different ciphertexts.

### Encryption workflow

```bash
# Create a project and add some secrets
envy init my-api --env dev --env prod
envy set my-api DB_PASSWORD=hunter2 API_KEY=sk-12345 --env dev

# Enable encryption
envy encrypt my-api

# All commands still work -- passphrase is prompted automatically
envy show my-api --reveal
envy load --project my-api --dry-run

# Use the env var to avoid prompts (e.g. in CI)
export ENVY_PASSPHRASE=my-secret-passphrase
envy get my-api DB_PASSWORD

# Change passphrase
envy rekey my-api

# Remove encryption entirely
envy decrypt my-api
```

Alternatively, create an encrypted project from the start:

```bash
envy init my-api --encrypt --env dev --env prod
envy set my-api DB_PASSWORD=hunter2 --env dev
```

## Environment variables

| Variable | Description |
|----------|-------------|
| `ENVY_CONFIG_DIR` | Override the default config directory (`~/.config/envy`). When set, projects are stored under `$ENVY_CONFIG_DIR/projects/`. |
| `ENVY_ENV` | Default environment for commands that accept `--env`. Overridden by `--env` flag and by a project's `default_env`. |
| `ENVY_PASSPHRASE` | Encryption passphrase for non-interactive use (CI, scripting). When unset, envy prompts on the terminal. |

## LLM reference

Run `envy --llm` to print a comprehensive plain-text reference covering all commands, flags, concepts, and workflows. The output is designed for LLMs to consume as context when assisting with envy usage.

## A note on scope

Envy is designed for engineers juggling multiple hobby or side projects who need a simple, local-first way to manage `.env` files. It is not intended as a production secret management solution.

For production workloads, consider a dedicated secrets manager such as:

- [HashiCorp Vault](https://www.vaultproject.io/) -- self-hosted or managed (HCP Vault), with fine-grained access policies and audit logging.
- [Doppler](https://www.doppler.com/) -- SaaS platform purpose-built for syncing secrets across environments and CI/CD pipelines.
- [Infisical](https://infisical.com/) -- open-source secret management with native integrations for most deployment platforms.
- Cloud-provider offerings: [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/), [Google Cloud Secret Manager](https://cloud.google.com/secret-manager), or [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault/).

## Licence

See [LICENSE](LICENSE).
