package main

import (
	"fmt"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/deploy"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ArgocdConfig struct {
	Enabled bool `json:"enabled" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
}

// PulumiConfig wraps the base configuration structure for the Pulumi stack.
//
// It embeds config.PulumiConfig to provide a consistent configuration interface
// across the application while allowing for future extensibility.
//
// To extend the configuration, add new fields to this struct.
// See: github.com/exivity/pulumiconfig
type PulumiConfig struct {
	config.PulumiConfig

	Argocd ArgocdConfig `json:"argocd" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackName := fmt.Sprintf("%s-%s", ctx.Project(), ctx.Stack())

		cfg, err := config.LoadConfig(ctx, &PulumiConfig{})
		if err != nil {
			return err
		}

		cluster, err := deploy.NewHetznerTalosKubernetesCluster(ctx, stackName, &cfg.PulumiConfig)
		if err != nil {
			return err
		}

		ctx.Export("kubeconfig", cluster.Kubeconfig.Kubeconfig.KubeconfigRaw)
		ctx.Export("talosconfig", cluster.TalosConfig)

		ctx.Export("argocdEnabled", pulumi.Bool(cfg.Argocd.Enabled))

		return nil
	})
}
