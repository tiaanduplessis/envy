# envy

A local-first Go CLI for managing `.env` files across projects from a centralised YAML config store.

Envy stores project configurations as YAML files under `~/.config/envy/projects/`. Each project gets its own file. Variables are organised by named environments (e.g. `dev`, `staging`, `prod`) and optional monorepo subpaths. When you run `envy load`, it merges the appropriate variables and writes a `.env` file to your working directory.

## Installation

Requires Go 1.25.7 or later.

```bash
go install github.com/tiaanduplessis/envy/cmd/envy@latest
```

Or build from source:

```bash
git clone https://github.com/tiaanduplessis/envy.git
cd envy
make build
# binary is at bin/envy
```

## Config store

All project configurations live as YAML files in a single directory:

```
~/.config/envy/
  projects/
    foo.yaml
    bar.yaml
    my-monorepo.yaml
```

The base config directory defaults to `~/.config/envy` (using `os.UserConfigDir()` on the current platform). Override it by setting the `ENVY_CONFIG_DIR` environment variable:

```bash
export ENVY_CONFIG_DIR=/path/to/custom/config
# Projects will be stored under /path/to/custom/config/projects/
```

## Data model

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
  prod:
    DATABASE_URL: "postgres://prod-db:5432/myapp"
    API_KEY: "prod-key-789"
    DEBUG: "false"

paths:
  services/api:
    dev:
      PORT: "3000"
      SERVICE_NAME: "api"
      DATABASE_URL: "postgres://localhost:5432/api_dev"
    staging:
      PORT: "3000"
      SERVICE_NAME: "api"
  services/worker:
    dev:
      PORT: "3001"
      SERVICE_NAME: "worker"
      QUEUE_URL: "redis://localhost:6379"
```

Project names must start with an alphanumeric character and contain only alphanumeric characters, hyphens, or underscores. The name maps directly to the filename on disk (`my-app` becomes `my-app.yaml`).

All values are stored as strings. Values can optionally be encrypted at rest -- see [Encryption](#encryption) below.

## Variable inheritance

When loading variables for a path (e.g. `services/api`) in a given environment (e.g. `dev`):

1. Start with all variables from `environments.dev` (the root-level set).
2. Merge `paths.services/api.dev` on top. Path-level values override root-level values for the same key.
3. The merged result is what gets written to `.env`.

This means subdirectories automatically inherit every root variable and only need to declare what differs. If a path has no overrides for the target environment, the root variables are returned without error.

## Environment resolution order

When a command needs to determine which environment to use, the following precedence applies:

1. The `--env` flag (highest priority).
2. The `ENVY_ENV` environment variable.
3. The `default_env` field in the project config.
4. `dev` as a final fallback.

## Commands

### envy init

Create a new project configuration.

```bash
# Minimal: creates project with a single "dev" environment
envy init my-app

# Multiple environments
envy init my-app --env dev --env staging --env prod

# With monorepo subpath stubs
envy init my-monorepo --path services/api --path services/worker

# Set a specific default environment
envy init my-app --env dev --env staging --default-env staging

# Create an encrypted project from the start
envy init my-app --encrypt
```

| Flag | Description |
|------|-------------|
| `--env <name>` | Environments to create (repeatable). Defaults to `dev` if omitted. |
| `--path <subdir>` | Monorepo subpath stubs to create (repeatable). |
| `--default-env <name>` | Set the default environment. Defaults to the first `--env` value, or `dev`. |
| `--encrypt` | Enable encryption on the new project. Prompts for a passphrase. |

Fails if a project with the same name already exists.

### envy set

Add or update variables in a project.

```bash
# Set root-level variables in the default environment
envy set my-app DB_HOST=localhost DB_PORT=5432

# Set in a specific environment
envy set my-app DB_HOST=staging-db --env staging

# Set path-level overrides for the default environment
envy set my-app PORT=3000 SERVICE_NAME=api --path services/api

# Set path-level overrides for a specific environment
envy set my-app PORT=3000 --path services/api --env prod
```

| Flag | Description |
|------|-------------|
| `--env <name>` | Target environment. Uses the environment resolution order if omitted. |
| `--path <subdir>` | Target a monorepo subdirectory instead of the root level. |

Each `KEY=VALUE` argument is an upsert -- safe to re-run.

### envy get

Retrieve a single variable's resolved value (with inheritance applied).

```bash
# Get from the default environment, root level
envy get my-app DB_HOST

# Get from a specific environment
envy get my-app DB_HOST --env staging

# Get with path inheritance applied
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
# Write .env using the default environment, root-level variables
envy load --project my-app

# Use a specific environment
envy load --project my-app --env staging

# Include path-level overrides (merged with root)
envy load --project my-app --path services/api

# Custom output filename
envy load --project my-app --env prod --output .env.prod

# Output in export format (export KEY="value")
envy load --project my-app --format export

# Overwrite without confirmation prompt
envy load --project my-app --force

