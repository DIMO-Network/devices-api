package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -source device_definitions_service.go -destination mocks/device_definitions_service_mock.go

type DeviceDefinitionService interface {
	FindDeviceDefinitionByMMY(ctx context.Context, mk, model string, year int) (*ddgrpc.GetDeviceDefinitionItemResponse, error)
	UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error
	PullDrivlyData(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error)
	PullVincarioValuation(ctx context.Context, userDeiceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error)
	GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*ddgrpc.DeviceMake, error)
	GetMakeByTokenID(ctx context.Context, tokenID *big.Int) (*ddgrpc.DeviceMake, error)
	GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) ([]*ddgrpc.GetDeviceDefinitionItemResponse, error)
	GetDeviceDefinitionByID(ctx context.Context, id string) (*ddgrpc.GetDeviceDefinitionItemResponse, error)
	GetIntegrations(ctx context.Context) ([]*ddgrpc.Integration, error)
	GetIntegrationByID(ctx context.Context, id string) (*ddgrpc.Integration, error)
	GetIntegrationByVendor(ctx context.Context, vendor string) (*ddgrpc.Integration, error)
	GetIntegrationByFilter(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error)
	CreateIntegration(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error)
	DecodeVIN(ctx context.Context, vin string, model string, year int, countryCode string) (*ddgrpc.DecodeVinResponse, error)
}

type deviceDefinitionService struct {
	dbs                 func() *db.ReaderWriter
	drivlySvc           DrivlyAPIService
	vincarioSvc         VincarioAPIService
	log                 *zerolog.Logger
	nhtsaSvc            INHTSAService
	definitionsGRPCAddr string
	googleMapsAPIKey    string
}

func NewDeviceDefinitionService(DBS func() *db.ReaderWriter, log *zerolog.Logger, nhtsaService INHTSAService, settings *config.Settings) DeviceDefinitionService {
	return &deviceDefinitionService{
		dbs:                 DBS,
		log:                 log,
		nhtsaSvc:            nhtsaService,
		drivlySvc:           NewDrivlyAPIService(settings, DBS),
		definitionsGRPCAddr: settings.DefinitionsGRPCAddr,
		googleMapsAPIKey:    settings.GoogleMapsAPIKey,
		vincarioSvc:         NewVincarioAPIService(settings, log),
	}
}

func (d *deviceDefinitionService) CreateIntegration(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error) {

	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	integration, err := definitionsClient.CreateIntegration(ctx, &ddgrpc.CreateIntegrationRequest{
		Vendor: vendor,
		Type:   integrationType,
		Style:  style,
	})

	if err != nil {
		return nil, err
	}

	return &ddgrpc.Integration{Id: integration.Id, Vendor: vendor, Type: integrationType, Style: style}, nil
}

// GetDeviceDefinitionsByIDs calls device definitions api via GRPC to get the definition. idea for testing: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
// if not found or other error from server, the error contains the grpc status code that can be interpreted for different conditions. example in api.GrpcErrorToFiber
func (d *deviceDefinitionService) GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) ([]*ddgrpc.GetDeviceDefinitionItemResponse, error) {

	if len(ids) == 0 {
		return nil, errors.New("Device Definition Ids is required")
	}

	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	definitions, err2 := definitionsClient.GetDeviceDefinitionByID(ctx, &ddgrpc.GetDeviceDefinitionRequest{
		Ids: ids,
	})

	if err2 != nil {
		return nil, err2
	}

	return definitions.GetDeviceDefinitions(), nil
}

func (d *deviceDefinitionService) DecodeVIN(ctx context.Context, vin string, model string, year int, countryCode string) (*ddgrpc.DecodeVinResponse, error) {
	if len(vin) != 17 {
		return nil, errors.New("VIN must be 17 chars")
	}

	client, conn, err := d.getVINDecodeGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err2 := client.DecodeVin(ctx, &ddgrpc.DecodeVinRequest{
		Vin:        vin,
		KnownModel: model,
		KnownYear:  int32(year),
		Country:    countryCode,
	})

	if err2 != nil {
		return nil, err2
	}

	return resp, nil
}

// GetDeviceDefinitionByID is a helper for calling GetDeviceDefinitionsByIDs with one id.
func (d *deviceDefinitionService) GetDeviceDefinitionByID(ctx context.Context, id string) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	resp, err := d.GetDeviceDefinitionsByIDs(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, status.Error(codes.NotFound, "No definition with that id.")
	}

	return resp[0], nil
}

