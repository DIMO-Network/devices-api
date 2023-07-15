package services

import (
	"context"
	dagrpc "github.com/DIMO-Network/device-data-api/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type deviceDataService struct {
	logger *zerolog.Logger
	client dagrpc.UserDeviceDataServiceClient
}

type DeviceDataService interface {
	GetDeviceData(ctx context.Context, userDeviceID, deviceDefinitionID, deviceStyleID string, privilegeIDs []int64) (*dagrpc.UserDeviceDataResponse, error)
}

func NewDeviceDataService(deviceDataGrpcAddr string, logger *zerolog.Logger) DeviceDataService {
	uddcon, err := grpc.Dial(deviceDataGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed dialing device-data-api.")
	}
	client := dagrpc.NewUserDeviceDataServiceClient(uddcon)

	return &deviceDataService{client: client, logger: logger}
}

func (ddd *deviceDataService) GetDeviceData(ctx context.Context, userDeviceID, deviceDefinitionID, deviceStyleID string, privilegeIDs []int64) (*dagrpc.UserDeviceDataResponse, error) {
	data, err := ddd.client.GetUserDeviceData(ctx, &dagrpc.UserDeviceDataRequest{
		UserDeviceId:       userDeviceID,
		DeviceDefinitionId: deviceDefinitionID,
		DeviceStyleId:      deviceStyleID,
		PrivilegeIds:       privilegeIDs,
	})
	return data, err
}

// todo add tests for nft controller and the status controller
