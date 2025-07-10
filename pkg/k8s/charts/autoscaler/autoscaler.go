package autoscaler

import (
	"encoding/base64"
	"encoding/json"
)

// HCloudClusterConfig is the root of the Hetzner cluster‐config.
// It corresponds to the top-level JSON object.
type HCloudClusterConfig struct {
	ImagesForArch ImagesForArch               `json:"imagesForArch"`
	NodeConfigs   map[string]HCloudNodeConfig `json:"nodeConfigs"`
}

// ImagesForArch selects which image to use per CPU architecture.
type ImagesForArch struct {
	ARM64 string `json:"arm64"` // e.g. "ubuntu-20.04"
	AMD64 string `json:"amd64"` // e.g. "ubuntu-20.04"
}

// HCloudNodeConfig holds the per-pool cloud-init, labels and taints.
type HCloudNodeConfig struct {
	CloudInit string            `json:"cloudInit"` // raw cloud-init YAML (not double-base64'd)
	Labels    map[string]string `json:"labels"`
	Taints    []Taint           `json:"taints"`
}

// Taint maps exactly to a Kubernetes taint spec.
type Taint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"` // e.g. "NoExecute"
}

// ToBase64JSON marshals the HCloudClusterConfig to JSON and returns
// a Base64 encoding of that JSON.
func (c *HCloudClusterConfig) ToBase64JSON() (string, error) {
	// 1) Marshal to JSON
	rawJSON, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	// 2) Base64-encode the JSON
	encoded := base64.StdEncoding.EncodeToString(rawJSON)
	return encoded, nil
}