// GetIntegrations calls device definitions integrations api via GRPC to get the definition. idea for testing: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
func (d *deviceDefinitionService) GetIntegrations(ctx context.Context) ([]*ddgrpc.Integration, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	definitions, err := definitionsClient.GetIntegrations(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc endpoint GetIntegrations")
	}

	return definitions.GetIntegrations(), nil
}

// GetIntegrationByID get integration from grpc by id
func (d *deviceDefinitionService) GetIntegrationByID(ctx context.Context, id string) (*ddgrpc.Integration, error) {
	allIntegrations, err := d.GetIntegrations(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc to get integrations")
	}
	var integration *ddgrpc.Integration
	for _, in := range allIntegrations {
		if in.Id == id {
			integration = in
		}
	}
	if integration == nil {
		return nil, fmt.Errorf("no integration with id %s found in the %d existing", id, len(allIntegrations))
	}

	return integration, nil
}

func (d *deviceDefinitionService) GetIntegrationByFilter(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error) {
	allIntegrations, err := d.GetIntegrations(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc to get integrations")
	}
	var integration *ddgrpc.Integration
	for _, in := range allIntegrations {
		if in.Type == integrationType && in.Vendor == vendor && in.Style == style {
			integration = in
		}
	}
	if integration == nil {
		return nil, nil
	}

	return integration, nil
}

func (d *deviceDefinitionService) GetMakeByTokenID(ctx context.Context, tokenID *big.Int) (*ddgrpc.DeviceMake, error) {
	client, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return client.GetDeviceMakeByTokenID(ctx, &ddgrpc.GetDeviceMakeByTokenIdRequest{TokenId: tokenID.String()})
}

func (d *deviceDefinitionService) GetIntegrationByVendor(ctx context.Context, vendor string) (*ddgrpc.Integration, error) {
	allIntegrations, err := d.GetIntegrations(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc to get integrations")
	}
	var integration *ddgrpc.Integration
	for _, in := range allIntegrations {
		if in.Vendor == vendor {
			integration = in
		}
	}
	if integration == nil {
		return nil, fmt.Errorf("no integration with vendor %s found in the %d existing", vendor, len(allIntegrations))
	}

	return integration, nil
}

// FindDeviceDefinitionByMMY builds and execs query to find device definition for MMY, calling out via gRPC. Includes compatible integrations.
func (d *deviceDefinitionService) FindDeviceDefinitionByMMY(ctx context.Context, mk, model string, year int) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// question: does this load the integrations? it should
	dd, err := definitionsClient.GetDeviceDefinitionByMMY(ctx, &ddgrpc.GetDeviceDefinitionByMMYRequest{
		Make:  mk,
		Model: model,
		Year:  int32(year),
	})

	if err != nil {
		return nil, err
	}

	return dd, nil
}

