package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
	err := setCommandCompatibility(ctx, p.pdb, p.ddSvc)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed during command compatibility fill.")
	}
	p.logger.Info().Msg("Finished setting command compatibility.")

	return subcommands.ExitSuccess
}

var teslaEnabledCommands = []string{constants.DoorsLock, constants.DoorsUnlock, constants.TrunkOpen, constants.FrunkOpen, constants.ChargeLimit}

func setCommandCompatibility(ctx context.Context, pdb db.Store, ddSvc services.DeviceDefinitionService) error {

	if err := setCommandCompatTesla(ctx, pdb, ddSvc); err != nil {
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
