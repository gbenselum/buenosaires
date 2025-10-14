# cmd/run.go

## Overview

**Location**: `/cmd/run.go`  
**Package**: `cmd`  
**Purpose**: Main monitoring command that watches Git repository for changes and executes plugins

## Description

This is the **core file** of Buenos Aires. It implements the monitoring loop that:
- Watches a Git repository for changes
- Detects new/modified shell scripts and container files
- Executes appropriate plugins
- Tracks status and logs output

**Complexity**: ~400 lines  
**Importance**: ⭐⭐⭐⭐⭐ (Critical)

## Constants

```go
const (
    DefaultPollInterval = 10 * time.Second          // Check repo every 10s
    MaxScriptSize = 10 * 1024 * 1024               // 10MB limit
    DefaultScriptTimeout = 5 * time.Minute          // Script timeout
    DirPerm  = 0755                                 // Directory permissions
    FilePerm = 0644                                 // File permissions
)
```

## Command Definition

```go
var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Run the buenosaires monitor",
    Long:  `This command starts the buenosaires monitor...`,
    Run: func(cmd *cobra.Command, args []string) {
        // Main monitoring logic
    },
}
```

## Main Execution Flow

### 1. Initialization Phase

```go
// Load global config
globalConfig, err := config.LoadGlobalConfig()
if err != nil {
    log.Fatalf("Failed to load global config: %v", err)
}

// Load status file
status, err := status.LoadStatus(".")
if err != nil {
    log.Fatalf("Failed to load status file: %v", err)
}

// Start web server if enabled
if globalConfig.GUI.Enabled {
    addr := fmt.Sprintf(":%d", globalConfig.GUI.Port)
    go web.StartServer(addr, globalConfig.LogDir)
}

// Open git repository
repo, err := git.PlainOpen(".")
if err != nil {
    log.Fatalf("Failed to open repository: %v", err)
}
```

### 2. Monitoring Loop

```go
for {
    // Fetch latest changes
    err := repo.Fetch(&git.FetchOptions{})
    
    // Get latest commit
    branchRef, err := repo.Reference(branchRefName, true)
    latestCommitHash := branchRef.Hash()
    
    // If commit changed, process changes
    if latestCommitHash != lastCommitHash {
        // Detect and process changes
        processChanges()
        
        lastCommitHash = latestCommitHash
    }
    
    time.Sleep(DefaultPollInterval)
}
```

### 3. Change Processing

For each file change:

```go
// Check if it's a new shell script
if isNewShellScript(change) {
    processShellScript(change)
}

// Check if it's a new/modified container file
if isNewOrModifiedContainerFile(change) {
    processContainerFile(change)
}
```

## Helper Functions

### isNewShellScript()

```go
func isNewShellScript(change *object.Change) bool {
    action, err := change.Action()
    if err != nil {
        return false
    }
    return action == merkletrie.Insert && strings.HasSuffix(change.To.Name, ".sh")
}
```

**Purpose**: Detect new shell script files  
**Criteria**: INSERT action + `.sh` extension

### isNewOrModifiedContainerFile()

```go
func isNewOrModifiedContainerFile(change *object.Change) bool {
    action, err := change.Action()
    if err != nil {
        return false
    }
    
    path := change.To.Name
    if !strings.HasPrefix(path, "Containers/") && 
       !strings.HasPrefix(path, "containers/") {
        return false
    }
    
    filename := filepath.Base(path)
    isContainerFile := filename == "Dockerfile" || filename == "Containerfile"
    
    return (action == merkletrie.Insert || action == merkletrie.Modify) && isContainerFile
}
```

**Purpose**: Detect container files  
**Criteria**: 
- In `Containers/` directory
- Named `Dockerfile` or `Containerfile`
- INSERT or MODIFY action

## Shell Script Processing

```go
// Get file content
file, err := latestTree.File(scriptName)
content, err := file.Contents()

// Check size
if len(content) > MaxScriptSize {
    log.Printf("Script %s exceeds maximum size", scriptName)
    status.UpdateScriptStatus(scriptName, "failure", ...)
    continue
}

// Create temp file
tmpfile, err := os.CreateTemp("", "script-*.sh")
tmpfile.Write([]byte(content))
tmpfile.Close()

// Lint and validate
plugin := shell.ShellPlugin{}
lintOutput, err := plugin.LintAndValidate(tmpfileName)
if err != nil {
    log.Printf("Validation failed: %v", err)
    status.UpdateScriptStatus(scriptName, "failure", ...)
    os.Remove(tmpfileName)
    continue
}

// Execute
execOutput, err := plugin.Run(tmpfileName, repoConfig.AllowSudo)
if err != nil {
    status.UpdateScriptStatus(scriptName, "success", "skipped", "failure", "failure")
} else {
    status.UpdateScriptStatus(scriptName, "success", "skipped", "success", "success")
}

// Clean up
os.Remove(tmpfileName)

// Write log
logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", scriptName))
logContent := fmt.Sprintf("--- LINT OUTPUT ---\n%s\n--- EXECUTION OUTPUT ---\n%s", 
                         lintOutput, execOutput)
os.WriteFile(logFile, []byte(logContent), FilePerm)
```

## Container Processing

