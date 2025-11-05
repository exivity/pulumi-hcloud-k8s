package config

// FirewallRuleConfig allows specifying custom firewall rules with full control over direction, protocol, and allowed IPs.
type FirewallRuleConfig struct {
	// Direction of the Firewall Rule. Required. Valid values: "in", "out"
	Direction string `json:"direction" validate:"required,oneof=in out"`

	// Protocol of the Firewall Rule. Required. Valid values: "tcp", "udp", "icmp", "gre", "esp"
	Protocol string `json:"protocol" validate:"required,oneof=tcp udp icmp gre esp"`

	// Description of the firewall rule (optional)
	Description string `json:"description,omitempty"`

	// Port of the Firewall Rule. Required when protocol is tcp or udp.
	// You can use "any" to allow all ports for the specific protocol.
	// Port ranges are also possible: "80-85" allows all ports between 80 and 85.
	Port string `json:"port,omitempty"`

	// SourceIps lists IPs or CIDRs that are allowed within this Firewall Rule (when direction is "in")
	SourceIps []string `json:"source_ips,omitempty" validate:"dive,cidr"`

	// DestinationIps lists IPs or CIDRs that are allowed within this Firewall Rule (when direction is "out")
	DestinationIps []string `json:"destination_ips,omitempty" validate:"dive,cidr"`
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
