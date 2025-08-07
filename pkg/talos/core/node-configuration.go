package core

import (
	"encoding/base64"
	"fmt"

	core_config "github.com/exivity/pulumi-hcloud-k8s/pkg/config"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/hetzner/meta"
	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/config"
)

type NodeConfigurationArgs struct {
	// BootstrapEnable is true if bootstrap is enabled
	// this is used to enable the bootstrap service like encryption. This can only use once and should not be apply for updates
	BootstrapEnable bool
	// ServerNodeType is the type of the server node
	ServerNodeType meta.ServerNodeType
	// Subnet is the subnet for the cluster
	Subnet string
	// PodSubnets is the pod subnets for the cluster
	PodSubnets string
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
	// EnableLocalStorage is true if local storage support is enabled
	EnableLocalStorage bool
	// Registries is the registries configuration for the Talos image
	Registries *core_config.RegistriesConfig
}

func NewNodeConfiguration(args *NodeConfigurationArgs) (string, error) {
	var adminKubeconfig *config.AdminKubeconfigConfig
	if args.CertLifetime != nil {
		adminKubeconfig = &config.AdminKubeconfigConfig{
			CertLifetime: *args.CertLifetime,
		}
	}

	configPatch := config.TalosConfig{
		Cluster: &config.ClusterConfig{
			ExternalCloudProvider: &config.ExternalCloudProviderConfig{
				Enabled: true,
			},
			Network: &config.ClusterNetworkConfig{
				PodSubnets: []string{args.PodSubnets},
			},
			Discovery: &config.ClusterDiscoveryConfig{
				Enabled: true, // Enable discovery, required for network encryption via kube span
			},
			AllowSchedulingOnControlPlanes: args.AllowSchedulingOnControlPlanes,
			AdminKubeconfig:                adminKubeconfig,
		},
		Machine: &config.MachineConfig{
			Type:            string(args.ServerNodeType),
			NodeLabels:      args.NodeLabels,
			NodeAnnotations: args.NodeAnnotations,
			Network: &config.NetworkConfig{
				Interfaces: []config.Device{
					{
						Interface: "eth1",
						DHCP:      true,
					},
				},
				Nameservers: args.Nameservers,
				KubeSpan: &config.NetworkKubeSpan{
					Enabled: true, // Enable kube span (wireguard)
				},
			},
			Kubelet: &config.KubeletConfig{
				NodeIP: &config.KubeletNodeIPConfig{
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
				ExtraMounts: []config.ExtraMount{},
			},
		},
	}

	if args.EnableLocalStorage {
		configPatch.Machine.Kubelet.ExtraMounts = append(configPatch.Machine.Kubelet.ExtraMounts, config.ExtraMount{
			Destination: "/var/mnt",
			Type:        "bind",
			Source:      "/var/mnt",
			Options:     []string{"bind", "rshared", "rw"},
		})
	}

	if args.BootstrapEnable {
		// Enable encryption by default
		// TODO: add options to configure more secure encryption
		// See: https://www.talos.dev/v1.9/talos-guides/configuration/disk-encryption/#luks2
		configPatch.Machine.SystemDiskEncryption = &config.SystemDiskEncryptionConfig{
			State: &config.EncryptionConfig{
				Provider: "luks2",
				Keys: []config.EncryptionKey{
					{
						NodeID: &config.EncryptionKeyNodeID{},
						Slot:   0,
					},
				},
			},
		}
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

		configPatch.Machine.Kernel = &config.KernelConfig{
			Modules: []config.KernelModuleConfig{
				{Name: "nvme_tcp"},
				{Name: "vfio_pci"},
				{Name: "uio_pci_generic"},
			},
		}

		configPatch.Machine.Kubelet.ExtraMounts = append(configPatch.Machine.Kubelet.ExtraMounts, config.ExtraMount{
			Destination: "/var/lib/longhorn",
			Type:        "bind",
			Source:      "/var/lib/longhorn",
			Options:     []string{"bind", "rshared", "rw"},
		})
	}

	configPatch.Machine.Registries = toRegistriesConfig(args.Registries)

	return configPatch.YAML()
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

func toRegistriesConfig(args *core_config.RegistriesConfig) *config.RegistriesConfig {
	out := &config.RegistriesConfig{
		Mirrors: map[string]config.RegistryMirrorConfig{},
		Config:  map[string]config.RegistryConfig{},
	}

	if args == nil {
		return out
	}

	for key, mirror := range args.Mirrors {
		out.Mirrors[key] = config.RegistryMirrorConfig{
			Endpoints:    mirror.Endpoints,
			OverridePath: mirror.OverridePath,
			SkipFallback: mirror.SkipFallback,
		}
	}

	for key, registry := range args.Config {
		var tls *config.RegistryTLSConfig
		if registry.TLS != nil {
			var clientIdentity *config.PEMEncodedCertificateAndKey
			if registry.TLS.ClientIdentity != nil {
				clientIdentity = &config.PEMEncodedCertificateAndKey{
					CRT: registry.TLS.ClientIdentity.CRT,
					Key: registry.TLS.ClientIdentity.Key,
				}
			}

			tls = &config.RegistryTLSConfig{
				ClientIdentity:     clientIdentity,
				CA:                 registry.TLS.CA,
				InsecureSkipVerify: registry.TLS.InsecureSkipVerify,
			}
		}

		var auth *config.RegistryAuthConfig
		if registry.Auth != nil {
			auth = &config.RegistryAuthConfig{
				Username:      registry.Auth.Username,
				Password:      registry.Auth.Password,
				Auth:          registry.Auth.Auth,
				IdentityToken: registry.Auth.IdentityToken,
			}
		}

		out.Config[key] = config.RegistryConfig{
			TLS:  tls,
			Auth: auth,
		}
	}

	return out
}
