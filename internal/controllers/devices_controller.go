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
		DeviceDefinitionID:     dd.DeviceDefinitionId,
		DefinitionID:           dd.Id,
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
			Type:  "Vehicle",
			Make:  dd.Make.Name,
			Model: dd.Model,
			Year:  int(dd.Year),
		},
		//Metadata: dd.Metadata,
		Verified: dd.Verified,
	}
	//// vehicle info
	//var vi map[string]services.DeviceVehicleInfo
	//rp.VehicleInfo = vi[vehicleInfoJSONNode]

	// sub_models
	for _, style := range dd.DeviceStyles {
		rp.Type.SubModels = append(rp.Type.SubModels, style.SubModel)
	}
	// temporary until mobile app stops using this stuff
	if rp.DeviceMake.Name == "Tesla" {
		// add only tesla
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Tesla", "Asia"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Tesla", "West Asia"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Tesla", "South America"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Tesla", "Oceania"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Tesla", "Europe"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Tesla", "Americas"))
	} else if rp.Type.Year > 2005 {
		// add hw options, Americas, USA, Europe
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("AutoPi", "Americas"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("AutoPi", "Europe"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Ruptela", "Americas"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Ruptela", "Europe"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Macaron", "Americas"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("Macaron", "Europe"))
	}
	if rp.DeviceMake.Name != "Tesla" && rp.Type.Year > 2018 {
		// add smartcar
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("SmartCar", "Europe"))
		rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, buildCompatibleIntegration("SmartCar", "Americas"))
	}

	return rp, nil
}

type DeviceRp struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

// buildCompatibleIntegration temporary until mobile app stops using this stuff
func buildCompatibleIntegration(vendor, region string) services.DeviceCompatibility {
	dc := services.DeviceCompatibility{}
	switch vendor {
	case "AutoPi":
		dc = services.DeviceCompatibility{
			ID:    "27qftVRWQYpVDcO5DltO5Ojbjxk",
			Type:  "Hardware",
			Style: "Addon",
		}
	case "SmartCar":
		dc = services.DeviceCompatibility{
			ID:    "22N2xaPOq2WW2gAHBHd0Ikn4Zob",
			Type:  "API",
			Style: "Webhook",
		}
	case "Tesla":
		dc = services.DeviceCompatibility{
			ID:    "26A5Dk3vvvQutjSyF0Jka2DP5lg",
			Type:  "API",
			Style: "OEM",
		}
	case "Ruptela":
		dc = services.DeviceCompatibility{
			ID:    "2lcaMFuCO0HJIUfdq8o780Kx5n3",
			Type:  "Hardware",
			Style: "Addon",
		}
	case "Macaron":
		dc = services.DeviceCompatibility{
			ID:    "2ULfuC8U9dOqRshZBAi0lMM1Rrx",
			Type:  "Hardware",
			Style: "Addon",
		}
	}
	dc.Vendor = vendor
	dc.Region = region

	return dc
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
