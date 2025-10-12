# Buenos Aires

[![CI](https://github.com/your-username/buenosaires/actions/workflows/ci.yml/badge.svg)](https://github.com/your-username/buenosaires/actions/workflows/ci.yml)
[![Snyk Security](https://snyk.io/test/github/your-username/buenosaires/badge.svg)](https://snyk.io/test/github/your-username/buenosaires)

Buenos Aires is a Go-based tool for monitoring repositories and applying GitOps principles to your shell scripts. It watches a specified branch for new `.sh` files and executes them based on a set of configurable rules.

## Installation

To install Buenos Aires, you'll need to have Go installed on your system. Then, you can use the following command to install the tool:

```bash
go install github.com/your-username/buenosaires@latest
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
