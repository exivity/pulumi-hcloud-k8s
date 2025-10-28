package firewall

import (
	"errors"
	"fmt"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	// ErrDirectionRequired is returned when direction field is missing
	ErrDirectionRequired = errors.New("direction is required")
	// ErrProtocolRequired is returned when protocol field is missing
	ErrProtocolRequired = errors.New("protocol is required")
	// ErrPortRequired is returned when port is required but missing
	ErrPortRequired = errors.New("port is required for tcp/udp protocol")
	// ErrSourceIPsRequired is returned when source_ips is required but missing
	ErrSourceIPsRequired = errors.New("source_ips is required for direction 'in'")
	// ErrDestinationIPsRequired is returned when destination_ips is required but missing
	ErrDestinationIPsRequired = errors.New("destination_ips is required for direction 'out'")
)

// ControlplaneFirewallArgs holds parameters for the control-plane firewall.
type ControlplaneFirewallArgs struct {
	// VpnCidrs lists VPN network CIDRs allowed to access control-plane API & trustd.
	VpnCidrs []string

	// OpenAPIToEveryone opens control-plane API Talos API & trustd (50000-50001) to all IPs.
	OpenAPIToEveryone bool

	// ExposeKubernetesAPIWithoutLoadBalancer enables Kubernetes API (port 6443) access for non-loadbalancer setups.
	// This should ONLY be enabled when control plane is configured without a load balancer.
	// Behavior:
	//   - If VpnCidrs are configured: port 6443 will be exposed only to those specific CIDRs
	//   - If VpnCidrs are empty: port 6443 will be exposed to everyone (0.0.0.0/0, ::/0)
	ExposeKubernetesAPIWithoutLoadBalancer bool

	// CustomRules allows opening additional ports to specific CIDRs (e.g., 80/443 for MetalLB).
	CustomRules []CustomFirewallRuleArg
}

// NewControlplaneFirewall creates an Hetzner firewall for control-plane nodes.
func NewControlplaneFirewall(ctx *pulumi.Context, name string, args *ControlplaneFirewallArgs, opts ...pulumi.ResourceOption) (*hcloud.Firewall, error) {
	rules := hcloud.FirewallRuleArray{}

	// VPN access for control-plane ports
	for _, cidr := range args.VpnCidrs {
		src := toPulumiIPList([]string{cidr})
		for _, p := range []string{"50000", "50001"} {
			rules = append(rules, &hcloud.FirewallRuleArgs{
				Description: pulumi.String("VPN: Open Talos API"),
				Direction:   pulumi.String("in"),
				Protocol:    pulumi.String("tcp"),
				Port:        pulumi.String(p), SourceIps: src,
			})
		}
		// Also open Kubernetes API (6443) to VPN CIDRs for non-loadbalancer setup
		if args.ExposeKubernetesAPIWithoutLoadBalancer {
			rules = append(rules, &hcloud.FirewallRuleArgs{
				Description: pulumi.String("VPN: Open Kubernetes API (non-LB setup)"),
				Direction:   pulumi.String("in"),
				Protocol:    pulumi.String("tcp"),
				Port:        pulumi.String("6443"),
				SourceIps:   src,
			})
		}
	}

	// Optionally open Talos API to all
	if args.OpenAPIToEveryone {
		all := toPulumiIPList([]string{"0.0.0.0/0", "::/0"})
		for _, p := range []string{"50000", "50001"} {
			rules = append(rules, &hcloud.FirewallRuleArgs{
				Description: pulumi.String("Public: Open Talos API"),
				Direction:   pulumi.String("in"),
				Protocol:    pulumi.String("tcp"),
				Port:        pulumi.String(p),
				SourceIps:   all,
			})
		}
	}

	// Open Kubernetes API for non-loadbalancer setup when no VPN CIDRs are configured
	// If VPN CIDRs exist, the API is already exposed to them in the loop above
	if args.ExposeKubernetesAPIWithoutLoadBalancer && len(args.VpnCidrs) == 0 {
		all := toPulumiIPList([]string{"0.0.0.0/0", "::/0"})
		rules = append(rules, &hcloud.FirewallRuleArgs{
			Description: pulumi.String("Public: Open Kubernetes API (non-LB setup)"),
			Direction:   pulumi.String("in"),
			Protocol:    pulumi.String("tcp"),
			Port:        pulumi.String("6443"),
			SourceIps:   all,
		})
	}

	// Add custom rules (e.g., for MetalLB)
	customRules, err := processCustomRules(args.CustomRules, "controlplane")
	if err != nil {
		return nil, err
	}
	rules = append(rules, customRules...)

	// Create the firewall
	fw, err := hcloud.NewFirewall(ctx, name, &hcloud.FirewallArgs{
		Name:   pulumi.String(name),
		Labels: meta.NewLabels(ctx, &meta.ServerLabelsArgs{ServerNodeType: meta.ControlPlaneNode}),
		Rules:  rules,
	}, opts...)
	if err != nil {
		return nil, err
	}
	return fw, nil
}

// WorkerFirewallArgs holds parameters for the worker-node firewall.
type WorkerFirewallArgs struct {
	// VpnCidrs lists VPN network CIDRs allowed to access Talos API.
	VpnCidrs []string

	// OpenAPIToEveryone opens Talos API (port 50000) to all IPs.
	OpenAPIToEveryone bool

	// CustomRules allows opening additional ports to specific CIDRs (e.g., 80/443 for MetalLB).
	CustomRules []CustomFirewallRuleArg
}

