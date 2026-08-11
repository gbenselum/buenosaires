# Status Contract

**Feature**: [spec.md](../spec.md) | **File**: contracts/status-contract.md

## File: `.buenosaires/status.json`

```json
{
  "scripts": {
    "<script-path>": {
      "lint_status": "pending|success|failure",
      "test_status": "pending|success|failure|skipped",
      "run_status": "pending|success|failure",
      "timestamp": "<RFC3339>",
      "overall_status": "pending|success|failure"
    }
  },
  "last_commit": "<sha> | omitted"
}
```

## Semantics

- `last_commit` empty or absent ⇒ **initial sync**: every `.sh` file under the scanned folder is processed.
- Base commit referenced by `last_commit` unavailable (shallow clone/GC) ⇒ fallback to initial sync, never fatal.
- **Insert**: process unless that script's `overall_status` is already `success` (restart/initial-sync dedupe).
- **Modify**: always re-process (fixes and new versions re-deploy).
- **Delete**: remove the script's status entry.
- Path keys are repo-relative (`shell/deploy.sh`); the path is sanitized on every read/write (`sanitizeRepoPath` — traversal MUST be rejected).
- File mode 0600; directory `.buenosaires/` mode 0750.
- Corrupt/truncated JSON: MUST NOT silently wipe other scripts' state — treat as recoverable error (log + preserve backup or start empty with warning per decision in tasks).

## Lifecycle Transitions (single script)

```
pending → (lint) → failure | success
success → (run)  → success | failure   (final overall = run outcome)
```
Test phase is `skipped` for the shell plugin (no test runner exists); asset `tests_passed` mirrors lint outcome (documented decision — removes finding B1/I2).
