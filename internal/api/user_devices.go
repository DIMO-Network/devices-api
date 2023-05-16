package api

import (
	"context"
	"database/sql"
	"errors"

	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewUserDeviceService(dbs func() *db.ReaderWriter, settings *config.Settings, hardwareTemplateService autopi.HardwareTemplateService, logger *zerolog.Logger, deviceDefSvc services.DeviceDefinitionService, eventService services.EventService) pb.UserDeviceServiceServer {
	return &userDeviceService{dbs: dbs,
		logger:                  logger,
		settings:                settings,
		hardwareTemplateService: hardwareTemplateService,
		deviceDefSvc:            deviceDefSvc,
		eventService:            eventService,
	}
}

type userDeviceService struct {
	pb.UnimplementedUserDeviceServiceServer
	dbs                     func() *db.ReaderWriter
	hardwareTemplateService autopi.HardwareTemplateService
	logger                  *zerolog.Logger
	settings                *config.Settings
	deviceDefSvc            services.DeviceDefinitionService
	eventService            services.EventService
}

func (s *userDeviceService) GetUserDevice(ctx context.Context, req *pb.GetUserDeviceRequest) (*pb.UserDevice, error) {
	dbDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(req.Id),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenAutopiUnit)),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
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

func (s *userDeviceService) GetUserDeviceByTokenId(ctx context.Context, req *pb.GetUserDeviceByTokenIdRequest) (*pb.UserDevice, error) { //nolint

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

func (s *userDeviceService) ListUserDevicesForUser(ctx context.Context, req *pb.ListUserDevicesForUserRequest) (*pb.ListUserDevicesForUserResponse, error) {
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
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenAutopiUnit)),
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

func (s *userDeviceService) ApplyHardwareTemplate(ctx context.Context, req *pb.ApplyHardwareTemplateRequest) (*pb.ApplyHardwareTemplateResponse, error) {
	resp, err := s.hardwareTemplateService.ApplyHardwareTemplate(ctx, req)
	if err != nil {
		s.logger.Err(err).Str("autopi_unit_id", req.AutoApiUnitId).Str("user_device_id", req.UserDeviceId).Msgf("failed to apply hardware template id %s", req.HardwareTemplateId)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, err
}

func (s *userDeviceService) CreateTemplate(_ context.Context, req *pb.CreateTemplateRequest) (*pb.CreateTemplateResponse, error) {
	resp, err := s.hardwareTemplateService.CreateTemplate(req)
	if err != nil {
		s.logger.Err(err).Str("template name", req.Name).Msgf("failed to create template %s", req.Name)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, err
}

func (s *userDeviceService) RegisterUserDeviceFromVIN(ctx context.Context, req *pb.RegisterUserDeviceFromVINRequest) (*pb.RegisterUserDeviceFromVINResponse, error) {
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

	if len(resp.DeviceDefinitionId) == 0 {
		s.logger.Warn().
			Str("vin", vin).
			Str("user_id", req.UserDeviceId).
			Msg("unable to decode vin for customer request to create vehicle")
		return nil, status.Error(codes.Internal, err.Error())
	}

	// attach device def to user
	dd, err := s.deviceDefSvc.GetDeviceDefinitionByID(ctx, resp.DeviceDefinitionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tx, err := s.dbs().Writer.DB.BeginTx(ctx, nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return nil, err
	}

	// future refactor: udc controller has a createUserDevice
	userDeviceID := ksuid.New().String()
	// register device for the user
	ud := models.UserDevice{
		ID:                 userDeviceID,
		UserID:             req.UserDeviceId,
		DeviceDefinitionID: dd.DeviceDefinitionId,
		VinIdentifier:      null.StringFrom(vin),
		CountryCode:        null.StringFrom(req.CountryCode),
		VinConfirmed:       true,
		Metadata:           null.JSON{}, // todo set powertrain
	}
	if len(resp.DeviceStyleId) > 0 {
		ud.DeviceStyleID = null.StringFrom(resp.DeviceStyleId)
	}
	err = ud.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("could not create user device for def_id: %s", dd.DeviceDefinitionId).Error())
	}

	err = tx.Commit() // commmit the transaction
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.eventService.Emit(&services.Event{
		Type:    constants.UserDeviceCreationEventType,
		Subject: req.UserDeviceId,
		Source:  "devices-api",
		Data: services.UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    req.UserDeviceId,
			Device: services.UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year), // Odd.
			},
		},
	})

	if err != nil {
		s.logger.Err(err).Msg("Failed emitting device creation event")
	}

	return &pb.RegisterUserDeviceFromVINResponse{Created: true}, err
}

