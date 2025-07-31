package main

import (
	"fmt"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/deploy"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackName := fmt.Sprintf("%s-%s", ctx.Project(), ctx.Stack())

		// Load base infrastructure configuration
		baseCfg, err := config.LoadConfig(ctx)
		if err != nil {
			return err
		}

		cluster, err := deploy.NewHetznerTalosKubernetesCluster(ctx, stackName, baseCfg)
		if err != nil {
			return err
		}

		ctx.Export("kubeconfig", cluster.Kubeconfig.Kubeconfig.KubeconfigRaw)
		ctx.Export("talosconfig", cluster.TalosConfig)

		return nil
	})
}
