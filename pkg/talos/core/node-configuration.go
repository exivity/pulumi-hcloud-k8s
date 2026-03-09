package core

import (
	"encoding/base64"
	"fmt"
	"sort"

	core_config "github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config/core"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config/registry"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config/volume"
)

type NodeConfigurationArgs struct {
	// ServerNodeType is the type of the server node
	ServerNodeType meta.ServerNodeType
	// DNSDomain is the DNS domain for the cluster, defaults to "cluster.local" if not provided
	DNSDomain *string
	// Subnet is the subnet for the cluster
	Subnet string
	// PodSubnets is the pod subnets for the cluster
	PodSubnets string
	// ServiceSubnet is the service subnets for the cluster
	ServiceSubnet *string
	// NodeLabels is the labels for the node
	NodeLabels map[string]string
	// NodeAnnotations is the annotations for the node
	NodeAnnotations map[string]string
	// NodeTaints is the taints for the node
	NodeTaints []core_config.Taint
	// EnableLonghornSupport is true if longhorn support is enabled
	EnableLonghornSupport bool
	// SecretboxEncryptionSecret is a base64-encoded 32-byte key used to encrypt Kubernetes secrets at rest in etcd
	SecretboxEncryptionSecret *string
	// CertLifetime is the admin kubeconfig certificate lifetime
	CertLifetime *string
	// AllowSchedulingOnControlPlanes is true if scheduling on control planes is allowed
	AllowSchedulingOnControlPlanes bool
	// Nameservers is the list of DNS servers to use for the cluster
	// If not provided, defaults are used
	Nameservers []string
	// LocalStorageFolders is a list of folders to make accessible for local storage
	LocalStorageFolders []string
	// Registries is the registries configuration for the Talos image
	Registries *core_config.RegistriesConfig
	// ExtraManifests is a list of URLs that point to additional manifests
	// These will get automatically deployed as part of the bootstrap
	ExtraManifests []string
	// ExtraManifestHeaders is a map of key value pairs that will be added while fetching the ExtraManifests
	ExtraManifestHeaders map[string]string
	// InlineManifests is a list of inline Kubernetes manifests
	// These will get automatically deployed as part of the bootstrap
	InlineManifests []core_config.ClusterInlineManifest
	// EnableHetznerCCMExtraManifest enables installation of Hetzner CCM via Talos extra manifests
	EnableHetznerCCMExtraManifest bool
	// EnableKubeSpan can be used to encrypt the traffic with wireguard. This works well with flannel, but it is recommended to disable when using a CNI like Cilium.
	EnableKubeSpan bool
	// CNI is the CNI configuration for the cluster.
	CNI *core_config.CNIConfig
	// DiskEncryption configures disk encryption for system partitions.
	DiskEncryption *core_config.DiskEncryptionConfig
}

func NewNodeConfiguration(args *NodeConfigurationArgs) ([]string, error) {
	nodeConfig := newMainTalosConfig(args)
	nodeConfigYAML, err := nodeConfig.YAML()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Talos config YAML: %w", err)
	}

	configs := []string{nodeConfigYAML}

	volumeConfigs := newVolumeConfigs(args)

	for _, vc := range volumeConfigs {
		vcYAML, err := vc.YAML()
		if err != nil {
			return nil, fmt.Errorf("failed to generate Volume config YAML: %w", err)
		}
		configs = append(configs, vcYAML)
	}

	registryConfigs, err := newRegistryConfigs(args.Registries)
	if err != nil {
		return nil, err
	}
	configs = append(configs, registryConfigs...)

	return configs, nil
}

