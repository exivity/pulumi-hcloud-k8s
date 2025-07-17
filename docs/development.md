# Development Guide

## Repository Structure

- `main.go` — Pulumi entrypoint
- `pkg/` — Go packages for Hetzner, Talos, Kubernetes
- `Makefile` — Automation for build, lint, test, deploy
- `docs/` — Documentation

## Common Tasks

- **Install dependencies:**

  ```sh
  make download
  ```

- **Format code:**

  ```sh
  make fmt
  ```

- **Lint:**

  ```sh
  make lint
  ```

- **Test:**

  ```sh
  make test
  ```

- **Generate kubeconfig/talosconfig:**

  ```sh
  make kubeconfig
  make talosconfig
  ```

## Testing Cookiecutter Template

```sh
make test-cookiecutter
```

## Contributing

- Fork the repo, create a branch, submit a PR.
- Follow Go best practices and keep code modular.

---
See [Dependencies](dependencies.md) for required tools.
