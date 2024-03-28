package rpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared"

	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/devices-api/internal/services/issuer"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
)

func NewUserDeviceRPCService(
	dbs func() *db.ReaderWriter,
	settings *config.Settings,
	hardwareTemplateService autopi.HardwareTemplateService,
	logger *zerolog.Logger,
	deviceDefSvc services.DeviceDefinitionService,
	eventService services.EventService,
	vcIss *issuer.Issuer,
	userDeviceService services.UserDeviceService,
	teslaTaskService services.TeslaTaskService,
	smartcarTaskSvc services.SmartcarTaskService,
) pb.UserDeviceServiceServer {
	return &userDeviceRPCServer{dbs: dbs,
		logger:                  logger,
		settings:                settings,
		hardwareTemplateService: hardwareTemplateService,
		deviceDefSvc:            deviceDefSvc,
		eventService:            eventService,
		vcIss:                   vcIss,
		userDeviceSvc:           userDeviceService,
		teslaTaskService:        teslaTaskService,
		smartcarTaskSvc:         smartcarTaskSvc,
	}
}

// userDeviceRPCServer is the grpc server implementation for the proto services
type userDeviceRPCServer struct {
	pb.UnimplementedUserDeviceServiceServer
	dbs                     func() *db.ReaderWriter
	hardwareTemplateService autopi.HardwareTemplateService
	logger                  *zerolog.Logger
	settings                *config.Settings
	deviceDefSvc            services.DeviceDefinitionService
	eventService            services.EventService
	vcIss                   *issuer.Issuer
	userDeviceSvc           services.UserDeviceService
	teslaTaskService        services.TeslaTaskService
	smartcarTaskSvc         services.SmartcarTaskService
}

func (s *userDeviceRPCServer) GetUserDevice(ctx context.Context, req *pb.GetUserDeviceRequest) (*pb.UserDevice, error) {
	dbDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(req.Id),
		qm.Load(
			qm.Rels(models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.VehicleTokenAftermarketDevice),
		),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(
			qm.Rels(
				models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.Claim,
			),
		),
		qm.Load(
			qm.Rels(models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.VehicleTokenSyntheticDevice),
		),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No device with that ID found.")
		}
		s.logger.Err(err).Str("userDeviceId", req.Id).Msg("Database failure retrieving device.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	return s.deviceModelToAPI(dbDevice), nil
}

func (s *userDeviceRPCServer) GetUserDeviceByVIN(ctx context.Context, req *pb.GetUserDeviceByVINRequest) (*pb.UserDevice, error) {
	dbDevice, err := models.UserDevices(
		models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(req.Vin)),
		qm.OrderBy("vin_confirmed desc"),
		qm.Load(
			qm.Rels(models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.VehicleTokenAftermarketDevice),
		),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(
			qm.Rels(
				models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.Claim,
			),
		),
		qm.Load(
			qm.Rels(models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.VehicleTokenSyntheticDevice),
		),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No device with that VIN found.")
		}
		s.logger.Err(err).Str("userDeviceVIN", req.Vin).Msg("Database failure retrieving device by VIN.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	return s.deviceModelToAPI(dbDevice), nil
}

func (s *userDeviceRPCServer) GetUserDeviceByEthAddr(ctx context.Context, req *pb.GetUserDeviceByEthAddrRequest) (*pb.UserDevice, error) {

	aftermarketDevice, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.EthereumAddress.EQ(req.EthAddr),
		qm.Load(models.AftermarketDeviceRels.VehicleToken),
	).One(ctx, s.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No AftermarketDevice found for the given Ethereum address")
		}
		s.logger.Err(err).Msg("Error finding AftermarketDevice")
		return nil, status.Error(codes.Internal, "Failed to fetch AftermarketDevice.")
	}

	if aftermarketDevice.R == nil || aftermarketDevice.R.VehicleToken == nil {
		return nil, status.Error(codes.NotFound, "No VehicleToken associated with the AftermarketDevice")
	}

	userDeviceID := aftermarketDevice.R.VehicleToken.UserDeviceID
	if !userDeviceID.Valid {
		return nil, status.Error(codes.NotFound, "No UserDeviceID found in VehicleNFT")
	}

	dbDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID.String),
		qm.Load(
			qm.Rels(models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.VehicleTokenAftermarketDevice),
		),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(
			qm.Rels(
				models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.Claim,
			),
		),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenSyntheticDevice)),
	).One(ctx, s.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No device with that ID found.")
		}
		s.logger.Err(err).Str("userDeviceId", userDeviceID.String).Msg("Database failure retrieving device.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	return s.deviceModelToAPI(dbDevice), nil
}

