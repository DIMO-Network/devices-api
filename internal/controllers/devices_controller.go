package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	qm "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DevicesController struct {
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
	NHTSASvc services.INHTSAService
	log      *zerolog.Logger
}

// NewDevicesController constructor
func NewDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger, nhtsaSvc services.INHTSAService) DevicesController {
	return DevicesController{
		Settings: settings,
		DBS:      dbs,
		NHTSASvc: nhtsaSvc,
		log:      logger,
	}
}

// LookupDeviceDefinitionByVIN godoc
// @Description decodes a VIN by first looking it up on our DB, and then calling out to external sources. If it does call out, it will backfill our DB
// @Tags 	device-definitions
// @Produce json
// @Param 	vin path string true "VIN eg. 5YJ3E1EA6MF873863"
// @Success 200 {object} controllers.DeviceDefinition
// @Router  /device-definitions/vin/{vin} [get]
func (d *DevicesController) LookupDeviceDefinitionByVIN(c *fiber.Ctx) error {
	vin := c.Params("vin")
	if len(vin) != 17 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_message": "vin is not 17 characters",
		})
	}
	squishVin := vin[:10]
	dd, err := models.DeviceDefinitions(
		qm.Where("vin_first_10 = ?", squishVin),
		qm.Where("verified = true"),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load("DeviceIntegrations.Integration")).
		One(c.Context(), d.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			decodedVIN, err := d.NHTSASvc.DecodeVIN(vin)
			if err != nil {
				return errorResponseHandler(c, err, fiber.StatusInternalServerError)
			}
			rp := NewDeviceDefinitionFromNHTSA(decodedVIN)
			// save to database, if error just log do not block, execute in go func routine to not block
			go func() {
				dbDevice := NewDbModelFromDeviceDefinition(rp, &squishVin)
				dbDevice.Verified = true
				dbDevice.Source = null.StringFrom("NHTSA")

				err = dbDevice.Insert(context.Background(), d.DBS().Writer, boil.Infer())
				if err != nil {
					d.log.Error().Err(err).Msg("error inserting device definition to db")
				}
			}()
			return c.JSON(fiber.Map{
				"device_definition": rp,
			})
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	rp := NewDeviceDefinitionFromDatabase(dd)
	return c.JSON(fiber.Map{
		"device_definition": rp,
	})
}

// GetAllDeviceMakeModelYears godoc
// @Description returns a json tree of Makes, models, and years
// @Tags 	device-definitions
// @Produce json
// @Success 200 {object} controllers.DeviceMMYRoot
// @Router  /device-definitions/all [get]
func (d *DevicesController) GetAllDeviceMakeModelYears(c *fiber.Ctx) error {
	all, err := models.DeviceDefinitions(qm.Where("verified = true")).All(c.Context(), d.DBS().Reader)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	var mmy []DeviceMMYRoot
	for _, dd := range all {
		idx := indexOfMake(mmy, dd.Make)
		// append make if not found
		if idx == -1 {
			mmy = append(mmy, DeviceMMYRoot{
				Make:   dd.Make,
				Models: []DeviceModels{{Model: dd.Model, Years: []DeviceModelYear{{Year: dd.Year, DeviceDefinitionID: dd.ID}}}},
			})
		} else {
			// attach model or year to existing make, lookup model
			idx2 := indexOfModel(mmy[idx].Models, dd.Model)
			if idx2 == -1 {
				// append model if not found
				mmy[idx].Models = append(mmy[idx].Models, DeviceModels{
					Model: dd.Model,
					Years: []DeviceModelYear{{Year: dd.Year, DeviceDefinitionID: dd.ID}},
				})
			} else {
				// make and model already found, just add year
				mmy[idx].Models[idx2].Years = append(mmy[idx].Models[idx2].Years, DeviceModelYear{Year: dd.Year, DeviceDefinitionID: dd.ID})
			}
		}
	}

	return c.JSON(fiber.Map{
		"makes": mmy,
	})
}

func (d *DevicesController) GetDeviceDefinitionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if len(id) != 27 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_message": "invalid device_definition_id",
		})
	}
	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load("DeviceIntegrations.Integration")).
		One(c.Context(), d.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	rp := NewDeviceDefinitionFromDatabase(dd)
	return c.JSON(fiber.Map{
		"device_definition": rp,
	})
}

func (d *DevicesController) GetIntegrationsByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if len(id) != 27 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_message": "invalid dvice definition id",
		})
	}
	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load("DeviceIntegrations.Integration")).
		One(c.Context(), d.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	// build object for integrations that have all the info
	var integrations []DeviceCompatibility
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			integrations = append(integrations, DeviceCompatibility{
				ID:           di.R.Integration.ID,
				Type:         di.R.Integration.Type,
				Style:        di.R.Integration.Style,
				Vendor:       di.R.Integration.Vendors,
				Country:      di.Country,
				Capabilities: string(di.Capabilities.JSON),
			})
		}
	}
	return c.JSON(fiber.Map{
		"compatible_integrations": integrations,
	})
}

