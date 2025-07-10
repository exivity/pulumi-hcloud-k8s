package compute

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/network"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/cli"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/core"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
)

var (
	// talosUpgradeQueue is a queue of nodes to be upgraded
	// this is used to upgrade the nodes in the node pool
	// in a controlled manner
	// the nodes are upgraded one by one
	talosUpgradeQueue = []pulumi.Resource{}

	// ErrAutoScalerNotSupportedForControlPlane indicates that auto-scaler nodes are not supported for control plane node pools
	ErrAutoScalerNotSupportedForControlPlane = errors.New("auto-scaler nodes are not supported for control plane node pools")
)

type NodePoolArgs struct {
	// Count is the number of nodes in the pool
	Count int
	// ServerSize is the server type to use for the nodes
	ServerSize string
	// Images are the images to use for the nodes
	Images *image.Images
	// Arch is the architecture of the nodes
	Arch image.CPUArchitecture
	// Region is the region of the nodes
	Region string
	// NodePoolName is the name of the node pool
	NodePoolName *string
	// ServerNodeType is the type of server node
	ServerNodeType meta.ServerNodeType
	// EnableBackup is whether to enable backups for the nodes
	EnableBackup bool
	// MachineConfigurationManager generates the machine configuration for the nodes
	MachineConfigurationManager *core.MachineConfigurationManager
	// ConfigPatchesBootstrap are the talos config patches to use for the nodes to apply for creation
	ConfigPatchesBootstrap pulumi.StringArrayInput
	// ConfigPatches are the talos config patches to use for the nodes to apply after creation
	ConfigPatches pulumi.StringArrayInput
	// PlacementGroup is the placement group to use for the nodes
	// this is optional and can be nil
	PlacementGroup *hcloud.PlacementGroup
	// Network is the network to use for the nodes
	Network *network.Network
	// Firewall is the firewall to use for the nodes
	Firewall *hcloud.Firewall
}

type NodePool struct {
	// NodePoolName is the name of the node pool
	NodePoolName string
	// ServerNodeType is the type of server node
	ServerNodeType meta.ServerNodeType
	// MachineConfigurationManager generates the machine configuration for the nodes
	MachineConfigurationManager *core.MachineConfigurationManager
	// ConfigPatches are the talos config patches to use for the nodes
	ConfigPatches pulumi.StringArrayInput
	// Nodes of the node pool
	Nodes []*hcloud.Server
	// AutoScalerNodes are the nodes in the node pool that are part of the auto-scaler
	AutoScalerNodes []hcloud.GetServersServer
}

// NewNodePool creates a new node pool in Hetzner Cloud.
// A node pool can be a control plane or a worker pool.
func NewNodePool(ctx *pulumi.Context, name string, args *NodePoolArgs, opts ...pulumi.ResourceOption) (*NodePool, error) {
	img, err := args.Images.GetImageByArch(args.Arch)
	if err != nil {
		return nil, err
	}

	var pg pulumi.IntPtrInput
	if args.PlacementGroup != nil {
		pg = args.PlacementGroup.ID().ApplyT(func(id pulumi.ID) *int {
			intID, _ := strconv.Atoi(string(id))
			return &intID
		}).(pulumi.IntPtrOutput)
	}

	machineConfiguration := args.MachineConfigurationManager.NewMachineConfiguration(ctx, &core.MachineConfigurationArgs{
		ServerNodeType: args.ServerNodeType,
		ConfigPatches:  args.ConfigPatchesBootstrap,
	})

	nodes := make([]*hcloud.Server, args.Count)

	for i := 0; i < args.Count; i++ {
		nodeName := fmt.Sprintf("%s-%d", name, i)
		nodeName = strings.ToLower(nodeName) // Hetzner CCM requires nodes to have lowercase names

		server, err := hcloud.NewServer(ctx, nodeName, &hcloud.ServerArgs{
			Name:       pulumi.String(nodeName),
			Image:      img.GetBuildID(),
			ServerType: pulumi.String(args.ServerSize),
			Location:   pulumi.String(args.Region),
			Backups:    pulumi.Bool(args.EnableBackup),
			Labels: meta.NewLabels(ctx, &meta.ServerLabelsArgs{
				ServerNodeType: args.ServerNodeType,
				Region:         &args.Region,
				Arch:           &args.Arch,
				NodePoolName:   args.NodePoolName,
			}),
			PublicNets: hcloud.ServerPublicNetArray{
				&hcloud.ServerPublicNetArgs{
					Ipv4Enabled: pulumi.Bool(true),
					Ipv6Enabled: pulumi.Bool(true),
				},
			},
			UserData:               machineConfiguration,
			ShutdownBeforeDeletion: pulumi.BoolPtr(true),
			PlacementGroupId:       pg,
			FirewallIds: pulumi.IntArray{
				args.Firewall.ID().ApplyT(func(id pulumi.ID) int {
					idInt, _ := strconv.Atoi(string(id))
					return idInt
				}).(pulumi.IntOutput),
			},
		}, append(opts,
			pulumi.AdditionalSecretOutputs([]string{"userData"}),
			pulumi.IgnoreChanges([]string{"userData", "image"}),
		)...)
		if err != nil {
			return nil, err
		}

		// attach the server to the network
		_, err = hcloud.NewServerNetwork(ctx, nodeName, &hcloud.ServerNetworkArgs{
			ServerId: server.ID().ApplyT(func(id pulumi.ID) int {
				idInt, _ := strconv.Atoi(string(id))
				return idInt
			}).(pulumi.IntOutput),
			NetworkId: args.Network.Network.ID().ApplyT(func(id pulumi.ID) int {
				idInt, _ := strconv.Atoi(string(id))
				return idInt
			}).(pulumi.IntOutput),
		}, append(opts,
			pulumi.Parent(server),
		)...)
		if err != nil {
			return nil, err
		}

		nodes[i] = server
	}

	if args.NodePoolName == nil {
		nodePoolName := fmt.Sprintf("%s-%s-%s", args.ServerNodeType, args.Region, args.ServerSize)
		args.NodePoolName = &nodePoolName
	}

	return &NodePool{
		NodePoolName:                *args.NodePoolName,
		ServerNodeType:              args.ServerNodeType,
		MachineConfigurationManager: args.MachineConfigurationManager,
		ConfigPatches:               args.ConfigPatches,
		Nodes:                       nodes,
	}, nil
}

