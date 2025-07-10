package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// YAML marshals the TalosConfig to YAML.
func (tc *TalosConfig) YAML() (string, error) {
	out, err := yaml.Marshal(tc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal TalosConfig: %w", err)
	}
	return string(out), nil
}

// TalosConfig is the top-level structure: version, debug, machine, cluster.
type TalosConfig struct {
	Debug   bool           `yaml:"debug,omitempty"` // Enable verbose logging
	Machine *MachineConfig `yaml:"machine,omitempty"`
	Cluster *ClusterConfig `yaml:"cluster,omitempty"`
}

// MachineConfig represents the "machine:" section of Talos config.
type MachineConfig struct {
	Type                     string                       `yaml:"type,omitempty"`  // "controlplane" or "worker"
	Token                    string                       `yaml:"token,omitempty"` // Bootstrap token
	CA                       *PEMEncodedCertificateAndKey `yaml:"ca,omitempty"`    // root CA for PKI
	AcceptedCAs              []PEMEncodedCertificate      `yaml:"acceptedCAs,omitempty"`
	CertSANs                 []string                     `yaml:"certSANs,omitempty"`
	ControlPlane             *MachineControlPlaneConfig   `yaml:"controlPlane,omitempty"`
	Kubelet                  *KubeletConfig               `yaml:"kubelet,omitempty"`
	Pods                     []interface{}                `yaml:"pods,omitempty"` // static pods as unstructured
	Network                  *NetworkConfig               `yaml:"network,omitempty"`
	Disks                    []MachineDisk                `yaml:"disks,omitempty"`   // partition/format extra disks
	Install                  *InstallConfig               `yaml:"install,omitempty"` // installation config
	Files                    []MachineFile                `yaml:"files,omitempty"`   // user-defined files
	Env                      map[string]string            `yaml:"env,omitempty"`     // environment vars
	Time                     *TimeConfig                  `yaml:"time,omitempty"`
	Sysctls                  map[string]string            `yaml:"sysctls,omitempty"` // sysctl
	Sysfs                    map[string]string            `yaml:"sysfs,omitempty"`
	Registries               *RegistriesConfig            `yaml:"registries,omitempty"`
	SystemDiskEncryption     *SystemDiskEncryptionConfig  `yaml:"systemDiskEncryption,omitempty"`
	Features                 *FeaturesConfig              `yaml:"features,omitempty"`
	Udev                     *UdevConfig                  `yaml:"udev,omitempty"`
	Logging                  *LoggingConfig               `yaml:"logging,omitempty"`
	Kernel                   *KernelConfig                `yaml:"kernel,omitempty"`
	SeccompProfiles          []MachineSeccompProfile      `yaml:"seccompProfiles,omitempty"`
	BaseRuntimeSpecOverrides interface{}                  `yaml:"baseRuntimeSpecOverrides,omitempty"`
	NodeLabels               map[string]string            `yaml:"nodeLabels,omitempty"`
	NodeAnnotations          map[string]string            `yaml:"nodeAnnotations,omitempty"`
	NodeTaints               map[string]string            `yaml:"nodeTaints,omitempty"`
}

// PEMEncodedCertificateAndKey is a base64-encoded certificate + key.
type PEMEncodedCertificateAndKey struct {
	CRT string `yaml:"crt"`
	Key string `yaml:"key"`
}

// PEMEncodedCertificate is a base64-encoded certificate.
type PEMEncodedCertificate struct {
	CRT string `yaml:"crt"`
}

// MachineControlPlaneConfig is machine-specific controlplane config (e.g., disabling scheduler).
type MachineControlPlaneConfig struct {
	ControllerManager *MachineControllerManagerConfig `yaml:"controllerManager,omitempty"`
	Scheduler         *MachineSchedulerConfig         `yaml:"scheduler,omitempty"`
}

// MachineControllerManagerConfig toggles the kube-controller-manager on this node.
type MachineControllerManagerConfig struct {
	Disabled bool `yaml:"disabled"`
}

// MachineSchedulerConfig toggles the kube-scheduler on this node.
type MachineSchedulerConfig struct {
	Disabled bool `yaml:"disabled"`
}

