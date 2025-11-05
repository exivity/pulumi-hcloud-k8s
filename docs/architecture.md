# Architecture

This project provisions a Kubernetes cluster on Hetzner Cloud using Talos Linux and Pulumi (Go).

## Key Components

- **Pulumi (Go):** Infrastructure as code, reusable as a library or CLI
- **Hetzner Cloud:** Virtual machines, networking, firewalls, and load balancers
- **Talos Linux:** Secure, immutable OS for Kubernetes nodes
- **Helm:** Deploys in-cluster components (CCM, CSI, autoscaler, etc)

## Infrastructure Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                    Hetzner Cloud VPC                       │
│                                                             │
│  ┌─────────────────┐    ┌──────────────────────────────────┐│
│  │  Load Balancer  │    │          Control Plane          ││
│  │   (Kubernetes   │────│  ┌─────┐ ┌─────┐ ┌─────┐       ││
│  │    API Server)  │    │  │ CP1 │ │ CP2 │ │ CP3 │       ││
│  └─────────────────┘    │  └─────┘ └─────┘ └─────┘       ││
│                         │   (Multi-region deployment)     ││
│                         └──────────────────────────────────┘│
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                Worker Node Pools                       ││
│  │  ┌─────┐ ┌─────┐ ┌─────┐     ┌─────┐ ┌─────┐           ││
│  │  │ W1  │ │ W2  │ │ W3  │ ... │ Wn  │ │Auto │           ││
│  │  └─────┘ └─────┘ └─────┘     └─────┘ │Scale│           ││
│  │                                      └─────┘           ││
│  └─────────────────────────────────────────────────────────┘│
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                  Firewall Rules                        ││
│  │  • Kubernetes API (6443)                               ││
│  │  • Talos API (50000)                                   ││
│  │  • Node-to-node communication                          ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Node Pools

### Control Plane

- **High Availability:** Configurable multi-region deployment across fsn1, nbg1, hel1
- **Placement Groups:** Anti-affinity rules ensure nodes are on different physical hosts
- **Load Balancer:** Highly available Kubernetes API server with automatic failover
- **Talos Configuration:** Automated machine configuration and bootstrapping

### Worker Pools

- **Multiple Pools:** Support for different server types and architectures
- **Auto Scaling:** Kubernetes Cluster Autoscaler integration
- **Flexible Sizing:** ARM64 (cax) and AMD64 (cx) server types
- **Regional Distribution:** Deploy across multiple Hetzner regions

## Networking

### Private Networking

- **VPC:** Dedicated private network with custom subnets
- **Private IPs:** All inter-node communication uses private networking
- **Firewall Rules:** Automated security group management

### Public Access

- **Load Balancer:** Public endpoint for Kubernetes API server
- **Optional Talos API:** Configurable CIDR-based access control
- **Secure by Default:** Minimal public exposure

## Storage

### Block Storage

- **Hetzner CSI:** Native integration with Hetzner Cloud Volumes
- **Encryption:** Optional volume encryption at rest
- **Dynamic Provisioning:** Automatic PVC fulfillment

### Distributed Storage (Optional)

- **Longhorn:** Cloud-native distributed storage
- **Replication:** Data redundancy across nodes
- **Snapshots & Backups:** Built-in data protection

## Security

### OS-Level Security

- **Talos Linux:** Immutable, minimal attack surface
- **No SSH:** API-only access to nodes
- **Secure Boot:** Hardware-verified boot process
- **Read-only Root:** Filesystem immutability

### Kubernetes Security

- **etcd Encryption:** Optional secrets encryption at rest
- **Network Policies:** Configurable pod-to-pod communication
- **RBAC:** Role-based access control
- **Certificate Management:** Automated TLS certificate lifecycle

## Extensibility

### Go Library Usage

```go
package main

import (
    "github.com/exivity/pulumi-hcloud-k8s/pkg/deploy"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        cluster, err := deploy.NewHetznerTalosKubernetesCluster(ctx, "my-cluster", cfg)
        return err
    })
}
```

### Modular Configuration

- **Composable:** Mix and match components as needed
- **Helm Values:** Override default chart configurations
- **Custom Resources:** Extend with additional Kubernetes resources

---

See [Configuration](configuration.md) for detailed configuration options and [Examples](examples.md) for common deployment patterns.
