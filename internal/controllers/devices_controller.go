package controllers

import (
	"fmt"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared/db"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

type DevicesController struct {
	settings        *config.Settings
	dbs             func() *db.ReaderWriter
	nhtsaSvc        services.INHTSAService
	deviceDefSvc    services.DeviceDefinitionService
	deviceDefIntSvc services.DeviceDefinitionIntegrationService
	log             *zerolog.Logger
}

// NewDevicesController constructor
func NewDevicesController(settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, nhtsaSvc services.INHTSAService, ddSvc services.DeviceDefinitionService, ddIntSvc services.DeviceDefinitionIntegrationService) DevicesController {

	return DevicesController{
		settings:        settings,
		dbs:             dbs,
		nhtsaSvc:        nhtsaSvc,
		log:             logger,
		deviceDefSvc:    ddSvc,
		deviceDefIntSvc: ddIntSvc,
	}
}

// GetDeviceDefinitionByID godoc
// @Description gets a specific device definition by id, adds autopi integration on the fly if does not have it and year > cutoff
// @Tags        device-definitions
// @Produce     json
// @Param       id  path     string true "device definition id, KSUID format"
// @Success     200 {object} services.DeviceDefinition
// @Router      /device-definitions/:id [get]
func (d *DevicesController) GetDeviceDefinitionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if len(id) != 27 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "invalid device_definition_id",
		})
	}
	deviceDefinitionResponse, err := d.deviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{id})
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+id)
	}

	if len(deviceDefinitionResponse) == 0 {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("device definition with id %s not found", id))
	}

	dd := deviceDefinitionResponse[0]
	rp, err := NewDeviceDefinitionFromGRPC(dd)
	if err != nil {
		return errors.Wrapf(err, "could not convert device def for api response %+v", dd)
	}

	return c.JSON(fiber.Map{
		"deviceDefinition": rp,
	})
}

// GetDeviceIntegrationsByID godoc
// @Description gets all the available integrations for a device definition. Includes the capabilities of the device with the integration
// @Tags        device-definitions
// @Produce     json
// @Param       id  path     string true "device definition id, KSUID format"
// @Success     200 {object} []services.DeviceCompatibility
// @Router      /device-definitions/{id}/integrations [get]
func (d *DevicesController) GetDeviceIntegrationsByID(c *fiber.Ctx) error {
	// todo check if this method is even still used
	id := c.Params("id")
	if len(id) != 27 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "invalid device definition id",
		})
	}

	deviceDefinitionResponse, err := d.deviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{id})

	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to get definition id: "+id)
	}

	dd := deviceDefinitionResponse[0]
	// build object for integrations that have all the info
	integrations := make([]services.DeviceCompatibility, len(dd.DeviceIntegrations))
	for i, di := range dd.DeviceIntegrations {
		integrations[i] = services.DeviceCompatibility{
			ID:     di.Integration.Id,
			Type:   di.Integration.Type,
			Style:  di.Integration.Style,
			Vendor: di.Integration.Vendor,
			Region: di.Region,
		}
	}

	return c.JSON(fiber.Map{
		"compatibleIntegrations": integrations,
	})
}

// GetDeviceDefinitionByMMY godoc
// @Description gets a specific device definition by make model and year
// @Tags        device-definitions
// @Produce     json
// @Param       make  query    string true "make eg TESLA"
// @Param       model query    string true "model eg MODEL Y"
// @Param       year  query    string true "year eg 2021"
// @Success     200   {object} services.DeviceDefinition
// @Router      /device-definitions [get]
func (d *DevicesController) GetDeviceDefinitionByMMY(c *fiber.Ctx) error {
	mk := c.Query("make")
	model := c.Query("model")
	year := c.Query("year")
	if mk == "" || model == "" || year == "" {
		return helpers.ErrorResponseHandler(c, errors.New("make, model, and year are required"), fiber.StatusBadRequest)
	}
	yrInt, err := strconv.Atoi(year)
	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	dd, err := d.deviceDefSvc.FindDeviceDefinitionByMMY(c.Context(), mk, model, yrInt)

	if err != nil {
		return helpers.GrpcErrorToFiber(err, fmt.Sprintf("device with %s %s %s failed", mk, model, year))
	}

	// sometimes dd can empty nil.
	if dd == nil {
		return helpers.ErrorResponseHandler(c, errors.Wrapf(err, "device with %s %s %s not found", mk, model, year), fiber.StatusNotFound)
	}

	rp, err := NewDeviceDefinitionFromGRPC(dd)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"deviceDefinition": rp,
	})
}

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

type DeviceMMYRoot struct {
	Make   string         `json:"make"`
	Models []DeviceModels `json:"models"`
}

type DeviceModels struct {
	Model string            `json:"model"`
	Years []DeviceModelYear `json:"years"`
}

type DeviceModelYear struct {
	Year               int16  `json:"year"`
	DeviceDefinitionID string `json:"id"`
}
