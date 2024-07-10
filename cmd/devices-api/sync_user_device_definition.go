package main

import (
	"context"
	"flag"
	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
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

func (s *syncUserDeviceDeviceDefinitionCmd) SetFlags(set *flag.FlagSet) {

}

func (s *syncUserDeviceDeviceDefinitionCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {

	err := s.syncUserDevicesWithTableland(ctx)

	if err != nil {
		s.logger.Fatal().Err(err).Msg("failed to sync user devices with tableland")
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (s *syncUserDeviceDeviceDefinitionCmd) syncUserDevicesWithTableland(ctx context.Context) error {
	dbs := s.pdb.DBS()

	userDevices, err := models.UserDevices().All(ctx, dbs.Reader)

	if err != nil {
		s.logger.Fatal().Err(err).Msg("failed to get user devices")
		return err
	}

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

	for _, dd := range resp.DeviceDefinitions {
		for _, ud := range userDevices {
			if strings.EqualFold(dd.DeviceDefinitionId, ud.DeviceDefinitionID) {
				ud.DefinitionID = dd.NameSlug

				_, err := ud.Update(ctx, dbs.Writer, boil.Infer())
				if err != nil {
					s.logger.Error().Err(err).Msg("failed to update user device")
					return err
				}

			}
		}
	}

	return nil
}