// FindNodePoolAutoScalerNodes finds the nodes in the node pool that are part of the auto-scaler.
// This fetch the Hetzner API to add the nodes which are created by the auto-scaler.
func (n *NodePool) FindNodePoolAutoScalerNodes(ctx *pulumi.Context, opts ...pulumi.InvokeOption) error {
	if n.ServerNodeType == meta.ControlPlaneNode {
		return fmt.Errorf("node pool %s: %w", n.NodePoolName, ErrAutoScalerNotSupportedForControlPlane)
	}

	selector := fmt.Sprintf("%s=%s,!project", meta.NodePoolLabel, n.NodePoolName)

	nodes, err := hcloud.GetServers(ctx, &hcloud.GetServersArgs{
		WithSelector: pulumi.StringRef(selector),
	}, opts...)
	if err != nil {
		return err
	}

	n.AutoScalerNodes = nodes.Servers

	return nil
}

// ApplyConfigPatches applies the config patches to the nodes in the node pool.
func (n *NodePool) ApplyConfigPatches(ctx *pulumi.Context, opts ...pulumi.ResourceOption) error {
	machineConfiguration := n.MachineConfigurationManager.NewMachineConfiguration(ctx, &core.MachineConfigurationArgs{
		ServerNodeType: n.ServerNodeType,
		ConfigPatches:  n.ConfigPatches,
	})

	for i, node := range n.Nodes {
		_, err := machine.NewConfigurationApply(ctx, fmt.Sprintf("%s-%d", n.NodePoolName, i), &machine.ConfigurationApplyArgs{
			ClientConfiguration:       n.MachineConfigurationManager.Secrets.ClientConfiguration,
			MachineConfigurationInput: machineConfiguration,
			Node:                      node.Ipv4Address,
			ConfigPatches:             n.ConfigPatches,
		}, pulumi.DependsOn([]pulumi.Resource{node}), pulumi.Parent(node))
		if err != nil {
			return err
		}
	}

	for _, node := range n.AutoScalerNodes {
		_, err := machine.NewConfigurationApply(ctx, node.Name, &machine.ConfigurationApplyArgs{
			ClientConfiguration:       n.MachineConfigurationManager.Secrets.ClientConfiguration,
			MachineConfigurationInput: machineConfiguration,
			Node:                      pulumi.String(node.Ipv4Address),
			ConfigPatches:             n.ConfigPatches,
		}, opts...)
		if err != nil {
			return err
		}
	}
	return nil
}

type UpgradeTalosArgs struct {
	// Talosconfig is the talos configuration to use for the Talos CLI
	Talosconfig pulumi.StringOutput
	// TalosVersion is the version of Talos to upgrade to
	TalosVersion string
	// Images are the images to use for the upgrade
	Images *image.Images
}

func (n *NodePool) UpgradeTalos(ctx *pulumi.Context, args *UpgradeTalosArgs) error {
	for i, node := range n.Nodes {
		upgradeTalos, err := cli.UpgradeTalos(ctx, fmt.Sprintf("%s-%d", n.NodePoolName, i), &cli.UpgradeTalosArgs{
			Talosconfig:     args.Talosconfig,
			TalosVersion:    args.TalosVersion,
			Images:          args.Images,
			NodeIpv4Address: node.Ipv4Address,
			NodeImage:       node.Image,
		}, pulumi.Parent(node), pulumi.DependsOn(talosUpgradeQueue))
		if err != nil {
			return err
		}
		talosUpgradeQueue = append(talosUpgradeQueue, upgradeTalos)
	}

	for _, node := range n.AutoScalerNodes {
		upgradeTalos, err := cli.UpgradeTalos(ctx, node.Name, &cli.UpgradeTalosArgs{
			Talosconfig:     args.Talosconfig,
			TalosVersion:    args.TalosVersion,
			Images:          args.Images,
			NodeIpv4Address: pulumi.String(node.Ipv4Address).ToStringOutput(),
			NodeImage:       pulumi.StringPtr(node.Image).ToStringPtrOutput(),
		}, pulumi.DependsOn(talosUpgradeQueue))
		if err != nil {
			return err
		}
		talosUpgradeQueue = append(talosUpgradeQueue, upgradeTalos)
	}

	return nil
}
