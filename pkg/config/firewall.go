package config

// FirewallRuleConfig allows specifying custom firewall rules (port and allowed CIDRs).
type FirewallRuleConfig struct {
	Port  string   `json:"port" validate:"required"`
	CIDRs []string `json:"cidrs" validate:"dive,cidr"`
}

// FirewallConfig holds settings for Hetzner Cloud Firewall configuration,
// mapping to ControlplaneFirewallArgs and WorkerFirewallArgs in the firewall package.
type FirewallConfig struct {
	// VpnCidrs lists VPN network CIDRs allowed to access control-plane API & trustd.
	// When load balancer is disabled, these CIDRs also control access to the Kubernetes API (port 6443).
	// If empty and load balancer is disabled, Kubernetes API will be exposed to all IPs (0.0.0.0/0, ::/0).
	VpnCidrs []string `json:"vpn_cidrs" validate:"dive,cidr"`

	// OpenTalosAPI opens Talos API to all IPs.
	// Controlplane port: 50000 & 5001
	// Worker port: 50000
	OpenTalosAPI bool `json:"open_talos_api"`

	// CustomRulesControlplane allows opening additional ports to specific CIDRs for control plane nodes (e.g., 80/443 for MetalLB).
	CustomRulesControlplane []FirewallRuleConfig `json:"custom_rules_controlplane"`

	// CustomRulesWorker allows opening additional ports to specific CIDRs for worker nodes (e.g., 80/443 for MetalLB).
	CustomRulesWorker []FirewallRuleConfig `json:"custom_rules_worker"`
}