# Preview to stdout without writing a file
envy load --project my-app --dry-run

# Write .env files for all configured paths at once
envy load --project my-monorepo --all-paths --force

# Preview all paths without writing
envy load --project my-monorepo --all-paths --dry-run
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

### envy update

Read an existing `.env` file and update the stored config from it.

```bash
# Read .env from the current directory, update the default environment
envy update --project my-app

# Update a specific environment and path
envy update --project my-app --env staging --path services/api

# Only add new keys; do not overwrite existing values
envy update --project my-app --merge

# Read from a specific file instead of .env
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
# Default output
envy list
# my-app              (envs: dev, staging, prod)
# my-monorepo         (envs: dev, staging, prod) [2 path(s)]

# JSON output
envy list --json

# Project names only, one per line
envy list --quiet
```

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON. |
| `--quiet` | Print project names only, one per line. |

### envy show

Display a project's configuration.

```bash
# Full project overview (values masked by default)
envy show my-app

# Show a specific environment's variables
envy show my-app --env dev

# Show resolved variables for a subpath (with inheritance)
envy show my-app --env dev --path services/api

# Reveal actual values
envy show my-app --reveal

# JSON output
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
# Prompts for confirmation
envy delete my-app

# Skip confirmation
envy delete my-app --force
```

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt. |

This deletes the project YAML file from disk. The operation cannot be undone.

### envy env

Manage environments within a project. This command has four subcommands.

#### envy env add

```bash
envy env add my-app staging
```

Adds an empty environment to the project. Fails if the environment already exists.

#### envy env remove

```bash
# Prompts for confirmation
envy env remove my-app staging

# Skip confirmation
envy env remove my-app staging --force
```

Removes the environment from both root-level `environments` and all `paths`.

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt. |

#### envy env list

```bash
envy env list my-app
# dev
# staging
# prod
```

Lists all environment names in the project, sorted alphabetically.

#### envy env copy

```bash
envy env copy my-app --from dev --to staging
```

Copies all root-level variables from the source environment to the destination. If the destination environment does not exist, it is created. Existing keys in the destination are overwritten.

| Flag | Description |
|------|-------------|
| `--from <name>` | Source environment (required). |
| `--to <name>` | Destination environment (required). |

### envy diff

Compare environments or compare the local `.env` file against stored config.

```bash
# Compare two stored environments
envy diff my-app --env dev --env staging

# Compare local .env against a stored environment
envy diff my-app --env dev

# Show actual values in the diff output
envy diff my-app --env dev --env staging --reveal
```

| Flag | Description |
|------|-------------|
| `--env <name>` | Environment(s) to compare (repeatable). One `--env` compares local `.env` against that environment. Two `--env` flags compare those two environments. |
| `--reveal` | Show actual values instead of masked output. |

Output uses `+` for added keys, `-` for removed keys, and `~` for changed values.

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

### envy encrypt

Enable encryption on an existing project.

```bash
envy encrypt my-app
# Prompts for passphrase (twice, for confirmation)
```

All existing variable values are encrypted in place. The project's YAML file is rewritten with the `encryption` block and `ENC:`-prefixed values.

Fails if the project is already encrypted.

### envy decrypt

Disable encryption and store values as plaintext.

```bash
envy decrypt my-app
# Prompts for passphrase
```

All values are decrypted and the `encryption` block is removed from the YAML file.

Fails if the project is not encrypted.

### envy rekey

Change the encryption passphrase.

```bash
envy rekey my-app
# Prompts for current passphrase, then new passphrase (twice)
```

Decrypts all values with the old passphrase, generates a new salt, and re-encrypts with the new passphrase.

Fails if the project is not encrypted.

### Encryption workflow example

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
# Values are encrypted on disk immediately
```

## Example workflow

```bash
# 1. Create a project with multiple environments and paths
envy init my-monorepo --env dev --env staging --env prod \
    --path services/api --path services/worker

# 2. Set shared root-level variables
envy set my-monorepo DATABASE_URL=postgres://localhost/dev DEBUG=true --env dev
envy set my-monorepo DATABASE_URL=postgres://staging-db/app --env staging
envy set my-monorepo DATABASE_URL=postgres://prod-db/app --env prod

# 3. Set service-specific overrides
envy set my-monorepo PORT=3000 SERVICE_NAME=api --path services/api
envy set my-monorepo PORT=3001 SERVICE_NAME=worker QUEUE_URL=redis://localhost \
    --path services/worker

# 4. Generate .env for the API service in the dev environment
cd ~/code/my-monorepo/services/api
envy load --project my-monorepo --path services/api --env dev
# .env now contains merged root dev vars + api-specific overrides

# 5. After editing .env locally, push changes back to the store
envy update --project my-monorepo --path services/api --env dev

# 6. Compare environments
envy diff my-monorepo --env dev --env staging

# 7. Check a single resolved value
envy get my-monorepo PORT --path services/api
# 3000
```

## Licence

See [LICENSE](LICENSE).
