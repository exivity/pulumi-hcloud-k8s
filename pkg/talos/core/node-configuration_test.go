package core

import (
	"reflect"
	"strings"
	"testing"

	core_config "github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config/core"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config/registry"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config/volume"
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

func Test_newRegistryConfigs(t *testing.T) {
	tests := []struct {
		name    string
		args    *core_config.RegistriesConfig
		wantLen int
		wantErr bool
		verify  func(t *testing.T, configs []string)
	}{
		{
			name:    "nil input",
			args:    nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "empty input",
			args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{},
				Config:  map[string]core_config.RegistryConfig{},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "single mirror",
			args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{
					"docker.io": {
						Endpoints: []core_config.RegistryMirrorEndpoint{
							{URL: "https://mirror.gcr.io", OverridePath: true},
						},
						SkipFallback: false,
					},
				},
				Config: map[string]core_config.RegistryConfig{},
			},
			wantLen: 1,
			wantErr: false,
			verify: func(t *testing.T, configs []string) {
				var mirrorCfg registry.RegistryMirrorConfig
				err := yaml.Unmarshal([]byte(configs[0]), &mirrorCfg)
				assert.NoError(t, err)
				assert.Equal(t, "RegistryMirrorConfig", mirrorCfg.Kind)
				assert.Equal(t, "docker.io", mirrorCfg.Name)
				assert.Len(t, mirrorCfg.Endpoints, 1)
				assert.Equal(t, "https://mirror.gcr.io", mirrorCfg.Endpoints[0].URL)
				assert.True(t, mirrorCfg.Endpoints[0].OverridePath)
			},
		},
		{
			name: "single registry config with TLS and Auth",
			args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{},
				Config: map[string]core_config.RegistryConfig{
					"quay.io": {
						TLS: &core_config.RegistryTLSConfig{
							ClientIdentity: &core_config.PEMEncodedCertificateAndKey{
								Cert: "crtdata",
								Key:  "keydata",
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
			},
			wantLen: 2, // TLS + Auth
			wantErr: false,
			verify: func(t *testing.T, configs []string) {
				// First config should be TLS
				var tlsCfg registry.RegistryTLSConfig
				err := yaml.Unmarshal([]byte(configs[0]), &tlsCfg)
				assert.NoError(t, err)
				assert.Equal(t, "RegistryTLSConfig", tlsCfg.Kind)
				assert.Equal(t, "quay.io", tlsCfg.Name)
				assert.Equal(t, "ca-data", tlsCfg.CA)
				assert.True(t, tlsCfg.InsecureSkipVerify)
				assert.Equal(t, "crtdata", tlsCfg.ClientIdentity.Cert)
				assert.Equal(t, "keydata", tlsCfg.ClientIdentity.Key)

				// Second config should be Auth
				var authCfg registry.RegistryAuthConfig
				err = yaml.Unmarshal([]byte(configs[1]), &authCfg)
				assert.NoError(t, err)
				assert.Equal(t, "RegistryAuthConfig", authCfg.Kind)
				assert.Equal(t, "quay.io", authCfg.Name)
				assert.Equal(t, "user", authCfg.Username)
				assert.Equal(t, "pass", authCfg.Password)
			},
		},
		{
			name: "multiple mirrors and registries",
			args: &core_config.RegistriesConfig{
				Mirrors: map[string]core_config.RegistryMirrorConfig{
					"docker.io": {
						Endpoints: []core_config.RegistryMirrorEndpoint{
							{URL: "https://mirror1"},
							{URL: "https://mirror2"},
						},
						SkipFallback: true,
					},
					"gcr.io": {
						Endpoints: []core_config.RegistryMirrorEndpoint{
							{URL: "https://gcr-mirror", OverridePath: true},
						},
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
			},
			wantLen: 4, // 2 mirrors + 1 auth (docker.io) + 1 TLS (gcr.io)
			wantErr: false,
			verify: func(t *testing.T, configs []string) {
				// Mirrors come first (sorted by key), then TLS/Auth configs (sorted by key)
				// docker.io mirror
				assert.True(t, strings.Contains(configs[0], "docker.io"))
				assert.True(t, strings.Contains(configs[0], "RegistryMirrorConfig"))
				// gcr.io mirror
				assert.True(t, strings.Contains(configs[1], "gcr.io"))
				assert.True(t, strings.Contains(configs[1], "RegistryMirrorConfig"))
				// docker.io auth
				assert.True(t, strings.Contains(configs[2], "docker.io"))
				assert.True(t, strings.Contains(configs[2], "RegistryAuthConfig"))
				// gcr.io TLS
				assert.True(t, strings.Contains(configs[3], "gcr.io"))
				assert.True(t, strings.Contains(configs[3], "RegistryTLSConfig"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newRegistryConfigs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("newRegistryConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Len(t, got, tt.wantLen)
			if tt.verify != nil {
				tt.verify(t, got)
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
		want []core.ClusterInlineManifest
	}{
		{
			name: "empty manifests",
			args: args{
				manifests: []core_config.ClusterInlineManifest{},
			},
			want: []core.ClusterInlineManifest{},
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
			want: []core.ClusterInlineManifest{
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
			want: []core.ClusterInlineManifest{
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
			want: []core.ClusterInlineManifest{
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
			want: []core.ClusterInlineManifest{
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
		want *core.CNIConfig
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
			want: &core.CNIConfig{
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
			want: &core.CNIConfig{
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
		want   *core.TalosConfig
		verify func(t *testing.T, cfg *core.TalosConfig)
	}{
		{
			name: "basic controlplane",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.ControlPlaneNode,
				Subnet:         "10.0.0.0/24",
				PodSubnets:     "10.244.0.0/16",
			},
			verify: func(t *testing.T, cfg *core.TalosConfig) {
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
			verify: func(t *testing.T, cfg *core.TalosConfig) {
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
			verify: func(t *testing.T, cfg *core.TalosConfig) {
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
			verify: func(t *testing.T, cfg *core.TalosConfig) {
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
			verify: func(t *testing.T, cfg *core.TalosConfig) {
				assert.Equal(t, "1024", cfg.Machine.Sysctls["vm.nr_hugepages"])
				assert.Contains(t, cfg.Machine.Kernel.Modules, core.KernelModuleConfig{Name: "nvme_tcp"})
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
			verify: func(t *testing.T, cfg *core.TalosConfig) {
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
			verify: func(t *testing.T, cfg *core.TalosConfig) {
				assert.Equal(t, "key=val:NoSchedule", cfg.Machine.Kubelet.ExtraArgs["register-with-taints"])
			},
		},
		{
			name: "registries not in main config",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.WorkerNode,
				Subnet:         "10.0.0.0/24",
				PodSubnets:     "10.244.0.0/16",
				Registries: &core_config.RegistriesConfig{
					Mirrors: map[string]core_config.RegistryMirrorConfig{
						"docker.io": {Endpoints: []core_config.RegistryMirrorEndpoint{
							{URL: "https://mirror.gcr.io"},
						}},
					},
				},
			},
			verify: func(t *testing.T, cfg *core.TalosConfig) {
				// Registries should no longer be embedded in the main TalosConfig;
				// they are now emitted as separate config documents.
				assert.Nil(t, cfg.Machine.Registries)
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
		verify  func(t *testing.T, configs []string)
	}{
		{
			name: "returns all config files",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.ControlPlaneNode,
				Subnet:         "10.0.0.0/24",
				PodSubnets:     "10.244.0.0/16",
			},
			wantLen: 1,
			wantErr: false,
			verify: func(t *testing.T, configs []string) {
				var talosConfig core.TalosConfig
				err := yaml.Unmarshal([]byte(configs[0]), &talosConfig)
				assert.NoError(t, err, "failed to unmarshal TalosConfig")
				assert.Equal(t, "controlplane", talosConfig.Machine.Type)
			},
		},
		{
			name: "with disk encryption",
			args: &NodeConfigurationArgs{
				ServerNodeType: meta.ControlPlaneNode,
				Subnet:         "10.0.0.0/24",
				PodSubnets:     "10.244.0.0/16",
				DiskEncryption: &core_config.DiskEncryptionConfig{
					EncryptState:     true,
					EncryptEphemeral: true,
					Keys: []core_config.EncryptionKeyConfig{
						{
							Slot:   0,
							NodeID: &core_config.EncryptionKeyNodeID{},
						},
					},
				},
			},
			wantLen: 3,
			wantErr: false,
			verify: func(t *testing.T, configs []string) {
				// Verify STATE volume config
				var stateConfig volume.VolumeConfig
				err := yaml.Unmarshal([]byte(configs[1]), &stateConfig)
				assert.NoError(t, err)
				assert.Equal(t, "STATE", stateConfig.Name)
				assert.Equal(t, "luks2", stateConfig.Encryption.Provider)
				assert.Len(t, stateConfig.Encryption.Keys, 1)
				assert.NotNil(t, stateConfig.Encryption.Keys[0].NodeID)

				// Verify EPHEMERAL volume config
				var ephemeralConfig volume.VolumeConfig
				err = yaml.Unmarshal([]byte(configs[2]), &ephemeralConfig)
				assert.NoError(t, err)
				assert.Equal(t, "EPHEMERAL", ephemeralConfig.Name)
				assert.Equal(t, "luks2", ephemeralConfig.Encryption.Provider)
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
						"docker.io": {
							Endpoints: []core_config.RegistryMirrorEndpoint{
								{URL: "https://mirror.gcr.io"},
							},
						},
					},
					Config: map[string]core_config.RegistryConfig{
						"mirror.gcr.io": {
							Auth: &core_config.RegistryAuthConfig{
								Username: "user",
								Password: "pass",
							},
						},
					},
				},
			},
			wantLen: 3, // 1 main + 1 mirror + 1 auth
			wantErr: false,
			verify: func(t *testing.T, configs []string) {
				// First config is the main Talos config
				var talosConfig core.TalosConfig
				err := yaml.Unmarshal([]byte(configs[0]), &talosConfig)
				assert.NoError(t, err)
				assert.Nil(t, talosConfig.Machine.Registries)

				// Second config is the mirror
				var mirrorCfg registry.RegistryMirrorConfig
				err = yaml.Unmarshal([]byte(configs[1]), &mirrorCfg)
				assert.NoError(t, err)
				assert.Equal(t, "RegistryMirrorConfig", mirrorCfg.Kind)
				assert.Equal(t, "docker.io", mirrorCfg.Name)

				// Third config is the auth
				var authCfg registry.RegistryAuthConfig
				err = yaml.Unmarshal([]byte(configs[2]), &authCfg)
				assert.NoError(t, err)
				assert.Equal(t, "RegistryAuthConfig", authCfg.Kind)
				assert.Equal(t, "mirror.gcr.io", authCfg.Name)
			},
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
				if tt.verify != nil {
					tt.verify(t, got)
				}
			}
		})
	}
}
