# Configuration Contract

**Feature**: [spec.md](../spec.md) | **File**: contracts/config-contract.md

## Global Configuration (`~/.buenosaires/config.toml`)

Created by `buenosaires install`. Loaded by `run` — a missing or unparsable file is a **fatal** startup error with a clear message.

```toml
user = "default"          # required non-empty
log_dir = "logs"          # optional; relative to repo root
branch = "main"           # required non-empty
sync_interval = 180       # optional; 0 ⇒ 180
repository_url = "https://github.com/owner/repo.git"  # required for first clone
allow_sudo = false        # host-operator gate (two-gate model)

[gui]
enabled = false
port = 9099               # used when enabled
```

## Repository Configuration (`config.toml` in repo root)

Loaded each poll cycle; missing file is NOT an error (defaults apply). Values override global defaults when non-empty.

```toml
user = "default"      # optional override
log_dir = "logs"      # optional override
allow_sudo = false    # repo opt-in — effective ONLY if global allow_sudo is also true

[plugins.shell]
enabled = true
folder_to_scan = "shell"   # optional; default = plugin name
```

## Validation Rules (to be implemented — closes finding U1)

- Global config: `user` and `branch` MUST be non-empty; `repository_url` MUST be non-empty when the workspace contains no repository to open.
- `gui.port` MUST be in range 1–65535 (invalid ⇒ warn and default to 9099).
- Unknown plugin keys are ignored (forward compatibility).
- `folder_to_scan` empty ⇒ plugin name (e.g. `shell`).

## Install Command Contract

Prompts: username, log folder, branch, repository URL (default `https://github.com/gbenselum/buenosaires_test`), web GUI enable (y/n), port (default 9099). All prompts accept empty input where a default exists; stdin read errors MUST be surfaced, not silently ignored (closes finding in install.go).