```go
// Verify file exists
_, err := latestTree.File(containerPath)

// Get container name from directory
containerDir := filepath.Dir(containerPath)
containerName := filepath.Base(containerDir)

// Lint and validate
dockerPlugin := docker.DockerPlugin{}
fullContainerPath := filepath.Join(".", containerPath)
lintOutput, err := dockerPlugin.LintAndValidate(fullContainerPath)
if err != nil {
    status.UpdateScriptStatus(statusKey, "failure", ...)
    continue
}

// Build (and optionally run)
imageTag := repoConfig.Docker.DefaultTag
if imageTag == "" {
    imageTag = "latest"
}

imageName := containerName
if repoConfig.Docker.ImagePrefix != "" {
    imageName = repoConfig.Docker.ImagePrefix + imageName
}

execOutput, err := dockerPlugin.Run(fullContainerPath, imageName, imageTag, 
                                    repoConfig.Docker.AutoRun)
if err != nil {
    status.UpdateScriptStatus(statusKey, "success", "skipped", "failure", "failure")
} else {
    status.UpdateScriptStatus(statusKey, "success", "skipped", "success", "success")
}

// Write log
logFile := filepath.Join(logDir, fmt.Sprintf("%s-container.log", containerName))
logContent := fmt.Sprintf("--- LINT OUTPUT ---\n%s\n--- BUILD/RUN OUTPUT ---\n%s", 
                         lintOutput, execOutput)
os.WriteFile(logFile, []byte(logContent), FilePerm)
```

## Security Features

### 1. Script Size Limit
```go
if len(content) > MaxScriptSize {
    // Reject and log
}
```

### 2. Sudo Audit Logging
```go
if repoConfig.AllowSudo {
    log.Printf("WARNING: Executing script %s with sudo privileges", scriptName)
}
```

### 3. Temporary File Cleanup
```go
tmpfileName := tmpfile.Name()
defer os.Remove(tmpfileName)  // Not used - cleaned up explicitly
os.Remove(tmpfileName)         // Explicit cleanup after execution
```

### 4. Path Validation
Container files must be in `Containers/` directory - prevents arbitrary file execution.

## Error Handling

### Fatal Errors (stop application)
- Cannot load global config
- Cannot open git repository
- Cannot get branch reference

### Non-Fatal Errors (log and continue)
- Fetch failures (network issues)
- Individual script/container failures
- Log write failures

Example:
```go
if err := status.SaveStatus("."); err != nil {
    log.Printf("Failed to save status: %v", err)
    // Continue anyway
}
```

## Status Updates

Every operation updates status:

```go
// Initial
status.UpdateScriptStatus(name, "pending", "skipped", "pending", "pending")

// After lint
status.UpdateScriptStatus(name, "success", "skipped", "pending", "pending")

// After execution
status.UpdateScriptStatus(name, "success", "skipped", "success", "success")
```

Each update:
- Increments generation counter
- Updates timestamps
- Saves to disk

## Concurrency

### Web Server
Runs in goroutine:
```go
go web.StartServer(addr, globalConfig.LogDir)
```

### Main Loop
Single-threaded, sequential processing:
- Simple to reason about
- No race conditions
- Status updates are thread-safe (mutex protected)

## Dependencies

### External
- `github.com/go-git/go-git/v5` - Git operations
- `github.com/spf13/cobra` - CLI framework

### Internal
- `buenosaires/internal/config` - Configuration
- `buenosaires/internal/status` - Status tracking
- `buenosaires/internal/web` - Web server
- `buenosaires/plugins/shell` - Shell plugin
- `buenosaires/plugins/docker` - Docker plugin

## Configuration Used

### Global Config
- `Branch`: Branch to monitor
- `LogDir`: Default log directory
- `Plugins`: Enabled plugins
- `GUI`: Web interface settings

### Repo Config
- `LogDir`: Override global log directory
- `AllowSudo`: Allow sudo execution
- `Docker`: Docker plugin settings

## Logging

All operations logged to console:
- `log.Printf()`: Informational messages
- `log.Fatalf()`: Fatal errors (exits)

Log messages include:
- Script/container names
- Validation results
- Execution status
- Warnings (sudo usage)

## Performance Characteristics

### Polling Frequency
- Default: 10 seconds
- Configurable via `DefaultPollInterval`

### Processing Time
- Sequential: One change at a time
- Bounded by script execution time
- No timeout enforcement (future enhancement)

### Resource Usage
- Memory: O(n) for n tracked items
- CPU: Minimal when idle
- Network: Git fetch every poll
- Disk: Log accumulation

## Future Enhancements

Potential improvements:
- [ ] Parallel processing of changes
- [ ] Configurable poll interval
- [ ] Script timeout enforcement
- [ ] Selective file watching (inotify)
- [ ] Graceful shutdown (signal handling)
- [ ] Metrics collection

## Troubleshooting

### Script Not Detected
- Check filename ends with `.sh`
- Verify file is new (INSERT action)
- Ensure on correct branch

### Container Not Detected
- Check path: `Containers/<name>/Dockerfile`
- Verify case sensitivity
- Check plugin enabled

### High CPU Usage
- Check poll interval
- Verify git operations not failing
- Check script execution times

## Related Files

- [cmd/install.go.md](./install.go.md) - Installation command
- [internal/config/config.go.md](../internal/config/config.go.md) - Configuration
- [internal/status/status.go.md](../internal/status/status.go.md) - Status tracking
- [plugins/shell/shell.go.md](../plugins/shell/shell.go.md) - Shell plugin
- [plugins/docker/docker.go.md](../plugins/docker/docker.go.md) - Docker plugin

---

**Last Updated**: 2025-10-14  
**Lines of Code**: ~400  
**Complexity**: High
