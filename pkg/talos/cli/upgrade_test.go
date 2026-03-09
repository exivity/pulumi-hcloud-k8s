package cli

import (
	_ "embed"
	"path/filepath"
	"testing"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/exivity/pulumi-hcloud-upload-image/sdk/go/pulumi-hcloud-upload-image/hcloudimages"
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
	type args struct {
		name    string
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				name:    "talos-upgrade-version.sh",
				content: []byte("test content"),
			},
			want:    filepath.Join(".pulumi-tmp", "talos-upgrade-version.sh"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := writeScriptToProjectTmp(tt.args.name, tt.args.content)
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

func TestNewUpgradeTalos(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
		setupArgs    func(ctx *pulumi.Context) (*UpgradeTalosArgs, error)
		validate     func(t *testing.T, result *UpgradeTalos)
		wantErr      bool
	}{
		{
			name:         "default",
			resourceName: "test-node",
			setupArgs: func(ctx *pulumi.Context) (*UpgradeTalosArgs, error) {
				// Create mock images manually
				armSnapshot := &hcloudimages.UploadedImage{
					ImageId: pulumi.Int(12345).ToIntOutput(),
				}
				x86Snapshot := &hcloudimages.UploadedImage{
					ImageId: pulumi.Int(12345).ToIntOutput(),
				}

				images := &image.Images{
					ARM:          &image.Image{Snapshot: armSnapshot},
					X86:          &image.Image{Snapshot: x86Snapshot},
					TalosImageID: "v1.0.0",
				}

				return &UpgradeTalosArgs{
					Talosconfig:                   pulumi.Sprintf("talosconfig-content"),
					TalosVersion:                  "v1.2.3",
					Images:                        images,
					NodeIpv4Address:               pulumi.Sprintf("1.2.3.4"),
					NodeImage:                     pulumi.String("node-image").ToStringPtrOutput(),
					RemoveNodeFromClusterOnDelete: false,
				}, nil
			},
			validate: func(t *testing.T, result *UpgradeTalos) {
				assert.NotNil(t, result.Upgrade)
				assert.Nil(t, result.Delete)
				pulumi.All(result.Upgrade.Environment).ApplyT(func(args []interface{}) error {
					env := args[0].(map[string]string)
					assert.Equal(t, "talosconfig-content", env["TALOSCONFIG_VALUE"])
					assert.Equal(t, "v1.2.3", env["TALOS_VERSION"])
					assert.Equal(t, "v1.0.0", env["TALOS_IMAGE"])
					assert.Equal(t, "12345", env["ARM_IMAGE"])
					assert.Equal(t, "12345", env["X86_IMAGE"])
					assert.Equal(t, "test-node", env["NODE_NAME"])
					assert.Equal(t, "1.2.3.4", env["NODE_IP"])
					assert.Equal(t, "node-image", env["NODE_IMAGE"])
					return nil
				})
			},
			wantErr: false,
		},
		{
			name:         "with delete command",
			resourceName: "test-node-with-delete",
			setupArgs: func(ctx *pulumi.Context) (*UpgradeTalosArgs, error) {
				// Create mock images manually
				armSnapshot := &hcloudimages.UploadedImage{
					ImageId: pulumi.Int(12345).ToIntOutput(),
				}
				x86Snapshot := &hcloudimages.UploadedImage{
					ImageId: pulumi.Int(12345).ToIntOutput(),
				}

				images := &image.Images{
					ARM:          &image.Image{Snapshot: armSnapshot},
					X86:          &image.Image{Snapshot: x86Snapshot},
					TalosImageID: "v1.0.0",
				}

				return &UpgradeTalosArgs{
					Talosconfig:                   pulumi.Sprintf("talosconfig-content"),
					TalosVersion:                  "v1.2.3",
					Images:                        images,
					NodeIpv4Address:               pulumi.Sprintf("1.2.3.4"),
					NodeImage:                     pulumi.String("node-image").ToStringPtrOutput(),
					RemoveNodeFromClusterOnDelete: true,
				}, nil
			},
			validate: func(t *testing.T, result *UpgradeTalos) {
				assert.NotNil(t, result.Upgrade)
				assert.NotNil(t, result.Delete)
				pulumi.All(result.Delete.Environment).ApplyT(func(args []interface{}) error {
					env := args[0].(map[string]string)
					assert.Equal(t, "talosconfig-content", env["TALOSCONFIG_VALUE"])
					assert.Equal(t, "1.2.3.4", env["NODE_IP"])
					return nil
				})
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				args, err := tt.setupArgs(ctx)
				if err != nil {
					return err
				}
				got, err := NewUpgradeTalos(ctx, tt.resourceName, args)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewUpgradeTalos() error = %v, wantErr %v", err, tt.wantErr)
					return nil
				}
				if !tt.wantErr {
					tt.validate(t, got)
				}
				return nil
			}, pulumi.WithMocks("project", "stack", mocks(0)))
			assert.NoError(t, err)
		})
	}
}