// GetOrCreateMake gets the make from the db or creates it if not found. optional tx - if not passed in uses db writer
func (d *deviceDefinitionService) GetOrCreateMake(ctx context.Context, _ boil.ContextExecutor, makeName string) (*ddgrpc.DeviceMake, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// question: does this load the integrations? it should
	dm, err := definitionsClient.CreateDeviceMake(ctx, &ddgrpc.CreateDeviceMakeRequest{
		Name: makeName,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc endpoint CreateDeviceMake")
	}

	return &ddgrpc.DeviceMake{Id: dm.Id, Name: makeName}, nil
}

// UpdateDeviceDefinitionFromNHTSA (deprecated) pulls vin info from nhtsa, and updates the device definition metadata if the MMY from nhtsa matches ours, and the Source is not NHTSA verified
func (d *deviceDefinitionService) UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error {

	deviceDefinitionResponse, err := d.GetDeviceDefinitionsByIDs(ctx, []string{deviceDefinitionID})
	if err != nil {
		return err
	}

	if len(deviceDefinitionResponse) == 0 {
		return errors.New("Device definition empty")
	}

	dbDeviceDef := deviceDefinitionResponse[0]

	nhtsaDecode, err := d.nhtsaSvc.DecodeVIN(vin)
	if err != nil {
		return err
	}
	dd := NewDeviceDefinitionFromNHTSA(nhtsaDecode)
	if dd.Type.Make == dbDeviceDef.Make.Name && dd.Type.Model == dbDeviceDef.Type.Model && int16(dd.Type.Year) == int16(dbDeviceDef.Type.Year) {
		if idx := slices.IndexFunc(dbDeviceDef.ExternalIds, func(c *ddgrpc.ExternalID) bool { return c.Vendor == "NHTSA" }); !(dbDeviceDef.Verified && idx != -1) {
			definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
			if err != nil {
				return err
			}
			defer conn.Close()

			_, err = definitionsClient.UpdateDeviceDefinition(ctx, &ddgrpc.UpdateDeviceDefinitionRequest{
				DeviceDefinitionId: dbDeviceDef.DeviceDefinitionId,
				Verified:           true,
				Source:             "NHTSA",
				Year:               dbDeviceDef.Type.Year,
				Model:              dbDeviceDef.Type.Model,
				ImageUrl:           dbDeviceDef.ImageUrl,
			})

			if err != nil {
				return err
			}

		}
	} else {
		// just log for now if no MMY match.
		d.log.Warn().Msgf("No MMY match between deviceDefinitionID: %s and NHTSA for VIN: %s, %s", deviceDefinitionID, vin, dd.Name)
	}

	return nil
}

func (d *deviceDefinitionService) PullVincarioValuation(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error) {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return ErrorDataPullStatus, errors.Errorf("invalid VIN %s", vin)
	}

	// make sure userdevice exists
	ud, err := models.FindUserDevice(ctx, d.dbs().Reader, userDeviceID)
	if err != nil {
		return ErrorDataPullStatus, err
	}
	// do not pull for USA
	if strings.EqualFold(ud.CountryCode.String, "USA") {
		return SkippedDataPullStatus, nil
	}

	// check repull window
	existingPricingData, _ := models.ExternalVinData(
		models.ExternalVinDatumWhere.Vin.EQ(vin),
		models.ExternalVinDatumWhere.VincarioMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)

	// just return if already pulled recently for this VIN, but still need to insert never pulled vin - should be uncommon scenario
	if existingPricingData != nil && existingPricingData.UpdatedAt.Add(repullWindow).After(time.Now()) {
		return SkippedDataPullStatus, nil
	}

	externalVinData := &models.ExternalVinDatum{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(deviceDefinitionID),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}

	valuation, err := d.vincarioSvc.GetMarketValuation(vin)

	if err != nil {
		return ErrorDataPullStatus, errors.Wrap(err, "error pulling market data from vincario")
	}
	err = externalVinData.VincarioMetadata.Marshal(valuation)
	if err != nil {
		return ErrorDataPullStatus, errors.Wrap(err, "error marshalling vincario responset")
	}
	err = externalVinData.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return ErrorDataPullStatus, errors.Wrap(err, "error inserting external_vin_data for vincario")
	}

	return PulledValuationVincarioStatus, nil
}

const MilesToKmFactor = 1.609344 // there is 1.609 kilometers in a mile. const should probably be KmToMilesFactor
const EstMilesPerYear = 12000.0

type ValuationRequestData struct {
	Mileage *float64 `json:"mileage,omitempty"`
	ZipCode *string  `json:"zipCode,omitempty"`
}

type DataPullStatusEnum string

const (
	// PulledInfoAndValuationStatus means we pulled vin, edmunds, build and valuations
	PulledInfoAndValuationStatus DataPullStatusEnum = "PulledAll"
	// PulledValuationDrivlyStatus means we only pulled offers and pricing
	PulledValuationDrivlyStatus   DataPullStatusEnum = "PulledValuations"
	PulledValuationVincarioStatus DataPullStatusEnum = "PulledValuationVincario"
	SkippedDataPullStatus         DataPullStatusEnum = "Skipped"
	ErrorDataPullStatus           DataPullStatusEnum = "Error"
)

