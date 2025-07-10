# FAQ

**Q: Can I use this as a Go library in my own Pulumi project?**
A: Yes! Import the Go module and use the cluster constructs in your own Pulumi Go code.

**Q: How do I update Talos or Kubernetes versions?**
A: Change the `image_version` and `kubernetes_version` in your stack YAML, then run `pulumi up`.

**Q: How do I add more node pools or change server types?**
A: Edit the `node_pools` section in your stack config.

**Q: Where can I find example configurations?**
A: See [Examples](examples.md) and the sample `Pulumi.*.yaml` files in the repo.

---
For more, open an issue or PR on GitHub.
