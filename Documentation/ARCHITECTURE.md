# Buenos Aires Architecture

## Overview

Buenos Aires is a GitOps automation tool built in Go that monitors Git repositories for changes and automatically executes scripts or deploys containers based on detected changes.

## System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Buenos Aires                          │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │              Main Application (main.go)            │ │
│  │                                                    │ │
│  │  ┌──────────────────────────────────────────────┐ │ │
│  │  │          Command Layer (cmd/)                │ │ │
│  │  │  ┌──────────┐  ┌─────────┐  ┌────────────┐  │ │ │
│  │  │  │ install  │  │   run   │  │    root    │  │ │ │
│  │  │  └──────────┘  └─────────┘  └────────────┘  │ │ │
│  │  └──────────────────────────────────────────────┘ │ │
│  │                                                    │ │
│  │  ┌──────────────────────────────────────────────┐ │ │
│  │  │         Internal Packages (internal/)        │ │ │
│  │  │  ┌──────────┐  ┌─────────┐  ┌────────────┐  │ │ │
│  │  │  │  config  │  │ status  │  │    web     │  │ │ │
│  │  │  └──────────┘  └─────────┘  └────────────┘  │ │ │
│  │  └──────────────────────────────────────────────┘ │ │
│  │                                                    │ │
│  │  ┌──────────────────────────────────────────────┐ │ │
│  │  │          Plugin System (plugins/)            │ │ │
│  │  │  ┌──────────┐                ┌────────────┐  │ │ │
│  │  │  │  shell   │                │   docker   │  │ │ │
│  │  │  └──────────┘                └────────────┘  │ │ │
│  │  └──────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
         │                        │
         ▼                        ▼
┌────────────────┐     ┌─────────────────────┐
│  Git Repository│     │  Docker Daemon      │
│  (.git)        │     │                     │
└────────────────┘     └─────────────────────┘
```

## Core Components

### 1. Application Entry Point

**File**: `main.go`

**Responsibility**: Initialize the application and invoke the command layer.

**Flow**:
```
main() → cmd.Execute() → Command routing
```

### 2. Command Layer

**Package**: `cmd/`

**Components**:
- `root.go`: Base command setup using Cobra
- `install.go`: Interactive installation and configuration
- `run.go`: Main monitoring and execution loop

**Flow**:
```
User runs command
    ↓
Cobra parses arguments
    ↓
Execute appropriate command handler
    ↓
Command uses internal packages & plugins
```

### 3. Internal Packages

#### Configuration (`internal/config`)

**Purpose**: Manage global and repository-specific configuration

**Key Types**:
- `GlobalConfig`: User-wide settings (~/.buenosaires/config.toml)
- `RepoConfig`: Repository-specific settings (config.toml)
- `DockerConfig`: Docker plugin configuration

**Responsibilities**:
- Load/save TOML configuration files
- Provide configuration to other components
- Manage default values

#### Status Tracking (`internal/status`)

**Purpose**: Track execution status of scripts and containers

**Key Types**:
- `Status`: Overall status container
- `ScriptStatus`: Individual script/container status

**Features**:
- Thread-safe read/write operations
- Generation tracking
- Timestamp management
- Persistent JSON storage

**Storage**: `.buenosaires/status.json`

#### Web Server (`internal/web`)

**Purpose**: Provide web GUI for viewing logs

**Features**:
- List all log files
- View individual log contents
- Path traversal protection
- Simple HTML templates

**Endpoints**:
- `/` - List all logs
- `/logs/<filename>` - View specific log

### 4. Plugin System

**Package**: `plugins/`

**Design**: Simple interface-based plugin system

#### Shell Plugin (`plugins/shell`)

**Purpose**: Validate and execute shell scripts

**Methods**:
- `LintAndValidate()`: Syntax check + shellcheck
- `Run()`: Execute script with bash

**Process**:
```
Detect .sh file
    ↓
