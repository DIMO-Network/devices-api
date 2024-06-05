package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/sdtask"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
)

type populateSDInfoTopicCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	producer  sarama.SyncProducer
	container dependencyContainer
}

func (*populateSDInfoTopicCmd) Name() string     { return "populate-sd-info-topic" }
func (*populateSDInfoTopicCmd) Synopsis() string { return "populate-sd-info-topic args to stdout." }
func (*populateSDInfoTopicCmd) Usage() string {
	return `populate-sd-info-topic:
	populate-sd-info-topic args.
  `
}

// nolint
func (p *populateSDInfoTopicCmd) SetFlags(f *flag.FlagSet) {

}

func (p *populateSDInfoTopicCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	err := remakeSDInfoTopic(&p.settings, p.pdb, p.producer)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

func remakeSDInfoTopic(settings *config.Settings, pdb db.Store, producer sarama.SyncProducer) error {
	ctx := context.Background()

	udais, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNotNull(),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to retrieve active polling jobs: %w", err)
	}

	for _, udai := range udais {
		sd := udai.R.UserDevice.R.VehicleTokenSyntheticDevice
		if sd == nil || sd.TokenID.IsZero() {
			continue
		}

		tokenID, _ := sd.TokenID.Int64()
		vehicleTokenID, _ := sd.VehicleTokenID.Int64()
		integrationTokenID, _ := sd.IntegrationTokenID.Int64()

		out := sdtask.SyntheticDevice{
			TokenID:            int(tokenID),
			Address:            common.BytesToAddress(sd.WalletAddress),
			IntegrationTokenID: int(integrationTokenID),
			WalletChildNumber:  sd.WalletChildNumber,
			VehicleTokenID:     int(vehicleTokenID),
		}

		b, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("couldn't marshal payload for synthetic device %d: %w", tokenID, err)
		}

		msg := &sarama.ProducerMessage{
			Topic: settings.SDInfoTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(b),
		}
		if _, _, err := producer.SendMessage(msg); err != nil {
			return fmt.Errorf("couldn't send message for synthetic device %d: %w", tokenID, err)
		}
	}

	return nil
}
