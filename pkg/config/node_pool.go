package config

import "github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"

// AutoScalerConfig defines min/max worker count.
type AutoScalerConfig struct {
	MinCount int `json:"min_count" validate:"min=0"`
	MaxCount int `json:"max_count" validate:"gtfield=MinCount"`
}

type Taint struct {
	Key    string `json:"key" validate:"required,excludes=/"`
	Value  string `json:"value" validate:"required"`
	Effect string `json:"effect" validate:"required,oneof=NoSchedule NoExecute PreferNoSchedule"`
}

// NodePoolConfig holds a set of identical worker nodes.
type NodePoolConfig struct {
	Name string `json:"name" validate:"required"`

	// Count is the number of nodes in the pool. Those nodes will are deployed through pulumi and autoscaler can not remove them.
	Count int `json:"count"`

	// AutoScaler is the configuration for the autoscaler.
	AutoScaler *AutoScalerConfig `json:"auto_scaler"`

	ServerSize string                `json:"server_size" validate:"required"`
	Arch       image.CPUArchitecture `json:"arch" validate:"omitempty,oneof=amd64 arm64"`
	Region     string                `json:"region" validate:"required"`

	Labels      map[string]string `json:"labels" validate:"dive,keys,excludes=/"`
	Annotations map[string]string `json:"annotations" validate:"dive,keys,excludes=/"`
	// Taints will only be applied, when a node is created. Update of taints is not supported by talos.
	Taints []Taint `json:"taints" validate:"dive"`
}

// NodePoolsConfig holds a list of worker node pools.
type NodePoolsConfig struct {
	NodePools []NodePoolConfig `json:"node_pools" validate:"dive,required"`

	// SkipAutoScalerDiscovery skips the FindNodePoolAutoScalerNodes call for all node pools.
	// This is useful for development environments where auto-scaler nodes may not exist
	// or when testing without a full cluster setup.
	// WARNING: This should NEVER be enabled in production as it will prevent
	// proper management of auto-scaler created nodes.
	SkipAutoScalerDiscovery bool `json:"skip_auto_scaler_discovery" validate:"env=SKIP_AUTO_SCALER_DISCOVERY"`
}
