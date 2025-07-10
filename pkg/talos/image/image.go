package image

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//go:embed hcloud.pkr.hcl
var hcloudPkrHcl []byte

const packerFileName = "hcloud.pkr.hcl"
const packerFilePerm = 0o600

var ErrUnknownArchitecture = errors.New("unknown architecture")

// CPUArchitecture represents the CPU architecture of the image
type CPUArchitecture string

const (
	// ArchARM represents the ARM architecture
	ArchARM CPUArchitecture = "arm64"
	// ArchX86 represents the x86 architecture
	ArchX86 CPUArchitecture = "amd64"
)

type ImagesArgs struct {
	// Hetzner Token is the Hetzner Cloud API token.
	HetznerToken string
	// TalosVersion is the version of Talos to upload.
	TalosVersion string
	// TalosImageID is the ID of the Talos image to upload.
	TalosImageID string
	// ARMServerSize is the server type to use for the image upload. The size muss match the architecture.
	ARMServerSize string
	// X86ServerSize is the server type to use for the image upload. The size muss match the architecture.
	X86ServerSize string
	// ImageBuildRegion is the region to use for the image upload.
	// This is the region where the server will be created.
	ImageBuildRegion string
}

// Images represents the uploaded Talos images for both architectures
type Images struct {
	ARM *Image
	X86 *Image
	// TalosImageID is the ID of the Talos image to upload.
	TalosImageID string
}

// NewImages uploads Talos images for both architectures to Hetzner Cloud
func NewImages(ctx *pulumi.Context, args *ImagesArgs, opts ...pulumi.ResourceOption) (*Images, error) {
	arm, err := NewImage(ctx, &ImageArgs{
		HetznerToken:     args.HetznerToken,
		TalosVersion:     args.TalosVersion,
		TalosImageID:     args.TalosImageID,
		Arch:             ArchARM,
		ServerSize:       args.ARMServerSize,
		ImageBuildRegion: args.ImageBuildRegion,
	}, opts...)
	if err != nil {
		return nil, err
	}

	x86, err := NewImage(ctx, &ImageArgs{
		HetznerToken:     args.HetznerToken,
		TalosVersion:     args.TalosVersion,
		TalosImageID:     args.TalosImageID,
		Arch:             ArchX86,
		ServerSize:       args.X86ServerSize,
		ImageBuildRegion: args.ImageBuildRegion,
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &Images{
		ARM:          arm,
		X86:          x86,
		TalosImageID: args.TalosImageID,
	}, nil
}

type ImageArgs struct {
	// Hetzner Token is the Hetzner Cloud API token.
	HetznerToken string
	// TalosVersion is the version of Talos to upload.
	TalosVersion string
	// TalosImageID is the ID of the Talos image to upload.
	TalosImageID string
	// Arch is the architecture of the image to upload. Must be either "amd64" or "arm64".
	Arch CPUArchitecture
	// ServerSize is the server type to use for the image upload. The size muss match the architecture.
	// Like "cx22" for "amd64" or "cax11" for "arm64".
	// All available server types can be found here https://www.hetzner.com/cloud/
	ServerSize string
	// ImageBuildRegion is the region to use for the image upload.
	// This is the region where the server will be created.
	ImageBuildRegion string
}

// Image represents the Packer command to upload a Talos image to Hetzner Cloud
type Image struct {
	Command *local.Command
}

// writePackerFileToProjectRoot writes the embedded hcloud.pkr.hcl to the project root
func writePackerFileToProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	file := filepath.Join(cwd, packerFileName)
	if err := os.WriteFile(file, hcloudPkrHcl, packerFilePerm); err != nil {
		return "", err
	}
	return file, nil
}

// NewImage uploads a Talos image to Hetzner Cloud using Packer.
func NewImage(ctx *pulumi.Context, args *ImageArgs, opts ...pulumi.ResourceOption) (*Image, error) {
	if args.Arch != ArchARM && args.Arch != ArchX86 {
		return nil, ErrUnknownArchitecture
	}

	if _, err := writePackerFileToProjectRoot(); err != nil {
		return nil, err
	}

	snapshotName := fmt.Sprintf("talos-%s-%s", args.Arch, args.TalosVersion)

	command, err := local.NewCommand(ctx, snapshotName, &local.CommandArgs{
		Create: pulumi.String(fmt.Sprintf(`packer init . && packer build -var 'talos_image_id=%s' -var 'talos_version=%s' -var 'arch=%s' -var 'server_type=%s' -var 'server_location=%s' -var 'snapshot_name=%s' .`, args.TalosImageID, args.TalosVersion, args.Arch, args.ServerSize, args.ImageBuildRegion, snapshotName)),
		Delete: pulumi.String(fmt.Sprintf(`
IMAGE_ID=$(go run github.com/hetznercloud/cli/cmd/hcloud image list --type snapshot -o json | jq -r '.[] | select(.description=="%s") | .id')
if [ -n "$IMAGE_ID" ]; then
    go run github.com/hetznercloud/cli/cmd/hcloud image delete "$IMAGE_ID"
    echo "Deleted image with ID: $IMAGE_ID"
else
    echo "Image not found, skipping delete."
fi
`, snapshotName)),
		Environment: pulumi.StringMap{
			"HCLOUD_TOKEN": pulumi.String(args.HetznerToken),
		},
	}, append(opts, pulumi.ReplaceOnChanges([]string{"create", "delete"}))...)
	if err != nil {
		return nil, err
	}

	return &Image{Command: command}, nil
}

func (u *Image) GetBuildID() pulumi.StringOutput {
	return u.Command.Stdout.ApplyT(func(output string) string {
		// Define a regex pattern to capture the build ID (e.g., "id=217055019")
		re := regexp.MustCompile(`\(ID:\s*(\d+)\)`)
		matches := re.FindStringSubmatch(output)

		if len(matches) > 1 {
			return matches[1] // Return the captured ID
		}
		return "" // Return empty if no match
	}).(pulumi.StringOutput)
}

func (u *Image) GetBuildIDString() string {
	var buildID string
	var wg sync.WaitGroup
	wg.Add(1)

	u.GetBuildID().ApplyT(func(output string) string {
		buildID = output
		defer wg.Done()
		return output
	})

	wg.Wait()
	return buildID
}

func (i *Images) GetImageByArch(arch CPUArchitecture) (*Image, error) {
	switch arch {
	case ArchARM:
		return i.ARM, nil
	case ArchX86:
		return i.X86, nil
	default:
		return nil, ErrUnknownArchitecture
	}
}
