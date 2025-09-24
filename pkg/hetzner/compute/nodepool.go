package compute

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
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
			Image:      pulumi.Sprintf("%d", img.ImageId()),
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
func (n *NodePool) ApplyConfigPatches(ctx *pulumi.Context, opts ...pulumi.ResourceOption) ([]*machine.ConfigurationApply, error) {
	machineConfiguration := n.MachineConfigurationManager.NewMachineConfiguration(ctx, &core.MachineConfigurationArgs{
		ServerNodeType: n.ServerNodeType,
		ConfigPatches:  n.ConfigPatches,
	})

	configurationApplies := []*machine.ConfigurationApply{}

	for i, node := range n.Nodes {
		configurationApply, err := machine.NewConfigurationApply(ctx, fmt.Sprintf("%s-%d", n.NodePoolName, i), &machine.ConfigurationApplyArgs{
			ClientConfiguration:       n.MachineConfigurationManager.Secrets.ClientConfiguration,
			MachineConfigurationInput: machineConfiguration,
			Node:                      node.Ipv4Address,
			ConfigPatches:             n.ConfigPatches,
		}, append(opts, pulumi.Parent(node), pulumi.DependsOn(talosUpgradeQueue))...)
		if err != nil {
			return nil, err
		}
		configurationApplies = append(configurationApplies, configurationApply)
	}

	for _, node := range n.AutoScalerNodes {
		configurationApply, err := machine.NewConfigurationApply(ctx, node.Name, &machine.ConfigurationApplyArgs{
			ClientConfiguration:       n.MachineConfigurationManager.Secrets.ClientConfiguration,
			MachineConfigurationInput: machineConfiguration,
			Node:                      pulumi.String(node.Ipv4Address),
			ConfigPatches:             n.ConfigPatches,
		}, opts...)
		if err != nil {
			return nil, err
		}
		configurationApplies = append(configurationApplies, configurationApply)
	}
	return configurationApplies, nil
}

type UpgradeTalosArgs struct {
	// Talosconfig is the talos configuration to use for the Talos CLI
	Talosconfig pulumi.StringOutput
	// TalosVersion is the version of Talos to upgrade to
	TalosVersion string
	// Images are the images to use for the upgrade
	Images *image.Images
}

func (n *NodePool) UpgradeTalos(ctx *pulumi.Context, args *UpgradeTalosArgs, opts ...pulumi.ResourceOption) ([]pulumi.Resource, error) {
	for i, node := range n.Nodes {
		upgradeTalos, err := cli.UpgradeTalos(ctx, fmt.Sprintf("%s-%d", n.NodePoolName, i), &cli.UpgradeTalosArgs{
			Talosconfig:     args.Talosconfig,
			TalosVersion:    args.TalosVersion,
			Images:          args.Images,
			NodeIpv4Address: node.Ipv4Address,
			NodeImage:       node.Image,
		}, append(opts, pulumi.Parent(node), pulumi.DependsOn(talosUpgradeQueue))...)
		if err != nil {
			return nil, err
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
		}, append(opts, pulumi.DependsOn(talosUpgradeQueue))...)
		if err != nil {
			return nil, err
		}
		talosUpgradeQueue = append(talosUpgradeQueue, upgradeTalos)
	}

	return talosUpgradeQueue, nil
}

