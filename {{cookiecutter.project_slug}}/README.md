# {{ cookiecutter.project_name }}

[![Made with Pulumi](https://img.shields.io/badge/Made%20with-Pulumi-5F43E9?logo=pulumi&logoColor=white)](https://www.pulumi.com/)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![exivity/pulumi-hcloud-k8s](https://img.shields.io/github/stars/exivity/pulumi-hcloud-k8s?style=social&label=exivity%2Fpulumi-hcloud-k8s)](https://github.com/exivity/pulumi-hcloud-k8s)

> This template uses the [exivity/pulumi-hcloud-k8s](https://github.com/exivity/pulumi-hcloud-k8s) library for Hetzner Kubernetes deployments with Pulumi and Talos.

{{ cookiecutter.description }}

## Usage

This project is generated from a Cookiecutter template for a Pulumi Go deployment on Hetzner using Talos.

### Prerequisites

- [Go](https://golang.org/doc/install)
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [talosctl](https://www.talos.dev/v1.0/introduction/getting-started/#installing-talosctl)

### Quickstart

1. Initialize Go module (done automatically by the template):

   ```sh
   go mod init {{ cookiecutter.go_module_path }}
   ```

2. Download dependencies:

   ```sh
   make download
   ```

3. Set up your Pulumi stack:

   ```sh
   pulumi stack init {{ cookiecutter.pulumi_stack }}
   ```

4. Configure your secrets and settings in `Pulumi.{{ cookiecutter.pulumi_stack }}.yaml`.

5. Deploy:

   ```sh
   pulumi up
   ```

### Makefile Targets

- `make download` - Download Go dependencies
- `make tidy` - Clean up Go modules
- `make fmt` - Format code
- `make lint` - Lint code (requires golangci-lint)
- `make test` - Run tests
- `make kubeconfig` - Export kubeconfig from Pulumi stack
- `make talosconfig` - Export Talos config from Pulumi stack
- `make k9s` - Run k9s with the current kubeconfig
- `make talosctl` - Run talosctl with the current config
- `make clean` - Remove build artifacts

---

Generated with [Cookiecutter](https://cookiecutter.readthedocs.io/en/latest/).
