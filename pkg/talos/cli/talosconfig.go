package cli

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TalosConfigurationArgs struct {
	Context           string
	Endpoints         []pulumi.StringOutput
	Nodes             []pulumi.StringOutput
	CACertificate     pulumi.StringOutput
	ClientCertificate pulumi.StringOutput
	ClientKey         pulumi.StringOutput
}

// NewTalosConfiguration generates configuration for talosctl.
func NewTalosConfiguration(args *TalosConfigurationArgs) pulumi.StringOutput {
	return pulumi.Sprintf(`{"context": "%s", "contexts": {"%s": {"endpoints": %s, "nodes": %s, "ca": "%s", "crt": "%s", "key": "%s"}}}`,
		args.Context, args.Context, marshalJSONList(args.Endpoints), marshalJSONList(args.Nodes), args.CACertificate, args.ClientCertificate, args.ClientKey,
	)
}

func marshalJSONList(items []pulumi.StringOutput) pulumi.StringOutput {
	if len(items) == 0 {
		return pulumi.String("[]").ToStringOutput()
	}

	jsonList := pulumi.String("[").ToStringOutput()
	for i, item := range items {
		jsonList = pulumi.Sprintf(`%s"%s"`, jsonList, item)
		if i < len(items)-1 {
			jsonList = pulumi.Sprintf("%s, ", jsonList)
		}
	}
	jsonList = pulumi.Sprintf("%s]", jsonList)

	return jsonList
}
