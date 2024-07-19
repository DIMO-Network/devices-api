package main

import (
	"context"
	"flag"
	"strings"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type syncUserDeviceDeviceDefinitionCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*syncUserDeviceDeviceDefinitionCmd) Name() string { return "sync-tableland" }
func (*syncUserDeviceDeviceDefinitionCmd) Synopsis() string {
	return "sync User Device device definition to tableland id."
}
func (*syncUserDeviceDeviceDefinitionCmd) Usage() string {
	return `sync-tableland:
	sync User Device device definition to tableland id.
  `
}

func (s *syncUserDeviceDeviceDefinitionCmd) SetFlags(_ *flag.FlagSet) {

}

func (s *syncUserDeviceDeviceDefinitionCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	err := s.syncUserDevicesWithTableland(ctx)

	if err != nil {
		s.logger.Fatal().Err(err).Msg("failed to sync user devices with tableland")
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (s *syncUserDeviceDeviceDefinitionCmd) syncUserDevicesWithTableland(ctx context.Context) error {

	conn, err := grpc.Dial(s.settings.DefinitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Fatal().Err(err).Msg("failed to connect to device definitions grpc")
		return err
	}
	defer conn.Close()

	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	resp, err := definitionsClient.GetDeviceDefinitions(ctx, &emptypb.Empty{})
	if err != nil {
		s.logger.Fatal().Err(err).Msg("failed to get device definitions")
		return err
	}

	dbs := s.pdb.DBS()

	err = ProcessDeviceDefinitions(ctx, dbs, resp.DeviceDefinitions)
	if err != nil {
		return err
	}

	return nil
}

func ProcessDeviceDefinitions(ctx context.Context, dbs *db.ReaderWriter, deviceDefinitions []*ddgrpc.GetDeviceDefinitionItemResponse) error {
	cursor := ""
	hasMore := true

	for hasMore {
		userDevices, err := models.UserDevices(
			models.UserDeviceWhere.DefinitionID.IsNull(),
			models.UserDeviceWhere.ID.GT(cursor),
			qm.Limit(5000),
			qm.OrderBy(models.UserDeviceColumns.ID),
		).All(ctx, dbs.Reader)

		if err != nil {
			return err
		}

		if len(userDevices) == 0 {
			break
		}

		cursor = userDevices[len(userDevices)-1].ID

		for _, dd := range deviceDefinitions {
			for _, ud := range userDevices {
				if strings.EqualFold(dd.DeviceDefinitionId, ud.DeviceDefinitionID) {
					ud.DefinitionID = null.StringFrom(dd.NameSlug)

					_, err := ud.Update(ctx, dbs.Writer, boil.Infer())
					if err != nil {
						return err
					}
				}
			}
		}

		if len(userDevices) < 5000 {
			hasMore = false
		}
	}

	return nil
}
