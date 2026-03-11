package registry

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const apiVersion = "v1alpha1"

// RegistryMirrorConfig configures an image registry mirror.
// Generated based on Talos v1.12 documentation:
// https://docs.siderolabs.com/talos/v1.12/reference/configuration/cri/registrymirrorconfig
type RegistryMirrorConfig struct {
	APIVersion   string             `yaml:"apiVersion" validate:"required,eq=v1alpha1"`
	Kind         string             `yaml:"kind" validate:"required,eq=RegistryMirrorConfig"`
	Name         string             `yaml:"name" validate:"required"`
	Endpoints    []RegistryEndpoint `yaml:"endpoints" validate:"required,dive"`
	SkipFallback bool               `yaml:"skipFallback,omitempty"`
}

// RegistryEndpoint defines a registry mirror endpoint.
type RegistryEndpoint struct {
	URL          string `yaml:"url" validate:"required"`
	OverridePath bool   `yaml:"overridePath,omitempty"`
}

// YAML marshals the RegistryMirrorConfig to YAML.
func (c *RegistryMirrorConfig) YAML() (string, error) {
	if c.APIVersion == "" {
		c.APIVersion = apiVersion
	}
	if c.Kind == "" {
		c.Kind = "RegistryMirrorConfig"
	}

	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	out, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal RegistryMirrorConfig: %w", err)
	}
	return string(out), nil
}

// RegistryAuthConfig configures authentication for a registry endpoint.
// Generated based on Talos v1.12 documentation:
// https://docs.siderolabs.com/talos/v1.12/reference/configuration/cri/registryauthconfig
type RegistryAuthConfig struct {
	APIVersion    string `yaml:"apiVersion" validate:"required,eq=v1alpha1"`
	Kind          string `yaml:"kind" validate:"required,eq=RegistryAuthConfig"`
	Name          string `yaml:"name" validate:"required"`
	Username      string `yaml:"username,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Auth          string `yaml:"auth,omitempty"`
	IdentityToken string `yaml:"identityToken,omitempty"`
}

// YAML marshals the RegistryAuthConfig to YAML.
func (c *RegistryAuthConfig) YAML() (string, error) {
	if c.APIVersion == "" {
		c.APIVersion = apiVersion
	}
	if c.Kind == "" {
		c.Kind = "RegistryAuthConfig"
	}

	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	out, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal RegistryAuthConfig: %w", err)
	}
	return string(out), nil
}

// RegistryTLSConfig configures TLS for a registry endpoint.
// Generated based on Talos v1.12 documentation:
// https://docs.siderolabs.com/talos/v1.12/reference/configuration/cri/registrytlsconfig
type RegistryTLSConfig struct {
	APIVersion         string             `yaml:"apiVersion" validate:"required,eq=v1alpha1"`
	Kind               string             `yaml:"kind" validate:"required,eq=RegistryTLSConfig"`
	Name               string             `yaml:"name" validate:"required"`
	ClientIdentity     *CertificateAndKey `yaml:"clientIdentity,omitempty"`
	CA                 string             `yaml:"ca,omitempty"`
	InsecureSkipVerify bool               `yaml:"insecureSkipVerify,omitempty"`
}

// CertificateAndKey holds PEM-encoded certificate and key.
type CertificateAndKey struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

// YAML marshals the RegistryTLSConfig to YAML.
func (c *RegistryTLSConfig) YAML() (string, error) {
	if c.APIVersion == "" {
		c.APIVersion = apiVersion
	}
	if c.Kind == "" {
		c.Kind = "RegistryTLSConfig"
	}

	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	out, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal RegistryTLSConfig: %w", err)
	}
	return string(out), nil
}
