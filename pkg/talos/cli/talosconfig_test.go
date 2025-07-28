package cli

import (
	"testing"

	"github.com/exivity/pulumiconfig/pkg/pulumitest"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestNewTalosConfiguration(t *testing.T) {
	type args struct {
		args *TalosConfigurationArgs
	}
	tests := []struct {
		name string
		args args
		want pulumi.StringOutput
	}{
		{
			name: "basic config",
			args: args{
				args: &TalosConfigurationArgs{
					Context:           "test-cluster",
					Endpoints:         []pulumi.StringOutput{pulumi.String("1.2.3.4:6443").ToStringOutput()},
					Nodes:             []pulumi.StringOutput{pulumi.String("1.2.3.4").ToStringOutput()},
					CACertificate:     pulumi.String("ca-cert").ToStringOutput(),
					ClientCertificate: pulumi.String("client-cert").ToStringOutput(),
					ClientKey:         pulumi.String("client-key").ToStringOutput(),
				},
			},
			want: pulumi.Sprintf(`{"context": "%s", "contexts": {"%s": {"endpoints": %s, "nodes": %s, "ca": "%s", "crt": "%s", "key": "%s"}}}`,
				"test-cluster", "test-cluster", `["1.2.3.4:6443"]`, `["1.2.3.4"]`, "ca-cert", "client-cert", "client-key"),
		},
		{
			name: "different context",
			args: args{
				args: &TalosConfigurationArgs{
					Context:           "prod-cluster",
					Endpoints:         []pulumi.StringOutput{pulumi.String("10.0.0.1:6443").ToStringOutput()},
					Nodes:             []pulumi.StringOutput{pulumi.String("10.0.0.1").ToStringOutput()},
					CACertificate:     pulumi.String("prod-ca").ToStringOutput(),
					ClientCertificate: pulumi.String("prod-client-cert").ToStringOutput(),
					ClientKey:         pulumi.String("prod-client-key").ToStringOutput(),
				},
			},
			want: pulumi.Sprintf(`{"context": "%s", "contexts": {"%s": {"endpoints": %s, "nodes": %s, "ca": "%s", "crt": "%s", "key": "%s"}}}`,
				"prod-cluster", "prod-cluster", `["10.0.0.1:6443"]`, `["10.0.0.1"]`, "prod-ca", "prod-client-cert", "prod-client-key"),
		},
		{
			name: "multiple endpoints and nodes",
			args: args{
				args: &TalosConfigurationArgs{
					Context: "multi-cluster",
					Endpoints: []pulumi.StringOutput{
						pulumi.String("1.2.3.4:6443").ToStringOutput(),
						pulumi.String("1.2.3.5:6443").ToStringOutput(),
					},
					Nodes: []pulumi.StringOutput{
						pulumi.String("1.2.3.4").ToStringOutput(),
						pulumi.String("1.2.3.5").ToStringOutput(),
						pulumi.String("1.2.3.6").ToStringOutput(),
					},
					CACertificate:     pulumi.String("multi-ca").ToStringOutput(),
					ClientCertificate: pulumi.String("multi-client-cert").ToStringOutput(),
					ClientKey:         pulumi.String("multi-client-key").ToStringOutput(),
				},
			},
			want: pulumi.Sprintf(`{"context": "%s", "contexts": {"%s": {"endpoints": %s, "nodes": %s, "ca": "%s", "crt": "%s", "key": "%s"}}}`,
				"multi-cluster", "multi-cluster", `["1.2.3.4:6443", "1.2.3.5:6443"]`, `["1.2.3.4", "1.2.3.5", "1.2.3.6"]`, "multi-ca", "multi-client-cert", "multi-client-key"),
		},
		{
			name: "empty endpoints and nodes",
			args: args{
				args: &TalosConfigurationArgs{
					Context:           "empty-cluster",
					Endpoints:         []pulumi.StringOutput{},
					Nodes:             []pulumi.StringOutput{},
					CACertificate:     pulumi.String("empty-ca").ToStringOutput(),
					ClientCertificate: pulumi.String("empty-client-cert").ToStringOutput(),
					ClientKey:         pulumi.String("empty-client-key").ToStringOutput(),
				},
			},
			want: pulumi.Sprintf(`{"context": "%s", "contexts": {"%s": {"endpoints": %s, "nodes": %s, "ca": "%s", "crt": "%s", "key": "%s"}}}`,
				"empty-cluster", "empty-cluster", `[]`, `[]`, "empty-ca", "empty-client-cert", "empty-client-key"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTalosConfiguration(tt.args.args)
			pulumitest.AssertStringOutputEqual(t, got, tt.want)
		})
	}
}
