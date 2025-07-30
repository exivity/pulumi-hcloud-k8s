# Development Guide

This guide covers development workflows for contributors to the pulumi-hcloud-k8s project.

## Repository Structure

```text
├── main.go                     # Example Pulumi program
├── pkg/                        # Go packages
│   ├── config/                 # Configuration structs
│   ├── deploy/                 # Deployment logic
│   ├── hetzner/                # Hetzner Cloud resources
│   ├── k8s/                    # Kubernetes resources
│   ├── talos/                  # Talos-specific logic
│   └── validators/             # Input validation
├── docs/                       # Documentation
├── {{cookiecutter.project_slug}}/  # Cookiecutter template
└── Makefile                    # Development automation
```

## Common Tasks

### Download Dependencies

```sh
make download
```

### Lint Code

```sh
make lint
```

### Run Tests

```sh
make test
```

### Export Cluster Credentials

```sh
make kubeconfig
make talosconfig
```

## Testing Cookiecutter Template

Test the cookiecutter template generation:

```sh
make test-cookiecutter
```

This will:

1. Generate a test project using cookiecutter
2. Run linting on the generated project
3. Clean up the test project

## Contributing

This project welcomes contributions! Please ensure you have the required tools installed as listed in the main [README](../README.md#prerequisites).

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting: `make lint test`
5. Test the cookiecutter template: `make test-cookiecutter`
6. Submit a pull request

### Code Style

- Follow Go conventions and use `gofmt`
- Run `make lint` to check for issues
- Add tests for new functionality
