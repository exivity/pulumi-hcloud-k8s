# Architecture

This project provisions a Kubernetes cluster on Hetzner Cloud using Talos Linux and Pulumi (Go).

## Key Components

- **Pulumi (Go):** Infrastructure as code, reusable as a library or CLI.
- **Hetzner Cloud:** Virtual machines, networking, firewalls, and load balancers.
- **Talos Linux:** Secure, immutable OS for Kubernetes nodes.
- **Helm:** Deploys in-cluster components (CCM, CSI, autoscaler, etc).

## Node Pools

- **Control Plane:** Highly available, configurable node pools.
- **Workers:** Multiple pools, autoscaling supported.

## Networking

- Private VPC, firewall rules, VPN support, and public/private endpoints.

## Extensibility

- Use as a Pulumi Go library or standalone project.
- Modular configuration for custom needs.

---
See [Configuration](configuration.md) for all options.
