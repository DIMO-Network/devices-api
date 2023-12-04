package main

import (
	"context"
	"encoding/json"
	"flag"
	"time"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
)

type remakeUserDeviceTokenTableCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	container dependencyContainer
}

type MapData struct {
	UserDeviceID   string `json:"userDeviceId"`
	VehicleTokenID int    `json:"vehicleTokenId"`
}

func (*remakeUserDeviceTokenTableCmd) Name() string { return "remake-user-device-token-table" }
func (*remakeUserDeviceTokenTableCmd) Synopsis() string {
	return "remake-user-device-token-table args to stdout."
}
func (*remakeUserDeviceTokenTableCmd) Usage() string {
	return `remake-user-device-token-table:
	remake-user-device-token-table args.
  `
}

// nolint
func (p *remakeUserDeviceTokenTableCmd) SetFlags(f *flag.FlagSet) {

}

func (p *remakeUserDeviceTokenTableCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := remakeUserDeviceTokenTable(ctx, p.pdb, p.container.getKafkaProducer())
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running AutoPi Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

// remakeAftermarketTopic re-populates the autopi ingest registrar topic based on data we have in user_device_api_integrations
func remakeUserDeviceTokenTable(ctx context.Context, pdb db.Store, producer sarama.SyncProducer) error {
	db := pdb.DBS().Reader

	vns, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.IsNotNull(),
		models.VehicleNFTWhere.UserDeviceID.IsNotNull(),
	).All(ctx, db)
	if err != nil {
		return err
	}

	for _, vn := range vns {
		tokenID, _ := vn.TokenID.Int64()

		out := &shared.CloudEvent[MapData]{
			ID:          ksuid.New().String(),
			Source:      "user-device-token-mapping-processor",
			SpecVersion: "1.0",
			Subject:     vn.UserDeviceID.String,
			Time:        time.Now(),
			Type:        "zone.dimo.device.token",
			Data: MapData{
				UserDeviceID:   vn.UserDeviceID.String,
				VehicleTokenID: int(tokenID),
			},
		}

		b, _ := json.Marshal(out)

		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: "table.device.token.mapping",
			Key:   sarama.StringEncoder(vn.UserDeviceID.String),
			Value: sarama.ByteEncoder(b),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
