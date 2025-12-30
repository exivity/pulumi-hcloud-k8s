package cli

import (
	_ "embed"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
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

func TestTalosConfigPath(t *testing.T) {
	type args struct {
		project string
		stack   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default",
			args: args{
				project: "test-project",
				stack:   "test-stack",
			},
			want: "test-stack.talosconfig.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				if got := TalosConfigPath(ctx); got != tt.want {
					t.Errorf("TalosConfigPath() = %v, want %v", got, tt.want)
				}
				return nil
			}, pulumi.WithMocks(tt.args.project, tt.args.stack, mocks(0)))
			assert.NoError(t, err)
		})
	}
}

func Test_writeScriptToProjectTmp(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "default",
			want:    filepath.Join(".pulumi-tmp", "talos-upgrade-version.sh"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := writeScriptToProjectTmp()
			if (err != nil) != tt.wantErr {
				t.Errorf("writeScriptToProjectTmp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("writeScriptToProjectTmp() = %v, want %v", got, tt.want)
			}
		})
	}
}
