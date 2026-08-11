# AGENTS.md

Buenos Aires: a Go (1.24) GitOps tool that watches a git branch, validates `.sh` scripts (`bash -n` + `shellcheck`), and executes them. Cobra CLI (`install`, `run`); `main.go` → `cmd.Execute()`. Module is declared as `module buenosaires`, so imports are `buenosaires/...` (not `github.com/gbenselum/...`).

## Commands

- `make check` — the full local gate: `golangci-lint run ./...` + `gosec ./...` + `go test ./...`. Mirrors CI; must pass before PRs.
- `make hooks` — `pre-commit run --all-files` (run `pre-commit install` once). `make fmt` — `gofmt -w .`.
- `go test ./...` requires `bash` and `shellcheck` in PATH — the shell plugin tests exec them. CI installs shellcheck via apt.
- To run the monitor locally: `buenosaires install` first (creates `~/.buenosaires/config.toml`; `run` fails without it), then `buenosaires run` from an **empty directory or an existing clone** of the monitored repo — it clones into `$PWD` if neither.

## Architecture

- `cmd/run.go` holds the core loop: fetch → diff `origin/<branch>` ref vs `last_commit` persisted in `.buenosaires/status.json` → process changes. Change detection reads remote refs and git objects, **not the worktree** — a dirty worktree or failed pull never stops monitoring.
- `internal/config` — global config (`~/.buenosaires/config.toml`) + per-repo `config.toml` (repo overrides global `user`/`log_dir`/`allow_sudo`).
- `internal/status` — `.buenosaires/status.json`: per-script status + `last_commit`. Empty `last_commit` ⇒ initial sync of all existing scripts; a missing base commit also falls back to initial sync.
- `internal/web` — PatternFly UI; `go:embed assets/patternfly.min.css` (keep the file; removing it breaks the build). Serves logs + `plugins/shell/assets/{name}.json`. Web server errors are logged, never fatal.
- `plugins/shell` — the only plugin. Per-script asset JSON in `plugins/shell/assets/` (gitignored).
- `src/` is empty; ignore it.

## Processing semantics (easy to get wrong)

- Modify ⇒ always re-process. Insert ⇒ skipped only if that script's overall status is already `success`. Delete ⇒ removes the script's status entry.
- Scripts are materialized from the git tree at their real repo path (preserving committed file mode) and run via `bash <path>` — relative paths inside scripts resolve against the repo layout.
- `allow_sudo` is a two-gate: global AND repo config must both be `true`; a repo cannot escalate on its own.
- Default poll interval is `sync_interval` = 180s.
- Runtime artifacts (`.buenosaires/`, `plugins/*/assets/`, `*.log`) are gitignored — never commit them.

## Conventions

- Snyk and SonarQube are handled by GitHub/SonarCloud — do not add them to CI.
- gosec runs locally; intentional exceptions use `#nosec` comments (11 in the tree). Keep them with their justifications.
- `.golangci.yml` is v2 format. govet's `inline` analyzer is deliberately disabled (false positives when the local Go is newer than the module's `go 1.24.3`); don't re-enable.
- New plugins must follow the shell plugin contract: `enabled` bool, `folder_to_scan` key (defaults to plugin name), per-asset JSON tracking `generation` (incremented per modification), `last_run`, lint/test/run status, event, user, run_duration, status, commit_hash.
- Path-traversal hardening (web log/asset handlers, duplicated `sanitizeRepoPath` in config/status) is deliberate; preserve it.

## Reference

- `memory/application_overview.md` is the maintained project-knowledge file — keep it in sync when behavior changes.
- CI (`.github/workflows/ci.yml`): test, lint (golangci-lint v2.12.2), gosec (v2.28.0) on push to `main` and PRs; `build-and-push` publishes `ghcr.io/<owner>/buenosaires:latest`. Docker run details are in the README.
