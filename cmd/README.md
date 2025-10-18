# Command-Line Interface

This directory contains the source code for the command-line interface (CLI) of the application. The CLI is built using the Cobra library.

## Files

- `root.go`: This file initializes the root command of the CLI. It sets up the basic command structure and handles the execution of subcommands.
- `run.go`: This file implements the `run` command, which is responsible for the main functionality of the application. It contains the logic for monitoring the repository, detecting changes, and executing the appropriate actions.
- `install.go`: This file implements the `install` command, which is responsible for installing and configuring the application.
