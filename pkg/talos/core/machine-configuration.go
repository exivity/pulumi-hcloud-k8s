package core

import (
	"fmt"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/lb"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
)

type MachineConfigurationManagerArgs struct {
	// ControlplaneLoadBalancer is the control plane load balancer
	ControlplaneLoadBalancer *lb.Controlplane
	// SingleControlPlaneNodeIP is the IP address of a single control plane node (used when load balancer is disabled)
	SingleControlPlaneNodeIP pulumi.StringInput
	// TalosVersion is the version of Talos to use
	TalosVersion string
	// KubernetesVersion is the version of Kubernetes to use
	KubernetesVersion string
}

type MachineConfigurationManager struct {
	// ClusterName is the name of the cluster
	ClusterName string
	// Talos Linux secrets for the cluster
	Secrets *machine.Secrets
	// ControlplaneLoadBalancer is the control plane load balancer
	ControlplaneLoadBalancer *lb.Controlplane
	// SingleControlPlaneNodeIP is the IP address of a single control plane node (used when load balancer is disabled)
	SingleControlPlaneNodeIP pulumi.StringInput
	// TalosVersion is the version of Talos to use
	TalosVersion string
	// KubernetesVersion is the version of Kubernetes to use
	KubernetesVersion string
}

func NewMachineConfigurationManager(ctx *pulumi.Context, name string, args *MachineConfigurationManagerArgs, opts ...pulumi.ResourceOption) (*MachineConfigurationManager, error) {
	// secrets are the encryption keys for the cluster
	secrets, err := machine.NewSecrets(ctx, fmt.Sprintf("%s-secret", name), &machine.SecretsArgs{}, opts...)
	if err != nil {
		return nil, err
	}

	return &MachineConfigurationManager{
		ClusterName:              name,
		Secrets:                  secrets,
		ControlplaneLoadBalancer: args.ControlplaneLoadBalancer,
		SingleControlPlaneNodeIP: args.SingleControlPlaneNodeIP,
		TalosVersion:             args.TalosVersion,
		KubernetesVersion:        args.KubernetesVersion,
	}, nil
}

type MachineConfigurationArgs struct {
	// ServerNodeType is the type of server node
	ServerNodeType meta.ServerNodeType
	// ConfigPatches is the configuration patches to apply to the machine
	ConfigPatches pulumi.StringArrayInput
}

// NewMachineConfiguration generates a new machine configuration for the cluster
// A MachineConfiguration is needed to give a new hetzner server as UserData
// Like UserData: NewMachineConfiguration()
func (c *MachineConfigurationManager) NewMachineConfiguration(ctx *pulumi.Context, args *MachineConfigurationArgs) (pulumi.StringOutput, error) {
	configuration := machine.GetConfigurationOutputArgs{
		ClusterName:       pulumi.String(c.ClusterName),
		MachineType:       pulumi.String(args.ServerNodeType),
		MachineSecrets:    c.Secrets.MachineSecrets,
		TalosVersion:      pulumi.String(c.TalosVersion),
		KubernetesVersion: pulumi.String(c.KubernetesVersion),
		ConfigPatches:     args.ConfigPatches,
		Docs:              pulumi.BoolPtr(false),
		Examples:          pulumi.BoolPtr(false),
	}

	// If we have a single control plane IP, use that
	if c.SingleControlPlaneNodeIP != nil {
		configuration.ClusterEndpoint = pulumi.Sprintf("https://%s:%d", c.SingleControlPlaneNodeIP, lb.ControlPlaneLoadBalancerPort)
	}

	// If we have a load balancer, prefer that over single node IP
	if c.ControlplaneLoadBalancer != nil {
		configuration.ClusterEndpoint = pulumi.Sprintf("https://%s:%d", c.ControlplaneLoadBalancer.LoadBalancer.Ipv4, lb.ControlPlaneLoadBalancerPort)
	}

	if configuration.ClusterEndpoint == nil {
		return pulumi.StringOutput{}, fmt.Errorf("either ControlplaneLoadBalancer or SingleControlPlaneNodeIP must be set")
	}

	return machine.GetConfigurationOutput(ctx, configuration,
		pulumi.Parent(c.Secrets),
	).MachineConfiguration(), nil
}

// SetSingleControlPlaneNodeIP sets the IP address of the first control plane node.
// This method should only be called when SingleControlPlaneNodeIP is nil (i.e., when load balancer is disabled).
// It allows setting the control plane endpoint after the first control plane node is created.
func (c *MachineConfigurationManager) SetSingleControlPlaneNodeIP(ip pulumi.StringInput) {
	c.SingleControlPlaneNodeIP = ip
}

// HasClusterEndpoint checks if a cluster endpoint is available.
// Returns true if either a control plane load balancer is configured
// or a single control plane node IP is set. This helps determine
// whether the configuration is for a load balancer or non-loadbalancer setup.
func (c *MachineConfigurationManager) HasClusterEndpoint() bool {
	return c.ControlplaneLoadBalancer != nil || c.SingleControlPlaneNodeIP != nil
}
