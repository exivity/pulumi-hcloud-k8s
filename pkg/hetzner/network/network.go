package network

import (
	"strconv"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type NetworkArgs struct {
	// NetworkZone is the network zone for the network, like "eu-central"
	NetworkZone string
	// CIDR is the IP range for the network, like 10.0.0.0/8
	CIDR string
	// Subnet is the IP range for the network subnet, like 10.128.1.0/24
	Subnet string
}

type Network struct {
	Network       *hcloud.Network
	NetworkSubnet *hcloud.NetworkSubnet
}

func NewNetwork(ctx *pulumi.Context, name string, args *NetworkArgs, opts ...pulumi.ResourceOption) (*Network, error) {
	network, err := hcloud.NewNetwork(ctx, name, &hcloud.NetworkArgs{
		Name:    pulumi.String(name),
		IpRange: pulumi.String(args.CIDR),
		Labels:  meta.NewLabels(ctx, &meta.ServerLabelsArgs{ServerNodeType: meta.NoneNode}),
	}, opts...)
	if err != nil {
		return nil, err
	}

	networkSubnet, err := hcloud.NewNetworkSubnet(ctx, name, &hcloud.NetworkSubnetArgs{
		NetworkId: network.ID().ApplyT(func(id pulumi.ID) int {
			idInt, _ := strconv.Atoi(string(id))
			return idInt
		}).(pulumi.IntOutput),
		Type:        pulumi.String("cloud"),
		NetworkZone: pulumi.String(args.NetworkZone),
		IpRange:     pulumi.String(args.Subnet),
	}, append(opts,
		pulumi.Parent(network),
	)...)
	if err != nil {
		return nil, err
	}

	return &Network{
		Network:       network,
		NetworkSubnet: networkSubnet,
	}, nil
}
