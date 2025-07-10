package provider

import (
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ProviderArgs struct {
	// Token is the Hetzner Cloud API token
	Token string
}

func NewProvider(ctx *pulumi.Context, name string, args *ProviderArgs, opts ...pulumi.ResourceOption) (*hcloud.Provider, error) {
	return hcloud.NewProvider(ctx, name, &hcloud.ProviderArgs{
		Token: pulumi.String(args.Token),
	}, opts...)
}
