package rpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"

	mtpgrpc "github.com/DIMO-Network/meta-transaction-processor/pkg/grpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/ericlagergren/decimal"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DIMO-Network/devices-api/internal/services/autopi"
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
	userDeviceSvc           services.UserDeviceService
	teslaTaskService        services.TeslaTaskService
	smartcarTaskSvc         services.SmartcarTaskService
}

func (s *userDeviceRPCServer) GetUserDevice(ctx context.Context, req *pb.GetUserDeviceRequest) (*pb.UserDevice, error) {
	dbDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(req.Id),
		qm.Load(models.UserDeviceRels.VehicleTokenAftermarketDevice),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
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
		qm.Load(models.UserDeviceRels.VehicleTokenAftermarketDevice),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
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

	if aftermarketDevice.R.VehicleToken == nil {
		return nil, status.Error(codes.NotFound, "No VehicleToken associated with the AftermarketDevice")
	}

	userDeviceID := aftermarketDevice.R.VehicleToken.ID
	dbDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.VehicleTokenAftermarketDevice),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
	).One(ctx, s.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No device with that ID found.")
		}
		s.logger.Err(err).Str("userDeviceId", userDeviceID).Msg("Database failure retrieving device.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	return s.deviceModelToAPI(dbDevice), nil
}

func (s *userDeviceRPCServer) GetUserDeviceByTokenId(ctx context.Context, req *pb.GetUserDeviceByTokenIdRequest) (*pb.UserDevice, error) { // nolint
	tknID := types.NewNullDecimal(decimal.New(req.TokenId, 0))

	dbDevice, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tknID),
		qm.Load(models.UserDeviceRels.VehicleTokenAftermarketDevice),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No device with that token Id found.")
		}
		s.logger.Err(err).Int64("tokenId", req.TokenId).Msg("Database failure retrieving device.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	return s.deviceModelToAPI(dbDevice), nil
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
			models.UserDeviceWhere.UserID.EQ(req.UserId),
			qm.Or2(models.UserDeviceWhere.OwnerAddress.EQ(null.BytesFrom(common.HexToAddress(req.EthereumAddress).Bytes()))),
		}
	}

	query = append(query,
		qm.Load(models.UserDeviceRels.VehicleTokenAftermarketDevice),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
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
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if resp.Year < 2008 {
		s.logger.Warn().
			Str("vin", vin).
			Str("year", fmt.Sprint(resp.Year)).
			Str("user_id", common.BytesToAddress(req.OwnerAddress).Hex()).
			Msg("VIN is too old")

		return nil, status.Errorf(codes.InvalidArgument, "VIN %s from year %v is too old", vin, resp.Year)
	}

	if len(resp.DefinitionId) == 0 {
		s.logger.Warn().
			Str("vin", vin).
			Str("user_id", common.BytesToAddress(req.OwnerAddress).Hex()).
			Msg("unable to decode vin for customer request to create vehicle")
		return nil, status.Error(codes.Internal, "Unable to decode VIN")
	}

	// attach device def to user
	tx, err := s.dbs().Writer.BeginTx(ctx, nil)
	defer tx.Rollback() // nolint
	if err != nil {
		return nil, err
	}

	_, _, err = s.userDeviceSvc.CreateUserDeviceByOwner(ctx, resp.DefinitionId, resp.DeviceStyleId, req.CountryCode, vin, req.OwnerAddress)
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
		UserDeviceId: dbDevice.UserDeviceID,
		UserId:       dbDevice.R.UserDevice.UserID,
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
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
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
		Id:            ud.ID,
		UserId:        ud.UserID,
		DeviceStyleId: ud.DeviceStyleID.Ptr(),
		OptedInAt:     nullTimeToPB(ud.OptedInAt),
		Integrations:  make([]*pb.UserDeviceIntegration, len(ud.R.UserDeviceAPIIntegrations)),
		VinConfirmed:  ud.VinConfirmed,
		DefinitionId:  ud.DefinitionID,
	}

	if !ud.TokenID.IsZero() {
		out.TokenId = s.toUint64(ud.TokenID)
		if ud.OwnerAddress.Valid {
			out.OwnerAddress = ud.OwnerAddress.Bytes
		}

		if amnft := ud.R.VehicleTokenAftermarketDevice; amnft != nil {
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

		if sd := ud.R.VehicleTokenSyntheticDevice; sd != nil && !sd.TokenID.IsZero() {
			stk, _ := sd.TokenID.Uint64()
			iTkID, _ := sd.IntegrationTokenID.Uint64()
			wc := sd.WalletChildNumber
			out.SyntheticDevice = &pb.SyntheticDevice{
				TokenId:             stk,
				IntegrationTokenId:  iTkID,
				WalletChildNumberId: uint64(wc),
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

	totalNFT, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.IsNotNull()).Count(ctx, s.dbs().Reader)

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

	conn, err := grpc.NewClient(s.settings.MetaTransactionProcessorGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to meta transaction processor: %w", err)
	}

	client := mtpgrpc.NewMetaTransactionServiceClient(conn)
	// call to meta transaction connector to clear the transaction and get id

	response, err := client.CleanStuckMetaTransactions(ctx, &emptypb.Empty{})

	if err != nil {
		return nil, fmt.Errorf("failed to clear meta transaction: %w", err)
	}

	metaTransaction, err := models.MetaTransactionRequests(
		models.MetaTransactionRequestWhere.ID.EQ(response.Id),
		models.MetaTransactionRequestWhere.Status.NEQ("Confirmed"),
		models.MetaTransactionRequestWhere.Hash.IsNull(),
		qm.Load(models.MetaTransactionRequestRels.MintRequestUserDevice),
		qm.Load(models.MetaTransactionRequestRels.BurnRequestUserDevice),
		qm.Load(models.MetaTransactionRequestRels.MintRequestSyntheticDevice),
		qm.Load(models.MetaTransactionRequestRels.BurnRequestSyntheticDevice),
		qm.Load(models.MetaTransactionRequestRels.ClaimMetaTransactionRequestAftermarketDevice),
		qm.Load(models.MetaTransactionRequestRels.PairRequestAftermarketDevice),
		qm.Load(models.MetaTransactionRequestRels.UnpairRequestAftermarketDevice),
	).One(ctx, s.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no overdue meta transaction")
		}
		return nil, fmt.Errorf("failed to select transaction to clear: %w", err)
	}

	if metaTransaction.R != nil {

		if metaTransaction.R.MintRequestUserDevice != nil {
			vehicleNft := metaTransaction.R.MintRequestUserDevice

			vehicleNft.MintRequestID = null.StringFromPtr(nil)
			_, err = vehicleNft.Update(ctx, s.dbs().Writer, boil.Infer())
			if err != nil {
				return nil, fmt.Errorf("failed to update vehicleNft %s: %w", vehicleNft.TokenID, err)
			}
		}

		if metaTransaction.R.BurnRequestUserDevice != nil {
			vehicleNft := metaTransaction.R.BurnRequestUserDevice

			vehicleNft.BurnRequestID = null.StringFromPtr(nil)
			_, err = vehicleNft.Update(ctx, s.dbs().Writer, boil.Infer())
			if err != nil {
				return nil, fmt.Errorf("failed to update vehicleNft %s: %w", vehicleNft.TokenID, err)
			}
		}

		// ? TODO: what should we do with synthetic devices?
		if metaTransaction.R.MintRequestSyntheticDevice != nil {
			s.logger.Warn().Msg(fmt.Sprintf("Could not delete Meta transaction cause is related to synthetic device %s", metaTransaction.R.MintRequestSyntheticDevice.TokenID))
			return &pb.ClearMetaTransactionRequestsResponse{Id: metaTransaction.ID}, nil
		}

		if metaTransaction.R.BurnRequestSyntheticDevice != nil {
			syntheticDevice := metaTransaction.R.BurnRequestSyntheticDevice

			syntheticDevice.BurnRequestID = null.StringFromPtr(nil)
			_, err = syntheticDevice.Update(ctx, s.dbs().Writer, boil.Infer())
			if err != nil {
				return nil, fmt.Errorf("failed to update syntheticDevice %s: %w", syntheticDevice.TokenID, err)
			}
		}

		if metaTransaction.R.ClaimMetaTransactionRequestAftermarketDevice != nil {
			aftermarketDevice := metaTransaction.R.ClaimMetaTransactionRequestAftermarketDevice

			aftermarketDevice.ClaimMetaTransactionRequestID = null.StringFromPtr(nil)
			_, err = aftermarketDevice.Update(ctx, s.dbs().Writer, boil.Infer())
			if err != nil {
				return nil, fmt.Errorf("failed to update aftermarketDevice %s: %w", aftermarketDevice.TokenID, err)
			}
		}

		if metaTransaction.R.PairRequestAftermarketDevice != nil {
			aftermarketDevice := metaTransaction.R.PairRequestAftermarketDevice

			aftermarketDevice.PairRequestID = null.StringFromPtr(nil)
			_, err = aftermarketDevice.Update(ctx, s.dbs().Writer, boil.Infer())
			if err != nil {
				return nil, fmt.Errorf("failed to update aftermarketDevice %s: %w", aftermarketDevice.TokenID, err)
			}
		}

		if metaTransaction.R.UnpairRequestAftermarketDevice != nil {
			aftermarketDevice := metaTransaction.R.UnpairRequestAftermarketDevice

			aftermarketDevice.UnpairRequestID = null.StringFromPtr(nil)
			_, err = aftermarketDevice.Update(ctx, s.dbs().Writer, boil.Infer())
			if err != nil {
				return nil, fmt.Errorf("failed to update aftermarketDevice %s: %w", aftermarketDevice.TokenID, err)
			}
		}
	}

	_, err = metaTransaction.Delete(ctx, s.dbs().Writer)
	if err != nil {
		return nil, fmt.Errorf("failed to Delete transaction %s: %w", metaTransaction.ID, err)
	}

	return &pb.ClearMetaTransactionRequestsResponse{Id: metaTransaction.ID}, nil
}

func (s *userDeviceRPCServer) StopUserDeviceIntegration(ctx context.Context, req *pb.StopUserDeviceIntegrationRequest) (*emptypb.Empty, error) {
	log := s.logger.With().
		Str("userDeviceId", req.UserDeviceId).
		Str("integrationId", req.IntegrationId).Logger()
	log.Info().Msg("stopping user device integration polling")

	apiInt, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(req.UserDeviceId),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(req.IntegrationId),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		log.Err(err).Msg("failed to retrieve integration")
		return nil, fmt.Errorf("failed to retrieve integration %s for user device %s: %w", req.IntegrationId, req.UserDeviceId, err)
	}

	if apiInt.R.UserDevice == nil {
		log.Info().Msg("failed to find user device")
		return nil, fmt.Errorf("failed to find user device %s for integration %s", req.UserDeviceId, req.IntegrationId)
	}

	integ, err := s.deviceDefSvc.GetIntegrationByID(ctx, req.IntegrationId)
	if err != nil {
		log.Err(err).Msg("deviceDefSvc error getting integration by id")
		return nil, fmt.Errorf("deviceDefSvc error getting integration by id: %w", err)
	}

	if !apiInt.TaskID.Valid {
		log.Info().Msg("failed to stop device integration polling; invalid task id")
		return nil, fmt.Errorf("failed to stop device integration polling; invalid task id")
	}

	switch integ.Vendor {
	case constants.SmartCarVendor:
		err = s.smartcarTaskSvc.StopPoll(apiInt)
		if err != nil {
			log.Err(err).Msg("failed to stop smartcar poll")
			return nil, fmt.Errorf("failed to stop smartcar poll: %w", err)
		}
	case constants.TeslaVendor:
		err = s.teslaTaskService.StopPoll(apiInt)
		if err != nil {
			log.Err(err).Msg("failed to stop tesla poll")
			return nil, fmt.Errorf("failed to stop tesla poll: %w", err)
		}
	default:
		log.Info().Str("vendor", integ.Vendor).Msg("stop user integration poll not implemented for vendor")
		return nil, fmt.Errorf("stop user integration poll not implemented for vendor %s", integ.Vendor)
	}

	if apiInt.Status == models.UserDeviceAPIIntegrationStatusAuthenticationFailure {
		log.Info().Msgf("integration authentication status is already %s", models.UserDeviceAPIIntegrationStatusAuthenticationFailure)
		return nil, fmt.Errorf("integration authentication status is already %s", models.UserDeviceAPIIntegrationStatusAuthenticationFailure)
	}

	apiInt.Status = models.UserDeviceAPIIntegrationStatusAuthenticationFailure
	if _, err := apiInt.Update(ctx, s.dbs().Writer, boil.Infer()); err != nil {
		log.Err(err).Msgf("failed to update integration table; task id: %s", apiInt.TaskID.String)
		return nil, fmt.Errorf("failed to update integration table; task id: %s; %w", apiInt.TaskID.String, err)
	}

	log.Info().Msg("integration polling stopped")
	return &emptypb.Empty{}, nil
}

