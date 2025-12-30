package config

// ImageGeneratorSizes holds the two supported VM sizes.
type ImageGeneratorSizes struct {
	X86 string `json:"x86" validate:"default=cx23"`
	ARM string `json:"arm" validate:"default=cax11"`
}

// RegistriesConfig sets registry mirrors, auth, etc.
type RegistriesConfig struct {
	Mirrors map[string]RegistryMirrorConfig `json:"mirrors"`
	Config  map[string]RegistryConfig       `json:"config"`
}

// RegistryMirrorConfig configures a single mirror.
type RegistryMirrorConfig struct {
	Endpoints    []string `json:"endpoints"`
	OverridePath bool     `json:"overridePath"`
	SkipFallback bool     `json:"skipFallback"`
}

// RegistryConfig for TLS & auth in container registries.
type RegistryConfig struct {
	TLS  *RegistryTLSConfig  `json:"tls"`
	Auth *RegistryAuthConfig `json:"auth"`
}

// RegistryTLSConfig for mutual TLS or skipping verification.
type RegistryTLSConfig struct {
	ClientIdentity     *PEMEncodedCertificateAndKey `json:"clientIdentity"`
	CA                 string                       `json:"ca"`
	InsecureSkipVerify bool                         `json:"insecureSkipVerify"`
}

// RegistryAuthConfig for basic or token authentication.
type RegistryAuthConfig struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Auth          string `json:"auth"`
	IdentityToken string `json:"identityToken"`
}

// PEMEncodedCertificateAndKey is a base64-encoded certificate + key.
type PEMEncodedCertificateAndKey struct {
	CRT string `json:"crt"`
	Key string `json:"key"`
}

// ClusterInlineManifest represents an inline Kubernetes manifest.
type ClusterInlineManifest struct {
	// Name of the manifest.
	Name string `json:"name" validate:"required"`
	// Manifest contents as a string.
	Contents string `json:"contents" validate:"required"`
}

// CNIConfig holds the CNI configuration for the cluster.
type CNIConfig struct {
	// Name of the CNI to use. Can be "flannel", "custom", or "none".
	Name string `json:"name" validate:"required,oneof=flannel custom none"`
	// URLs of the CNI manifests to apply if Name is "custom".
	URLs []string `json:"urls,omitempty" validate:"required_if=Name custom"`
}

// ProxyConfig holds the proxy configuration for the cluster.
type ProxyConfig struct {
	// Disabled disables the kube-proxy.
	Disabled bool `json:"disabled,omitempty"`
}

// EncryptionKeyNodeID configuration.
type EncryptionKeyNodeID struct{}

// EncryptionKeyConfig defines a single encryption key.
type EncryptionKeyConfig struct {
	// Slot is the LUKS2 keyslot index; LUKS2 supports keyslots 0–31.
	Slot   int                  `json:"slot" validate:"min=0,max=31"`
	NodeID *EncryptionKeyNodeID `json:"node_id,omitempty"`
}

// DiskEncryptionConfig configures disk encryption.
type DiskEncryptionConfig struct {
	// EncryptState enables encryption for the STATE partition.
	EncryptState bool `json:"encrypt_state"`
	// EncryptEphemeral enables encryption for the EPHEMERAL partition.
	EncryptEphemeral bool `json:"encrypt_ephemeral"`
	// Keys is a list of encryption keys to use.
	Keys []EncryptionKeyConfig `json:"keys" validate:"dive"`
}

