package cluster

import (
	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/network"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/k8s/charts/autoscaler"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/k8s/charts/ccm"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/k8s/charts/csi"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/k8s/charts/kubeletservingcertapprover"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/k8s/charts/longhorn"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/k8s/charts/metricsserver"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/core"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/cluster"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ApplicationsArgs struct {
	Cfg                         *config.PulumiConfig
	Kubeconfig                  *cluster.Kubeconfig
	Network                     *network.Network
	Images                      *image.Images
	MachineConfigurationManager *core.MachineConfigurationManager
	FirewallWorker              *hcloud.Firewall
}

type Applications struct {
	Provider                   *kubernetes.Provider
	HcloudSecret               *corev1.Secret
	Longhorn                   *longhorn.Longhorn
	CloudControlManager        *ccm.CloudControlManager
	CSI                        *csi.CSI
	ClusterAutoscaler          *autoscaler.ClusterAutoscaler
	AutoscalerConfiguration    *autoscaler.AutoscalerConfiguration
	KubeletServingCertApprover *kubeletservingcertapprover.KubeletServingCertApprover
	MetricServer               *metricsserver.MetricServer
}

func NewApplications(ctx *pulumi.Context, name string, args *ApplicationsArgs, opts ...pulumi.ResourceOption) (*Applications, error) { //nolint:cyclop // TODO: refactor
	out := &Applications{}
	var err error

	out.Provider, err = kubernetes.NewProvider(ctx, "k8s", &kubernetes.ProviderArgs{
		Kubeconfig:        args.Kubeconfig.KubeconfigRaw,
		ClusterIdentifier: pulumi.StringPtr("hcloud-talos-k8s"),
	}, pulumi.Parent(args.Kubeconfig))
	if err != nil {
		return nil, err
	}

	// default k8s pulumi opts, define kubernetes provider & parent
	opts = append(opts,
		pulumi.Provider(out.Provider),
		pulumi.Parent(out.Provider),
		pulumi.DependsOn([]pulumi.Resource{
			args.Kubeconfig,
			out.Provider,
		}),
	)

	if args.Cfg.Kubernetes.HCloudToken != "" {
		out.HcloudSecret, err = corev1.NewSecret(ctx, "hcloud-secret", &corev1.SecretArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("hcloud"),
				Namespace: pulumi.String("kube-system"),
			},
			StringData: pulumi.StringMap{
				"token":   pulumi.String(args.Cfg.Kubernetes.HCloudToken),
				"network": args.Network.Network.ID(),
			},
		},
			opts...,
		)
		if err != nil {
			return nil, err
		}
		opts = append(opts, pulumi.DependsOn([]pulumi.Resource{out.HcloudSecret}))
	}

	if args.Cfg.Kubernetes.HetznerCCM != nil && args.Cfg.Kubernetes.HetznerCCM.Enabled {
		out.CloudControlManager, err = ccm.NewCloudControlManager(ctx, &ccm.CloudControlManagerArgs{
			Network:    args.Network,
			Values:     args.Cfg.Kubernetes.HetznerCCM.Values,
			Version:    args.Cfg.Kubernetes.HetznerCCM.Version,
			PodSubnets: args.Cfg.Network.PodSubnets,
		},
			opts...,
		)
		if err != nil {
			return nil, err
		}
	}

	if args.Cfg.Kubernetes.CSI != nil && args.Cfg.Kubernetes.CSI.Enabled {
		out.CSI, err = csi.NewCSI(ctx, &csi.CSIArgs{
			Values:                args.Cfg.Kubernetes.CSI.Values,
			Version:               args.Cfg.Kubernetes.CSI.Version,
			EncryptedSecret:       args.Cfg.Kubernetes.CSI.EncryptedSecret,
			IsDefaultStorageClass: args.Cfg.Kubernetes.CSI.IsDefaultStorageClass,
			ReclaimPolicy:         args.Cfg.Kubernetes.CSI.ReclaimPolicy,
		},
			opts...,
		)
		if err != nil {
			return nil, err
		}
	}

	// Deploy autoscaler configuration if any node pool has autoscaling enabled
	// This configuration is needed even if the Helm chart is not deployed
	if err := deployAutoscaler(ctx, out, args, opts); err != nil {
		return nil, err
	}

	if args.Cfg.Kubernetes.KubeletServingCertApprover != nil && args.Cfg.Kubernetes.KubeletServingCertApprover.Enabled {
		out.KubeletServingCertApprover, err = kubeletservingcertapprover.New(ctx, &kubeletservingcertapprover.Args{
			Version: args.Cfg.Kubernetes.KubeletServingCertApprover.Version,
		},
			opts...,
		)
		if err != nil {
			return nil, err
		}
	}

	if args.Cfg.Kubernetes.KubernetesMetricsServer != nil && args.Cfg.Kubernetes.KubernetesMetricsServer.Enabled {
		out.MetricServer, err = metricsserver.New(ctx, &metricsserver.Args{
			Values:  args.Cfg.Kubernetes.KubernetesMetricsServer.Values,
			Version: args.Cfg.Kubernetes.KubernetesMetricsServer.Version,
		},
			opts...,
		)
		if err != nil {
			return nil, err
		}
	}

	if args.Cfg.Kubernetes.Longhorn != nil && args.Cfg.Kubernetes.Longhorn.Enabled {
		out.Longhorn, err = longhorn.NewLonghorn(ctx, &longhorn.LonghornArgs{
			Values:  args.Cfg.Kubernetes.Longhorn.Values,
			Version: args.Cfg.Kubernetes.Longhorn.Version,
		},
			opts...,
		)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// deployAutoscaler handles deployment of cluster autoscaler and its configuration.
// If autoscaling is configured (or ForceDeployAutoScalerConfig is set), it deploys
// either the full Helm chart (if enabled) or just the configuration (if chart is disabled).
func deployAutoscaler(ctx *pulumi.Context, out *Applications, args *ApplicationsArgs, opts []pulumi.ResourceOption) error {
	// Check if any node pool has autoscaling configured
	hasAutoScaling := false
	for _, pool := range args.Cfg.NodePools.NodePools {
		if pool.AutoScaler != nil {
			hasAutoScaling = true
			break
		}
	}

	// Skip if no autoscaling is configured and force deploy is not set
	if !hasAutoScaling && !args.Cfg.NodePools.ForceDeployAutoScalerConfig {
		return nil
	}

	var err error
	autoscalerArgs := &autoscaler.AutoscalerConfigurationArgs{
		Images:                      args.Images,
		MachineConfigurationManager: args.MachineConfigurationManager,
		NodePools:                   args.Cfg.NodePools.NodePools,
		Subnet:                      args.Cfg.Network.Subnet,
		PodSubnets:                  args.Cfg.Network.PodSubnets,
		DNSDomain:                   args.Cfg.Network.DNSDomain,
		EnableLonghorn:              args.Cfg.Talos.EnableLonghorn,
		LocalStorageFolders:         args.Cfg.Talos.LocalStorageFolders,
		Network:                     args.Network,
		Nameservers:                 args.Cfg.Network.Nameservers,
		HcloudToken:                 args.Cfg.Kubernetes.HCloudToken,
		Firewall:                    args.FirewallWorker,
		EnableKubeSpan:              args.Cfg.Talos.EnableKubeSpan,
		CNI:                         args.Cfg.Talos.CNI,
	}

	if args.Cfg.Kubernetes.ClusterAutoScaler != nil && args.Cfg.Kubernetes.ClusterAutoScaler.Enabled {
		out.ClusterAutoscaler, err = autoscaler.NewClusterAutoscaler(ctx, &autoscaler.ClusterAutoscalerArgs{
			Values:                      args.Cfg.Kubernetes.ClusterAutoScaler.Values,
			Version:                     args.Cfg.Kubernetes.ClusterAutoScaler.Version,
			Images:                      autoscalerArgs.Images,
			MachineConfigurationManager: autoscalerArgs.MachineConfigurationManager,
			NodePools:                   autoscalerArgs.NodePools,
			Subnet:                      autoscalerArgs.Subnet,
			PodSubnets:                  autoscalerArgs.PodSubnets,
			DNSDomain:                   args.Cfg.Network.DNSDomain,
			ServiceSubnet:               args.Cfg.Network.ServiceSubnet,
			EnableLonghorn:              autoscalerArgs.EnableLonghorn,
			LocalStorageFolders:         autoscalerArgs.LocalStorageFolders,
			Network:                     autoscalerArgs.Network,
			Nameservers:                 autoscalerArgs.Nameservers,
			HcloudToken:                 autoscalerArgs.HcloudToken,
			Firewall:                    autoscalerArgs.Firewall,
			EnableKubeSpan:              autoscalerArgs.EnableKubeSpan,
			CNI:                         autoscalerArgs.CNI,
		},
			opts...,
		)
		return err
	}

	// If the Helm chart is not enabled but autoscaling is configured,
	// deploy only the configuration (secrets and node configs)
	out.AutoscalerConfiguration, err = autoscaler.DeployAutoscalerConfiguration(ctx, autoscalerArgs, opts...)
	return err
}
