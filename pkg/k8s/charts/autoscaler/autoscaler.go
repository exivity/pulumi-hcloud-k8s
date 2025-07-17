package autoscaler

import (
	"crypto/sha256"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// HCloudClusterConfig is the root of the Hetzner cluster‐config.
// It corresponds to the top-level JSON object.
type HCloudClusterConfig struct {
	ImagesForArch ImagesForArch               `json:"imagesForArch"`
	NodeConfigs   map[string]HCloudNodeConfig `json:"nodeConfigs"`
}

// ImagesForArch selects which image to use per CPU architecture.
type ImagesForArch struct {
	ARM64 pulumi.IntOutput `json:"arm64"`
	AMD64 pulumi.IntOutput `json:"amd64"`
}

// HCloudNodeConfig holds the per-pool cloud-init, labels and taints.
type HCloudNodeConfig struct {
	CloudInit pulumi.StringOutput `json:"cloudInit"` // raw cloud-init YAML (not double-base64'd)
	Labels    map[string]string   `json:"labels"`
	Taints    []Taint             `json:"taints"`
}

// Taint maps exactly to a Kubernetes taint spec.
type Taint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"` // e.g. "NoExecute"
}

// ToJSON marshals the HCloudClusterConfig to JSON for handle pulumi serialization.
func (c *HCloudClusterConfig) ToJSON() pulumi.StringOutput {

	nodeConfigs := map[string]interface{}{}
	for name, config := range c.NodeConfigs {
		nodeConfigs[name] = map[string]interface{}{
			"cloudInit": config.CloudInit,
			"labels":    config.Labels,
			"taints":    config.Taints,
		}
	}

	return pulumi.JSONMarshal(map[string]interface{}{
		"imagesForArch": map[string]interface{}{
			"arm64": pulumi.Sprintf("%d", c.ImagesForArch.ARM64),
			"amd64": pulumi.Sprintf("%d", c.ImagesForArch.AMD64),
		},
		"nodeConfigs": nodeConfigs,
	})
}

func hashJSON(jsonStr pulumi.StringOutput) pulumi.StringOutput {
	// Use pulumi.Apply to compute the hash of the JSON string
	return jsonStr.ApplyT(func(s string) (string, error) {
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
		return hash, nil
	}).(pulumi.StringOutput)
}