// DeleteVehicle Tries to stops synthetic device tasks using above method, deletes the vehicle from web2
func (s *userDeviceRPCServer) DeleteVehicle(ctx context.Context, req *pb.DeleteVehicleRequest) (*emptypb.Empty, error) {
	ti := new(big.Int).SetUint64(req.TokenId)
	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(tid),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
	).One(ctx, s.dbs().Writer)
	if err != nil {
		return nil, fmt.Errorf("failed to find vehicle %d: %w", req.TokenId, err)
	}

	// check if has synthetic device, and if so stop jobs
	synthDevice, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.VehicleTokenID.EQ(tid)).One(ctx, s.dbs().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error checking synthetic device: %s", err.Error())
	}
	// UDAI and synthetic should be a unit more or less
	if synthDevice != nil {
		// stop the tasks for each
		for _, udai := range userDevice.R.UserDeviceAPIIntegrations {
			_, err := s.StopUserDeviceIntegration(ctx, &pb.StopUserDeviceIntegrationRequest{
				UserDeviceId:  userDevice.ID,
				IntegrationId: udai.IntegrationID,
			})
			// ideally this doesn't happen but we should log regardless
			if err != nil {
				s.logger.Error().Err(err).Uint64("vehicleTokenId", req.TokenId).Msg("failed to stop user device integration from a delete vehicle request")
			}
		}
		// delete the synthetic device
		_, err = models.SyntheticDevices(
			models.SyntheticDeviceWhere.VehicleTokenID.EQ(tid),
		).DeleteAll(ctx, s.dbs().Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to delete synthetic device: %w", err)
		}
	}
	// delete the vehicle, web2 only, we'll still have web3 records.
	_, err = userDevice.Delete(ctx, s.dbs().Writer)
	if err != nil {
		return nil, fmt.Errorf("failed to set delete userDevice %s: %w", userDevice.TokenID, err)
	}
	s.logger.Info().Uint64("vehicleTokenId", req.TokenId).Msgf("successfully deleted vehicle via grpc call")

	return &emptypb.Empty{}, nil
}

// DeleteUnMintedUserDevice deletes user_device records that have not been minted
func (s *userDeviceRPCServer) DeleteUnMintedUserDevice(ctx context.Context, req *pb.DeleteUnMintedUserDeviceRequest) (*emptypb.Empty, error) {
	log := s.logger.With().
		Str("userDeviceId", req.UserDeviceId).
		Logger()

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(req.UserDeviceId),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
	).One(ctx, s.dbs().Writer)
	if err != nil {
		return nil, fmt.Errorf("failed to find vehicle %s: %w", req.UserDeviceId, err)
	}
	if !userDevice.TokenID.IsZero() {
		return nil, fmt.Errorf("cannot delete user device %s, it has a token id", req.UserDeviceId)
	}
	_, err = userDevice.Delete(ctx, s.dbs().Writer)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user device %s : %w", req.UserDeviceId, err)
	}
	log.Info().Msg("deleted unminted user device")
	return &emptypb.Empty{}, nil
}
