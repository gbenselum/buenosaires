# Docker Plugin Documentation

## Overview

The Docker plugin extends Buenos Aires with container-based GitOps capabilities, allowing you to automatically build and deploy Docker containers alongside shell script automation.

## Features

### Core Capabilities

- ✅ **Automatic Detection**: Monitors `Containers/` folder for Dockerfile/Containerfile changes
- ✅ **Validation**: Lints Dockerfiles with `hadolint` (if installed)
- ✅ **Building**: Automatically builds Docker images on commit
- ✅ **Optional Deployment**: Can automatically run containers (configurable)
- ✅ **Status Tracking**: Full integration with Buenos Aires status system
- ✅ **Logging**: Complete build and deployment logs

### Supported File Types

- `Dockerfile` (standard Docker format)
- `Containerfile` (Podman/OCI format)

## Directory Structure

The Docker plugin expects containers to be organized in a `Containers/` folder at the repository root:

```
myrepo/
├── config.toml
├── Containers/
│   ├── webapp/
│   │   ├── Dockerfile
│   │   ├── app.py
│   │   └── requirements.txt
│   ├── api/
│   │   ├── Containerfile
│   │   └── server.js
│   └── worker/
│       ├── Dockerfile
│       └── worker.rb
└── deploy.sh
```

## Configuration

### Repository Configuration

Add the following to your repository's `config.toml`:

```toml
[docker]
# Enable the Docker plugin (default: true)
enabled = true

# Automatically run containers after building (default: false)
# SECURITY: Only enable in trusted environments
auto_run = false

# Default tag for Docker images (default: "latest")
default_tag = "latest"

# Prefix for image names (optional)
# Example: "mycompany/" -> mycompany/webapp:latest
image_prefix = ""
```

### Configuration Options Explained

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable/disable the Docker plugin |
| `auto_run` | bool | `false` | Automatically start containers after building |
| `default_tag` | string | `"latest"` | Tag to use for built images |
| `image_prefix` | string | `""` | Prefix added to all image names |

## How It Works

### Detection

When a commit is pushed to the monitored branch, Buenos Aires:

1. Scans for changes in the `Containers/` folder
2. Looks for `Dockerfile` or `Containerfile` in subdirectories
3. Detects both new files (INSERT) and modifications (MODIFY)

### Processing Pipeline

For each detected container file:

1. **Validation**
   - Check file exists and is accessible
   - Run `hadolint` linting (if available)
   - Fail fast on critical errors

2. **Building**
   - Use directory name as image name
   - Apply image prefix if configured
   - Apply default tag
   - Execute `docker build` with full context

3. **Optional Deployment** (if `auto_run = true`)
   - Remove old container if exists
   - Run new container in detached mode
   - Name container: `<imagename>-<tag>`

4. **Logging**
   - Save lint output
   - Save build output
   - Save run output (if applicable)
   - Write to `<logdir>/<containername>-container.log`

5. **Status Tracking**
   - Update status in `.buenosaires/status.json`
   - Track with key: `container:<name>`
   - Record generation, timestamps, and status

## Image Naming

### Basic Naming

The subdirectory name becomes the image name:

```
Containers/webapp/Dockerfile -> webapp:latest
Containers/api/Dockerfile    -> api:latest
```

### With Tag Configuration

```toml
default_tag = "v1.0"
```

Results in:
```
Containers/webapp/Dockerfile -> webapp:v1.0
```

### With Prefix Configuration

```toml
image_prefix = "mycompany/"
default_tag = "prod"
```

Results in:
```
Containers/webapp/Dockerfile -> mycompany/webapp:prod
```

## GitOps Workflow Examples

### Example 1: Deploy a Web Application

1. **Create container directory**:
   ```bash
   mkdir -p Containers/webapp
   ```

2. **Add Dockerfile**:
   ```dockerfile
   # Containers/webapp/Dockerfile
   FROM python:3.9-slim
   WORKDIR /app
   COPY requirements.txt .
   RUN pip install -r requirements.txt
   COPY . .
   CMD ["python", "app.py"]
   ```

3. **Add application files**:
   ```bash
   # Add your app.py, requirements.txt, etc.
   ```

