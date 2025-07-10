package meta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type mocks int

// Create the mock.
func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := args.Inputs.Mappable()
	return args.Name + "_id", resource.NewPropertyMapFromMap(outputs), nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := map[string]interface{}{}
	return resource.NewPropertyMapFromMap(outputs), nil
}

func TestNewLabels(t *testing.T) {
	type args struct {
		project string
		stack   string
		args    *ServerLabelsArgs
	}
	tests := []struct {
		name string
		args args
		want pulumi.StringMap
	}{
		{
			name: "minimal labels",
			args: args{
				project: "testproj",
				stack:   "teststack",
				args:    &ServerLabelsArgs{},
			},
			want: pulumi.StringMap{
				"project": pulumi.String("testproj"),
				"stack":   pulumi.String("teststack"),
			},
		},
		{
			name: "all fields set",
			args: args{
				project: "proj",
				stack:   "stack",
				args: &ServerLabelsArgs{
					ServerNodeType: ControlPlaneNode,
					Region:         ptrString("eu-central"),
					Arch:           ptrArch("amd64"),
					NodePoolName:   ptrString("pool1"),
				},
			},
			want: pulumi.StringMap{
				"project":     pulumi.String("proj"),
				"stack":       pulumi.String("stack"),
				"type":        pulumi.String("controlplane"),
				"region":      pulumi.String("eu-central"),
				"arch":        pulumi.String("amd64"),
				NodePoolLabel: pulumi.String("pool1"),
			},
		},
		{
			name: "only node type and node pool",
			args: args{
				project: "p",
				stack:   "s",
				args: &ServerLabelsArgs{
					ServerNodeType: WorkerNode,
					NodePoolName:   ptrString("npool"),
				},
			},
			want: pulumi.StringMap{
				"project":     pulumi.String("p"),
				"stack":       pulumi.String("s"),
				"type":        pulumi.String("worker"),
				NodePoolLabel: pulumi.String("npool"),
			},
		},
		{
			name: "region and arch only",
			args: args{
				project: "prj",
				stack:   "stk",
				args: &ServerLabelsArgs{
					Region: ptrString("us-east"),
					Arch:   ptrArch("arm64"),
				},
			},
			want: pulumi.StringMap{
				"project": pulumi.String("prj"),
				"stack":   pulumi.String("stk"),
				"region":  pulumi.String("us-east"),
				"arch":    pulumi.String("arm64"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				if got := NewLabels(ctx, tt.args.args); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewLabels() = %v, want %v", got, tt.want)
				}
				return nil
			}, pulumi.WithMocks(tt.args.project, tt.args.stack, mocks(0)))
			assert.NoError(t, err)
		})
	}
}

// helper functions for test args
func ptrString(s string) *string { return &s }
func ptrArch(a string) *image.CPUArchitecture {
	arch := image.CPUArchitecture(a)
	return &arch
}
