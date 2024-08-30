package services

import (
	"context"

	"github.com/DIMO-Network/shared/privileges"

	dagrpc "github.com/DIMO-Network/device-data-api/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -source device_data_service.go -destination mocks/device_data_service_mock.go
type deviceDataService struct {
	logger *zerolog.Logger
	client dagrpc.UserDeviceDataServiceClient
}

type DeviceDataService interface {
	GetDeviceData(ctx context.Context, userDeviceID, deviceDefinitionID, deviceStyleID string, privilegeIDs []privileges.Privilege) (*dagrpc.UserDeviceDataResponse, error)
	GetRawDeviceData(ctx context.Context, userDeviceID, integrationID string) (*dagrpc.RawDeviceDataResponse, error)
}

func NewDeviceDataService(deviceDataGrpcAddr string, logger *zerolog.Logger) DeviceDataService {
	uddcon, err := grpc.NewClient(deviceDataGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed dialing device-data-api.")
	}
	client := dagrpc.NewUserDeviceDataServiceClient(uddcon)

	return &deviceDataService{client: client, logger: logger}
}

func (ddd *deviceDataService) GetDeviceData(ctx context.Context, userDeviceID, deviceDefinitionID, deviceStyleID string, privilegeIDs []privileges.Privilege) (*dagrpc.UserDeviceDataResponse, error) {
	int64Slice := make([]int64, len(privilegeIDs))
	for i, v := range privilegeIDs {
		int64Slice[i] = int64(v)
	}
	data, err := ddd.client.GetUserDeviceData(ctx, &dagrpc.UserDeviceDataRequest{
		UserDeviceId:       userDeviceID,
		DeviceDefinitionId: deviceDefinitionID,
		DeviceStyleId:      deviceStyleID,
		PrivilegeIds:       int64Slice,
	})
	return data, err
}

func (ddd *deviceDataService) GetRawDeviceData(ctx context.Context, userDeviceID, integrationID string) (*dagrpc.RawDeviceDataResponse, error) {
	data, err := ddd.client.GetRawDeviceData(ctx, &dagrpc.RawDeviceDataRequest{
		UserDeviceId:  userDeviceID,
		IntegrationId: &integrationID,
	})
	return data, err
}