4. **Commit and push**:
   ```bash
   git add Containers/webapp/
   git commit -m "Add webapp container"
   git push origin main
   ```

5. **Buenos Aires will**:
   - Detect the new Dockerfile
   - Lint with hadolint
   - Build image: `webapp:latest`
   - Log the output
   - Track deployment status

### Example 2: Update an Existing Container

1. **Modify Dockerfile**:
   ```bash
   vim Containers/api/Dockerfile
   # Make your changes
   ```

2. **Commit and push**:
   ```bash
   git add Containers/api/Dockerfile
   git commit -m "Update API container to use Node 18"
   git push origin main
   ```

3. **Buenos Aires will**:
   - Detect the modification
   - Validate and lint
   - Rebuild image: `api:latest`
   - Increment generation counter
   - Update status

### Example 3: Multi-Container Application

```
Containers/
├── frontend/
│   ├── Dockerfile
│   └── (React app files)
├── backend/
│   ├── Dockerfile
│   └── (Node.js API files)
└── database/
    ├── Dockerfile
    └── (PostgreSQL config)
```

Each container will be built independently when its Dockerfile changes.

## Status Tracking

### Status File

Containers are tracked in `.buenosaires/status.json`:

```json
{
  "scripts": {
    "container:webapp": {
      "lint_status": "success",
      "test_status": "skipped",
      "run_status": "success",
      "timestamp": "2025-10-14T10:30:00Z",
      "overall_status": "success",
      "generation": 3,
      "first_deploy_date": "2025-10-10T14:20:00Z",
      "current_version_date": "2025-10-14T10:30:00Z"
    }
  }
}
```

### Fields Explained

- **lint_status**: Result of hadolint validation
- **run_status**: Result of docker build (and run if enabled)
- **overall_status**: Combined status
- **generation**: Number of times this container has been redeployed
- **first_deploy_date**: When container was first deployed
- **current_version_date**: When current version was deployed

## Logging

### Log Files

Logs are saved to the configured log directory:

```
logs/
├── deploy.sh.log           # Shell script logs
└── webapp-container.log     # Container logs
```

### Log Format

```
--- LINT OUTPUT ---
(hadolint output)

--- BUILD/RUN OUTPUT ---
(docker build output)
(docker run output if auto_run enabled)
```

## Security Considerations

### Auto-Run Safety

⚠️ **WARNING**: `auto_run` is disabled by default for security.

**Only enable `auto_run` when**:
- You control all Dockerfiles in the repository
- You trust all contributors
- You're in a sandboxed environment
- You understand the security implications

**Risks of enabling `auto_run`**:
- Containers run with Docker daemon privileges
- Malicious containers could compromise the host
- Resource exhaustion attacks possible
- Network exposure if ports are published

### Best Practices

1. **Review all Dockerfiles** before enabling auto_run
2. **Use image scanning** tools (Snyk, Trivy, etc.)
3. **Implement resource limits** in Dockerfiles
4. **Use minimal base images** (alpine, distroless)
5. **Don't run as root** inside containers
6. **Pin dependency versions** explicitly
7. **Enable hadolint** for best practices enforcement

## Linting with Hadolint

### Installation

```bash
# macOS
brew install hadolint

# Linux
wget -O /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
chmod +x /usr/local/bin/hadolint

# Docker
docker pull hadolint/hadolint
```

### What Hadolint Checks

- Dockerfile best practices
- Common mistakes
- Security issues
- Performance optimizations
- Style consistency

### Example Linting Output

```
Containers/webapp/Dockerfile:1 DL3006 warning: Always tag the version of an image explicitly
Containers/webapp/Dockerfile:3 DL3042 warning: Avoid the use of cache directory with RUN
Linting completed with warnings.
```

## Troubleshooting

### Container Not Detected

**Problem**: Dockerfile changes not triggering builds

**Solutions**:
1. Ensure file is in `Containers/<name>/Dockerfile`
2. Check that docker plugin is enabled in config
3. Verify file is committed and pushed
4. Check logs for errors

### Build Failures

**Problem**: Docker build fails

