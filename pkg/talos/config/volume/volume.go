package volume

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// VolumeConfig represents the volume configuration.
// Generated based on Talos v1.12 documentation:
// https://docs.siderolabs.com/talos/v1.12/reference/configuration/block/volumeconfig
//
// To update for new Talos versions, check the documentation for changes in the
// VolumeConfig structure and update the fields accordingly.
type VolumeConfig struct {
	APIVersion   string            `yaml:"apiVersion" validate:"required,eq=v1alpha1"`
	Kind         string            `yaml:"kind" validate:"required,eq=VolumeConfig"`
	Name         string            `yaml:"name" validate:"required,oneof=STATE EPHEMERAL IMAGE-CACHE"`
	Provisioning *ProvisioningSpec `yaml:"provisioning,omitempty" validate:"omitempty"`
	Encryption   *EncryptionSpec   `yaml:"encryption,omitempty" validate:"omitempty"`
}

// ProvisioningSpec describes how the volume is provisioned.
type ProvisioningSpec struct {
	DiskSelector *DiskSelector `yaml:"diskSelector,omitempty" validate:"omitempty"`
	Grow         bool          `yaml:"grow,omitempty"`
	MinSize      string        `yaml:"minSize,omitempty"` // ByteSize
	MaxSize      string        `yaml:"maxSize,omitempty"` // Size
}

// DiskSelector selects a disk for the volume.
type DiskSelector struct {
	Match string `yaml:"match,omitempty" validate:"required"` // CEL expression
}

// EncryptionSpec represents volume encryption settings.
type EncryptionSpec struct {
	Provider  string          `yaml:"provider,omitempty" validate:"omitempty,eq=luks2"` // e.g., luks2
	Keys      []EncryptionKey `yaml:"keys,omitempty" validate:"omitempty,dive"`
	Cipher    string          `yaml:"cipher,omitempty" validate:"omitempty"`
	KeySize   uint            `yaml:"keySize,omitempty"`
	BlockSize uint64          `yaml:"blockSize,omitempty"`
	Options   []string        `yaml:"options,omitempty"`
}

// EncryptionKey represents configuration for disk encryption key.
type EncryptionKey struct {
	Slot        int                  `yaml:"slot" validate:"gte=0"`
	Static      *EncryptionKeyStatic `yaml:"static,omitempty" validate:"omitempty"`
	NodeID      *EncryptionKeyNodeID `yaml:"nodeID,omitempty" validate:"omitempty"`
	KMS         *EncryptionKeyKMS    `yaml:"kms,omitempty" validate:"omitempty"`
	TPM         *EncryptionKeyTPM    `yaml:"tpm,omitempty" validate:"omitempty"`
	LockToState bool                 `yaml:"lockToState,omitempty"`
}

// EncryptionKeyStatic represents throw away key type.
type EncryptionKeyStatic struct {
	Passphrase string `yaml:"passphrase,omitempty" validate:"required"`
}

// EncryptionKeyNodeID represents deterministically generated key from the node UUID and PartitionLabel.
type EncryptionKeyNodeID struct{}

// EncryptionKeyKMS represents a key that is generated and then sealed/unsealed by the KMS server.
type EncryptionKeyKMS struct {
	Endpoint string `yaml:"endpoint,omitempty" validate:"required,url"`
}

// EncryptionKeyTPM represents a key that is generated and then sealed/unsealed by the TPM.
type EncryptionKeyTPM struct {
	Options                       *EncryptionKeyTPMOptions `yaml:"options,omitempty" validate:"omitempty"`
	CheckSecurebootStatusOnEnroll bool                     `yaml:"checkSecurebootStatusOnEnroll,omitempty"`
}

// EncryptionKeyTPMOptions represents the options for TPM-based key protection.
type EncryptionKeyTPMOptions struct {
	PCRs []int `yaml:"pcrs,omitempty"`
}

// YAML marshals the VolumeConfig to YAML.
func (vc *VolumeConfig) YAML() (string, error) {
	// Ensure APIVersion and Kind are set
	if vc.APIVersion == "" {
		vc.APIVersion = "v1alpha1"
	}
	if vc.Kind == "" {
		vc.Kind = "VolumeConfig"
	}

	validate := validator.New()
	if err := validate.Struct(vc); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	out, err := yaml.Marshal(vc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal VolumeConfig: %w", err)
	}
	return string(out), nil
}
