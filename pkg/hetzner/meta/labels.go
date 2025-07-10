package meta

import (
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ServerNodeType represents the type of server node
type ServerNodeType string

const (
	// ControlPlaneNode is a control plane node
	ControlPlaneNode ServerNodeType = "controlplane"
	// WorkerNode is a worker node
	WorkerNode ServerNodeType = "worker"
	// NoneNode is a node without a specific type
	NoneNode ServerNodeType = ""
	// NodePoolLabel is the label used to identify the node pool
	NodePoolLabel = "hcloud/node-group"
)

// LabelsArgs are the arguments for the Labels function
type ServerLabelsArgs struct {
	// ServerNodeType is the type of the server node
	ServerNodeType ServerNodeType
	// Region is the region of the server
	Region *string
	// Arch is the architecture of the server
	Arch *image.CPUArchitecture
	// NodePoolName is the name of the node pool
	NodePoolName *string
}

// NewLabels generates the labels for a hetzner resource like Server and LoadBalancer
func NewLabels(ctx *pulumi.Context, args *ServerLabelsArgs) pulumi.StringMap {
	labels := pulumi.StringMap{
		"stack":   pulumi.String(ctx.Stack()),
		"project": pulumi.String(ctx.Project()),
	}

	if args.ServerNodeType != "" {
		labels["type"] = pulumi.String(string(args.ServerNodeType))
	}

	if args.NodePoolName != nil {
		labels[NodePoolLabel] = pulumi.String(*args.NodePoolName)
	}
	if args.Region != nil {
		labels["region"] = pulumi.String(*args.Region)
	}
	if args.Arch != nil {
		labels["arch"] = pulumi.String(*args.Arch)
	}

	return labels
}
