package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/shared/db"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

type setCommandCompatibilityCmd struct {
	logger       zerolog.Logger
	settings     config.Settings
	pdb          db.Store
	eventService services.EventService
	ddSvc        services.DeviceDefinitionService
	container    dependencyContainer
}

func (*setCommandCompatibilityCmd) Name() string     { return "set-command-compat" }
func (*setCommandCompatibilityCmd) Synopsis() string { return "set-command-compat args to stdout." }
func (*setCommandCompatibilityCmd) Usage() string {
	return `set-command-compat [] <some text>:
	set-command-compat args.
  `
}

// nolint
func (p *setCommandCompatibilityCmd) SetFlags(f *flag.FlagSet) {

}

func (p *setCommandCompatibilityCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	p.eventService = services.NewEventService(&p.logger, &p.settings, p.container.getKafkaProducer())
	err := setCommandCompatibility(ctx, &p.settings, p.pdb, p.ddSvc)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed during command compatibility fill.")
	}
	p.logger.Info().Msg("Finished setting command compatibility.")

	return subcommands.ExitSuccess
}

var teslaEnabledCommands = []string{constants.DoorsLock, constants.DoorsUnlock, constants.TrunkOpen, constants.FrunkOpen, constants.ChargeLimit}

func setCommandCompatibility(ctx context.Context, settings *config.Settings, pdb db.Store, ddSvc services.DeviceDefinitionService) error {

	if err := setCommandCompatTesla(ctx, pdb, ddSvc); err != nil {
		return err
	}

	err := setCommandCompatSmartcar(ctx, settings, pdb, ddSvc)

	if err != nil {
		return err
	}

	return nil
}

func setCommandCompatTesla(ctx context.Context, pdb db.Store, ddSvc services.DeviceDefinitionService) error {
	teslaInt, err := ddSvc.GetIntegrationByVendor(ctx, constants.TeslaVendor)
	if err != nil {
		return err
	}

	teslaUDAIs, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(teslaInt.Id),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	for _, tu := range teslaUDAIs {
		md := new(services.UserDeviceAPIIntegrationsMetadata)
		if err := tu.Metadata.Unmarshal(md); err != nil {
			return err
		}

		md.Commands = &services.UserDeviceAPIIntegrationsMetadataCommands{Enabled: teslaEnabledCommands}

		if err := tu.Metadata.Marshal(md); err != nil {
			return err
		}

		if _, err := tu.Update(ctx, pdb.DBS().Writer, boil.Whitelist("metadata")); err != nil {
			return err
		}
	}

	return nil
}

func setCommandCompatSmartcar(ctx context.Context, settings *config.Settings, pdb db.Store, ddSvc services.DeviceDefinitionService) error {
	scInt, err := ddSvc.GetIntegrationByVendor(ctx, constants.SmartCarVendor)
	if err != nil {
		return err
	}

	scUDAIs, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scInt.Id),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice), // Need VIN and country.
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	for _, su := range scUDAIs {
		country := constants.FindCountry(su.R.UserDevice.CountryCode.String)
		doors, err := checkSmartcarDoorCompatibility(settings, su.R.UserDevice.VinIdentifier.String, country.Alpha2)
		if err != nil {
			log.Err(err).Msg("Error getting compat")
			continue
		}
		if !doors {
			continue
		}
		md := new(services.UserDeviceAPIIntegrationsMetadata)
		if err := su.Metadata.Unmarshal(md); err != nil {
			return err
		}

		if md.Commands == nil {
			md.Commands = new(services.UserDeviceAPIIntegrationsMetadataCommands)
		}

		if len(md.Commands.Enabled) != 0 {
			continue
		}

		md.Commands.Capable = []string{constants.DoorsLock, constants.DoorsUnlock}

		if err := su.Metadata.Marshal(md); err != nil {
			return err
		}

		if _, err := su.Update(ctx, pdb.DBS().Writer, boil.Whitelist("metadata")); err != nil {
			return err
		}
	}

	return nil
}

type capResp struct {
	Capabilities []struct {
		Permission string `json:"permission"`
		Capable    bool   `json:"capable"`
	} `json:"capabilities"`
}

func checkSmartcarDoorCompatibility(settings *config.Settings, vin, countryAlpha2 string) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.smartcar.com/v2.0/compatibility?vin=%s&scope=control_security&country=%s", vin, countryAlpha2), nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(settings.SmartcarClientID+":"+settings.SmartcarClientSecret)))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status code %d", resp.StatusCode)
	}

	rb := new(capResp)
	if err := json.NewDecoder(resp.Body).Decode(rb); err != nil {
		return false, err
	}

	if len(rb.Capabilities) == 0 {
		return false, errors.New("no capabilities in response")
	}

	return rb.Capabilities[0].Capable, nil
}