func (s *userDeviceRPCServer) GetUserDeviceByTokenId(ctx context.Context, req *pb.GetUserDeviceByTokenIdRequest) (*pb.UserDevice, error) { // nolint

	tknID := types.NewNullDecimal(decimal.New(req.TokenId, 0))

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tknID),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No device with that token ID found.")
		}
		s.logger.Err(err).Int64("tokenID", req.TokenId).Msg("Database failure retrieving device.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	out := &pb.UserDevice{
		Id:           nft.UserDeviceID.String,
		TokenId:      s.toUint64(nft.TokenID),
		OwnerAddress: nft.OwnerAddress.Bytes,
	}

	return out, nil
}

func (s *userDeviceRPCServer) ListUserDevicesForUser(ctx context.Context, req *pb.ListUserDevicesForUserRequest) (*pb.ListUserDevicesForUserResponse, error) {
	var query []qm.QueryMod

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing userID paramter")
	}

	if req.EthereumAddress == "" {
		query = []qm.QueryMod{
			models.UserDeviceWhere.UserID.EQ(req.UserId),
		}
	} else {
		query = []qm.QueryMod{
			qm.LeftOuterJoin("devices_api." + models.TableNames.VehicleNFTS + " ON " + models.VehicleNFTTableColumns.UserDeviceID + " = " + models.UserDeviceTableColumns.ID),
			models.UserDeviceWhere.UserID.EQ(req.UserId),
			qm.Or2(models.VehicleNFTWhere.OwnerAddress.EQ(null.BytesFrom(common.HexToAddress(req.EthereumAddress).Bytes()))),
		}
	}

	query = append(query,
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenAftermarketDevice)),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenSyntheticDevice)),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.OrderBy(models.UserDeviceTableColumns.CreatedAt+" DESC"),
	)

	devices, err := models.UserDevices(query...).All(ctx, s.dbs().Reader)
	if err != nil {
		s.logger.Err(err).Str("userId", req.UserId).Msg("Database failure retrieving user's devices.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	out := make([]*pb.UserDevice, len(devices))

	for i := 0; i < len(devices); i++ {
		out[i] = s.deviceModelToAPI(devices[i])
	}

	return &pb.ListUserDevicesForUserResponse{UserDevices: out}, nil
}

func (s *userDeviceRPCServer) ApplyHardwareTemplate(ctx context.Context, req *pb.ApplyHardwareTemplateRequest) (*pb.ApplyHardwareTemplateResponse, error) {
	resp, err := s.hardwareTemplateService.ApplyHardwareTemplate(ctx, req)
	if err != nil {
		s.logger.Err(err).Str("autopi_unit_id", req.AutoApiUnitId).Str("user_device_id", req.UserDeviceId).Msgf("failed to apply hardware template id %s", req.HardwareTemplateId)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, err
}

func (s *userDeviceRPCServer) CreateTemplate(_ context.Context, req *pb.CreateTemplateRequest) (*pb.CreateTemplateResponse, error) {
	resp, err := s.hardwareTemplateService.CreateTemplate(req)
	if err != nil {
		s.logger.Err(err).Str("template name", req.Name).Msgf("failed to create template %s", req.Name)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, err
}

// RegisterUserDeviceFromVIN used from admin to add a vehicle to a user when we can't get the VIN via hardware
func (s *userDeviceRPCServer) RegisterUserDeviceFromVIN(ctx context.Context, req *pb.RegisterUserDeviceFromVINRequest) (*pb.RegisterUserDeviceFromVINResponse, error) {
	country := constants.FindCountry(req.CountryCode)
	if country == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid countryCode field or country not supported: %s", req.CountryCode)
	}
	// check for duplicate vin, future: refactor with user_devices_controler fromsmartcar, fromvin
	vin := strings.ToUpper(req.Vin)

	hasConflict := false

	if s.settings.IsProduction() {
		conflict, err := models.UserDevices(
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(ctx, s.dbs().Reader)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Internal, err.Error())
		}

		hasConflict = conflict
	}

	if hasConflict {
		return nil, status.Errorf(codes.AlreadyExists, "VIN %s in use by a previously connected device", vin)
	}

	resp, err := s.deviceDefSvc.DecodeVIN(ctx, vin, "", 0, "")

	if resp.Year < 2008 {
		s.logger.Warn().
			Str("vin", vin).
			Str("year", fmt.Sprint(resp.Year)).
			Str("user_id", req.UserDeviceId).
			Msg("VIN is too old")

		return nil, status.Errorf(codes.InvalidArgument, "VIN %s from year %v is too old", vin, resp.Year)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(resp.DeviceDefinitionId) == 0 {
		s.logger.Warn().
			Str("vin", vin).
			Str("user_id", req.UserDeviceId).
			Msg("unable to decode vin for customer request to create vehicle")
		return nil, status.Error(codes.Internal, "Unable to decode VIN")
	}

	// attach device def to user
	dd, err := s.deviceDefSvc.GetDeviceDefinitionByID(ctx, resp.DeviceDefinitionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tx, err := s.dbs().Writer.DB.BeginTx(ctx, nil)
	defer tx.Rollback() // nolint
	if err != nil {
		return nil, err
	}

	_, _, err = s.userDeviceSvc.CreateUserDevice(ctx, dd.DeviceDefinitionId, resp.DeviceStyleId, req.CountryCode, req.UserDeviceId, &vin, nil, req.VinConfirmed)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterUserDeviceFromVINResponse{Created: true}, err
}

func (s *userDeviceRPCServer) UpdateDeviceIntegrationStatus(ctx context.Context, req *pb.UpdateDeviceIntegrationStatusRequest) (*pb.UserDevice, error) {
	logger := s.logger.With().Str("integrationId", req.IntegrationId).Str("userDeviceId", req.UserDeviceId).Logger()

	apiIntegration, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(req.IntegrationId),
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(req.UserDeviceId),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		logger.Err(err).Msg("Couldn't retrieve integration.")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No UserDeviceAPIIntegrations with that ID found.")
		}
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	if req.Status != apiIntegration.Status {
		apiIntegration.Status = req.Status
		if _, err := apiIntegration.Update(ctx, s.dbs().Writer, boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Status)); err != nil {
			logger.Info().Msgf("Failed to update integration status to %s.", req.Status)
			return nil, status.Error(codes.Internal, "failed to update API integration")
		}
		logger.Info().Msgf("Updated integration status to %s.", req.Status)
	}

	return s.GetUserDevice(ctx, &pb.GetUserDeviceRequest{Id: req.UserDeviceId})
}

//nolint:all
func (s *userDeviceRPCServer) GetUserDeviceByAutoPIUnitId(ctx context.Context, req *pb.GetUserDeviceByAutoPIUnitIdRequest) (*pb.UserDeviceAutoPIUnitResponse, error) {
	dbDevice, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.Serial.EQ(null.StringFrom(req.Id)),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No UserDeviceAPIIntegrations with that ID found.")
		}
		s.logger.Err(err).Str("autoPIUnitId", req.Id).Msg("Database failure retrieving UserDeviceAPIIntegrations.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	result := &pb.UserDeviceAutoPIUnitResponse{
		UserDeviceId:       dbDevice.UserDeviceID,
		DeviceDefinitionId: dbDevice.R.UserDevice.DeviceDefinitionID,
		UserId:             dbDevice.R.UserDevice.UserID,
	}

	if dbDevice.R.UserDevice.DeviceStyleID.Valid {
		result.DeviceStyleId = dbDevice.R.UserDevice.DeviceStyleID.String
	}

	return result, nil
}