// DeployControlPlanePools deploys all control plane node pools
func DeployControlPlanePools(ctx *pulumi.Context, cfg *config.PulumiConfig, images *image.Images, net *network.Network, cpPg *hcloud.PlacementGroup, machineConfigurationManager *core.MachineConfigurationManager, firewallCp *hcloud.Firewall, hetznerProvider *hcloud.Provider) ([]*NodePool, error) {
	cpPools := []*NodePool{}

	for _, pool := range cfg.ControlPlane.NodePools {
		cpNodeConfigurationBootstrap, err := core.NewNodeConfiguration(&core.NodeConfigurationArgs{
			ServerNodeType:                 meta.ControlPlaneNode,
			Subnet:                         cfg.Network.Subnet,
			PodSubnets:                     cfg.Network.PodSubnets,
			EnableLonghornSupport:          cfg.Talos.EnableLonghorn,
			EnableLocalStorage:             cfg.Talos.EnableLocalStorage,
			Nameservers:                    cfg.Network.Nameservers,
			SecretboxEncryptionSecret:      cfg.Talos.SecretboxEncryptionSecret,
			AllowSchedulingOnControlPlanes: cfg.Talos.AllowSchedulingOnControlPlanes,
			BootstrapEnable:                true,
			NodeLabels:                     pool.Labels,
			NodeTaints:                     pool.Taints,
			NodeAnnotations:                pool.Annotations,
			Registries:                     cfg.Talos.Registries,
			CertLifetime:                   cfg.Talos.CertLifetime,
			ExtraManifests:                 cfg.Talos.ExtraManifests,
			ExtraManifestHeaders:           cfg.Talos.ExtraManifestHeaders,
			InlineManifests:                cfg.Talos.InlineManifests,
			EnableHetznerCCMExtraManifest:  cfg.Talos.EnableHetznerCCMExtraManifest,
			EnableKubeSpan:                 cfg.Talos.EnableKubeSpan,
			CNI:                            cfg.Talos.CNI,
		})
		if err != nil {
			return nil, err
		}

		cpNodeConfiguration, err := core.NewNodeConfiguration(&core.NodeConfigurationArgs{
			ServerNodeType:                 meta.ControlPlaneNode,
			Subnet:                         cfg.Network.Subnet,
			PodSubnets:                     cfg.Network.PodSubnets,
			EnableLonghornSupport:          cfg.Talos.EnableLonghorn,
			EnableLocalStorage:             cfg.Talos.EnableLocalStorage,
			Nameservers:                    cfg.Network.Nameservers,
			SecretboxEncryptionSecret:      cfg.Talos.SecretboxEncryptionSecret,
			AllowSchedulingOnControlPlanes: cfg.Talos.AllowSchedulingOnControlPlanes,
			NodeLabels:                     pool.Labels,
			NodeTaints:                     pool.Taints,
			NodeAnnotations:                pool.Annotations,
			Registries:                     cfg.Talos.Registries,
			CertLifetime:                   cfg.Talos.CertLifetime,
			ExtraManifests:                 cfg.Talos.ExtraManifests,
			ExtraManifestHeaders:           cfg.Talos.ExtraManifestHeaders,
			InlineManifests:                cfg.Talos.InlineManifests,
			EnableHetznerCCMExtraManifest:  cfg.Talos.EnableHetznerCCMExtraManifest,
			EnableKubeSpan:                 cfg.Talos.EnableKubeSpan,
			CNI:                            cfg.Talos.CNI,
		})
		if err != nil {
			return nil, err
		}

		cpPool, err := NewNodePool(ctx, fmt.Sprintf("controlplane-%s-%s", pool.Region, pool.ServerSize), &NodePoolArgs{
			Count:                       pool.Count,
			ServerSize:                  pool.ServerSize,
			Images:                      images,
			Arch:                        pool.Arch,
			Region:                      pool.Region,
			ServerNodeType:              meta.ControlPlaneNode,
			PlacementGroup:              cpPg,
			Network:                     net,
			EnableBackup:                pool.EnableBackup,
			MachineConfigurationManager: machineConfigurationManager,
			ConfigPatchesBootstrap: pulumi.StringArray{
				pulumi.String(cpNodeConfigurationBootstrap),
			},
			ConfigPatches: pulumi.StringArray{
				pulumi.String(cpNodeConfiguration),
			},
			Firewall: firewallCp,
		},
			pulumi.Parent(cpPg),
			pulumi.Provider(hetznerProvider),
			pulumi.DependsOn([]pulumi.Resource{firewallCp}),
		)
		if err != nil {
			return nil, err
		}

		cpPools = append(cpPools, cpPool)
	}

	return cpPools, nil
}

