# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and development commands

```bash
make build          # builds binary to bin/envy
make test           # runs all tests with race detector
make lint           # runs go vet
make fmt-check      # checks code formatting
make man            # generates man pages to man/
make clean          # removes bin/
make release LEVEL=patch  # tags and pushes a release (patch|minor|major)
go test -race ./internal/config/...   # run tests for a single package
go test -race -run TestSetCmd ./internal/cli/...  # run a single test
```

## Architecture

Envy is a local-first CLI for managing `.env` files from a centralised YAML config store at `~/.config/envy/projects/`. Built with Go, Cobra for CLI, and `gopkg.in/yaml.v3` for persistence.

### Package layout

- **`cmd/envy`** -- Entrypoint. Creates a `config.Store`, wires `crypto.GetPassphrase` as the passphrase provider, and passes the store to `cli.NewRootCmd`.
- **`internal/config`** -- Core data model and persistence:
  - `Project` struct holds environments (root-level vars per env) and paths (per-subpath overrides per env). `Environments` is `map[string]map[string]string` (env -> key -> value). `Paths` is `map[string]map[string]map[string]string` (path -> env -> key -> value). Optional `Encryption *EncryptionConfig` field enables per-project encryption. Optional `EnvFiles map[string]string` maps environments to custom output filenames for `load`.
  - `Store` handles YAML file CRUD on disk. `Load`/`Save` transparently decrypt/encrypt values for encrypted projects. `LoadRaw`/`SaveRaw` bypass encryption for commands that manage it directly (encrypt, decrypt, rekey).
  - `ResolveEnv` implements the environment resolution order (flag > `ENVY_ENV` > `default_env` > `"dev"`).
  - `ResolveVars` merges root-level vars with path-level overrides for a given env+path.
- **`internal/crypto`** -- Encryption primitives and passphrase handling:
  - AES-256-GCM encryption with per-value random nonces. Values stored as `ENC:<base64>` in YAML.
  - Argon2id key derivation from passphrase + salt. Parameters stored per-project in `EncryptionConfig.Params`.
  - `GetPassphrase` resolves via `ENVY_PASSPHRASE` env var or interactive terminal prompt. `GetPassphraseWithConfirm` prompts twice for new passphrases.
- **`internal/scan`** -- Directory scanner for `.env` file discovery. Maps filenames to environment names via conventions (e.g. `.env.staging` -> `staging`), resolves conflicts when multiple files map to the same environment, and skips common non-project directories.
- **`internal/cli`** -- One file per subcommand. Each exports a `NewXxxCmd(store)` constructor. All commands receive a `*config.Store` via closure (no globals). Test helpers in `helpers_test.go` provide `setupTestStore` (temp dir), `setupEncryptedTestStore` (with fixed passphrase func), and `executeCommand` (captures stdout/stderr).
- **`internal/dotenv`** -- `.env` parser and writer. Parser handles bare, single-quoted, double-quoted values, `export` prefix, inline comments. Writer handles quoting and escape sequences.
- **`internal/util`** -- `ConfigDir`/`ProjectsDir` for path resolution (respects `ENVY_CONFIG_DIR`), `MaskValue`/`FormatKeyValue` for display formatting.

### Key patterns

- All CLI tests use `t.TempDir()` for isolated stores -- no shared test state. Encryption tests use `t.Setenv("ENVY_PASSPHRASE", ...)` to avoid terminal prompts.
- Commands that mutate state call `store.Save(project)` after modifications and set `project.UpdatedAt`.
- Variable inheritance: path-level vars are merged on top of root-level env vars at resolution time, not at storage time. The store keeps them separate.
- Encryption is transparent: `Store.Load` decrypts, `Store.Save` encrypts a deep copy (caller's in-memory Project stays plaintext). Commands that manage encryption directly (encrypt, decrypt, rekey) use `LoadRaw`/`SaveRaw` to bypass the automatic encrypt/decrypt.
- CLI commands that need injectable I/O for testing use a `newXxxCmd(store, fn)` internal constructor alongside the public `NewXxxCmd(store)`. The public constructor passes `nil` for the func, which falls back to the production implementation.

### Code comments

- Do not add godoc comments that merely restate the function/type name. For example, `// Delete removes a project file from disk.` on `func (s *Store) Delete(name string) error` adds nothing. Only add a doc comment when it conveys information not already obvious from the signature (edge cases, defaults, side effects, format details, priority order).
- Do not add inline comments that describe what the next line of code does (e.g. `// Ensure the environment exists` before a nil-map check, or `// Strip optional export prefix` before `strings.HasPrefix(line, "export ")`).
- Do not add section-label comments in tests (e.g. `// Create`, `// Verify round-trip`) when the code is self-explanatory.
- Comments are welcome when they explain **why** something is done, document non-obvious behaviour, describe algorithm intent, or note caveats the code cannot express through naming or structure alone.