// PullDrivlyData pulls vin info from drivly, and inserts a record with the data.
// Will only pull if haven't in last 2 weeks. Does not re-pull VIN info, updates DD metadata, sets the device_style_id using the edmunds data pulled.
func (d *deviceDefinitionService) PullDrivlyData(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error) {
	const repullWindow = time.Hour * 24 * 14
	if len(vin) != 17 {
		return ErrorDataPullStatus, errors.Errorf("invalid VIN %s", vin)
	}

	deviceDef, err := d.GetDeviceDefinitionByID(ctx, deviceDefinitionID)
	if err != nil {
		return ErrorDataPullStatus, err
	}
	localLog := d.log.With().Str("vin", vin).Str("deviceDefinitionID", deviceDefinitionID).Logger()

	existingVINData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.Vin.EQ(vin),
		models.ExternalVinDatumWhere.VinMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)

	if err != nil {
		return ErrorDataPullStatus, err
	}

	// make sure userdevice exists
	ud, err := models.FindUserDevice(ctx, d.dbs().Reader, userDeviceID)
	if err != nil {
		return ErrorDataPullStatus, err
	}

	// by this point we know we might need to insert drivly raw json data
	externalVinData := &models.ExternalVinDatum{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(deviceDef.DeviceDefinitionId),
		Vin:                vin,
		UserDeviceID:       null.StringFrom(userDeviceID),
	}

	// should probably move this up to top as our check for never pulled, then seperate call to get latest pull date for repullWindow check
	if existingVINData != nil && existingVINData.VinMetadata.Valid {
		var vinInfo map[string]interface{}
		err = existingVINData.VinMetadata.Unmarshal(&vinInfo)
		if err != nil {
			return ErrorDataPullStatus, errors.Wrap(err, "unable to unmarshal vin metadata")
		}
		// update the device attributes via gRPC
		err2 := d.updateDeviceDefAttrs(ctx, deviceDef, vinInfo)
		if err2 != nil {
			return ErrorDataPullStatus, err2
		}
	}

	// determine if want to pull pricing data
	existingPricingData, _ := models.ExternalVinData(
		models.ExternalVinDatumWhere.Vin.EQ(vin),
		models.ExternalVinDatumWhere.PricingMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).
		One(context.Background(), d.dbs().Writer)
	// just return if already pulled recently for this VIN, but still need to insert never pulled vin - should be uncommon scenario
	if existingPricingData != nil && existingPricingData.UpdatedAt.Add(repullWindow).After(time.Now()) {
		localLog.Info().Msgf("already pulled pricing data for vin %s, skipping", vin)
		return SkippedDataPullStatus, nil
	}

	// get mileage for the drivly request
	deviceMileage, err := d.getDeviceMileage(userDeviceID, int(deviceDef.Type.Year))
	if err != nil {
		return ErrorDataPullStatus, err
	}

	reqData := ValuationRequestData{
		Mileage: deviceMileage,
	}

	udMD := new(UserDeviceMetadata)
	_ = ud.Metadata.Unmarshal(udMD)

	if udMD.PostalCode == nil {
		lat, long := d.getDeviceLatLong(userDeviceID)
		localLog.Info().Msgf("lat long found: %f, %f", lat, long)
		if lat != 0 && long != 0 {
			gl, err := GeoDecodeLatLong(lat, long, d.googleMapsAPIKey)
			if err != nil {
				localLog.Err(err).Msgf("failed to GeoDecode lat long %f, %f", lat, long)
			}
			if gl != nil {
				// update UD, ignore if fails doesn't matter
				udMD.PostalCode = &gl.PostalCode
				udMD.GeoDecodedCountry = &gl.Country
				udMD.GeoDecodedStateProv = &gl.AdminAreaLevel1
				_ = ud.Metadata.Marshal(udMD)
				_, err = ud.Update(ctx, d.dbs().Writer, boil.Whitelist(models.UserDeviceColumns.Metadata, models.UserDeviceColumns.UpdatedAt))
				if err != nil {
					localLog.Err(err).Msg("failed to update user_device.metadata with geodecode info")
				}
				localLog.Info().Msgf("GeoDecoded a lat long: %+v", gl)
			}
		}
	}

	if udMD.PostalCode != nil {
		reqData.ZipCode = udMD.PostalCode
	}
	_ = externalVinData.RequestMetadata.Marshal(reqData)

	// only pull offers and pricing on every pull.
	offer, err := d.drivlySvc.GetOffersByVIN(vin, &reqData)
	if err == nil {
		_ = externalVinData.OfferMetadata.Marshal(offer)
	}
	pricing, err := d.drivlySvc.GetVINPricing(vin, &reqData)
	if err == nil {
		_ = externalVinData.PricingMetadata.Marshal(pricing)
	}

	// check on edmunds data so we can get the style id
	edmundsExists, _ := models.ExternalVinData(models.ExternalVinDatumWhere.UserDeviceID.EQ(null.StringFrom(ud.ID)),
		models.ExternalVinDatumWhere.EdmundsMetadata.IsNotNull()).Exists(ctx, d.dbs().Reader)
	if !edmundsExists {
		// extra optional data that only needs to be pulled once.
		edmunds, err := d.drivlySvc.GetEdmundsByVIN(vin) // this is source data that will only be available after pulling vin + pricing
		if err == nil {
			_ = externalVinData.EdmundsMetadata.Marshal(edmunds)
		}
		// fill in edmunds style_id in our user_device if it exists and not already set. None of these seen as bad errors so just logs
		if edmunds != nil && ud.DeviceStyleID.IsZero() {
			d.setUserDeviceStyleFromEdmunds(ctx, edmunds, ud)
			localLog.Info().Msgf("set device_style_id for ud id %s", ud.ID)
		} else {
			localLog.Warn().Msgf("could not set edmunds style id. edmunds data exists: %v. ud style_id already set: %v", edmunds != nil, !ud.DeviceStyleID.IsZero())
		}
	}

	err = externalVinData.Insert(ctx, d.dbs().Writer, boil.Infer())
	if err != nil {
		return ErrorDataPullStatus, err
	}

	defer appmetrics.DrivlyIngestTotalOps.Inc()

	return PulledValuationDrivlyStatus, nil
}

