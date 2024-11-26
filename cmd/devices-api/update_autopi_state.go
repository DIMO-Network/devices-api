package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/DIMO-Network/devices-api/internal/constants"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

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

func (*updateStateCmd) Name() string { return "autopi-notify-status" }
func (*updateStateCmd) Synopsis() string {
	return "syncs our autopi integration status to their cloud. also syncs Name + MMY"
}
func (*updateStateCmd) Usage() string {
	return `autopi-notify-status`
}

// nolint
func (p *updateStateCmd) SetFlags(f *flag.FlagSet) {

}

func (p *updateStateCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	autoPiSvc := services.NewAutoPiAPIService(&p.settings, p.pdb.DBS)
	ddSvc := services.NewDeviceDefinitionService(p.pdb.DBS, &p.logger, &p.settings)
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
	// get all autopi paired devices
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
		reg := ""
		ci := constants.FindCountry(apiInt.R.UserDevice.CountryCode.String)
		if ci != nil {
			reg = ci.Region
		}
		err := autoPiSvc.UpdateState(apiInt.ExternalID.String, apiInt.Status, apiInt.R.UserDevice.CountryCode.String, reg)
		if err != nil {
			logger.Err(err).Msgf("failed to update status when calling autopi api for deviceId: %s", apiInt.ExternalID.String)
		} else {
			logger.Info().Msgf("successfully updated state for %s", apiInt.ExternalID.String)
		}
		time.Sleep(500)
		// also update the AP vehicle Call Name to make it easier to find in AP dashboard
		autoPiDevice, err := autoPiSvc.GetDeviceByUnitID(apiInt.Serial.String)
		if err == nil {
			dd, _ := deviceDefSvc.GetDeviceDefinitionBySlug(ctx, apiInt.R.UserDevice.DefinitionID)
			nm := services.BuildCallName(apiInt.R.UserDevice.Name.Ptr(), dd)
			err = autoPiSvc.PatchVehicleProfile(autoPiDevice.Vehicle.ID, services.PatchVehicleProfile{
				CallName: &nm,
			})
			if err != nil {
				logger.Err(err).Msgf("unable to patch vehicle profile. unitID: %s, callname: %s", apiInt.Serial.String, nm)
			} else {
				logger.Info().Msgf("also updated callname to: %s", nm)
			}
		} else {
			logger.Err(err).Msgf("could not get device by unitID: %s", apiInt.Serial.String)
		}
	}

	return nil
}
