package cio

import (
	"context"
	"errors"

	pb_accounts "github.com/DIMO-Network/accounts-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/models"
	analytics "github.com/customerio/cdp-analytics-go"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

const SoftwareConnectionExpiredEvent = "software_connection_expired"

type AccountsClient interface {
	GetAccount(context.Context, *pb_accounts.GetAccountRequest, ...grpc.CallOption) (*pb_accounts.Account, error)
}

type Service struct {
	client     analytics.Client
	logger     *zerolog.Logger
	acctClient AccountsClient
}

func New(cioKey string, acctClient AccountsClient, logger *zerolog.Logger) (*Service, error) {
	client, err := analytics.NewWithConfig(cioKey, analytics.Config{})
	if err != nil {
		return nil, err
	}

	return &Service{
		client:     client,
		acctClient: acctClient,
		logger:     logger,
	}, nil

}

func (s *Service) SoftwareDisconnectionEvent(ctx context.Context, udai *models.UserDeviceAPIIntegration) error {
	if udai.R.UserDevice.TokenID.IsZero() {
		return errors.New("vehicle is not minted")
	}

	if udai.R.UserDevice.OwnerAddress.IsZero() {
		return errors.New("no owner address")
	}

	vehicleTokenID, ok := udai.R.UserDevice.TokenID.Int64()
	if !ok {
		return errors.New("failed to parse vehicle token id")
	}

	sd := udai.R.UserDevice.R.VehicleTokenSyntheticDevice
	if sd == nil {
		return errors.New("no synthetic device associcated with api integration")
	}

	sdWallet := common.BytesToAddress(sd.WalletAddress)

	integTokenID, ok := sd.IntegrationTokenID.Int64()
	if !ok {
		return errors.New("failed to parse integration token id")
	}

	userAddr := common.BytesToAddress(udai.R.UserDevice.OwnerAddress.Bytes)
	account, err := s.acctClient.GetAccount(ctx, &pb_accounts.GetAccountRequest{
		WalletAddress: userAddr.Bytes(),
	})
	if err != nil {
		s.logger.Err(err).Str("user_address", userAddr.Hex()).Msg("failed to get account by wallet address from accounts api")
		return err
	}

	return s.client.Enqueue(
		analytics.Track{
			UserId:     account.GetId(),
			Event:      SoftwareConnectionExpiredEvent,
			Properties: analytics.NewProperties().Set("integration_id", integTokenID).Set("vehicle_id", vehicleTokenID).Set("device_id", sdWallet),
		},
	)
}
