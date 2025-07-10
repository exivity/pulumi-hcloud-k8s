package compute

import (
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
)

type PlacementGroupArgs struct {
	ServerNodeType meta.ServerNodeType
}

type PlacementGroup struct {
	PlacementGroup *hcloud.PlacementGroup
}

func NewPlacementGroup(ctx *pulumi.Context, name string, args *PlacementGroupArgs, opts ...pulumi.ResourceOption) (*hcloud.PlacementGroup, error) {
	return hcloud.NewPlacementGroup(ctx, name, &hcloud.PlacementGroupArgs{
		Name:   pulumi.String(name),
		Type:   pulumi.String("spread"),
		Labels: meta.NewLabels(ctx, &meta.ServerLabelsArgs{ServerNodeType: args.ServerNodeType}),
	}, opts...)
}
