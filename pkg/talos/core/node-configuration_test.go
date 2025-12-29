package core

import (
	"reflect"
	"testing"

	core_config "github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func stringPtr(s string) *string {
	return &s
}

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
			want: nil,
		},
		{
			name: "empty input",
			args: args{args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{},
				Config:  map[string]core_config.RegistryConfig{},
			}},
			want: nil,
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

func Test_toInlineManifests(t *testing.T) {
	type args struct {
		manifests []core_config.ClusterInlineManifest
	}
	tests := []struct {
		name string
		args args
		want []config.ClusterInlineManifest
	}{
		{
			name: "empty manifests",
			args: args{
				manifests: []core_config.ClusterInlineManifest{},
			},
			want: []config.ClusterInlineManifest{},
		},
		{
			name: "single manifest",
			args: args{
				manifests: []core_config.ClusterInlineManifest{
					{
						Name: "namespace-ci",
						Contents: `apiVersion: v1
kind: Namespace
metadata:
  name: ci`,
					},
				},
			},
			want: []config.ClusterInlineManifest{
				{
					Name: "namespace-ci",
					Contents: `apiVersion: v1
kind: Namespace
metadata:
  name: ci`,
				},
			},
		},
		{
			name: "multiple manifests",
			args: args{
				manifests: []core_config.ClusterInlineManifest{
					{
						Name: "namespace-prod",
						Contents: `apiVersion: v1
kind: Namespace
metadata:
  name: production`,
					},
					{
						Name: "configmap-app",
						Contents: `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  config.yaml: |
    app:
      name: myapp
      port: 8080`,
					},
				},
			},
			want: []config.ClusterInlineManifest{
				{
					Name: "namespace-prod",
					Contents: `apiVersion: v1
kind: Namespace
metadata:
  name: production`,
				},
				{
					Name: "configmap-app",
					Contents: `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  config.yaml: |
    app:
      name: myapp
      port: 8080`,
				},
			},
		},
		{
			name: "manifest with complex yaml",
			args: args{
				manifests: []core_config.ClusterInlineManifest{
					{
						Name: "deployment-app",
						Contents: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
        env:
        - name: ENV_VAR
          value: "production"`,
					},
				},
			},
			want: []config.ClusterInlineManifest{
				{
					Name: "deployment-app",
					Contents: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
        env:
        - name: ENV_VAR
          value: "production"`,
				},
			},
		},
		{
			name: "manifest with special characters",
			args: args{
				manifests: []core_config.ClusterInlineManifest{
					{
						Name: "secret-config",
						Contents: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm`,
					},
				},
			},
			want: []config.ClusterInlineManifest{
				{
					Name: "secret-config",
					Contents: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toInlineManifests(tt.args.manifests); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toInlineManifests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toCNIConfig(t *testing.T) {
	tests := []struct {
		name string
		cni  *core_config.CNIConfig
		want *config.CNIConfig
	}{
		{
			name: "nil input",
			cni:  nil,
			want: nil,
		},
		{
			name: "valid config",
			cni: &core_config.CNIConfig{
				Name: "custom-cni",
				URLs: []string{"https://example.com/cni.yaml"},
			},
			want: &config.CNIConfig{
				Name: "custom-cni",
				URLs: []string{"https://example.com/cni.yaml"},
			},
		},
		{
			name: "empty urls",
			cni: &core_config.CNIConfig{
				Name: "none",
				URLs: []string{},
			},
			want: &config.CNIConfig{
				Name: "none",
				URLs: []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toCNIConfig(tt.cni); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCNIConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newMainTalosConfig(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		args   *NodeConfigurationArgs
		want   *config.TalosConfig
		verify func(t *testing.T, cfg *config.TalosConfig)
	}{
		{
			name: "basic controlplane",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.ControlPlaneNode,
				Subnet:         "10.0.0.0/24",
				PodSubnets:     "10.244.0.0/16",
			},
			verify: func(t *testing.T, cfg *config.TalosConfig) {
				assert.Equal(t, "controlplane", cfg.Machine.Type)
				assert.Equal(t, []string{"10.0.0.0/24"}, cfg.Machine.Kubelet.NodeIP.ValidSubnets)
				assert.Equal(t, []string{"10.244.0.0/16"}, cfg.Cluster.Network.PodSubnets)
				assert.True(t, cfg.Cluster.ExternalCloudProvider.Enabled)
				assert.Empty(t, cfg.Cluster.ExternalCloudProvider.Manifests)
			},
		},
		{
			name: "basic worker",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.WorkerNode,
				Subnet:         "10.0.1.0/24",
				PodSubnets:     "10.244.0.0/16",
			},
			verify: func(t *testing.T, cfg *config.TalosConfig) {
				assert.Equal(t, "worker", cfg.Machine.Type)
			},
		},
		{
			name: "with optional fields",
			args: &NodeConfigurationArgs{
				ServerNodeType:                 meta.ControlPlaneNode,
				Subnet:                         "10.0.0.0/24",
				PodSubnets:                     "10.244.0.0/16",
				DNSDomain:                      stringPtr("example.com"),
				ServiceSubnet:                  stringPtr("10.96.0.0/12"),
				CertLifetime:                   stringPtr("8760h"),
				SecretboxEncryptionSecret:      stringPtr("secret-key"),
				AllowSchedulingOnControlPlanes: true,
			},
			verify: func(t *testing.T, cfg *config.TalosConfig) {
				assert.Equal(t, "example.com", cfg.Cluster.Network.DNSDomain)
				assert.Equal(t, []string{"10.96.0.0/12"}, cfg.Cluster.Network.ServiceSubnets)
				assert.Equal(t, "8760h", cfg.Cluster.AdminKubeconfig.CertLifetime)
				// "secret-key" base64 encoded is "c2VjcmV0LWtleQ=="
				assert.Equal(t, "c2VjcmV0LWtleQ==", cfg.Cluster.SecretboxEncryptionSecret)
				assert.True(t, cfg.Cluster.AllowSchedulingOnControlPlanes)
			},
		},
		{
			name: "with extra manifests",
			args: &NodeConfigurationArgs{
				ServerNodeType:                meta.ControlPlaneNode,
				Subnet:                        "10.0.0.0/24",
				PodSubnets:                    "10.244.0.0/16",
				ExtraManifests:                []string{"https://example.com/manifest.yaml"},
				EnableHetznerCCMExtraManifest: true,
			},
			verify: func(t *testing.T, cfg *config.TalosConfig) {
				assert.Contains(t, cfg.Cluster.ExtraManifests, "https://example.com/manifest.yaml")
				assert.Len(t, cfg.Cluster.ExternalCloudProvider.Manifests, 2)
				assert.Contains(t, cfg.Cluster.ExternalCloudProvider.Manifests[0], "ccm-networks.yaml")
			},
		},
		{
			name: "with longhorn support",
			args: &NodeConfigurationArgs{
				ServerNodeType:        meta.WorkerNode,
				Subnet:                "10.0.0.0/24",
				PodSubnets:            "10.244.0.0/16",
				EnableLonghornSupport: true,
			},
			verify: func(t *testing.T, cfg *config.TalosConfig) {
				assert.Equal(t, "1024", cfg.Machine.Sysctls["vm.nr_hugepages"])
				assert.Contains(t, cfg.Machine.Kernel.Modules, config.KernelModuleConfig{Name: "nvme_tcp"})
				found := false
				for _, mount := range cfg.Machine.Kubelet.ExtraMounts {
					if mount.Destination == "/var/lib/longhorn" {
						found = true
						break
					}
				}
				assert.True(t, found, "longhorn mount not found")
			},
		},
		{
			name: "with local storage",
			args: &NodeConfigurationArgs{
				ServerNodeType:      meta.WorkerNode,
				Subnet:              "10.0.0.0/24",
				PodSubnets:          "10.244.0.0/16",
				LocalStorageFolders: []string{"/data/local"},
			},
			verify: func(t *testing.T, cfg *config.TalosConfig) {
				found := false
				for _, mount := range cfg.Machine.Kubelet.ExtraMounts {
					if mount.Destination == "/data/local" {
						found = true
						break
					}
				}
				assert.True(t, found, "local storage mount not found")
			},
		},
		{
			name: "with taints",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.WorkerNode,
				Subnet:         "10.0.0.0/24",
				PodSubnets:     "10.244.0.0/16",
				NodeTaints: []core_config.Taint{
					{Key: "key", Value: "val", Effect: "NoSchedule"},
				},
			},
			verify: func(t *testing.T, cfg *config.TalosConfig) {
				assert.Equal(t, "key=val:NoSchedule", cfg.Machine.Kubelet.ExtraArgs["register-with-taints"])
			},
		},
		{
			name: "with registries",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.WorkerNode,
				Subnet:         "10.0.0.0/24",
				PodSubnets:     "10.244.0.0/16",
				Registries: &core_config.RegistriesConfig{
					Mirrors: map[string]core_config.RegistryMirrorConfig{
						"docker.io": {Endpoints: []string{"https://mirror.gcr.io"}},
					},
				},
			},
			verify: func(t *testing.T, cfg *config.TalosConfig) {
				assert.NotNil(t, cfg.Machine.Registries)
				assert.Contains(t, cfg.Machine.Registries.Mirrors, "docker.io")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newMainTalosConfig(tt.args)
			if tt.verify != nil {
				tt.verify(t, got)
			}
		})
	}
}

func TestNewNodeConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		args    *NodeConfigurationArgs
		wantLen int
		wantErr bool
	}{
		{
			name: "returns all config files",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.ControlPlaneNode,
				Subnet:         "10.0.0.0/24",
				PodSubnets:     "10.244.0.0/16",
			},
			wantLen: 3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewNodeConfiguration(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNodeConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Len(t, got, tt.wantLen)

				// Verify TalosConfig
				var talosConfig config.TalosConfig
				err = yaml.Unmarshal([]byte(got[0]), &talosConfig)
				assert.NoError(t, err, "failed to unmarshal TalosConfig")
				assert.Equal(t, "controlplane", talosConfig.Machine.Type)

				// Verify VolumeConfig
				var volumeConfig config.VolumeConfig
				err = yaml.Unmarshal([]byte(got[1]), &volumeConfig)
				assert.NoError(t, err, "failed to unmarshal VolumeConfig")

				// Verify UserVolumeConfig
				var userVolumeConfig config.UserVolumeConfig
				err = yaml.Unmarshal([]byte(got[2]), &userVolumeConfig)
				assert.NoError(t, err, "failed to unmarshal UserVolumeConfig")
			}
		})
	}
}