LintAndValidate()
    ├─ bash -n (syntax)
    └─ shellcheck (linting)
    ↓
Run()
    └─ bash (or sudo bash)
    ↓
Capture output & status
```

#### Docker Plugin (`plugins/docker`)

**Purpose**: Build and deploy Docker containers

**Methods**:
- `LintAndValidate()`: Hadolint validation
- `Build()`: Build Docker image
- `Run()`: Build + optionally run container
- `FindContainerFile()`: Locate Dockerfile/Containerfile

**Process**:
```
Detect Dockerfile in Containers/
    ↓
LintAndValidate()
    └─ hadolint
    ↓
Build()
    └─ docker build
    ↓
Run() (if auto_run enabled)
    └─ docker run
    ↓
Capture output & status
```

## Data Flow

### Installation Flow

```
User: buenosaires install
    ↓
Prompt for configuration
    ├─ Username
    ├─ Log directory
    ├─ Branch to monitor
    └─ Web GUI settings
    ↓
Create GlobalConfig
    ↓
Save to ~/.buenosaires/config.toml
```

### Monitoring Flow

```
User: buenosaires run
    ↓
Load GlobalConfig
    ↓
Load/Create Status
    ↓
Start Web Server (if enabled)
    ↓
Open Git Repository
    ↓
Enter Monitoring Loop:
    ┌─────────────────────────────────┐
    │ Fetch from remote               │
    │         ↓                       │
    │ Get latest commit               │
    │         ↓                       │
    │ Compare with last processed     │
    │         ↓                       │
    │ Detect changes:                 │
    │  ├─ New .sh files               │
    │  └─ New/modified Dockerfiles    │
    │         ↓                       │
    │ For each change:                │
    │  ├─ Load RepoConfig             │
    │  ├─ Validate (plugin)           │
    │  ├─ Execute (plugin)            │
    │  ├─ Update Status               │
    │  └─ Write Logs                  │
    │         ↓                       │
    │ Sleep (poll interval)           │
    │         ↓                       │
    └─────────────────────────────────┘
         (repeat)
```

### Script Processing Flow

```
New .sh file detected
    ↓
Check status (skip if already successful)
    ↓
Update status: pending
    ↓
Create temp file with content
    ↓
ShellPlugin.LintAndValidate()
    ├─ Success → continue
    └─ Failure → mark failed, stop
    ↓
Update status: lint success
    ↓
ShellPlugin.Run()
    ├─ Success → mark success
    └─ Failure → mark failed
    ↓
Update status: final
    ↓
Write log file
    ↓
Clean up temp file
```

### Container Processing Flow

```
Dockerfile detected in Containers/
    ↓
Check status (skip if already successful)
    ↓
Update status: pending
    ↓
DockerPlugin.LintAndValidate()
    ├─ Success → continue
    └─ Failure → mark failed, stop
    ↓
Update status: lint success
    ↓
DockerPlugin.Run()
    ├─ Build image
    └─ Optionally run container
    ├─ Success → mark success
    └─ Failure → mark failed
    ↓
Update status: final
    ↓
Write log file
```

## Concurrency Model

### Thread Safety

**Status Updates**: Protected by `sync.RWMutex`
```go
status.mu.Lock()
// Update status
status.mu.Unlock()
```

**Web Server**: Runs in separate goroutine
```go
go web.StartServer(addr, logDir)
```

**Main Loop**: Single-threaded sequential processing
- Fetches are synchronous
- Changes processed one at a time
- Prevents race conditions

### Design Rationale

- **Simplicity**: Easier to reason about and debug
- **Safety**: No concurrent file operations
- **Sufficient**: Poll interval provides natural batching
- **Future**: Can be parallelized if needed

## Storage

### Configuration Files

```
~/.buenosaires/
└── config.toml          # Global configuration

<repository>/
├── config.toml          # Repository configuration
└── .buenosaires/
    └── status.json      # Status tracking
