# Pulumi Hetzner Kubernetes (Talos)

Deploy and manage Kubernetes clusters on Hetzner Cloud using Talos Linux, powered by Golang Pulumi. This project provides reusable infrastructure code to provision, configure, and manage experimental-grade Kubernetes clusters on Hetzner.

## Features

- **Automated Cluster Provisioning:** Use Pulumi (in Go) to declaratively manage Hetzner resources and Talos-based Kubernetes clusters.
- **Customizable Node Pools:** Define control plane and worker node pools with flexible sizing and configuration.
- **Infrastructure as Code:** All cluster resources, networking, and firewall rules are managed in code.
- **Makefile Automation:** Common tasks (build, lint, test, deploy) are automated via `make`.
- **Talos Image Creation:** Uses [hcloud-upload-image](https://github.com/apricote/hcloud-upload-image) to create and upload Talos images to Hetzner Cloud.

## Requirements

Install the following tools before using this project:

- [Go](https://golang.org/) (for Pulumi program)
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/)
- [Talosctl](https://www.talos.dev/docs/latest/introduction/installation/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Cookiecutter](https://cookiecutter.readthedocs.io/en/latest/) (for project scaffolding)
- [golangci-lint](https://golangci-lint.run/) (for linting)

Optional:

- [k9s](https://k9scli.io/) (Kubernetes CLI UI)
- [helm](https://helm.sh/) (for Helm chart management)

## Quickstart

This guide helps you get started with a new Hetzner Talos Kubernetes cluster using the Cookiecutter template.

### Prerequisites

- [Go](https://go.dev/doc/install)
- [Pulumi CLI](www.pulumi.com/docs/iac/download-install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
- [talosctl](https://www.talos.dev/v1.10/talos-guides/install/talosctl/)
- [Cookiecutter](https://cookiecutter.readthedocs.io/en/latest/README.html#installation)

### Create a New Project

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
