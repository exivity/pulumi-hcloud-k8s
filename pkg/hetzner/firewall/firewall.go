package firewall

import (
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ControlplaneFirewallArgs holds parameters for the control-plane firewall.
type ControlplaneFirewallArgs struct {
	// VpnCidrs lists VPN network CIDRs allowed to access control-plane API & trustd.
	VpnCidrs []string

	// OpenAPIToEveryone opens control-plane API Talos API & trustd (50000-50001) to all IPs.
	OpenAPIToEveryone bool

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
	}

	// Optionally open to all
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

	// Add custom rules (e.g., for MetalLB)
	for _, rule := range args.CustomRules {
		if len(rule.CIDRs) == 0 || rule.Port == "" {
			continue // skip invalid
		}
		rules = append(rules, &hcloud.FirewallRuleArgs{
			Description: pulumi.Sprintf("Custom: Open port %s", rule.Port),
			Direction:   pulumi.String("in"),
			Protocol:    pulumi.String("tcp"),
			Port:        pulumi.String(rule.Port),
			SourceIps:   toPulumiIPList(rule.CIDRs),
		})
	}

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
	for _, rule := range args.CustomRules {
		if len(rule.CIDRs) == 0 || rule.Port == "" {
			continue // skip invalid
		}
		rules = append(rules, &hcloud.FirewallRuleArgs{
			Description: pulumi.Sprintf("Custom: Open port %s", rule.Port),
			Direction:   pulumi.String("in"),
			Protocol:    pulumi.String("tcp"),
			Port:        pulumi.String(rule.Port),
			SourceIps:   toPulumiIPList(rule.CIDRs),
		})
	}

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

// CustomFirewallRuleArg represents a custom port and allowed CIDRs.
type CustomFirewallRuleArg struct {
	Port  string
	CIDRs []string
}

// toPulumiIPList converts a slice of CIDRs to pulumi.StringArray.
func toPulumiIPList(cidrs []string) pulumi.StringArray {
	arr := pulumi.StringArray{}
	for _, c := range cidrs {
		arr = append(arr, pulumi.String(c))
	}
	return arr
}
