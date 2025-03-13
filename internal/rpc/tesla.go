package rpc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func NewTeslaRPCService(
	dbs func() *db.ReaderWriter,
	settings *config.Settings,
	cipher shared.Cipher,
	teslaAPI services.TeslaFleetAPIService,
	logger *zerolog.Logger,
) pb.TeslaServiceServer {
	return &teslaRPCServer{
		dbs:      dbs,
		logger:   logger,
		settings: settings,
		cipher:   cipher,
		teslaAPI: teslaAPI,
	}
}

// userDeviceRPCServer is the grpc server implementation for the proto services
type teslaRPCServer struct {
	pb.UnimplementedTeslaServiceServer
	dbs      func() *db.ReaderWriter
	logger   *zerolog.Logger
	settings *config.Settings
	cipher   shared.Cipher
	teslaAPI services.TeslaFleetAPIService
}

func convertBoolRef(b *bool) *wrapperspb.BoolValue {
	if b == nil {
		return nil
	}
	return wrapperspb.Bool(*b)
}

func (s *teslaRPCServer) GetPollingInfo(ctx context.Context, req *pb.GetPollingInfoRequest) (*pb.GetPollingInfoResponse, error) {
	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.TaskID.EQ(null.StringFrom(req.TaskId)),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No Tesla task with that id found.")
		}
		return nil, err
	}

	var meta services.UserDeviceAPIIntegrationsMetadata
	if err := udai.Metadata.Unmarshal(&meta); err != nil {
		return nil, err
	}

	return &pb.GetPollingInfoResponse{
		DiscountedData:        convertBoolRef(meta.TeslaDiscountedData),
		FleetTelemetryCapable: convertBoolRef(meta.TeslaFleetTelemetryCapable),
	}, nil
}

func (s *teslaRPCServer) GetFleetStatus(ctx context.Context, req *pb.GetFleetStatusRequest) (*pb.GetFleetStatusResponse, error) {
	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(req.VehicleTokenId, 0))),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg")),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No Vehicle with that token id found.")
		}
		return nil, err
	}

	if len(ud.R.UserDeviceAPIIntegrations) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "No Tesla integration found.")
	}

	udai := ud.R.UserDeviceAPIIntegrations[0]

	var metadata services.UserDeviceAPIIntegrationsMetadata
	err = udai.Metadata.Unmarshal(&metadata)
	if err != nil {
		return nil, err
	}

	if metadata.TeslaVIN == "" {
		return nil, status.Error(codes.FailedPrecondition, "No VIN attached to integration.")
	}
	vin := metadata.TeslaVIN

	token, err := s.cipher.Decrypt(ud.R.UserDeviceAPIIntegrations[0].AccessToken.String)
	if err != nil {
		return nil, err
	}

	res, err := s.teslaAPI.VirtualKeyConnectionStatus(ctx, token, vin)
	if err != nil {
		return nil, err
	}

	return &pb.GetFleetStatusResponse{
		KeyPaired:                      res.KeyPaired,
		FirmwareVersion:                res.FirmwareVersion,
		VehicleCommandProtocolRequired: res.VehicleCommandProtocolRequired,
		FleetTelemetryVersion:          res.FleetTelemetryVersion,
		DiscountedDeviceData:           res.DiscountedDeviceData,
	}, nil
}

func (s *teslaRPCServer) GetFleetTelemetryConfig(ctx context.Context, req *pb.GetFleetTelemetryConfigRequest) (*pb.GetFleetTelemetryConfigResponse, error) {
	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(req.VehicleTokenId, 0))),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg")),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No Vehicle with that token id found.")
		}
		return nil, err
	}

	if len(ud.R.UserDeviceAPIIntegrations) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "No Tesla integration found.")
	}

	udai := ud.R.UserDeviceAPIIntegrations[0]

	var metadata services.UserDeviceAPIIntegrationsMetadata
	err = udai.Metadata.Unmarshal(&metadata)
	if err != nil {
		return nil, err
	}

	if metadata.TeslaVIN == "" {
		return nil, status.Error(codes.FailedPrecondition, "No VIN attached to integration.")
	}
	vin := metadata.TeslaVIN

	token, err := s.cipher.Decrypt(ud.R.UserDeviceAPIIntegrations[0].AccessToken.String)
	if err != nil {
		return nil, err
	}

	res, err := s.teslaAPI.GetTelemetrySubscriptionStatus(ctx, token, vin)
	if err != nil {
		return nil, err
	}

	return &pb.GetFleetTelemetryConfigResponse{
		Configured:   res.Configured,
		Synced:       res.Synced,
		KeyPaired:    res.KeyPaired,
		LimitReached: res.LimitReached,
	}, nil
}

func (s *teslaRPCServer) ConfigureFleetTelemetry(ctx context.Context, req *pb.ConfigureFleetTelemetryRequest) (*pb.ConfigureFleetTelemetryResponse, error) {
	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(req.VehicleTokenId, 0))),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg")),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No Vehicle with that token id found.")
		}
		return nil, err
	}

	if len(ud.R.UserDeviceAPIIntegrations) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "No Tesla integration found.")
	}

	udai := ud.R.UserDeviceAPIIntegrations[0]

	var metadata services.UserDeviceAPIIntegrationsMetadata
	err = udai.Metadata.Unmarshal(&metadata)
	if err != nil {
		return nil, err
	}

	if metadata.TeslaVIN == "" {
		return nil, status.Error(codes.FailedPrecondition, "No VIN attached to integration.")
	}
	vin := metadata.TeslaVIN

	token, err := s.cipher.Decrypt(ud.R.UserDeviceAPIIntegrations[0].AccessToken.String)
	if err != nil {
		return nil, err
	}

	err = s.teslaAPI.SubscribeForTelemetryData(ctx, token, vin)
	if err != nil {
		return nil, err
	}

	return &pb.ConfigureFleetTelemetryResponse{}, nil
}

func (s *teslaRPCServer) GetScopes(ctx context.Context, req *pb.GetScopesRequest) (*pb.GetScopesResponse, error) {
	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(req.VehicleTokenId, 0))),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg")),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No Vehicle with that token id found.")
		}
		return nil, err
	}

	if len(ud.R.UserDeviceAPIIntegrations) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "No Tesla integration found.")
	}

	token, err := s.cipher.Decrypt(ud.R.UserDeviceAPIIntegrations[0].AccessToken.String)
	if err != nil {
		return nil, err
	}

	var claims partialTeslaClaims
	_, _, err = jwt.NewParser().ParseUnverified(token, &claims)
	if err != nil {
		return nil, err
	}

	return &pb.GetScopesResponse{Scopes: claims.Scopes}, nil
}

type partialTeslaClaims struct {
	jwt.RegisteredClaims
	Scopes []string `json:"scp"`
}
