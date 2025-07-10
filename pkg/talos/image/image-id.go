package image

const (
	// talosImageIDHetznerDefault is the default Talos image ID for Hetzner Cloud
	// This image is created by the Talos image factory.
	talosImageIDHetznerDefault = "376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba"
	// talosImageIDHetznerLonghorn is the default Talos image ID for Hetzner Cloud with Longhorn support
	// This image is created by the Talos image factory.
	// customization:
	// 	 systemExtensions:
	// 	 	 officialExtensions:
	// 			 - siderolabs/iscsi-tools
	// 			 - siderolabs/util-linux-tools
	talosImageLonghorn = "613e1592b2da41ae5e265e8789429f22e121aab91cb4deb6bc3c0b6262961245"
)

type TalosImageIDArgs struct {
	// OverwriteTalosImageID is the ID of the talos image factory to use
	OverwriteTalosImageID *string
	// EnableLonghornSupport is a flag to enable longhorn support for the cluster.
	// This will create a longhorn storage class and a longhorn CSI driver.
	EnableLonghornSupport bool
}

func NewTalosImageID(args *TalosImageIDArgs) string {
	if args.OverwriteTalosImageID != nil {
		return *args.OverwriteTalosImageID
	}

	if args.EnableLonghornSupport {
		return talosImageLonghorn
	}

	return talosImageIDHetznerDefault
}
