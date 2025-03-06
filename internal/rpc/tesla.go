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

	var out *wrapperspb.BoolValue

	if tdd := meta.TeslaDiscountedData; tdd != nil {
		out = wrapperspb.Bool(*tdd)
	}

	return &pb.GetPollingInfoResponse{DiscountedData: out}, nil
}

//   rpc GetFleetStatus(GetFleetStatusRequest) returns (GetFleetStatusResponse);

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
