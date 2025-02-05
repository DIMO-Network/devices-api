package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
)

const (
	TeslaIntegrationID = "26A5Dk3vvvQutjSyF0Jka2DP5lg"
	TeslaMappingTopic  = "topic.device.integration.mapping.tesla"
)

type populateTeslaTelemetryMapCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	producer  sarama.SyncProducer
	container dependencyContainer
}

func (*populateTeslaTelemetryMapCmd) Name() string { return "populate-tesla-telemetry-map" }
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
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(TeslaIntegrationID),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve active polling jobs: %w", err)
	}

	type userDeviceIDData struct {
		UserDeviceID   string `json:"userDeviceId"`
		VehicleTokenID int64  `json:"vehicleTokenID"`
		Type           string `json:"type"`
	}

	for _, udai := range udais {
		if !udai.R.UserDevice.VinConfirmed {
			continue
		}

		if !udai.R.UserDevice.VinIdentifier.Valid || len(udai.R.UserDevice.VinIdentifier.String) != 17 {
			logger.Warn().Str("userDeviceId", udai.UserDeviceID).Msg("Weird VIN, skipping.")
			continue
		}

		VIN := udai.R.UserDevice.VinIdentifier.String

		if udai.R.UserDevice.TokenID.IsZero() {
			logger.Warn().Str("userDeviceId", udai.UserDeviceID).Str("vin", VIN).Msg("invalid vehicle token id")
			continue
		}

		vID, ok := udai.R.UserDevice.TokenID.Int64()
		if !ok {
			logger.Warn().Str("userDeviceId", udai.UserDeviceID).Str("vin", VIN).Msg("failed to parse vehicle token id")
			continue
		}

		udid := userDeviceIDData{
			UserDeviceID:   udai.UserDeviceID,
			VehicleTokenID: vID,
			Type:           "Add",
		}

		b, err := json.Marshal(shared.CloudEvent[userDeviceIDData]{Data: udid})
		if err != nil {
			return err
		}

		msg := &sarama.ProducerMessage{
			Topic: TeslaMappingTopic,
			Key:   sarama.StringEncoder(VIN),
			Value: sarama.ByteEncoder(b),
		}
		if _, _, err := producer.SendMessage(msg); err != nil {
			return fmt.Errorf("couldn't send message for vin %s; userDeviceID: %s: %w", VIN, udai.UserDeviceID, err)
		}
	}

	return nil
}
