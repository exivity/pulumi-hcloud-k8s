package config

// ChartConfig represents a Helm chart override.
type ChartConfig struct {
	Enabled bool                    `json:"enabled"`
	Version *string                 `json:"version"`
	Values  *map[string]interface{} `json:"values"`
}

// CSIChartConfig represents CSI driver configuration.
type CSIChartConfig struct {
	ChartConfig

	// IsDefaultStorageClass is the default storage class
	IsDefaultStorageClass bool `json:"is_default_storage_class"`

	// EncryptedSecret is the encrypted secret used to encrypt the CSI driver
	// This is required when enable CSI driver
	EncryptedSecret string `json:"encrypted_secret" validate:"required"`

	// ReclaimPolicy controls what happens to volumes when PVCs are deleted.
	// Defaults to "Delete" and must be either "Delete" or "Retain".
	ReclaimPolicy string `json:"reclaim_policy" validate:"default=Delete,oneof=Delete Retain"`
}

// KubeletServingCertApproverConfig configures the Kubelet Serving Certificate Approver.
type KubeletServingCertApproverConfig struct {
	Enabled bool `json:"enabled"`
	// Tag or branch of the Kubelet Serving Certificate Approver like v0.9.1 or main
	// See: https://github.com/alex1989hu/kubelet-serving-cert-approver/tags
	Version string `json:"version" default:"main"`
}

// Kubernetes Metrics Server is a lightweight, efficient, and scalable metrics server for Kubernetes.

type KubernetesMetricsServerChartConfig struct {
	ChartConfig
}

// KubernetesConfig configures in‑cluster Hetzner components.
type KubernetesConfig struct {
	// Token passed to CCM, CSI driver and autoscaler
	// This is required when enable CCM, CSI driver or autoscaler
	HCloudToken string `json:"hcloud_token" validate:"env=K8S_HCLOUD_TOKEN"`

	HetznerCCM        *ChartConfig    `json:"hetzner_ccm"`
	CSI               *CSIChartConfig `json:"csi"`
	ClusterAutoScaler *ChartConfig    `json:"cluster_auto_scaler"`
	// Longhorn is the configuration for the Longhorn chart
	// Longhorn needs to be enabled in the Talos config
	Longhorn *ChartConfig `json:"longhorn"`

	// Kubelet Serving Certificate Approver
	KubeletServingCertApprover *KubeletServingCertApproverConfig `json:"kubelet_serving_cert_approver"`

	// Kubernetes Metrics Server
	// Metrics Server is a scalable, efficient source of container resource metrics for Kubernetes built-in autoscaling pipelines.
	// Requires the Kubelet Serving Certificate Approver to be enabled.
	// See: https://github.com/kubernetes-sigs/metrics-server
	KubernetesMetricsServer *KubernetesMetricsServerChartConfig `json:"kubernetes_metrics_server"`
}
