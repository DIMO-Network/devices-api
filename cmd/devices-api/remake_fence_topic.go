package main

import (
	"context"
	"encoding/json"
	"flag"
	"time"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type remakeFenceTopicCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	producer  sarama.SyncProducer
	container dependencyContainer
}

func (*remakeFenceTopicCmd) Name() string     { return "remake-fence-topic" }
func (*remakeFenceTopicCmd) Synopsis() string { return "remake-fence-topic args to stdout." }
func (*remakeFenceTopicCmd) Usage() string {
	return `remake-fence-topic:
	remake-fence-topic args.
  `
}

// nolint
func (p *remakeFenceTopicCmd) SetFlags(f *flag.FlagSet) {

}

func (p *remakeFenceTopicCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	err := remakeFenceTopic(&p.settings, p.pdb, p.producer)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

func remakeFenceTopic(settings *config.Settings, pdb db.Store, producer sarama.SyncProducer) error {
	ctx := context.Background()

	rels, err := models.UserDeviceToGeofences(
		qm.Load(models.UserDeviceToGeofenceRels.Geofence),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	deviceIDToIndexes := make(map[string]*shared.StringSet)

	for _, rel := range rels {
		if _, ok := deviceIDToIndexes[rel.UserDeviceID]; !ok {
			deviceIDToIndexes[rel.UserDeviceID] = shared.NewStringSet()
		}
		for _, ind := range rel.R.Geofence.H3Indexes {
			deviceIDToIndexes[rel.UserDeviceID].Add(ind)
		}
	}

	for userDeviceID, indexes := range deviceIDToIndexes {
		if indexes.Len() == 0 {
			continue
		}
		ce := shared.CloudEvent[controllers.FenceData]{
			ID:          ksuid.New().String(),
			Source:      "devices-api",
			SpecVersion: "1.0",
			Subject:     userDeviceID,
			Time:        time.Now(),
			Type:        controllers.PrivacyFenceEventType,
			Data: controllers.FenceData{
				H3Indexes: indexes.Slice(),
			},
		}
		b, err := json.Marshal(ce)
		if err != nil {
			return err
		}
		msg := &sarama.ProducerMessage{
			Topic: settings.PrivacyFenceTopic,
			Key:   sarama.StringEncoder(userDeviceID),
			Value: sarama.ByteEncoder(b),
		}
		if _, _, err := producer.SendMessage(msg); err != nil {
			return err
		}
	}

	return nil
}
