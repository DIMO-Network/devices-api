package registry

import (
	"reflect"
	"testing"
)

func TestNameHash(t *testing.T) {

	tests := []struct {
		name    string
		dcnName string
		want    string
	}{
		{
			name:    "hash name",
			dcnName: "reddy.dimo",
			want:    "0x4979a77f26d0ed69c59eb658798efd7b35e5d06a86b5b0e6ebc12dfdf4829260",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NameHash(tt.dcnName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NameHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

// todo test for GetExpiration
