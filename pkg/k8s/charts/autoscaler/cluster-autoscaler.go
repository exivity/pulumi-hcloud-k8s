package autoscaler

import (
	"dario.cat/mergo"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/network"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/core"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ClusterAutoscalerArgs struct {
	// Values are the values to use for the chart
	Values *map[string]interface{}
	// Version is the version of the chart to use
	// The version must be available in the chart repository.
	// If not set, the latest version will be used.
	Version *string `json:"version"`
	// Images are the images to use for the nodes
	Images *image.Images

	MachineConfigurationManager *core.MachineConfigurationManager
	NodePools                   []config.NodePoolConfig
	Subnet                      string
	PodSubnets                  string
	DNSDomain                   *string
	ServiceSubnet               *string
	EnableLonghorn              bool
	LocalStorageFolders         []string
	// Registries is the registries configuration for the Talos image
	Registries  *config.RegistriesConfig
	Network     *network.Network
	Nameservers []string
	HcloudToken string
	// Firewall is the firewall to use for the nodes
	Firewall       *hcloud.Firewall
	EnableKubeSpan bool
	// CNI is the CNI configuration for the cluster.
	CNI *config.CNIConfig
}

type ClusterAutoscaler struct {
	Chart                       *helmv4.Chart
	AutoscalerSecret            *corev1.Secret
	AutoscalerClusterConfigHash pulumi.StringOutput
	AutoscalerClusterConfig     *corev1.Secret
	AutoscalingGroupsConfig     pulumi.Array
}

// AutoscalerConfigurationArgs contains the arguments needed to deploy autoscaler configuration
type AutoscalerConfigurationArgs struct {
	Images                      *image.Images
	MachineConfigurationManager *core.MachineConfigurationManager
	NodePools                   []config.NodePoolConfig
	Subnet                      string
	PodSubnets                  string
	DNSDomain                   *string
	ServiceSubnet               *string
	EnableLonghorn              bool
	LocalStorageFolders         []string
	Registries                  *config.RegistriesConfig
	Network                     *network.Network
	Nameservers                 []string
	HcloudToken                 string
	Firewall                    *hcloud.Firewall
	EnableKubeSpan              bool
	CNI                         *config.CNIConfig
}

// AutoscalerConfiguration holds the deployed autoscaler configuration resources
type AutoscalerConfiguration struct {
	AutoscalerSecret        *corev1.Secret
	AutoscalerClusterConfig *corev1.Secret
	ClusterConfigJSON       pulumi.StringOutput
	ClusterConfigJSONHash   pulumi.StringOutput
	AutoscalingGroups       pulumi.Array
}

// DeployAutoscalerConfiguration deploys the autoscaler configuration (secrets and node configs)
// This can be used independently of whether the Helm chart is deployed or not
func DeployAutoscalerConfiguration(ctx *pulumi.Context, args *AutoscalerConfigurationArgs, opts ...pulumi.ResourceOption) (*AutoscalerConfiguration, error) { //nolint:cyclop,funlen
	imgARM, err := args.Images.GetImageByArch(image.ArchARM)
	if err != nil {
		return nil, err
	}
	imgX86, err := args.Images.GetImageByArch(image.ArchX86)
	if err != nil {
		return nil, err
	}

	nodeConfigs := map[string]HCloudNodeConfig{}
	autoscalingGroups := pulumi.Array{}
	for _, pool := range args.NodePools {
		workerNodeConfiguration, err := core.NewNodeConfiguration(&core.NodeConfigurationArgs{
			ServerNodeType:        meta.WorkerNode,
			Subnet:                args.Subnet,
			PodSubnets:            args.PodSubnets,
			DNSDomain:             args.DNSDomain,
			ServiceSubnet:         args.ServiceSubnet,
			NodeLabels:            pool.Labels,
			NodeTaints:            pool.Taints,
			NodeAnnotations:       pool.Annotations,
			EnableLonghornSupport: args.EnableLonghorn,
			LocalStorageFolders:   args.LocalStorageFolders,
			Registries:            args.Registries,
			EnableKubeSpan:        args.EnableKubeSpan,
			Nameservers:           args.Nameservers,
			CNI:                   args.CNI,
		})
		if err != nil {
			return nil, err
		}

		workerMachineConfiguration, err := args.MachineConfigurationManager.NewMachineConfiguration(ctx, &core.MachineConfigurationArgs{
			ServerNodeType: meta.WorkerNode,
			ConfigPatches: pulumi.StringArray{
				pulumi.String(workerNodeConfiguration),
			},
		})
		if err != nil {
			return nil, err
		}

		nodeConfig := HCloudNodeConfig{
			CloudInit: workerMachineConfiguration,
			Labels:    map[string]string{},
			Taints:    []Taint{},
		}

		for key, value := range pool.Labels {
			nodeConfig.Labels[key] = value
		}
		for key, value := range pool.Annotations {
			nodeConfig.Labels[key] = value
		}
		for _, taint := range pool.Taints {
			nodeConfig.Taints = append(nodeConfig.Taints, Taint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: taint.Effect,
			})
		}

		nodeConfigs[pool.Name] = nodeConfig

		if pool.AutoScaler == nil {
			continue
		}

		autoscalingGroups = append(autoscalingGroups, pulumi.Map{
			"name":         pulumi.String(pool.Name),
			"minSize":      pulumi.Int(pool.AutoScaler.MinCount),
			"maxSize":      pulumi.Int(pool.AutoScaler.MaxCount),
			"instanceType": pulumi.String(pool.ServerSize),
			"region":       pulumi.String(pool.Region),
		})
	}

	clusterConfig := HCloudClusterConfig{
		ImagesForArch: ImagesForArch{
			ARM64: imgARM.ImageId(),
			AMD64: imgX86.ImageId(),
		},
		NodeConfigs: nodeConfigs,
	}
	clusterConfigJSON := clusterConfig.ToJSON()
	clusterConfigJSONHash := hashJSON(clusterConfigJSON)

	autoscalerSecret, err := corev1.NewSecret(ctx, "hcloud-autoscaler", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("hcloud-autoscaler"),
			Namespace: pulumi.String("kube-system"),
		},
		StringData: pulumi.StringMap{
			"HCLOUD_TOKEN":    pulumi.String(args.HcloudToken),
			"HCLOUD_NETWORK":  args.Network.Network.ID(),
			"HCLOUD_FIREWALL": args.Firewall.ID(),
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	autoscalerClusterConfig, err := corev1.NewSecret(ctx, "hcloud-autoscaler-cluster-config", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("hcloud-autoscaler-cluster-config"),
			Namespace: pulumi.String("kube-system"),
		},
		StringData: pulumi.StringMap{
			"HCLOUD_CLUSTER_CONFIG": clusterConfigJSON,
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &AutoscalerConfiguration{
		AutoscalerSecret:        autoscalerSecret,
		AutoscalerClusterConfig: autoscalerClusterConfig,
		ClusterConfigJSON:       clusterConfigJSON,
		ClusterConfigJSONHash:   clusterConfigJSONHash,
		AutoscalingGroups:       autoscalingGroups,
	}, nil
}

func NewClusterAutoscaler(ctx *pulumi.Context, args *ClusterAutoscalerArgs, opts ...pulumi.ResourceOption) (*ClusterAutoscaler, error) {
	// Deploy autoscaler configuration (secrets and node configs)
	autoscalerConfig, err := DeployAutoscalerConfiguration(ctx, &AutoscalerConfigurationArgs{
		Images:                      args.Images,
		MachineConfigurationManager: args.MachineConfigurationManager,
		NodePools:                   args.NodePools,
		Subnet:                      args.Subnet,
		PodSubnets:                  args.PodSubnets,
		DNSDomain:                   args.DNSDomain,
		ServiceSubnet:               args.ServiceSubnet,
		EnableLonghorn:              args.EnableLonghorn,
		LocalStorageFolders:         args.LocalStorageFolders,
		Registries:                  args.Registries,
		Network:                     args.Network,
		Nameservers:                 args.Nameservers,
		HcloudToken:                 args.HcloudToken,
		Firewall:                    args.Firewall,
		EnableKubeSpan:              args.EnableKubeSpan,
		CNI:                         args.CNI,
	}, opts...)
	if err != nil {
		return nil, err
	}

	preDefineValues := pulumi.Map{
		"cloudProvider": pulumi.String("hetzner"),
		"envFromSecret": pulumi.String("hcloud-autoscaler"),
		"extraArgs": pulumi.Map{
			"cloud-provider": pulumi.String("hetzner"),
			// Specifies the utilization threshold below which nodes are considered for scale-down.
			// For example, a value of 0.5 means nodes with less than 50% utilization can be scaled down.
			"scale-down-utilization-threshold": pulumi.String("0.5"),
			// Enables the scale-down feature, allowing the autoscaler to remove unused nodes
			"scale-down-enabled": pulumi.Bool(true),
			// do not scale nodes down if they use local storage
			"skip-nodes-with-local-storage": pulumi.Bool(true),
		},
		"autoscalingGroups": autoscalerConfig.AutoscalingGroups,
		"autoDiscovery": pulumi.Map{
			"enabled": pulumi.Bool(false),
		},
		"extraEnv": pulumi.StringMap{
			"HCLOUD_CLUSTER_CONFIG_FILE": pulumi.String("/etc/kubernetes/hcloud_cluster_config/HCLOUD_CLUSTER_CONFIG"),
			"HCLOUD_CLUSTER_CONFIG_HASH": autoscalerConfig.ClusterConfigJSONHash, // Hash of the cluster config to detect changes
		},
		"extraVolumes": pulumi.Array{
			pulumi.Map{
				"name": pulumi.String("hcloud-config"),
				"secret": pulumi.Map{
					"secretName": pulumi.String("hcloud-autoscaler-cluster-config"),
				},
			},
		},
		"extraVolumeMounts": pulumi.Array{
			pulumi.Map{
				"name":      pulumi.String("hcloud-config"),
				"mountPath": pulumi.String("/etc/kubernetes/hcloud_cluster_config"),
				"readOnly":  pulumi.Bool(true),
			},
		},
	}

	values := pulumi.Map{}
	if args.Values != nil {
		values = pulumi.ToMap(*args.Values)
	}

	err = mergo.Merge(&values, preDefineValues, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	clusterAutoscaler, err := helmv4.NewChart(ctx, "cluster-autoscaler", &helmv4.ChartArgs{
		Chart:     pulumi.String("cluster-autoscaler"),
		Namespace: pulumi.String("kube-system"),
		RepositoryOpts: &helmv4.RepositoryOptsArgs{
			Repo: pulumi.String("https://kubernetes.github.io/autoscaler"),
		},
		Version: pulumi.StringPtrFromPtr(args.Version),
		Values:  values,
	}, append(opts,
		pulumi.Parent(autoscalerConfig.AutoscalerSecret),
		pulumi.DependsOn([]pulumi.Resource{
			autoscalerConfig.AutoscalerSecret,
			autoscalerConfig.AutoscalerClusterConfig,
		}),
	)...)
	if err != nil {
		return nil, err
	}

	return &ClusterAutoscaler{
		Chart:                       clusterAutoscaler,
		AutoscalerSecret:            autoscalerConfig.AutoscalerSecret,
		AutoscalerClusterConfigHash: autoscalerConfig.ClusterConfigJSONHash,
		AutoscalerClusterConfig:     autoscalerConfig.AutoscalerClusterConfig,
		AutoscalingGroupsConfig:     autoscalerConfig.AutoscalingGroups,
	}, nil
}
