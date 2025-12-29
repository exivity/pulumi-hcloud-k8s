package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// VolumeConfig represents the volume configuration.
type VolumeConfig struct {
	// TODO: Add fields
}

// YAML marshals the VolumeConfig to YAML.
func (vc *VolumeConfig) YAML() (string, error) {
	out, err := yaml.Marshal(vc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal VolumeConfig: %w", err)
	}
	return string(out), nil
}
