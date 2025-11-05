package main

import (
	"fmt"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/deploy"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ExtendedConfig contains additional configuration fields beyond the base PulumiConfig.
//
// This struct holds extended configuration that's specific to this deployment.
// The base infrastructure configuration is handled separately via config.PulumiConfig.
//
// To extend the configuration, add new fields to this struct.
// See: github.com/exivity/pulumiconfig
type ExtendedConfig struct {
	// Add your extended configuration fields here
	// Example: MyCustomField MyCustomConfig `json:"my_custom_field" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackName := fmt.Sprintf("%s-%s", ctx.Project(), ctx.Stack())

		// Load base infrastructure configuration
		baseCfg, err := config.LoadConfig(ctx)
		if err != nil {
			return err
		}

		// Optional: Modify baseCfg after loading from Pulumi stack
		// This allows you to override or extend configuration programmatically
		// Examples:
		// 1. Override specific values:
		//    baseCfg.Talos.ImageVersion = "v1.11.3"
		//    baseCfg.ControlPlane.NodePools[0].Count = 5
		//
		// 2. Add conditional logic:
		//    if ctx.Stack() == "production" {
		//        baseCfg.ControlPlane.LoadBalancerType = "lb31"
		//    }
		//
		// 3. Alternative: Skip pulumiconfig entirely and build config programmatically:
		//    baseCfg := &config.PulumiConfig{
		//        Hetzner: config.HetznerConfig{Token: "your-token"},
		//        Talos: config.TalosConfig{ImageVersion: "v1.11.3"},
		//        // ... other fields
		//    }

		// Load extended configuration (uncomment and modify when you add extended fields)
		// extendedCfg := &ExtendedConfig{}
		// err = pulumiconfig.GetConfig(ctx, extendedCfg)
		// if err != nil {
		//     return err
		// }

		cluster, err := deploy.NewHetznerTalosKubernetesCluster(ctx, stackName, baseCfg)
		if err != nil {
			return err
		}

		ctx.Export("kubeconfig", cluster.Kubeconfig.Kubeconfig.KubeconfigRaw)
		ctx.Export("talosconfig", cluster.TalosConfig)

		return nil
	})
}
