package config

// NetworkConfig holds the VPC and CIDRs for the cluster.
type NetworkConfig struct {
	// e.g. "eu-central", "us-east", "us-west", "ap-southeast"
	Zone string `json:"zone" validate:"default=eu-central"`

	// Main network CIDR
	CIDR   string `json:"cidr" validate:"default=10.128.0.0/9"`
	Subnet string `json:"subnet" validate:"default=10.128.1.0/24"`

	PodSubnets string `json:"pod_subnets" validate:"default=172.20.0.0/16"`

	// DNS domain for the cluster, defaults to "cluster.local" if not provided
	DNSDomain *string `json:"dns_domain"`

	// Service subnet for the cluster, defaults to "10.96.0.0/12" if not provided
	ServiceSubnet *string `json:"service_subnet"`

	// Custom nameservers for the cluster nodes
	// If not provided, defaults to Quad9 and Google Public DNS
	// Example: ["9.9.9.9", "2620:fe::fe", "8.8.8.8", "2001:4860:4860::8888"]
	Nameservers []string `json:"nameservers"`
}
