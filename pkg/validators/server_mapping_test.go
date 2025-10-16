package validators

import (
	"reflect"
	"testing"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
	"github.com/go-playground/validator/v10"
)

// Test structs that mimic the config structs to avoid import cycles
type testNodePoolConfig struct {
	Name       string                `json:"name"`
	ServerSize string                `json:"server_size"`
	Arch       image.CPUArchitecture `json:"arch"`
	Region     string                `json:"region"`
}

type testControlPlaneNodePoolConfig struct {
	Count      int                   `json:"count"`
	ServerSize string                `json:"server_size"`
	Arch       image.CPUArchitecture `json:"arch"`
	Region     string                `json:"region"`
}

type testControlPlaneConfig struct {
	LoadBalancerType string                           `json:"load_balancer_type"`
	NodePools        []testControlPlaneNodePoolConfig `json:"node_pools"`
}

// Mock StructLevel implementation for testing
type mockStructLevel struct {
	current reflect.Value
}

func (m *mockStructLevel) Current() reflect.Value {
	return m.current
}

func (m *mockStructLevel) Parent() reflect.Value {
	return reflect.Value{}
}

func (m *mockStructLevel) Top() reflect.Value {
	return reflect.Value{}
}

func (m *mockStructLevel) ExtractType(field reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return field, field.Kind(), true
}

func (m *mockStructLevel) ReportError(value interface{}, namespace string, structNamespace string, tag string, param string) {
	// Mock implementation - do nothing for tests
}

func (m *mockStructLevel) ReportValidationErrors(relativeNamespace string, relativeActualNamespace string, errs validator.ValidationErrors) {
	// Mock implementation - do nothing for tests
}

func (m *mockStructLevel) Validator() *validator.Validate {
	return validator.New()
}

