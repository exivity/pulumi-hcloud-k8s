package compute_test

import (
	"testing"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/compute"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/core"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type mocks int

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := args.Inputs.Mappable()
	if args.TypeToken == "hcloud:index/server:Server" {
		outputs["ipv4Address"] = "10.0.0.1"
	}
	return args.Name + "_id", resource.NewPropertyMapFromMap(outputs), nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	if args.Token == "hcloud:index/getServers:getServers" {
		return resource.NewPropertyMapFromMap(map[string]interface{}{
			"servers": []interface{}{
				map[string]interface{}{
					"id":   123,
					"name": "worker-pool-autoscaler-1",
				},
			},
		}), nil
	}
	if args.Token == "talos:machine/getConfiguration:getConfiguration" {
		return resource.NewPropertyMapFromMap(map[string]interface{}{
			"machineConfiguration": "mock-configuration",
		}), nil
	}
	return args.Args, nil
}

func TestNodePool_FindNodePoolAutoScalerNodes(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		tests := []struct {
			name     string
			nodePool *compute.NodePool
			wantErr  bool
		}{
			{
				name: "ControlPlaneNode",
				nodePool: &compute.NodePool{
					NodePoolName:   "control-plane",
					ServerNodeType: meta.ControlPlaneNode,
				},
				wantErr: true,
			},
			{
				name: "WorkerNode",
				nodePool: &compute.NodePool{
					NodePoolName:   "worker-pool",
					ServerNodeType: meta.WorkerNode,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotErr := tt.nodePool.FindNodePoolAutoScalerNodes(ctx)
				if (gotErr != nil) != tt.wantErr {
					t.Errorf("FindNodePoolAutoScalerNodes() error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
				if !tt.wantErr && len(tt.nodePool.AutoScalerNodes) == 0 {
					t.Errorf("FindNodePoolAutoScalerNodes() expected autoscaler nodes, got none")
				}
			})
		}
		return nil
	}, pulumi.WithMocks("project", "stack", mocks(0)))

	if err != nil {
		t.Fatalf("pulumi.RunErr failed: %v", err)
	}
}

func TestNodePool_ApplyConfigPatches(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Setup MachineConfigurationManager
		mcm, err := core.NewMachineConfigurationManager(ctx, "test-cluster", &core.MachineConfigurationManagerArgs{
			SingleControlPlaneNodeIP: pulumi.String("1.2.3.4"),
			TalosVersion:             "v1.0.0",
			KubernetesVersion:        "v1.24.0",
		})
		if err != nil {
			return err
		}

		// Create a mock node
		node, err := hcloud.NewServer(ctx, "test-node", &hcloud.ServerArgs{
			ServerType: pulumi.String("cx11"),
			Image:      pulumi.String("ubuntu-20.04"),
		})
		if err != nil {
			return err
		}

		tests := []struct {
			name      string
			nodePool  *compute.NodePool
			wantErr   bool
			wantCount int
		}{
			{
				name: "WorkerNode",
				nodePool: &compute.NodePool{
					NodePoolName:                "worker-pool",
					ServerNodeType:              meta.WorkerNode,
					MachineConfigurationManager: mcm,
					Nodes:                       []*hcloud.Server{node},
				},
				wantErr:   false,
				wantCount: 1,
			},
			{
				name: "WorkerNodeWithAutoScaler",
				nodePool: &compute.NodePool{
					NodePoolName:                "worker-pool-autoscaler",
					ServerNodeType:              meta.WorkerNode,
					MachineConfigurationManager: mcm,
					Nodes:                       []*hcloud.Server{node},
					AutoScalerNodes: []hcloud.GetServersServer{
						{
							Name:        "autoscaler-node-1",
							Ipv4Address: "10.0.0.2",
						},
					},
				},
				wantErr:   false,
				wantCount: 2,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, gotErr := tt.nodePool.ApplyConfigPatches(ctx)
				if (gotErr != nil) != tt.wantErr {
					t.Errorf("ApplyConfigPatches() error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
				if !tt.wantErr && len(got) != tt.wantCount {
					t.Errorf("ApplyConfigPatches() expected %d result, got %d", tt.wantCount, len(got))
				}
			})
		}
		return nil
	}, pulumi.WithMocks("project", "stack", mocks(0)))

	if err != nil {
		t.Fatalf("pulumi.RunErr failed: %v", err)
	}
}
