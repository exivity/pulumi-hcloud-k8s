# Examples

> **🔐 Security Note:** All examples show realistic YAML configuration files. Hetzner tokens must be set separately as secrets using:
>
> ```sh
> # Infrastructure token for managing Hetzner Cloud resources
> pulumi config set --path hcloud-k8s:hetzner.token --secret
> # Kubernetes token for CCM, CSI, and Cluster Autoscaler (can be the same token)
> pulumi config set --path hcloud-k8s:kubernetes.hcloud_token --secret
> ```
>
> These commands will prompt you to enter the token values securely. Both tokens are required when using Kubernetes features.

## Minimal Stack Configuration

```yaml
config:
  # Tokens are set as secrets via CLI and won't appear in this file
  hcloud-k8s:talos:
    image_version: v1.10.5
    kubernetes_version: "1.33.0"
  hcloud-k8s:control_plane:
    node_pools:
      - count: 1
        server_size: cax11
        arch: arm64
        region: fsn1
  hcloud-k8s:node_pools:
    node_pools:
      - name: core
        count: 1
        server_size: cx22
        arch: amd64
        region: fsn1
```

## Production: High Availability with Multiple Node Pools

```yaml
config:
  # Tokens are set as secrets via CLI and won't appear in this file
  hcloud-k8s:talos:
    image_version: v1.10.5
    kubernetes_version: "1.33.0"
    api_allowed_cidrs: "10.0.0.0/8,192.168.0.0/16"
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
        server_size: cx22
        arch: amd64
        region: fsn1
        auto_scaler:
          enabled: true
          min_count: 1
          max_count: 5
      - name: compute
        count: 1
        server_size: cax41
        arch: arm64
        region: nbg1
        auto_scaler:
          enabled: true
          min_count: 0
          max_count: 3
  hcloud-k8s:kubernetes:
    hetzner_ccm:
      enabled: true
    hetzner_csi:
      enabled: true
    cluster_autoscaler:
      enabled: true
    metrics_server:
      enabled: true
    kubelet_cert_approver:
      enabled: true
```

## Advanced: With Longhorn Storage

```yaml
config:
  # Tokens are set as secrets via CLI and won't appear in this file
  hcloud-k8s:talos:
    image_version: v1.10.5
    kubernetes_version: "1.33.0"
  hcloud-k8s:control_plane:
    enable_ha: true
    node_pools:
      - count: 3
        server_size: cax11
        arch: arm64
        region: fsn1
  hcloud-k8s:node_pools:
    node_pools:
      - name: storage
        count: 3
        server_size: cx22
        arch: amd64
        region: fsn1
        auto_scaler:
          enabled: true
          min_count: 3
          max_count: 6
  hcloud-k8s:kubernetes:
    hetzner_ccm:
      enabled: true
    longhorn:
      enabled: true
    cluster_autoscaler:
      enabled: true
```

## Notes

- **Security**: Always use `pulumi config set --path <path> --secret` to set Hetzner tokens. Never store tokens in plain text YAML files
- **Token Requirements**: Both `hetzner.token` and `kubernetes.hcloud_token` must be set when using Kubernetes features. You can use the same token value for both
- **Versions**: Always check the [Talos Support Matrix](https://www.talos.dev/v1.10/introduction/support-matrix/) for compatible Kubernetes versions
- **Architecture**: ARM64 instances (cax series) are generally more cost-effective
- **Longhorn**: Requires at least 3 worker nodes for proper replication
- **High Availability**: Enable `control_plane.enable_ha` for production deployments

See [Configuration](configuration.md) for all available options.
