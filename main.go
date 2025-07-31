package main

import (
	"fmt"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/deploy"
	"github.com/exivity/pulumiconfig/pkg/pulumiconfig"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ArgocdConfig struct {
	Enabled bool `json:"enabled" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
}

// ExtendedConfig contains additional configuration fields beyond the base PulumiConfig.
//
// This struct holds extended configuration that's specific to this deployment.
// The base infrastructure configuration is handled separately via config.PulumiConfig.
//
// To extend the configuration, add new fields to this struct.
// See: github.com/exivity/pulumiconfig
type ExtendedConfig struct {
	Argocd ArgocdConfig `json:"argocd" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackName := fmt.Sprintf("%s-%s", ctx.Project(), ctx.Stack())

		// Load base infrastructure configuration
		baseCfg, err := config.LoadConfig(ctx)
		if err != nil {
			return err
		}

		// Load extended configuration
		extendedCfg := &ExtendedConfig{}
		err = pulumiconfig.GetConfig(ctx, extendedCfg)
		if err != nil {
			return err
		}

		cluster, err := deploy.NewHetznerTalosKubernetesCluster(ctx, stackName, baseCfg)
		if err != nil {
			return err
		}

		ctx.Export("kubeconfig", cluster.Kubeconfig.Kubeconfig.KubeconfigRaw)
		ctx.Export("talosconfig", cluster.TalosConfig)
		ctx.Export("argocdEnabled", pulumi.Bool(extendedCfg.Argocd.Enabled))

		return nil
	})
}
