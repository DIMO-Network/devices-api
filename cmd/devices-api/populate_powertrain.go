package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type populateUSAPowertrainCmd struct {
	logger       zerolog.Logger
	settings     config.Settings
	pdb          db.Store
	nhtsaService services.INHTSAService
}

func (*populateUSAPowertrainCmd) Name() string     { return "populate-usa-powertrain" }
func (*populateUSAPowertrainCmd) Synopsis() string { return "populate-usa-powertrain args to stdout." }
func (*populateUSAPowertrainCmd) Usage() string {
	return `populate-usa-powertrain:
	populate-usa-powertrain args.
  `
}

// nolint
func (p *populateUSAPowertrainCmd) SetFlags(f *flag.FlagSet) {

}

func (p *populateUSAPowertrainCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.logger.Info().Msg("Populating USA powertrain data from VINs")
	err := populateUSAPowertrain(ctx, &p.logger, p.pdb, p.nhtsaService)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error filling in powertrain data.")
	}
	return subcommands.ExitSuccess
}

func populateUSAPowertrain(ctx context.Context, logger *zerolog.Logger, pdb db.Store, nhtsaService services.INHTSAService) error {
	devices, err := models.UserDevices(
		models.UserDeviceWhere.CountryCode.EQ(null.StringFrom("USA")),
		models.UserDeviceWhere.VinIdentifier.IsNotNull(),
	).All(ctx, pdb.DBS().Writer)
	if err != nil {
		return err
	}

	process := func(device *models.UserDevice) error {
		resp, err := nhtsaService.DecodeVIN(device.VinIdentifier.String)
		if err != nil {
			return err
		}
		dt, err := resp.DriveType()
		if err != nil {
			return err
		}

		md := new(services.UserDeviceMetadata)
		if err := device.Metadata.Unmarshal(md); err != nil {
			return err
		}

		md.PowertrainType = &dt
		if err := device.Metadata.Marshal(md); err != nil {
			return err
		}
		if _, err := device.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}
		return nil
	}

	for _, device := range devices {
		if err := process(device); err != nil {
			logger.Err(err).Str("userDeviceId", device.ID).Str("vin", device.VinIdentifier.String).Msg("Failed to update powertrain metadata.")
		}
	}

	return nil
}
