package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
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

type partialTeslaClaims struct {
	jwt.RegisteredClaims
	Scopes []string `json:"scp"`
}

func checkVirtualKeys(settings *config.Settings, pdb db.Store, logger *zerolog.Logger, cipher shared.Cipher) error {
	userDeviceID := os.Args[2]

	log := logger.With().Str("userDeviceId", userDeviceID).Logger()

	fleetAPI, err := services.NewTeslaFleetAPIService(settings, logger)
	if err != nil {
		return err
	}

	ctx := context.Background()

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve Tesla jobs: %w", err)
	}

	if !udai.AccessToken.Valid || !udai.AccessExpiresAt.Valid || udai.R.UserDevice.R.VehicleTokenSyntheticDevice == nil || udai.R.UserDevice.R.VehicleTokenSyntheticDevice.TokenID.IsZero() {
		// Never active, don't care.
		return nil
	}

	if udai.Status == models.UserDeviceAPIIntegrationStatusAuthenticationFailure || udai.AccessExpiresAt.Time.Before(time.Now()) {
		log.Info().Msg("In authentication failure.")
		return nil
	}

	var md services.UserDeviceAPIIntegrationsMetadata
	err = udai.Metadata.Unmarshal(&md)
	if err != nil {
		return fmt.Errorf("couldn't unmarshal metadata: %w", err)
	}

	if md.TeslaAPIVersion != 2 {
		log.Info().Msg("Still on Owner API.")
		return nil
	}

	token, err := cipher.Decrypt(udai.AccessToken.String)
	if err != nil {
		return fmt.Errorf("couldn't decrypt access token: %w", err)
	}

	var claims partialTeslaClaims
	_, _, err = jwt.NewParser().ParseUnverified(token, &claims)
	if err != nil {
		return fmt.Errorf("couldn't parse access token: %w", err)
	}

	for _, scope := range []string{"vehicle_device_data", "vehicle_location"} {
		if !slices.Contains(claims.Scopes, scope) {
			log.Info().Msg("Missing required scope.")
			return nil
		}
	}

	fs, err := fleetAPI.VirtualKeyConnectionStatus(ctx, token, udai.R.UserDevice.VinIdentifier.String)
	if err != nil {
		log.Err(err).Msg("Error checking virtual key status.")
		return err
	}

	if !fs.DiscountedDeviceData && !fs.KeyPaired {
		log.Info().Msg("Missing virtual key.")
		return nil
	}

	return nil
}