// KubeletConfig for machine.kubelet, e.g. nodeLabels, nodeTaints, clusterDNS, etc.
type KubeletConfig struct {
	Image                               string               `yaml:"image,omitempty"`
	ClusterDNS                          []string             `yaml:"clusterDNS,omitempty"`
	ExtraArgs                           map[string]string    `yaml:"extraArgs,omitempty"`
	ExtraMounts                         []ExtraMount         `yaml:"extraMounts,omitempty"`
	ExtraConfig                         interface{}          `yaml:"extraConfig,omitempty"`
	CredentialProviderConfig            interface{}          `yaml:"credentialProviderConfig,omitempty"`
	DefaultRuntimeSeccompProfileEnabled bool                 `yaml:"defaultRuntimeSeccompProfileEnabled,omitempty"`
	RegisterWithFQDN                    bool                 `yaml:"registerWithFQDN,omitempty"`
	NodeIP                              *KubeletNodeIPConfig `yaml:"nodeIP,omitempty"`
	SkipNodeRegistration                bool                 `yaml:"skipNodeRegistration,omitempty"`
	DisableManifestsDirectory           bool                 `yaml:"disableManifestsDirectory,omitempty"`
}

// ExtraMount describes an extra volume mount for the kubelet container.
type ExtraMount struct {
	Destination string           `yaml:"destination"`
	Type        string           `yaml:"type"`
	Source      string           `yaml:"source"`
	Options     []string         `yaml:"options"`
	UIDMappings []LinuxIDMapping `yaml:"uidMappings,omitempty"`
	GIDMappings []LinuxIDMapping `yaml:"gidMappings,omitempty"`
}

// LinuxIDMapping config for user namespaces.
type LinuxIDMapping struct {
	ContainerID uint32 `yaml:"containerID"`
	HostID      uint32 `yaml:"hostID"`
	Size        uint32 `yaml:"size"`
}

// KubeletNodeIPConfig sets which IP the kubelet uses, e.g., filtering subnets.
type KubeletNodeIPConfig struct {
	ValidSubnets []string `yaml:"validSubnets"`
}

// NetworkConfig is machine.network
type NetworkConfig struct {
	Hostname            string           `yaml:"hostname,omitempty"`
	Interfaces          []Device         `yaml:"interfaces,omitempty"`
	Nameservers         []string         `yaml:"nameservers,omitempty"`
	SearchDomains       []string         `yaml:"searchDomains,omitempty"`
	ExtraHostEntries    []ExtraHost      `yaml:"extraHostEntries,omitempty"`
	KubeSpan            *NetworkKubeSpan `yaml:"kubespan,omitempty"`
	DisableSearchDomain bool             `yaml:"disableSearchDomain,omitempty"`
}

// Device represents a network interface/bond/bridge.
type Device struct {
	Interface      string                 `yaml:"interface,omitempty"`
	DeviceSelector *NetworkDeviceSelector `yaml:"deviceSelector,omitempty"`
	Addresses      []string               `yaml:"addresses,omitempty"`
	Routes         []Route                `yaml:"routes,omitempty"`
	Bond           *Bond                  `yaml:"bond,omitempty"`
	Bridge         *Bridge                `yaml:"bridge,omitempty"`
	BridgePort     *BridgePort            `yaml:"bridgePort,omitempty"`
	VLANs          []Vlan                 `yaml:"vlans,omitempty"`
	MTU            int                    `yaml:"mtu,omitempty"`
	DHCP           bool                   `yaml:"dhcp,omitempty"`
	Ignore         bool                   `yaml:"ignore,omitempty"`
	Dummy          bool                   `yaml:"dummy,omitempty"`
	DHCPOptions    *DHCPOptions           `yaml:"dhcpOptions,omitempty"`
	Wireguard      *DeviceWireguardConfig `yaml:"wireguard,omitempty"`
	VIP            *DeviceVIPConfig       `yaml:"vip,omitempty"`
}

// NetworkDeviceSelector picks a network device by busPath, hardwareAddr, etc.
type NetworkDeviceSelector struct {
	BusPath       string `yaml:"busPath,omitempty"`
	HardwareAddr  string `yaml:"hardwareAddr,omitempty"`
	PermanentAddr string `yaml:"permanentAddr,omitempty"`
	PciID         string `yaml:"pciID,omitempty"`
	Driver        string `yaml:"driver,omitempty"`
	Physical      bool   `yaml:"physical,omitempty"`
}