//nolint:all
func (s *userDeviceService) GetUserDeviceByAutoPIUnitId(ctx context.Context, req *pb.GetUserDeviceByAutoPIUnitIdRequest) (*pb.UserDeviceAutoPIUnitResponse, error) {
	dbDevice, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.AutopiUnitID.EQ(null.StringFrom(req.Id)),
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

func (s *userDeviceService) GetAllUserDeviceValuation(ctx context.Context, _ *emptypb.Empty) (*pb.ValuationResponse, error) {

	euroToUsd := 1.10

	query := `select sum(evd.retail_price) as totalRetail,
					 sum(evd.drivly_price) as totalDrivly,
					 from
                             (
								select distinct on (vin) vin, 
														pricing_metadata, 
														jsonb_path_query(evd.pricing_metadata, '$.retail.kelley.book')::decimal as retail_price,
														jsonb_path_query(evd.vincario_metadata, '$.market_price.price_avg')::decimal as drivly_price,
														created_at
       							from external_vin_data evd 
								order by vin, created_at desc
							) as evd;`

	queryGrowth := `select sum(evd.retail_price) as totalRetail,
					 sum(evd.drivly_price) as totalDrivly,
					 from
						(
							select distinct on (vin) vin, 
													pricing_metadata, 
													jsonb_path_query(evd.pricing_metadata, '$.retail.kelley.book')::decimal as retail_price, 
													jsonb_path_query(evd.vincario_metadata, '$.market_price.price_avg')::decimal as drivly_price,
													created_at
							from external_vin_data evd 
							where created_at > current_date - 7
							order by vin, created_at desc
						) as evd;`

	type Result struct {
		TotalRetail null.Float64 `boil:"totalRetail"`
		TotalDrivly null.Float64 `boil:"totalDrivly"`
	}
	var total Result
	var lastWeek Result

	err := queries.Raw(query).Bind(ctx, s.dbs().Reader, &total)
	if err != nil {
		s.logger.Err(err).Msg("Database failure retrieving total valuation.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	err = queries.Raw(queryGrowth).Bind(ctx, s.dbs().Reader, &lastWeek)
	if err != nil {
		s.logger.Err(err).Msg("Database failure retrieving last week valuation.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	totalValuation := 0.0
	growthPercentage := 0.0

	if !total.TotalRetail.IsZero() {
		totalValuation = total.TotalRetail.Float64
	}

	if !total.TotalDrivly.IsZero() {
		totalValuation += total.TotalDrivly.Float64 * euroToUsd
	}

	if totalValuation > 0 {

		totalLastWeek := 0.0

		if !lastWeek.TotalRetail.IsZero() {
			totalLastWeek = lastWeek.TotalRetail.Float64
		}

		if !lastWeek.TotalDrivly.IsZero() {
			totalLastWeek += lastWeek.TotalDrivly.Float64 * euroToUsd
		}

		growthPercentage = (totalLastWeek / totalValuation) * 100
	}

	// todo: get an average valuation per vehicle, and multiply for whatever count of vehicles we did not get value for

	return &pb.ValuationResponse{
		Total:            float32(totalValuation),
		GrowthPercentage: float32(growthPercentage),
	}, nil
}

func (s *userDeviceService) deviceModelToAPI(ud *models.UserDevice) *pb.UserDevice {
	out := &pb.UserDevice{
		Id:                 ud.ID,
		UserId:             ud.UserID,
		DeviceDefinitionId: ud.DeviceDefinitionID,
		DeviceStyleId:      ud.DeviceStyleID.Ptr(),
		OptedInAt:          nullTimeToPB(ud.OptedInAt),
		Integrations:       make([]*pb.UserDeviceIntegration, len(ud.R.UserDeviceAPIIntegrations)),
	}

	if vnft := ud.R.VehicleNFT; vnft != nil {
		out.TokenId = s.toUint64(vnft.TokenID)
		if vnft.OwnerAddress.Valid {
			out.OwnerAddress = vnft.OwnerAddress.Bytes
		}

		if amnft := vnft.R.VehicleTokenAutopiUnit; amnft != nil {
			out.AftermarketDeviceTokenId = s.toUint64(amnft.TokenID)

			if amnft.Beneficiary.Valid {
				out.AftermarketDeviceBeneficiaryAddress = amnft.Beneficiary.Bytes
			} else {
				out.AftermarketDeviceBeneficiaryAddress = vnft.OwnerAddress.Bytes
			}
		}
	}

	for i, udai := range ud.R.UserDeviceAPIIntegrations {
		out.Integrations[i] = &pb.UserDeviceIntegration{Id: udai.IntegrationID, Status: udai.Status}
	}

	if ud.VinConfirmed {
		out.Vin = &ud.VinIdentifier.String
	}

	return out
}

func (s *userDeviceService) GetClaimedVehiclesGrowth(ctx context.Context, _ *emptypb.Empty) (*pb.ClaimedVehiclesGrowth, error) {
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
func (s *userDeviceService) toUint64(dec types.NullDecimal) *uint64 {
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
