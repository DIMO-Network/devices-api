package main

import (
	"context"
	"flag"
	"fmt"
	"time"

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

	err := s.processDeviceDefinitions(ctx)

	if err != nil {
		s.logger.Fatal().Err(err).Msg("failed to sync user devices with tableland")
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (s *syncUserDeviceDeviceDefinitionCmd) processDeviceDefinitions(ctx context.Context) error {
	cursor := ""
	hasMore := true

	conn, err := grpc.NewClient(s.settings.DefinitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to close connection")
		}
	}(conn)

	for hasMore {
		userDevices, err := models.UserDevices(
			models.UserDeviceWhere.DefinitionID.IsNull(),
			models.UserDeviceWhere.ID.GT(cursor),
			qm.Limit(1000),
			qm.OrderBy(models.UserDeviceColumns.ID),
		).All(ctx, s.pdb.DBS().Writer)

		if err != nil {
			return err
		}

		if len(userDevices) == 0 {
			break
		}

		cursor = userDevices[len(userDevices)-1].ID

		for _, ud := range userDevices {

			deviceDefinitions, err := s.getDeviceDefinitionByID(ctx, conn, ud.DeviceDefinitionID)

			if err != nil {
				s.logger.Err(err).Msgf("failed to sync user device definition to tableland id: dd_id %s", ud.DeviceDefinitionID)
				continue
			}

			dd := deviceDefinitions.DeviceDefinitions[0]

			ud.DefinitionID = null.StringFrom(dd.NameSlug)

			_, err = ud.Update(ctx, s.pdb.DBS().Writer, boil.Whitelist(models.UserDeviceColumns.DefinitionID, models.UserDeviceColumns.UpdatedAt))

			if err != nil {
				s.logger.Err(err).Msgf("failed to udpdate user device with tableland id: user device id: %s", ud.ID)
			}

			time.Sleep(100 * time.Millisecond)
		}
		fmt.Println("processed 1000 user devices")

		if len(userDevices) < 1000 {
			hasMore = false
		}
	}

	return nil
}

func (s *syncUserDeviceDeviceDefinitionCmd) getDeviceDefinitionByID(ctx context.Context, conn *grpc.ClientConn, deviceDefinitionID string) (*ddgrpc.GetDeviceDefinitionResponse, error) {
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	resp, err := definitionsClient.GetDeviceDefinitionByID(ctx, &ddgrpc.GetDeviceDefinitionRequest{Ids: []string{deviceDefinitionID}})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
