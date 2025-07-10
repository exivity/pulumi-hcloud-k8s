package longhorn

import (
	"dario.cat/mergo"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type LonghornArgs struct {
	// Values are the values to use for the chart
	Values *map[string]interface{}
	// Version is the version of the chart to use
	// The version must be available in the chart repository.
	// If not set, the latest version will be used.
	Version *string `json:"version"`
}

type Longhorn struct {
	Namespace *corev1.Namespace
	Chart     *helmv4.Chart
}

func NewLonghorn(ctx *pulumi.Context, args *LonghornArgs, opts ...pulumi.ResourceOption) (*Longhorn, error) {
	longhornNS, err := corev1.NewNamespace(ctx, "longhorn-system", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("longhorn-system"),
			Labels: pulumi.StringMap{
				"pod-security.kubernetes.io/enforce": pulumi.String("privileged"),
				"pod-security.kubernetes.io/audit":   pulumi.String("privileged"),
				"pod-security.kubernetes.io/warn":    pulumi.String("privileged"),
			},
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	preDefineValues := pulumi.Map{
		"csi": pulumi.Map{
			"kubeletRootDir": pulumi.String("/var/lib/kubelet"),
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

	longhorn, err := helmv4.NewChart(ctx, "longhorn", &helmv4.ChartArgs{
		Chart:     pulumi.String("longhorn"),
		Namespace: longhornNS.Metadata.Name(),
		RepositoryOpts: &helmv4.RepositoryOptsArgs{
			Repo: pulumi.String("https://charts.longhorn.io"),
		},
		Version: pulumi.StringPtrFromPtr(args.Version),
		Values:  values,
	}, append(opts,
		pulumi.Parent(longhornNS),
	)...)
	if err != nil {
		return nil, err
	}

	return &Longhorn{
		Namespace: longhornNS,
		Chart:     longhorn,
	}, nil
}
