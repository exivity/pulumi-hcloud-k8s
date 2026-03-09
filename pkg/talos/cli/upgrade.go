package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//go:embed talos-upgrade-version.sh
var talosUpgradeScript []byte

//go:embed talos-delete-node.sh
var talosDeleteScript []byte

const scriptSubdir = ".pulumi-tmp"
const (
	dirPerm    = 0o700
	filePerm   = 0o600
	scriptPerm = 0o700 // permission for executable scripts
)

// UpgradeTalosArgs are the arguments for the UpgradeTalos function
type UpgradeTalosArgs struct {
	// Talosconfig is the Talos configuration
	Talosconfig pulumi.StringOutput
	// TalosVersion is the version of Talos to upgrade to
	TalosVersion string
	// Images is the image information for the upgrade
	Images *image.Images
	// NodeIpv4Address is the IPv4 address of the node
	NodeIpv4Address pulumi.StringOutput
	// NodeImage is the image of the node
	// This is used to determine the image ID for the upgrade
	NodeImage pulumi.StringPtrOutput
	// Protection is the protection status of the node
	Protection bool
	// RemoveNodeFromClusterOnDelete determines whether the node should be removed from the cluster before deletion
	RemoveNodeFromClusterOnDelete bool
}

// UpgradeTalos upgrades the Talos version on a node
type UpgradeTalos struct {
	Upgrade *local.Command
	Delete  *local.Command
}

// DeleteTalosArgs are the arguments for the DeleteTalos function
type DeleteTalosArgs struct {
	// Talosconfig is the Talos configuration
	Talosconfig pulumi.StringOutput
	// NodeIpv4Address is the IPv4 address of the node
	NodeIpv4Address pulumi.StringOutput
}

func TalosConfigPath(ctx *pulumi.Context) string {
	return fmt.Sprintf("%s.talosconfig.json", ctx.Stack())
}

// writeScriptToProjectTmp writes the embedded script to a persistent project subfolder and returns its path
func writeScriptToProjectTmp(name string, content []byte) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cwd, scriptSubdir)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return "", err
	}
	file := filepath.Join(dir, name)
	if err := os.WriteFile(file, content, filePerm); err != nil {
		return "", err
	}
	// Set executable permission on the script
	if err := os.Chmod(file, scriptPerm); err != nil {
		return "", err
	}
	return filepath.Join(scriptSubdir, name), nil
}

// UpgradeTalos upgrades the Talos version on a node
func NewUpgradeTalos(ctx *pulumi.Context, name string, args *UpgradeTalosArgs, opts ...pulumi.ResourceOption) (*UpgradeTalos, error) {
	armImage, err := args.Images.GetImageByArch(image.ArchARM)
	if err != nil {
		return nil, err
	}

	x86Image, err := args.Images.GetImageByArch(image.ArchX86)
	if err != nil {
		return nil, err
	}

	upgradeScriptPath, err := writeScriptToProjectTmp("talos-upgrade-version.sh", talosUpgradeScript)
	if err != nil {
		return nil, err
	}

	deleteScriptPath, err := writeScriptToProjectTmp("talos-delete-node-from-cluster.sh", talosDeleteScript)
	if err != nil {
		return nil, err
	}

	upgrade, err := local.NewCommand(ctx, fmt.Sprintf("upgrade-talos-%s", name), &local.CommandArgs{
		Create: pulumi.String(upgradeScriptPath),
		Environment: pulumi.StringMap{
			"TALOSCONFIG":       pulumi.String(TalosConfigPath(ctx)),
			"TALOSCONFIG_VALUE": args.Talosconfig,
			"TALOS_VERSION":     pulumi.String(args.TalosVersion),
			"TALOS_IMAGE":       pulumi.String(args.Images.TalosImageID),
			"ARM_IMAGE":         pulumi.Sprintf("%d", armImage.ImageId()),
			"X86_IMAGE":         pulumi.Sprintf("%d", x86Image.ImageId()),
			"NODE_NAME":         pulumi.String(name),
			"NODE_IP":           args.NodeIpv4Address,
			"NODE_IMAGE": args.NodeImage.ApplyT(func(image *string) string {
				return *image
			}).(pulumi.StringOutput),
		},
		Triggers: pulumi.Array{
			pulumi.String(args.TalosVersion),
			armImage.ImageId(),
			x86Image.ImageId(),
			args.NodeIpv4Address,
			args.NodeImage,
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	var delete *local.Command
	if args.RemoveNodeFromClusterOnDelete {
		delete, err = local.NewCommand(ctx, fmt.Sprintf("delete-talos-%s", name), &local.CommandArgs{
			Delete: pulumi.String(deleteScriptPath),
			Environment: pulumi.StringMap{
				"TALOSCONFIG":       pulumi.String(TalosConfigPath(ctx)),
				"TALOSCONFIG_VALUE": args.Talosconfig,
				"NODE_IP":           args.NodeIpv4Address,
			},
			Triggers: pulumi.Array{
				args.NodeIpv4Address,
			},
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	return &UpgradeTalos{
		Upgrade: upgrade,
		Delete:  delete,
	}, nil
}