// NewWorkerFirewall creates an Hetzner firewall for worker nodes.
func NewWorkerFirewall(ctx *pulumi.Context, name string, args *WorkerFirewallArgs, opts ...pulumi.ResourceOption) (*hcloud.Firewall, error) {
	rules := hcloud.FirewallRuleArray{}

	// VPN access to Talos API
	for _, cidr := range args.VpnCidrs {
		rules = append(rules, &hcloud.FirewallRuleArgs{
			Description: pulumi.String("VPN: Open Talos API"),
			Direction:   pulumi.String("in"),
			Protocol:    pulumi.String("tcp"),
			Port:        pulumi.String("50000"),
			SourceIps:   toPulumiIPList([]string{cidr})},
		)
	}

	// Open Talos API to everyone
	if args.OpenAPIToEveryone {
		all := toPulumiIPList([]string{"0.0.0.0/0", "::/0"})
		rules = append(rules, &hcloud.FirewallRuleArgs{
			Description: pulumi.String("Public: Open Talos API"),
			Direction:   pulumi.String("in"),
			Protocol:    pulumi.String("tcp"),
			Port:        pulumi.String("50000"),
			SourceIps:   all,
		})
	}

	// Add custom rules (e.g., for MetalLB)
	customRules, err := processCustomRules(args.CustomRules, "worker")
	if err != nil {
		return nil, err
	}
	rules = append(rules, customRules...)

	fw, err := hcloud.NewFirewall(ctx, name, &hcloud.FirewallArgs{
		Name:   pulumi.String(name),
		Labels: meta.NewLabels(ctx, &meta.ServerLabelsArgs{ServerNodeType: meta.WorkerNode}),
		Rules:  rules,
	}, opts...)
	if err != nil {
		return nil, err
	}
	return fw, nil
}

// CustomFirewallRuleArg represents a custom firewall rule with full control over direction, protocol, and allowed IPs.
type CustomFirewallRuleArg struct {
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

// ToCustomFirewallRuleArgs converts config.FirewallRuleConfig to CustomFirewallRuleArg
func ToCustomFirewallRuleArgs(rules []config.FirewallRuleConfig) []CustomFirewallRuleArg {
	out := make([]CustomFirewallRuleArg, 0, len(rules))
	for _, r := range rules {
		out = append(out, CustomFirewallRuleArg{
			Direction:      r.Direction,
			Protocol:       r.Protocol,
			Description:    r.Description,
			Port:           r.Port,
			SourceIps:      r.SourceIps,
			DestinationIps: r.DestinationIps,
		})
	}
	return out
}

// processCustomRules validates and converts custom firewall rules to Pulumi firewall rule args
func processCustomRules(rules []CustomFirewallRuleArg, nodeType string) (hcloud.FirewallRuleArray, error) {
	result := hcloud.FirewallRuleArray{}

	for i, rule := range rules {
		if err := validateCustomRule(rule, nodeType, i); err != nil {
			return nil, err
		}

		ruleArgs, err := buildFirewallRuleArgs(rule, nodeType, i)
		if err != nil {
			return nil, err
		}

		result = append(result, ruleArgs)
	}

	return result, nil
}

// validateCustomRule validates required fields for a custom firewall rule
func validateCustomRule(rule CustomFirewallRuleArg, nodeType string, index int) error {
	if rule.Direction == "" {
		return fmt.Errorf("%s custom rule %d: %w", nodeType, index, ErrDirectionRequired)
	}
	if rule.Protocol == "" {
		return fmt.Errorf("%s custom rule %d: %w", nodeType, index, ErrProtocolRequired)
	}
	// Port is required for tcp/udp protocols
	if (rule.Protocol == "tcp" || rule.Protocol == "udp") && rule.Port == "" {
		return fmt.Errorf("%s custom rule %d: %w", nodeType, index, ErrPortRequired)
	}
	return nil
}

// buildFirewallRuleArgs constructs a Pulumi firewall rule from a custom rule
func buildFirewallRuleArgs(rule CustomFirewallRuleArg, nodeType string, index int) (*hcloud.FirewallRuleArgs, error) {
	ruleArgs := &hcloud.FirewallRuleArgs{
		Direction: pulumi.String(rule.Direction),
		Protocol:  pulumi.String(rule.Protocol),
	}

	// Set description
	if rule.Description != "" {
		ruleArgs.Description = pulumi.String(rule.Description)
	} else {
		ruleArgs.Description = pulumi.Sprintf("Custom: %s %s", rule.Protocol, rule.Port)
	}

	// Set port if applicable
	if rule.Port != "" {
		ruleArgs.Port = pulumi.String(rule.Port)
	}

	// Set source or destination IPs based on direction
	if err := setRuleIPs(ruleArgs, rule, nodeType, index); err != nil {
		return nil, err
	}

	return ruleArgs, nil
}

// setRuleIPs sets source or destination IPs based on rule direction
func setRuleIPs(ruleArgs *hcloud.FirewallRuleArgs, rule CustomFirewallRuleArg, nodeType string, index int) error {
	switch rule.Direction {
	case "in":
		if len(rule.SourceIps) == 0 {
			return fmt.Errorf("%s custom rule %d: %w", nodeType, index, ErrSourceIPsRequired)
		}
		ruleArgs.SourceIps = toPulumiIPList(rule.SourceIps)
	case "out":
		if len(rule.DestinationIps) == 0 {
			return fmt.Errorf("%s custom rule %d: %w", nodeType, index, ErrDestinationIPsRequired)
		}
		ruleArgs.DestinationIps = toPulumiIPList(rule.DestinationIps)
	}
	return nil
}

// toPulumiIPList converts a slice of CIDRs to pulumi.StringArray.
func toPulumiIPList(cidrs []string) pulumi.StringArray {
	arr := pulumi.StringArray{}
	for _, c := range cidrs {
		arr = append(arr, pulumi.String(c))
	}
	return arr
}
