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

### Initial Setup

1. Initialize or select your Pulumi stack:

   **For a new local stack:**

   ```sh
   pulumi stack init {{ cookiecutter.pulumi_stack }}
   ```

   **For a new organization stack:**

   ```sh
   pulumi stack init <org-name>/{{ cookiecutter.pulumi_stack }}
   ```

   **If the stack already exists, select it:**

   ```sh
   pulumi stack select {{ cookiecutter.pulumi_stack }}
   # or for organization stacks:
   pulumi stack select <org-name>/{{ cookiecutter.pulumi_stack }}
   ```

2. Configure your Hetzner Cloud API tokens:

   ```sh
   # Set the Hetzner Cloud API token for managing resources
   pulumi config set --path hcloud-k8s:hetzner.token --secret
   # Set the Hetzner Cloud API token for deploying on Kubernetes
   pulumi config set --path hcloud-k8s:kubernetes.hcloud_token --secret
   ```

3. Review and customize your configuration in `Pulumi.{{ cookiecutter.pulumi_stack }}.yaml`.

### Deploy Your Cluster

Deploy the cluster (requires two steps):

```sh
pulumi up --yes
```

This first deployment creates the Kubernetes cluster, node pools, and Talos configuration. It will not install applications or Helm charts on the first run since the cluster must be up and running first.

After the first deployment completes, run again to install the applications:

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
kubectl --kubeconfig ./{{ cookiecutter.pulumi_stack }}.kubeconfig.yml get nodes
```

#### Access with k9s (Recommended)

Launch k9s to manage your cluster:

```sh
make k9s
```

#### Access with Talos

Use talosctl to manage Talos nodes:

```sh
make talosctl dashboard
```

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
- **kubectl**: `kubectl --kubeconfig ./{{ cookiecutter.pulumi_stack }}.kubeconfig.yml get pods -A`
- **Talos Dashboard**: `make talosctl dashboard`

#### Troubleshooting

- **Check cluster status**: `make talosctl health`
- **View cluster logs**: `make talosctl logs`
- **Check Pulumi state**: `pulumi stack`

For more configuration options, see the [Configuration Documentation](https://github.com/exivity/pulumi-hcloud-k8s/blob/main/docs/configuration.md).

### Makefile Targets

- `make download` - Download Go dependencies
- `make tidy` - Clean up Go modules
- `make fmt` - Format code
- `make kubeconfig` - Export kubeconfig from Pulumi stack
- `make talosconfig` - Export Talos config from Pulumi stack
- `make k9s` - Run k9s with the current kubeconfig
- `make talosctl` - Run talosctl with the current config
- `make lint` - Lint code (requires golangci-lint)
- `make test` - Run tests
- `make clean` - Remove build artifacts
