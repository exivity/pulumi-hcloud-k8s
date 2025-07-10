package core

import (
	"reflect"
	"testing"

	core_config "github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config"
)

func Test_toTalosTaints(t *testing.T) {
	type args struct {
		taints []core_config.Taint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty taints",
			args: args{
				taints: []core_config.Taint{},
			},
			want: "",
		},
		{
			name: "single taint",
			args: args{
				taints: []core_config.Taint{
					{
						Key:    "key1",
						Value:  "value1",
						Effect: "NoSchedule",
					},
				},
			},
			want: "key1=value1:NoSchedule",
		},
		{
			name: "multiple taints",
			args: args{
				taints: []core_config.Taint{
					{
						Key:    "key1",
						Value:  "value1",
						Effect: "NoSchedule",
					},
					{
						Key:    "key2",
						Value:  "value2",
						Effect: "NoExecute",
					},
				},
			},
			want: "key1=value1:NoSchedule,key2=value2:NoExecute",
		},
		{
			name: "taints with empty value",
			args: args{
				taints: []core_config.Taint{
					{
						Key:    "node-role",
						Value:  "",
						Effect: "NoSchedule",
					},
				},
			},
			want: "node-role=:NoSchedule",
		},
		{
			name: "complex taint combinations",
			args: args{
				taints: []core_config.Taint{
					{
						Key:    "special-workloads",
						Value:  "true",
						Effect: "PreferNoSchedule",
					},
					{
						Key:    "node-role.kubernetes.io/master",
						Value:  "",
						Effect: "NoSchedule",
					},
					{
						Key:    "dedicated",
						Value:  "monitoring",
						Effect: "NoExecute",
					},
				},
			},
			want: "special-workloads=true:PreferNoSchedule,node-role.kubernetes.io/master=:NoSchedule,dedicated=monitoring:NoExecute",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toTalosTaints(tt.args.taints); got != tt.want {
				t.Errorf("toTalosTaints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toRegistriesConfig(t *testing.T) {
	type args struct {
		args *core_config.RegistriesConfig
	}
	tests := []struct {
		name string
		args args
		want *config.RegistriesConfig
	}{
		{
			name: "nil input",
			args: args{args: nil},
			want: &config.RegistriesConfig{
				Mirrors: map[string]config.RegistryMirrorConfig{},
				Config:  map[string]config.RegistryConfig{},
			},
		},
		{
			name: "empty input",
			args: args{args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{},
				Config:  map[string]core_config.RegistryConfig{},
			}},
			want: &config.RegistriesConfig{
				Mirrors: map[string]config.RegistryMirrorConfig{},
				Config:  map[string]config.RegistryConfig{},
			},
		},
		{
			name: "single mirror",
			args: args{args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{
					"docker.io": {
						Endpoints:    []string{"https://mirror.gcr.io"},
						OverridePath: true,
						SkipFallback: false,
					},
				},
				Config: map[string]core_config.RegistryConfig{},
			}},
			want: &config.RegistriesConfig{
				Mirrors: map[string]config.RegistryMirrorConfig{
					"docker.io": {
						Endpoints:    []string{"https://mirror.gcr.io"},
						OverridePath: true,
						SkipFallback: false,
					},
				},
				Config: map[string]config.RegistryConfig{},
			},
		},
		{
			name: "single registry config with TLS and Auth",
			args: args{args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{},
				Config: map[string]core_config.RegistryConfig{
					"quay.io": {
						TLS: &core_config.RegistryTLSConfig{
							ClientIdentity: &core_config.PEMEncodedCertificateAndKey{
								CRT: "crtdata",
								Key: "keydata",
							},
							CA:                 "ca-data",
							InsecureSkipVerify: true,
						},
						Auth: &core_config.RegistryAuthConfig{
							Username:      "user",
							Password:      "pass",
							Auth:          "authstr",
							IdentityToken: "token",
						},
					},
				},
			}},
			want: &config.RegistriesConfig{
				Mirrors: map[string]config.RegistryMirrorConfig{},
				Config: map[string]config.RegistryConfig{
					"quay.io": {
						TLS: &config.RegistryTLSConfig{
							ClientIdentity: &config.PEMEncodedCertificateAndKey{
								CRT: "crtdata",
								Key: "keydata",
							},
							CA:                 "ca-data",
							InsecureSkipVerify: true,
						},
						Auth: &config.RegistryAuthConfig{
							Username:      "user",
							Password:      "pass",
							Auth:          "authstr",
							IdentityToken: "token",
						},
					},
				},
			},
		},
		{
			name: "multiple mirrors and registries",
			args: args{args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{
					"docker.io": {
						Endpoints:    []string{"https://mirror1", "https://mirror2"},
						OverridePath: false,
						SkipFallback: true,
					},
					"gcr.io": {
						Endpoints:    []string{"https://gcr-mirror"},
						OverridePath: true,
						SkipFallback: false,
					},
				},
				Config: map[string]core_config.RegistryConfig{
					"docker.io": {
						TLS: nil,
						Auth: &core_config.RegistryAuthConfig{
							Username:      "dockeruser",
							Password:      "dockerpass",
							Auth:          "docker-auth",
							IdentityToken: "docker-token",
						},
					},
					"gcr.io": {
						TLS: &core_config.RegistryTLSConfig{
							ClientIdentity:     nil,
							CA:                 "gcr-ca",
							InsecureSkipVerify: false,
						},
						Auth: nil,
					},
				},
			}},
			want: &config.RegistriesConfig{
				Mirrors: map[string]config.RegistryMirrorConfig{
					"docker.io": {
						Endpoints:    []string{"https://mirror1", "https://mirror2"},
						OverridePath: false,
						SkipFallback: true,
					},
					"gcr.io": {
						Endpoints:    []string{"https://gcr-mirror"},
						OverridePath: true,
						SkipFallback: false,
					},
				},
				Config: map[string]config.RegistryConfig{
					"docker.io": {
						TLS: nil,
						Auth: &config.RegistryAuthConfig{
							Username:      "dockeruser",
							Password:      "dockerpass",
							Auth:          "docker-auth",
							IdentityToken: "docker-token",
						},
					},
					"gcr.io": {
						TLS: &config.RegistryTLSConfig{
							ClientIdentity:     nil,
							CA:                 "gcr-ca",
							InsecureSkipVerify: false,
						},
						Auth: nil,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toRegistriesConfig(tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toRegistriesConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