func (d *deviceDefinitionService) updateDeviceDefAttrs(ctx context.Context, deviceDef *ddgrpc.GetDeviceDefinitionItemResponse, vinInfo map[string]any) error {
	deviceAttributes := buildDeviceAttributes(deviceDef.DeviceAttributes, vinInfo)

	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = definitionsClient.UpdateDeviceDefinition(ctx, &ddgrpc.UpdateDeviceDefinitionRequest{
		DeviceDefinitionId: deviceDef.DeviceDefinitionId,
		DeviceAttributes:   deviceAttributes,
	})
	if err != nil {
		return err
	}
	return nil
}

// buildDeviceAttributes returns list of set attributes based on what already exists and vinInfo pulled from drivly. based on a predetermined list
func buildDeviceAttributes(existingDeviceAttrs []*ddgrpc.DeviceTypeAttribute, vinInfo map[string]any) []*ddgrpc.DeviceTypeAttributeRequest {
	// TODO: replace seekAttributes with a better solution based on device_types.attributes
	seekAttributes := map[string]string{
		// {device attribute, must match device_types.properties}: {vin info from drivly}
		"mpg_city":               "mpgCity",
		"mpg_highway":            "mpgHighway",
		"mpg":                    "mpg",
		"base_msrp":              "msrpBase",
		"fuel_tank_capacity_gal": "fuelTankCapacityGal",
		"fuel_type":              "fuel",
		"wheelbase":              "wheelbase",
		"generation":             "generation",
		"number_of_doors":        "doors",
		"manufacturer_code":      "manufacturerCode",
		"driven_wheels":          "drive",
	}

	addedAttrCount := 0
	// build array of already present device_attributes and remove any already set satisfactorily from seekAttributes map
	var deviceAttributes []*ddgrpc.DeviceTypeAttributeRequest //nolint
	for _, attr := range existingDeviceAttrs {
		deviceAttributes = append(deviceAttributes, &ddgrpc.DeviceTypeAttributeRequest{
			Name:  attr.Name,
			Value: attr.Value,
		})
		// todo: 0 value attributes could be decimal form in string eg. 0.00000 . Convert value to int, and then compare to 0 again?
		if _, exists := seekAttributes[attr.Name]; exists && attr.Value != "" && attr.Value != "0" {
			// already set, no longer seeking it
			delete(seekAttributes, attr.Name)
		}
	}
	// iterate over remaining attributes
	for k, attr := range seekAttributes {
		if v, ok := vinInfo[attr]; ok && v != nil {
			val := fmt.Sprintf("%v", v)
			// lookup the existing device attribute and set it if exists
			existing := false
			for _, attribute := range deviceAttributes {
				if attribute.Name == k {
					attribute.Value = val
					existing = true
					break
				}
			}
			if !existing {
				deviceAttributes = append(deviceAttributes, &ddgrpc.DeviceTypeAttributeRequest{
					Name:  k,
					Value: val,
				})
			}
			addedAttrCount++
		}
	}
	if addedAttrCount == 0 {
		deviceAttributes = nil
	}
	return deviceAttributes
}

