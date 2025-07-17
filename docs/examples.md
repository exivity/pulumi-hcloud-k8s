# Examples

## Minimal Stack Configuration

```yaml
config:
  hcloud-k8s:hetzner:
    token: <your-hcloud-token>
  hcloud-k8s:talos:
    image_version: v1.10.3
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

## Advanced: Multiple Node Pools, Autoscaler, Longhorn

```yaml
config:
  hcloud-k8s:hetzner:
    token: <your-hcloud-token>
  hcloud-k8s:talos:
    image_version: v1.10.3
    kubernetes_version: "1.33.0"
    enable_longhorn: true
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
        auto_scaler:
          min_count: 1
          max_count: 3
  hcloud-k8s:kubernetes:
    hetzner_ccm:
      enabled: true
      version: 1.23.0
    longhorn:
      enabled: true
      version: 1.8.1
```

See [Configuration](configuration.md) for all options.