// Route is a static route for an interface.
type Route struct {
	Network string `yaml:"network"`
	Gateway string `yaml:"gateway,omitempty"`
	Source  string `yaml:"source,omitempty"`
	Metric  uint32 `yaml:"metric,omitempty"`
	MTU     uint32 `yaml:"mtu,omitempty"`
}

// Bond represents bond config (802.3ad, etc).
type Bond struct {
	Interfaces      []string                `yaml:"interfaces,omitempty"`
	DeviceSelectors []NetworkDeviceSelector `yaml:"deviceSelectors,omitempty"`
	Mode            string                  `yaml:"mode,omitempty"`
	XmitHashPolicy  string                  `yaml:"xmitHashPolicy,omitempty"`
	LacpRate        string                  `yaml:"lacpRate,omitempty"`
	AdActorSystem   string                  `yaml:"adActorSystem,omitempty"`
	ArpValidate     string                  `yaml:"arpValidate,omitempty"`
	ArpAllTargets   string                  `yaml:"arpAllTargets,omitempty"`
	Primary         string                  `yaml:"primary,omitempty"`
	PrimaryReselect string                  `yaml:"primaryReselect,omitempty"`
	FailOverMac     string                  `yaml:"failOverMac,omitempty"`
	AdSelect        string                  `yaml:"adSelect,omitempty"`
	Miimon          uint32                  `yaml:"miimon,omitempty"`
	Updelay         uint32                  `yaml:"updelay,omitempty"`
	Downdelay       uint32                  `yaml:"downdelay,omitempty"`
	ArpInterval     uint32                  `yaml:"arpInterval,omitempty"`
	ResendIgmp      uint32                  `yaml:"resendIgmp,omitempty"`
	MinLinks        uint32                  `yaml:"minLinks,omitempty"`
	LpInterval      uint32                  `yaml:"lpInterval,omitempty"`
	PacketsPerSlave uint32                  `yaml:"packetsPerSlave,omitempty"`
	NumPeerNotif    uint8                   `yaml:"numPeerNotif,omitempty"`
	TlbDynamicLb    uint8                   `yaml:"tlbDynamicLb,omitempty"`
	AllSlavesActive uint8                   `yaml:"allSlavesActive,omitempty"`
	UseCarrier      bool                    `yaml:"useCarrier,omitempty"`
	AdActorSysPrio  uint16                  `yaml:"adActorSysPrio,omitempty"`
	AdUserPortKey   uint16                  `yaml:"adUserPortKey,omitempty"`
	PeerNotifyDelay uint32                  `yaml:"peerNotifyDelay,omitempty"`
}

// Bridge is a software bridge.
type Bridge struct {
	Interfaces []string    `yaml:"interfaces,omitempty"`
	STP        *STP        `yaml:"stp,omitempty"`
	VLAN       *BridgeVLAN `yaml:"vlan,omitempty"`
}

// STP config for bridging.
type STP struct {
	Enabled bool `yaml:"enabled"`
}

// BridgeVLAN config for VLAN bridging.
type BridgeVLAN struct {
	VLANFiltering bool `yaml:"vlanFiltering"`
}

// BridgePort config for bridging a port.
type BridgePort struct {
	Master string `yaml:"master"`
}

// Vlan represents a VLAN sub-interface.
type Vlan struct {
	Addresses   []string         `yaml:"addresses,omitempty"`
	Routes      []Route          `yaml:"routes,omitempty"`
	DHCP        bool             `yaml:"dhcp,omitempty"`
	VLANID      uint16           `yaml:"vlanId,omitempty"`
	MTU         uint32           `yaml:"mtu,omitempty"`
	VIP         *DeviceVIPConfig `yaml:"vip,omitempty"`
	DHCPOptions *DHCPOptions     `yaml:"dhcpOptions,omitempty"`
}

// DHCPOptions for DHCP customizations.
type DHCPOptions struct {
	RouteMetric uint32 `yaml:"routeMetric,omitempty"`
	IPv4        bool   `yaml:"ipv4,omitempty"`
	IPv6        bool   `yaml:"ipv6,omitempty"`
	DUIDv6      string `yaml:"duidv6,omitempty"`
}

