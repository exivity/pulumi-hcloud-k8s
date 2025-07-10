package config

// HetznerConfig configures hetzner related parts
type HetznerConfig struct {
	// API token for Hetzner Cloud
	Token string `json:"token" validate:"env=HCLOUD_TOKEN"`
}
