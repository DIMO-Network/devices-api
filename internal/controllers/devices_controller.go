package controllers

import (
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

func NewDeviceDefinitionFromGRPC(dd *grpc.GetDeviceDefinitionItemResponse) (services.DeviceDefinition, error) {
	if dd.Make == nil {
		return services.DeviceDefinition{}, errors.New("required DeviceMake relation is not set")
	}
	deviceAttributes := make([]services.DeviceAttribute, len(dd.DeviceAttributes))
	for i, attr := range dd.DeviceAttributes {
		deviceAttributes[i] = services.DeviceAttribute{
			Name:  attr.Name,
			Value: attr.Value,
		}
	}
	rp := services.DeviceDefinition{
		DeviceDefinitionID:     dd.NameSlug,
		Name:                   dd.Name,
		ImageURL:               &dd.ImageUrl,
		CompatibleIntegrations: []services.DeviceCompatibility{},
		DeviceMake: services.DeviceMake{
			ID:              dd.Make.Id,
			Name:            dd.Make.Name,
			LogoURL:         null.StringFrom(dd.Make.LogoUrl),
			OemPlatformName: null.StringFrom(dd.Make.OemPlatformName),
		},
		DeviceAttributes: deviceAttributes,
		Type: services.DeviceType{
			Type:  dd.Type.Type,
			Make:  dd.Type.Make,
			Model: dd.Type.Model,
			Year:  int(dd.Type.Year),
		},
		//Metadata: dd.Metadata,
		Verified: dd.Verified,
	}
	//// vehicle info
	//var vi map[string]services.DeviceVehicleInfo
	//rp.VehicleInfo = vi[vehicleInfoJSONNode]

	// compatible integrations
	rp.CompatibleIntegrations = DeviceCompatibilityFromDB(dd.DeviceIntegrations)
	// sub_models
	rp.Type.SubModels = dd.Type.SubModels

	return rp, nil
}

type DeviceRp struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

// DeviceCompatibilityFromDB returns list of compatibility representation from device integrations db slice, assumes integration relation loaded
func DeviceCompatibilityFromDB(dbDIS []*grpc.DeviceIntegration) []services.DeviceCompatibility {
	if len(dbDIS) == 0 {
		return []services.DeviceCompatibility{}
	}
	compatibilities := make([]services.DeviceCompatibility, len(dbDIS))
	for i, di := range dbDIS {
		compatibilities[i] = services.DeviceCompatibility{
			ID:     di.Integration.Id,
			Type:   di.Integration.Type,
			Style:  di.Integration.Style,
			Vendor: di.Integration.Vendor,
			Region: di.Region,
			//Capabilities: di.Capabilities,
		}
	}
	return compatibilities
}
