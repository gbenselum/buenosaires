# Shell Plugin Documentation

## Overview

The Shell plugin is the core automation component of Buenos Aires, enabling GitOps-based shell script execution. It automatically detects, validates, and executes shell scripts committed to your repository, providing a secure and auditable way to manage infrastructure automation.

## Features

### Core Capabilities

- ✅ **Automatic Detection**: Monitors repository for new `.sh` files
- ✅ **Syntax Validation**: Checks script syntax with `bash -n`
- ✅ **Linting**: Validates best practices with `shellcheck`
- ✅ **Safe Execution**: Configurable sudo support with audit logging
- ✅ **Status Tracking**: Full integration with Buenos Aires status system
- ✅ **Comprehensive Logging**: Captures lint and execution output

### Supported File Types

- `.sh` files (Bash shell scripts)

## How It Works

### Detection

When a commit is pushed to the monitored branch, Buenos Aires:

1. Scans for new files with `.sh` extension
2. Only processes INSERT actions (new files)
3. Skips scripts already successfully processed

### Processing Pipeline

For each detected shell script:

1. **Syntax Validation**
   - Run `bash -n` to check for syntax errors
   - Fail fast on invalid syntax
   - Prevent execution of broken scripts

2. **Linting** (if shellcheck is installed)
   - Check for common mistakes
   - Identify potential bugs
   - Enforce best practices
   - Allow warnings (exit code 1) to pass
   - Fail on errors (exit code > 1)

3. **Execution**
   - Run with `bash` (or `sudo bash` if allowed)
   - Capture all output (stdout + stderr)
   - Log execution results

4. **Logging**
   - Save lint output
   - Save execution output
   - Write to `<logdir>/<scriptname>.log`

5. **Status Tracking**
   - Update status in `.buenosaires/status.json`
   - Track generation, timestamps, and status
   - Record success/failure for each phase

## Configuration

### Global Configuration

Set during installation (`buenosaires install`):

```toml
# ~/.buenosaires/config.toml
user = "appuser"
log_dir = "/var/log/buenosaires"
branch = "main"

[plugins]
shell = true
```

### Repository Configuration

Add to your repository's `config.toml`:

```toml
# The user to run the scripts as
user = "default"

# The directory to save logs to, relative to the repository root
log_dir = "logs"

# Whether to allow scripts to be run with sudo
allow_sudo = false
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `user` | string | - | User to run scripts as |
| `log_dir` | string | - | Directory for log files |
| `allow_sudo` | bool | `false` | Allow sudo execution |

## GitOps Workflow

### Example 1: Deploy Application

1. **Create deployment script**:
   ```bash
   # deploy.sh
   #!/bin/bash
   set -euo pipefail
   
   echo "Deploying application..."
   cd /opt/myapp
   git pull origin main
   npm install
   pm2 reload myapp
   echo "Deployment complete!"
   ```

2. **Commit and push**:
   ```bash
   git add deploy.sh
   git commit -m "Add deployment script"
   git push origin main
   ```

3. **Buenos Aires will**:
   - Detect `deploy.sh`
   - Check syntax with `bash -n`
   - Lint with `shellcheck`
   - Execute the deployment
   - Log all output to `logs/deploy.sh.log`
   - Track status with generation counter

### Example 2: System Maintenance

```bash
#!/bin/bash
# cleanup.sh - Clean up old logs and temporary files

set -euo pipefail

# Clean old log files (older than 30 days)
find /var/log/myapp -name "*.log" -mtime +30 -delete

# Clean temporary files
rm -rf /tmp/myapp-*

# Clear apt cache
apt-get clean

echo "Cleanup completed successfully"
```

**Commit and push** → Buenos Aires executes → Logs saved → Status tracked

### Example 3: Configuration Update

```bash
#!/bin/bash
# update-nginx-config.sh

set -euo pipefail

# Backup current config
cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup

# Update configuration
cat > /etc/nginx/sites-available/myapp <<'EOF'
server {
    listen 80;
    server_name example.com;
    
    location / {
        proxy_pass http://localhost:3000;
    }
}
EOF

# Test configuration
nginx -t

# Reload nginx
systemctl reload nginx

echo "Nginx configuration updated"
```

**Note**: This requires `allow_sudo = true` for system file access.

## Validation with Shellcheck

### Installation

```bash
# macOS
brew install shellcheck

# Ubuntu/Debian
apt-get install shellcheck

# Fedora/RHEL
dnf install shellcheck

