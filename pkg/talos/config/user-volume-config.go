package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// UserVolumeConfig represents the user volume configuration.
type UserVolumeConfig struct {
	// TODO: Add fields
}

// YAML marshals the UserVolumeConfig to YAML.
func (uvc *UserVolumeConfig) YAML() (string, error) {
	out, err := yaml.Marshal(uvc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal UserVolumeConfig: %w", err)
	}
	return string(out), nil
}
