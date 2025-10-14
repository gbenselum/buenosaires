# Quick Reference Guide

## Installation

```bash
# Install Buenos Aires
go install github.com/gbenselum/buenosaires@latest

# Run interactive setup
buenosaires install

# Output: Creates ~/.buenosaires/config.toml
```

## Basic Commands

```bash
# Start monitoring
buenosaires run

# Get help
buenosaires --help
buenosaires install --help
buenosaires run --help
```

## Directory Structure

```
myrepo/
├── config.toml              # Repository configuration
├── deploy.sh                # Shell scripts (monitored)
├── Containers/              # Container definitions
│   ├── webapp/
│   │   └── Dockerfile
│   └── api/
│       └── Containerfile
├── logs/                    # Execution logs
└── .buenosaires/
    └── status.json          # Status tracking
```

## Configuration

### Minimal Repository Config

```toml
# config.toml
log_dir = "logs"
allow_sudo = false

[docker]
enabled = true
auto_run = false
```

### Global Config Location

```
~/.buenosaires/config.toml
```

## Plugin Usage

### Shell Plugin

**Create script**:
```bash
#!/bin/bash
set -euo pipefail
echo "Hello from Buenos Aires!"
```

**Commit and push**:
```bash
git add deploy.sh
git commit -m "Add deployment script"
git push origin main
```

**Result**: Script validated and executed automatically

### Docker Plugin

**Create container**:
```bash
mkdir -p Containers/webapp
cat > Containers/webapp/Dockerfile <<EOF
FROM nginx:alpine
COPY index.html /usr/share/nginx/html/
EOF
```

**Commit and push**:
```bash
git add Containers/
git commit -m "Add webapp container"
git push origin main
```

**Result**: Image built automatically (webapp:latest)

## Status Tracking

### View Status

```bash
cat .buenosaires/status.json
```

### Status Fields

```json
{
  "scripts": {
    "deploy.sh": {
      "lint_status": "success",
      "run_status": "success",
      "overall_status": "success",
      "generation": 3,
      "first_deploy_date": "2025-10-01T10:00:00Z",
      "current_version_date": "2025-10-14T15:30:00Z"
    }
  }
}
```

## Logging

### Log Location

```
logs/<scriptname>.log
logs/<containername>-container.log
```

### View Logs

```bash
# Tail script log
tail -f logs/deploy.sh.log

# View container log
cat logs/webapp-container.log
```

### Log Format

```
--- LINT OUTPUT ---
(validation output)

--- EXECUTION OUTPUT ---
(script/build output)
```

## Web GUI

### Enable Web Interface

```toml
# ~/.buenosaires/config.toml
[gui]
enabled = true
port = 9099
```

### Access

```
http://localhost:9099
```

### Features

- List all log files
- View log contents
- Auto-refresh

## Common Patterns

### Deployment Script

```bash
#!/bin/bash
set -euo pipefail

echo "Deploying application..."

# Pull latest code
git pull origin main

# Install dependencies
npm install

# Restart service
pm2 reload myapp

echo "Deployment complete!"
```

### Multi-Stage Container

```dockerfile
# Containers/app/Dockerfile
FROM node:18 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
```

### Configuration Update Script

```bash
#!/bin/bash
set -euo pipefail

# Backup existing config
cp /etc/myapp/config.yml /etc/myapp/config.yml.backup

# Deploy new config
cat > /etc/myapp/config.yml <<EOF
server:
  port: 8080
  host: 0.0.0.0
EOF

# Reload service
systemctl reload myapp
```

## Troubleshooting

### Script Not Executing

**Check**:
```bash
# Verify file extension
ls -la *.sh

# Check Buenos Aires is running
ps aux | grep buenosaires

# View logs
tail -f logs/*.log
```

### Container Not Building

**Check**:
```bash
# Verify path
ls -la Containers/*/Dockerfile

# Check Docker plugin enabled
cat config.toml | grep docker

# Test Docker manually
docker build -t test Containers/webapp/
```

