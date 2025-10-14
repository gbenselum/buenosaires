# Pull Request: Security Fixes, Code Improvements, Enhanced JSON Structure, and Docker Plugin

## Summary

This PR addresses all critical security vulnerabilities, modernizes the codebase, enhances the JSON structure for better tracking and monitoring capabilities, **and introduces a comprehensive Docker plugin for container-based GitOps workflows**.

### 🔒 Security Improvements

- **Fixed path traversal vulnerability** in web server log viewer
  - Added `filepath.Base()` sanitization
  - Validated paths stay within log directory
  - Restricted access to `.log` files only
- **Added script size validation** (10MB limit) to prevent resource exhaustion
- **Enhanced sudo execution logging** for better audit trails
- **Improved input validation** across all user inputs

### 📊 Enhanced JSON Structure

**New fields added to `ScriptStatus`:**
- `generation` (int): Tracks deployment count, increments on each redeployment
- `first_deploy_date` (timestamp): Records when script was first deployed
- `current_version_date` (timestamp): Tracks current version deployment time

**Benefits:**
- Complete deployment history tracking
- Better audit trail for compliance
- Monitor script evolution over time
- Identify frequently updated scripts

### 🧵 Concurrency & Thread Safety

- Added `sync.RWMutex` for thread-safe status access
- New thread-safe `GetScriptStatus()` method
- Protected all status updates with proper locking

### 🔧 Code Modernization

**Deprecated Function Replacements:**
- Replaced all `io/ioutil` functions with `os` package equivalents
- `ioutil.ReadFile` → `os.ReadFile`
- `ioutil.WriteFile` → `os.WriteFile`
- `ioutil.ReadDir` → `os.ReadDir`
- `ioutil.TempFile` → `os.CreateTemp`
- `ioutil.TempDir` → `os.MkdirTemp`

**Configuration Constants Added:**
```go
DefaultPollInterval = 10 * time.Second
MaxScriptSize = 10 * 1024 * 1024  // 10MB
DefaultScriptTimeout = 5 * time.Minute
DirPerm = 0755
FilePerm = 0644
```

### 🐛 Bug Fixes

1. **Resource Leak Fixed**
   - Fixed temporary file accumulation in script processing loop
   - Moved cleanup outside of defer to execute immediately
   
2. **Error Handling Improvements**
   - Added validation for all user inputs
   - Port range validation (1024-65535)
   - Empty input detection
   - Better error messages

3. **Status File Robustness**
   - Added nil checks for Scripts map
   - Prevents potential panics

### 🐳 NEW: Docker Plugin

**Complete container-based GitOps automation!**

The new Docker plugin extends Buenos Aires with powerful container deployment capabilities:

**Features:**
- 🔍 Automatic detection of Dockerfile/Containerfile in `Containers/` folder
- ✅ Validation and linting with `hadolint` (optional)
- 🏗️ Automatic Docker image building on commit
- 🚀 Optional container deployment (configurable via `auto_run`)
- 📊 Full status tracking with generation counters
- 📝 Comprehensive build and deployment logging

**Workflow:**
1. Create `Containers/<name>/Dockerfile` in your repository
2. Commit and push to monitored branch
3. Buenos Aires automatically validates, builds, and optionally deploys
4. Track status in `.buenosaires/status.json`

**Configuration:**
```toml
[docker]
enabled = true       # Enable Docker plugin
auto_run = false     # Auto-start containers (disabled for safety)
default_tag = "latest"
image_prefix = ""    # Optional: "mycompany/"
```

**Image Naming:**
- `Containers/webapp/Dockerfile` → `webapp:latest`
- With prefix: `mycompany/webapp:latest`

**Security:**
- `auto_run` defaults to `false` for safety
- Full audit trail for all builds
- Separate status tracking for containers
- Supports both Dockerfile and Containerfile formats

See **[DOCKER_PLUGIN.md](./DOCKER_PLUGIN.md)** for complete documentation with examples, API reference, troubleshooting, and security best practices.

### 📈 Changes Overview

**Commits:**
- Security fixes and code improvements
- Enhanced JSON structure
- Docker plugin implementation
- Comprehensive documentation

**Files Changed:**
```
16 files changed, 1477 insertions(+), 52 deletions(-)
```

**Modified Files:**
- `cmd/install.go` - Enhanced input validation, Docker plugin support
- `cmd/run.go` - Security fixes, resource management, Docker integration
- `internal/status/status.go` - Enhanced JSON structure, thread safety
- `internal/web/server.go` - Path traversal fix
- `internal/config/config.go` - Docker configuration support
- `internal/config/config_test.go` - Test improvements
- `internal/web/server_test.go` - Test modernization
- `plugins/shell/shell_test.go` - Test updates
- `README.md` - Docker plugin documentation
- `config.toml.example` - Docker configuration examples

**New Files:**
- `plugins/docker/docker.go` - Docker plugin implementation
- `plugins/docker/docker_test.go` - Docker plugin tests
- `IMPROVEMENTS.md` - Security and code improvements documentation
- `DOCKER_PLUGIN.md` - Complete Docker plugin documentation
- `PR_DESCRIPTION.md` - This file (can be deleted after PR creation)

## Testing

✅ All tests pass:
- `internal/config` tests
- `internal/web` tests  
- Application builds successfully
- `go vet` passes without warnings

## Migration

**No manual migration required.** Changes are fully backward compatible:
- Existing `status.json` files will be auto-upgraded
- New fields will be populated on next script update
- No API breaking changes

## Documentation

See [IMPROVEMENTS.md](./IMPROVEMENTS.md) for detailed documentation including:
- Complete list of changes
- Security recommendations
- Performance impact analysis
- Future enhancement suggestions

## Checklist

- [x] Code builds successfully
- [x] All tests pass
- [x] No `go vet` warnings
- [x] Security vulnerabilities fixed
- [x] Backward compatible
- [x] Documentation added
- [x] Proper error handling
- [x] Thread-safe implementation

## Breaking Changes

None. All changes are backward compatible.

---

**Ready for review!** 🚀

All critical security issues have been addressed, code is modernized, and the enhanced JSON structure provides better tracking capabilities for GitOps workflows.
