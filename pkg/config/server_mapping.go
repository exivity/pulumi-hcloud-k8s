package config

import (
	"fmt"
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
func ValidateAndSetArchForControlPlane(sl validator.StructLevel) {
	_, ok := sl.Current().Interface().(ControlPlaneConfig)
	if !ok {
		sl.ReportError(nil, "", "", "controlplane_type_assertion_failed", "")
		return
	}

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
func ValidateAndSetArchForNodePool(sl validator.StructLevel) {
	fmt.Printf("DEBUG: ValidateAndSetArchForNodePool called\n")
	nodePool, ok := sl.Current().Interface().(NodePoolConfig)
	if !ok {
		sl.ReportError(nil, "", "", "nodepool_type_assertion_failed", "")
		return
	}

	// If Arch is not set (empty), auto-detect from ServerSize
	if nodePool.Arch == "" {
		// Use reflection to modify the original struct
		val := sl.Current()
		if val.Kind() == reflect.Struct && val.CanSet() {
			archField := val.FieldByName("Arch")
			if archField.IsValid() && archField.CanSet() {
				detectedArch := GetArchFromServerSize(nodePool.ServerSize)
				archField.Set(reflect.ValueOf(detectedArch))
			}
		}
	}
}