// DeviceWireguardConfig config for wireguard.
type DeviceWireguardConfig struct {
	PrivateKey   string                `yaml:"privateKey,omitempty"`
	ListenPort   int                   `yaml:"listenPort,omitempty"`
	FirewallMark int                   `yaml:"firewallMark,omitempty"`
	Peers        []DeviceWireguardPeer `yaml:"peers,omitempty"`
}

// DeviceWireguardPeer defines a peer's publicKey, endpoint, allowedIPs, etc.
type DeviceWireguardPeer struct {
	PublicKey                   string   `yaml:"publicKey"`
	Endpoint                    string   `yaml:"endpoint,omitempty"`
	PersistentKeepaliveInterval string   `yaml:"persistentKeepaliveInterval,omitempty"`
	AllowedIPs                  []string `yaml:"allowedIPs,omitempty"`
}

// DeviceVIPConfig sets a shared IP.
type DeviceVIPConfig struct {
	IP           string                 `yaml:"ip,omitempty"`
	EquinixMetal *VIPEquinixMetalConfig `yaml:"equinixMetal,omitempty"`
	HCloud       *VIPHCloudConfig       `yaml:"hcloud,omitempty"`
}

// VIPEquinixMetalConfig for Equinix VIP management.
type VIPEquinixMetalConfig struct {
	APIToken string `yaml:"apiToken,omitempty"`
}

// VIPHCloudConfig for Hetzner Cloud VIP management.
type VIPHCloudConfig struct {
	APIToken string `yaml:"apiToken,omitempty"`
}

// ExtraHost is an entry in /etc/hosts.
type ExtraHost struct {
	IP      string   `yaml:"ip"`
	Aliases []string `yaml:"aliases"`
}

// NetworkKubeSpan config for KubeSpan overlay.
type NetworkKubeSpan struct {
	Enabled                     bool             `yaml:"enabled"`
	AdvertiseKubernetesNetworks bool             `yaml:"advertiseKubernetesNetworks,omitempty"`
	AllowDownPeerBypass         bool             `yaml:"allowDownPeerBypass,omitempty"`
	HarvestExtraEndpoints       bool             `yaml:"harvestExtraEndpoints,omitempty"`
	MTU                         uint32           `yaml:"mtu,omitempty"`
	Filters                     *KubeSpanFilters `yaml:"filters,omitempty"`
}

// KubeSpanFilters for advanced endpoint filtering.
type KubeSpanFilters struct {
	Endpoints []string `yaml:"endpoints,omitempty"`
}

// MachineDisk describes an extra disk to partition/format.
type MachineDisk struct {
	Device     string          `yaml:"device"`
	Partitions []DiskPartition `yaml:"partitions,omitempty"`
}

// DiskPartition is a partition on a disk.
type DiskPartition struct {
	Size       DiskSize `yaml:"size,omitempty"`
	Mountpoint string   `yaml:"mountpoint,omitempty"`
}

// DiskSize can be string or numeric; used by parted.
type DiskSize interface{}

// InstallConfig describes how to install Talos onto disk (ISO/PXE).
type InstallConfig struct {
	Disk              string                   `yaml:"disk,omitempty"`
	DiskSelector      *InstallDiskSelector     `yaml:"diskSelector,omitempty"`
	ExtraKernelArgs   []string                 `yaml:"extraKernelArgs,omitempty"`
	Image             string                   `yaml:"image,omitempty"`
	Extensions        []InstallExtensionConfig `yaml:"extensions,omitempty"`
	Wipe              bool                     `yaml:"wipe,omitempty"`
	LegacyBIOSSupport bool                     `yaml:"legacyBIOSSupport,omitempty"`
}

// InstallDiskSelector picks a disk by size, model, etc.
type InstallDiskSelector struct {
	Size     string `yaml:"size,omitempty"`
	Name     string `yaml:"name,omitempty"`
	Model    string `yaml:"model,omitempty"`
	Serial   string `yaml:"serial,omitempty"`
	Modalias string `yaml:"modalias,omitempty"`
	UUID     string `yaml:"uuid,omitempty"`
	WWID     string `yaml:"wwid,omitempty"`
	Type     string `yaml:"type,omitempty"`
	BusPath  string `yaml:"busPath,omitempty"`
}

