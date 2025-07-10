package ccm

import (
	"dario.cat/mergo"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/network"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CloudControlManagerArgs struct {
	// Network is the network to use for the cluster
	Network *network.Network
	// PodSubnets is the pod subnets to use for the cluster
	PodSubnets string
	// Values are the values to use for the chart
	Values *map[string]interface{}
	// Version is the version of the chart to use
	// The version must be available in the chart repository.
	// If not set, the latest version will be used.
	Version *string `json:"version"`
}

type CloudControlManager struct {
	Chart *helmv4.Chart
}

func NewCloudControlManager(ctx *pulumi.Context, args *CloudControlManagerArgs, opts ...pulumi.ResourceOption) (*CloudControlManager, error) {
	// hcloudLoadBalancersDisablePrivateIngress := !args.EnablePrivateIngress

	preDefineValues := pulumi.Map{
		"env": pulumi.Map{
			"HCLOUD_LOAD_BALANCERS_NETWORK_ZONE": pulumi.Map{
				"value": args.Network.NetworkSubnet.NetworkZone,
			},
			"HCLOUD_LOAD_BALANCERS_USE_PRIVATE_IP": pulumi.Map{
				"value": pulumi.String("true"),
			},
		},
		"networking": pulumi.Map{
			"enabled":     pulumi.Bool(true),
			"clusterCIDR": pulumi.String(args.PodSubnets),
		},
	}

	values := pulumi.Map{}
	if args.Values != nil {
		values = pulumi.ToMap(*args.Values)
	}

	err := mergo.Merge(&values, preDefineValues, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	ccmChart, err := helmv4.NewChart(ctx, "hcloud-cloud-controller-manager", &helmv4.ChartArgs{
		Chart:     pulumi.String("hcloud-cloud-controller-manager"),
		Namespace: pulumi.String("kube-system"),
		RepositoryOpts: &helmv4.RepositoryOptsArgs{
			Repo: pulumi.String("https://charts.hetzner.cloud"),
		},
		Version: pulumi.StringPtrFromPtr(args.Version),
		Values:  values,
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &CloudControlManager{
		Chart: ccmChart,
	}, nil
}
