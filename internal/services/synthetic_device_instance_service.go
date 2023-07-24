package services

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	pb "github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SyntheticWalletInstanceService interface {
	SignHash(ctx context.Context, childNumber uint32, hash []byte) ([]byte, error)
	GetAddress(ctx context.Context, childNumber uint32) ([]byte, error)
}

type syntheticWalletInstanceService struct {
	client pb.SyntheticWalletClient
}

func NewSyntheticWalletInstanceService(settings *config.Settings) (SyntheticWalletInstanceService, error) {
	conn, err := grpc.Dial(settings.SyntheticWalletGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &syntheticWalletInstanceService{client: pb.NewSyntheticWalletClient(conn)}, nil
}

func (v *syntheticWalletInstanceService) GetAddress(ctx context.Context, childNumber uint32) ([]byte, error) {
	res, err := v.client.GetAddress(ctx, &pb.GetAddressRequest{
		ChildNumber: childNumber,
	})
	if err != nil {
		return nil, err
	}

	return res.Address, nil
}

func (v *syntheticWalletInstanceService) SignHash(ctx context.Context, childNumber uint32, hash []byte) ([]byte, error) {
	res, err := v.client.SignHash(ctx, &pb.SignHashRequest{
		ChildNumber: childNumber,
		Hash:        hash,
	})
	if err != nil {
		return nil, err
	}

	return res.Signature, nil
}
