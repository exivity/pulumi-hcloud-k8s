package compute

import (
	"testing"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/core"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/exivity/pulumi-hcloud-upload-image/sdk/go/pulumi-hcloud-upload-image/hcloudimages"
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
	if args.TypeToken == "hcloud-upload-image:index:UploadedImage" {
		outputs["imageId"] = 12345
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
			nodePool *NodePool
			wantErr  bool
		}{
			{
				name: "ControlPlaneNode",
				nodePool: &NodePool{
					NodePoolName:   "control-plane",
					ServerNodeType: meta.ControlPlaneNode,
				},
				wantErr: true,
			},
			{
				name: "WorkerNode",
				nodePool: &NodePool{
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
			nodePool  *NodePool
			wantErr   bool
			wantCount int
		}{
			{
				name: "WorkerNode",
				nodePool: &NodePool{
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
				nodePool: &NodePool{
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

func TestNodePool_UpgradeTalos(t *testing.T) {
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

		// Create mock images
		armSnapshot, err := hcloudimages.NewUploadedImage(ctx, "arm-image", &hcloudimages.UploadedImageArgs{
			HcloudToken:  pulumi.String("token"),
			Architecture: pulumi.String("arm"),
			ImageUrl:     pulumi.String("http://example.com/image.raw.xz"),
			ServerType:   pulumi.String("cax11"),
			Location:     pulumi.String("fsn1"),
		})
		if err != nil {
			return err
		}
		x86Snapshot, err := hcloudimages.NewUploadedImage(ctx, "x86-image", &hcloudimages.UploadedImageArgs{
			HcloudToken:  pulumi.String("token"),
			Architecture: pulumi.String("x86"),
			ImageUrl:     pulumi.String("http://example.com/image.raw.xz"),
			ServerType:   pulumi.String("cx21"),
			Location:     pulumi.String("fsn1"),
		})
		if err != nil {
			return err
		}

		images := &image.Images{
			ARM:          &image.Image{Snapshot: armSnapshot},
			X86:          &image.Image{Snapshot: x86Snapshot},
			TalosImageID: "v1.0.0",
		}

		tests := []struct {
			name      string
			nodePool  *NodePool
			args      *UpgradeTalosArgs
			wantErr   bool
			wantCount int
		}{
			{
				name: "WorkerNode",
				nodePool: &NodePool{
					NodePoolName:                "worker-pool",
					ServerNodeType:              meta.WorkerNode,
					MachineConfigurationManager: mcm,
					Nodes:                       []*hcloud.Server{node},
				},
				args: &UpgradeTalosArgs{
					Talosconfig:  pulumi.Sprintf("talosconfig"),
					TalosVersion: "v1.2.3",
					Images:       images,
				},
				wantErr:   false,
				wantCount: 1,
			},
			{
				name: "WorkerNodeWithAutoScaler",
				nodePool: &NodePool{
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
				args: &UpgradeTalosArgs{
					Talosconfig:  pulumi.Sprintf("talosconfig"),
					TalosVersion: "v1.2.3",
					Images:       images,
				},
				wantErr:   false,
				wantCount: 2,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Reset the queue
				talosUpgradeQueue = []pulumi.Resource{}

				got, gotErr := tt.nodePool.UpgradeTalos(ctx, tt.args)
				if (gotErr != nil) != tt.wantErr {
					t.Errorf("UpgradeTalos() error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
				if !tt.wantErr && len(got) != tt.wantCount {
					t.Errorf("UpgradeTalos() expected %d result, got %d", tt.wantCount, len(got))
				}
			})
		}
		return nil
	}, pulumi.WithMocks("project", "stack", mocks(0)))

	if err != nil {
		t.Fatalf("pulumi.RunErr failed: %v", err)
	}
}
