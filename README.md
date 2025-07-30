# Pulumi Hetzner Kubernetes (Talos)

Deploy and manage Kubernetes clusters on Hetzner Cloud using Talos Linux, powered by Golang Pulumi. This project provides reusable infrastructure code to provision, configure, and manage **experimental-grade** Kubernetes clusters on Hetzner.

## Features

- **Automated Cluster Provisioning:** Use Pulumi (in Go) to declaratively manage Hetzner resources and Talos-based Kubernetes clusters.
- **Customizable Node Pools:** Define control plane and worker node pools with flexible sizing and configuration.
- **Infrastructure as Code:** All cluster resources, networking, and firewall rules are managed in code.
- **Go Package Distribution:** This project is distributed as a Go library that can be updated via `go mod` commands, with automated dependency management through tools like Dependabot or Renovate Bot.
- **Makefile Automation:** Common tasks (build, lint, test, deploy) are automated via `make`.
- **Talos Image Creation:** Uses [hcloud-upload-image](https://github.com/apricote/hcloud-upload-image) to create and upload Talos images to Hetzner Cloud.

## 🔋 Batteries Included

This **experimental** project comes with pre-configured Kubernetes components that integrate with Hetzner Cloud. While functional, these components are designed for testing and development environments:

### **Cluster Management & Autoscaling**

- **🚀 Cluster Autoscaler:** Automatically scales worker nodes based on pod resource demands, with configurable min/max limits and utilization thresholds
- **📊 Kubernetes Metrics Server:** Provides container resource metrics for Horizontal Pod Autoscaling (HPA) and other autoscaling pipelines
- **🔐 Kubelet Serving Certificate Approver:** Automatically approves kubelet serving certificates for secure metrics collection

### **Hetzner Cloud Integration**

- **☁️ Hetzner Cloud Controller Manager (CCM):** Native integration for Hetzner load balancers, volumes, and networking
- **💾 Hetzner CSI Driver:** Persistent volume support with encryption, configurable storage classes, and automatic volume provisioning
- **🔥 Firewall Management:** Automated firewall rules for cluster communication and optional public API access

### **Storage Solutions**

- **📦 Longhorn Distributed Storage (Optional):** Cloud-native distributed block storage with replication, snapshots, and backup capabilities when enabled in configuration

### **Security & Networking**

- **🔒 etcd Encryption at Rest:** Optional Kubernetes secrets encryption using secretbox encryption
- **🌐 Private Networking:** VPC with custom subnets, private IPs, and secure inter-node communication
- **🔧 Custom Registry Support:** Configure private container registries with authentication and TLS

### **High Availability & Reliability**

- **⚖️ Multi-Region Control Plane:** Deploy control plane nodes across multiple Hetzner regions for maximum availability
- **🔄 Control Plane Placement Groups:** Anti-affinity rules ensure control plane nodes are distributed across different physical hosts (worker nodes do not use placement groups)
- **🎯 Control Plane Load Balancer:** Highly available Kubernetes API server with automatic failover
- **📋 Node Taints & Labels:** Flexible workload scheduling with custom node labeling and tainting

### **Multi-Architecture Support**

- **🏗️ ARM64 & AMD64:** Full support for both x86_64 and ARM64 architectures with automatic image selection
- **🖼️ Talos Image Factory:** Automatic building and uploading of architecture-specific Talos images

All Helm chart components are configured with sensible defaults but remain fully customizable through Helm values and configuration overrides.

## Quickstart

This guide helps you get started with a new Hetzner Talos Kubernetes cluster using the Cookiecutter template.

### Prerequisites

- [Go](https://go.dev/doc/install)
- [Pulumi CLI](www.pulumi.com/docs/iac/download-install/)
- [talosctl](https://www.talos.dev/v1.10/talos-guides/install/talosctl/)
- [Cookiecutter](https://cookiecutter.readthedocs.io/en/latest/README.html#installation)
  
Optional:

- [k9s](https://k9scli.io/) (Kubernetes CLI UI)

### Installation Instructions

#### macOS (using Homebrew)

```sh
# Install all required tools
brew install cookiecutter
brew install go
brew install pulumi/tap/pulumi
brew install siderolabs/tap/talosctl

# Verify installations
go version
pulumi version
talosctl version --client
```

#### Linux

For Linux users, please refer to the official installation guides:

- [Go](https://go.dev/doc/install)
- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [talosctl](https://www.talos.dev/v1.10/talos-guides/install/talosctl/)

### Create a New Project

> **💡 Recommendation:** Use a dedicated Hetzner Cloud project for each cluster deployment to ensure proper resource isolation.

```sh
cookiecutter https://github.com/exivity/pulumi-hcloud-k8s
cd <your-project-slug>
make download
```

### Configure Your Stack

1. Initialize your Pulumi stack:

   ```sh
   pulumi stack init dev
   ```

2. Add your Hetzner Cloud API token:

   ```sh
   # Set the Hetzner Cloud API token for managing resources
   pulumi config set --path hcloud-k8s:hetzner.token --secret
   # Set the Hetzner Cloud API token for deploying on Kubernetes
   pulumi config set --path hcloud-k8s:kubernetes.hcloud_token --secret
   ```

### Deploy

```sh
pulumi up --yes
```

This first deployment will create the Kubernetes cluster, node pools, and Talos configuration. But it will not install any applications or Helm charts.
For that the cluster must be up and running first. So wait for the deployment to finish (failed) and then proceed with installing applications.

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

---
See [Configuration](docs/configuration.md) for all available options.

## Project Structure

- `main.go` — Pulumi entrypoint (cluster definition)
- `pkg/` — Go packages for Hetzner, Talos, and Kubernetes resource abstractions
- `Makefile` — Automation for build, test, lint, and deployment
- `docs/` — Documentation and diagrams

---

**Note:** This project is under active development. Pulumi stack configuration files are included for development purposes and may be removed in future releases. Contributions and feedback are welcome!