func newMainTalosConfig(args *NodeConfigurationArgs) *core.TalosConfig { //nolint:funlen // lengthy function due to config mapping
	var adminKubeconfig *core.AdminKubeconfigConfig
	if args.CertLifetime != nil {
		adminKubeconfig = &core.AdminKubeconfigConfig{
			CertLifetime: *args.CertLifetime,
		}
	}

	ccmExtraManifests := []string{}
	if args.EnableHetznerCCMExtraManifest {
		ccmExtraManifests = []string{
			"https://raw.githubusercontent.com/hetznercloud/hcloud-cloud-controller-manager/refs/heads/main/deploy/ccm-networks.yaml",
			"https://raw.githubusercontent.com/hetznercloud/hcloud-cloud-controller-manager/refs/heads/main/deploy/ccm.yaml",
		}
	}

	clusterNetwork := &core.ClusterNetworkConfig{
		PodSubnets: []string{args.PodSubnets},
		CNI:        toCNIConfig(args.CNI),
	}

	if args.DNSDomain != nil {
		clusterNetwork.DNSDomain = *args.DNSDomain
	}

	if args.ServiceSubnet != nil {
		clusterNetwork.ServiceSubnets = []string{*args.ServiceSubnet}
	}

	configPatch := core.TalosConfig{
		Cluster: &core.ClusterConfig{
			ExternalCloudProvider: &core.ExternalCloudProviderConfig{
				Enabled:   true,
				Manifests: ccmExtraManifests,
			},
			Network: clusterNetwork,
			Discovery: &core.ClusterDiscoveryConfig{
				Enabled: true, // Enable discovery, required for network encryption via kube span
			},
			AllowSchedulingOnControlPlanes: args.AllowSchedulingOnControlPlanes,
			AdminKubeconfig:                adminKubeconfig,
			ExtraManifests:                 args.ExtraManifests,
			ExtraManifestHeaders:           args.ExtraManifestHeaders,
			InlineManifests:                toInlineManifests(args.InlineManifests),
		},
		Machine: &core.MachineConfig{
			Type:            string(args.ServerNodeType),
			NodeLabels:      args.NodeLabels,
			NodeAnnotations: args.NodeAnnotations,
			Network: &core.NetworkConfig{
				Interfaces: []core.Device{
					{
						Interface: "eth1",
						DHCP:      true,
					},
				},
				Nameservers: args.Nameservers,
				KubeSpan: &core.NetworkKubeSpan{
					Enabled: args.EnableKubeSpan, // Enable kube span (wireguard)
				},
			},
			Kubelet: &core.KubeletConfig{
				NodeIP: &core.KubeletNodeIPConfig{
					ValidSubnets: []string{
						args.Subnet,
					},
				},
				ExtraArgs: map[string]string{
					"register-with-taints": toTalosTaints(args.NodeTaints),
					// enable kubelet certificate rotation
					// This is required for deploying a metric server
					// See: https://www.talos.dev/v1.11/kubernetes-guides/configuration/deploy-metrics-server/
					"rotate-server-certificates": "true",
				},
				ExtraMounts: []core.ExtraMount{},
			},
		},
	}

	for _, folder := range args.LocalStorageFolders {
		configPatch.Machine.Kubelet.ExtraMounts = append(configPatch.Machine.Kubelet.ExtraMounts, core.ExtraMount{
			Destination: folder,
			Type:        "bind",
			Source:      folder,
			Options:     []string{"bind", "rshared", "rw"},
		})
	}

	if args.SecretboxEncryptionSecret != nil {
		// Ensure the secret is base64 encoded before setting
		encodedSecret := base64.StdEncoding.EncodeToString([]byte(*args.SecretboxEncryptionSecret))
		configPatch.Cluster.SecretboxEncryptionSecret = encodedSecret
	}

	if args.EnableLonghornSupport {
		configPatch.Machine.Sysctls = map[string]string{
			"vm.nr_hugepages": "1024",
		}

		configPatch.Machine.Kernel = &core.KernelConfig{
			Modules: []core.KernelModuleConfig{
				{Name: "nvme_tcp"},
				{Name: "vfio_pci"},
				{Name: "uio_pci_generic"},
			},
		}

		configPatch.Machine.Kubelet.ExtraMounts = append(configPatch.Machine.Kubelet.ExtraMounts, core.ExtraMount{
			Destination: "/var/lib/longhorn",
			Type:        "bind",
			Source:      "/var/lib/longhorn",
			Options:     []string{"bind", "rshared", "rw"},
		})
	}

	return &configPatch
}

func toCNIConfig(cni *core_config.CNIConfig) *core.CNIConfig {
	if cni == nil {
		return nil
	}

	return &core.CNIConfig{
		Name: cni.Name,
		URLs: cni.URLs,
	}
}

func toTalosTaints(taints []core_config.Taint) string {
	t := ""
	for _, taint := range taints {
		t += fmt.Sprintf("%s=%s:%s,", taint.Key, taint.Value, taint.Effect)
	}
	if len(t) > 0 {
		t = t[:len(t)-1]
	}
	return t
}

