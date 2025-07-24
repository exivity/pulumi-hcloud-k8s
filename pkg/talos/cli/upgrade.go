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
	Images       *image.Images
	// NodeIpv4Address is the IPv4 address of the node
	NodeIpv4Address pulumi.StringOutput
	// NodeImage is the image of the node
	// This is used to determine the image ID for the upgrade
	NodeImage pulumi.StringPtrOutput
}

func TalosConfigPath(ctx *pulumi.Context) string {
	return fmt.Sprintf("%s.talosconfig.json", ctx.Stack())
}

// writeScriptToProjectTmp writes the embedded script to a persistent project subfolder and returns its path
func writeScriptToProjectTmp() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cwd, scriptSubdir)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return "", err
	}
	file := filepath.Join(dir, "talos-upgrade-version.sh")
	if err := os.WriteFile(file, talosUpgradeScript, filePerm); err != nil {
		return "", err
	}
	// Set executable permission on the script
	if err := os.Chmod(file, scriptPerm); err != nil {
		return "", err
	}
	return file, nil
}

// UpgradeTalos upgrades the Talos version on a node
func UpgradeTalos(ctx *pulumi.Context, name string, args *UpgradeTalosArgs, opts ...pulumi.ResourceOption) (*local.Command, error) {
	armImage, err := args.Images.GetImageByArch(image.ArchARM)
	if err != nil {
		return nil, err
	}

	x86Image, err := args.Images.GetImageByArch(image.ArchX86)
	if err != nil {
		return nil, err
	}

	scriptPath, err := writeScriptToProjectTmp()
	if err != nil {
		return nil, err
	}

	return local.NewCommand(ctx, fmt.Sprintf("upgrade-talos-%s", name), &local.CommandArgs{
		Create: pulumi.String(scriptPath),
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
}
