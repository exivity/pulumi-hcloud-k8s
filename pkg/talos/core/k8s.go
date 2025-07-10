package core

import (
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/cluster"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
)

type KubeconfigArgs struct {
	// CertificateRenewalDuration is the duration for which the certificate is valid
	CertificateRenewalDuration string
	// FirstControlPlane is the first control plane node
	FirstControlPlane *hcloud.Server
	// Talos Linux secrets for the cluster
	Secrets *machine.Secrets
}

type Kubeconfig struct {
	Bootstrap  *machine.Bootstrap
	Kubeconfig *cluster.Kubeconfig
}

func NewKubeconfig(ctx *pulumi.Context, args *KubeconfigArgs, opts ...pulumi.ResourceOption) (*Kubeconfig, error) {
	bootstrap, err := machine.NewBootstrap(ctx, "bootstrap", &machine.BootstrapArgs{
		Node:                args.FirstControlPlane.Ipv4Address,
		ClientConfiguration: args.Secrets.ClientConfiguration,
	}, append(opts,
		pulumi.Parent(args.Secrets),
		pulumi.IgnoreChanges([]string{"node"}),
	)...)
	if err != nil {
		return nil, err
	}

	k8s, err := cluster.NewKubeconfig(ctx, "kubeconfigResource", &cluster.KubeconfigArgs{
		ClientConfiguration: &cluster.KubeconfigClientConfigurationArgs{
			CaCertificate:     bootstrap.ClientConfiguration.CaCertificate(),
			ClientCertificate: bootstrap.ClientConfiguration.ClientCertificate(),
			ClientKey:         bootstrap.ClientConfiguration.ClientKey(),
		},
		Node:                       args.FirstControlPlane.Ipv4Address,
		CertificateRenewalDuration: pulumi.String(args.CertificateRenewalDuration),
		Endpoint:                   args.FirstControlPlane.Ipv4Address,
	},
		pulumi.Parent(bootstrap),
		pulumi.IgnoreChanges([]string{"node"}), // Ignore changes to the node address, as it may change after initial creation
	)
	if err != nil {
		return nil, err
	}

	return &Kubeconfig{
		Bootstrap:  bootstrap,
		Kubeconfig: k8s,
	}, nil
}
