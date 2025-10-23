package lb

import (
	"fmt"
	"strconv"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/network"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// ControlPlaneLoadBalancerPort is the port the control plane load balancer listens on
	ControlPlaneLoadBalancerPort = 6443
)

// ControlplaneArgs are the arguments for the NewControlplane function
type ControlplaneArgs struct {
	// DisableLoadBalancer disables the creation of the load balancer
	DisableLoadBalancer bool
	// LoadBalancerType is the type of load balancer to create
	LoadBalancerType string
	// Hetzner Cloud network to use for the load balancer
	Network *network.Network
	// Location is the location to create the load balancer in
	Location *string
}

// Controlplane represents a control plane load balancer
//
// The control plane load balancer is a Hetzner Cloud load balancer that exposes the Kubernetes API server (port 6443)
// to clients and worker nodes. It provides a single, stable endpoint for the API, enabling high availability (HA)
// and seamless failover between multiple control plane nodes. In production, this is critical for cluster reliability.
type Controlplane struct {
	// LoadBalancer is the Hetzner Cloud load balancer
	LoadBalancer *hcloud.LoadBalancer
	// Service is the Hetzner Cloud load balancer service
	Service *hcloud.LoadBalancerService
	// Target is the Hetzner Cloud load balancer target
	Target *hcloud.LoadBalancerTarget
	// LoadBalancerNetwork is the Hetzner Cloud load balancer network
	LoadBalancerNetwork *hcloud.LoadBalancerNetwork
}

// NewControlplane creates a new control plane load balancer
func NewControlplane(ctx *pulumi.Context, name string, args *ControlplaneArgs, opts ...pulumi.ResourceOption) (*Controlplane, error) {
	// If load balancer is disabled, return nil
	if args.DisableLoadBalancer {
		return nil, nil
	}

	resourceName := fmt.Sprintf("%s-controlplane", name)

	lbArgs := &hcloud.LoadBalancerArgs{
		Name:             pulumi.String(resourceName),
		LoadBalancerType: pulumi.String(args.LoadBalancerType),
		Labels:           meta.NewLabels(ctx, &meta.ServerLabelsArgs{ServerNodeType: meta.ControlPlaneNode}),
	}
	if args.Location != nil {
		lbArgs.Location = pulumi.StringPtrFromPtr(args.Location)
	} else {
		lbArgs.NetworkZone = args.Network.NetworkSubnet.NetworkZone
	}
	loadBalancer, err := hcloud.NewLoadBalancer(ctx, resourceName, lbArgs, opts...)
	if err != nil {
		return nil, err
	}

	service, err := hcloud.NewLoadBalancerService(ctx, resourceName, &hcloud.LoadBalancerServiceArgs{
		LoadBalancerId:  loadBalancer.ID(),
		Protocol:        pulumi.String("tcp"),
		ListenPort:      pulumi.Int(ControlPlaneLoadBalancerPort),
		DestinationPort: pulumi.Int(ControlPlaneLoadBalancerPort),
	}, append(opts,
		pulumi.Parent(loadBalancer),
		pulumi.DependsOn([]pulumi.Resource{loadBalancer}),
	)...)
	if err != nil {
		return nil, err
	}

	loadBalancerNetwork, err := hcloud.NewLoadBalancerNetwork(ctx, resourceName, &hcloud.LoadBalancerNetworkArgs{
		LoadBalancerId: loadBalancer.ID().ApplyT(func(id pulumi.ID) int {
			idInt, _ := strconv.Atoi(string(id))
			return idInt
		},
		).(pulumi.IntOutput),
		SubnetId: args.Network.NetworkSubnet.ID(),
	}, append(opts,
		pulumi.Parent(loadBalancer),
		pulumi.DependsOn([]pulumi.Resource{loadBalancer}),
	)...)
	if err != nil {
		return nil, err
	}

	target, err := hcloud.NewLoadBalancerTarget(ctx, resourceName, &hcloud.LoadBalancerTargetArgs{
		Type:           pulumi.String("label_selector"),
		LoadBalancerId: loadBalancer.ID().ApplyT(strconv.Atoi).(pulumi.IntOutput),
		LabelSelector:  pulumi.Sprintf("type=controlplane,stack=%s,project=%s", ctx.Stack(), ctx.Project()),
		UsePrivateIp:   pulumi.Bool(true),
	}, append(opts,
		pulumi.Parent(loadBalancer),
		pulumi.DependsOn([]pulumi.Resource{service, loadBalancerNetwork}),
	)...)
	if err != nil {
		return nil, err
	}

	return &Controlplane{
		LoadBalancer:        loadBalancer,
		Service:             service,
		Target:              target,
		LoadBalancerNetwork: loadBalancerNetwork,
	}, nil
}
