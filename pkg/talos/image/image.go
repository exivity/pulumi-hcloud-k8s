package image

import (
	"errors"
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/exivity/pulumi-hcloud-upload-image/sdk/go/pulumi-hcloud-upload-image/hcloudimages"
)

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
	// EnableARMImageUpload enables the upload of the ARM image.
	EnableARMImageUpload bool
	// EnableX86ImageUpload enables the upload of the x86 image.
	EnableX86ImageUpload bool
	// TalosVersion is the version of Talos to upload.
	TalosVersion string
	// TalosImageID is the ID of the Talos image to upload.
	TalosImageID string
	// ARMServerSize is the server type to use for the image upload. The size muss match the architecture.
	ARMServerSize string
	// X86ServerSize is the server type to use for the image upload. The size muss match the architecture.
	X86ServerSize string
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
	var err error
	var arm *Image
	if args.EnableARMImageUpload {
		arm, err = NewImage(ctx, &ImageArgs{
			HetznerToken: args.HetznerToken,
			TalosVersion: args.TalosVersion,
			TalosImageID: args.TalosImageID,
			Arch:         ArchARM,
			ServerSize:   args.ARMServerSize,
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	var x86 *Image
	if args.EnableX86ImageUpload {
		x86, err = NewImage(ctx, &ImageArgs{
			HetznerToken: args.HetznerToken,
			TalosVersion: args.TalosVersion,
			TalosImageID: args.TalosImageID,
			Arch:         ArchX86,
			ServerSize:   args.X86ServerSize,
		}, opts...)
		if err != nil {
			return nil, err
		}
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

// Image represents the uploaded Talos image
// It contains the Hetzner Cloud image snapshot.
type Image struct {
	Snapshot *hcloudimages.UploadedImage
}

// NewImage uploads a Talos image to Hetzner Cloud using the hcloud-upload-image package.
func NewImage(ctx *pulumi.Context, args *ImageArgs, opts ...pulumi.ResourceOption) (*Image, error) {
	if args.Arch != ArchARM && args.Arch != ArchX86 {
		return nil, ErrUnknownArchitecture
	}

	// set arch for uploaded image
	var arch string
	switch args.Arch {
	case ArchARM:
		arch = "arm"
	case ArchX86:
		arch = "x86"
	default:
		return nil, ErrUnknownArchitecture
	}

	name := fmt.Sprintf("talos-%s-%s", arch, args.TalosVersion)

	snapshot, err := hcloudimages.NewUploadedImage(ctx, name, &hcloudimages.UploadedImageArgs{
		Description:      pulumi.Sprintf("%s - %s", name, time.Now().Format(time.RFC3339)),
		HcloudToken:      pulumi.String(args.HetznerToken),
		Architecture:     pulumi.String(arch),
		ImageUrl:         pulumi.Sprintf("https://factory.talos.dev/image/%s/%s/hcloud-%s.raw.xz", args.TalosImageID, args.TalosVersion, args.Arch),
		ImageCompression: pulumi.StringPtr("xz"),
		ServerType:       pulumi.String(args.ServerSize),
		Labels: pulumi.StringMap{
			"talos-version": pulumi.String(args.TalosVersion),
			"arch":          pulumi.String(string(args.Arch)),
			"stack":         pulumi.String(ctx.Stack()),
			"project":       pulumi.String(ctx.Project()),
		},
	}, append(opts,
		pulumi.IgnoreChanges([]string{"description"}))...,
	)
	if err != nil {
		return nil, err
	}

	return &Image{Snapshot: snapshot}, nil
}

func (i *Image) ImageId() pulumi.IntOutput {
	if i != nil {
		return i.Snapshot.ImageId
	}

	return pulumi.Int(0).ToIntOutput()
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

// DetectRequiredArchitecturesFromList determines which architectures are needed from a list of architectures
func DetectRequiredArchitecturesFromList(architectures []CPUArchitecture) (enableARM, enableX86 bool) {
	for _, arch := range architectures {
		switch arch {
		case ArchARM:
			enableARM = true
		case ArchX86:
			enableX86 = true
		}
	}
	return enableARM, enableX86
}