func (s *userDeviceRPCServer) GetAllUserDevice(req *pb.GetAllUserDeviceRequest, stream pb.UserDeviceService_GetAllUserDeviceServer) error {
	ctx := context.Background()
	all, err := models.UserDevices(
		models.UserDeviceWhere.VinConfirmed.EQ(true),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenSyntheticDevice)),
	).All(ctx, s.dbs().Reader)
	if err != nil {
		s.logger.Err(err).Msg("Database failure retrieving all user devices.")
		return status.Error(codes.Internal, "Internal error.")
	}

	if len(req.Wmi) == 3 {
		wmi := strings.ToUpper(req.Wmi)
		s.logger.Info().Msgf("WMI filter set: %s", wmi)
		filtered := models.UserDeviceSlice{}
		for _, device := range all {
			if len(device.VinIdentifier.String) > 3 && device.VinIdentifier.String[:3] == wmi {
				filtered = append(filtered, device)
			}
		}
		all = filtered
	}

	for _, item := range all {
		if err := stream.Send(s.deviceModelToAPI(item)); err != nil {
			return err
		}
	}

	return nil
}

func decimalToUint(x types.Decimal) uint64 {
	y, _ := x.Uint64()
	return y
}

func (s *userDeviceRPCServer) deviceModelToAPI(ud *models.UserDevice) *pb.UserDevice {
	out := &pb.UserDevice{
		Id:                 ud.ID,
		UserId:             ud.UserID,
		DeviceDefinitionId: ud.DeviceDefinitionID,
		DeviceStyleId:      ud.DeviceStyleID.Ptr(),
		OptedInAt:          nullTimeToPB(ud.OptedInAt),
		Integrations:       make([]*pb.UserDeviceIntegration, len(ud.R.UserDeviceAPIIntegrations)),
		VinConfirmed:       ud.VinConfirmed,
	}

	if vnft := ud.R.VehicleNFT; vnft != nil {
		out.TokenId = s.toUint64(vnft.TokenID)
		if vnft.OwnerAddress.Valid {
			out.OwnerAddress = vnft.OwnerAddress.Bytes
		}

		if amnft := vnft.R.VehicleTokenAftermarketDevice; amnft != nil {
			out.AftermarketDevice = &pb.AftermarketDevice{
				Serial:              amnft.Serial,
				UserId:              amnft.UserID.Ptr(),
				TokenId:             decimalToUint(amnft.TokenID),
				ManufacturerTokenId: decimalToUint(amnft.DeviceManufacturerTokenID),
			}

			if amnft.OwnerAddress.Valid {
				out.AftermarketDevice.OwnerAddress = amnft.OwnerAddress.Bytes

				if amnft.Beneficiary.Valid {
					out.AftermarketDevice.Beneficiary = amnft.Beneficiary.Bytes
				} else {
					out.AftermarketDevice.Beneficiary = amnft.OwnerAddress.Bytes
				}
			}

			// These fields have been deprecated but are populated for backwards compatibility.
			out.AftermarketDeviceBeneficiaryAddress = out.AftermarketDevice.Beneficiary //nolint:staticcheck
			out.AftermarketDeviceTokenId = &out.AftermarketDevice.TokenId               //nolint:staticcheck
		}

		if vc := vnft.R.Claim; vc != nil {
			out.LatestVinCredential = &pb.VinCredential{
				Id:         vc.ClaimID,
				Expiration: timestamppb.New(vc.ExpirationDate),
			}
		}

		if sd := vnft.R.VehicleTokenSyntheticDevice; sd != nil && !sd.TokenID.IsZero() {
			stk, _ := sd.TokenID.Uint64()
			iTkID, _ := sd.IntegrationTokenID.Uint64()
			out.SyntheticDevice = &pb.SyntheticDevice{
				TokenId:            stk,
				IntegrationTokenId: iTkID,
			}
		}
	}

	for i, udai := range ud.R.UserDeviceAPIIntegrations {
		out.Integrations[i] = &pb.UserDeviceIntegration{Id: udai.IntegrationID, Status: udai.Status}
		if udai.ExternalID.Valid {
			out.Integrations[i].ExternalId = udai.ExternalID.String
		}
	}

	if ud.VinConfirmed {
		out.Vin = &ud.VinIdentifier.String
	}

	if ud.CountryCode.Valid {
		out.CountryCode = ud.CountryCode.String
	}

	// metadata properties
	md := services.UserDeviceMetadata{}
	if err := ud.Metadata.Unmarshal(&md); err != nil {
		s.logger.Error().Msgf("Could not unmarshal userdevice metadata for device: %s", ud.ID)
	}

	if md.PowertrainType != nil {
		out.PowerTrainType = md.PowertrainType.String()
	}
	if md.CANProtocol != nil {
		out.CANProtocol = *md.CANProtocol
	}
	if md.PostalCode != nil {
		out.PostalCode = *md.PostalCode
	}
	if md.GeoDecodedStateProv != nil {
		out.GeoDecodedStateProv = *md.GeoDecodedStateProv
	}
	if md.GeoDecodedCountry != nil {
		out.GeoDecodedCountry = *md.GeoDecodedCountry
	}

	return out
}

