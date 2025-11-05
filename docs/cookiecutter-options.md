# Cookiecutter Template Options

This document describes all the available options when using the cookiecutter template to generate a new Pulumi Hetzner Kubernetes (Talos) project.

## Basic Project Configuration

### `project_name`

**Default:** `"My Awesome Project"`
**Description:** The human-readable name of your project. This will be converted to a slug for file names and directory structure.

### `project_slug`

**Auto-generated from:** `project_name`
**Description:** URL-friendly version of the project name, used for directory and file naming.

### `description`

**Default:** `"Behold My Awesome Project!"`
**Description:** A brief description of your project for documentation purposes.

### `go_module_path`

**Default:** `"github.com/my-org/{{ cookiecutter.project_slug }}"`
**Description:** The Go module path for your project. Update the organization name as needed.

## Pulumi Configuration

### `pulumi_project_name`

**Default:** `"{{ cookiecutter.project_slug }}"`
**Description:** The name of the Pulumi project.

### `pulumi_org`

**Default:** `""` (empty string)
**Description:** The Pulumi organization name. If not set, it will use not a org. Only relevant when using pulumi.com backend in a org. Can be updated after project creation.

### `pulumi_stack`

**Default:** `"dev"`
**Description:** The default Pulumi stack name.

## Hetzner Configuration

### `hetzner_token`

**Default:** `""` (empty string)
**Description:** Your Hetzner Cloud API token for managing resources. If not set, it will try to use the env var `HCLOUD_TOKEN`. If both are not set, the deployment will fail. Can be updated after project creation.

### `hetzner_cluster_token`

**Default:** `""` (empty string)
**Description:** Your Hetzner Cloud API token for deploying Kubernetes resources (CCM, CSI, Cluster Autoscaler). This is required when enabling any Kubernetes features that integrate with Hetzner Cloud. You can use the same token as `hetzner_token`. If not set, you must configure it after project creation using `pulumi config set --path hcloud-k8s:kubernetes.hcloud_token --secret`.

## Talos & Kubernetes Configuration

### `talos_api_allowed_cidrs`

**Default:** `""` (empty string)
**Description:** Comma-separated list of CIDR blocks allowed to access the Talos API. If empty, Talos API will be open to all IPs. Example: `"10.0.0.0/8,192.168.1.0/24"`. Can be updated after project creation.

### `talos_version`

**Default:** `"v1.11.3"`
**Description:** The version of Talos Linux to use. See the [Talos Support Matrix](https://www.talos.dev/v1.11/introduction/support-matrix/) for supported Kubernetes versions.

### `kubernetes_version`

**Default:** `"1.34.0"`
**Description:** The Kubernetes version to deploy. Must be compatible with the selected Talos version.

## Control Plane Configuration

### `controlplane_enable_ha`

**Default:** `false`
**Description:** Enable high availability control plane across multiple regions (fsn1, nbg1, hel1). When enabled, deploys 3 control plane nodes across different regions.

### `controlplane_disable_load_balancer`

**Default:** `false`
**Description:** Disable the control plane load balancer. When set to `true`, the cluster will be configured to access the Kubernetes API directly via the first control plane node's IP address instead of through a load balancer. This is useful for single-node clusters or cost-sensitive deployments. Note: When the load balancer is disabled, the Kubernetes API port (6443) will be exposed through the firewall rules.

### `controlplane_server_size`

**Default:** `"cax11"`
**Description:** The Hetzner server type for control plane nodes. Available options include cax11 (ARM64, cost-effective) and other Hetzner server types.

## Worker Pool Configuration

### `worker_pool_name`

**Default:** `"worker"`
**Description:** The name of the default worker node pool.

### `worker_pool_count`

**Default:** `2`
**Description:** The initial number of worker nodes managed by Pulumi.

### `worker_pool_server_size`

**Default:** `"cax31"`
**Description:** The Hetzner server type for worker nodes.

### `worker_pool_region`

**Default:** `"hel1"`
**Description:** The region where worker nodes will be deployed. Available options include hel1, fsn1, nbg1, and ash1.

### `worker_pool_auto_scale_max`

**Default:** `3`
**Description:** Maximum number of nodes the autoscaler can scale up to when cluster autoscaler is enabled.

## Kubernetes Components

### `enable_longhorn`

**Default:** `false`
**Description:** Enable Longhorn distributed storage for persistent volumes with replication, snapshots, and backup capabilities.

### `enable_hetzner_csi`

**Default:** `true`
**Description:** Enable the Hetzner CSI driver for persistent volume support with native Hetzner Cloud integration.

### `enable_cluster_autoscaler`

**Default:** `true`
**Description:** Enable the Kubernetes Cluster Autoscaler component for automatic node scaling based on resource demands.

### `enable_kubelet_cert_approver`

**Default:** `true`
**Description:** Enable automatic approval of kubelet serving certificates for secure metrics collection.

### `enable_metrics_server`

**Default:** `true`
**Description:** Enable the Kubernetes Metrics Server for Horizontal Pod Autoscaling (HPA) and resource monitoring.

## Usage Examples

### Development Cluster

```bash
cookiecutter https://github.com/exivity/pulumi-hcloud-k8s \
  --no-input \
  project_name="My Dev Cluster" \
  controlplane_enable_ha=false \
  worker_pool_count=1 \
  enable_longhorn=false
```

### Production Cluster

```bash
cookiecutter https://github.com/exivity/pulumi-hcloud-k8s \
  --no-input \
  project_name="Production Cluster" \
  controlplane_enable_ha=true \
  worker_pool_count=3 \
  worker_pool_auto_scale_max=10 \
  enable_longhorn=true
```

### High-Performance Cluster

```bash
cookiecutter https://github.com/exivity/pulumi-hcloud-k8s \
  --no-input \
  project_name="HPC Cluster" \
  controlplane_server_size=cax31 \
  worker_pool_server_size=cax41 \
  enable_longhorn=true
```

## Notes

- **Version Compatibility:** Always check the [Talos Support Matrix](https://www.talos.dev/v1.10/introduction/support-matrix/) for the latest supported Kubernetes versions.
- **Cost Optimization:** ARM64 instances (cax series) are generally more cost-effective than x86 instances.
- **High Availability:** For production deployments, enable `controlplane_enable_ha=true` for multi-region control planes.
- **Security:** Configure `talos_api_allowed_cidrs` to restrict Talos API access to specific IP ranges for production environments.
- **Storage:** Longhorn provides distributed storage but requires at least 3 worker nodes for proper operation.