**Solutions**:
1. Check Dockerfile syntax
2. Ensure build context has all required files
3. Verify Docker daemon is running
4. Check hadolint warnings for issues
5. Review build logs in `logs/<name>-container.log`

### Permission Issues

**Problem**: Cannot connect to Docker daemon

**Solutions**:
1. Ensure user is in docker group: `sudo usermod -aG docker $USER`
2. Restart Docker daemon
3. Check Docker socket permissions

## Advanced Usage

### Custom Build Arguments

Currently not supported. To add build arguments:

1. Modify `plugins/docker/docker.go`
2. Update `Build()` method to accept build args
3. Pass via `docker build --build-arg`

### Multi-Stage Builds

Fully supported! Use standard multi-stage Dockerfile syntax:

```dockerfile
FROM node:18 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
```

### Private Registries

To push to private registries:

1. Configure `image_prefix` with registry URL:
   ```toml
   image_prefix = "registry.example.com/myteam/"
   ```

2. Ensure Docker is logged in to the registry

3. Add a post-build script to push:
   ```bash
   docker push registry.example.com/myteam/webapp:latest
   ```

## API Reference

### DockerPlugin Methods

#### `LintAndValidate(containerFilePath string) (string, error)`

Validates a Dockerfile/Containerfile using hadolint.

**Parameters**:
- `containerFilePath`: Path to Dockerfile/Containerfile

**Returns**:
- `string`: Linting output
- `error`: Error if validation fails

#### `Build(containerFilePath, imageName, imageTag string) (string, error)`

Builds a Docker image.

**Parameters**:
- `containerFilePath`: Path to Dockerfile/Containerfile
- `imageName`: Name for the image
- `imageTag`: Tag for the image

**Returns**:
- `string`: Build output
- `error`: Error if build fails

#### `Run(containerFilePath, imageName, imageTag string, autoRun bool) (string, error)`

Builds and optionally runs a Docker container.

**Parameters**:
- `containerFilePath`: Path to Dockerfile/Containerfile
- `imageName`: Name for the image
- `imageTag`: Tag for the image
- `autoRun`: Whether to run the container after building

**Returns**:
- `string`: Combined build and run output
- `error`: Error if operation fails

#### `FindContainerFile(dir string) (string, error)`

Finds a Dockerfile or Containerfile in a directory.

**Parameters**:
- `dir`: Directory to search

**Returns**:
- `string`: Path to found container file
- `error`: Error if no file found

## Testing

### Running Tests

```bash
# Run all Docker plugin tests
go test ./plugins/docker -v

# Run specific test
go test ./plugins/docker -v -run TestDockerPlugin_LintAndValidate
```

### Test Coverage

Current test coverage includes:
- ✅ Linting and validation
- ✅ File detection (Dockerfile and Containerfile)
- ✅ Error handling for missing files
- ✅ Build operations (requires Docker)
- ✅ Run operations (requires Docker)

### Manual Testing

1. Create test repository:
   ```bash
   mkdir test-repo && cd test-repo
   git init
   ```

2. Add config:
   ```toml
   # config.toml
   log_dir = "logs"
   [docker]
   enabled = true
   auto_run = false
   ```

3. Create container:
   ```bash
   mkdir -p Containers/test
   cat > Containers/test/Dockerfile <<EOF
   FROM alpine:latest
   CMD ["echo", "Hello from Buenos Aires!"]
   EOF
   ```

4. Commit and test:
   ```bash
   git add .
   git commit -m "Add test container"
   buenosaires run
   ```

## Future Enhancements

Potential improvements for future versions:

- [ ] Support for docker-compose.yml files
- [ ] Custom build arguments configuration
- [ ] Health checks for running containers
- [ ] Automatic image cleanup for old versions
- [ ] Integration with container registries
- [ ] Support for BuildKit features
- [ ] Parallel container builds
- [ ] Custom container networking
- [ ] Volume management
- [ ] Environment variable injection

## Contributing

To contribute to the Docker plugin:

1. Follow the existing code structure in `plugins/docker/`
2. Add tests for all new features
3. Update this documentation
4. Ensure backward compatibility
5. Follow Go best practices

## License

Same as Buenos Aires main project.
