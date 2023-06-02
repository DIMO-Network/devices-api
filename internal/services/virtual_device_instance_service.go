package services

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared/db"
	pbvirt "github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VirtualDeviceInstanceService interface {
	GetAddress(ctx context.Context, req *pbvirt.GetAddressRequest) (*pbvirt.GetAddressResponse, error)
	SignHash(ctx context.Context, childNumber uint32, hash []byte) ([]byte, error)
}

type virtualDeviceInstanceService struct {
	dbs                           func() *db.ReaderWriter
	virtualDeviceInstanceGRPCAddr string
}

func NewVirtualDeviceInstanceService(DBS func() *db.ReaderWriter, settings *config.Settings) VirtualDeviceInstanceService {
	return &virtualDeviceInstanceService{
		dbs:                           DBS,
		virtualDeviceInstanceGRPCAddr: settings.VirtualDeviceInstanceGrpcAddr,
	}
}

func (v *virtualDeviceInstanceService) GetAddress(ctx context.Context, req *pbvirt.GetAddressRequest) (*pbvirt.GetAddressResponse, error) {
	_, conn, err := v.getVirtualDeviceGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return nil, nil
}

func (v *virtualDeviceInstanceService) SignHash(ctx context.Context, childNumber uint32, hash []byte) ([]byte, error) {
	client, conn, err := v.getVirtualDeviceGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := client.SignHash(ctx, &pbvirt.SignHashRequest{
		ChildNumber: childNumber,
		Hash:        hash,
	})

	if err != nil {
		return nil, err
	}

	return res.Signature, nil
}

// getDeviceDefsIntGrpcClient instanties new connection with client to dd service. You must defer conn.close from returned connection
func (v *virtualDeviceInstanceService) getVirtualDeviceGrpcClient() (pbvirt.SyntheticWalletClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(v.virtualDeviceInstanceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	virtualDeviceClient := pbvirt.NewSyntheticWalletClient(conn)
	return virtualDeviceClient, conn, nil
}
