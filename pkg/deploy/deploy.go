package deploy

import (
	"fmt"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/compute"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/firewall"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/lb"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/network"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/provider"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/cli"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/core"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type HetznerTalosKubernetesCluster struct {
	Kubeconfig          *core.Kubeconfig
	TalosConfig         pulumi.StringOutput
	ClusterApplications *ClusterApplications
}

// NewHetznerTalosKubernetesCluster creates a new Hetzner Talos Kubernetes cluster with the given name and configuration.
// It sets up the necessary Hetzner provider, images, network, control plane load balancer, placement group,
// machine configuration manager, firewalls, control plane and worker node pools, and Kubernetes provider.
// It also applies Talos upgrades and exports the kubeconfig and talosconfig.
func NewHetznerTalosKubernetesCluster(ctx *pulumi.Context, name string, cfg *config.PulumiConfig) (*HetznerTalosKubernetesCluster, error) { //nolint:cyclop,funlen // TODO: refactor
	out := &HetznerTalosKubernetesCluster{}

	hetznerProvider, err := provider.NewProvider(ctx, "hetzner", &provider.ProviderArgs{
		Token: cfg.Hetzner.Token,
	})
	if err != nil {
		return nil, err
	}

	imageID := image.NewTalosImageID(&image.TalosImageIDArgs{
		OverwriteTalosImageID: cfg.Talos.ImageIDOverride,
		EnableLonghornSupport: cfg.Talos.EnableLonghorn,
	})

	// Extract all architectures from both control plane and worker node pools
	var architectures []image.CPUArchitecture //nolint:prealloc
	for _, pool := range cfg.ControlPlane.NodePools {
		architectures = append(architectures, pool.Arch)
	}
	for _, pool := range cfg.NodePools.NodePools {
		architectures = append(architectures, pool.Arch)
	}
	enableARMImages, enableX86Images := image.DetectRequiredArchitecturesFromList(architectures)

	images, err := image.NewImages(ctx, &image.ImagesArgs{
		HetznerToken:         cfg.Hetzner.Token,
		EnableARMImageUpload: enableARMImages,
		EnableX86ImageUpload: enableX86Images,
		TalosVersion:         cfg.Talos.ImageVersion,
		TalosImageID:         imageID,
		ARMServerSize:        cfg.Talos.GeneratorSizes.ARM,
		X86ServerSize:        cfg.Talos.GeneratorSizes.X86,
	}, pulumi.Parent(hetznerProvider))
	if err != nil {
		return nil, err
	}

	net, err := network.NewNetwork(ctx, "talos-network", &network.NetworkArgs{
		NetworkZone: cfg.Network.Zone,
		CIDR:        cfg.Network.CIDR,
		Subnet:      cfg.Network.Subnet,
	}, pulumi.Parent(hetznerProvider), pulumi.Provider(hetznerProvider))
	if err != nil {
		return nil, err
	}

	cpLb, err := lb.NewControlplane(ctx, "controlplane-lb", &lb.ControlplaneArgs{
		LoadBalancerType: cfg.ControlPlane.LoadBalancerType,
		Network:          net,
	}, pulumi.Parent(net.Network), pulumi.Provider(hetznerProvider))
	if err != nil {
		return nil, err
	}

	cpPg, err := compute.NewPlacementGroup(ctx, "controlplane-placement-group", &compute.PlacementGroupArgs{
		ServerNodeType: meta.ControlPlaneNode,
	}, pulumi.Parent(net.Network), pulumi.Provider(hetznerProvider))
	if err != nil {
		return nil, err
	}

	machineConfigurationManager, err := core.NewMachineConfigurationManager(ctx, name, &core.MachineConfigurationManagerArgs{
		ControlplaneLoadBalancer: cpLb,
		TalosVersion:             cfg.Talos.ImageVersion,
		KubernetesVersion:        cfg.Talos.KubernetesVersion,
	})
	if err != nil {
		return nil, err
	}

	firewallCp, err := firewall.NewControlplaneFirewall(ctx, "fw-controlplane", &firewall.ControlplaneFirewallArgs{
		VpnCidrs:          cfg.Firewall.VpnCidrs,
		OpenAPIToEveryone: cfg.Firewall.OpenTalosAPI,
		CustomRules:       toCustomFirewallRuleArgs(cfg.Firewall.CustomRulesControlplane),
	}, pulumi.Provider(hetznerProvider))
	if err != nil {
		return nil, err
	}

	firewallWorker, err := firewall.NewWorkerFirewall(ctx, "fw-worker", &firewall.WorkerFirewallArgs{
		VpnCidrs:          cfg.Firewall.VpnCidrs,
		OpenAPIToEveryone: cfg.Firewall.OpenTalosAPI,
		CustomRules:       toCustomFirewallRuleArgs(cfg.Firewall.CustomRulesWorker),
	}, pulumi.Provider(hetznerProvider))
	if err != nil {
		return nil, err
	}

	cpPools, err := deployControlPlanePools(ctx, cfg, images, net, cpPg, machineConfigurationManager, firewallCp, hetznerProvider)
	if err != nil {
		return nil, err
	}

	// TODO: remove bootstrap creation from k8s package
	out.Kubeconfig, err = core.NewKubeconfig(ctx, &core.KubeconfigArgs{
		CertificateRenewalDuration: cfg.Talos.K8sCertificateRenewalDuration,
		FirstControlPlane:          cpPools[0].Nodes[0],
		Secrets:                    machineConfigurationManager.Secrets,
	},
		pulumi.DependsOn([]pulumi.Resource{
			cpPools[0].Nodes[0],
			cpLb.LoadBalancer,
			cpLb.Service,
			cpLb.Target,
			cpLb.LoadBalancerNetwork,
		}),
	)
	if err != nil {
		return nil, err
	}

	workerPools, err := deployWorkerPools(ctx, cfg, images, net, machineConfigurationManager, firewallWorker, hetznerProvider)
	if err != nil {
		return nil, err
	}

	endpoints := []pulumi.StringOutput{}
	nodes := []pulumi.StringOutput{}
	for _, cpPool := range cpPools {
		for _, node := range cpPool.Nodes {
			endpoints = append(endpoints, node.Ipv4Address)
			nodes = append(nodes, node.Ipv4Address)
		}
	}

	for _, workerPool := range workerPools {
		for _, node := range workerPool.Nodes {
			nodes = append(nodes, node.Ipv4Address)
		}
	}

	out.TalosConfig = cli.NewTalosConfiguration(&cli.TalosConfigurationArgs{
		Context:           machineConfigurationManager.ClusterName,
		Endpoints:         endpoints,
		Nodes:             nodes,
		CACertificate:     out.Kubeconfig.Bootstrap.ClientConfiguration.CaCertificate(),
		ClientCertificate: out.Kubeconfig.Bootstrap.ClientConfiguration.ClientCertificate(),
		ClientKey:         out.Kubeconfig.Bootstrap.ClientConfiguration.ClientKey(),
	})

	// Apply configuration patches to all nodes
	err = applyConfigPatchesToAllPools(ctx, cpPools, workerPools, hetznerProvider)
	if err != nil {
		return nil, err
	}

	// Upgrade Talos on all nodes
	err = upgradeTalosOnAllPools(ctx, cpPools, workerPools, cfg.Talos.ImageVersion, images, out.TalosConfig)
	if err != nil {
		return nil, err
	}

	out.ClusterApplications, err = NewClusterApplications(ctx, name, &ClusterApplicationsArgs{
		Cfg:                         cfg,
		Kubeconfig:                  out.Kubeconfig.Kubeconfig,
		Network:                     net,
		Images:                      images,
		MachineConfigurationManager: machineConfigurationManager,
		FirewallWorker:              firewallWorker,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// toCustomFirewallRuleArgs converts config.FirewallRuleConfig to firewall.CustomFirewallRuleArg
func toCustomFirewallRuleArgs(rules []config.FirewallRuleConfig) []firewall.CustomFirewallRuleArg {
	out := make([]firewall.CustomFirewallRuleArg, 0, len(rules))
	for _, r := range rules {
		out = append(out, firewall.CustomFirewallRuleArg{
			Port:  r.Port,
			CIDRs: r.CIDRs,
		})
	}
	return out
}

// deployControlPlanePools deploys all control plane node pools
func deployControlPlanePools(ctx *pulumi.Context, cfg *config.PulumiConfig, images *image.Images, net *network.Network, cpPg *hcloud.PlacementGroup, machineConfigurationManager *core.MachineConfigurationManager, firewallCp *hcloud.Firewall, hetznerProvider *hcloud.Provider) ([]*compute.NodePool, error) {
	cpPools := []*compute.NodePool{}

	for _, pool := range cfg.ControlPlane.NodePools {
		cpNodeConfigurationBootstrap, err := core.NewNodeConfiguration(&core.NodeConfigurationArgs{
			ServerNodeType:                 meta.ControlPlaneNode,
			Subnet:                         cfg.Network.Subnet,
			PodSubnets:                     cfg.Network.PodSubnets,
			EnableLonghornSupport:          cfg.Talos.EnableLonghorn,
			EnableLocalStorage:             cfg.Talos.EnableLocalStorage,
			SecretboxEncryptionSecret:      cfg.Talos.SecretboxEncryptionSecret,
			AllowSchedulingOnControlPlanes: cfg.Talos.AllowSchedulingOnControlPlanes,
			BootstrapEnable:                true,
			NodeLabels:                     pool.Labels,
			NodeTaints:                     pool.Taints,
			NodeAnnotations:                pool.Annotations,
			Registries:                     cfg.Talos.Registries,
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
			SecretboxEncryptionSecret:      cfg.Talos.SecretboxEncryptionSecret,
			AllowSchedulingOnControlPlanes: cfg.Talos.AllowSchedulingOnControlPlanes,
			NodeLabels:                     pool.Labels,
			NodeTaints:                     pool.Taints,
			NodeAnnotations:                pool.Annotations,
			Registries:                     cfg.Talos.Registries,
		})
		if err != nil {
			return nil, err
		}

		cpPool, err := compute.NewNodePool(ctx, fmt.Sprintf("controlplane-%s-%s", pool.Region, pool.ServerSize), &compute.NodePoolArgs{
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

// deployWorkerPools deploys all worker node pools
func deployWorkerPools(ctx *pulumi.Context, cfg *config.PulumiConfig, images *image.Images, net *network.Network, machineConfigurationManager *core.MachineConfigurationManager, firewallWorker *hcloud.Firewall, hetznerProvider *hcloud.Provider) ([]*compute.NodePool, error) {
	workerPools := []*compute.NodePool{}

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
			BootstrapEnable:       true,
			Registries:            cfg.Talos.Registries,
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
			Registries:            cfg.Talos.Registries,
		})
		if err != nil {
			return nil, err
		}

		workerPool, err := compute.NewNodePool(ctx, pool.Name, &compute.NodePoolArgs{
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

		err = workerPool.FindNodePoolAutoScalerNodes(ctx, pulumi.Provider(hetznerProvider))
		if err != nil {
			return nil, err
		}

		workerPools = append(workerPools, workerPool)
	}

	return workerPools, nil
}

// applyConfigPatchesToAllPools applies configuration patches to all node pools
func applyConfigPatchesToAllPools(ctx *pulumi.Context, cpPools []*compute.NodePool, workerPools []*compute.NodePool, hetznerProvider *hcloud.Provider) error {
	// Apply config patches to control plane pools
	for _, cpPool := range cpPools {
		err := cpPool.ApplyConfigPatches(ctx, pulumi.Provider(hetznerProvider))
		if err != nil {
			return err
		}
	}

	// Apply config patches to worker pools
	for _, workerPool := range workerPools {
		err := workerPool.ApplyConfigPatches(ctx, pulumi.Provider(hetznerProvider))
		if err != nil {
			return err
		}
	}

	return nil
}

// upgradeTalosOnAllPools upgrades Talos on all node pools
func upgradeTalosOnAllPools(ctx *pulumi.Context, cpPools []*compute.NodePool, workerPools []*compute.NodePool, talosVersion string, images *image.Images, talosConfig pulumi.StringOutput) error {
	// Upgrade control plane pools
	for _, cpPool := range cpPools {
		err := cpPool.UpgradeTalos(ctx, &compute.UpgradeTalosArgs{
			Talosconfig:  talosConfig,
			TalosVersion: talosVersion,
			Images:       images,
		})
		if err != nil {
			return err
		}
	}

	// Upgrade worker pools
	for _, workerPool := range workerPools {
		err := workerPool.UpgradeTalos(ctx, &compute.UpgradeTalosArgs{
			Talosconfig:  talosConfig,
			TalosVersion: talosVersion,
			Images:       images,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
