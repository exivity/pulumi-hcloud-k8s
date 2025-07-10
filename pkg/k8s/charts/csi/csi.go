package csi

import (
	"dario.cat/mergo"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CSIArgs struct {
	// Values are the values to use for the chart
	Values *map[string]interface{}
	// Version is the version of the chart to use
	// The version must be available in the chart repository.
	// If not set, the latest version will be used.
	Version *string `json:"version"`
	// EncryptedSecret is the encrypted secret used to encrypt the CSI driver
	EncryptedSecret string `json:"encrypted_secret" validate:"required"`
	// IsDefaultStorageClass is the default storage class
	IsDefaultStorageClass bool
	// ReclaimPolicy controls what happens to volumes when PVCs are deleted.
	ReclaimPolicy string
}

type CSI struct {
	Chart *helmv4.Chart
}

func NewCSI(ctx *pulumi.Context, args *CSIArgs, opts ...pulumi.ResourceOption) (*CSI, error) {
	encryptionSecret, err := corev1.NewSecret(ctx, "encryption-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("encryption-secret"),
			Namespace: pulumi.String("kube-system"),
		},
		StringData: pulumi.StringMap{
			"encryption-passphrase": pulumi.String(args.EncryptedSecret),
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	preDefineValues := pulumi.Map{
		"storageClasses": pulumi.Array{
			pulumi.Map{
				"name":                pulumi.String("hcloud-volumes"),
				"defaultStorageClass": pulumi.Bool(args.IsDefaultStorageClass),
				"reclaimPolicy":       pulumi.String(args.ReclaimPolicy),
				"extraParameters": pulumi.Map{
					"csi.storage.k8s.io/node-publish-secret-name":      encryptionSecret.Metadata.Name(),
					"csi.storage.k8s.io/node-publish-secret-namespace": encryptionSecret.Metadata.Namespace(),
				},
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

	ccmChart, err := helmv4.NewChart(ctx, "hcloud-csi", &helmv4.ChartArgs{
		Chart:     pulumi.String("hcloud-csi"),
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

	return &CSI{
		Chart: ccmChart,
	}, nil
}
