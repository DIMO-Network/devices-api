package services

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/DIMO-Network/shared"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -source device_definitions_service.go -destination mocks/device_definitions_service_mock.go
type DeviceDefinitionService interface {
	FindDeviceDefinitionByMMY(ctx context.Context, mk, model string, year int) (*ddgrpc.GetDeviceDefinitionItemResponse, error)
	GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*ddgrpc.DeviceMake, error)
	GetMakeByTokenID(ctx context.Context, tokenID *big.Int) (*ddgrpc.DeviceMake, error)
	GetIntegrations(ctx context.Context) ([]*ddgrpc.Integration, error)
	GetIntegrationByID(ctx context.Context, id string) (*ddgrpc.Integration, error)
	GetIntegrationByVendor(ctx context.Context, vendor string) (*ddgrpc.Integration, error)
	GetIntegrationByFilter(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error)
	CreateIntegration(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error)
	DecodeVIN(ctx context.Context, vin string, model string, year int, countryCode string) (*ddgrpc.DecodeVinResponse, error)
	GetIntegrationByTokenID(ctx context.Context, tokenID uint64) (*ddgrpc.Integration, error)
	GetDeviceStyleByID(ctx context.Context, id string) (*ddgrpc.DeviceStyle, error)
	// GetDeviceDefinitionBySlug get definition by new slug id. definitions api internally looks up info in tableland sqllite
	GetDeviceDefinitionBySlug(ctx context.Context, definitionID string) (*ddgrpc.GetDeviceDefinitionItemResponse, error)
	//go:generate mockgen -source device_definitions_service.go -destination ./device_definition_service_mock_test.go -package=services
}

type deviceDefinitionService struct {
	dbs                 func() *db.ReaderWriter
	log                 *zerolog.Logger
	definitionsGRPCAddr string
	googleMapsAPIKey    string
	identityAPI         IdentityAPI
}

func NewDeviceDefinitionService(DBS func() *db.ReaderWriter, log *zerolog.Logger, settings *config.Settings) DeviceDefinitionService {
	return &deviceDefinitionService{
		dbs:                 DBS,
		log:                 log,
		definitionsGRPCAddr: settings.DefinitionsGRPCAddr,
		googleMapsAPIKey:    settings.GoogleMapsAPIKey,
		identityAPI:         NewIdentityAPIService(log, settings),
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

func (d *deviceDefinitionService) DecodeVIN(ctx context.Context, vin string, model string, year int, countryCode string) (*ddgrpc.DecodeVinResponse, error) {
	if len(vin) < 13 || len(vin) > 17 {
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

// GetIntegrationByID get integration from grpc by NFT tokenID
func (d *deviceDefinitionService) GetIntegrationByTokenID(ctx context.Context, tokenID uint64) (*ddgrpc.Integration, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	integration, err := definitionsClient.GetIntegrationByTokenID(ctx, &ddgrpc.GetIntegrationByTokenIDRequest{TokenId: tokenID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc endpoint GetIntegrationByTokenID")
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

func (d *deviceDefinitionService) GetDeviceDefinitionBySlug(_ context.Context, definitionID string) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	if len(definitionID) == 0 {
		return nil, errors.New("Definition ID is required")
	}

	def, err := d.identityAPI.GetDefinition(strings.TrimSpace(definitionID))
	if err != nil {
		return nil, err
	}

	attrs := make([]*ddgrpc.DeviceTypeAttribute, len(def.Attributes))
	for i, attribute := range def.Attributes {
		attrs[i] = &ddgrpc.DeviceTypeAttribute{
			Name:  attribute.Name,
			Value: attribute.Value,
		}
	}

	return &ddgrpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: def.LegacyID,
		Name:               def.Manufacturer.Name + " " + def.Model + strconv.Itoa(def.Year),
		ImageUrl:           def.ImageURI,
		Verified:           true,
		DeviceIntegrations: nil,
		Make: &ddgrpc.DeviceMake{
			Name:     def.Manufacturer.Name,
			TokenId:  uint64(def.Manufacturer.TokenID),
			NameSlug: shared.SlugString(def.Manufacturer.Name),
		},
		DeviceAttributes:   attrs,
		HardwareTemplateId: "130",
		NameSlug:           definitionID,
		Year:               int32(def.Year),
		Model:              def.Model,
		Ksuid:              def.LegacyID,
		Id:                 definitionID,
	}, nil
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

const MilesToKmFactor = 1.609344 // there is 1.609 kilometers in a mile. const should probably be KmToMilesFactor
const EstMilesPerYear = 12000.0

type ValuationRequestData struct {
	Mileage *float64 `json:"mileage,omitempty"`
	ZipCode *string  `json:"zipCode,omitempty"`
}

type DataPullStatusEnum string

func (d *deviceDefinitionService) GetDeviceStyleByID(ctx context.Context, id string) (*ddgrpc.DeviceStyle, error) {
	definitionsClient, conn, err := d.getDeviceDefsGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ds, err := definitionsClient.GetDeviceStyleByID(ctx, &ddgrpc.GetDeviceStyleByIDRequest{
		Id: id,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to call grpc endpoint GetDeviceStyleByID")
	}

	return ds, nil
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

// getDeviceDefsGrpcClient instanties new connection with client to dd service. You must defer conn.close from returned connection
func (d *deviceDefinitionService) getDeviceDefsGrpcClient() (ddgrpc.DeviceDefinitionServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(d.definitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	return definitionsClient, conn, nil
}

func (d *deviceDefinitionService) getVINDecodeGrpcClient() (ddgrpc.VinDecoderServiceClient, *grpc.ClientConn, error) {
	// we may need to increase timeout for this request
	conn, err := grpc.NewClient(d.definitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	client := ddgrpc.NewVinDecoderServiceClient(conn)
	return client, conn, nil
}

func ConvertPowerTrainStringToPowertrain(value string) PowertrainType {
	switch value {
	case "HEV":
		return HEV
	case "PHEV":
		return PHEV
	case "BEV":
		return BEV
	case "ICE":
		return ICE
	case "FCEV":
		return FCEV
	default:
		return ICE
	}
}
