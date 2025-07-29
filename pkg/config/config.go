package config

import (
	"github.com/exivity/pulumi-hcloud-k8s/pkg/validators"
	"github.com/exivity/pulumiconfig/pkg/pulumiconfig"
	"github.com/go-playground/validator/v10"
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
			Validate: ValidateHcloudToken,
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

// ValidateHcloudToken checks if HCloudToken is set when required by enabled features.
func ValidateHcloudToken(sl validator.StructLevel) {
	// Attempt to type assert the struct being validated to *PulumiConfig
	cfg, ok := sl.Current().Interface().(PulumiConfig)
	if !ok {
		// If the type assertion fails, report a validation error and return
		sl.ReportError(nil, "", "", "pulumiconfig_type_assertion_failed", "")
		return
	}

	// Only check for HCloudToken if it is nil
	if cfg.Kubernetes.HCloudToken == "" {
		// If HetznerCCM is enabled, HCloudToken is required
		if cfg.Kubernetes.HetznerCCM != nil && cfg.Kubernetes.HetznerCCM.Enabled {
			sl.ReportError(cfg.Kubernetes.HCloudToken, "HCloudToken", "HCloudToken", "required_with_hetznerccm", "")
		}
		// If CSI is enabled, HCloudToken is required
		if cfg.Kubernetes.CSI != nil && cfg.Kubernetes.CSI.Enabled {
			sl.ReportError(cfg.Kubernetes.HCloudToken, "HCloudToken", "HCloudToken", "required_with_csi", "")
		}
		// If ClusterAutoScaler is enabled, HCloudToken is required
		if cfg.Kubernetes.ClusterAutoScaler != nil && cfg.Kubernetes.ClusterAutoScaler.Enabled {
			sl.ReportError(cfg.Kubernetes.HCloudToken, "HCloudToken", "HCloudToken", "required_with_clusterautoscaler", "")
		}
	}
}
