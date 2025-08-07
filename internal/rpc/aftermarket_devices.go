package rpc

import (
	"context"

	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewAftermarketDeviceService(dbs func() *db.ReaderWriter, logger *zerolog.Logger) pb.AftermarketDeviceServiceServer {
	return &aftermarketDeviceService{dbs: dbs, logger: logger}
}

type aftermarketDeviceService struct {
	pb.UnimplementedAftermarketDeviceServiceServer
	dbs    func() *db.ReaderWriter
	logger *zerolog.Logger
}

func (s *aftermarketDeviceService) ListAftermarketDevicesForUser(ctx context.Context, req *pb.ListAftermarketDevicesForUserRequest) (*pb.ListAftermarketDevicesForUserResponse, error) {
	units, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.UserID.EQ(null.StringFrom(req.UserId)),
	).All(ctx, s.dbs().Reader)
	if err != nil {
		s.logger.Err(err).Str("userId", req.UserId).Str("method", "ListAftermarketDevicesForUser").Msg("Database failure.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	out := make([]*pb.AftermarketDevice, len(units))

	for i, unit := range units {
		out[i] = &pb.AftermarketDevice{
			Serial:              unit.Serial,
			UserId:              &req.UserId,
			TokenId:             decimalToUint(unit.TokenID),
			ManufacturerTokenId: decimalToUint(unit.DeviceManufacturerTokenID),
		}

		if unit.OwnerAddress.Valid {
			out[i].OwnerAddress = unit.OwnerAddress.Bytes

			if unit.Beneficiary.Valid {
				out[i].Beneficiary = unit.Beneficiary.Bytes
			} else {
				out[i].Beneficiary = unit.OwnerAddress.Bytes
			}
		}
	}

	return &pb.ListAftermarketDevicesForUserResponse{AftermarketDevices: out}, nil
}