### Permission Issues

**Check**:
```bash
# Verify log directory permissions
ls -ld logs/

# Check sudo settings
grep allow_sudo config.toml

# View Buenos Aires logs
journalctl -u buenosaires -f  # (if running as service)
```

## Security Checklist

- [ ] `allow_sudo = false` (unless required)
- [ ] `auto_run = false` for Docker (unless trusted)
- [ ] Review all scripts before commit
- [ ] Use branch protection rules
- [ ] Monitor audit logs
- [ ] Restrict repository write access
- [ ] Keep log directory secure (700 permissions)
- [ ] Regular security updates

## Performance Tuning

### Adjust Poll Interval

Default: 10 seconds

**Modify in source**:
```go
// cmd/run.go
const DefaultPollInterval = 30 * time.Second  // Slower polling
```

### Manage Log Size

```bash
# Rotate logs
logrotate /var/log/buenosaires/*.log

# Delete old logs
find logs/ -mtime +30 -delete

# Compress logs
gzip logs/*.log
```

## Integration Examples

### With Systemd

```ini
# /etc/systemd/system/buenosaires.service
[Unit]
Description=Buenos Aires GitOps Monitor
After=network.target

[Service]
Type=simple
User=buenosaires
WorkingDirectory=/opt/myrepo
ExecStart=/usr/local/bin/buenosaires run
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable buenosaires
sudo systemctl start buenosaires
sudo systemctl status buenosaires
```

### With Docker

```bash
# Build image
docker build -t buenosaires .

# Run container
docker run -it --rm \
  -v $(pwd):/app \
  -v ~/.buenosaires:/home/appuser/.buenosaires \
  -v /var/run/docker.sock:/var/run/docker.sock \
  buenosaires run
```

### With GitHub Actions

```yaml
# .github/workflows/validate.yml
name: Validate Scripts

on: [push, pull_request]

jobs:
  shellcheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run ShellCheck
        run: shellcheck *.sh
      
  hadolint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Lint Dockerfiles
        uses: hadolint/hadolint-action@v2.0.0
        with:
          dockerfile: Containers/*/Dockerfile
```

## Best Practices

### Shell Scripts

✅ **Do**:
- Use `set -euo pipefail`
- Validate inputs
- Use explicit paths
- Add error handling
- Log all actions

❌ **Don't**:
- Hardcode secrets
- Ignore errors
- Use interactive prompts
- Rely on external state

### Docker Containers

✅ **Do**:
- Use specific base image tags
- Multi-stage builds
- Minimal base images
- Non-root user
- Health checks

❌ **Don't**:
- Use `latest` tag in FROM
- Run as root
- Include secrets in image
- Build unnecessary files

### GitOps Workflow

✅ **Do**:
- Code review all changes
- Test locally first
- Use feature branches
- Tag releases
- Document changes

❌ **Don't**:
- Commit directly to main
- Skip testing
- Ignore linter warnings
- Deploy without review

## Keyboard Shortcuts (Web GUI)

| Key | Action |
|-----|--------|
| ↑↓ | Navigate logs |
| Enter | Open log |
| Esc | Back to list |
| R | Refresh |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 130 | Interrupted (Ctrl+C) |

## File Locations

| Item | Location |
|------|----------|
| Binary | `/usr/local/bin/buenosaires` |
| Global Config | `~/.buenosaires/config.toml` |
| Repo Config | `<repo>/config.toml` |
| Status File | `<repo>/.buenosaires/status.json` |
| Logs | `<repo>/logs/` (configurable) |

## Environment Variables

Currently none. Future:
- `BUENOSAIRES_CONFIG` - Override config path
- `BUENOSAIRES_LOG_LEVEL` - Set log level
- `BUENOSAIRES_POLL_INTERVAL` - Override poll interval

## Version Information

```bash
# View version (future)
buenosaires version

# Check for updates
go get -u github.com/gbenselum/buenosaires
```

---

**Tip**: Bookmark this page for quick reference!

**Last Updated**: 2025-10-14
