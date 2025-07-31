package validators

import (
	"reflect"
	"testing"

	validatorV10 "github.com/go-playground/validator/v10"
)

// Test structs that mimic the config structs without validation tags to avoid validation issues
type testHetznerConfig struct {
	Token string `json:"token"`
}

type testChartConfig struct {
	Enabled bool `json:"enabled"`
}

type testCSIChartConfig struct {
	testChartConfig
}

type testKubernetesConfig struct {
	HCloudToken       string              `json:"hcloud_token"`
	HetznerCCM        *testChartConfig    `json:"hetzner_ccm"`
	CSI               *testCSIChartConfig `json:"csi"`
	ClusterAutoScaler *testChartConfig    `json:"cluster_auto_scaler"`
}

type testPulumiConfig struct {
	Hetzner    testHetznerConfig    `json:"hetzner"`
	Kubernetes testKubernetesConfig `json:"kubernetes"`
}

// Mock StructLevel implementation for testing HCloud token validation
type mockStructLevelForHCloud struct {
	current    reflect.Value
	errorCount int
}

func (m *mockStructLevelForHCloud) Current() reflect.Value {
	return m.current
}

func (m *mockStructLevelForHCloud) Parent() reflect.Value {
	return reflect.Value{}
}

func (m *mockStructLevelForHCloud) Top() reflect.Value {
	return reflect.Value{}
}

func (m *mockStructLevelForHCloud) ExtractType(field reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return field, field.Kind(), true
}

func (m *mockStructLevelForHCloud) ReportError(value interface{}, namespace string, structNamespace string, tag string, param string) {
	m.errorCount++
}

func (m *mockStructLevelForHCloud) ReportValidationErrors(relativeNamespace string, relativeActualNamespace string, errs validatorV10.ValidationErrors) {
	m.errorCount++
}

func (m *mockStructLevelForHCloud) Validator() *validatorV10.Validate {
	return validatorV10.New()
}

func TestValidateHcloudToken(t *testing.T) {
	type args struct {
		cfg *testPulumiConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid - hcloud token provided",
			args: args{
				cfg: &testPulumiConfig{
					Kubernetes: testKubernetesConfig{
						HCloudToken: "test-token",
						HetznerCCM: &testChartConfig{
							Enabled: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - fallback to hetzner token when kubernetes token is empty",
			args: args{
				cfg: &testPulumiConfig{
					Hetzner: testHetznerConfig{
						Token: "hetzner-token",
					},
					Kubernetes: testKubernetesConfig{
						HCloudToken: "",
						HetznerCCM: &testChartConfig{
							Enabled: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - no token needed when no features enabled",
			args: args{
				cfg: &testPulumiConfig{
					Kubernetes: testKubernetesConfig{
						HCloudToken: "",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - no token needed when features are disabled",
			args: args{
				cfg: &testPulumiConfig{
					Kubernetes: testKubernetesConfig{
						HCloudToken: "",
						HetznerCCM: &testChartConfig{
							Enabled: false,
						},
						CSI: &testCSIChartConfig{
							testChartConfig: testChartConfig{
								Enabled: false,
							},
						},
						ClusterAutoScaler: &testChartConfig{
							Enabled: false,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid - no tokens available but HetznerCCM enabled",
			args: args{
				cfg: &testPulumiConfig{
					Hetzner: testHetznerConfig{
						Token: "",
					},
					Kubernetes: testKubernetesConfig{
						HCloudToken: "",
						HetznerCCM: &testChartConfig{
							Enabled: true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid - no tokens available but CSI enabled",
			args: args{
				cfg: &testPulumiConfig{
					Kubernetes: testKubernetesConfig{
						HCloudToken: "",
						CSI: &testCSIChartConfig{
							testChartConfig: testChartConfig{
								Enabled: true,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid - no tokens available but ClusterAutoScaler enabled",
			args: args{
				cfg: &testPulumiConfig{
					Kubernetes: testKubernetesConfig{
						HCloudToken: "",
						ClusterAutoScaler: &testChartConfig{
							Enabled: true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid - no tokens available but multiple features enabled",
			args: args{
				cfg: &testPulumiConfig{
					Kubernetes: testKubernetesConfig{
						HCloudToken: "",
						HetznerCCM: &testChartConfig{
							Enabled: true,
						},
						CSI: &testCSIChartConfig{
							testChartConfig: testChartConfig{
								Enabled: true,
							},
						},
						ClusterAutoScaler: &testChartConfig{
							Enabled: true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid - nil pointers for optional features",
			args: args{
				cfg: &testPulumiConfig{
					Kubernetes: testKubernetesConfig{
						HCloudToken:       "",
						HetznerCCM:        nil,
						CSI:               nil,
						ClusterAutoScaler: nil,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock StructLevel to test the validator function directly
			mock := &mockStructLevelForHCloud{
				current:    reflect.ValueOf(*tt.args.cfg),
				errorCount: 0,
			}

			// Call the validator function directly
			ValidateHcloudToken(mock)

			hasError := mock.errorCount > 0
			if tt.wantErr != hasError {
				t.Errorf("ValidateHcloudToken() error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}

func Test_hasValidHetznerToken(t *testing.T) {
	type args struct {
		current reflect.Value
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid - hetzner token exists",
			args: args{
				current: reflect.ValueOf(testPulumiConfig{
					Hetzner: testHetznerConfig{
						Token: "valid-token",
					},
				}),
			},
			want: true,
		},
		{
			name: "invalid - hetzner token is empty",
			args: args{
				current: reflect.ValueOf(testPulumiConfig{
					Hetzner: testHetznerConfig{
						Token: "",
					},
				}),
			},
			want: false,
		},
		{
			name: "invalid - hetzner field missing",
			args: args{
				current: reflect.ValueOf(struct {
					Kubernetes testKubernetesConfig `json:"kubernetes"`
				}{
					Kubernetes: testKubernetesConfig{
						HCloudToken: "test-token",
					},
				}),
			},
			want: false,
		},
		{
			name: "invalid - hetzner token field missing",
			args: args{
				current: reflect.ValueOf(struct {
					Hetzner struct {
						SomeOtherField string `json:"other"`
					} `json:"hetzner"`
				}{
					Hetzner: struct {
						SomeOtherField string `json:"other"`
					}{
						SomeOtherField: "value",
					},
				}),
			},
			want: false,
		},
		{
			name: "invalid - empty struct",
			args: args{
				current: reflect.ValueOf(struct{}{}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasValidHetznerToken(tt.args.current); got != tt.want {
				t.Errorf("hasValidHetznerToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
