# Plugin & Asset Contract

**Feature**: [spec.md](../spec.md) | **File**: contracts/plugin-contract.md

## Plugin Interface

Every plugin MUST:

1. Expose an `enabled` boolean in its `[plugins.<name>]` repo-config section (constitution §VII).
2. Support a `folder_to_scan` key; when empty, default to the plugin's own name (e.g. `shell`).
3. Track each managed script/file in a per-asset JSON file under `plugins/<name>/assets/`, mirroring the script's repo-relative path (`plugins/shell/assets/shell/deploy.sh.json`).
4. Implement validation (lint) and execution phases; only scripts passing validation are executed.
5. Write asset metadata with the schema below.

## Asset JSON Schema

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

| Field | Contract |
|-------|----------|
| `generation` | Integer; incremented by 1 each time the asset is modified by a run |
| `last_run` | RFC3339 timestamp of the most recent processing |
| `lint_passed` | Bool; outcome of the validation phase |
| `tests_passed` | Bool; `true` for the shell plugin (no test runner; mirrors lint outcome) |
| `event` | Free text: validation/lint output or execution output |
| `user` | Effective user from config |
| `run_duration` | Duration string (`1.2s`) — **string form only**; numeric JSON values are rejected, never interpreted as nanoseconds (closes finding I1) |
| `status` | `pending`/`success`/`failure` for the run phase |
| `commit_hash` | SHA of the commit that introduced the processed version |

## Path Safety (non-negotiable, constitution §II)

- Asset paths are derived from git-tree paths; `..` components, absolute paths, and empty names MUST be rejected.
- Web asset lookup matches by basename across the plugin's assets tree; the requested name MUST be a single path segment (no `/`, `\`, `..`).
