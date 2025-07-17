package image

import "testing"

func TestDetectRequiredArchitecturesFromList(t *testing.T) {
	type args struct {
		architectures []CPUArchitecture
	}
	tests := []struct {
		name          string
		args          args
		wantEnableARM bool
		wantEnableX86 bool
	}{
		{
			name: "empty list",
			args: args{
				architectures: []CPUArchitecture{},
			},
			wantEnableARM: false,
			wantEnableX86: false,
		},
		{
			name: "only ARM architecture",
			args: args{
				architectures: []CPUArchitecture{ArchARM},
			},
			wantEnableARM: true,
			wantEnableX86: false,
		},
		{
			name: "only x86 architecture",
			args: args{
				architectures: []CPUArchitecture{ArchX86},
			},
			wantEnableARM: false,
			wantEnableX86: true,
		},
		{
			name: "both architectures",
			args: args{
				architectures: []CPUArchitecture{ArchARM, ArchX86},
			},
			wantEnableARM: true,
			wantEnableX86: true,
		},
		{
			name: "multiple ARM entries",
			args: args{
				architectures: []CPUArchitecture{ArchARM, ArchARM, ArchARM},
			},
			wantEnableARM: true,
			wantEnableX86: false,
		},
		{
			name: "multiple x86 entries",
			args: args{
				architectures: []CPUArchitecture{ArchX86, ArchX86},
			},
			wantEnableARM: false,
			wantEnableX86: true,
		},
		{
			name: "mixed architectures with duplicates",
			args: args{
				architectures: []CPUArchitecture{ArchARM, ArchX86, ArchARM, ArchX86},
			},
			wantEnableARM: true,
			wantEnableX86: true,
		},
		{
			name: "unknown architecture (should not enable either)",
			args: args{
				architectures: []CPUArchitecture{CPUArchitecture("unknown")},
			},
			wantEnableARM: false,
			wantEnableX86: false,
		},
		{
			name: "mixed with unknown architecture",
			args: args{
				architectures: []CPUArchitecture{ArchARM, CPUArchitecture("unknown"), ArchX86},
			},
			wantEnableARM: true,
			wantEnableX86: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnableARM, gotEnableX86 := DetectRequiredArchitecturesFromList(tt.args.architectures)
			if gotEnableARM != tt.wantEnableARM {
				t.Errorf("DetectRequiredArchitecturesFromList() gotEnableARM = %v, want %v", gotEnableARM, tt.wantEnableARM)
			}
			if gotEnableX86 != tt.wantEnableX86 {
				t.Errorf("DetectRequiredArchitecturesFromList() gotEnableX86 = %v, want %v", gotEnableX86, tt.wantEnableX86)
			}
		})
	}
}
