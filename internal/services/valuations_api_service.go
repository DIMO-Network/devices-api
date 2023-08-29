package services

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	vrpc "github.com/DIMO-Network/valuations-api/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -source valuations_api_service.go -destination mocks/valuations_api_service_mock.go -package mock_services

type ValuationsAPIService interface {
	GetUserDeviceValuations(ctx context.Context, userDeviceID string) (*vrpc.DeviceValuation, error)
	GetUserDeviceOffers(ctx context.Context, userDeviceID string) (*vrpc.DeviceOffer, error)
}

type valuationsAPIService struct {
	settings *config.Settings
	log      *zerolog.Logger
}

func NewValuationsAPIService(settings *config.Settings, log *zerolog.Logger) ValuationsAPIService {
	return &valuationsAPIService{
		settings: settings,
		log:      log,
	}
}

func (va *valuationsAPIService) GetUserDeviceValuations(ctx context.Context, userDeviceID string) (*vrpc.DeviceValuation, error) {

	valuationsClient, conn, err := va.connectToValuationsGRPCClient()

	if err != nil {
		va.log.Error().Err(err).Msg("failed to connect to valuations service")
		return nil, err
	}

	defer conn.Close()

	valuations, err := valuationsClient.GetUserDeviceValuation(ctx, &vrpc.DeviceValuationRequest{
		UserDeviceId: userDeviceID,
	})

	if err != nil {
		va.log.Error().Err(err).Msg("failed to get valuations")
		return nil, err
	}

	return valuations, nil
}

func (va *valuationsAPIService) GetUserDeviceOffers(ctx context.Context, userDeviceID string) (*vrpc.DeviceOffer, error) {
	valuationsClient, conn, err := va.connectToValuationsGRPCClient()

	if err != nil {
		va.log.Error().Err(err).Msg("failed to connect to valuations service")
		return nil, err
	}

	defer conn.Close()

	offers, err := valuationsClient.GetUserDeviceOffer(ctx, &vrpc.DeviceOfferRequest{
		UserDeviceId: userDeviceID,
	})

	if err != nil {
		va.log.Error().Err(err).Msg("failed to get offers")
		return nil, err
	}

	return offers, nil
}

func (va *valuationsAPIService) connectToValuationsGRPCClient() (vrpc.ValuationsServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(va.settings.ValuationsAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := vrpc.NewValuationsServiceClient(conn)

	return client, conn, nil
}
