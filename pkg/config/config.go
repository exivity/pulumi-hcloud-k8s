package config

import (
	"github.com/exivity/pulumi-hcloud-k8s/pkg/validators"
	"github.com/exivity/pulumiconfig/pkg/pulumiconfig"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Config is the root of the entire deployment manifest.
type PulumiConfig struct {
	Hetzner      HetznerConfig      `json:"hetzner" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
	Network      NetworkConfig      `json:"network" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
	Firewall     FirewallConfig     `json:"firewall" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
	Talos        TalosConfig        `json:"talos" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
	ControlPlane ControlPlaneConfig `json:"control_plane" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
	NodePools    NodePoolsConfig    `json:"node_pools" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
	Kubernetes   KubernetesConfig   `json:"kubernetes" pulumiConfigNamespace:"hcloud-k8s-esc" overrideConfigNamespace:"hcloud-k8s"`
}

// LoadConfig loads the config from the pulumi stack.
func LoadConfig(ctx *pulumi.Context) (*PulumiConfig, error) {
	cfg := &PulumiConfig{}
	err := pulumiconfig.GetConfig(ctx, cfg, GetCustomValidations()...)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func GetCustomValidations() []pulumiconfig.Validator {
	return []pulumiconfig.Validator{
		pulumiconfig.StructValidation{
			Struct:   PulumiConfig{},
			Validate: validators.ValidateHcloudToken,
		},
		pulumiconfig.StructValidation{
			Struct:   ControlPlaneConfig{},
			Validate: validators.ValidateAndSetArchForControlPlane,
		},
		pulumiconfig.StructValidation{
			Struct:   NodePoolConfig{},
			Validate: validators.ValidateAndSetArchForNodePool,
		},
	}
}
