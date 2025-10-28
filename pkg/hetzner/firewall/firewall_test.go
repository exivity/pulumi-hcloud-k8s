package firewall_test

import (
	"reflect"
	"testing"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/firewall"
)

func TestToCustomFirewallRuleArgs(t *testing.T) {
	tests := []struct {
		name  string // description of this test case
		rules []config.FirewallRuleConfig
		want  []firewall.CustomFirewallRuleArg
	}{
		{
			name:  "empty rules",
			rules: []config.FirewallRuleConfig{},
			want:  []firewall.CustomFirewallRuleArg{},
		},
		{
			name: "single inbound TCP rule",
			rules: []config.FirewallRuleConfig{
				{
					Direction:   "in",
					Protocol:    "tcp",
					Description: "Allow HTTP",
					Port:        "80",
					SourceIps:   []string{"0.0.0.0/0", "::/0"},
				},
			},
			want: []firewall.CustomFirewallRuleArg{
				{
					Direction:   "in",
					Protocol:    "tcp",
					Description: "Allow HTTP",
					Port:        "80",
					SourceIps:   []string{"0.0.0.0/0", "::/0"},
				},
			},
		},
		{
			name: "single outbound UDP rule",
			rules: []config.FirewallRuleConfig{
				{
					Direction:      "out",
					Protocol:       "udp",
					Description:    "Allow DNS",
					Port:           "53",
					DestinationIps: []string{"8.8.8.8/32"},
				},
			},
			want: []firewall.CustomFirewallRuleArg{
				{
					Direction:      "out",
					Protocol:       "udp",
					Description:    "Allow DNS",
					Port:           "53",
					DestinationIps: []string{"8.8.8.8/32"},
				},
			},
		},
		{
			name: "multiple rules with different protocols",
			rules: []config.FirewallRuleConfig{
				{
					Direction: "in",
					Protocol:  "tcp",
					Port:      "443",
					SourceIps: []string{"10.0.0.0/8"},
				},
				{
					Direction: "in",
					Protocol:  "icmp",
					SourceIps: []string{"0.0.0.0/0"},
				},
				{
					Direction: "in",
					Protocol:  "udp",
					Port:      "8080-8090",
					SourceIps: []string{"192.168.1.0/24"},
				},
			},
			want: []firewall.CustomFirewallRuleArg{
				{
					Direction: "in",
					Protocol:  "tcp",
					Port:      "443",
					SourceIps: []string{"10.0.0.0/8"},
				},
				{
					Direction: "in",
					Protocol:  "icmp",
					SourceIps: []string{"0.0.0.0/0"},
				},
				{
					Direction: "in",
					Protocol:  "udp",
					Port:      "8080-8090",
					SourceIps: []string{"192.168.1.0/24"},
				},
			},
		},
		{
			name: "rule with port range",
			rules: []config.FirewallRuleConfig{
				{
					Direction:   "in",
					Protocol:    "tcp",
					Description: "Allow port range",
					Port:        "80-85",
					SourceIps:   []string{"0.0.0.0/0"},
				},
			},
			want: []firewall.CustomFirewallRuleArg{
				{
					Direction:   "in",
					Protocol:    "tcp",
					Description: "Allow port range",
					Port:        "80-85",
					SourceIps:   []string{"0.0.0.0/0"},
				},
			},
		},
		{
			name: "rule with GRE protocol",
			rules: []config.FirewallRuleConfig{
				{
					Direction: "in",
					Protocol:  "gre",
					SourceIps: []string{"10.0.0.1/32"},
				},
			},
			want: []firewall.CustomFirewallRuleArg{
				{
					Direction: "in",
					Protocol:  "gre",
					SourceIps: []string{"10.0.0.1/32"},
				},
			},
		},
		{
			name: "rule with ESP protocol",
			rules: []config.FirewallRuleConfig{
				{
					Direction:   "in",
					Protocol:    "esp",
					Description: "IPsec ESP",
					SourceIps:   []string{"172.16.0.0/16"},
				},
			},
			want: []firewall.CustomFirewallRuleArg{
				{
					Direction:   "in",
					Protocol:    "esp",
					Description: "IPsec ESP",
					SourceIps:   []string{"172.16.0.0/16"},
				},
			},
		},
		{
			name: "rule with both IPv4 and IPv6",
			rules: []config.FirewallRuleConfig{
				{
					Direction: "in",
					Protocol:  "tcp",
					Port:      "22",
					SourceIps: []string{"0.0.0.0/0", "::/0", "10.0.0.0/8"},
				},
			},
			want: []firewall.CustomFirewallRuleArg{
				{
					Direction: "in",
					Protocol:  "tcp",
					Port:      "22",
					SourceIps: []string{"0.0.0.0/0", "::/0", "10.0.0.0/8"},
				},
			},
		},
		{
			name: "outbound rule with multiple destinations",
			rules: []config.FirewallRuleConfig{
				{
					Direction:      "out",
					Protocol:       "tcp",
					Port:           "443",
					DestinationIps: []string{"1.1.1.1/32", "8.8.8.8/32"},
				},
			},
			want: []firewall.CustomFirewallRuleArg{
				{
					Direction:      "out",
					Protocol:       "tcp",
					Port:           "443",
					DestinationIps: []string{"1.1.1.1/32", "8.8.8.8/32"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firewall.ToCustomFirewallRuleArgs(tt.rules)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToCustomFirewallRuleArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
