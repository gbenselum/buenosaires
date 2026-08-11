# Data Model: Core Rebuild

**Date**: 2026-08-11
**Feature**: [spec.md](./spec.md)

Entities derived from spec.md §Key Entities and the existing implementation. All storage is filesystem-based.

## 1. Global Configuration (`~/.buenosaires/config.toml`)

| Field | Type | Purpose |
|-------|------|---------|
| `user` | string | Default user for running scripts |
| `log_dir` | string | Default log directory |
| `branch` | string | Git branch to monitor (e.g. `main`) |
| `sync_interval` | int | Poll interval in seconds (default 180) |
| `repository_url` | string | Remote repository to monitor |
| `allow_sudo` | bool | Host-operator gate for privileged execution |
| `gui.enabled` | bool | Enable web interface |
| `gui.port` | int | Web interface port (default 9099) |

## 2. Repository Configuration (repo `config.toml`)

| Field | Type | Purpose |
|-------|------|---------|
| `user` | string | Overrides global user |
| `log_dir` | string | Overrides global log directory |
| `allow_sudo` | bool | Repo-side sudo opt-in (effective only if global also true) |
| `plugins.<name>.enabled` | bool | Plugin on/off switch |
| `plugins.<name>.folder_to_scan` | string | Folder scanned for assets (default: plugin name) |

## 3. Execution Status (`.buenosaires/status.json`)

```json
{
  "scripts": {
    "shell/deploy.sh": {
      "lint_status": "success",
      "test_status": "skipped",
      "run_status": "success",
      "timestamp": "2026-08-11T12:00:00Z",
      "overall_status": "success"
    }
  },
  "last_commit": "a1b2c3d4..."
}
```

- `Status.scripts`: map of script path → `ScriptStatus`.
- `Status.last_commit`: last fully processed commit hash; **empty ⇒ initial sync** of all existing scripts; missing base commit ⇒ fallback to initial sync.
- Status values: `pending` | `success` | `failure` | `skipped`.
- **Test phase semantics (decision)**: the shell plugin does not run a test phase; `test_status` is recorded as `skipped` and asset `tests_passed` reflects the lint outcome. This must be documented in code and README to remove the current ambiguity (finding B1).

## 4. Script Asset (`plugins/<plugin>/assets/<script-path>.json`)

```json
{
  "generation": 1,
  "last_run": "2026-08-11T12:00:00Z",
  "lint_passed": true,
  "tests_passed": true,
  "event": "Linting completed without errors. Tests passed.",
  "user": "default",
  "run_duration": "1.2s",
  "status": "success",
  "commit_hash": "a1b2c3d4..."
}
```

- `generation`: incremented once per asset modification (per run that touched the asset).
- `run_duration`: custom JSON encoding as duration string (`1.2s`); numeric JSON input must NOT be interpreted as nanoseconds (fix finding I1 — string form only).
- Location mirrors script path: `plugins/shell/assets/shell/deploy.sh.json`.

## 5. Log Entry (`<log_dir>/<script-basename>.log`)

Two sections, concatenated:

```text
--- LINT OUTPUT ---
<bash -n + shellcheck output>

--- EXECUTION OUTPUT ---
<script stdout+stderr>
```

- File mode 0600; directory mode 0750.
- Note: the log filename uses the script **basename** (folder prefix is lost); the web asset lookup searches the assets tree by basename to compensate.

## 6. Relationships

```
GlobalConfig 1──* RepoConfig (per monitored repo)
RepoConfig 1──* PluginConfig (per plugin)
Status 1──* ScriptStatus (per script path)
Script (in git tree) 1──1 Asset (per plugin assets dir)
Script 1──* LogEntry (per execution, last-write-wins)
```
