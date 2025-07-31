package validators

import (
	"github.com/go-playground/validator/v10"
)

// ValidateHcloudToken checks if HCloudToken is set when required by enabled features.
// This function works with any struct that has the same field structure as config.PulumiConfig.
func ValidateHcloudToken(sl validator.StructLevel) {
	// Get the Kubernetes field using reflection
	kubernetesField := sl.Current().FieldByName("Kubernetes")
	if !kubernetesField.IsValid() {
		sl.ReportError(nil, "Kubernetes", "Kubernetes", "field_not_found", "")
		return
	}

	// Extract HCloudToken
	hcloudTokenField := kubernetesField.FieldByName("HCloudToken")
	if !hcloudTokenField.IsValid() {
		sl.ReportError(nil, "HCloudToken", "HCloudToken", "field_not_found", "")
		return
	}

	hcloudToken := hcloudTokenField.String()

	// Only check for HCloudToken if it is empty
	if hcloudToken == "" {
		// Check HetznerCCM
		if ccmField := kubernetesField.FieldByName("HetznerCCM"); ccmField.IsValid() && !ccmField.IsNil() {
			if enabledField := ccmField.Elem().FieldByName("Enabled"); enabledField.IsValid() && enabledField.Bool() {
				sl.ReportError(hcloudToken, "HCloudToken", "HCloudToken", "required_with_hetznerccm", "")
			}
		}

		// Check CSI
		if csiField := kubernetesField.FieldByName("CSI"); csiField.IsValid() && !csiField.IsNil() {
			if enabledField := csiField.Elem().FieldByName("Enabled"); enabledField.IsValid() && enabledField.Bool() {
				sl.ReportError(hcloudToken, "HCloudToken", "HCloudToken", "required_with_csi", "")
			}
		}

		// Check ClusterAutoScaler
		if casField := kubernetesField.FieldByName("ClusterAutoScaler"); casField.IsValid() && !casField.IsNil() {
			if enabledField := casField.Elem().FieldByName("Enabled"); enabledField.IsValid() && enabledField.Bool() {
				sl.ReportError(hcloudToken, "HCloudToken", "HCloudToken", "required_with_clusterautoscaler", "")
			}
		}
	}
}
