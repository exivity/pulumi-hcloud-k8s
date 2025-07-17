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
	EnableLonghorn              bool
	Network                     *network.Network
	HcloudToken                 string
	// Firewall is the firewall to use for the nodes
	Firewall *hcloud.Firewall
}

type ClusterAutoscaler struct {
	Chart *helmv4.Chart
}

func NewClusterAutoscaler(ctx *pulumi.Context, args *ClusterAutoscalerArgs, opts ...pulumi.ResourceOption) (*ClusterAutoscaler, error) { //nolint:cyclop,funlen // TODO: refactor
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
		if pool.AutoScaler == nil {
			continue
		}

		workerNodeConfiguration, err := core.NewNodeConfiguration(&core.NodeConfigurationArgs{
			ServerNodeType:        meta.WorkerNode,
			Subnet:                args.Subnet,
			PodSubnets:            args.PodSubnets,
			NodeLabels:            pool.Labels,
			NodeTaints:            pool.Taints,
			NodeAnnotations:       pool.Annotations,
			EnableLonghornSupport: args.EnableLonghorn,
			BootstrapEnable:       true,
		})
		if err != nil {
			return nil, err
		}

		workerMachineConfiguration := args.MachineConfigurationManager.NewMachineConfiguration(ctx, &core.MachineConfigurationArgs{
			ServerNodeType: meta.WorkerNode,
			ConfigPatches: pulumi.StringArray{
				pulumi.String(workerNodeConfiguration),
			},
		})

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
		for _, taint := range pool.Taints { // TODO: does this make sense?
			nodeConfig.Taints = append(nodeConfig.Taints, Taint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: taint.Effect,
			})
		}

		nodeConfigs[pool.Name] = nodeConfig

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
			ARM64: imgARM.Snapshot.ImageId,
			AMD64: imgX86.Snapshot.ImageId,
		},
		NodeConfigs: nodeConfigs,
	}
	clusterConfigJSON, err := clusterConfig.ToJSON()
	if err != nil {
		return nil, err
	}

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
		"autoscalingGroups": autoscalingGroups,
		"autoDiscovery": pulumi.Map{
			"enabled": pulumi.Bool(false),
		},
	}

	values := pulumi.Map{
		"extraEnv": pulumi.StringMap{
			"HCLOUD_CLUSTER_CONFIG_FILE": pulumi.String("/etc/kubernetes/hcloud_cluster_config.json"),
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
				"mountPath": pulumi.String("/etc/kubernetes/hcloud_cluster_config.json"),
				"subPath":   pulumi.String("hcloud-autoscaler-cluster-config"),
			},
		},
	}
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
		pulumi.Parent(autoscalerSecret),
		pulumi.DependsOn([]pulumi.Resource{
			autoscalerSecret,
			autoscalerClusterConfig,
		}),
	)...)
	if err != nil {
		return nil, err
	}

	return &ClusterAutoscaler{
		Chart: clusterAutoscaler,
	}, nil
}
