This application is a GitOps tool written in Go that automatically deploys using plugins.

The project uses GitHub Actions as a pipeline to build and test.

The sync interval for repository polling is configurable in the global `config.toml` via `sync_interval` and defaults to 180 seconds.

The application has a web GUI for viewing logs, which can be enabled and configured in the global `config.toml`. The UI is built with PatternFly.

Do not add SAST tools like Snyk or SonarQube to the CI pipeline, as they are handled by GitHub.

The local user for script execution is specified in the repository's `config.toml`.

The global configuration (`~/.buenosaires/config.toml`) includes a `RepositoryURL` field to specify the remote repository to monitor.

The `shellcheck` package is a required dependency for running tests and can be installed with `sudo apt-get install -y shellcheck`.

Repository-specific configuration is stored in a `config.toml` file in the repository's root.

The project uses `go test ./...` to run tests.

The project requires Go version 1.24 or higher.

The project uses a multi-stage Dockerfile to build a minimal runtime image.

Global configuration is stored in `~/.buenosaires/config.toml`.

The asset JSON format should be documented with an example in the main README.md.

The 'generation' field in an asset's JSON should be incremented each time the asset is modified in the repository.

Each plugin's configuration should support a 'folder_to_scan' key.

The project's GitHub Actions workflow (`.github/workflows/ci.yml`) is configured to run tests on pushes to the `main` branch.

The application clones the configured repository into the working directory on the first run and pulls for updates on subsequent runs.

The default folder to scan for assets for a plugin is a directory with the same name as the plugin (e.g., './shell').

The shell plugin's assets are tracked in individual JSON files located in the `plugins/shell/assets/` directory.

The shell plugin uses `shellcheck` for linting shell scripts.

Every plugin should have its own configuration section in the config.toml file.

The project uses the Cobra library for its command-line interface.

Plugins are enabled on a per-repository basis in the `config.toml` file within their specific configuration section (e.g., `[plugins.shell]`) by setting the `enabled = true` flag.

An asset's JSON file must be stored in a separate folder for each plugin and track: generation, last_run, tests_passed, lint_passed, event, user, run_duration, status, and the git commit hash.

Each plugin's configuration must include an `enabled` boolean field to act as an on/off switch.

The application uses a `.buenosaires/status.json` file in the repository root to track the status of processed scripts.

The project uses the `go-git` library for Git operations.

The 'event' field in an asset's JSON should store the output from linting and tests.

The default repository URL is https://github.com/gbenselum/buenosaires_test.

The application supports a plugin architecture, with the first plugin being for shell scripting.

Each plugin must maintain a JSON file for every script or file it manages, referred to as an 'asset'.

The web server (`internal/web/server.go`) validates log file paths to prevent directory traversal attacks.

The SonarQube project key is `gbenselum_buenosaires`.
