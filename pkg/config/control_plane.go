package config

import "github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"

// ControlPlaneConfig defines control‑plane node settings.
type ControlPlaneConfig struct {
	// Disable load balancer deployment.
	//
	// This option is intended for development and testing only. It disables
	// automatic creation of the Hetzner load balancer so clusters can be
	// brought up in environments where a load balancer is not available or
	// desired (for example, single-node development clusters, CI, or local
	// integration tests). It is NOT recommended for production workloads
	// because it bypasses the high‑availability and traffic distribution
	// guarantees provided by the load balancer.
	DisableLoadBalancer bool `json:"disable_load_balancer"`

	// Hetzner load‑balancer type (e.g. "lb11")
	LoadBalancerType string `json:"load_balancer_type" validate:"default=lb11"`

	// Location to create the load balancer in (e.g. "nbg1", "fsn1", "hel1").
	// If not set, the location of the network will be used.
	LoadBalancerLocation *string `json:"load_balancer_location"`

	// Node pool settings
	NodePools []ControlPlaneNodePoolConfig `json:"node_pools" validate:"required"`
}

type ControlPlaneNodePoolConfig struct {
	// Number of control‑plane nodes
	Count int `json:"count" validate:"default=1,min=1"`

	// Hetzner server type (e.g. "cx23", "cax11")
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
