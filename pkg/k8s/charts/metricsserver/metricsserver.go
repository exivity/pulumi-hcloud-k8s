package metricsserver

import (
	"dario.cat/mergo"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Args struct {
	// Values are the values to use for the chart
	Values *map[string]interface{}
	// Version is the version of the chart to use
	// The version must be available in the chart repository.
	// If not set, the latest version will be used.
	Version *string `json:"version"`
}

type MetricServer struct {
	Chart *helmv4.Chart
}

func New(ctx *pulumi.Context, args *Args, opts ...pulumi.ResourceOption) (*MetricServer, error) {
	preDefineValues := pulumi.Map{}

	values := pulumi.Map{}
	if args.Values != nil {
		values = pulumi.ToMap(*args.Values)
	}

	err := mergo.Merge(&values, preDefineValues, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	ccmChart, err := helmv4.NewChart(ctx, "metrics-server", &helmv4.ChartArgs{
		Chart:     pulumi.String("metrics-server"),
		Namespace: pulumi.String("kube-system"),
		RepositoryOpts: &helmv4.RepositoryOptsArgs{
			Repo: pulumi.String("https://kubernetes-sigs.github.io/metrics-server/"),
		},
		Version: pulumi.StringPtrFromPtr(args.Version),
		Values:  values,
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &MetricServer{
		Chart: ccmChart,
	}, nil
}