// TalosConfig contains all Talos Linux image & version settings.
type TalosConfig struct {
	// If set, overrides the ID of the Talos image on Hetzner
	ImageIDOverride *string `json:"image_id_override"`

	// ImageGenerationLocation is the location where the image will be generated.
	// Defaults to "fsn1".
	ImageGenerationLocation string `json:"image_generation_location" validate:"default=fsn1"`

	// Talos image version (GitHub tag)
	ImageVersion string `json:"image_version" validate:"required"`

	// Kubernetes version per Talos support matrix
	KubernetesVersion string `json:"kubernetes_version" validate:"required"`

	// K8sCertificateRenewalDuration is the look-ahead window before a client
	// certificate's NotAfter timestamp when the Talos provider will
	// automatically renew that certificate.
	// Must be a valid Go duration string (see https://pkg.go.dev/time#ParseDuration).
	// Examples:
	//
	//	"10m"   - 10 minutes before expiry
	//	"24h"   - 24 hours  before expiry
	//	"168h"  -  7 days   before expiry
	//	"720h"  - 30 days   before expiry (default)
	//
	// Defaults to "720h" (30 days).
	K8sCertificateRenewalDuration string `json:"k8s_certificate_renewal_duration" validate:"default=720h"`

	// VM sizes for building x86 & ARM images
	GeneratorSizes ImageGeneratorSizes `json:"generator_sizes"`

	// Whether to enable Longhorn CSI support
	EnableLonghorn bool `json:"enable_longhorn"`

	// Registry configuration for image registries
	Registries *RegistriesConfig `json:"registries"`

	// LocalStorageFolders is a list of folders to make accessible for local storage in Talos.
	// Each folder will be mounted as a bind mount with rshared and rw options.
	// See: https://www.talos.dev/v1.7/kubernetes-guides/configuration/local-storage/
	// Example: ["/var/mnt", "/var/local-storage"]
	LocalStorageFolders []string `json:"local_storage_folders"`

	// Enable workers on your control plane
	// This is useful for testing, but not recommended for production
	AllowSchedulingOnControlPlanes bool `json:"allow_scheduling_on_control_planes"`

	// SecretboxEncryptionSecret is a base64-encoded 32-byte key used to encrypt Kubernetes secrets at rest in etcd
	// When set, Talos will enable encryption of Secret objects in etcd
	// example creation: pulumi config set --path hcloud-k8s:talos.secretbox_encryption_secret --secret $(pwgen 32 1)
	SecretboxEncryptionSecret *string `json:"secretbox_encryption_secret" validate:"omitempty,len=32"`

	// CertLifetime is the admin kubeconfig certificate lifetime (default is 1 year).
	// Field format accepts any Go time.Duration format ('1h' for one hour, '10m' for ten minutes).
	// Examples:
	//
	//	"24h"         - 24 hours (1 day)
	//	"168h"        - 168 hours (7 days)
	//	"720h"        - 720 hours (30 days)
	//	"8760h"       - 8760 hours (1 year, default)
	//
	// Defaults to "8760h" (1 year).
	// Note: When modifying this value, it is recommended to also update K8sCertificateRenewalDuration
	// to ensure certificates are renewed well before the kubeconfig expires.
	CertLifetime *string `json:"cert_lifetime" validate:"omitempty"`

	// ExtraManifests is a list of URLs that point to additional manifests.
	// These will get automatically deployed as part of the bootstrap.
	// Examples:
	//
	//	- "https://www.example.com/manifest1.yaml"
	//	- "https://www.example.com/manifest2.yaml"
	ExtraManifests []string `json:"extra_manifests"`

	// ExtraManifestHeaders is a map of key value pairs that will be added while fetching the ExtraManifests.
	// Examples:
	//
	//	Token: "1234567"
	//	X-ExtraInfo: "info"
	ExtraManifestHeaders map[string]string `json:"extra_manifest_headers"`

	// InlineManifests is a list of inline Kubernetes manifests.
	// These will get automatically deployed as part of the bootstrap.
	InlineManifests []ClusterInlineManifest `json:"inline_manifests"`

	// EnableHetznerCCMExtraManifest enables installation of Hetzner Cloud Controller Manager via Talos extra manifests.
	// If enabled, the following manifests will be installed:
	//   - https://raw.githubusercontent.com/hetznercloud/hcloud-cloud-controller-manager/refs/heads/main/deploy/ccm-networks.yaml
	//   - https://raw.githubusercontent.com/hetznercloud/hcloud-cloud-controller-manager/refs/heads/main/deploy/ccm.yaml
	// Disabled by default. If enabled, do not enable HetznerCCM Helm chart in KubernetesConfig.
	EnableHetznerCCMExtraManifest bool `json:"enable_hetzner_ccm_extra_manifest"`

	// EnableKubeSpan can be used to encrypt the traffic with wireguard. This works well with flannel, but it is recommended to disable when using a CNI like Cilium.
	EnableKubeSpan bool `json:"enable_kubespan"`

	// CNI configuration for the cluster.
	CNI *CNIConfig `json:"cni"`

	// DiskEncryption configures disk encryption for system partitions.
	DiskEncryption *DiskEncryptionConfig `json:"disk_encryption"`
}