// InstallExtensionConfig references a system extension image.
type InstallExtensionConfig struct {
	Image string `yaml:"image"`
}

// MachineFile describes a user file to be placed on the node.
type MachineFile struct {
	Content     string `yaml:"content"`
	Permissions uint32 `yaml:"permissions,omitempty"`
	Path        string `yaml:"path"`
	Op          string `yaml:"op"` // create, append, overwrite
}

// TimeConfig describes machine time sync (NTP).
type TimeConfig struct {
	Disabled    bool     `yaml:"disabled,omitempty"`
	Servers     []string `yaml:"servers,omitempty"`
	BootTimeout string   `yaml:"bootTimeout,omitempty"`
}

// RegistriesConfig sets registry mirrors, auth, etc.
type RegistriesConfig struct {
	Mirrors map[string]RegistryMirrorConfig `yaml:"mirrors,omitempty"`
	Config  map[string]RegistryConfig       `yaml:"config,omitempty"`
}

// RegistryMirrorConfig configures a single mirror.
type RegistryMirrorConfig struct {
	Endpoints    []string `yaml:"endpoints,omitempty"`
	OverridePath bool     `yaml:"overridePath,omitempty"`
	SkipFallback bool     `yaml:"skipFallback,omitempty"`
}

// RegistryConfig for TLS & auth in container registries.
type RegistryConfig struct {
	TLS  *RegistryTLSConfig  `yaml:"tls,omitempty"`
	Auth *RegistryAuthConfig `yaml:"auth,omitempty"`
}

// RegistryTLSConfig for mutual TLS or skipping verification.
type RegistryTLSConfig struct {
	ClientIdentity     *PEMEncodedCertificateAndKey `yaml:"clientIdentity,omitempty"`
	CA                 string                       `yaml:"ca,omitempty"`
	InsecureSkipVerify bool                         `yaml:"insecureSkipVerify,omitempty"`
}

// RegistryAuthConfig for basic or token authentication.
type RegistryAuthConfig struct {
	Username      string `yaml:"username,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Auth          string `yaml:"auth,omitempty"`
	IdentityToken string `yaml:"identityToken,omitempty"`
}

// SystemDiskEncryptionConfig configures ephemeral/state partition encryption.
type SystemDiskEncryptionConfig struct {
	State     *EncryptionConfig `yaml:"state,omitempty"`
	Ephemeral *EncryptionConfig `yaml:"ephemeral,omitempty"`
}

// EncryptionConfig is partition-level encryption settings.
type EncryptionConfig struct {
	Provider  string          `yaml:"provider,omitempty"`
	Keys      []EncryptionKey `yaml:"keys,omitempty"`
	Cipher    string          `yaml:"cipher,omitempty"`
	KeySize   uint            `yaml:"keySize,omitempty"`
	BlockSize uint64          `yaml:"blockSize,omitempty"`
	Options   []string        `yaml:"options,omitempty"`
}

// EncryptionKey is one key entry for disk encryption.
type EncryptionKey struct {
	Static *EncryptionKeyStatic `yaml:"static,omitempty"`
	NodeID *EncryptionKeyNodeID `yaml:"nodeID,omitempty"`
	KMS    *EncryptionKeyKMS    `yaml:"kms,omitempty"`
	Slot   int                  `yaml:"slot,omitempty"`
	TPM    *EncryptionKeyTPM    `yaml:"tpm,omitempty"`
}

// EncryptionKeyStatic is a stored passphrase.
type EncryptionKeyStatic struct {
	Passphrase string `yaml:"passphrase"`
}

// EncryptionKeyNodeID is derived from node UUID/PartitionLabel.
type EncryptionKeyNodeID struct{}

// EncryptionKeyKMS is sealed/unsealed by a KMS service.
type EncryptionKeyKMS struct {
	Endpoint string `yaml:"endpoint,omitempty"`
}

// EncryptionKeyTPM is sealed/unsealed by a TPM.
type EncryptionKeyTPM struct {
	CheckSecurebootStatusOnEnroll bool `yaml:"checkSecurebootStatusOnEnroll,omitempty"`
}

// FeaturesConfig toggles RBAC, stableHostname, kubernetesTalosAPIAccess, etc.
type FeaturesConfig struct {
	RBAC                     bool                            `yaml:"rbac,omitempty"`
	StableHostname           bool                            `yaml:"stableHostname,omitempty"`
	KubernetesTalosAPIAccess *KubernetesTalosAPIAccessConfig `yaml:"kubernetesTalosAPIAccess,omitempty"`
	ApidCheckExtKeyUsage     bool                            `yaml:"apidCheckExtKeyUsage,omitempty"`
	DiskQuotaSupport         bool                            `yaml:"diskQuotaSupport,omitempty"`
	KubePrism                *KubePrism                      `yaml:"kubePrism,omitempty"`
	HostDNS                  *HostDNSConfig                  `yaml:"hostDNS,omitempty"`
	ImageCache               *ImageCacheConfig               `yaml:"imageCache,omitempty"`
	NodeAddressSortAlgorithm string                          `yaml:"nodeAddressSortAlgorithm,omitempty"`
}

// KubernetesTalosAPIAccessConfig enables Talos API from pods.
type KubernetesTalosAPIAccessConfig struct {
	Enabled                     bool     `yaml:"enabled"`
	AllowedRoles                []string `yaml:"allowedRoles,omitempty"`
	AllowedKubernetesNamespaces []string `yaml:"allowedKubernetesNamespaces,omitempty"`
}

// KubePrism sets up a local proxy load balancer.
type KubePrism struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port,omitempty"`
}

