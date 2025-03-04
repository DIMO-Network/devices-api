package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

type checkTelemetryCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
	cipher   shared.Cipher
}

func (*checkTelemetryCmd) Name() string { return "check-telemetry" }
func (*checkTelemetryCmd) Synopsis() string {
	return "xdd"
}
func (*checkTelemetryCmd) Usage() string {
	return `xpp
  `
}

// nolint
func (p *checkTelemetryCmd) SetFlags(f *flag.FlagSet) {

}

func (p *checkTelemetryCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := checkVirtualKeys(&p.settings, p.pdb, &p.logger, p.cipher)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error checking Telemetry capability.")
	}
	return subcommands.ExitSuccess
}

func checkVirtualKeys(settings *config.Settings, pdb db.Store, logger *zerolog.Logger, cipher shared.Cipher) error {
	userDeviceID := os.Args[2]

	fleetAPI, err := services.NewTeslaFleetAPIService(settings, logger)
	if err != nil {
		return err
	}

	ctx := context.Background()

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve Tesla jobs: %w", err)
	}

	if udai.AccessExpiresAt.Time.Before(time.Now()) {
		return fmt.Errorf("access token expired %s ago", time.Since(udai.AccessExpiresAt.Time))
	}

	var md services.UserDeviceAPIIntegrationsMetadata
	err = udai.Metadata.Unmarshal(&md)
	if err != nil {
		return err
	}

	token, err := cipher.Decrypt(udai.AccessToken.String)
	if err != nil {
		return fmt.Errorf("couldn't decrypt access token: %w", err)
	}

	teslaID, err := strconv.Atoi(udai.ExternalID.String)
	if err != nil {
		return err
	}

	v, err := fleetAPI.GetVehicle(ctx, token, teslaID)
	if err != nil {
		return err
	}

	// ss, err := fleetAPI.GetTelemetrySubscriptionStatus(ctx, token, teslaID)
	// if err != nil {
	// 	return err
	// }

	fs, err := fleetAPI.VirtualKeyConnectionStatus(ctx, token, v.VIN)
	if err != nil {
		return err
	}

	md.TeslaVIN = v.VIN
	md.TeslaDiscountedData = &fs.DiscountedDeviceData

	if err := udai.Metadata.Marshal(md); err != nil {
		return err
	}

	_, err = udai.Update(ctx, pdb.DBS().Writer, boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Metadata, models.UserDeviceAPIIntegrationColumns.UpdatedAt))
	if err != nil {
		return err
	}

	logger.Info().Str("userDeviceId", userDeviceID).Interface("vehicle", v).Interface("fleet", fs).Msg("Returned information.")

	return nil
}
