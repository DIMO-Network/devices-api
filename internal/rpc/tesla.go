package rpc

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
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
