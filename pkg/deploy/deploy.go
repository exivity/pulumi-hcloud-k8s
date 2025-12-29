package deploy

import (
	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/compute"
	hfirewall "github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/firewall"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/lb"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/network"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/provider"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/k8s/cluster"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/cli"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/core"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type HetznerTalosKubernetesCluster struct {
	Kubeconfig          *core.Kubeconfig
	TalosConfig         pulumi.StringOutput
	ClusterApplications *cluster.Applications
	ControlPlanePools   []*compute.NodePool
	WorkerPools         []*compute.NodePool
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
		HetznerToken:            cfg.Hetzner.Token,
		EnableARMImageUpload:    enableARMImages,
		EnableX86ImageUpload:    enableX86Images,
		TalosVersion:            cfg.Talos.ImageVersion,
		TalosImageID:            imageID,
		ARMServerSize:           cfg.Talos.GeneratorSizes.ARM,
		X86ServerSize:           cfg.Talos.GeneratorSizes.X86,
		ImageGenerationLocation: cfg.Talos.ImageGenerationLocation,
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
		DisableLoadBalancer: cfg.ControlPlane.DisableLoadBalancer,
		LoadBalancerType:    cfg.ControlPlane.LoadBalancerType,
		Network:             net,
		Location:            cfg.ControlPlane.LoadBalancerLocation,
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

	firewallCp, err := hfirewall.NewControlplaneFirewall(ctx, "fw-controlplane", &hfirewall.ControlplaneFirewallArgs{
		VpnCidrs:                               cfg.Firewall.VpnCidrs,
		OpenAPIToEveryone:                      cfg.Firewall.OpenTalosAPI,
		ExposeKubernetesAPIWithoutLoadBalancer: cfg.ControlPlane.DisableLoadBalancer,
		CustomRules:                            hfirewall.ToCustomFirewallRuleArgs(cfg.Firewall.CustomRulesControlplane),
	}, pulumi.Provider(hetznerProvider))
	if err != nil {
		return nil, err
	}

	firewallWorker, err := hfirewall.NewWorkerFirewall(ctx, "fw-worker", &hfirewall.WorkerFirewallArgs{
		VpnCidrs:          cfg.Firewall.VpnCidrs,
		OpenAPIToEveryone: cfg.Firewall.OpenTalosAPI,
		CustomRules:       hfirewall.ToCustomFirewallRuleArgs(cfg.Firewall.CustomRulesWorker),
	}, pulumi.Provider(hetznerProvider))
	if err != nil {
		return nil, err
	}

	cpPools, err := compute.DeployControlPlanePools(ctx, cfg, images, net, cpPg, machineConfigurationManager, firewallCp, hetznerProvider)
	if err != nil {
		return nil, err
	}
	out.ControlPlanePools = cpPools

	workerPools, err := compute.DeployWorkerPools(ctx, cfg, images, net, machineConfigurationManager, firewallWorker, hetznerProvider)
	if err != nil {
		return nil, err
	}
	out.WorkerPools = workerPools

	workerPoolDependsOn := []pulumi.Resource{}
	for _, cpPool := range cpPools {
		for _, node := range cpPool.Nodes {
			workerPoolDependsOn = append(workerPoolDependsOn, node)
		}
	}

	workerPoolDependsOn = append(workerPoolDependsOn,
		cpPools[0].Nodes[0],
	)

	if cpLb != nil {
		workerPoolDependsOn = append(workerPoolDependsOn,
			cpLb.LoadBalancer,
			cpLb.Service,
			cpLb.Target,
			cpLb.LoadBalancerNetwork,
		)
	}

	// Apply configuration patches to all nodes
	configurationApplies, err := compute.ApplyConfigPatchesToAllPools(ctx, cpPools, workerPools, hetznerProvider)
	if err != nil {
		return nil, err
	}

	// TODO: remove bootstrap creation from k8s package
	out.Kubeconfig, err = core.NewKubeconfig(ctx, &core.KubeconfigArgs{
		CertificateRenewalDuration: cfg.Talos.K8sCertificateRenewalDuration,
		FirstControlPlane:          cpPools[0].Nodes[0],
		Secrets:                    machineConfigurationManager.Secrets,
	},
		pulumi.DependsOn(workerPoolDependsOn),
		pulumi.DependsOn(configurationApplies),
	)
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

	// Upgrade Talos on all nodes
	upgradedNodes, err := compute.UpgradeTalosOnAllPools(ctx, cpPools, workerPools, cfg.Talos.ImageVersion, images, out.TalosConfig,
		pulumi.DependsOn(append(workerPoolDependsOn, out.Kubeconfig.Bootstrap)),
	)
	if err != nil {
		return nil, err
	}

	out.ClusterApplications, err = cluster.NewApplications(ctx, name, &cluster.ApplicationsArgs{
		Cfg:                         cfg,
		Kubeconfig:                  out.Kubeconfig.Kubeconfig,
		Network:                     net,
		Images:                      images,
		MachineConfigurationManager: machineConfigurationManager,
		FirewallWorker:              firewallWorker,
	},
		pulumi.DependsOn(upgradedNodes),
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}
