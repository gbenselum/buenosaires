# main.go

## Overview

**Location**: `/main.go`  
**Package**: `main`  
**Purpose**: Application entry point for Buenos Aires

## Description

This is the simplest file in the entire project, serving solely as the entry point for the Buenos Aires application. It delegates all functionality to the command layer.

## Code Structure

```go
package main

import (
    "buenosaires/cmd"
)

func main() {
    cmd.Execute()
}
```

## Responsibilities

1. **Import the command package**: Brings in the CLI command framework
2. **Invoke Execute()**: Starts the Cobra command processing

## Execution Flow

```
OS launches binary
    ↓
main() invoked
    ↓
cmd.Execute() called
    ↓
Cobra processes arguments
    ↓
Route to appropriate command (install/run/etc.)
```

## Dependencies

- **buenosaires/cmd**: Command layer package

## Why So Simple?

This follows the **separation of concerns** principle:
- `main.go` is just an entry point
- All logic lives in appropriate packages
- Makes testing easier (can test cmd package independently)
- Follows Go best practices

## Build Process

When compiled:
```bash
go build -o buenosaires main.go
```

Results in a standalone binary that can be executed:
```bash
./buenosaires [command]
```

## Related Files

- [cmd/root.go.md](./cmd/root.go.md) - Root command definition
- [cmd/install.go.md](./cmd/install.go.md) - Install command
- [cmd/run.go.md](./cmd/run.go.md) - Run command

## Notes

- No error handling needed here (delegated to cmd layer)
- No configuration loading (delegated to commands)
- No business logic (follows single responsibility principle)

---

**Last Updated**: 2025-10-14
