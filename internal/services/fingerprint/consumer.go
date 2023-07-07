package fingerprint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/issuer"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
			var event DeviceFingerprintCloudEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Couldn't parse fingerprint event.")
			} else {
				if err := c.Handle(session.Context(), &event); err != nil {
					c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Failed to process fingerprint event.")
				}
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func RunConsumer(ctx context.Context, settings *config.Settings, logger *zerolog.Logger, i *issuer.Issuer, dbs db.Store) error {
	kc := sarama.NewConfig()
	kc.Version = sarama.V3_3_1_0

	group, err := sarama.NewConsumerGroup(strings.Split(settings.KafkaBrokers, ","), settings.DeviceFingerprintConsumerGroup, kc)
	if err != nil {
		return err
	}

	c := &Consumer{logger: logger, iss: i, DBS: dbs}

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

func (c *Consumer) Handle(ctx context.Context, event *DeviceFingerprintCloudEvent) error {
	if !common.IsHexAddress(event.Subject) {
		return fmt.Errorf("subject %q not a valid address", event.Subject)
	}
	addr := common.HexToAddress(event.Subject)
	data, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	signature := common.FromHex(event.Signature)
	hash := crypto.Keccak256Hash(data)
	if recAddr, err := helpers.Ecrecover(hash, signature); err != nil {
		return err
	} else if recAddr != addr {
		err := errors.New("dervied address does not match expected")
		c.logger.Info().Err(err).Msg("invalid signature")
		return err
	}

	observedVIN, err := services.ExtractVIN(data)
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
	if err == nil {
		c.logger.Info().Msgf("Issued VIN credential for vehicle %d using device %s.", vn.TokenID, addr)
	}

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

type CloudEventHeaders struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	SpecVersion string `json:"specversion"`
	Subject     string `json:"subject"`
	Time        string `json:"time"`
	Type        string `json:"type"`
	// Signature is an extension https://github.com/cloudevents/spec/blob/main/cloudevents/documented-extensions.md
	Signature string `json:"signature"`
}

type DeviceFingerprintCloudEvent struct {
	CloudEventHeaders
	Data Data `json:"data"`
}

type Data struct {
	CommonData
	Vin      string  `json:"vin"`
	Protocol string  `json:"protocol"`
	Odometer float64 `json:"odometer,omitempty"`
}

// CommonData common properties we want to send with every data payload
type CommonData struct {
	RpiUptimeSecs  int     `json:"rpiUptimeSecs,omitempty"`
	BatteryVoltage float64 `json:"batteryVoltage,omitempty"`
	Timestamp      int64   `json:"timestamp"`
}
