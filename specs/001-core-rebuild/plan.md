# Implementation Plan: Core Rebuild

**Branch**: `feat/speckit-bootstrap` | **Date**: 2026-08-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-core-rebuild/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Rebuild the Buenos Aires GitOps tool from a clean, spec-driven baseline: a CLI that watches a git branch, validates committed shell scripts (syntax + lint) before executing them, tracks per-script status and asset metadata, writes per-script logs, and exposes an optional web interface. The technical approach is a single Go binary (module `buenosaires`) with a Cobra CLI (`install`, `run`), go-git for repository operations, TOML configuration (global + per-repo), filesystem-backed JSON state (`.buenosaires/status.json`, `plugins/*/assets/*.json`), and an embedded-PatternFly web server. The target layout already exists in the repository and is retained as the canonical structure.

## Technical Context

**Language/Version**: Go 1.24.3 (module `buenosaires`)

**Primary Dependencies**: github.com/spf13/cobra (CLI), github.com/go-git/go-git/v5 (repository operations), github.com/BurntSushi/toml (configuration)

**Storage**: Filesystem only — `~/.buenosaires/config.toml` (global), repo `config.toml` (overrides), `.buenosaires/status.json` (execution state), `plugins/*/assets/<path>.json` (per-asset metadata), `<log_dir>/*.log` (per-script logs)

**Testing**: `go test ./...` (unit + git-tree-based integration), shellcheck + bash for plugin validation tests; golangci-lint (v2.12.2) and gosec (v2.28.0) as quality gates

**Target Platform**: macOS/Linux CLI; multi-stage Alpine container (`appuser`, non-root) for deployment

**Project Type**: CLI tool (GitOps monitor) with optional embedded web UI

**Performance Goals**: Default poll interval 180s; single-host, single-repository scope; per-script processing must complete within one poll cycle for typical scripts

**Constraints**: Go 1.24.3 module directive; `govet` `inline` analyzer stays disabled (false positives with newer local toolchains); runtime deps `bash` + `shellcheck` required in CI and container; no Snyk/SonarQube in CI (external); path-traversal hardening preserved

**Scale/Scope**: Single operator host, one monitored repository, N scripts; no multi-user, no TLS, no remote web access in v1

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Evidence of compliance with the project constitution (v1.0.0):

- **I. Quality Gate**: plan mandates `make check` (golangci-lint + gosec + go test) as the completion gate for every implementation phase; CI mirrors it (Constitution §Development Workflow).
- **II. Security by Design**: all security-critical tasks preserve the two-gate `allow_sudo` model, path-traversal guards (`sanitizeRepoPath`, web handlers), and `#nosec` justifications.
- **III. Runtime Artifact Hygiene**: tasks enforce gitignore contract (`.buenosaires/`, `plugins/*/assets/`, `*.log`) and restrictive file permissions (0600/0750).
- **IV. Knowledge Synchronization**: final phase requires updating `memory/application_overview.md` in sync with any behavior changes.
- **V. External SAST Governance**: no task adds Snyk/SonarQube to CI.
- **VI. Toolchain Fidelity**: Go 1.24.3 directive retained; `govet.inline` stays disabled.
- **VII. Plugin Contract**: shell plugin tasks validate asset schema (generation, last_run, lint/test/run status, event, user, run_duration, status, commit_hash) and `folder_to_scan` defaulting.

Result: **PASS** — no violations. Complexity Tracking left empty.

## Project Structure

### Documentation (this feature)

```text
specs/001-core-rebuild/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 output — technology trade-offs
├── data-model.md        # Phase 1 output — entities & schemas
├── quickstart.md        # Phase 1 output — local validation guide
├── contracts/           # Phase 1 output — interface contracts
│   ├── config-contract.md      # Global + repo config file schemas
│   ├── status-contract.md      # .buenosaires/status.json schema
│   ├── plugin-contract.md      # Plugin + asset JSON contract
│   └── web-contract.md         # Web UI routes & security requirements
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
buenosaires/
├── main.go                  # Entry point → cmd.Execute()
├── cmd/
│   ├── root.go              # Cobra root command
│   ├── install.go           # Guided setup → global config
│   ├── run.go               # Monitor loop (fetch → diff → process → persist)
│   └── run_test.go          # Change-selection unit tests
├── internal/
│   ├── config/
│   │   ├── config.go        # Global/Repo/Plugin/GUI config + sanitizeRepoPath
│   │   └── config_test.go
│   ├── status/
│   │   ├── status.go        # Status/ScriptStatus persistence
│   │   └── (status_test.go) # NEW: required coverage
│   └── web/
│       ├── server.go        # PatternFly UI, log/asset handlers
│       ├── server_test.go
│       └── assets/patternfly.min.css   # go:embed — never remove
├── plugins/
│   ├── shell/
│   │   ├── shell.go         # LintAndValidate (bash -n + shellcheck) + Run
│   │   ├── asset.go         # Asset model
│   │   ├── duration.go      # time.Duration JSON wrapper
│   │   ├── shell_test.go
│   │   └── (integration_test.go)  # NEW: end-to-end script lifecycle
├── .github/workflows/ci.yml # test + lint + security + build-and-push
├── Dockerfile               # Multi-stage alpine, non-root appuser
├── Makefile                 # fmt/lint/sec/test/check/hooks
├── .golangci.yml            # v2 config (govet.inline disabled)
├── .pre-commit-config.yaml
├── memory/application_overview.md   # Maintained knowledge file
└── config.toml.example
```

**Structure Decision**: Single Go module (`buenosaires`) with the existing package layout retained as canonical. The rebuild preserves the proven architecture (remote-ref change detection, tree-materialized execution, two-gate sudo) and focuses spec-driven work on closing coverage gaps (status tests, lifecycle integration tests), hardening (config validation, signal handling), and documentation sync. No new top-level directories are introduced; `src/` remains empty and unused.

## Phases

| Phase | Scope | Outputs | Gate |
|-------|-------|---------|------|
| 0 | Research & baseline verification | research.md; confirm existing layout matches plan | Constitution re-check |
| 1 | Data model & contracts | data-model.md, contracts/*, quickstart.md | Contracts reviewed against spec |
| 2 | Task breakdown | tasks.md (via /speckit-tasks) | /speckit-analyze passes |
| 3 | Implementation | Code changes per tasks | make check green per task |
| 4 | Convergence & knowledge sync | memory/application_overview.md update; /speckit-converge | Final analyze + CI green |

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations — table intentionally empty.
