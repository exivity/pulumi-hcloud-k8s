package cli

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TalosConfigurationArgs struct {
	Context string
	// LoadBalancer Endpoint
	Endpoint          pulumi.StringOutput
	CACertificate     pulumi.StringOutput
	ClientCertificate pulumi.StringOutput
	ClientKey         pulumi.StringOutput
}

// NewTalosConfiguration generates configuration for talosctl.
func NewTalosConfiguration(args *TalosConfigurationArgs) pulumi.StringOutput {
	return pulumi.Sprintf(`{"context": "%s", "contexts": {"%s": {"endpoints": ["%s"], "ca": "%s", "crt": "%s", "key": "%s"}}}`,
		args.Context, args.Context, args.Endpoint, args.CACertificate, args.ClientCertificate, args.ClientKey,
	)
}