# From source
wget -qO- "https://github.com/koalaman/shellcheck/releases/download/stable/shellcheck-stable.linux.x86_64.tar.xz" | tar -xJv
cp shellcheck-stable/shellcheck /usr/local/bin/
```

### What Shellcheck Validates

- **SC2086**: Double quote to prevent word splitting
- **SC2164**: Use `cd ... || exit` in case cd fails
- **SC2046**: Quote to prevent word splitting
- **SC2155**: Declare and assign separately
- **SC2034**: Variable appears unused
- **And 200+ more checks...**

### Example Linting Output

**Script with issues**:
```bash
#!/bin/bash
cd /tmp
file=$1
cat $file
```

**Shellcheck output**:
```
deploy.sh:2:1: note: Use 'cd ... || exit' or 'cd ... || return' in case cd fails. [SC2164]
deploy.sh:4:5: warning: Double quote to prevent globbing and word splitting. [SC2086]
```

**Fixed script**:
```bash
#!/bin/bash
cd /tmp || exit 1
file="$1"
cat "$file"
```

### Handling Warnings vs Errors

The shell plugin treats shellcheck exit codes intelligently:

- **Exit 0**: No issues → Continue
- **Exit 1**: Warnings only → Continue with warnings logged
- **Exit > 1**: Errors found → Fail validation, don't execute

## Sudo Execution

### Security Considerations

⚠️ **WARNING**: `allow_sudo` should only be enabled in controlled environments.

**Risks**:
- Scripts run with full system privileges
- Malicious scripts can compromise the entire system
- Accidental commands can damage the system

**When to enable**:
- Trusted repository with access controls
- All contributors are verified
- Scripts are code-reviewed
- Audit logging is monitored

### Audit Logging

When `allow_sudo = true`, Buenos Aires logs:

```
WARNING: Executing script deploy.sh with sudo privileges
```

This appears in:
- Console output
- Status tracking
- System logs (if configured)

### Best Practices for Sudo Scripts

1. **Use `set -euo pipefail`**
   ```bash
   #!/bin/bash
   set -euo pipefail  # Exit on error, undefined vars, pipe failures
   ```

2. **Validate inputs**
   ```bash
   if [ $# -eq 0 ]; then
       echo "Usage: $0 <environment>"
       exit 1
   fi
   ```

3. **Use explicit paths**
   ```bash
   /usr/bin/systemctl restart nginx
   # Not: systemctl restart nginx
   ```

4. **Check before destructive operations**
   ```bash
   if [ -f /etc/important.conf ]; then
       cp /etc/important.conf /etc/important.conf.backup
   fi
   ```

5. **Log all actions**
   ```bash
   echo "$(date): Starting deployment" | tee -a /var/log/deploy.log
   ```

## Status Tracking

### Status File

Scripts are tracked in `.buenosaires/status.json`:

```json
{
  "scripts": {
    "deploy.sh": {
      "lint_status": "success",
      "test_status": "skipped",
      "run_status": "success",
      "timestamp": "2025-10-14T10:30:00Z",
      "overall_status": "success",
      "generation": 5,
      "first_deploy_date": "2025-10-01T14:20:00Z",
      "current_version_date": "2025-10-14T10:30:00Z"
    }
  }
}
```

### Fields Explained

- **lint_status**: Result of shellcheck validation
  - `pending`: Not yet validated
  - `success`: Passed validation
  - `failure`: Failed validation

- **test_status**: Currently `skipped` (reserved for future use)

- **run_status**: Result of script execution
  - `pending`: Not yet executed
  - `success`: Executed successfully
  - `failure`: Execution failed

- **overall_status**: Combined status of all phases

- **generation**: Number of times script has been updated/redeployed

- **first_deploy_date**: When script was first deployed

- **current_version_date**: When current version was deployed

- **timestamp**: Last status update time

### Status Lifecycle

```
New Script Added
    ↓
lint_status: pending
run_status: pending
overall_status: pending
    ↓
Validation
    ↓
lint_status: success/failure
    ↓
    └─→ If failure: overall_status: failure (stop)
    └─→ If success: Continue
        ↓
    Execution
        ↓
    run_status: success/failure
    overall_status: success/failure
        ↓
    Generation++
    current_version_date updated
```

## Logging

### Log Files

Logs are saved to the configured log directory:

```
logs/
├── deploy.sh.log
├── cleanup.sh.log
└── backup.sh.log
```

### Log Format

```
--- LINT OUTPUT ---
deploy.sh:10:5: note: Double quote to prevent globbing [SC2086]
Syntax check passed.
Linting completed.

--- EXECUTION OUTPUT ---
Deploying application...
Pulling latest code...
Installing dependencies...
Reloading application...
Deployment complete!
```

### Accessing Logs

**Via filesystem**:
```bash
cat logs/deploy.sh.log
```

**Via Web GUI** (if enabled):
```
http://localhost:9099/logs/deploy.sh.log
```

## Size Limits

### Maximum Script Size

Scripts are limited to **10MB** to prevent:
- Memory exhaustion
- Resource abuse
- Accidental large file commits

If a script exceeds this limit:
```
Script deploy.sh exceeds maximum size of 10485760 bytes
```

The script will be marked as failed and not executed.

## Error Handling

### Common Errors and Solutions

#### 1. Syntax Error

**Error**:
```
deploy.sh: line 5: syntax error near unexpected token `}'
syntax check failed
```

**Solution**: Fix the shell syntax
```bash
# Before (incorrect)
if [ -f file ]; then
    echo "exists"
}

# After (correct)
if [ -f file ]; then
    echo "exists"
fi
```

#### 2. Shellcheck Not Found

**Warning**:
```
Warning: shellcheck not found, skipping linting
```

**Solution**: Install shellcheck (see Installation section)

#### 3. Execution Failure

**Error**:
```
Failed to execute script deploy.sh: exit status 1
```

**Solution**: 
- Check script logs in `logs/deploy.sh.log`
- Review the execution output
- Fix any runtime errors in the script

#### 4. Permission Denied

**Error**:
```
bash: /opt/app/deploy.sh: Permission denied
```

**Solution**: Script file must be readable (created from git content automatically)

## Advanced Usage

### Script Templates

**Basic template**:
```bash
#!/bin/bash
set -euo pipefail

# Description: [What this script does]
# Requirements: [Any dependencies needed]
# Usage: [How to use this script]

main() {
    echo "Starting..."
    
    # Your code here
    
    echo "Complete!"
}

main "$@"
```

**Error handling template**:
```bash
#!/bin/bash
set -euo pipefail

# Trap errors
trap 'echo "Error on line $LINENO"' ERR

# Cleanup on exit
cleanup() {
    echo "Cleaning up..."
    # Cleanup code
}
trap cleanup EXIT

main() {
    # Your code here
}

main "$@"
```

### Environment Variables

Scripts can access standard environment variables:

```bash
#!/bin/bash
set -euo pipefail

echo "User: $USER"
echo "Home: $HOME"
echo "Path: $PATH"
echo "Working directory: $PWD"
```

### Script Arguments

Currently, scripts run without arguments. To add argument support, modify the repository configuration or use environment variables.

## Security Best Practices

### 1. Code Review

- Review all scripts before committing
- Use pull requests for script changes
- Require approval from team members

### 2. Access Control

- Limit repository write access
- Use branch protection rules
- Enable required status checks

### 3. Secrets Management

❌ **Never hardcode secrets**:
```bash
# BAD
PASSWORD="supersecret123"
```

✅ **Use environment variables**:
```bash
# GOOD
PASSWORD="${DB_PASSWORD:?DB_PASSWORD not set}"
```

✅ **Use secret management tools**:
```bash
# GOOD
PASSWORD=$(vault kv get -field=password secret/db)
```

### 4. Input Validation

```bash
#!/bin/bash
set -euo pipefail

ENV="${1:-}"

# Validate input
case "$ENV" in
    prod|staging|dev)
        echo "Deploying to $ENV"
        ;;
    *)
        echo "Invalid environment: $ENV"
        echo "Usage: $0 {prod|staging|dev}"
        exit 1
        ;;
esac
```

### 5. Idempotency

Make scripts safe to run multiple times:

```bash
#!/bin/bash
set -euo pipefail

# Check if already configured
if [ -f /etc/myapp/configured ]; then
    echo "Already configured, skipping"
    exit 0
fi

# Configure application
configure_app

# Mark as configured
touch /etc/myapp/configured
```

## Troubleshooting

### Script Not Detected

**Problem**: Script committed but not executed

**Solutions**:
1. Verify filename ends with `.sh`
2. Ensure file is committed (not just staged)
3. Check file is a new file (INSERT action)
4. Verify repository is on the monitored branch
5. Check Buenos Aires is running: `ps aux | grep buenosaires`

### Script Fails Validation

**Problem**: Shellcheck reports errors

**Solutions**:
1. Review shellcheck output in logs
2. Fix reported issues
3. Test locally: `shellcheck deploy.sh`
4. Commit fixed version

### Script Execution Hangs

**Problem**: Script never completes

**Solutions**:
1. Check for interactive prompts
2. Add timeout to long-running commands
3. Use `-y` flags for automatic confirmation
4. Review script logic for infinite loops

### Permission Issues with Sudo

**Problem**: Script needs root but `allow_sudo = false`

**Solutions**:
1. Enable `allow_sudo` in config (if appropriate)
2. Restructure script to not require sudo
3. Use pre-configured permissions/capabilities
4. Run Buenos Aires itself with appropriate permissions

## Testing

### Manual Testing

**1. Test syntax locally**:
```bash
bash -n deploy.sh
```

**2. Test with shellcheck**:
```bash
shellcheck deploy.sh
```

**3. Test execution**:
```bash
bash deploy.sh
```

**4. Test in Buenos Aires**:
```bash
# Commit and monitor
git add deploy.sh
git commit -m "Test deployment script"
git push origin main

# Watch logs
tail -f logs/deploy.sh.log
```

### Automated Testing

**Unit tests for scripts**:
```bash
#!/bin/bash
# test_deploy.sh

source deploy.sh

# Test individual functions
test_function_name() {
    result=$(function_name)
    if [ "$result" = "expected" ]; then
        echo "PASS: function_name"
    else
        echo "FAIL: function_name"
        return 1
    fi
}

test_function_name
```

## Integration with CI/CD

### Pre-commit Hooks

Validate scripts before committing:

```bash
# .git/hooks/pre-commit
#!/bin/bash

# Check all .sh files
for file in $(git diff --cached --name-only --diff-filter=ACM | grep '\.sh$'); do
    # Syntax check
    if ! bash -n "$file"; then
        echo "Syntax error in $file"
        exit 1
    fi
    
    # Shellcheck
    if command -v shellcheck &> /dev/null; then
        if ! shellcheck "$file"; then
            echo "Shellcheck failed for $file"
            exit 1
        fi
    fi
done
```

### GitHub Actions

```yaml
# .github/workflows/validate-scripts.yml
name: Validate Shell Scripts

on: [push, pull_request]

jobs:
  shellcheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run ShellCheck
        uses: ludeeus/action-shellcheck@master
        with:
          scandir: '.'
```

## API Reference

### ShellPlugin Methods

#### `LintAndValidate(scriptPath string) (string, error)`

Validates a shell script using bash syntax check and shellcheck.

**Parameters**:
- `scriptPath`: Path to shell script file

**Returns**:
- `string`: Combined output from bash and shellcheck
- `error`: Error if validation fails

**Example**:
```go
plugin := shell.ShellPlugin{}
output, err := plugin.LintAndValidate("/path/to/script.sh")
if err != nil {
    log.Printf("Validation failed: %v\n%s", err, output)
}
```

#### `Run(scriptPath string, allowSudo bool) (string, error)`

Executes a shell script.

**Parameters**:
- `scriptPath`: Path to shell script file
- `allowSudo`: Whether to run with sudo privileges

**Returns**:
- `string`: Combined stdout and stderr output
- `error`: Error if execution fails

**Example**:
```go
plugin := shell.ShellPlugin{}
output, err := plugin.Run("/path/to/script.sh", false)
if err != nil {
    log.Printf("Execution failed: %v\n%s", err, output)
}
```

## Performance Considerations

### Execution Time

- Scripts are executed sequentially
- Default timeout: 5 minutes (configurable)
- Long-running scripts should be backgrounded

### Resource Usage

- Memory: Limited by script content and output
- CPU: Limited by script operations
- Disk: Log files accumulate over time

### Optimization Tips

1. **Keep scripts focused**: One task per script
2. **Use efficient commands**: Prefer built-ins over external commands
3. **Clean up**: Remove old log files periodically
4. **Background long tasks**: Use systemd or cron for ongoing processes

## Migration Guide

### From Manual Execution

**Before**:
```bash
ssh server
sudo bash deploy.sh
```

**After**:
1. Add script to repository
2. Configure `config.toml`
3. Commit and push
4. Buenos Aires handles execution

### From Other Automation Tools

**From Ansible**:
- Convert playbooks to shell scripts
- Use shell script best practices
- Leverage Buenos Aires validation

**From Jenkins**:
- Move pipeline steps to shell scripts
- Commit scripts to repository
- Configure Buenos Aires monitoring

## Future Enhancements

Potential improvements for future versions:

- [ ] Script arguments/parameters support
- [ ] Environment variable injection from config
- [ ] Parallel script execution
- [ ] Script dependencies/ordering
- [ ] Rollback on failure
- [ ] Dry-run mode
- [ ] Script templates/generators
- [ ] Integration with testing frameworks

## Contributing

To contribute to the shell plugin:

1. Follow existing code structure in `plugins/shell/`
2. Add tests for new features
3. Update this documentation
4. Ensure backward compatibility
5. Follow Go best practices

## License

Same as Buenos Aires main project.
