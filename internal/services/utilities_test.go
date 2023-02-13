package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAndCleanUUID(t *testing.T) {

	tests := []struct {
		name  string
		uuid  string
		want  bool
		want1 string
	}{
		{
			name:  "validate an autopi style uuid",
			uuid:  "ca6c89f7-ddc9-9b1a-5f1b-e0333ecd38e0",
			want:  true,
			want1: "ca6c89f7-ddc9-9b1a-5f1b-e0333ecd38e0",
		},
		{
			name:  "validate another autopi style uuid",
			uuid:  "5972e817-5302-f94b-6001-0597656157b6",
			want:  true,
			want1: "5972e817-5302-f94b-6001-0597656157b6",
		},
		{
			name:  "validate, lower case and trim otherwise valid uuid",
			uuid:  "Ca6c89f7-Ddc9-9b1a-5f1b-e0333ecd38e0 ",
			want:  true,
			want1: "ca6c89f7-ddc9-9b1a-5f1b-e0333ecd38e0",
		},
		{
			name:  "invalid uuid, too short",
			uuid:  "ca6c89f7-ddc9-9b1a-5f1b-e0333ecd38e",
			want:  false,
			want1: "",
		},
		{
			name:  "invalid uuid, bad character",
			uuid:  "ca6c89f7-ddc9-9b1a-5f1b-e0333ecd38e*",
			want:  false,
			want1: "",
		},
		{
			name:  "invalid uuid, empty",
			uuid:  "",
			want:  false,
			want1: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ValidateAndCleanUUID(tt.uuid)
			assert.Equalf(t, tt.want, got, "ValidateAndCleanUUID(%v)", tt.uuid)
			assert.Equalf(t, tt.want1, got1, "ValidateAndCleanUUID(%v)", tt.uuid)
		})
	}
}
