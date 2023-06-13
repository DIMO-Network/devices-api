package services

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared/db"
	pb "github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SyntheticWalletInstanceService interface {
	SignHash(ctx context.Context, childNumber uint32, hash []byte) ([]byte, error)
	GetAddress(ctx context.Context, childNumber uint32) ([]byte, error)
}

type syntheticWalletInstanceService struct {
	dbs  func() *db.ReaderWriter
	grpc *grpcClient
}

type grpcClient struct {
	client pb.SyntheticWalletClient
	conn   *grpc.ClientConn
}

func NewSyntheticWalletInstanceService(DBS func() *db.ReaderWriter, settings *config.Settings) (SyntheticWalletInstanceService, error) {
	conn, err := grpc.Dial(settings.SyntheticWalletGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	virtualDeviceClient := pb.NewSyntheticWalletClient(conn)

	grpc := &grpcClient{
		conn:   conn,
		client: virtualDeviceClient,
	}

	return &syntheticWalletInstanceService{
		dbs:  DBS,
		grpc: grpc,
	}, nil
}

func (v *syntheticWalletInstanceService) GetAddress(ctx context.Context, childNumber uint32) ([]byte, error) {
	res, err := v.grpc.client.GetAddress(ctx, &pb.GetAddressRequest{
		ChildNumber: childNumber,
	})

	if err != nil {
		return nil, err
	}

	return res.Address, nil
}

func (v *syntheticWalletInstanceService) SignHash(ctx context.Context, childNumber uint32, hash []byte) ([]byte, error) {
	res, err := v.grpc.client.SignHash(ctx, &pb.SignHashRequest{
		ChildNumber: childNumber,
		Hash:        hash,
	})

	if err != nil {
		return nil, err
	}

	return res.Signature, nil
}
