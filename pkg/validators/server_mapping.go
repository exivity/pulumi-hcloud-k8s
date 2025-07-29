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
	// Get the underlying struct pointer to modify it
	val := sl.Current()
	if val.Kind() == reflect.Struct && val.CanSet() {
		nodePoolsField := val.FieldByName("NodePools")
		if nodePoolsField.IsValid() && nodePoolsField.CanSet() && nodePoolsField.Kind() == reflect.Slice {
			for i := 0; i < nodePoolsField.Len(); i++ {
				nodePoolItem := nodePoolsField.Index(i)
				if nodePoolItem.Kind() == reflect.Struct && nodePoolItem.CanSet() {
					// Get the actual values for checking
					serverSizeField := nodePoolItem.FieldByName("ServerSize")
					archField := nodePoolItem.FieldByName("Arch")

					if serverSizeField.IsValid() && archField.IsValid() && archField.CanSet() {
						serverSize := serverSizeField.String()
						currentArch := archField.String()

						if currentArch == "" {
							detectedArch := GetArchFromServerSize(serverSize)
							archField.Set(reflect.ValueOf(detectedArch))
						}
					}
				}
			}
		}
	}
}

// ValidateAndSetArchForNodePool validates and auto-sets the Arch field for NodePoolConfig
// Note: This function works with interface{} to avoid import cycles
func ValidateAndSetArchForNodePool(sl validator.StructLevel) {
	// Use reflection to get the Arch field value
	val := sl.Current()
	if val.Kind() == reflect.Struct && val.CanSet() {
		archField := val.FieldByName("Arch")
		serverSizeField := val.FieldByName("ServerSize")

		if archField.IsValid() && archField.CanSet() && serverSizeField.IsValid() {
			currentArch := archField.String()
			serverSize := serverSizeField.String()

			// If Arch is not set (empty), auto-detect from ServerSize
			if currentArch == "" {
				detectedArch := GetArchFromServerSize(serverSize)
				archField.Set(reflect.ValueOf(detectedArch))
			}
		}
	}
}