func (s *userDeviceRPCServer) GetClaimedVehiclesGrowth(ctx context.Context, _ *emptypb.Empty) (*pb.ClaimedVehiclesGrowth, error) {
	// Checking both that the nft exists and is linked to a device.

	query := `select count(1)
			  from devices_api.vehicle_nfts n 
			  inner join devices_api.meta_transaction_requests m
			  on n.mint_request_id = m.id
			  where n.user_device_id is not null and n.token_id is not null 
			  and m.created_at > current_date - 7;`

	var lastWeeksNFT struct {
		Count int `boil:"count"`
	}

	err := queries.Raw(query).Bind(ctx, s.dbs().Reader, &lastWeeksNFT)

	if err != nil {
		return nil, err
	}

	totalNFT, err := models.VehicleNFTS(models.VehicleNFTWhere.UserDeviceID.IsNotNull(),
		models.VehicleNFTWhere.TokenID.IsNotNull()).Count(ctx, s.dbs().Reader)

	growthPercentage := float32(0)

	if totalNFT > 0 {
		growthPercentage = (float32(lastWeeksNFT.Count) / float32(totalNFT)) * 100
	}

	if err != nil {
		return nil, err
	}

	return &pb.ClaimedVehiclesGrowth{
		TotalClaimedVehicles: totalNFT,
		GrowthPercentage:     growthPercentage,
	}, nil
}