// setUserDeviceStyleFromEdmunds given edmunds json, sets the device style_id in the user_device per what edmunds says.
// If errors just logs and continues, since non critical
func (d *deviceDefinitionService) setUserDeviceStyleFromEdmunds(ctx context.Context, edmunds map[string]interface{}, ud *models.UserDevice) {
	edmundsJSON, err := json.Marshal(edmunds)
	if err != nil {
		d.log.Err(err).Msg("could not marshal edmunds response to json")
		return
	}
	styleIDResult := gjson.GetBytes(edmundsJSON, "edmundsStyle.data.style.id")
	styleID := styleIDResult.String()
	if styleIDResult.Exists() && len(styleID) > 0 {

		definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
		if err != nil {
			return
		}
		defer conn.Close()

		deviceStyle, err := definitionsClient.GetDeviceStyleByExternalID(ctx, &ddgrpc.GetDeviceStyleByIDRequest{
			Id: styleID,
		})

		if err != nil {
			d.log.Err(err).Msgf("unable to find device_style for edmunds style_id %s", styleID)
			return
		}
		ud.DeviceStyleID = null.StringFrom(deviceStyle.Id) // set foreign key
		_, err = ud.Update(ctx, d.dbs().Writer, boil.Whitelist("updated_at", "device_style_id"))
		if err != nil {
			d.log.Err(err).Msgf("unable to update user_device_id %s with styleID %s", ud.ID, deviceStyle.Id)
			return
		}
	}
}

// getDeviceDefsGrpcClient instanties new connection with client to dd service. You must defer conn.close from returned connection
func (d *deviceDefinitionService) getDeviceDefsGrpcClient() (ddgrpc.DeviceDefinitionServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(d.definitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	return definitionsClient, conn, nil
}

func (d *deviceDefinitionService) getVINDecodeGrpcClient() (ddgrpc.VinDecoderServiceClient, *grpc.ClientConn, error) {
	// we may need to increase timeout for this request
	conn, err := grpc.Dial(d.definitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	client := ddgrpc.NewVinDecoderServiceClient(conn)
	return client, conn, nil
}

func (d *deviceDefinitionService) getDeviceMileage(udID string, modelYear int) (mileage *float64, err error) {
	var deviceMileage *float64

	// Get user device odometer
	deviceData, err := models.UserDeviceData(
		models.UserDeviceDatumWhere.UserDeviceID.EQ(udID),
		models.UserDeviceDatumWhere.Data.IsNotNull(),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(context.Background(), d.dbs().Writer)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		deviceOdometer := gjson.GetBytes(deviceData.Data.JSON, "odometer")
		if deviceOdometer.Exists() {
			deviceMileage = new(float64)
			*deviceMileage = deviceOdometer.Float() / MilesToKmFactor
		}
	}

	// Estimate mileage based on model year
	if deviceMileage == nil {
		deviceMileage = new(float64)
		yearDiff := time.Now().Year() - modelYear
		switch {
		case yearDiff > 0:
			// Past model year
			*deviceMileage = float64(yearDiff) * EstMilesPerYear
		case yearDiff == 0:
			// Current model year
			*deviceMileage = EstMilesPerYear / 2
		default:
			// Next model year
			*deviceMileage = 0
		}
	}

	return deviceMileage, nil
}

func (d *deviceDefinitionService) getDeviceLatLong(userDeviceID string) (lat, long float64) {
	deviceData, err := models.UserDeviceData(
		models.UserDeviceDatumWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceDatumWhere.Data.IsNotNull(),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(context.Background(), d.dbs().Writer)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return
		}
	} else {
		latitude := gjson.GetBytes(deviceData.Data.JSON, "latitude")
		longitude := gjson.GetBytes(deviceData.Data.JSON, "longitude")
		if latitude.Exists() && longitude.Exists() {
			lat = latitude.Float()
			long = longitude.Float()
			return
		}
	}
	return
}
