package main

import (
	"context"
	"flag"
	"fmt"
	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

type updateStateCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*updateStateCmd) Name() string     { return "autopi-notify-status" }
func (*updateStateCmd) Synopsis() string { return "autopi-notify-status args to stdout." }
func (*updateStateCmd) Usage() string {
	return `autopi-notify-status [] <some text>:
	autopi-notify-status args.
  `
}

// nolint
func (p *updateStateCmd) SetFlags(f *flag.FlagSet) {

}

func (p *updateStateCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	autoPiSvc := services.NewAutoPiAPIService(&p.settings, p.pdb.DBS)
	ddSvc := services.NewDeviceDefinitionService(p.pdb.DBS, &p.logger, nil, &p.settings)
	err := updateState(ctx, p.pdb, &p.logger, autoPiSvc, ddSvc)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("failed to sync autopi notify status")
	}
	p.logger.Info().Msg("success")

	return subcommands.ExitSuccess
}

// updateStatus re-populates the autopi ingest registrar topic based on data we have in user_device_api_integrations
func updateState(ctx context.Context, pdb db.Store, logger *zerolog.Logger, autoPiSvc services.AutoPiAPIService, deviceDefSvc services.DeviceDefinitionService) error {
	reader := pdb.DBS().Reader

	const (
		autopiInteg = "27qftVRWQYpVDcO5DltO5Ojbjxk"
	)

	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autopiInteg),
		models.UserDeviceAPIIntegrationWhere.ExternalID.IsNotNull(),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).All(ctx, reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}
	logger.Info().Msgf("found %d connected autopis to update status for", len(apiInts))

	for _, apiInt := range apiInts {
		err := autoPiSvc.UpdateState(apiInt.ExternalID.String, apiInt.Status)
		if err != nil {
			logger.Err(err).Msgf("failed to update status when calling autopi api for deviceId: %s", apiInt.ExternalID.String)
		} else {
			logger.Info().Msgf("successfully updated state for %s", apiInt.ExternalID.String)
		}
		time.Sleep(500)

		autoPiDevice, err := autoPiSvc.GetDeviceByUnitID(apiInt.Serial.String)
		if err == nil {
			dd, _ := deviceDefSvc.GetDeviceDefinitionByID(ctx, apiInt.R.UserDevice.DeviceDefinitionID)
			nm := buildCallName(apiInt.R.UserDevice.Name.Ptr(), dd)
			_ = autoPiSvc.PatchVehicleProfile(autoPiDevice.Vehicle.ID, services.PatchVehicleProfile{
				CallName: &nm,
			})
		}
	}

	return nil
}

func buildCallName(callName *string, dd *ddgrpc.GetDeviceDefinitionItemResponse) string {
	uniquer := ksuid.New().String()[6:10]
	if dd == nil {
		if callName != nil {
			return *callName
		}
		return uniquer
	}
	mmy := fmt.Sprintf("%d %s %s", dd.Type.Year, dd.Type.MakeSlug, dd.Type.ModelSlug)
	if callName == nil {
		return uniquer + ":" + mmy
	}
	return *callName + ":" + mmy
}
