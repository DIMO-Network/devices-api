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
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
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

type teslaGetVehicleRes struct {
	Response struct {
		VIN string `json:"vin"`
	}
}

type teslafleetStatusReq struct {
	VINs []string `json:"vins"`
}

type teslaFleetStatusRes struct {
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
	userDeviceID := os.Args[2]

	ctx := context.Background()

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve Tesla jobs: %w", err)
	}

	baseURL, err := url.ParseRequestURI(settings.TeslaFleetURL)
	if err != nil {
		panic(err)
	}

	token, err := cipher.Decrypt(udai.AccessToken.String)
	if err != nil {
		return fmt.Errorf("couldn't decrypt access token: %w", err)
	}

	var tgvRes teslaGetVehicleRes
	err = submitTeslaReq(
		"GET",
		baseURL.JoinPath("api/1/vehicles", udai.ExternalID.String).String(),
		token,
		nil,
		&tgvRes,
	)
	if err != nil {
		return fmt.Errorf("failed to double-check VIN: %w", err)
	}

	vin := tgvRes.Response.VIN

	logger.Info().Str("vin", vin).Msg("Retrieved VIN.")

	var tfsRes teslaFleetStatusRes
	err = submitTeslaReq(
		"POST",
		baseURL.JoinPath("api/1/vehicles/fleet_status").String(),
		token,
		teslafleetStatusReq{
			VINs: []string{vin},
		},
		&tfsRes,
	)
	if err != nil {
		return fmt.Errorf("failed to check virtual key status: %w", err)
	}

	logger.Info().Bool("keyPaired", len(tfsRes.Response.KeyPairedVINs) == 1).Str("firmwareVersion", tfsRes.Response.VehicleInfo[vin].FirmwareVersion).Bool("protocolRequired", tfsRes.Response.VehicleInfo[vin].VehicleCommandProtocolRequired).Msg("Checked virtual key status.")

	var claims partialTeslaClaims
	_, _, err = jwt.NewParser().ParseUnverified(token, &claims)
	if err != nil {
		return fmt.Errorf("couldn't parse JWT: %w", err)
	}

	logger.Info().Interface("scopes", claims.Scopes).Msg("Checked scopes.")

	return nil
}

type partialTeslaClaims struct {
	jwt.RegisteredClaims
	Scopes []string `json:"scp"`
}

func submitTeslaReq(method, uri, token string, reqBody any, respObj any) error {
	var buf io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, uri, buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	respBytes, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("status code %d", res.StatusCode)
	}

	if respObj != nil {
		err = json.Unmarshal(respBytes, respObj)
		if err != nil {
			return fmt.Errorf("couldn't unmarshal response body: %w", err)
		}
	}

	return nil
}