// toUint64 takes a nullable decimal and returns nil if there is no value, or
// a reference to the uint64 value of the decimal otherwise. If the value does not
// fit then we return nil and log.
func (s *userDeviceRPCServer) toUint64(dec types.NullDecimal) *uint64 {
	if dec.IsZero() {
		return nil
	}

	ui, ok := dec.Uint64()
	if !ok {
		s.logger.Error().Str("decimal", dec.String()).Msg("Value too large for uint64.")
		return nil
	}

	return &ui
}

func nullTimeToPB(t null.Time) *timestamppb.Timestamp {
	if !t.Valid {
		return nil
	}

	return timestamppb.New(t.Time)
}

func (s *userDeviceRPCServer) IssueVinCredential(ctx context.Context, req *pb.IssueVinCredentialRequest) (*pb.IssueVinCredentialResponse, error) {
	v, err := models.VehicleNFTS(models.VehicleNFTWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(int64(req.TokenId), 0)))).One(ctx, s.dbs().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "No vehicle with that id.")
		}
		return nil, err
	}

	if req.Vin != v.Vin {
		return nil, status.Error(codes.InvalidArgument, "Input and NFT VINs do not much.")
	}

	credID, err := s.vcIss.VIN(req.Vin, new(big.Int).SetUint64(req.TokenId), req.ExpiresAt.AsTime())
	if err != nil {
		s.logger.Err(err).Msg("Failed to create vin credential.")
		return nil, err
	}
	return &pb.IssueVinCredentialResponse{
		CredentialId: credID,
	}, nil
}

func (s *userDeviceRPCServer) UpdateUserDeviceMetadata(ctx context.Context, req *pb.UpdateUserDeviceMetadataRequest) (*emptypb.Empty, error) {
	ud, err := models.FindUserDevice(ctx, s.dbs().Reader, req.UserDeviceId)
	if err != nil {
		return nil, err
	}
	var udMd services.UserDeviceMetadata
	_ = ud.Metadata.Unmarshal(&udMd)

	if req.GeoDecodedStateProv != nil {
		udMd.GeoDecodedStateProv = req.GeoDecodedStateProv
	}
	if req.PostalCode != nil {
		udMd.PostalCode = req.PostalCode
	}
	if req.GeoDecodedCountry != nil {
		udMd.GeoDecodedCountry = req.GeoDecodedCountry
	}
	_ = ud.Metadata.Marshal(udMd)
	_, err = ud.Update(ctx, s.dbs().Writer, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *userDeviceRPCServer) ClearMetaTransactionRequests(ctx context.Context, _ *emptypb.Empty) (*pb.ClearMetaTransactionRequestsResponse, error) {
	currTime := time.Now()
	fifteenminsAgo := currTime.Add(-time.Minute * 15)

	m, err := models.MetaTransactionRequests(
		models.MetaTransactionRequestWhere.CreatedAt.LT(fifteenminsAgo),
		qm.OrderBy(models.MetaTransactionRequestColumns.CreatedAt+" ASC"),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no overdue meta transaction")
		}
		return nil, fmt.Errorf("failed to select transaction to clear: %w", err)
	}

	_, err = m.Delete(ctx, s.dbs().Writer)
	if err != nil {
		return nil, fmt.Errorf("failed to Delete transaction %s: %w", m.ID, err)
	}

	return &pb.ClearMetaTransactionRequestsResponse{Id: m.ID}, nil
}

