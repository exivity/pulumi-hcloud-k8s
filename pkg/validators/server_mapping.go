package validators

import (
	"reflect"
	"strings"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/go-playground/validator/v10"
)

// GetArchFromServerSize returns the CPU architecture for a given server size.
// It first tries to look up the exact server size in the map, then falls back
// to pattern matching based on the server type prefix.
func GetArchFromServerSize(serverSize string) image.CPUArchitecture {
	// Fallback to pattern matching for any new server types
	if strings.HasPrefix(serverSize, "cax") {
		return image.ArchARM
	}

	// All other prefixes (cx, ccx, cpx) are x86/AMD64
	if strings.HasPrefix(serverSize, "cx") ||
		strings.HasPrefix(serverSize, "ccx") ||
		strings.HasPrefix(serverSize, "cpx") {
		return image.ArchX86
	}

	// Default to AMD64 for unknown server types
	return image.ArchX86
}

// ValidateAndSetArchForControlPlane validates and auto-sets the Arch field for all ControlPlaneNodePoolConfig in ControlPlaneConfig
// Note: This function works with interface{} to avoid import cycles
func ValidateAndSetArchForControlPlane(sl validator.StructLevel) {
	val := sl.Current()
	if !isValidStructForModification(val) {
		return
	}

	nodePoolsField := val.FieldByName("NodePools")
	if !isValidSliceField(nodePoolsField) {
		return
	}

	processNodePoolSlice(nodePoolsField)
}

// isValidStructForModification checks if the value is a valid struct that can be modified
func isValidStructForModification(val reflect.Value) bool {
	return val.Kind() == reflect.Struct && val.CanSet()
}

// isValidSliceField checks if the field is a valid slice that can be modified
func isValidSliceField(field reflect.Value) bool {
	return field.IsValid() && field.CanSet() && field.Kind() == reflect.Slice
}

// processNodePoolSlice processes each node pool in the slice and sets the arch if needed
func processNodePoolSlice(nodePoolsField reflect.Value) {
	for i := 0; i < nodePoolsField.Len(); i++ {
		nodePoolItem := nodePoolsField.Index(i)
		if !isValidStructForModification(nodePoolItem) {
			continue
		}

		setArchIfEmpty(nodePoolItem)
	}
}

// setArchIfEmpty sets the Arch field based on ServerSize if it's currently empty
func setArchIfEmpty(nodePoolItem reflect.Value) {
	serverSizeField := nodePoolItem.FieldByName("ServerSize")
	archField := nodePoolItem.FieldByName("Arch")

	if !areFieldsValidForArchUpdate(serverSizeField, archField) {
		return
	}

	serverSize := serverSizeField.String()
	currentArch := archField.String()

	if currentArch == "" {
		detectedArch := GetArchFromServerSize(serverSize)
		archField.Set(reflect.ValueOf(detectedArch))
	}
}

// areFieldsValidForArchUpdate checks if both serverSize and arch fields are valid for updating
func areFieldsValidForArchUpdate(serverSizeField, archField reflect.Value) bool {
	return serverSizeField.IsValid() && archField.IsValid() && archField.CanSet()
}

// ValidateAndSetArchForNodePool validates and auto-sets the Arch field for NodePoolConfig
// Note: This function works with interface{} to avoid import cycles
func ValidateAndSetArchForNodePool(sl validator.StructLevel) {
	val := sl.Current()
	if !isValidStructForModification(val) {
		return
	}

	setArchIfEmpty(val)
}
