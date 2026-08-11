# Quickstart: Core Rebuild — Local Validation Guide

**Feature**: [spec.md](./spec.md)

## Prerequisites

- Go 1.24+, `bash`, `shellcheck`, `golangci-lint`, `gosec`, `pre-commit` (`brew install go shellcheck golangci-lint gosec pre-commit`)
- A test repository with a `config.toml` and at least one `.sh` file under the scanned folder (e.g. `shell/hello.sh`)

## Validation Flow (mirrors the spec's acceptance scenarios)

### 1. Install

```bash
make check            # gate: lint + security + tests all green
go build -o buenosaires . || go run . install
buenosaires install   # answer prompts (user, logs, branch main, repo URL, GUI y/n)
```

**Verify**: `~/.buenosaires/config.toml` exists with your answers.

### 2. First run / initial sync

```bash
mkdir -p /tmp/ba-workspace && cd /tmp/ba-workspace
buenosaires run       # clones the repo, performs initial sync
```

**Verify**: every existing `.sh` in the scanned folder is linted + executed once; `.buenosaires/status.json` created with per-script entries and `last_commit` set.

### 3. Script lifecycle (SC-002)

- **Insert**: push a valid script → it executes once; push the same script again unchanged → no re-run.
- **Modify**: change the script and push → re-validated and re-executed.
- **Invalid**: push a script with a syntax error → recorded failure, never executed; push a fix → executed.
- **Delete**: remove the script and push → its status entry is removed.

### 4. Sudo two-gate (SC-003)

| global allow_sudo | repo allow_sudo | Result |
|-------------------|-----------------|--------|
| false | false | no sudo |
| false | true | no sudo |
| true | false | no sudo |
| true | true | sudo |

### 5. Resilience (SC-006)

- Dirty worktree or failed `git pull` → monitoring continues.
- Failed fetch → retries next cycle.
- Web server failure → logged, monitor continues.
- Restart the monitor → no re-run of succeeded inserts (SC-004).

### 6. Web UI (SC-005)

Enable GUI in global config, `buenosaires run`, open `http://localhost:<port>`:
- Log list renders (empty list when nothing has run yet).
- Open a log → content + asset JSON accordion.
- Security headers present; traversal URLs rejected (400).

### 7. Container (SC-007)

```bash
docker build -t buenosaires .
docker run -it --rm --user "$(id -u):$(id -g)" \
  -v $(pwd):/app -v ~/.buenosaires:/home/appuser/.buenosaires \
  buenosaires run
```

**Verify**: runs as non-root; git/bash/shellcheck present.

## Teardown

```bash
rm -rf /tmp/ba-workspace
```
