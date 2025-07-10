package image

import "testing"

func strPtr(s string) *string {
	return &s
}

func TestNewTalosImageID(t *testing.T) {
	type args struct {
		args *TalosImageIDArgs
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default image id",
			args: args{args: &TalosImageIDArgs{}},
			want: talosImageIDHetznerDefault,
		},
		{
			name: "longhorn enabled",
			args: args{args: &TalosImageIDArgs{EnableLonghornSupport: true}},
			want: talosImageLonghorn,
		},
		{
			name: "overwrite image id",
			args: args{args: &TalosImageIDArgs{OverwriteTalosImageID: strPtr("custom-id-123")}},
			want: "custom-id-123",
		},
		{
			name: "overwrite image id takes precedence over longhorn",
			args: args{args: &TalosImageIDArgs{OverwriteTalosImageID: strPtr("custom-id-456"), EnableLonghornSupport: true}},
			want: "custom-id-456",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTalosImageID(tt.args.args); got != tt.want {
				t.Errorf("NewTalosImageID() = %v, want %v", got, tt.want)
			}
		})
	}
}
