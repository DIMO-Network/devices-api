package main

import (
	"context"
	"flag"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type autopiClearVINCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*autopiClearVINCmd) Name() string { return "autopi-clear-vin" }
func (*autopiClearVINCmd) Synopsis() string {
	return "iterates over all our autopi units and clears the VIN on autopi cloud if has a value."
}
func (*autopiClearVINCmd) Usage() string {
	return `autopi-clear-vin`
}

// nolint
func (p *autopiClearVINCmd) SetFlags(f *flag.FlagSet) {

}

func (p *autopiClearVINCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	clearVINFromAutopi(ctx, &p.logger, &p.settings, p.pdb)
	return subcommands.ExitSuccess
}

// clearVINFromAutopi iterates over all our known AutoPi units and sets the vin to blank, only if it has a value in their autopi profile.
func clearVINFromAutopi(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb db.Store) {
	// instantiate
	autoPiSvc := services.NewAutoPiAPIService(settings, pdb.DBS)

	// iterate all autopi units
	all, err := models.AutopiUnits().All(ctx, pdb.DBS().Reader)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to query db")
	}
	logger.Info().Msgf("processing %d autopi units", len(all))

	for _, unit := range all {
		innerLogger := logger.With().Str("autopiUnitID", unit.AutopiUnitID).Logger()

		autoPiDevice, err := autoPiSvc.GetDeviceByUnitID(unit.AutopiUnitID)
		if err != nil {
			innerLogger.Err(err).Msg("failed to call autopi api to get autoPiDevice")
			continue
		}
		if autoPiDevice != nil && len(autoPiDevice.Vehicle.Vin) > 1 {
			// call api svc to update profile, setting vin = ""
			err = autoPiSvc.PatchVehicleProfile(autoPiDevice.Vehicle.ID, services.PatchVehicleProfile{
				Vin: "",
			})
			if err != nil {
				// uh oh spaghettie oh
				innerLogger.Err(err).Msg("failed to set VIN on autopi service")
			} else {
				innerLogger.Info().Msgf("cleared vin for unit: %s", unit.AutopiUnitID)
			}
		}
	}

	logger.Info().Msg("all done")
}