func TestGetArchFromServerSize(t *testing.T) {
	type args struct {
		serverSize string
	}
	tests := []struct {
		name string
		args args
		want image.CPUArchitecture
	}{
		// ARM64 server types (cax series)
		{
			name: "cax11 should return ARM64",
			args: args{serverSize: "cax11"},
			want: image.ArchARM,
		},
		{
			name: "cax21 should return ARM64",
			args: args{serverSize: "cax21"},
			want: image.ArchARM,
		},
		{
			name: "cax31 should return ARM64",
			args: args{serverSize: "cax31"},
			want: image.ArchARM,
		},
		{
			name: "cax41 should return ARM64",
			args: args{serverSize: "cax41"},
			want: image.ArchARM,
		},

		// x86/AMD64 server types (cx series)
		{
			name: "cx23 should return x86",
			args: args{serverSize: "cx23"},
			want: image.ArchX86,
		},
		{
			name: "cx32 should return x86",
			args: args{serverSize: "cx32"},
			want: image.ArchX86,
		},
		{
			name: "cx42 should return x86",
			args: args{serverSize: "cx42"},
			want: image.ArchX86,
		},
		{
			name: "cx52 should return x86",
			args: args{serverSize: "cx52"},
			want: image.ArchX86,
		},
		{
			name: "cpx11 should return x86",
			args: args{serverSize: "cpx11"},
			want: image.ArchX86,
		},
		{
			name: "cpx21 should return x86",
			args: args{serverSize: "cpx21"},
			want: image.ArchX86,
		},
		{
			name: "cpx31 should return x86",
			args: args{serverSize: "cx32"},
			want: image.ArchX86,
		},
		{
			name: "cpx41 should return x86",
			args: args{serverSize: "cx42"},
			want: image.ArchX86,
		},
		{
			name: "cpx51 should return x86",
			args: args{serverSize: "cpx51"},
			want: image.ArchX86,
		},

		// Dedicated CPU instances (ccx series)
		{
			name: "ccx13 should return x86",
			args: args{serverSize: "ccx13"},
			want: image.ArchX86,
		},
		{
			name: "ccx23 should return x86",
			args: args{serverSize: "ccx23"},
			want: image.ArchX86,
		},
		{
			name: "ccx33 should return x86",
			args: args{serverSize: "ccx33"},
			want: image.ArchX86,
		},
		{
			name: "ccx43 should return x86",
			args: args{serverSize: "ccx43"},
			want: image.ArchX86,
		},
		{
			name: "ccx53 should return x86",
			args: args{serverSize: "ccx53"},
			want: image.ArchX86,
		},
		{
			name: "ccx63 should return x86",
			args: args{serverSize: "ccx63"},
			want: image.ArchX86,
		},

		// Compute optimized instances (cpx series)
		{
			name: "cpx11 should return x86",
			args: args{serverSize: "cpx11"},
			want: image.ArchX86,
		},
		{
			name: "cpx21 should return x86",
			args: args{serverSize: "cpx21"},
			want: image.ArchX86,
		},
		{
			name: "cpx31 should return x86",
			args: args{serverSize: "cpx31"},
			want: image.ArchX86,
		},
		{
			name: "cpx41 should return x86",
			args: args{serverSize: "cpx41"},
			want: image.ArchX86,
		},
		{
			name: "cpx51 should return x86",
			args: args{serverSize: "cpx51"},
			want: image.ArchX86,
		},

		// Pattern matching tests for future server types
		{
			name: "future cax server should return ARM64",
			args: args{serverSize: "cax99"},
			want: image.ArchARM,
		},
		{
			name: "future cx server should return x86",
			args: args{serverSize: "cx99"},
			want: image.ArchX86,
		},
		{
			name: "future ccx server should return x86",
			args: args{serverSize: "ccx99"},
			want: image.ArchX86,
		},
		{
			name: "future cpx server should return x86",
			args: args{serverSize: "cpx99"},
			want: image.ArchX86,
		},

		// Edge cases and unknown server types
		{
			name: "unknown server type should default to x86",
			args: args{serverSize: "unknown123"},
			want: image.ArchX86,
		},
		{
			name: "empty string should default to x86",
			args: args{serverSize: ""},
			want: image.ArchX86,
		},
		{
			name: "partial cax match should return x86",
			args: args{serverSize: "ca11"},
			want: image.ArchX86,
		},
		{
			name: "partial cx match should return x86",
			args: args{serverSize: "c22"},
			want: image.ArchX86,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetArchFromServerSize(tt.args.serverSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetArchFromServerSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndSetArchForControlPlane(t *testing.T) {
	tests := []struct {
		name           string
		input          testControlPlaneConfig
		expectedArches []image.CPUArchitecture
	}{
		{
			name: "single ARM64 node pool with empty arch",
			input: testControlPlaneConfig{
				LoadBalancerType: "lb11",
				NodePools: []testControlPlaneNodePoolConfig{
					{
						Count:      1,
						ServerSize: "cax31",
						Arch:       "", // empty, should be auto-detected
						Region:     "hel1",
					},
				},
			},
			expectedArches: []image.CPUArchitecture{image.ArchARM},
		},
		{
			name: "single x86 node pool with empty arch",
			input: testControlPlaneConfig{
				LoadBalancerType: "lb11",
				NodePools: []testControlPlaneNodePoolConfig{
					{
						Count:      1,
						ServerSize: "cx42",
						Arch:       "", // empty, should be auto-detected
						Region:     "hel1",
					},
				},
			},
			expectedArches: []image.CPUArchitecture{image.ArchX86},
		},
		{
			name: "multiple node pools with empty arches",
			input: testControlPlaneConfig{
				LoadBalancerType: "lb11",
				NodePools: []testControlPlaneNodePoolConfig{
					{
						Count:      1,
						ServerSize: "cax31",
						Arch:       "", // empty, should be auto-detected to ARM
						Region:     "hel1",
					},
					{
						Count:      1,
						ServerSize: "cx42",
						Arch:       "", // empty, should be auto-detected to x86
						Region:     "fsn1",
					},
					{
						Count:      1,
						ServerSize: "ccx33",
						Arch:       "", // empty, should be auto-detected to x86
						Region:     "nbg1",
					},
				},
			},
			expectedArches: []image.CPUArchitecture{image.ArchARM, image.ArchX86, image.ArchX86},
		},
		{
			name: "node pool with already set arch should not change",
			input: testControlPlaneConfig{
				LoadBalancerType: "lb11",
				NodePools: []testControlPlaneNodePoolConfig{
					{
						Count:      1,
						ServerSize: "cax31",
						Arch:       image.ArchX86, // explicitly set to x86, should not change
						Region:     "hel1",
					},
				},
			},
			expectedArches: []image.CPUArchitecture{image.ArchX86}, // should remain x86
		},
		{
			name: "mixed node pools - some with arch set, some without",
			input: testControlPlaneConfig{
				LoadBalancerType: "lb11",
				NodePools: []testControlPlaneNodePoolConfig{
					{
						Count:      1,
						ServerSize: "cax31",
						Arch:       "", // empty, should be auto-detected to ARM
						Region:     "hel1",
					},
					{
						Count:      1,
						ServerSize: "cx42",
						Arch:       image.ArchARM, // explicitly set to ARM, should not change
						Region:     "fsn1",
					},
				},
			},
			expectedArches: []image.CPUArchitecture{image.ArchARM, image.ArchARM},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the input that we can modify
			testInput := tt.input

			// Create mock struct level with the test input
			val := reflect.ValueOf(&testInput).Elem()
			mock := &mockStructLevel{current: val}

			// Call the validator
			ValidateAndSetArchForControlPlane(mock)

			// Check that the arches were set correctly
			if len(testInput.NodePools) != len(tt.expectedArches) {
				t.Fatalf("Expected %d node pools, got %d", len(tt.expectedArches), len(testInput.NodePools))
			}

			for i, expectedArch := range tt.expectedArches {
				if testInput.NodePools[i].Arch != expectedArch {
					t.Errorf("NodePool[%d]: expected arch %s, got %s", i, expectedArch, testInput.NodePools[i].Arch)
				}
			}
		})
	}
}

func TestValidateAndSetArchForNodePool(t *testing.T) {
	tests := []struct {
		name         string
		input        testNodePoolConfig
		expectedArch image.CPUArchitecture
	}{
		{
			name: "ARM64 server with empty arch should be auto-detected",
			input: testNodePoolConfig{
				Name:       "worker",
				ServerSize: "cax31",
				Arch:       "", // empty, should be auto-detected
				Region:     "hel1",
			},
			expectedArch: image.ArchARM,
		},
		{
			name: "x86 server with empty arch should be auto-detected",
			input: testNodePoolConfig{
				Name:       "worker",
				ServerSize: "cx42",
				Arch:       "", // empty, should be auto-detected
				Region:     "hel1",
			},
			expectedArch: image.ArchX86,
		},
		{
			name: "dedicated CPU server with empty arch should be auto-detected",
			input: testNodePoolConfig{
				Name:       "worker",
				ServerSize: "ccx33",
				Arch:       "", // empty, should be auto-detected
				Region:     "hel1",
			},
			expectedArch: image.ArchX86,
		},
		{
			name: "compute optimized server with empty arch should be auto-detected",
			input: testNodePoolConfig{
				Name:       "worker",
				ServerSize: "cpx41",
				Arch:       "", // empty, should be auto-detected
				Region:     "hel1",
			},
			expectedArch: image.ArchX86,
		},
		{
			name: "node pool with already set arch should not change",
			input: testNodePoolConfig{
				Name:       "worker",
				ServerSize: "cax31",
				Arch:       image.ArchX86, // explicitly set to x86, should not change
				Region:     "hel1",
			},
			expectedArch: image.ArchX86, // should remain x86
		},
		{
			name: "future ARM server type should be auto-detected",
			input: testNodePoolConfig{
				Name:       "worker",
				ServerSize: "cax99", // future server type
				Arch:       "",      // empty, should be auto-detected
				Region:     "hel1",
			},
			expectedArch: image.ArchARM,
		},
		{
			name: "future x86 server type should be auto-detected",
			input: testNodePoolConfig{
				Name:       "worker",
				ServerSize: "cx99", // future server type
				Arch:       "",     // empty, should be auto-detected
				Region:     "hel1",
			},
			expectedArch: image.ArchX86,
		},
		{
			name: "unknown server type should default to x86",
			input: testNodePoolConfig{
				Name:       "worker",
				ServerSize: "unknown123", // unknown server type
				Arch:       "",           // empty, should be auto-detected
				Region:     "hel1",
			},
			expectedArch: image.ArchX86,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the input that we can modify
			testInput := tt.input

			// Create mock struct level with the test input
			val := reflect.ValueOf(&testInput).Elem()
			mock := &mockStructLevel{current: val}

			// Call the validator
			ValidateAndSetArchForNodePool(mock)

			// Check that the arch was set correctly
			if testInput.Arch != tt.expectedArch {
				t.Errorf("Expected arch %s, got %s", tt.expectedArch, testInput.Arch)
			}
		})
	}
}
