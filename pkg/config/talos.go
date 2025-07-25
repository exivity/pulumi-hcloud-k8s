package config

// ImageGeneratorSizes holds the two supported VM sizes.
type ImageGeneratorSizes struct {
	X86 string `json:"x86" validate:"default=cx22"`
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

// TalosConfig contains all Talos Linux image & version settings.
type TalosConfig struct {
	// If set, overrides the ID of the Talos image on Hetzner
	ImageIDOverride *string `json:"image_id_override"`

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

	// EnableLocalStorage enables local storage support in Talos
	// See: https://www.talos.dev/v1.7/kubernetes-guides/configuration/local-storage/
	EnableLocalStorage bool `json:"enable_local_storage"`

	// Enable workers on your control plane
	// This is useful for testing, but not recommended for production
	AllowSchedulingOnControlPlanes bool `json:"allow_scheduling_on_control_planes"`

	// SecretboxEncryptionSecret is a base64-encoded 32-byte key used to encrypt Kubernetes secrets at rest in etcd
	// When set, Talos will enable encryption of Secret objects in etcd
	// example creation: pulumi config set --path hcloud-k8s:talos.secretbox_encryption_secret --secret $(pwgen 32 1)
	SecretboxEncryptionSecret *string `json:"secretbox_encryption_secret" validate:"omitempty,len=32"`
}