```

### Log Files

```
<repository>/
└── logs/                # Or configured directory
    ├── script1.sh.log
    ├── script2.sh.log
    └── webapp-container.log
```

## Security Architecture

### Defense Layers

1. **Input Validation**
   - Configuration validation
   - Script size limits
   - Path sanitization

2. **Execution Isolation**
   - Temporary file cleanup
   - Separate user execution
   - Optional sudo (disabled by default)

3. **Output Protection**
   - Path traversal prevention in web server
   - Log file sanitization
   - Status file validation

4. **Audit Trail**
   - All executions logged
   - Sudo usage warnings
   - Status tracking with timestamps

## Extension Points

### Adding New Plugins

1. Create package in `plugins/`
2. Implement validation method
3. Implement execution method
4. Add detection logic to `cmd/run.go`
5. Add configuration to `internal/config`
6. Update documentation

### Example Plugin Interface

```go
type Plugin interface {
    LintAndValidate(path string) (string, error)
    Run(path string, config Config) (string, error)
}
```

## Performance Characteristics

### Resource Usage

- **Memory**: O(n) where n = number of tracked scripts/containers
- **CPU**: Minimal (sleep 10s between polls)
- **Disk**: Accumulates logs over time
- **Network**: Git fetch every poll interval

### Scalability Limits

- **Scripts per repository**: Hundreds (limited by git tree size)
- **Concurrent executions**: 1 (sequential processing)
- **Repository size**: No practical limit
- **Poll frequency**: 10 seconds (configurable)

### Optimization Opportunities

1. Parallel plugin execution
2. Incremental git fetch
3. Log rotation
4. Status file compaction
5. Selective file watching

## Error Handling Strategy

### Error Categories

1. **Fatal Errors** (exit application)
   - Cannot load global config
   - Cannot open repository
   - Cannot get branch reference

2. **Recoverable Errors** (log and continue)
   - Fetch failures (network issues)
   - Individual script failures
   - Log write failures

3. **Expected Errors** (handle gracefully)
   - No changes detected
   - Already processed scripts
   - Missing optional tools (hadolint)

### Error Flow

```
Error occurs
    ↓
Is it fatal?
    ├─ Yes → log.Fatalf() → exit
    └─ No → log.Printf() → continue
```

## Dependencies

### External Go Modules

- **cobra**: CLI framework
- **go-git**: Git operations
- **toml**: Configuration parsing

### External Tools (Optional)

- **shellcheck**: Shell script linting
- **hadolint**: Dockerfile linting
- **docker**: Container operations

### System Requirements

- **Go**: 1.24.3+
- **Git**: Any recent version
- **Bash**: For shell script execution
- **Docker**: For container plugin

## Build and Deployment

### Build Process

```
go.mod → go build → binary
```

### Docker Build

```
Multi-stage build:
  Builder: Go compilation
  Runtime: Alpine + dependencies
```

### Deployment Options

1. **Binary**: Standalone executable
2. **Docker**: Container deployment
3. **SystemD**: Service management

## Monitoring and Observability

### Current Capabilities

- Console logging
- File-based logs
- Status tracking
- Web GUI for log viewing

### Future Enhancements

- Metrics (Prometheus)
- Structured logging
- Health endpoints
- Tracing support

## Design Principles

1. **Simplicity**: Easy to understand and modify
2. **Safety**: Fail-safe defaults, explicit security
3. **GitOps**: Everything in Git, declarative
4. **Extensibility**: Plugin-based architecture
5. **Reliability**: Stateful, recoverable, auditable

## Version Compatibility

### Configuration

- Backward compatible
- New fields have defaults
- Old configs continue to work

### Status Files

- Auto-upgrade on load
- New fields added as needed
- No manual migration required

### Plugins

- Independent versioning
- No breaking changes to interface
- Extensions via new methods

---

**Last Updated**: 2025-10-14  
**Version**: 1.0.0
