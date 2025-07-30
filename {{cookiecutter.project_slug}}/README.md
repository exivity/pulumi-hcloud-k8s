# {{ cookiecutter.project_name }}

[![Made with Pulumi](https://img.shields.io/badge/Made%20with-Pulumi-5F43E9?logo=pulumi&logoColor=white)](https://www.pulumi.com/)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![exivity/pulumi-hcloud-k8s](https://img.shields.io/github/stars/exivity/pulumi-hcloud-k8s?style=social&label=exivity%2Fpulumi-hcloud-k8s)](https://github.com/exivity/pulumi-hcloud-k8s)

{{ cookiecutter.description }}

> **💡 Recommendation:** Use a dedicated Hetzner Cloud project for each cluster deployment to ensure proper resource isolation and billing separation.

## 🔋 What's Included

This **experimental** project comes pre-configured with essential Kubernetes components for Hetzner Cloud:

- **🚀 Cluster Autoscaler:** Auto-scales worker nodes based on demand
- **☁️ Hetzner Cloud Controller Manager (CCM):** Native Hetzner integration  
- **💾 Hetzner CSI Driver:** Persistent volume support with encryption
- **📊 Kubernetes Metrics Server:** For HPA and resource monitoring
- **🔐 Kubelet Serving Certificate Approver:** Automated certificate management
- **🔥 Firewall Management:** Automated security rules
- **📦 Longhorn Storage (Optional):** Distributed block storage
- **🏗️ Multi-Architecture Support:** ARM64 & AMD64 with automatic image selection
- **� Go Package Distribution:** Built as a Go library that can be updated via `go mod` commands, with automated dependency management through tools like Dependabot or Renovate Bot

All components can be customized through Helm values and configuration overrides.

## Requirements

Install the following tools to manage this cluster:

- [Go](https://go.dev/doc/install) (for Pulumi program)
- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [talosctl](https://www.talos.dev/v1.10/talos-guides/install/talosctl/)
- [golangci-lint](https://golangci-lint.run/) (for linting)

Optional but recommended:

- [k9s](https://k9scli.io/) (Kubernetes CLI UI)
- [helm](https://helm.sh/) (for Helm chart management)

### Installation Instructions

#### macOS & Linux (using Homebrew)

```sh
# Install all required tools
brew install go pulumi kubectl k9s helm
brew install pulumi/tap/pulumi
brew install siderolabs/tap/talosctl

# Verify installations
go version && pulumi version && kubectl version --client && talosctl version
```

## Cluster Management

This section covers how to manage your Kubernetes cluster deployment.

### Configuration

To modify your cluster configuration after initial setup, edit `Pulumi.{{ cookiecutter.pulumi_stack }}.yaml` and run `pulumi up` to apply changes.

For detailed configuration options and advanced setup, see the [Configuration Documentation](https://github.com/exivity/pulumi-hcloud-k8s/blob/main/docs/configuration.md).

### Deploy Your Cluster

Deploy the cluster (requires two steps):

```sh
pulumi up --yes
```

**Note:** The deployment consists of two phases:

1. **Infrastructure Phase:** Creates Hetzner Cloud resources, Talos cluster, and node pools
2. **Kubernetes Phase:** Installs applications and Helm charts on the cluster

The first deployment will fail during the Kubernetes phase because the cluster needs time to fully boot and become ready. This is expected behavior since there's no built-in check to wait for cluster readiness.

After the first deployment completes (with failures), wait a few minutes for the cluster to fully initialize, then run the deployment again to install the Kubernetes applications:

```sh
pulumi up --yes
```

### Access Your Cluster

#### Get Cluster Credentials

Export kubeconfig and talosconfig:

```sh
make kubeconfig
make talosconfig
```

The kubeconfig will be saved as `{{ cookiecutter.pulumi_stack }}.kubeconfig.yml` and talosconfig as `{{ cookiecutter.pulumi_stack }}.talosconfig.json`.

#### Access with kubectl

```sh
make kubectl get nodes
```

#### Access with k9s (Recommended)

Launch k9s to manage your cluster:

```sh
make k9s
```

#### Access with Talos

Use talosctl to manage Talos nodes:

```sh
# Get cluster information
make talosctl cluster show

# Check cluster health  
make talosctl health --server=false

# View cluster resources
make talosctl get members

# Launch Talos dashboard (web UI)
make talosctl dashboard
```

For more Talos commands, see the [talosctl CLI reference](https://www.talos.dev/v1.10/reference/cli/#talosctl-dashboard).

### Common Operations

#### Update Cluster Configuration

1. Edit `Pulumi.{{ cookiecutter.pulumi_stack }}.yaml`
2. Run `pulumi up` to apply changes

#### Scale Node Pools

Edit the node pool configuration in your Pulumi YAML file and run:

```sh
pulumi up
```

#### Update Kubernetes/Talos Versions

1. Update versions in `Pulumi.{{ cookiecutter.pulumi_stack }}.yaml`
2. Run `pulumi up` to apply updates

#### Monitor Cluster

- **k9s**: `make k9s` (recommended)
- **kubectl**: `make kubectl get pods -A`
- **Talos services**: `make talosctl services`
- **Talos cluster info**: `make talosctl cluster show`

#### Troubleshooting

- **Check cluster health**: `make talosctl health --server=false`
- **Get cluster members**: `make talosctl get members`
- **View system services**: `make talosctl services`
- **Check node status**: `make talosctl get nodes`
- **View cluster logs**: `make talosctl logs controller-runtime`
- **Check Pulumi state**: `pulumi stack`

For more configuration options, see the [Configuration Documentation](https://github.com/exivity/pulumi-hcloud-k8s/blob/main/docs/configuration.md). Since this project is experimental, configuration options may change between versions - refer to the [pkg/config](https://github.com/exivity/pulumi-hcloud-k8s/tree/main/pkg/config) source code for the most up-to-date options.

### Makefile Targets

- `make download` - Download Go dependencies
- `make tidy` - Clean up Go modules
- `make fmt` - Format code
- `make kubeconfig` - Export kubeconfig from Pulumi stack
- `make talosconfig` - Export Talos config from Pulumi stack
- `make kubectl` - Run kubectl with the current kubeconfig
- `make k9s` - Run k9s with the current kubeconfig
- `make talosctl` - Run talosctl with the current config
- `make lint` - Lint code (requires golangci-lint)
- `make test` - Run tests
- `make clean` - Remove build artifacts