// HostDNSConfig configures the host DNS caching resolver.
type HostDNSConfig struct {
	Enabled              bool `yaml:"enabled"`
	ForwardKubeDNSToHost bool `yaml:"forwardKubeDNSToHost,omitempty"`
	ResolveMemberNames   bool `yaml:"resolveMemberNames,omitempty"`
}

// ImageCacheConfig for caching container images locally.
type ImageCacheConfig struct {
	LocalEnabled bool `yaml:"localEnabled"`
}

// UdevConfig describes custom udev rules.
type UdevConfig struct {
	Rules []string `yaml:"rules,omitempty"`
}

// LoggingConfig describes logging destinations.
type LoggingConfig struct {
	Destinations []LoggingDestination `yaml:"destinations,omitempty"`
}

// LoggingDestination is a syslog-like endpoint (tcp/udp).
type LoggingDestination struct {
	Endpoint  string            `yaml:"endpoint"`
	Format    string            `yaml:"format,omitempty"`
	ExtraTags map[string]string `yaml:"extraTags,omitempty"`
}

// KernelConfig loads modules or sets parameters.
type KernelConfig struct {
	Modules []KernelModuleConfig `yaml:"modules,omitempty"`
}

// KernelModuleConfig is a single kernel module to load.
type KernelModuleConfig struct {
	Name       string   `yaml:"name"`
	Parameters []string `yaml:"parameters,omitempty"`
}

// MachineSeccompProfile is a named seccomp profile.
type MachineSeccompProfile struct {
	Name  string      `yaml:"name"`
	Value interface{} `yaml:"value"`
}

// ================================
// CLUSTER SECTION
// ================================

