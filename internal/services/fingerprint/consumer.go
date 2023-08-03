package fingerprint

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services/issuer"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/kafka"
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

func NewConsumer(dbs db.Store, iss *issuer.Issuer, log *zerolog.Logger) *Consumer {
	return &Consumer{
		DBS:    dbs,
		logger: log,
		iss:    iss,
	}
}

func RunConsumer(ctx context.Context, settings *config.Settings, logger *zerolog.Logger, i *issuer.Issuer, dbs db.Store) error {
	consumer := NewConsumer(dbs, i, logger)

	if err := kafka.Consume(ctx, kafka.Config{
		Brokers: strings.Split(settings.KafkaBrokers, ","),
		Topic:   settings.DeviceFingerprintTopic,
		Group:   settings.DeviceFingerprintConsumerGroup,
	}, consumer.HandleDeviceFingerprint, logger); err != nil {
		logger.Fatal().Err(err).Msg("couldn't start device fingerprint consumer")
	}

	if err := kafka.Consume(ctx, kafka.Config{
		Brokers: strings.Split(settings.KafkaBrokers, ","),
		Topic:   settings.SyntheticFingerprintTopic,
		Group:   settings.SyntheticFingerprintConsumerGroup,
	}, consumer.HandleSyntheticFingerprint, logger); err != nil {
		logger.Fatal().Err(err).Msg("couldn't start synthetic fingerprint consumer")
	}

	logger.Info().Msg("Starting transaction request status listener.")

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

	observedVIN, err := ExtractVIN(event.Data)
	if err != nil {
		if err == ErrNoVIN {
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
		c.logger.Warn().Msgf("observed vin %s does not match verified vin %s for device %s", observedVIN, vn.Vin, vn.UserDeviceID.String)
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

func (c *Consumer) HandleSyntheticFingerprint(ctx context.Context, event *Event) error {
	if !common.IsHexAddress(event.Subject) {
		return fmt.Errorf("subject %q not a valid address", event.Subject)
	}

	observedVIN, err := ExtractVIN(event.Data)
	if err != nil {
		if err == ErrNoVIN {
			return nil
		}
		return fmt.Errorf("couldn't extract VIN: %w", err)
	}

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.WalletAddress.EQ(common.FromHex(event.Subject)),
		qm.Load(qm.Rels(models.SyntheticDeviceRels.VehicleToken, models.VehicleNFTRels.Claim)),
	).One(ctx, c.DBS.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed querying for synthetic device: %w", err)
	}

	vn := sd.R.VehicleToken
	if vn == nil {
		c.logger.Warn().Str("synthetic-wallet-addr", common.Bytes2Hex(sd.WalletAddress)).Msg("no vehicle nft associated with device")
		return nil
	}

	if observedVIN != vn.Vin {
		c.logger.Warn().Msgf("observed vin %s does not match verified vin %s for device %s", observedVIN, vn.Vin, vn.UserDeviceID.String)
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

var ErrNoVIN = errors.New("no VIN field")

var basicVINExp = regexp.MustCompile(`^[A-Z0-9]{17}$`)

// ExtractVIN extracts the vin field from a status update's data object.
// If this field is not present or fails basic validation, an error is returned.
// The function does clean up the input slightly.
func ExtractVIN(data []byte) (string, error) {
	partialData := new(struct {
		VIN *string `json:"vin"`
	})

	if err := json.Unmarshal(data, partialData); err != nil {
		return "", fmt.Errorf("failed parsing data field: %w", err)
	}

	if partialData.VIN == nil {
		return "", ErrNoVIN
	}

	// Minor cleaning.
	vin := strings.ToUpper(strings.ReplaceAll(*partialData.VIN, " ", ""))

	// We have seen crazy VINs like "\u000" before.
	if !basicVINExp.MatchString(vin) {
		return "", errors.New("invalid VIN")
	}

	return vin, nil
}
