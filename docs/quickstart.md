# Quickstart

This guide helps you get started with a new Hetzner Talos Kubernetes cluster using the Cookiecutter template.

## Prerequisites

- [Go](https://go.dev/doc/install)
- [Pulumi CLI](www.pulumi.com/docs/iac/download-install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
- [talosctl](https://www.talos.dev/v1.10/talos-guides/install/talosctl/)
- [Cookiecutter](https://cookiecutter.readthedocs.io/en/latest/README.html#installation)

## Create a New Project

```sh
cookiecutter https://github.com/exivity/pulumi-hcloud-k8s
cd <your-project-slug>
make download
```

## Configure Your Stack

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

## Deploy

   Note: The `-f` flag skips the preview step for first-time deployment.

```sh
pulumi up --yes -f
```

This first deployment will create the Kubernetes cluster, node pools, and Talos configuration. But it will not install any applications or Helm charts.
For that the cluster must be up and running first. So wait for the deployment to finish (failed) and then proceed with installing applications.

```sh
pulumi up --yes
```

## Access Your Cluster

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
See [Configuration](configuration.md) for all available options.
