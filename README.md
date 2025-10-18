# Buenos Aires

[![CI](https://github.com/gbenselum/buenosaires/actions/workflows/ci.yml/badge.svg)](https://github.com/gbenselum/buenosaires/actions/workflows/ci.yml)
[![Snyk Security](https://snyk.io/test/github/gbenselum/buenosaires/badge.svg)](https://snyk.io/test/github/gbenselum/buenosaires)
[![SonarQube](https://sonarcloud.io/api/project_badges/measure?project=your-project-key&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=your-project-key)
![Buenos Aires GitOps Logo](buenosaires_logo.png)

A `.snyk` file is included in this repository to allow for managing security policies, such as ignoring specific vulnerabilities. For more information, see the [Snyk documentation](https://docs.snyk.io/features/snyk-cli/policies/the-.snyk-file).

This project uses GitHub Actions to run a CI/CD pipeline that includes a SAST scan using Snyk. The pipeline is defined in the `.github/workflows/ci.yml` file.

Buenos Aires is a Go-based tool for monitoring repositories and applying GitOps principles to your shell scripts. It watches a specified branch for new `.sh` files and executes them based on a set of configurable rules.

## Installation

To install Buenos Aires, you'll need to have Go installed on your system. Then, you can use the following command to install the tool:

```bash
go install github.com/gbenselum/buenosaires@latest
```

After installing, you need to run the interactive setup to create the global configuration file:

```bash
buenosaires install
```

This will prompt you for the following information:
- **Username**: The default user for running scripts.
- **Log Directory**: The default directory for storing logs.
- **Branch**: The Git branch to monitor for new scripts (e.g., `main` or `master`).

This will create a configuration file at `~/.buenosaires/config.toml`.

## Usage

To use Buenos Aires to monitor a repository, you first need to create a `config.toml` file in the root of your repository. This file allows you to override the global settings and configure repository-specific behavior.

### Configuration Options

The `config.toml` file has the following sections:

#### Global Settings

These settings apply to all plugins and override the global configuration in `~/.buenosaires/config.toml`.

-   `user`: The user to run the scripts as.
-   `log_dir`: The directory to save logs to, relative to the repository root.
-   `allow_sudo`: Whether to allow scripts to be run with sudo.

#### Plugin Configuration

Each plugin has its own configuration section, which is defined by `[plugins.<plugin_name>]`. For example, the shell plugin is configured under `[plugins.shell]`.

-   `folder_to_scan`: The folder to scan for new scripts. If not specified, it defaults to `./<plugin_name>`.

### Example `config.toml`

Here's an example `config.toml` that enables the shell plugin and configures it to scan for scripts in the `scripts` folder:

```toml
user = "default"
log_dir = "logs"
allow_sudo = false

[plugins.shell]
folder_to_scan = "scripts"
```

Once you've configured your repository, you can start the monitor by running the following command from the root of your repository:

```bash
buenosaires run
```

Buenos Aires will then start monitoring the branch you specified during installation. When a new commit is pushed to that branch, it will scan for any new `.sh` files.

### GitOps Workflow

1.  **Commit a new script**: Create a new shell script (e.g., `deploy.sh`) and commit it to your repository.
2.  **Push to the monitored branch**: Push the commit to the branch that Buenos Aires is monitoring.
3.  **Linting and Validation**: Buenos Aires will automatically detect the new script and perform a dry run to validate it. This includes:
    -   **Syntax Check**: Using `bash -n` to check for syntax errors.
    -   **Linting**: Using `shellcheck` to identify potential issues.
4.  **Execution**: If the script passes the validation step, Buenos Aires will execute it using the shell plugin. The output of the script will be saved to the configured log directory.

This workflow allows you to manage your infrastructure and deployments through Git, with the assurance that your scripts are validated before they are executed.

## Running with Docker

You can also run Buenos Aires in a Docker container. This is a convenient way to run the tool without having to install Go or other dependencies on your host machine.

### Building the Docker Image

To build the Docker image, run the following command from the root of the repository:

```bash
docker build -t buenosaires .
```

### Running the Container

To run the `buenosaires` tool in a Docker container, you'll need to mount your repository and your global configuration file into the container.

First, make sure you have run `buenosaires install` on your host machine to create the global configuration file at `~/.buenosaires/config.toml`.

Then, you can run the container with the following command:

```bash
docker run -it --rm \
  -v $(pwd):/app \
  -v ~/.buenosaires:/home/appuser/.buenosaires \
  buenosaires run
```

This command does the following:
-   `docker run -it --rm`: Runs the container in interactive mode and removes it when it exits.
-   `-v $(pwd):/app`: Mounts the current directory (your repository) into the `/app` directory in the container.
-   `-v ~/.buenosaires:/home/appuser/.buenosaires`: Mounts your global configuration directory into the container.
-   `buenosaires run`: Runs the `run` command inside the container.

You can also run the `install` command in the container to create a new configuration file:

```bash
docker run -it --rm \
  -v ~/.buenosaires:/home/appuser/.buenosaires \
  buenosaires install
```

## Status Tracking

Buenos Aires keeps track of the scripts it has processed in a `.buenosaires/status.json` file in the root of your repository. This file contains the status of each script, including its linting, testing, and execution status. This file is automatically created and updated by the `run` command.

The `.buenosaires` directory is included in the `.gitignore` file, so the status file will not be committed to your repository.

## Asset Tracking

Each plugin is responsible for tracking its assets in a JSON file. An asset is any file or script that the plugin manages. The asset JSON file contains metadata about the asset, such as its generation, last run time, and test and linting status.

The following is an example of the asset JSON format for the shell plugin:

```json
{
  "generation": 1,
  "last_run": "2024-10-18T15:50:36.166157Z",
  "lint_passed": true,
  "tests_passed": true,
  "event": "Linting completed without errors. Tests passed.",
  "user": "testuser",
  "run_duration": "1.2s",
  "status": "success",
  "commit_hash": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"
}
```
