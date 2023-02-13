package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindRegionForCountry(t *testing.T) {
	tests := []struct {
		name        string
		countryCode string
		want        *CountryInfo
	}{
		{
			name:        "find existing country",
			countryCode: "USA",
			want: &CountryInfo{
				Name:          "United States of America",
				Alpha3:        "USA",
				Alpha2:        "US",
				Region:        "Americas",
				SubRegion:     "Northern America",
				RegionCode:    19,
				SubRegionCode: 21,
			},
		},
		{
			name:        "non existing country returns nil",
			countryCode: "CAC",
			want:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, FindCountry(tt.countryCode), "FindCountry(%v)", tt.countryCode)
		})
	}
}
