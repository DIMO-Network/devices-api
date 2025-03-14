package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type enableTelemetryCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
	cipher   shared.Cipher
}

func (*enableTelemetryCmd) Name() string { return "enable-telemetry" }
func (*enableTelemetryCmd) Synopsis() string {
	return "xdd"
}
func (*enableTelemetryCmd) Usage() string {
	return `xpp
  `
}

// nolint
func (p *enableTelemetryCmd) SetFlags(f *flag.FlagSet) {

}

func (p *enableTelemetryCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := enableTelemetry(&p.settings, p.pdb, &p.logger, p.cipher)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

func enableTelemetry(settings *config.Settings, pdb db.Store, logger *zerolog.Logger, cipher shared.Cipher) error {
	userDeviceID := os.Args[2]

	ctx := context.Background()

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	if !udai.R.UserDevice.VinConfirmed || udai.R.UserDevice.R.VehicleTokenSyntheticDevice == nil || udai.R.UserDevice.R.VehicleTokenSyntheticDevice.TokenID.IsZero() {
		return errors.New("synthetic not minted")
	}

	fleetAPI, err := services.NewTeslaFleetAPIService(settings, logger)
	if err != nil {
		return err
	}

	var md services.UserDeviceAPIIntegrationsMetadata
	err = udai.Metadata.Unmarshal(&md)
	if err != nil {
		logger.Warn().Str("userDeviceId", udai.UserDeviceID).Msg("Couldn't parse metadata, skipping.")
		return err
	}

	if md.TeslaAPIVersion != 2 {
		return fmt.Errorf("tesla version not %d", md.TeslaAPIVersion)
	}

	token, err := cipher.Decrypt(udai.AccessToken.String)
	if err != nil {
		return err
	}

	return fleetAPI.SubscribeForTelemetryData(ctx, token, udai.R.UserDevice.VinIdentifier.String)
}
