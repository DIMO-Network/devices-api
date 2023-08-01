package fingerprint

import (
	"context"
	"encoding/json"
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
				c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Couldn't parse fingerprint event.")
			} else {
				switch event.Type {
				case "zone.dimo.aftermarket.device.fingerprint":
					if err := c.HandleDeviceFingerprint(session.Context(), &event); err != nil {
						c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Failed to process device fingerprint event.")
					}
				case "zone.dimo.synthetic.device.fingerprint":
					if err := c.HandleSyntheticFingerprint(session.Context(), &event, string(message.Key[:])); err != nil {
						c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Failed to process synthetic fingerprint event.")
					}
				default:
					c.logger.Info().Int32("partition", message.Partition).Int64("offset", message.Offset).Str("type", event.Type).Msg("Unrecognized event type.")
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

	deviceGroup, err := sarama.NewConsumerGroup(strings.Split(settings.KafkaBrokers, ","), settings.DeviceFingerprintConsumerGroup, kc)
	if err != nil {
		return err
	}

	syntheticGroup, err := sarama.NewConsumerGroup(strings.Split(settings.KafkaBrokers, ","), settings.DeviceFingerprintConsumerGroup, kc)
	if err != nil {
		return err
	}

	c := &Consumer{logger: logger, iss: i, DBS: dbs}

	logger.Info().Msg("Starting transaction request status listener.")

	go func() {
		for {
			if err := deviceGroup.Consume(ctx, []string{settings.DeviceFingerprintTopic}, c); err != nil {
				logger.Warn().Err(err).Msgf("Consumer group session ended: %s", settings.DeviceFingerprintTopic)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	go func() {
		for {
			if err := syntheticGroup.Consume(ctx, []string{settings.SyntheticFingerprintTopic}, c); err != nil {
				logger.Warn().Err(err).Msgf("Consumer group session ended: %s", settings.SyntheticFingerprintTopic)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

func (c *Consumer) HandleDeviceFingerprint(ctx context.Context, event *Event) error {
	if !common.IsHexAddress(event.Subject) {
		return fmt.Errorf("subject %q not a valid address", event.Subject)
	}
	addr := common.HexToAddress(event.Subject)
	signature := common.FromHex(event.Signature)
	hash := crypto.Keccak256Hash(event.Data)

	if recAddr, err := helpers.Ecrecover(hash.Bytes(), signature); err != nil {
		return fmt.Errorf("failed to recover an address: %w", err)
	} else if recAddr != addr {
		return fmt.Errorf("recovered wrong address %s", recAddr)
	}

	observedVIN, err := services.ExtractVIN(event.Data)
	if err != nil {
		if err == services.ErrNoVIN {
			return nil
		}
		return fmt.Errorf("couldn't extract VIN: %w", err)
	}

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.EthereumAddress.EQ(addr.Bytes()),
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
		c.logger.Warn().Str("device", vn.UserDeviceID.String).Str("verified-vin", vn.Vin).Str("observed-vin", observedVIN).Msg("invalid vin")
		return nil
	}

	if vc := vn.R.Claim; vc != nil {
		weekEnd := NumToWeekEnd(GetWeekNum(time.Now()))
		if vc.ExpirationDate.After(weekEnd) {
			return nil
		}
	}

	if _, err := c.iss.VIN(observedVIN, vn.TokenID.Int(nil), time.Now().Add(8*24*time.Hour)); err != nil {
		return err
	}

	c.logger.Info().Str("device-addr", event.Subject).Msg("issued vin credential")

	return nil
}

func (c *Consumer) HandleSyntheticFingerprint(ctx context.Context, event *Event, key string) error {
	if !common.IsHexAddress(event.Subject) {
		return fmt.Errorf("subject %q not a valid address", event.Subject)
	}

	observedVIN, err := services.ExtractVIN(event.Data)
	if err != nil {
		if err == services.ErrNoVIN {
			return nil
		}
		return fmt.Errorf("couldn't extract VIN: %w", err)
	}

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(key),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.Claim)),
	).One(ctx, c.DBS.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed querying for device: %w", err)
	}

	vn := ud.R.VehicleNFT
	if vn == nil {
		return nil
	}

	if observedVIN != vn.Vin {
		c.logger.Warn().Str("device", ud.ID).Str("verified-vin", vn.Vin).Str("observed-vin", observedVIN).Msg("invalid vin")
		return nil
	}

	if vc := vn.R.Claim; vc != nil {
		weekEnd := NumToWeekEnd(GetWeekNum(time.Now()))
		if vc.ExpirationDate.After(weekEnd) {
			return nil
		}
	}

	if _, err := c.iss.VIN(observedVIN, vn.TokenID.Int(nil), time.Now().Add(8*24*time.Hour)); err != nil {
		return err
	}

	c.logger.Info().Str("device-addr", event.Subject).Msg("issued vin credential")

	return nil
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
