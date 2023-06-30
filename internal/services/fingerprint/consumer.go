package fingerprint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/issuer"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Event struct {
	shared.CloudEvent[json.RawMessage]
	Signature string `json:"signature"`
}

type Consumer struct {
	logger *zerolog.Logger
	iss    *issuer.Issuer
	DBS    db.Store
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			var event Event
			if err := json.Unmarshal(message.Value, &event); err != nil {
				c.logger.Err(err).Msgf("Couldn't parse fingerprint event at partition %d, offset %d.", message.Partition, message.Offset)
			} else {
				if err := c.Handle(session.Context(), &event); err != nil {
					c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Failed to update vin credential status.")
				}
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func RunConsumer(ctx context.Context, settings *config.Settings, logger *zerolog.Logger, i *issuer.Issuer) error {
	kc := sarama.NewConfig()
	kc.Version = sarama.V3_3_1_0

	group, err := sarama.NewConsumerGroup(strings.Split(settings.KafkaBrokers, ","), settings.DeviceFingerprintConsumerGroup, kc)
	if err != nil {
		return err
	}

	c := &Consumer{logger: logger, iss: i}

	logger.Info().Msg("Starting transaction request status listener.")

	go func() {
		for {
			if err := group.Consume(ctx, []string{settings.DeviceFingerprintTopic}, c); err != nil {
				logger.Warn().Err(err).Msg("Consumer group session ended.")
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

func (c *Consumer) Handle(ctx context.Context, event *Event) error {
	if !common.IsHexAddress(event.Subject) {
		return fmt.Errorf("subject %q not a valid address", event.Subject)
	}

	addr := common.HexToAddress(event.Subject)

	observedVIN, err := services.ExtractVIN(event.CloudEvent.Data)
	if err != nil {
		return fmt.Errorf("couldn't extract VIN: %w", err)
	}

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.EthereumAddress.EQ(null.BytesFrom(addr.Bytes())),
		qm.Load(qm.Rels(models.AftermarketDeviceRels.VehicleToken, models.VehicleNFTRels.Claim)),
	).One(ctx, c.DBS.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed querying for device: %w", err)
	}

	vn := ad.R.VehicleToken
	if vn == nil {
		return nil
	}

	if observedVIN != vn.Vin {
		c.logger.Warn().Msgf("Observed VIN %s for vehicle %d with VIN %s.", observedVIN, vn.TokenID, vn.Vin)
		return nil
	}

	if vc := vn.R.Claim; vc != nil {
		weekEnd := NumToWeekEnd(GetWeekNum(time.Now()))
		if vc.ExpirationDate.After(weekEnd) {
			return nil
		}
	}

	_, err = c.iss.VIN(observedVIN, vn.TokenID.Int(nil))
	return err
}

var startTime = time.Date(2022, time.January, 31, 5, 0, 0, 0, time.UTC)
var weekDuration = 7 * 24 * time.Hour

func GetWeekNum(t time.Time) int {
	sinceStart := t.Sub(startTime)
	weekNum := int(sinceStart.Truncate(weekDuration) / weekDuration)
	return weekNum
}

func NumToWeekEnd(n int) time.Time {
	return startTime.Add(time.Duration(n+1) * weekDuration)
}
