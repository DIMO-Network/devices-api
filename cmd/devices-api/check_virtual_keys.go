package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

type checkVirtualKeyCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
	cipher   shared.Cipher
}

func (*checkVirtualKeyCmd) Name() string { return "check-virtual-key" }
func (*checkVirtualKeyCmd) Synopsis() string {
	return "xdd"
}
func (*checkVirtualKeyCmd) Usage() string {
	return `xpp
  `
}

// nolint
func (p *checkVirtualKeyCmd) SetFlags(f *flag.FlagSet) {

}

func (p *checkVirtualKeyCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := checkVirtualKeys(&p.settings, p.pdb, &p.logger, p.cipher)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

type fleetStatusReq struct {
	VINs []string `json:"vins"`
}

type fleetStatusResp struct {
	Response struct {
		KeyPairedVINs []string `json:"key_paired_vins"`
		UnpairedVINs  []string `json:"unpaired_vins"`
		VehicleInfo   map[string]struct {
			FirmwareVersion                string `json:"firmware_version"`
			VehicleCommandProtocolRequired bool   `json:"vehicle_command_protocol_required"`
		} `json:"vehicle_info"`
	} `json:"response"`
}

func checkVirtualKeys(settings *config.Settings, pdb db.Store, logger *zerolog.Logger, cipher shared.Cipher) error {
	logger.Info().Msgf("Checking virtual key status.")

	ctx := context.Background()

	udais, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
		models.UserDeviceAPIIntegrationWhere.AccessToken.IsNotNull(),
		models.UserDeviceAPIIntegrationWhere.Metadata.IsNotNull(),
		models.UserDeviceAPIIntegrationWhere.AccessExpiresAt.GT(null.TimeFrom(time.Now())),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve Tesla jobs: %w", err)
	}
	baseURL, err := url.ParseRequestURI(settings.TeslaFleetURL)
	if err != nil {
		panic(err)
	}

	ur := baseURL.JoinPath("api/1/vehicles/fleet_status")

	checked := 0
	worked := 0

	for _, udai := range udais {
		if !udai.R.UserDevice.VinConfirmed {
			continue
		}

		if !udai.R.UserDevice.VinIdentifier.Valid || len(udai.R.UserDevice.VinIdentifier.String) != 17 {
			logger.Warn().Str("userDeviceId", udai.UserDeviceID).Msg("Weird VIN, skipping.")
			continue
		}

		var md services.UserDeviceAPIIntegrationsMetadata
		err := udai.Metadata.Unmarshal(&md)
		if err != nil {
			logger.Warn().Str("userDeviceId", udai.UserDeviceID).Msg("Couldn't parse metadata, skipping.")
			continue
		}

		if md.TeslaAPIVersion != 2 {
			continue
		}

		token, err := cipher.Decrypt(udai.AccessToken.String)
		if err != nil {
			return err
		}

		vKeyGood, err := checkVehicle(ur.String(), udai.R.UserDevice.VinIdentifier.String, token)
		if err != nil {
			logger.Err(err).Str("userDeviceId", udai.UserDeviceID).Msg("Failed to check status.")
		}

		checked++
		if vKeyGood {
			worked++
		}
	}

	logger.Info().Msgf("Checked %d, with %d working", checked, worked)

	return nil
}

func checkVehicle(uri, vin, token string) (bool, error) {
	reqBody := fleetStatusReq{
		VINs: []string{vin},
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqBytes))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status code %d", res.StatusCode)
	}

	respBytes, err := io.ReadAll(res.Body)

	var resBody fleetStatusResp

	if err := json.Unmarshal(respBytes, &resBody); err != nil {
		return false, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return len(resBody.Response.KeyPairedVINs) != 0, nil
}
