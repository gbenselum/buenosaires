# Configuration Guide

## Overview

Buenos Aires uses TOML configuration files for both global and repository-specific settings. This guide explains all configuration options and provides examples.

## Configuration Files

### Global Configuration

**Location**: `~/.buenosaires/config.toml`  
**Created by**: `buenosaires install` command  
**Scope**: User-wide settings

### Repository Configuration

**Location**: `<repository>/config.toml`  
**Created by**: Manual creation  
**Scope**: Repository-specific overrides

## Global Configuration Reference

### Complete Example

```toml
# ~/.buenosaires/config.toml

# User to run scripts as
user = "appuser"

# Default log directory (can be overridden per-repo)
log_dir = "/var/log/buenosaires"

# Git branch to monitor
branch = "main"

# Enabled plugins
[plugins]
shell = true
docker = true

# Web GUI configuration
[gui]
enabled = true
port = 9099
```

### Global Configuration Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `user` | string | Yes | - | Default user for script execution |
| `log_dir` | string | Yes | - | Default log directory path |
| `branch` | string | Yes | - | Git branch to monitor |
| `plugins.shell` | bool | No | `true` | Enable shell plugin |
| `plugins.docker` | bool | No | `true` | Enable Docker plugin |
| `gui.enabled` | bool | No | `false` | Enable web GUI |
| `gui.port` | int | No | `8080` | Web GUI port |

## Repository Configuration Reference

### Complete Example

```toml
# <repository>/config.toml

# User to run scripts as (overrides global)
user = "default"

# Log directory relative to repository root
log_dir = "logs"

# Allow sudo execution (security risk!)
allow_sudo = false

# Docker plugin configuration
[docker]
# Enable Docker plugin for this repository
enabled = true

# Automatically run containers after building
# WARNING: Security risk - only enable in trusted repos
auto_run = false

# Default tag for Docker images
default_tag = "latest"

# Prefix for image names (e.g., "mycompany/")
image_prefix = ""
```

### Repository Configuration Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `user` | string | No | Global user | User for script execution |
| `log_dir` | string | No | Global log_dir | Log directory path |
| `allow_sudo` | bool | No | `false` | Allow sudo execution |
| `docker.enabled` | bool | No | `true` | Enable Docker plugin |
| `docker.auto_run` | bool | No | `false` | Auto-run containers |
| `docker.default_tag` | string | No | `"latest"` | Default image tag |
| `docker.image_prefix` | string | No | `""` | Image name prefix |

## Configuration Examples

### Example 1: Simple Development Setup

**Global** (`~/.buenosaires/config.toml`):
```toml
user = "developer"
log_dir = "~/buenosaires-logs"
branch = "main"

[plugins]
shell = true
docker = true

[gui]
enabled = true
port = 9099
```

**Repository** (`config.toml`):
```toml
user = "default"
log_dir = "logs"
allow_sudo = false

[docker]
enabled = true
auto_run = false
default_tag = "dev"
image_prefix = "myapp/"
```

### Example 2: Production Setup

**Global**:
```toml
user = "buenosaires"
log_dir = "/var/log/buenosaires"
branch = "production"

[plugins]
shell = true
docker = true

[gui]
enabled = true
port = 8080
```

**Repository**:
```toml
user = "deployment"
log_dir = "/var/log/deployment"
allow_sudo = true  # Carefully controlled

[docker]
enabled = true
auto_run = true    # Auto-deploy in production
default_tag = "prod"
image_prefix = "registry.company.com/prod/"
```

### Example 3: Docker-Only Setup

**Global**:
```toml
user = "docker-user"
log_dir = "/var/log/containers"
branch = "main"

[plugins]
shell = false  # Disable shell scripts
docker = true

[gui]
enabled = false  # No web GUI needed
```

**Repository**:
```toml
log_dir = "container-logs"

[docker]
enabled = true
auto_run = false
default_tag = "latest"
image_prefix = "mycompany/"
```

### Example 4: Multi-Environment

**Staging Repository** (`config.toml`):
```toml
log_dir = "logs/staging"
allow_sudo = false

[docker]
enabled = true
auto_run = true  # Auto-deploy to staging
default_tag = "staging"
image_prefix = "staging/"
```

**Production Repository** (`config.toml`):
```toml
log_dir = "logs/production"
allow_sudo = true

[docker]
enabled = true
auto_run = false  # Manual deploy in prod
default_tag = "production"
image_prefix = "prod/"
```

## Configuration Precedence

When settings exist in both global and repository config:

1. **Repository config takes precedence** for:
   - `user`
   - `log_dir`
   - `allow_sudo`
   - All Docker settings

2. **Global config is always used** for:
   - `branch` (monitoring branch)
   - `plugins` (which plugins are available)
   - `gui` (web interface)

## Docker Configuration Deep Dive

### enabled

**Type**: `bool`  
**Default**: `true`

Controls whether the Docker plugin processes containers in this repository.

```toml
[docker]
enabled = false  # Skip all Dockerfile processing
```

**Use cases**:
- Temporarily disable container deployments
- Shell-only repositories
- Debugging

