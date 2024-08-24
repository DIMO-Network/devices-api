package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/models"
)

type populatePrivacyV2Topic struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	producer  sarama.SyncProducer
	container dependencyContainer
}

func (*populatePrivacyV2Topic) Name() string     { return "populate-privacy-topic-v2" }
func (*populatePrivacyV2Topic) Synopsis() string { return "populate-privacy-topic-v2 args to stdout." }
func (*populatePrivacyV2Topic) Usage() string {
	return `populate-privacy-topic-v2:
	populate-privacy-topic-v2 args.
  `
}

// nolint
func (p *populatePrivacyV2Topic) SetFlags(f *flag.FlagSet) {

}

func (p *populatePrivacyV2Topic) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	err := remakePrivacyV2Topic(&p.settings, p.pdb, p.producer, &p.logger)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

func remakePrivacyV2Topic(settings *config.Settings, pdb db.Store, producer sarama.SyncProducer, logger *zerolog.Logger) error {
	logger.Info().Msgf("Starting privacy processor v2 table population, sending to topic %s.", settings.PrivacyFenceTopicV2)

	ctx := context.Background()

	uds, err := models.UserDevices(models.UserDeviceWhere.TokenID.IsNotNull()).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to minted devices: %w", err)
	}

	for _, ud := range uds {
		fenceMap, err := models.UserDeviceToGeofences(
			models.UserDeviceToGeofenceWhere.UserDeviceID.EQ(ud.ID),
			qm.Load(models.UserDeviceToGeofenceRels.Geofence),
		).All(ctx, pdb.DBS().Reader)
		if err != nil {
			return fmt.Errorf("failed to find fences for device %s: %w", ud.ID, err)
		}

		if len(fenceMap) == 0 {
			continue
		}

		indexSet := shared.NewStringSet()

		for _, mapping := range fenceMap {
			for _, ind := range mapping.R.Geofence.H3Indexes {
				indexSet.Add(ind)
			}
		}

		if indexSet.Len() == 0 {
			continue
		}

		ce := shared.CloudEvent[controllers.FenceData]{
			ID:          ksuid.New().String(),
			Source:      "devices-api",
			SpecVersion: "1.0",
			Subject:     ud.ID,
			Time:        time.Now(),
			Type:        controllers.PrivacyFenceEventType,
			Data: controllers.FenceData{
				H3Indexes: indexSet.Slice(),
			},
		}
		b, err := json.Marshal(ce)
		if err != nil {
			return err
		}

		msg := &sarama.ProducerMessage{
			Topic: settings.PrivacyFenceTopicV2,
			Key:   sarama.StringEncoder(ud.TokenID.String()),
			Value: sarama.ByteEncoder(b),
		}
		if _, _, err := producer.SendMessage(msg); err != nil {
			return err
		}
	}

	return nil
}
