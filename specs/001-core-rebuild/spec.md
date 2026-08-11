# Feature Specification: Core Rebuild

**Feature Branch**: `001-core-rebuild`

**Created**: 2026-08-11

**Status**: Draft

**Input**: User description: "Rebuild the Buenos Aires GitOps tool from a clean specification — a command-line tool that watches a git branch, validates committed shell scripts, executes them, tracks their status, and exposes a web interface for inspecting results. Recreate all planning artifacts from scratch."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Operator installs and configures the tool (Priority: P1)

An operator downloads the tool and runs a guided setup: they provide their username, a default log folder, the git branch to monitor, and the repository URL. The tool persists these as global defaults so subsequent runs need no reconfiguration. The operator can optionally enable the web interface and choose its port during setup.

**Why this priority**: Without setup there is nothing to monitor; this is the entry point for every other journey and must work before anything else is valuable.

**Independent Test**: Can be fully tested by running the setup flow on a fresh machine and verifying the global configuration file is created with the provided values, then re-running the tool against an unconfigured environment to confirm a clear error is shown.

**Acceptance Scenarios**:

1. **Given** a fresh environment with no prior configuration, **When** the operator runs the guided setup and answers all prompts, **Then** the global configuration file is created with their username, log folder, branch, repository URL, and web UI preference.
2. **Given** a missing or invalid global configuration, **When** the operator starts the monitor, **Then** the tool reports a clear, actionable error instead of failing silently or misbehaving.

---

### User Story 2 - Operator monitors a repository branch (Priority: P1)

The operator runs the monitor from an empty workspace or an existing clone. On first run the monitor clones the configured repository and performs an initial sync: every existing shell script in the scanned folder is validated and executed. Afterwards, on every poll cycle, the monitor detects new commits on the monitored branch, validates newly inserted or modified shell scripts, executes those that pass validation, and cleans up tracking for deleted scripts. This continues indefinitely until stopped.

**Why this priority**: This is the core value of the product — GitOps-style automated execution of infrastructure scripts.

**Independent Test**: Can be fully verified by setting up a test repository with a known-good script, starting the monitor, and confirming the script is validated and executed on first run, then committing a modified version and confirming it re-executes.

**Acceptance Scenarios**:

1. **Given** a monitored repository with existing shell scripts in the scanned folder, **When** the monitor runs for the first time, **Then** every script is validated and executed, and the last processed commit is recorded.
2. **Given** a new commit that inserts a shell script, **When** the monitor's next poll detects it, **Then** the script is validated and, if valid, executed exactly once.
3. **Given** a commit that modifies an existing script, **When** detected, **Then** the script is re-validated and re-executed regardless of its previous status.
4. **Given** a commit that deletes a script, **When** detected, **Then** the script's tracked status is removed.
5. **Given** an invalid script (syntax error), **When** detected, **Then** it is never executed and its failure is recorded; a later fix is re-processed and executed.

---

### User Story 3 - Operator inspects execution results (Priority: P2)

After scripts run, the operator can inspect what happened: per-script logs containing validation and execution output, per-script status (pending/success/failure/skipped) with timestamps, and asset metadata (generation count, last run time, lint/test/run outcomes, user, duration, status, commit hash). If enabled, the web interface lists all logs, shows individual log contents, and displays asset metadata alongside each log.

**Why this priority**: Execution without observability is not operationally usable; this journey makes the tool trustworthy in practice.

**Independent Test**: Can be fully tested by running a script that fails validation and one that succeeds, then verifying each produces a log file and distinct status/asset metadata, and that the web interface shows both.

**Acceptance Scenarios**:

1. **Given** scripts that have been processed, **When** the operator opens the log directory, **Then** each script has a log file containing both validation output and execution output.
2. **Given** the web interface enabled, **When** the operator opens it, **Then** they see a list of log files, can open any log, and see the associated asset metadata in a structured view.
3. **Given** processing that produced failures, **When** the operator inspects status, **Then** each script shows an accurate overall status (success or failure) distinguishable from pending/skipped.

---

### User Story 4 - Developer deploys via git push (Priority: P2)

A developer manages infrastructure scripts in a git repository. They commit a script to the monitored branch and push; the tool automatically validates and executes it. Fixing a failed script and pushing again re-deploys it. This is the GitOps workflow: the git history is the source of truth for what runs.

**Why this priority**: This journey differentiates the product from manual execution and is the reason to run it continuously.

**Independent Test**: Can be fully tested by pushing a valid script, then a broken version, then a fix, and verifying each push produces the expected execution outcome.

**Acceptance Scenarios**:

1. **Given** a running monitor, **When** the developer pushes a valid script, **Then** it executes without manual intervention.
2. **Given** a previously failed script, **When** the developer pushes a fix, **Then** the fixed version is validated and executed.

---

### User Story 5 - Security-conscious operator controls privilege (Priority: P3)

Executing infrastructure scripts often requires elevated privileges, but the operator controls exactly when this is possible. Sudo-capable execution is only active when BOTH the host operator's global configuration explicitly allows it AND the repository's configuration opts in. A repository alone can never escalate privileges.

**Why this priority**: Security boundaries protect the host; this is a trust-enabling capability rather than a daily flow.

**Independent Test**: Can be fully tested by enabling sudo in the repository config only (must not run with sudo), enabling it globally only (must not run with sudo), and enabling it in both (must run with sudo).

**Acceptance Scenarios**:

1. **Given** a repository config that allows sudo but a global config that does not, **When** a script is executed, **Then** it runs without elevated privileges.
2. **Given** both global and repository configs allow sudo, **When** a script is executed, **Then** it runs with elevated privileges.

---

### Edge Cases

