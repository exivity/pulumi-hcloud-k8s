package validators

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

// ValidateHcloudToken checks if HCloudToken is set when required by enabled features.
// This function works with any struct that has the same field structure as config.PulumiConfig.
// If HCloudToken is empty, it falls back to using the Hetzner token from the Hetzner config.
func ValidateHcloudToken(sl validator.StructLevel) {
	kubernetesField := sl.Current().FieldByName("Kubernetes")
	if !kubernetesField.IsValid() {
		sl.ReportError(nil, "Kubernetes", "Kubernetes", "field_not_found", "")
		return
	}

	hcloudTokenField := kubernetesField.FieldByName("HCloudToken")
	if !hcloudTokenField.IsValid() {
		sl.ReportError(nil, "HCloudToken", "HCloudToken", "field_not_found", "")
		return
	}

	hcloudToken := hcloudTokenField.String()

	// If token is provided, validation passes
	if hcloudToken != "" {
		return
	}

	// Check if Hetzner token can be used as fallback
	if hasValidHetznerToken(sl.Current()) {
		return
	}

	// Check if any Kubernetes features requiring token are enabled
	validateRequiredFeatures(sl, kubernetesField, hcloudToken)
}

// hasValidHetznerToken checks if a valid Hetzner token exists as fallback
func hasValidHetznerToken(current reflect.Value) bool {
	hetznerField := current.FieldByName("Hetzner")
	if !hetznerField.IsValid() {
		return false
	}

	hetznerTokenField := hetznerField.FieldByName("Token")
	if !hetznerTokenField.IsValid() {
		return false
	}

	return hetznerTokenField.String() != ""
}

// validateRequiredFeatures checks if features requiring HCloud token are enabled and reports errors
func validateRequiredFeatures(sl validator.StructLevel, kubernetesField reflect.Value, hcloudToken string) {
	checkFeature(sl, kubernetesField, "HetznerCCM", "required_with_hetznerccm", hcloudToken)
	checkFeature(sl, kubernetesField, "CSI", "required_with_csi", hcloudToken)
	checkFeature(sl, kubernetesField, "ClusterAutoScaler", "required_with_clusterautoscaler", hcloudToken)
}

// checkFeature validates if a specific feature is enabled and reports error if token is missing
func checkFeature(sl validator.StructLevel, kubernetesField reflect.Value, fieldName, errorTag, hcloudToken string) {
	featureField := kubernetesField.FieldByName(fieldName)
	if !featureField.IsValid() || featureField.IsNil() {
		return
	}

	enabledField := featureField.Elem().FieldByName("Enabled")
	if !enabledField.IsValid() || !enabledField.Bool() {
		return
	}

	sl.ReportError(hcloudToken, "HCloudToken", "HCloudToken", errorTag, "")
}
