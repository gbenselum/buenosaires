# cmd/root.go

## Overview

**Location**: `/cmd/root.go`  
**Package**: `cmd`  
**Purpose**: Define the root Cobra command and command execution framework

## Description

This file sets up the base command structure using the Cobra CLI framework. It defines the root command that all subcommands (install, run) are attached to.

## Code Structure

```go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "buenosaires",
    Short: "A tool to monitor a repository",
    Long:  `buenosaires is a Go-based tool for monitoring repositories and running plugins.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

## Components

### rootCmd Variable

**Type**: `*cobra.Command`

**Fields**:
- `Use`: Command name ("buenosaires")
- `Short`: Brief description shown in help
- `Long`: Detailed description shown in detailed help

**Purpose**: Serves as the parent for all subcommands

### Execute() Function

**Signature**: `func Execute()`

**Purpose**: Entry point for command processing

**Flow**:
1. Calls `rootCmd.Execute()`
2. Cobra parses command-line arguments
3. Routes to appropriate subcommand
4. Returns error if command fails
5. Prints error and exits with code 1 on failure

## Cobra Framework

### What is Cobra?

Cobra is a popular Go library for creating CLI applications. It provides:
- Command structure and routing
- Flag parsing
- Help text generation
- Automatic completion

### Command Hierarchy

```
buenosaires (root)
    ├── install
    └── run
```

Each subcommand registers itself with `rootCmd.AddCommand()` in its `init()` function.

## Usage Examples

### Display Help

```bash
$ buenosaires --help
buenosaires is a Go-based tool for monitoring repositories and running plugins.

Usage:
  buenosaires [command]

Available Commands:
  help        Help about any command
  install     Install and configure buenosaires
  run         Run the buenosaires monitor

Flags:
  -h, --help   help for buenosaires

Use "buenosaires [command] --help" for more information about a command.
```

### Invalid Command

```bash
$ buenosaires invalid
Error: unknown command "invalid" for "buenosaires"
Run 'buenosaires --help' for usage.
```

## Error Handling

**Fatal Errors**: If command execution fails, the application:
1. Prints the error message to stdout
2. Exits with status code 1

**Example**:
```bash
$ buenosaires run
# (if config missing)
Failed to load global config: open /home/user/.buenosaires/config.toml: no such file or directory
exit status 1
```

## Extension Points

### Adding New Commands

To add a new command:

1. Create new file (e.g., `cmd/newcommand.go`)
2. Define the command:
   ```go
   var newCmd = &cobra.Command{
       Use:   "new",
       Short: "Description",
       Run: func(cmd *cobra.Command, args []string) {
           // Implementation
       },
   }
   ```
3. Register in `init()`:
   ```go
   func init() {
       rootCmd.AddCommand(newCmd)
   }
   ```

### Adding Flags

Global flags (available to all commands):
```go
func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}
```

Command-specific flags:
```go
func init() {
    installCmd.Flags().BoolVar(&force, "force", false, "force reinstall")
}
```

## Related Files

- [main.go.md](../main.go.md) - Entry point that calls Execute()
- [cmd/install.go.md](./install.go.md) - Install subcommand
- [cmd/run.go.md](./run.go.md) - Run subcommand

## Dependencies

### External
- `github.com/spf13/cobra` - CLI framework

### Internal
- None (base command)

## Testing

The root command itself doesn't need testing. Subcommands are tested independently.

## Best Practices

1. **Keep root simple**: No business logic in root
2. **Descriptive help**: Clear Short and Long descriptions
3. **Consistent naming**: Follow Go and Cobra conventions
4. **Error handling**: Always handle Execute() errors

## Future Enhancements

Potential improvements:
- Add version command
- Add config validation command
- Add status command
- Global flags for verbosity, config file path

---

**Last Updated**: 2025-10-14