// ClusterConfig is the \"cluster:\" portion of Talos config.
type ClusterConfig struct {
	ID     string `yaml:"id,omitempty"`
	Secret string `yaml:"secret,omitempty"`

	ControlPlane *ControlPlaneConfig   `yaml:"controlPlane,omitempty"`
	ClusterName  string                `yaml:"clusterName,omitempty"`
	Network      *ClusterNetworkConfig `yaml:"network,omitempty"`
	Token        string                `yaml:"token,omitempty"`

	AescbcEncryptionSecret    string `yaml:"aescbcEncryptionSecret,omitempty"`
	SecretboxEncryptionSecret string `yaml:"secretboxEncryptionSecret,omitempty"`

	CA             *PEMEncodedCertificateAndKey `yaml:"ca,omitempty"`
	AcceptedCAs    []PEMEncodedCertificate      `yaml:"acceptedCAs,omitempty"`
	AggregatorCA   *PEMEncodedCertificateAndKey `yaml:"aggregatorCA,omitempty"`
	ServiceAccount *PEMEncodedKey               `yaml:"serviceAccount,omitempty"`

	APIServer         *APIServerConfig         `yaml:"apiServer,omitempty"`
	ControllerManager *ControllerManagerConfig `yaml:"controllerManager,omitempty"`
	Proxy             *ProxyConfig             `yaml:"proxy,omitempty"`
	Scheduler         *SchedulerConfig         `yaml:"scheduler,omitempty"`
	Discovery         *ClusterDiscoveryConfig  `yaml:"discovery,omitempty"`
	Etcd              *EtcdConfig              `yaml:"etcd,omitempty"`
	CoreDNS           *CoreDNS                 `yaml:"coreDNS,omitempty"`

	ExternalCloudProvider *ExternalCloudProviderConfig `yaml:"externalCloudProvider,omitempty"`

	ExtraManifests       []string                `yaml:"extraManifests,omitempty"`
	ExtraManifestHeaders map[string]string       `yaml:"extraManifestHeaders,omitempty"`
	InlineManifests      []ClusterInlineManifest `yaml:"inlineManifests,omitempty"`

	AdminKubeconfig                *AdminKubeconfigConfig `yaml:"adminKubeconfig,omitempty"`
	AllowSchedulingOnControlPlanes bool                   `yaml:"allowSchedulingOnControlPlanes,omitempty"`
}

// PEMEncodedKey is a base64-encoded private key.
type PEMEncodedKey struct {
	Key string `yaml:"key"`
}

// ControlPlaneConfig sets the canonical endpoint and local port.
type ControlPlaneConfig struct {
	Endpoint           Endpoint `yaml:"endpoint,omitempty"`
	LocalAPIServerPort int      `yaml:"localAPIServerPort,omitempty"`
}

// Endpoint is typically a string: \"https://1.2.3.4:6443\".
type Endpoint struct {
	Endpoint string `yaml:"-"`
}

// UnmarshalYAML custom logic to store the raw string in Endpoint.
func (e *Endpoint) UnmarshalYAML(value *yaml.Node) error {
	var raw string
	if err := value.Decode(&raw); err != nil {
		return err
	}
	e.Endpoint = raw
	return nil
}

func (e Endpoint) MarshalYAML() (interface{}, error) {
	return e.Endpoint, nil
}

// ClusterNetworkConfig is cluster.network: cni, dnsDomain, podSubnets, etc.
type ClusterNetworkConfig struct {
	CNI            *CNIConfig `yaml:"cni,omitempty"`
	DNSDomain      string     `yaml:"dnsDomain,omitempty"`
	PodSubnets     []string   `yaml:"podSubnets,omitempty"`
	ServiceSubnets []string   `yaml:"serviceSubnets,omitempty"`
}

// CNIConfig can be flannel, custom, or none.
type CNIConfig struct {
	Name    string            `yaml:"name"`           // flannel, custom, none
	URLs    []string          `yaml:"urls,omitempty"` // if name=custom
	Flannel *FlannelCNIConfig `yaml:"flannel,omitempty"`
}

// FlannelCNIConfig for flannel-specific config.
type FlannelCNIConfig struct {
	ExtraArgs []string `yaml:"extraArgs,omitempty"`
}

// APIServerConfig for the Kubernetes apiserver.
type APIServerConfig struct {
	Image                    string                                `yaml:"image,omitempty"`
	ExtraArgs                map[string]string                     `yaml:"extraArgs,omitempty"`
	ExtraVolumes             []VolumeMountConfig                   `yaml:"extraVolumes,omitempty"`
	Env                      map[string]string                     `yaml:"env,omitempty"`
	CertSANs                 []string                              `yaml:"certSANs,omitempty"`
	DisablePodSecurityPolicy bool                                  `yaml:"disablePodSecurityPolicy,omitempty"`
	AdmissionControl         []AdmissionPluginConfig               `yaml:"admissionControl,omitempty"`
	AuditPolicy              interface{}                           `yaml:"auditPolicy,omitempty"`
	Resources                *ResourcesConfig                      `yaml:"resources,omitempty"`
	AuthorizationConfig      []AuthorizationConfigAuthorizerConfig `yaml:"authorizationConfig,omitempty"`
}

