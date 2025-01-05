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
	KeyPairedVINs []string `json:"key_paired_vins"`
	UnpairedVINs  []string `json:"unpaired_vins"`
	VehicleInfo   map[string]struct {
		FirmwareVersion                string `json:"firmware_version"`
		VehicleCommandProtocolRequired bool   `json:"vehicle_command_protocol_required"`
	} `json:"vehicle_info"`
}

func checkVirtualKeys(settings *config.Settings, pdb db.Store, logger *zerolog.Logger, cipher shared.Cipher) error {
	logger.Info().Msgf("Starting synthetic device job enrichment, sending to topic %s.", settings.SDInfoTopic)

	ctx := context.Background()

	udais, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
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

		req := fleetStatusReq{
			VINs: []string{udai.R.UserDevice.VinIdentifier.String},
		}

		outB, _ := json.Marshal(req)

		hreq, _ := http.NewRequest("POST", ur.String(), bytes.NewBuffer(outB))
		hreq.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(hreq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code %d", resp.StatusCode)
		}

		respB, _ := io.ReadAll(resp.Body)

		var outBody fleetStatusResp
		err = json.Unmarshal(respB, &outBody)
		if err != nil {
			return err
		}

		checked += 1
		if len(outBody.KeyPairedVINs) != 0 {
			worked += 1
		}
	}

	fmt.Printf("Checked %d, with %d working\n", checked, worked)

	return nil
}
