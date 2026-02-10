# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and development commands

```bash
make build          # builds binary to bin/envy
make test           # runs all tests with race detector
make lint           # runs go vet
make clean          # removes bin/
go test -race ./internal/config/...   # run tests for a single package
go test -race -run TestSetCmd ./internal/cli/...  # run a single test
```

## Architecture

Envy is a local-first CLI for managing `.env` files from a centralised YAML config store at `~/.config/envy/projects/`. Built with Go, Cobra for CLI, and `gopkg.in/yaml.v3` for persistence.

### Package layout

- **`cmd/envy`** -- Entrypoint. Creates a `config.Store` and passes it to `cli.NewRootCmd`.
- **`internal/config`** -- Core data model and persistence:
  - `Project` struct holds environments (root-level vars per env) and paths (per-subpath overrides per env). Both are nested `map[string]map[string]string`.
  - `Store` handles YAML file CRUD on disk.
  - `ResolveEnv` implements the environment resolution order (flag > `ENVY_ENV` > `default_env` > `"dev"`).
  - `ResolveVars` merges root-level vars with path-level overrides for a given env+path.
- **`internal/cli`** -- One file per subcommand. Each exports a `NewXxxCmd(store)` constructor. All commands receive a `*config.Store` via closure (no globals). Test helpers in `helpers_test.go` provide `setupTestStore` (temp dir) and `executeCommand` (captures stdout/stderr).
- **`internal/dotenv`** -- `.env` parser and writer. Parser handles bare, single-quoted, double-quoted values, `export` prefix, inline comments. Writer handles quoting and escape sequences.
- **`internal/util`** -- `ConfigDir`/`ProjectsDir` for path resolution (respects `ENVY_CONFIG_DIR`), `MaskValue`/`FormatKeyValue` for display formatting.

### Key patterns

- All CLI tests use `t.TempDir()` for isolated stores -- no shared test state.
- Commands that mutate state call `store.Save(project)` after modifications and set `project.UpdatedAt`.
- Variable inheritance: path-level vars are merged on top of root-level env vars at resolution time, not at storage time. The store keeps them separate.
