# Research: Core Rebuild

**Date**: 2026-08-11
**Feature**: [spec.md](./spec.md)

## Technology Trade-offs

### 1. Language: Go (retained)

| Option | Verdict |
|--------|---------|
| **Go 1.24** (chosen) | Static binary, excellent git integration via go-git, native concurrency for the poll loop, single-binary deployment, strong tooling (golangci-lint, gosec, go vet) |
| Python | Faster iteration, but requires runtime + dependency management on hosts and in containers; weaker static analysis gates |
| Rust | Strong safety, but slower iteration for a small CLI and no mature pure-Go git equivalent |

**Decision**: Retain Go 1.24.3 — the existing codebase, CI, and Docker toolchain already target it; no NEEDS CLARIFICATION remains.

### 2. CLI framework: Cobra (retained)

| Option | Verdict |
|--------|---------|
| **Cobra** (chosen) | Subcommand structure (`install`, `run`) with help generation, familiar in Go ecosystem, supports future subcommands |
| stdlib `flag` | Adequate for one flag but poor ergonomics for subcommand-based CLIs |
| urfave/cli | Valid alternative; Cobra already in use and consistent with the codebase |

**Decision**: Cobra v1.10.1.

### 3. Git operations: go-git (retained)

| Option | Verdict |
|--------|---------|
| **go-git/v5** (chosen) | Pure-Go repository access: remote-ref resolution, tree diffing, object reading — enables worktree-independent change detection (constitution §Security & Runtime Constraints) |
| Shelling out to `git` CLI | Adds a hard runtime dependency and makes worktree-independent diffing much harder; container would need full git plumbing |

**Decision**: go-git v5.16.3.

### 4. Configuration format: TOML (retained)

| Option | Verdict |
|--------|---------|
| **TOML** (chosen) | Human-editable, typed, native Go support via BurntSushi/toml; matches existing `~/.buenosaires/config.toml` and repo `config.toml` |
| YAML | More expressive but error-prone indentation rules and a larger dependency surface |
| JSON | Not user-friendly for hand-edited config |

**Decision**: TOML.

### 5. Shell validation policy

- **Syntax check**: `bash -n` — hard failure on syntax errors (script never executes).
- **Linting**: `shellcheck -s bash` — exit code 1 (warnings) is **non-fatal**; exit codes > 1 (errors) or tool-absent are **fatal**.
- Rationale: warnings are advisory; syntax errors and analyzer errors indicate the script cannot be trusted.

### 6. Web UI: embedded PatternFly (retained)

`go:embed` of `patternfly.min.css` keeps the UI dependency-free at runtime. The CSS file is build-critical and must never be removed (constitution §Security & Runtime Constraints). UI is informational only (log list/view, asset JSON), optional, and non-fatal on failure.

### 7. Sudo execution model (retained, constitution-mandated)

Two-gate: global `~/.buenosaires/config.toml` `allow_sudo = true` **AND** repo `config.toml` `allow_sudo = true`. A repository alone cannot escalate privileges. Requires passwordless sudo for the executing user (documented; container configures it for `appuser`).

## Open Questions

None — all specification items resolved; no `[NEEDS CLARIFICATION]` markers remain in spec.md.

## Reference

- `memory/application_overview.md` — maintained project knowledge (kept in sync per constitution §IV)
- `AGENTS.md` — operational conventions for the agent roster
