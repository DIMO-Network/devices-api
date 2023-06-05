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
}

type syntheticWalletInstanceService struct {
	dbs         func() *db.ReaderWriter
	serviceAddr string
}

func NewSyntheticWalletInstanceService(DBS func() *db.ReaderWriter, settings *config.Settings) SyntheticWalletInstanceService {
	return &syntheticWalletInstanceService{
		dbs:         DBS,
		serviceAddr: settings.SyntheticWalletGRPCAddr,
	}
}

func (v *syntheticWalletInstanceService) SignHash(ctx context.Context, childNumber uint32, hash []byte) ([]byte, error) {
	client, conn, err := v.getVirtualDeviceGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := client.SignHash(ctx, &pb.SignHashRequest{
		ChildNumber: childNumber,
		Hash:        hash,
	})

	if err != nil {
		return nil, err
	}

	return res.Signature, nil
}

// getDeviceDefsIntGrpcClient instanties new connection with client to dd service. You must defer conn.close from returned connection
func (v *syntheticWalletInstanceService) getVirtualDeviceGrpcClient() (pb.SyntheticWalletClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(v.serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	virtualDeviceClient := pb.NewSyntheticWalletClient(conn)
	return virtualDeviceClient, conn, nil
}
