package fingerprint

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/DIMO-Network/devices-api/internal/services"

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
	logger.Info().Msg("Starting transaction request status listener.")

	return nil
}

// DefaultCredDuration is the default lifetime for VIN credentials. It's meant to cover the
// current reward week, but contains an extra day for safety.
var DefaultCredDuration = 8 * 24 * time.Hour

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

	observedVIN := ""
	var err error
	if event.Source == "macaron/fingerprint" {
		observedVIN, err = ExtractVINMacaronType1(string(event.Data))
	} else {
		observedVIN, err = ExtractVIN(event.Data)
	}
	if err != nil {
		if errors.Is(err, ErrNoVIN) {
			return nil
		}
		return fmt.Errorf("couldn't extract VIN: %w", err)
	}

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.EthereumAddress.EQ(addr.Bytes()),
		qm.Load(qm.Rels(models.AftermarketDeviceRels.VehicleToken, models.UserDeviceRels.Claim)),
	).One(ctx, c.DBS.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed querying for device: %w", err)
	}

	ud := ad.R.VehicleToken
	if ud == nil {
		return nil
	}

	if observedVIN != ud.VinIdentifier.String {
		c.logger.Warn().Msgf("observed vin %s does not match verified vin %s for device %s", observedVIN, ud.VinIdentifier.String, ud.ID)
		return nil
	}

	if vc := ud.R.Claim; vc != nil {
		weekEnd := NumToWeekEnd(GetWeekNum(time.Now()))
		if vc.ExpirationDate.After(weekEnd) {
			return nil
		}
	}

	if _, err := c.iss.VIN(observedVIN, ud.TokenID.Int(nil), time.Now().Add(DefaultCredDuration)); err != nil {
		return err
	}

	md := services.UserDeviceMetadata{}
	if err = ad.R.VehicleToken.Metadata.Unmarshal(&md); err != nil {
		c.logger.Error().Msgf("Could not unmarshal userdevice metadata for device: %s", ad.R.VehicleToken.ID)
		return err
	}

	var protocol *string
	if md.CANProtocol == nil {
		if event.Source == "macaron/fingerprint" {
			protocol, err = ExtractProtocolMacaronType1(string(event.Data))
		} else {
			protocol, err = ExtractProtocol(event.Data)
		}
		if err != nil {
			c.logger.Error().Err(err)
		}

		if err == nil && protocol != nil {
			md.CANProtocol = protocol
		}
	}

	err = ad.R.VehicleToken.Metadata.Marshal(&md)
	if err != nil {
		c.logger.Error().Msgf("could not marshal userdevice metadata for device: %s", ad.R.VehicleToken.ID)
		if protocol != nil {
			appmetrics.FingerprintRequestCount.With(prometheus.Labels{"protocol": *protocol, "status": "Failed"}).Inc()
		}
		return err
	}

	if protocol != nil {
		appmetrics.FingerprintRequestCount.With(prometheus.Labels{"protocol": *protocol, "status": "Success"}).Inc()
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

// ExtractVINMacaronType1 extracts the vin field from message type 1
func ExtractVINMacaronType1(data string) (string, error) {
	// Decode base64 data
	decodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("failed to decode macaron base64 data: %w", err)
	}

	// Verify the length of decodedBytes: 1 byte header, 4 bytes timestamp, 8 bytes location, 1 byte protocol, 17 bytes VIN
	if len(decodedBytes) < 14+17 {
		return "", errors.New("decoded bytes too short to decode VIN from macaron")
	}
	// Extract VIN bytes
	vinStart := 1 + 4 + 8 + 1
	vinBytes := decodedBytes[vinStart : vinStart+17]
	vin := string(vinBytes)

	// We have seen crazy VINs like "\u000" before.
	if !basicVINExp.MatchString(vin) {
		return "", errors.New("invalid VIN from macaron")
	}

	return vin, nil
}

// ExtractProtocolMacaronType1 pulls out the can protocol from macaron message type 1
func ExtractProtocolMacaronType1(data string) (*string, error) {

	decodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}
	if len(decodedBytes) < 14 {
		return nil, errors.New("decoded bytes too short to decode protocol")
	}

	//Extract protocol
	protocolByte := decodedBytes[1+4+8]
	protocol := fmt.Sprintf("%02x", protocolByte)

	return &protocol, nil
}

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

func ExtractProtocol(data []byte) (*string, error) {
	partialData := new(struct {
		Protocol *string `json:"protocol"`
	})

	if err := json.Unmarshal(data, partialData); err != nil {
		return nil, fmt.Errorf("failed parsing data field: %w", err)
	}

	return partialData.Protocol, nil
}
