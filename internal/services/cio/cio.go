package cio

import (
	"context"

	pb_accounts "github.com/DIMO-Network/accounts-api/pkg/grpc"
	analytics "github.com/customerio/cdp-analytics-go"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

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

func (s *Service) SoftwareDisconnectionEvent(ctx context.Context, vehicleID uint64, address []byte, integrationID string) error {
	account, err := s.acctClient.GetAccount(ctx, &pb_accounts.GetAccountRequest{
		WalletAddress: address,
	})
	if err != nil {
		s.logger.Err(err).Str("wallet_address", common.Bytes2Hex(address)).Msg("failed to get account by wallet address from accounts api")
		return err
	}

	return s.client.Enqueue(
		analytics.Identify{
			UserId: account.GetId(),
			Traits: analytics.NewTraits().Set("integration_id", integrationID).Set("vehicle_id", vehicleID),
		},
	)
}
