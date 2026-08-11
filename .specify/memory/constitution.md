<!--
  SYNC IMPACT REPORT
  Version change: (none) -> 1.0.0 (initial ratification)
  Modified principles: n/a (fresh document)
  Added sections: Core Principles (I-VII), Security & Runtime Constraints,
                  Development Workflow & Quality Gates, Governance
  Removed sections: n/a
  Follow-up TODOs: none
-->

# Buenos Aires Constitution

## Core Principles

### I. Quality Gate is Non-Negotiable

Every change MUST pass the full local gate before it is merged: `golangci-lint run ./...`, `gosec ./...`, and `go test ./...` (the `make check` target). This gate MUST mirror the CI pipeline exactly — CI and local checks are never allowed to diverge.

### II. Security by Design

Security hardening is deliberate and MUST be preserved: duplicated path-traversal guards (`sanitizeRepoPath` in `internal/config` and `internal/status`, web log/asset handlers), `#nosec` exceptions MUST carry justifications and be kept with them, and the two-gate `allow_sudo` model (global AND repo opt-in) MUST NOT be weakened. A repository committed by anyone MUST NOT be able to escalate privileges on its own.

### III. Runtime Artifact Hygiene

Runtime state MUST NEVER be committed: `.buenosaires/`, `plugins/*/assets/`, and `*.log` are gitignored by contract. Logs and status files are sensitive execution records written with restrictive permissions (0600 for logs/status/assets; 0750 for directories).

### IV. Knowledge Synchronization

`memory/application_overview.md` is the maintained project-knowledge file and MUST stay in sync whenever behavior changes. It is the source of truth for setup context and configuration decisions; operational docs (README, AGENTS.md) MUST not contradict it.

### V. External SAST Governance

Snyk and SonarQube are handled by GitHub/SonarCloud and MUST NOT be added to the CI pipeline. Local security scanning is gosec's responsibility; the `.snyk` file manages Snyk policy exclusions only.

### VI. Toolchain Fidelity

The module targets Go 1.24.3. The `govet` `inline` analyzer MUST remain disabled in `.golangci.yml` (false positives when the local toolchain is newer than the module's go directive) and MUST NOT be re-enabled. `bash` and `shellcheck` are required dependencies for tests and MUST be installed in CI (apt) and the Docker runtime image (alpine packages).

### VII. Plugin Contract

New plugins MUST follow the shell plugin contract: an `enabled` boolean, a `folder_to_scan` key (defaulting to the plugin name), and per-asset JSON tracking `generation` (incremented per modification), `last_run`, lint/test/run status, event, user, run_duration, status, and commit_hash.

## Security & Runtime Constraints

- Scripts are materialized from the git tree at their real repository path (preserving committed file mode) and executed via `bash <path>` — never through a shell string, so no command injection is possible.
- Change detection reads remote refs and git objects, not the worktree — a dirty worktree or failed pull MUST NOT stop monitoring.
- Web server errors are logged, never fatal; a web GUI failure MUST NOT kill the monitor.
- The web interface MUST keep its embedded PatternFly stylesheet (`internal/web/assets/patternfly.min.css`); removing it breaks the build.
- The shell plugin's `allow_sudo` execution requires passwordless sudo configuration (documented in the Dockerfile) and is gated by the two-gate model.
- Runtime artifacts are written with restrictive permissions and NEVER committed.

## Development Workflow & Quality Gates

- Every PR MUST pass CI: tests, golangci-lint (v2.12.2), and gosec (v2.28.0) on push to `main` and pull requests.
- Pre-commit hooks enforce formatting, linting, security, tests, and meta checks (trailing whitespace, EOF newlines, YAML/JSON validity, merge conflicts, large files) on every commit.
- Intentional gosec exceptions use `#nosec` comments with justifications (11 in the tree); new exceptions require explicit rationale.
- `src/` is empty and ignored; do not place code there.
- New features are planned via Spec Kit artifacts (`specs/`) and executed through `/speckit-implement`; `/speckit-analyze` must pass before implementation.

## Governance

This constitution supersedes all other project guidance. Amendments require documentation, explicit approval, and a migration plan; version bumps follow semver (MAJOR for principle removals/redefinitions, MINOR for additions, PATCH for clarifications). Every PR/review MUST verify compliance with this constitution; complexity MUST be justified. The constitution is enforced during `/speckit-analyze` — conflicts are CRITICAL and require adjusting the spec, plan, or tasks, never diluting the principle.

**Version**: 1.0.0 | **Ratified**: 2026-08-11 | **Last Amended**: 2026-08-11