- **First run with existing scripts**: initial sync processes every existing script; scripts that already succeeded are not re-run on subsequent restarts.
- **Missing or unreadable status file**: the tool must recover cleanly (treat as fresh state or a clear error, never a crash loop).
- **Base commit unavailable** (shallow clone / garbage collection): the monitor falls back to an initial sync instead of failing.
- **Dirty worktree or failed pull**: monitoring continues from remote refs; a failed pull never stops the loop.
- **Failed fetch**: the monitor keeps polling against the last known remote ref and retries next cycle.
- **Script outside the scanned folder / non-script files**: never processed.
- **Folder prefix collisions** (e.g. `shell` vs `shellish`): only exact folder-prefixed scripts are processed.
- **Path traversal attempts in script names**: rejected outright.
- **Shellcheck warnings**: non-fatal; the script still executes. Shellcheck fatal errors (exit code > 1) or absent tooling: script fails validation.
- **Empty log directory / nothing has run yet**: web interface renders an empty list, not an error.
- **Port already in use / web server failure**: logged, never fatal to the monitor.
- **Corrupt or truncated JSON state**: must be handled without data loss of other scripts' state.
- **Rename of a script**: behaves as delete + insert; a renamed previously-failed script is re-processed.
- **Restart mid-poll**: the last recorded commit determines the next diff; no script runs twice unless its content changed.
- **Shell script with relative paths**: executes from its real repository location so relative references resolve.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST provide a guided setup command that collects username, log folder, branch, repository URL, and optional web interface preference (with port) and persists them as the global default configuration.
- **FR-002**: The tool MUST provide a monitor command that polls the configured repository at a configurable interval (default 180 seconds) and processes changes continuously until stopped.
- **FR-003**: Change detection MUST be based on the remote tracking ref after fetch, so a dirty local worktree or failed pull NEVER stops processing.
- **FR-004**: On first run (or when the last processed commit is unavailable), the monitor MUST perform an initial sync that processes every existing shell script in the scanned folder.
- **FR-005**: Inserted scripts MUST be processed unless that script's overall status is already success; modified scripts MUST always be re-processed; deleted scripts MUST have their status entry removed.
- **FR-006**: Before execution, every script MUST pass syntax validation and static linting; scripts that fail validation MUST NOT be executed, and their failure MUST be recorded.
- **FR-007**: The tool MUST execute scripts with the configured user context, and MUST support privileged execution only when both the global and repository configurations explicitly allow it.
- **FR-008**: The tool MUST persist per-script status (lint/test/run phases and overall) and the last processed commit, resuming incrementally across restarts.
- **FR-009**: The tool MUST track per-script asset metadata including generation (incremented per modification), last run time, lint/test/run outcomes, event text, user, duration, status, and commit hash.
- **FR-010**: The tool MUST write per-script log files containing validation output and execution output.
- **FR-011**: The web interface MUST list log files, display individual log contents, and show asset metadata alongside each log, with hardened security headers and path-traversal protection.
- **FR-012**: Repository configuration MUST override global defaults for user, log folder, and sudo policy, and MUST support enabling/disabling plugins and selecting each plugin's scanned folder (default: the plugin's own name).
- **FR-013**: The tool MUST be runnable as a containerized deployment that includes the runtime dependencies (git, shell interpreter, linter, privilege tooling) and runs as a non-root user.
- **FR-014**: The project MUST ship a CI/CD pipeline that runs tests, linting, and security scanning on every push and pull request, and publishes a container image on the default branch.

### Key Entities *(include if feature involves data)*

- **Script Asset**: A shell script tracked by the tool, identified by its repository path; carries metadata: generation, last run time, lint/test/run outcomes, event text, executing user, run duration, status, and originating commit.
- **Execution Status**: Per-script lifecycle record with lint status, test status, run status, timestamp, and overall status (pending/success/failure/skipped).
- **Global Configuration**: Operator-level defaults (username, log folder, branch, poll interval, repository URL, web UI preference, sudo policy) stored outside the repository.
- **Repository Configuration**: Per-repository overrides (user, log folder, sudo policy) plus plugin settings (enabled, scanned folder) committed with the repository.
- **Log Entry**: A recorded file containing validation and execution output for one script execution.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A fresh workspace pointed at a repository with existing scripts processes 100% of those scripts on first run without manual intervention.
- **SC-002**: The full script lifecycle works end-to-end: insert → validate → execute; modify → re-execute; delete → cleanup; invalid script → recorded failure, never executed.
- **SC-003**: Privileged execution occurs if and only if both the global and repository configurations allow it (verified for all four permission combinations).
- **SC-004**: A restart of the monitor re-processes zero previously succeeded scripts and resumes from the last processed commit with no data loss.
- **SC-005**: An operator can determine every script's last outcome (success/failure with reason) from logs and the web interface within one navigation.
- **SC-006**: The monitor keeps running through induced failures (dirty worktree, failed pull, failed fetch, web server error) and recovers automatically on the next cycle.
- **SC-007**: All local and CI quality gates (lint, security scan, tests) pass on the default branch, and a container image is published automatically on merge.

## Assumptions

- The tool targets operators with a build toolchain available for building from source; prebuilt container images are provided for runtime.
- Shell scripts target a POSIX-compatible environment; the shell interpreter and linter are installed on the host (or included in the container image).
- The monitor runs in a workspace that is either empty or an existing clone of the monitored repository; cloning into a non-empty unrelated directory is out of scope.
- This is a single-user, single-host tool: no multi-user authentication, TLS, or remote access to the web interface is required for v1 (the interface binds to the host and is protected by path validation and security headers).
- The web interface is optional and never required for monitoring to function.
- Privileged execution requires the operator's host to allow passwordless elevation for the executing user (documented in deployment guidance).
- Snyk and SonarQube governance is handled externally; they are not part of this feature's pipeline scope.
