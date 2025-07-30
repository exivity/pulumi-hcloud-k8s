package config

import "github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"

// ControlPlaneConfig defines control‑plane node settings.
type ControlPlaneConfig struct {
	// Hetzner load‑balancer type (e.g. "lb11")
	LoadBalancerType string `json:"load_balancer_type" validate:"default=lb11"`

	// Node pool settings
	NodePools []ControlPlaneNodePoolConfig `json:"node_pools" validate:"required"`
}

type ControlPlaneNodePoolConfig struct {
	// Number of control‑plane nodes
	Count int `json:"count" validate:"default=1,min=1"`

	// Hetzner server type (e.g. "cx22", "cax11")
	ServerSize string `json:"server_size" validate:"required"`

	// CPU architecture
	Arch image.CPUArchitecture `json:"arch" validate:"omitempty,oneof=amd64 arm64"`

	// Hetzner region
	Region string `json:"region" validate:"required"`

	// Daily backups, kept 7 days
	EnableBackup bool `json:"enable_backup"`

	Labels      map[string]string `json:"labels" validate:"dive,keys,excludes=/"`
	Annotations map[string]string `json:"annotations" validate:"dive,keys,excludes=/"`
	// Taints will only be applied, when a node is created. Update of taints is not supported by talos.
	Taints []Taint `json:"taints" validate:"dive"`
}
