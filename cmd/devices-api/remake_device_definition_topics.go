package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/google/subcommands"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type remakeDeviceDefinitionTopicsCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	producer  sarama.SyncProducer
	ddSvc     services.DeviceDefinitionService
	container dependencyContainer
}

func (*remakeDeviceDefinitionTopicsCmd) Name() string     { return "remake-dd-topics" }
func (*remakeDeviceDefinitionTopicsCmd) Synopsis() string { return "remake-dd-topics args to stdout." }
func (*remakeDeviceDefinitionTopicsCmd) Usage() string {
	return `remake-dd-topics:
	remake-dd-topics args.
  `
}

// nolint
func (p *remakeDeviceDefinitionTopicsCmd) SetFlags(f *flag.FlagSet) {

}

func (p *remakeDeviceDefinitionTopicsCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	err := remakeDeviceDefinitionTopics(ctx, &p.settings, p.pdb, p.producer, &p.logger, p.ddSvc)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error recreating device definition KTables.")
	}
	return subcommands.ExitSuccess
}

// remakeDeviceDefinitionTopics invokes [services.DeviceDefinitionRegistrar] for each user device
// with an integration.
func remakeDeviceDefinitionTopics(ctx context.Context, settings *config.Settings, pdb db.Store, producer sarama.SyncProducer, logger *zerolog.Logger, ddSvc services.DeviceDefinitionService) error {
	reg := services.NewDeviceDefinitionRegistrar(producer, settings)
	db := pdb.DBS().Reader

	// Find all integrations instances.
	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.ExternalID.IsNotNull(),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve integration instances: %w", err)
	}

	failures := 0

	ddIDs := shared.NewStringSet()

	for _, d := range apiInts {
		ddIDs.Add(d.R.UserDevice.DeviceDefinitionID)
	}

	// For each of these, register the device's device definition with the data pipeline.
	for _, apiInt := range apiInts {
		ddInfo, err := ddSvc.GetDeviceDefinitionByID(ctx, apiInt.R.UserDevice.DeviceDefinitionID)
		if err != nil {
			logger.Err(err).
				Str("userDeviceId", apiInt.UserDeviceID).
				Str("deviceDefinitionId", apiInt.R.UserDevice.DeviceDefinitionID).
				Msg("Failed to retrieve device definition.")
			continue
		}

		userDeviceID := apiInt.UserDeviceID

		region := ""

		if country := apiInt.R.UserDevice.CountryCode; country.Valid {
			countryData := constants.FindCountry(country.String)
			if countryData != nil {
				region = countryData.Region
			}
		}

		ddReg := services.DeviceDefinitionDTO{
			UserDeviceID:       userDeviceID,
			DeviceDefinitionID: ddInfo.DeviceDefinitionId,
			IntegrationID:      apiInt.IntegrationID,
			Make:               ddInfo.Type.Make,
			Model:              ddInfo.Type.Model,
			Year:               int(ddInfo.Type.Year),
			Region:             region,
			MakeSlug:           ddInfo.Type.MakeSlug,
			ModelSlug:          ddInfo.Type.ModelSlug,
		}

		err = reg.Register(ddReg)
		if err != nil {
			logger.Err(err).Str("userDeviceId", userDeviceID).Msg("Failed to register device's device definition.")
			failures++
		}
	}

	logger.Info().Int("attempted", len(apiInts)).Int("failed", failures).Msg("Finished device definition registration.")

	return nil
}
