package kubeletservingcertapprover

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Args struct {
	// Version is the version of the kubelet serving cert approver to use
	// The version must be a tag or branch of the kubelet serving cert approver
	// See: https://github.com/alex1989hu/kubelet-serving-cert-approver/tags
	Version string `json:"version"`
}

type KubeletServingCertApprover struct {
	Resource *yaml.ConfigFile
}

func New(ctx *pulumi.Context, args *Args, opts ...pulumi.ResourceOption) (*KubeletServingCertApprover, error) {
	r, err := yaml.NewConfigFile(ctx, "kubelet-serving-cert-approver",
		&yaml.ConfigFileArgs{
			File: fmt.Sprintf("https://raw.githubusercontent.com/alex1989hu/kubelet-serving-cert-approver/%s/deploy/standalone-install.yaml", args.Version),
		}, opts...)
	if err != nil {
		return nil, err
	}

	return &KubeletServingCertApprover{
		Resource: r,
	}, nil
}
