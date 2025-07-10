# Pulumi Stack Configuration

All cluster options are configured via Pulumi stack YAML files (e.g., `Pulumi.dev.yaml`).

## Main Configuration Sections

- `hcloud-k8s:hetzner`: Hetzner API token
- `hcloud-k8s:network`: VPC, CIDRs, nameservers
- `hcloud-k8s:firewall`: VPN CIDRs, open ports, custom rules
- `hcloud-k8s:talos`: Talos/Kubernetes versions, image settings
- `hcloud-k8s:control_plane`: Control plane node pools
- `hcloud-k8s:node_pools`: Worker node pools
- `hcloud-k8s:kubernetes`: In-cluster components (CCM, CSI, autoscaler, etc)

## Example

```yaml
config:
  hcloud-k8s:hetzner:
    token: <your-hcloud-token>
  hcloud-k8s:talos:
    image_version: v1.10.3
    kubernetes_version: "1.33.0"
  hcloud-k8s:control_plane:
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
  hcloud-k8s:kubernetes:
    hetzner_ccm:
      enabled: true
      version: 1.23.0
```

## All Options

See [pkg/config](../pkg/config/) for all available fields and validation rules. Each config struct is documented in code.

- [HetznerConfig](../pkg/config/hetzner.go)
- [NetworkConfig](../pkg/config/network.go)
- [FirewallConfig](../pkg/config/firewall.go)
- [TalosConfig](../pkg/config/talos.go)
- [ControlPlaneConfig](../pkg/config/control_plane.go)
- [NodePoolsConfig](../pkg/config/node_pool.go)
- [KubernetesConfig](../pkg/config/kubernetes.go)