### auto_run

**Type**: `bool`  
**Default**: `false`

⚠️ **Security Warning**: Automatically runs containers after building.

```toml
[docker]
auto_run = true  # Containers start automatically
```

**When enabled**:
- Builds Docker image
- Removes old container (if exists)
- Starts new container in detached mode

**When disabled** (default):
- Only builds Docker image
- Image ready for manual deployment

**Security implications**:
- Untrusted Dockerfiles could exploit system
- Use only in controlled environments
- Review all Dockerfiles before enabling

### default_tag

**Type**: `string`  
**Default**: `"latest"`

Tag applied to built Docker images.

```toml
[docker]
default_tag = "v1.2.3"
```

**Results in**:
- `Containers/webapp/Dockerfile` → `webapp:v1.2.3`

**Use cases**:
- Version tags: `"v1.0.0"`
- Environment tags: `"staging"`, `"production"`
- Build identifiers: `"build-123"`

### image_prefix

**Type**: `string`  
**Default**: `""`

Prefix added to all image names.

```toml
[docker]
image_prefix = "mycompany/"
```

**Results in**:
- `Containers/webapp/Dockerfile` → `mycompany/webapp:latest`

**Use cases**:
- Docker Hub usernames: `"username/"`
- Private registries: `"registry.company.com/team/"`
- Organization namespaces: `"myorg/"`

## Creating Configuration Files

### Interactive Installation

```bash
$ buenosaires install
Enter your username: appuser
Enter the folder to save logs: /var/log/buenosaires
Enter the branch to monitor (e.g., main): main
Enable Web GUI? (y/n): y
Enter the port for the Web GUI (e.g., 9099): 9099
Configuration saved successfully!
```

Creates: `~/.buenosaires/config.toml`

### Manual Repository Config

```bash
cd /path/to/repository
cat > config.toml <<EOF
user = "default"
log_dir = "logs"
allow_sudo = false

[docker]
enabled = true
auto_run = false
default_tag = "latest"
image_prefix = ""
EOF
```

## Configuration Validation

### Testing Configuration

```bash
# Test loading config
buenosaires run

# Check logs for errors
tail -f logs/*.log
```

### Common Issues

**1. Invalid TOML syntax**
```
Error: Near line 5: Expected key but found ']' instead
```

**Fix**: Check TOML syntax, ensure proper quoting

**2. Port already in use**
```
Failed to start web server: listen tcp :9099: bind: address already in use
```

**Fix**: Change GUI port or stop conflicting service

**3. Permission denied on log directory**
```
Failed to create log directory: permission denied
```

**Fix**: Ensure directory exists and is writable

## Security Best Practices

### Sudo Configuration

❌ **Avoid**:
```toml
allow_sudo = true  # Dangerous in untrusted repos
```

✅ **Better**:
```toml
allow_sudo = false  # Default, safer
```

Only enable sudo when:
- Repository is fully trusted
- All contributors are verified
- Scripts are code-reviewed
- Audit logs are monitored

### Docker Auto-Run

❌ **Avoid**:
```toml
[docker]
auto_run = true  # Risk of malicious containers
```

✅ **Better**:
```toml
[docker]
auto_run = false  # Build only, manual deployment
```

### Log Directory Permissions

✅ **Recommended**:
```toml
log_dir = "/var/log/buenosaires"  # Restricted directory
# Ensure: drwx------ (700) permissions
```

❌ **Avoid**:
```toml
log_dir = "/tmp/logs"  # World-writable, insecure
```

## Environment-Specific Configurations

### Development

```toml
[docker]
enabled = true
auto_run = false  # Manual testing
default_tag = "dev"
image_prefix = "dev/"
```

### Staging

```toml
[docker]
enabled = true
auto_run = true   # Auto-deploy for testing
default_tag = "staging"
image_prefix = "staging/"
```

### Production

```toml
[docker]
enabled = true
auto_run = false  # Controlled deployment
default_tag = "prod"
image_prefix = "registry.company.com/prod/"
```

## Troubleshooting

### Configuration Not Found

```
Failed to load global config: no such file or directory
```

**Solution**: Run `buenosaires install`

### Repository Config Ignored

**Check**:
1. File named exactly `config.toml`
2. In repository root
3. Valid TOML syntax
4. No merge conflicts

### Docker Settings Not Applied

**Check**:
1. Docker plugin enabled in global config
2. `docker.enabled = true` in repo config
3. Containers directory exists

## Advanced Configuration

### Multiple Repositories

Each repository can have its own `config.toml`:

```
~/projects/
├── app1/
│   ├── config.toml  # app1 settings
│   └── Containers/
├── app2/
│   ├── config.toml  # app2 settings
│   └── Containers/
└── app3/
    └── config.toml  # app3 settings (no containers)
```

### Shared Configuration

Use symbolic links for shared configs:

```bash
# Create base config
cat > ~/base-config.toml <<EOF
log_dir = "logs"
allow_sudo = false
EOF

# Link in repositories
cd /path/to/repo1
ln -s ~/base-config.toml config.toml
```

---

**Last Updated**: 2025-10-14