// VolumeMountConfig to define extra volumes in static pods.
type VolumeMountConfig struct {
	HostPath  string `yaml:"hostPath"`
	MountPath string `yaml:"mountPath"`
	Readonly  bool   `yaml:"readonly,omitempty"`
}

// AdmissionPluginConfig for admission plugins like PodSecurity.
type AdmissionPluginConfig struct {
	Name          string      `yaml:"name"`
	Configuration interface{} `yaml:"configuration,omitempty"`
}

// ResourcesConfig sets requests/limits.
type ResourcesConfig struct {
	Requests interface{} `yaml:"requests,omitempty"`
	Limits   interface{} `yaml:"limits,omitempty"`
}

// AuthorizationConfigAuthorizerConfig for Node, RBAC, or Webhook authorizers.
type AuthorizationConfigAuthorizerConfig struct {
	Type    string      `yaml:"type"`
	Name    string      `yaml:"name"`
	Webhook interface{} `yaml:"webhook,omitempty"`
}

// ControllerManagerConfig for the kube-controller-manager manifest.
type ControllerManagerConfig struct {
	Image        string              `yaml:"image,omitempty"`
	ExtraArgs    map[string]string   `yaml:"extraArgs,omitempty"`
	ExtraVolumes []VolumeMountConfig `yaml:"extraVolumes,omitempty"`
	Env          map[string]string   `yaml:"env,omitempty"`
	Resources    *ResourcesConfig    `yaml:"resources,omitempty"`
}

// ProxyConfig for the kube-proxy manifest.
type ProxyConfig struct {
	Disabled  bool              `yaml:"disabled,omitempty"`
	Image     string            `yaml:"image,omitempty"`
	Mode      string            `yaml:"mode,omitempty"`
	ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
}

// SchedulerConfig for the kube-scheduler manifest.
type SchedulerConfig struct {
	Image        string              `yaml:"image,omitempty"`
	ExtraArgs    map[string]string   `yaml:"extraArgs,omitempty"`
	ExtraVolumes []VolumeMountConfig `yaml:"extraVolumes,omitempty"`
	Env          map[string]string   `yaml:"env,omitempty"`
	Resources    *ResourcesConfig    `yaml:"resources,omitempty"`
	Config       interface{}         `yaml:"config,omitempty"` // custom scheduler config
}

// ClusterDiscoveryConfig toggles discovery & registries.
type ClusterDiscoveryConfig struct {
	Enabled    bool                       `yaml:"enabled,omitempty"`
	Registries *DiscoveryRegistriesConfig `yaml:"registries,omitempty"`
}

// DiscoveryRegistriesConfig includes `kubernetes` or `service`.
type DiscoveryRegistriesConfig struct {
	Kubernetes *RegistryKubernetesConfig `yaml:"kubernetes,omitempty"`
	Service    *RegistryServiceConfig    `yaml:"service,omitempty"`
}

// RegistryKubernetesConfig uses k8s for cluster membership.
type RegistryKubernetesConfig struct {
	Disabled bool `yaml:"disabled,omitempty"`
}

// RegistryServiceConfig uses an external service for cluster membership.
type RegistryServiceConfig struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Endpoint string `yaml:"endpoint,omitempty"`
}

// EtcdConfig for the etcd manifest.
type EtcdConfig struct {
	Image             string                       `yaml:"image,omitempty"`
	CA                *PEMEncodedCertificateAndKey `yaml:"ca,omitempty"`
	ExtraArgs         map[string]string            `yaml:"extraArgs,omitempty"`
	AdvertisedSubnets []string                     `yaml:"advertisedSubnets,omitempty"`
	ListenSubnets     []string                     `yaml:"listenSubnets,omitempty"`
}

// CoreDNS config for coredns installation.
type CoreDNS struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Image    string `yaml:"image,omitempty"`
}

// ExternalCloudProviderConfig sets external CCM usage.
type ExternalCloudProviderConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Manifests []string `yaml:"manifests,omitempty"`
}

// ClusterInlineManifest is an inline K8s YAML doc.
type ClusterInlineManifest struct {
	Name     string `yaml:"name"`
	Contents string `yaml:"contents"`
}

// AdminKubeconfigConfig sets admin kubeconfig certificate lifetime.
type AdminKubeconfigConfig struct {
	CertLifetime string `yaml:"certLifetime,omitempty"`
}
