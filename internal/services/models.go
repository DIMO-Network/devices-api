package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/volatiletech/null/v8"
)

// DeviceDefinition represents a device for to clients in generic form, ie. not specific to a user
type DeviceDefinition struct {
	DeviceDefinitionID string     `json:"deviceDefinitionId"`
	Name               string     `json:"name"`
	ImageURL           *string    `json:"imageUrl"`
	DeviceMake         DeviceMake `json:"make"`
	// CompatibleIntegrations has systems this vehicle can integrate with
	CompatibleIntegrations []DeviceCompatibility `json:"compatibleIntegrations"`
	Type                   DeviceType            `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo DeviceVehicleInfo `json:"vehicleData,omitempty"`
	// DeviceAttributes is a list of attributes for the device type as defined in device_types.properties
	DeviceAttributes []DeviceAttribute `json:"deviceAttributes,omitempty"`
	Metadata         interface{}       `json:"metadata"`
	Verified         bool              `json:"verified"`
}

type DeviceMake struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	LogoURL         null.String `json:"logo_url" swaggertype:"string"`
	OemPlatformName null.String `json:"oem_platform_name" swaggertype:"string"`
}

// DeviceCompatibility represents what systems we know this is compatible with
type DeviceCompatibility struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Style        string          `json:"style"`
	Vendor       string          `json:"vendor"`
	Region       string          `json:"region"`
	Country      string          `json:"country,omitempty"`
	Capabilities json.RawMessage `json:"capabilities"`
}

// DeviceType whether it is a vehicle or other type and basic information
type DeviceType struct {
	// Type is eg. Vehicle, E-bike, roomba
	Type      string   `json:"type"`
	Make      string   `json:"make"`
	Model     string   `json:"model"`
	Year      int      `json:"year"`
	SubModels []string `json:"subModels"`
}

// DeviceVehicleInfo represents some standard vehicle specific properties stored in the metadata json field in DB
type DeviceVehicleInfo struct {
	FuelType            string `json:"fuel_type,omitempty"`
	DrivenWheels        string `json:"driven_wheels,omitempty"`
	NumberOfDoors       string `json:"number_of_doors,omitempty"`
	BaseMSRP            int    `json:"base_msrp,omitempty"`
	EPAClass            string `json:"epa_class,omitempty"`
	VehicleType         string `json:"vehicle_type,omitempty"` // VehicleType PASSENGER CAR, from NHTSA
	MPGHighway          string `json:"mpg_highway,omitempty"`
	MPGCity             string `json:"mpg_city,omitempty"`
	FuelTankCapacityGal string `json:"fuel_tank_capacity_gal,omitempty"`
	MPG                 string `json:"mpg,omitempty"`
}

// DeviceAttribute represents some device type specific property stored in the metadata json field in DB
type DeviceAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Converters

// NewDeviceDefinitionFromNHTSA converts nhtsa response into our standard device definition struct
func NewDeviceDefinitionFromNHTSA(decodedVin *NHTSADecodeVINResponse) DeviceDefinition {
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
	dd.VehicleInfo = DeviceVehicleInfo{
		FuelType:      decodedVin.LookupValue("Fuel Type - Primary"),
		NumberOfDoors: decodedVin.LookupValue("Doors"),
		BaseMSRP:      msrp,
		VehicleType:   decodedVin.LookupValue("Vehicle Type"),
	}

	return dd
}

type PowertrainType string

const (
	ICE  PowertrainType = "ICE"
	HEV  PowertrainType = "HEV"
	PHEV PowertrainType = "PHEV"
	BEV  PowertrainType = "BEV"
	FCEV PowertrainType = "FCEV"
)

func (p PowertrainType) String() string {
	return string(p)
}

func (p *PowertrainType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Potentially an invalid value.
	switch bv := PowertrainType(s); bv {
	case ICE, HEV, PHEV, BEV, FCEV:
		*p = bv
		return nil
	default:
		return fmt.Errorf("unrecognized value: %s", s)
	}
}

// IntegrationsMetadata represents json stored in integrations table metadata jsonb column
type IntegrationsMetadata struct {
	AutoPiDefaultTemplateID      int                    `json:"autoPiDefaultTemplateId"`
	AutoPiPowertrainToTemplateID map[PowertrainType]int `json:"autoPiPowertrainToTemplateId,omitempty"`
}

// UserDeviceAPIIntegrationsMetadata represents json stored in user_device_api_integrations table metadata jsonb column
type UserDeviceAPIIntegrationsMetadata struct {
	AutoPiUnitID            *string                                    `json:"autoPiUnitId,omitempty"`
	AutoPiIMEI              *string                                    `json:"imei,omitempty"`
	AutoPiTemplateApplied   *int                                       `json:"autoPiTemplateApplied,omitempty"`
	AutoPiSubStatus         *string                                    `json:"autoPiSubStatus,omitempty"`
	AutoPiRegistrationError *string                                    `json:"autoPiRegistrationError,omitempty"`
	EnableTeslaLock         bool                                       `json:"enableTeslaLock,omitempty"`
	SmartcarEndpoints       []string                                   `json:"smartcarEndpoints,omitempty"`
	SmartcarUserID          *string                                    `json:"smartcarUserId,omitempty"`
	Commands                *UserDeviceAPIIntegrationsMetadataCommands `json:"commands,omitempty"`
	// CANProtocol is the protocol that was detected by edge-network from the autopi.
	CANProtocol     *string `json:"canProtocol,omitempty"`
	TeslaVehicleID  int     `json:"teslaVehicleId,omitempty"`
	TeslaAPIVersion int     `json:"teslaApiVersion,omitempty"`
}

type UserDeviceAPIIntegrationsMetadataCommands struct {
	Enabled []string `json:"enabled,omitempty"`
	Capable []string `json:"capable,omitempty"`
}

type UserDeviceMetadata struct {
	PowertrainType          *PowertrainType `json:"powertrainType,omitempty"`
	ElasticDefinitionSynced bool            `json:"elasticDefinitionSynced,omitempty"`
	ElasticRegionSynced     bool            `json:"elasticRegionSynced,omitempty"`
	PostalCode              *string         `json:"postal_code"`
	GeoDecodedCountry       *string         `json:"geoDecodedCountry"`
	GeoDecodedStateProv     *string         `json:"geoDecodedStateProv"`
	// CANProtocol is the protocol that was detected by edge-network from the autopi.
	CANProtocol *string `json:"canProtocol,omitempty"`
}

// AftermarketDeviceMetadata json metadata for table AftermarketDevice
type AftermarketDeviceMetadata struct {
	AutoPiDeviceID string `json:"autoPiDeviceId,omitempty"`
}

// todo: consider moving below to controllers and have service just return db object

// AutoPiCommandJob holds the autopi webhook jobs in a format for returning to clients
type AutoPiCommandJob struct {
	CommandJobID string               `json:"commandJobId"`
	CommandState string               `json:"commandState"`
	CommandRaw   string               `json:"commandRaw"`
	LastUpdated  *time.Time           `json:"lastUpdated"`
	Result       *AutoPiCommandResult `json:"result,omitempty"`
}

type ValuationDecodeCommand struct {
	VIN          string `json:"vin"`
	UserDeviceID string `json:"userDeviceId"`
}