func (s *userDeviceRPCServer) DeleteSyntheticDeviceIntegration(ctx context.Context, req *pb.DeleteSyntheticDeviceIntegrationsRequest) (*pb.DeleteSyntheticDeviceIntegrationResponse, error) {
	resp := &pb.DeleteSyntheticDeviceIntegrationResponse{}

	for _, deleteRequest := range req.DeviceIntegrations {

		s.logger.Info().
			Str("userDeviceId", deleteRequest.UserDeviceId).
			Str("integrationId", deleteRequest.IntegrationId).
			Msg("deleting integration on behalf of user")

		apiInt, err := models.UserDeviceAPIIntegrations(
			models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(deleteRequest.UserDeviceId),
			models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(deleteRequest.IntegrationId),
			qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
		).One(ctx, s.dbs().Reader)
		if err != nil {
			return nil, err
		}

		if apiInt.R.UserDevice == nil {
			return nil, fmt.Errorf("failed to find user device %s for integration %s", deleteRequest.UserDeviceId, deleteRequest.IntegrationId)
		}

		dd, err := s.deviceDefSvc.GetDeviceDefinitionByID(ctx, apiInt.R.UserDevice.DeviceDefinitionID)
		if err != nil {
			return nil, fmt.Errorf("deviceDefSvc error getting device definition by id %s: %w", apiInt.R.UserDevice.DeviceDefinitionID, err)
		}

		integ, err := s.deviceDefSvc.GetIntegrationByID(ctx, deleteRequest.IntegrationId)
		if err != nil {
			return nil, fmt.Errorf("deviceDefSvc error getting integration by id %s: %w", deleteRequest.IntegrationId, err)
		}

		switch integ.Vendor {
		case constants.SmartCarVendor:
			if apiInt.TaskID.Valid {
				err = s.smartcarTaskSvc.StopPoll(apiInt)
				if err != nil {
					return nil, fmt.Errorf("failed to stop smartcar poll: %w", err)
				}
			}
		case constants.TeslaVendor:
			if apiInt.TaskID.Valid {
				err = s.teslaTaskService.StopPoll(apiInt)
				if err != nil {
					return nil, fmt.Errorf("failed to stop tesla poll: %w", err)
				}
			}
		}

		_, err = apiInt.Delete(ctx, s.dbs().Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to delete integration: %w", err)
		}

		if err := s.eventService.Emit(&shared.CloudEvent[any]{
			Type:    "com.dimo.zone.device.integration.delete",
			Source:  "devices-api",
			Subject: apiInt.UserDeviceID,
			Data: services.UserDeviceIntegrationEvent{
				Timestamp: time.Now(),
				UserID:    apiInt.R.UserDevice.UserID,
				Device: services.UserDeviceEventDevice{
					ID:    apiInt.UserDeviceID,
					Make:  dd.Make.Name,
					Model: dd.Type.Model,
					Year:  int(dd.Type.Year),
				},
				Integration: services.UserDeviceEventIntegration{
					ID:     integ.Id,
					Type:   integ.Type,
					Style:  integ.Style,
					Vendor: integ.Vendor,
				},
			},
		}); err != nil {
			s.logger.Err(err).Msg("Failed to emit integration deletion")
		}
		resp.ImpactedUserDeviceIds = append(resp.ImpactedUserDeviceIds, deleteRequest.UserDeviceId)
	}

	return resp, nil
}
