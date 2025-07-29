package validators

import (
	"reflect"
	"testing"

	"github.com/exivity/pulumi-hcloud-k8s/pkg/talos/image"
)

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
			name: "cx22 should return x86",
			args: args{serverSize: "cx22"},
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
