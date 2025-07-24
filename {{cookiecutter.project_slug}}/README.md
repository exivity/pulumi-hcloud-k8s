# {{ cookiecutter.project_name }}

[![Made with Pulumi](https://img.shields.io/badge/Made%20with-Pulumi-5F43E9?logo=pulumi&logoColor=white)](https://www.pulumi.com/)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![exivity/pulumi-hcloud-k8s](https://img.shields.io/github/stars/exivity/pulumi-hcloud-k8s?style=social&label=exivity%2Fpulumi-hcloud-k8s)](https://github.com/exivity/pulumi-hcloud-k8s)

{{ cookiecutter.description }}

## Usage

This project deploys a Kubernetes cluster on Hetzner Cloud using Talos Linux and Pulumi (Go).

### Prerequisites

- [Go](https://go.dev/doc/install)
- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [talosctl](https://www.talos.dev/v1.10/talos-guides/install/talosctl/)

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

4. Configure your secrets and settings:

   Add your Hetzner Cloud API token:

   ```sh
   # Set the Hetzner Cloud API token for managing resources
   pulumi config set --path hcloud-k8s:hetzner.token --secret
   # Set the Hetzner Cloud API token for deploying on Kubernetes
   pulumi config set --path hcloud-k8s:kubernetes.hcloud_token --secret
   ```

   Then edit `Pulumi.{{ cookiecutter.pulumi_stack }}.yaml` for additional configuration.

5. Deploy (requires two steps):

   ```sh
   pulumi up --yes
   ```

   This first deployment creates the Kubernetes cluster, node pools, and Talos configuration. It will not install applications or Helm charts on the first run since the cluster must be up and running first.

   After the first deployment completes, run again to install the applications:

   ```sh
   pulumi up --yes
   ```

### Access Your Cluster

Export kubeconfig and talosconfig:

```sh
make kubeconfig
make talosconfig
```

Access the cluster using `k9s`:

```sh
make k9s
```

For more configuration options, see the [Configuration Documentation](https://github.com/exivity/pulumi-hcloud-k8s/blob/main/docs/configuration.md).

### Makefile Targets

- `make download` - Download Go dependencies
- `make tidy` - Clean up Go modules
- `make fmt` - Format code
- `make kubeconfig` - Export kubeconfig from Pulumi stack
- `make talosconfig` - Export Talos config from Pulumi stack
- `make k9s` - Run k9s with the current kubeconfig
- `make talosctl` - Run talosctl with the current config
- `make clean` - Remove build artifacts
