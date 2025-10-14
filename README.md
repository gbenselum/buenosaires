# Buenos Aires

[![CI](https://github.com/gbenselum/buenosaires/actions/workflows/ci.yml/badge.svg)](https://github.com/gbenselum/buenosaires/actions/workflows/ci.yml)
[![Snyk Security](https://snyk.io/test/github/gbenselum/buenosaires/badge.svg)](https://snyk.io/test/github/gbenselum/buenosaires)
[![SonarQube](https://sonarcloud.io/api/project_badges/measure?project=your-project-key&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=your-project-key)

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

Here's an example `config.toml`:

```toml
# The user to run the scripts as.
user = "default"

# The directory to save logs to, relative to the repository root.
log_dir = "logs"

# Whether to allow scripts to be run with sudo.
allow_sudo = false

# Docker plugin configuration
[docker]
enabled = true
auto_run = false
default_tag = "latest"
image_prefix = ""
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

## Docker Plugin

Buenos Aires includes a Docker plugin that automatically builds and deploys containers from your repository. This enables container-based GitOps workflows alongside shell script automation.

### How It Works

1. **Create a `Containers` folder** in your repository root
2. **Add subdirectories** for each container (e.g., `Containers/webapp`, `Containers/api`)
3. **Place a `Dockerfile` or `Containerfile`** in each subdirectory
4. **Commit and push** to the monitored branch

Buenos Aires will automatically:
- Detect new or modified container files
- Lint and validate with `hadolint` (if installed)
- Build the Docker image
- Optionally run the container (if `auto_run` is enabled)
- Track the deployment status

### Configuration

Add Docker configuration to your repository's `config.toml`:

```toml
[docker]
# Enable the Docker plugin (default: true)
enabled = true

# Automatically run containers after building (default: false for safety)
# When false, only builds the image. When true, also runs the container.
auto_run = false

# Default tag for Docker images (default: "latest")
default_tag = "latest"

# Prefix for image names (optional, e.g., "mycompany/")
image_prefix = ""
```

### Example Repository Structure

```
myrepo/
├── config.toml
├── Containers/
│   ├── webapp/
│   │   ├── Dockerfile
│   │   └── (app files)
│   └── api/
│       ├── Dockerfile
│       └── (api files)
└── deploy.sh
```

### Container Naming

The Docker plugin uses the subdirectory name as the container/image name. For example:
- `Containers/webapp/Dockerfile` → image: `webapp:latest`
- `Containers/api/Containerfile` → image: `api:latest`

With `image_prefix = "mycompany/"`:
- `Containers/webapp/Dockerfile` → image: `mycompany/webapp:latest`

### GitOps Workflow for Containers

1. **Create a new container**: Add a new directory in `Containers/` with a Dockerfile/Containerfile
2. **Commit and push** to the monitored branch
3. **Validation**: Buenos Aires validates the Dockerfile with `hadolint` (if available)
4. **Build**: The Docker image is built automatically
5. **Deploy**: If `auto_run` is enabled, the container starts automatically
6. **Track**: Status is recorded in `.buenosaires/status.json`

### Security Note

By default, `auto_run` is set to `false` to prevent automatic execution of containers. Only enable this in trusted environments where you control all container definitions.

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
