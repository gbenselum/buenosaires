# Buenos Aires Documentation

Welcome to the comprehensive documentation for Buenos Aires - a GitOps tool for monitoring repositories and automating shell scripts and container deployments.

## Documentation Structure

This documentation is organized into the following sections:

### 📚 Core Documentation

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - System architecture and design principles
- **[GETTING_STARTED.md](./GETTING_STARTED.md)** - Quick start guide for new users
- **[CONFIGURATION.md](./CONFIGURATION.md)** - Complete configuration reference

### 🔌 Plugin Documentation

- **[SHELL_PLUGIN.md](../SHELL_PLUGIN.md)** - Shell script automation plugin
- **[DOCKER_PLUGIN.md](../DOCKER_PLUGIN.md)** - Docker container deployment plugin

### 📁 File-by-File Documentation

#### Core Application
- **[main.go.md](./files/main.go.md)** - Application entry point
- **[Dockerfile.md](./files/Dockerfile.md)** - Container build configuration

#### Commands
- **[cmd/root.go.md](./files/cmd/root.go.md)** - Root command definition
- **[cmd/install.go.md](./files/cmd/install.go.md)** - Installation command
- **[cmd/run.go.md](./files/cmd/run.go.md)** - Main monitoring command

#### Internal Packages
- **[internal/config/config.go.md](./files/internal/config/config.go.md)** - Configuration management
- **[internal/status/status.go.md](./files/internal/status/status.go.md)** - Status tracking
- **[internal/web/server.go.md](./files/internal/web/server.go.md)** - Web GUI server

#### Plugins
- **[plugins/shell/shell.go.md](./files/plugins/shell/shell.go.md)** - Shell plugin implementation
- **[plugins/docker/docker.go.md](./files/plugins/docker/docker.go.md)** - Docker plugin implementation

### 🔧 Development Documentation

- **[DEVELOPMENT.md](./DEVELOPMENT.md)** - Development setup and guidelines
- **[TESTING.md](./TESTING.md)** - Testing strategy and test documentation
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Contribution guidelines

### 🔐 Security Documentation

- **[SECURITY.md](./SECURITY.md)** - Security considerations and best practices

## Quick Links

### For Users
- [Installation Guide](./GETTING_STARTED.md#installation)
- [Basic Usage](./GETTING_STARTED.md#basic-usage)
- [Configuration Examples](./CONFIGURATION.md#examples)
- [Troubleshooting](./TROUBLESHOOTING.md)

### For Developers
- [Architecture Overview](./ARCHITECTURE.md)
- [Development Setup](./DEVELOPMENT.md#setup)
- [Creating Plugins](./DEVELOPMENT.md#creating-plugins)
- [Testing](./TESTING.md)

### For DevOps
- [GitOps Workflow](./GITOPS.md)
- [Security Best Practices](./SECURITY.md)
- [Production Deployment](./PRODUCTION.md)

## Navigation

Each documentation file includes:
- **Purpose**: What the file/component does
- **Key Concepts**: Important concepts to understand
- **API Reference**: Functions, methods, and types
- **Examples**: Practical usage examples
- **Related Files**: Links to related documentation

## Documentation Standards

All documentation follows these standards:
- **Markdown format** for readability
- **Code examples** with syntax highlighting
- **Cross-references** to related documentation
- **Version compatibility** notes where applicable
- **Last updated** date in each file

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/gbenselum/buenosaires/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gbenselum/buenosaires/discussions)
- **Email**: Project maintainers

## Contributing to Documentation

Documentation improvements are always welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on:
- Fixing typos and errors
- Adding examples
- Improving clarity
- Translating documentation

---

**Last Updated**: 2025-10-14  
**Version**: 1.0.0