func indexOfMake(makes []DeviceMMYRoot, make string) int {
	for i, root := range makes {
		if root.Make == make {
			return i
		}
	}
	return -1
}
func indexOfModel(models []DeviceModels, model string) int {
	for i, m := range models {
		if m.Model == model {
			return i
		}
	}
	return -1
}

const vehicleInfoJSONNode = "vehicle_info"

func NewDeviceDefinitionFromDatabase(dd *models.DeviceDefinition) DeviceDefinition {
	rp := DeviceDefinition{
		DeviceDefinitionID:     dd.ID,
		Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.Make, dd.Model),
		ImageURL:               "",
		CompatibleIntegrations: []DeviceCompatibility{},
		Type: DeviceType{
			Type:     "Vehicle",
			Make:     dd.Make,
			Model:    dd.Model,
			Year:     int(dd.Year),
			SubModel: dd.SubModel.String,
		},
		Metadata: string(dd.Metadata.JSON),
	}
	// vehicle info
	var vi map[string]services.DeviceVehicleInfo
	if err := dd.Metadata.Unmarshal(&vi); err == nil {
		rp.VehicleInfo = vi[vehicleInfoJSONNode]
	}
	// compatible integrations
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			rp.CompatibleIntegrations = append(rp.CompatibleIntegrations, DeviceCompatibility{
				ID:      di.R.Integration.ID,
				Type:    di.R.Integration.Type,
				Style:   di.R.Integration.Style,
				Vendor:  di.R.Integration.Vendors,
				Country: di.Country,
			})
		}
	}

	return rp
}

// NewDbModelFromDeviceDefinition converts a DeviceDefinition response object to a new database model for the given squishVin
func NewDbModelFromDeviceDefinition(dd DeviceDefinition, squishVin *string) *models.DeviceDefinition {
	dbDevice := models.DeviceDefinition{
		ID:         ksuid.New().String(),
		VinFirst10: null.StringFromPtr(squishVin),
		Make:       dd.Type.Make,
		Model:      dd.Type.Model,
		Year:       int16(dd.Type.Year),
		SubModel:   null.StringFrom(dd.Type.SubModel),
	}
	_ = dbDevice.Metadata.Marshal(map[string]interface{}{vehicleInfoJSONNode: dd.VehicleInfo})

	return &dbDevice
}

type DeviceRp struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

func NewDeviceDefinitionFromNHTSA(decodedVin *services.NHTSADecodeVINResponse) DeviceDefinition {
	dd := DeviceDefinition{}
	yr, _ := strconv.Atoi(decodedVin.LookupValue("Model Year"))
	msrp, _ := strconv.Atoi(decodedVin.LookupValue("Base Price ($)"))
	dd.Type = DeviceType{
		Type:  "Vehicle",
		Make:  decodedVin.LookupValue("Make"),
		Model: decodedVin.LookupValue("Model"),
		Year:  yr,
	}
	dd.Name = fmt.Sprintf("%d %s %s", dd.Type.Year, dd.Type.Make, dd.Type.Model)
	dd.VehicleInfo = services.DeviceVehicleInfo{
		FuelType:      decodedVin.LookupValue("Fuel Type - Primary"),
		NumberOfDoors: decodedVin.LookupValue("Doors"),
		BaseMSRP:      msrp,
		VehicleType:   decodedVin.LookupValue("Vehicle Type"),
	}

	return dd
}

// DeviceCompatibilityFromDB returns list of compatibility representation from device integrations db slice, assumes integration relation loaded
func DeviceCompatibilityFromDB(dbDIS models.DeviceIntegrationSlice) []DeviceCompatibility {
	var compatibilities []DeviceCompatibility
	if len(dbDIS) == 0 {
		return compatibilities
	}
	for _, di := range dbDIS {
		compatibilities = append(compatibilities, DeviceCompatibility{
			ID:           di.IntegrationID,
			Type:         di.R.Integration.Type,
			Style:        di.R.Integration.Style,
			Vendor:       di.R.Integration.Vendors,
			Country:      di.Country,
			Capabilities: string(di.Capabilities.JSON),
		})
	}
	return compatibilities
}

type DeviceDefinition struct {
	DeviceDefinitionID string `json:"device_definition_id"`
	Name               string `json:"name"`
	ImageURL           string `json:"image_url"`
	// CompatibleIntegrations has systems this vehicle can integrate with
	CompatibleIntegrations []DeviceCompatibility `json:"compatible_integrations"`
	Type                   DeviceType            `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo services.DeviceVehicleInfo `json:"vehicle_data,omitempty"`
	Metadata    interface{}                `json:"metadata"`
}

// DeviceCompatibility represents what systems we know this is compatible with
type DeviceCompatibility struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Style        string `json:"style"`
	Vendor       string `json:"vendor"`
	Country      string `json:"country"`
	Capabilities string `json:"capabilities,omitempty"`
}

// DeviceType whether it is a vehicle or other type and basic information
type DeviceType struct {
	// Type is eg. Vehicle, E-bike, roomba
	Type     string `json:"type"`
	Make     string `json:"make"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
	SubModel string `json:"sub_model"`
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
