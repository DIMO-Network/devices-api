package services

import (
	"testing"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/stretchr/testify/assert"
)

func Test_buildDeviceAttributes(t *testing.T) {
	existingDeviceAttrs := []*ddgrpc.DeviceTypeAttribute{
		{
			Name:  "mpg_city",
			Value: "15", // does not get set
		},
		{
			Name:  "mpg_highway",
			Value: "0", // does get set
		},
		{
			Name:  "mpg",
			Value: "", // does get set
		},
		{
			Name:  "driven_wheels",
			Value: "4",
		},
		{
			Name:  "something_else",
			Value: "dontsetme",
		},
	}
	// this object uses the drivly keys
	vinInfo := map[string]any{
		"mpgCity":             "14", // should not set even though different
		"mpgHighway":          "25",
		"mpg":                 "20",
		"msrpBase":            "35000",
		"fuelTankCapacityGal": 22.6,
		"fuel":                "gasoline",
		"wheelbase":           "220",
		"generation":          "2",
	}
	attributes := buildDeviceAttributes(existingDeviceAttrs, vinInfo)

	// assertions for each attribute
	assert.Len(t, attributes, 10)
	assert.Equal(t, "15", findAttribute(attributes, "mpg_city"))
	assert.Equal(t, "25", findAttribute(attributes, "mpg_highway"))
	assert.Equal(t, "20", findAttribute(attributes, "mpg"))
	assert.Equal(t, "35000", findAttribute(attributes, "base_msrp"))
	assert.Equal(t, "22.6", findAttribute(attributes, "fuel_tank_capacity_gal"))
	assert.Equal(t, "gasoline", findAttribute(attributes, "fuel_type"))
	assert.Equal(t, "220", findAttribute(attributes, "wheelbase"))
	assert.Equal(t, "2", findAttribute(attributes, "generation"))
	assert.Equal(t, "dontsetme", findAttribute(attributes, "something_else"))
	assert.Equal(t, "4", findAttribute(attributes, "driven_wheels"))
}

func findAttribute(attributes []*ddgrpc.DeviceTypeAttributeRequest, name string) string {
	for _, attribute := range attributes {
		if attribute.Name == name {
			return attribute.Value
		}
	}
	return ""
}