// DeployWorkerPools deploys all worker node pools
func DeployWorkerPools(ctx *pulumi.Context, cfg *config.PulumiConfig, images *image.Images, net *network.Network, machineConfigurationManager *core.MachineConfigurationManager, firewallWorker *hcloud.Firewall, hetznerProvider *hcloud.Provider) ([]*NodePool, error) {
	workerPools := []*NodePool{}

	for _, pool := range cfg.NodePools.NodePools {
		if pool.Labels == nil {
			pool.Labels = map[string]string{}
		}
		if pool.Annotations == nil {
			pool.Annotations = map[string]string{}
		}

		workerNodeConfigurationBootstrap, err := core.NewNodeConfiguration(&core.NodeConfigurationArgs{
			ServerNodeType:        meta.WorkerNode,
			Subnet:                cfg.Network.Subnet,
			PodSubnets:            cfg.Network.PodSubnets,
			NodeLabels:            pool.Labels,
			NodeTaints:            pool.Taints,
			NodeAnnotations:       pool.Annotations,
			EnableLonghornSupport: cfg.Talos.EnableLonghorn,
			EnableLocalStorage:    cfg.Talos.EnableLocalStorage,
			Nameservers:           cfg.Network.Nameservers,
			BootstrapEnable:       true,
			Registries:            cfg.Talos.Registries,
			EnableKubeSpan:        cfg.Talos.EnableKubeSpan,
			CNI:                   cfg.Talos.CNI,
		})
		if err != nil {
			return nil, err
		}

		workerNodeConfiguration, err := core.NewNodeConfiguration(&core.NodeConfigurationArgs{
			ServerNodeType:        meta.WorkerNode,
			Subnet:                cfg.Network.Subnet,
			PodSubnets:            cfg.Network.PodSubnets,
			NodeLabels:            pool.Labels,
			NodeTaints:            pool.Taints,
			NodeAnnotations:       pool.Annotations,
			EnableLonghornSupport: cfg.Talos.EnableLonghorn,
			EnableLocalStorage:    cfg.Talos.EnableLocalStorage,
			Nameservers:           cfg.Network.Nameservers,
			Registries:            cfg.Talos.Registries,
			EnableKubeSpan:        cfg.Talos.EnableKubeSpan,
			CNI:                   cfg.Talos.CNI,
		})
		if err != nil {
			return nil, err
		}

		workerPool, err := NewNodePool(ctx, pool.Name, &NodePoolArgs{
			Count:                       pool.Count,
			ServerSize:                  pool.ServerSize,
			Images:                      images,
			Arch:                        pool.Arch,
			Region:                      pool.Region,
			NodePoolName:                &pool.Name,
			ServerNodeType:              meta.WorkerNode,
			Network:                     net,
			MachineConfigurationManager: machineConfigurationManager,
			ConfigPatchesBootstrap: pulumi.StringArray{
				pulumi.String(workerNodeConfigurationBootstrap),
			},
			ConfigPatches: pulumi.StringArray{
				pulumi.String(workerNodeConfiguration),
			},
			Firewall: firewallWorker,
		},
			pulumi.Provider(hetznerProvider),
			pulumi.DependsOn([]pulumi.Resource{firewallWorker}),
		)
		if err != nil {
			return nil, err
		}

		// Skip auto-scaler node discovery if configured for all node pools
		if !cfg.NodePools.SkipAutoScalerDiscovery {
			err = workerPool.FindNodePoolAutoScalerNodes(ctx, pulumi.Provider(hetznerProvider))
			if err != nil {
				return nil, err
			}
		}

		workerPools = append(workerPools, workerPool)
	}

	return workerPools, nil
}

// ApplyConfigPatchesToAllPools applies configuration patches to all node pools
func ApplyConfigPatchesToAllPools(ctx *pulumi.Context, cpPools []*NodePool, workerPools []*NodePool, hetznerProvider *hcloud.Provider, opts ...pulumi.ResourceOption) ([]pulumi.Resource, error) {
	configurationApplies := []*machine.ConfigurationApply{}

	// Apply config patches to control plane pools
	for _, cpPool := range cpPools {
		configurationApply, err := cpPool.ApplyConfigPatches(ctx,
			append(opts, pulumi.DependsOn(talosUpgradeQueue))...,
		)
		if err != nil {
			return nil, err
		}
		configurationApplies = append(configurationApplies, configurationApply...)
	}

	// Apply config patches to worker pools
	for _, workerPool := range workerPools {
		configurationApply, err := workerPool.ApplyConfigPatches(ctx,
			append(opts, pulumi.DependsOn(talosUpgradeQueue))...,
		)
		if err != nil {
			return nil, err
		}
		configurationApplies = append(configurationApplies, configurationApply...)
	}

	out := []pulumi.Resource{}
	for _, c := range configurationApplies {
		out = append(out, c)
	}

	return out, nil
}

// UpgradeTalosOnAllPools upgrades Talos on all node pools
func UpgradeTalosOnAllPools(ctx *pulumi.Context, cpPools []*NodePool, workerPools []*NodePool, talosVersion string, images *image.Images, talosConfig pulumi.StringOutput, opts ...pulumi.ResourceOption) ([]pulumi.Resource, error) {
	talosUpgradeQueue := []pulumi.Resource{}

	// Upgrade control plane pools
	for _, cpPool := range cpPools {
		talosUpgradeQueuePool, err := cpPool.UpgradeTalos(ctx, &UpgradeTalosArgs{
			Talosconfig:  talosConfig,
			TalosVersion: talosVersion,
			Images:       images,
		}, append(opts, pulumi.DependsOn(talosUpgradeQueue))...)
		if err != nil {
			return nil, err
		}
		talosUpgradeQueue = append(talosUpgradeQueue, talosUpgradeQueuePool...)
	}

	// Upgrade worker pools
	for _, workerPool := range workerPools {
		talosUpgradeQueuePool, err := workerPool.UpgradeTalos(ctx, &UpgradeTalosArgs{
			Talosconfig:  talosConfig,
			TalosVersion: talosVersion,
			Images:       images,
		}, append(opts, pulumi.DependsOn(talosUpgradeQueue))...)
		if err != nil {
			return nil, err
		}
		talosUpgradeQueue = append(talosUpgradeQueue, talosUpgradeQueuePool...)
	}

	return talosUpgradeQueue, nil
}
