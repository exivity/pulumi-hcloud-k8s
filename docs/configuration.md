# Pulumi Stack Configuration

All cluster options are configured via Pulumi stack YAML files (e.g., `Pulumi.dev.yaml`).

## Main Configuration Sections

### Hetzner Cloud Settings

Configure your Hetzner Cloud API tokens using the Pulumi CLI (recommended for security):

```sh
# Set the Hetzner Cloud API token for managing infrastructure resources
pulumi config set --path hcloud-k8s:hetzner.token --secret

# Set the Hetzner Cloud API token for Kubernetes components (CCM, CSI, Cluster Autoscaler)
# This is required when enabling any Kubernetes features that integrate with Hetzner Cloud
pulumi config set --path hcloud-k8s:kubernetes.hcloud_token --secret
```

These commands will prompt you to enter the token values securely without exposing them in your shell history.

> **⚠️ Important:** Both tokens are required if you plan to use Kubernetes features like Hetzner CCM, CSI, or Cluster Autoscaler. You can use the same token value for both fields, but both must be explicitly set.

**Alternative (not recommended for production):** You can also set tokens directly in YAML, but this should only be done for local development when you understand the security implications:

```yaml
config:
  hcloud-k8s:hetzner:
    token: <your-hcloud-token>
  hcloud-k8s:kubernetes:
    hcloud_token: <your-hcloud-token>  # Can be the same token as above
```

> **⚠️ Security Warning:** Never commit plain text tokens to version control. Always use `pulumi config set --secret` for sensitive values.

### Talos Configuration

Configure Talos Linux version and Kubernetes version:

```yaml
config:
  hcloud-k8s:talos:
    image_version: v1.11.3
    kubernetes_version: "1.34.0"
    api_allowed_cidrs: "10.0.0.0/8,192.168.0.0/16"  # Optional
```

### Control Plane Configuration

Configure control plane nodes:

```yaml
config:
  hcloud-k8s:control_plane:
    enable_ha: true  # Deploy across multiple regions
    node_pools:
      - count: 3
        server_size: cax11
        arch: arm64
        region: fsn1
```

Disable load balancer (development only)

If you are running a small development or test cluster you can disable the
automatic Hetzner load balancer creation by setting `disable_load_balancer`.
This is useful for single-node clusters, CI jobs, or local integration tests
where a cloud load balancer is not available or desired. Example:

```yaml
config:
  hcloud-k8s:control_plane:
    disable_load_balancer: true
```

Important: Disabling the load balancer is intended only for development and
testing. It is not recommended for production workloads because it bypasses
the high-availability and traffic distribution guarantees provided by a
properly configured load balancer.

### Worker Node Pools

Configure worker node pools:

```yaml
config:
  hcloud-k8s:node_pools:
    node_pools:
      - name: core
        count: 2
        server_size: cx23
        arch: amd64
        region: fsn1
        auto_scaler:
          enabled: true
          min_count: 1
          max_count: 5
```

### Kubernetes Components

Enable and configure Kubernetes components:

```yaml
config:
  hcloud-k8s:kubernetes:
    hetzner_ccm:
      enabled: true
      version: "1.23.0"
    hetzner_csi:
      enabled: true
    cluster_autoscaler:
      enabled: true
    metrics_server:
      enabled: true
    kubelet_cert_approver:
      enabled: true
    longhorn:
      enabled: false  # Optional distributed storage
```

## Complete Example

> **Note:** This example shows configuration structure. Hetzner tokens are set as secrets via CLI and won't appear in the YAML file. Both tokens are required when using Kubernetes features.

```yaml
config:
  # Tokens are set as secrets via: pulumi config set --path hcloud-k8s:hetzner.token --secret
  # Tokens are set as secrets via: pulumi config set --path hcloud-k8s:kubernetes.hcloud_token --secret
  hcloud-k8s:talos:
    image_version: v1.11.3
    kubernetes_version: "1.34.0"
  hcloud-k8s:control_plane:
    enable_ha: true
    node_pools:
      - count: 3
        server_size: cax11
        arch: arm64
        region: fsn1
  hcloud-k8s:node_pools:
    node_pools:
      - name: core
        count: 2
        server_size: cx23
        arch: amd64
        region: fsn1
        auto_scaler:
          enabled: true
          min_count: 1
          max_count: 5
  hcloud-k8s:kubernetes:
    hetzner_ccm:
      enabled: true
      version: "1.23.0"
    longhorn:
      enabled: true
```

## Configuration Reference

> **⚠️ Note:** This project is experimental and the configuration API is evolving rapidly. Configuration options may change between versions.

For the most up-to-date configuration options, refer to:

- **Go structs**: See [pkg/config](../pkg/config/) for all available fields and validation rules
- **Cookiecutter options**: See [cookiecutter-options.md](cookiecutter-options.md) for template configuration
- **Examples**: See [examples.md](examples.md) for common configuration patterns

Each configuration struct in the Go code includes detailed documentation and validation rules.
