# Pulumi Hetzner Kubernetes (Talos)

Deploy and manage Kubernetes clusters on Hetzner Cloud using Talos Linux, powered by Golang Pulumi. This project provides reusable infrastructure code to provision, configure, and manage experimental-grade Kubernetes clusters on Hetzner.

## Features

- **Automated Cluster Provisioning:** Use Pulumi (in Go) to declaratively manage Hetzner resources and Talos-based Kubernetes clusters.
- **Customizable Node Pools:** Define control plane and worker node pools with flexible sizing and configuration.
- **Infrastructure as Code:** All cluster resources, networking, and firewall rules are managed in code.
- **Image Management:** Use HashiCorp Packer to build custom Talos images for Hetzner.
- **Makefile Automation:** Common tasks (build, lint, test, deploy) are automated via `make`.

## Requirements

Install the following tools before using this project:

- [Go](https://golang.org/) (for Pulumi program)
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/)
- [Talosctl](https://www.talos.dev/docs/latest/introduction/installation/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [HashiCorp Packer](https://www.packer.io/downloads) (for custom images)
- [Cookiecutter](https://cookiecutter.readthedocs.io/en/latest/) (for project scaffolding)
- [golangci-lint](https://golangci-lint.run/) (for linting)

Optional:

- [k9s](https://k9scli.io/) (Kubernetes CLI UI)
- [helm](https://helm.sh/) (for Helm chart management)

## Quickstart

See [docs/quickstart.md](docs/quickstart.md) for the full Getting Started guide.

## Project Structure

- `main.go` — Pulumi entrypoint (cluster definition)
- `pkg/` — Go packages for Hetzner, Talos, and Kubernetes resource abstractions
- `hcloud.pkr.hcl` — Packer template for Talos image
- `Makefile` — Automation for build, test, lint, and deployment
- `docs/` — Documentation and diagrams

---

**Note:** This project is under active development. Pulumi stack configuration files are included for development purposes and may be removed in future releases. Contributions and feedback are welcome!
