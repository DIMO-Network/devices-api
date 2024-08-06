package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
)

type populateTeslaTelemetryMapCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	producer  sarama.SyncProducer
	container dependencyContainer
}

func (*populateTeslaTelemetryMapCmd) Name() string { return "populate-sd-info-topic" }
func (*populateTeslaTelemetryMapCmd) Synopsis() string {
	return "populate-sd-info-topic args to stdout."
}
func (*populateTeslaTelemetryMapCmd) Usage() string {
	return `populate-tesla-telemetry-map:
	populate-tesla-telemetry-map args.
  `
}

// nolint
func (p *populateTeslaTelemetryMapCmd) SetFlags(f *flag.FlagSet) {

}

func (p *populateTeslaTelemetryMapCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	err := remakeTeslaTelemTopic(&p.settings, p.pdb, p.producer, &p.logger)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

func remakeTeslaTelemTopic(settings *config.Settings, pdb db.Store, producer sarama.SyncProducer, logger *zerolog.Logger) error {
	logger.Info().Msgf("Starting synthetic device job enrichment, sending to topic %s.", settings.SDInfoTopic)

	ctx := context.Background()

	udais, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve active polling jobs: %w", err)
	}

	type userDeviceIDData struct {
		UserDeviceID string `json:"userDeviceId"`
		Type         string `json:"type"`
	}

	for _, udai := range udais {
		if !udai.R.UserDevice.VinConfirmed {
			continue
		}

		if !udai.R.UserDevice.VinIdentifier.Valid || len(udai.R.UserDevice.VinIdentifier.String) != 17 {
			logger.Warn().Str("userDeviceId", udai.UserDeviceID).Msg("Weird VIN, skipping.")
			continue
		}

		udid := userDeviceIDData{
			UserDeviceID: udai.UserDeviceID,
			Type:         "Add",
		}

		b, err := json.Marshal(shared.CloudEvent[userDeviceIDData]{Data: udid})
		if err != nil {
			return err
		}

		msg := &sarama.ProducerMessage{
			Topic: "topic.device.integration.mapping.tesla",
			Key:   sarama.StringEncoder(udai.R.UserDevice.VinIdentifier.String),
			Value: sarama.ByteEncoder(b),
		}
		if _, _, err := producer.SendMessage(msg); err != nil {
			return fmt.Errorf("couldn't send message for UD %s: %w", udai.UserDeviceID, err)
		}
	}

	return nil
}
