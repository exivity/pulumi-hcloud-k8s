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
					Endpoint:          pulumi.String("1.2.3.4:6443").ToStringOutput(),
					CACertificate:     pulumi.String("ca-cert").ToStringOutput(),
					ClientCertificate: pulumi.String("client-cert").ToStringOutput(),
					ClientKey:         pulumi.String("client-key").ToStringOutput(),
				},
			},
			want: pulumi.Sprintf(`{"context": "%s", "contexts": {"%s": {"endpoints": ["%s"], "ca": "%s", "crt": "%s", "key": "%s"}}}`,
				"test-cluster", "test-cluster", "1.2.3.4:6443", "ca-cert", "client-cert", "client-key"),
		},
		{
			name: "different context",
			args: args{
				args: &TalosConfigurationArgs{
					Context:           "prod-cluster",
					Endpoint:          pulumi.String("10.0.0.1:6443").ToStringOutput(),
					CACertificate:     pulumi.String("prod-ca").ToStringOutput(),
					ClientCertificate: pulumi.String("prod-client-cert").ToStringOutput(),
					ClientKey:         pulumi.String("prod-client-key").ToStringOutput(),
				},
			},
			want: pulumi.Sprintf(`{"context": "%s", "contexts": {"%s": {"endpoints": ["%s"], "ca": "%s", "crt": "%s", "key": "%s"}}}`,
				"prod-cluster", "prod-cluster", "10.0.0.1:6443", "prod-ca", "prod-client-cert", "prod-client-key"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTalosConfiguration(tt.args.args)
			pulumitest.AssertStringOutputEqual(t, got, tt.want)
		})
	}
}