func newRegistryConfigs(args *core_config.RegistriesConfig) ([]string, error) {
	if args == nil {
		return nil, nil
	}

	var configs []string

	// Sort mirror keys for deterministic output
	mirrorKeys := make([]string, 0, len(args.Mirrors))
	for key := range args.Mirrors {
		mirrorKeys = append(mirrorKeys, key)
	}
	sort.Strings(mirrorKeys)

	for _, key := range mirrorKeys {
		mirror := args.Mirrors[key]
		endpoints := make([]registry.RegistryEndpoint, len(mirror.Endpoints))
		for i, ep := range mirror.Endpoints {
			endpoints[i] = registry.RegistryEndpoint{
				URL:          ep.URL,
				OverridePath: ep.OverridePath,
			}
		}

		cfg := &registry.RegistryMirrorConfig{
			Name:         key,
			Endpoints:    endpoints,
			SkipFallback: mirror.SkipFallback,
		}
		yamlStr, err := cfg.YAML()
		if err != nil {
			return nil, fmt.Errorf("failed to generate RegistryMirrorConfig YAML for %q: %w", key, err)
		}
		configs = append(configs, yamlStr)
	}

	// Sort config keys for deterministic output
	configKeys := make([]string, 0, len(args.Config))
	for key := range args.Config {
		configKeys = append(configKeys, key)
	}
	sort.Strings(configKeys)

	for _, key := range configKeys {
		regCfg := args.Config[key]

		if regCfg.TLS != nil {
			var clientIdentity *registry.CertificateAndKey
			if regCfg.TLS.ClientIdentity != nil {
				clientIdentity = &registry.CertificateAndKey{
					Cert: regCfg.TLS.ClientIdentity.Cert,
					Key:  regCfg.TLS.ClientIdentity.Key,
				}
			}

			tlsCfg := &registry.RegistryTLSConfig{
				Name:               key,
				ClientIdentity:     clientIdentity,
				CA:                 regCfg.TLS.CA,
				InsecureSkipVerify: regCfg.TLS.InsecureSkipVerify,
			}
			yamlStr, err := tlsCfg.YAML()
			if err != nil {
				return nil, fmt.Errorf("failed to generate RegistryTLSConfig YAML for %q: %w", key, err)
			}
			configs = append(configs, yamlStr)
		}

		if regCfg.Auth != nil {
			authCfg := &registry.RegistryAuthConfig{
				Name:          key,
				Username:      regCfg.Auth.Username,
				Password:      regCfg.Auth.Password,
				Auth:          regCfg.Auth.Auth,
				IdentityToken: regCfg.Auth.IdentityToken,
			}
			yamlStr, err := authCfg.YAML()
			if err != nil {
				return nil, fmt.Errorf("failed to generate RegistryAuthConfig YAML for %q: %w", key, err)
			}
			configs = append(configs, yamlStr)
		}
	}

	return configs, nil
}

func toInlineManifests(manifests []core_config.ClusterInlineManifest) []core.ClusterInlineManifest {
	out := make([]core.ClusterInlineManifest, len(manifests))
	for i, manifest := range manifests {
		out[i] = core.ClusterInlineManifest{
			Name:     manifest.Name,
			Contents: manifest.Contents,
		}
	}
	return out
}

func newVolumeConfigs(args *NodeConfigurationArgs) []*volume.VolumeConfig {
	if args.DiskEncryption == nil {
		return nil
	}

	var configs []*volume.VolumeConfig

	if args.DiskEncryption.EncryptState {
		configs = append(configs, &volume.VolumeConfig{
			Name: "STATE",
			Encryption: &volume.EncryptionSpec{
				Provider: "luks2",
				Keys:     toEncryptionKeys(args.DiskEncryption.Keys),
			},
		})
	}

	if args.DiskEncryption.EncryptEphemeral {
		configs = append(configs, &volume.VolumeConfig{
			Name: "EPHEMERAL",
			Encryption: &volume.EncryptionSpec{
				Provider: "luks2",
				Keys:     toEncryptionKeys(args.DiskEncryption.Keys),
			},
		})
	}

	return configs
}

func toEncryptionKeys(keys []core_config.EncryptionKeyConfig) []volume.EncryptionKey {
	out := make([]volume.EncryptionKey, len(keys))
	for i, key := range keys {
		out[i] = volume.EncryptionKey{
			Slot: key.Slot,
		}
		if key.NodeID != nil {
			out[i].NodeID = &volume.EncryptionKeyNodeID{}
		}
	}
	return out
}
