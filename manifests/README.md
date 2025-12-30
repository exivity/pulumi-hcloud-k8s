# Talos Extra Manifests Generator

This directory contains tools to generate Kubernetes manifests that can be used as Talos extra manifests during cluster bootstrap.

## Overview

The manifest generator creates properly configured manifests for common Kubernetes components:

- **ArgoCD**: GitOps continuous delivery tool
- **Hetzner CCM**: Cloud Controller Manager for Hetzner Cloud
- **Cilium**: Advanced CNI with optional kube-proxy replacement

## Quick Start

### Generate All Manifests

```bash
cd manifests
make build
```

This generates all component manifests in the `talos-manifests/` directory.

### Generate Specific Components

```bash
# Generate only ArgoCD
make argocd

# Generate Hetzner CCM (both variants)
make hetzner-ccm

# Generate only private network variant
make hetzner-ccm-private

# Generate Cilium without kube-proxy
make cilium-no-kube-proxy
```

## Component Details

### ArgoCD

Generates a complete ArgoCD installation manifest.

**Configuration:**

- `ARGOCD_VERSION`: ArgoCD version (default: v2.13.2)
- `ARGOCD_NAMESPACE`: Namespace for deployment (default: argocd)

**Example:**

```bash
make argocd ARGOCD_VERSION=v2.13.2 ARGOCD_NAMESPACE=my-argocd
```

**Output:** `talos-manifests/argocd.yaml`

### Hetzner Cloud Controller Manager

Generates Hetzner CCM manifests for cloud integration.

**Variants:**

- **Public**: For clusters using public networking only
- **Private**: For clusters using Hetzner private networks

**Configuration:**

- `HETZNER_CCM_NAMESPACE`: Namespace for deployment (default: kube-system)

**Examples:**

```bash
# Both variants
make hetzner-ccm

# Only public network variant
make hetzner-ccm-public

# Only private network variant
make hetzner-ccm-private HETZNER_CCM_NAMESPACE=hetzner-system
```

**Outputs:**

- `talos-manifests/hetzner-ccm-public.yaml`
- `talos-manifests/hetzner-ccm-private.yaml`

### Cilium

Generates Cilium CNI manifests using Helm templates.

**Prerequisites:** Helm must be installed

**Variants:**

- **With kube-proxy**: Cilium + kube-proxy (traditional setup)
- **Without kube-proxy**: Cilium replaces kube-proxy functionality (recommended)

**Configuration:**

- `CILIUM_VERSION`: Cilium version (default: 1.16.5)
- `CILIUM_NAMESPACE`: Namespace for deployment (default: kube-system)

**Examples:**

```bash
# Both variants
make cilium

# Only with kube-proxy
make cilium-kube-proxy CILIUM_VERSION=1.16.5

# Only without kube-proxy (full replacement)
make cilium-no-kube-proxy
```

**Outputs:**

- `talos-manifests/cilium-kube-proxy.yaml`
- `talos-manifests/cilium-no-kube-proxy.yaml`

## Using with Talos

There are two ways to use these manifests with Talos:

### Option 1: Extra Manifests (URLs)

Host the generated manifests on a web server and reference them:

```bash
# Generate manifests
make build

# Host them (example using Python)
cd ../talos-manifests
python3 -m http.server 8000

# Configure in Pulumi
pulumi config set --path hcloud-k8s:talos.extra_manifests[0] http://localhost:8000/argocd.yaml
pulumi config set --path hcloud-k8s:talos.extra_manifests[1] http://localhost:8000/hetzner-ccm-private.yaml
pulumi config set --path hcloud-k8s:talos.extra_manifests[2] http://localhost:8000/cilium-no-kube-proxy.yaml
```

### Option 2: Inline Manifests

Use the manifest contents directly in your Pulumi configuration:

```bash
# Generate manifest
make argocd

# Create inline manifest configuration
cat > inline-argocd.json << 'EOF'
{
  "name": "argocd",
  "contents": "$(cat ../talos-manifests/argocd.yaml | sed 's/"/\\"/g' | tr '\n' ' ')"
}
EOF

# Configure in Pulumi (manual step - edit your YAML config)
# Add to Pulumi.yaml:
# talos:
#   inline_manifests:
#     - name: argocd
#       contents: |
#         <paste manifest content here>
```

### Option 3: From Main Makefile

You can also run the manifest commands from the root of the project:

```bash
# From project root
make manifests          # Generate all manifests
make manifests-help     # Show manifest help
make manifests-clean    # Clean manifests
```

## Configuration Examples

### Complete Setup with Cilium

If you're using Cilium as your CNI, configure Talos accordingly:

```bash
# Set CNI to custom (Cilium will be deployed via extra manifests)
pulumi config set --path hcloud-k8s:talos.cni.name custom
pulumi config set --path hcloud-k8s:talos.cni.urls[0] ""  # Empty since using extra manifests

# Generate and configure Cilium
cd manifests
make cilium-no-kube-proxy

# Host or inline the manifest as shown above
```

### Setup with Hetzner CCM

```bash
# Disable the Helm-based CCM in Kubernetes config
pulumi config set hcloud-k8s:kubernetes.enable_hetzner_ccm false

# Enable via extra manifests instead
pulumi config set hcloud-k8s:talos.enable_hetzner_ccm_extra_manifest false

# Generate and configure Hetzner CCM manually
cd manifests
make hetzner-ccm-private  # or hetzner-ccm-public

# Host or inline the manifest
```

### Complete GitOps Setup with ArgoCD

```bash
# Generate ArgoCD manifest
cd manifests
make argocd

# Configure as extra manifest
pulumi config set --path hcloud-k8s:talos.extra_manifests[0] https://your-server.com/argocd.yaml

# After cluster creation, access ArgoCD
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Get initial password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

## Cleaning Up

```bash
# Clean all generated manifests
make clean

# Or from project root
make manifests-clean
```

## Advanced Usage

### Custom Versions

Each component supports version customization:

```bash
make argocd ARGOCD_VERSION=v2.12.0
make cilium CILIUM_VERSION=1.15.0
```

### Custom Namespaces

Deploy components to custom namespaces:

```bash
make argocd ARGOCD_NAMESPACE=gitops
make hetzner-ccm HETZNER_CCM_NAMESPACE=cloud-controller
make cilium CILIUM_NAMESPACE=networking
```

### Combining Multiple Components

```bash
# Generate all with custom configuration
make argocd ARGOCD_NAMESPACE=gitops
make hetzner-ccm-private
make cilium-no-kube-proxy CILIUM_VERSION=1.16.5

# All manifests are now in talos-manifests/ directory
ls -lh ../talos-manifests/
```

## Troubleshooting

### Helm Not Found (Cilium)

The Cilium component requires Helm to be installed:

```bash
# macOS
brew install helm

# Linux
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### Manifest Download Fails

Ensure you have internet connectivity and the upstream repositories are accessible:

- ArgoCD: <https://github.com/argoproj/argo-cd>
- Hetzner CCM: <https://github.com/hetznercloud/hcloud-cloud-controller-manager>
- Cilium: <https://helm.cilium.io/>

### Namespace Conflicts

If manifests reference different namespaces than expected, ensure you're passing the `NAMESPACE` parameter consistently.

## Additional Resources

- [Talos Extra Manifests Documentation](https://www.talos.dev/v1.7/kubernetes-guides/configuration/deploy-workloads/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Hetzner CCM Documentation](https://github.com/hetznercloud/hcloud-cloud-controller-manager)
- [Cilium Documentation](https://docs.cilium.io/)
