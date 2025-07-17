package autoscaler

import (
	"encoding/json"

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

func (c *HCloudNodeConfig) ToJSON() (pulumi.StringOutput, error) {
	labels, err := json.Marshal(c.Labels)
	if err != nil {
		return pulumi.String("").ToStringOutput(), err
	}
	taints, err := json.Marshal(c.Taints)
	if err != nil {
		return pulumi.String("").ToStringOutput(), err
	}

	return pulumi.Sprintf(`{
	"cloudInit": "%s",
	"labels": %s,
	"taints": %s
}`,
		c.CloudInit,
		labels,
		taints,
	), nil
}

// ToJSON marshals the HCloudClusterConfig to JSON for handle pulumi serialization.
func (c *HCloudClusterConfig) ToJSON() (pulumi.StringOutput, error) {
	nodeConfiguration := pulumi.Sprintf("")
	nodeConfigurationInitialized := false

	for node, nodeConfig := range c.NodeConfigs {
		jsonNodeConfig, err := nodeConfig.ToJSON()
		if err != nil {
			return pulumi.String("").ToStringOutput(), err
		}
		nodeJson := pulumi.Sprintf(`"%s": %s`, node, jsonNodeConfig)

		if nodeConfigurationInitialized == false {
			nodeConfiguration = nodeJson
			nodeConfigurationInitialized = true
		} else {
			nodeConfiguration = pulumi.Sprintf(`%s, %s`, nodeConfiguration, nodeJson)
		}
	}

	return pulumi.Sprintf(`{
    "imagesForArch": {
        "arm64": "%d",
        "amd64": "%d"
    },
    "nodeConfigs": %s
}`,
		c.ImagesForArch.ARM64,
		c.ImagesForArch.AMD64,
		nodeConfiguration,
	), nil
}
