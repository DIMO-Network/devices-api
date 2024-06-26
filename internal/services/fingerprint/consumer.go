package fingerprint

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/DIMO-Network/devices-api/internal/services"

	"github.com/pkg/errors"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
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
	DBS    db.Store
}

func NewConsumer(dbs db.Store, log *zerolog.Logger) *Consumer {
	return &Consumer{
		DBS:    dbs,
		logger: log,
	}
}

func RunConsumer(ctx context.Context, settings *config.Settings, logger *zerolog.Logger, dbs db.Store) error {
	consumer := NewConsumer(dbs, logger)

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

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.EthereumAddress.EQ(addr.Bytes()),
		qm.Load(models.AftermarketDeviceRels.VehicleToken),
	).One(ctx, c.DBS.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed querying for device: %w", err)
	}

	ud := ad.R.VehicleToken
	if ud == nil {
		return nil
	}

	var md services.UserDeviceMetadata
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

func ExtractProtocol(data []byte) (*string, error) {
	partialData := new(struct {
		Protocol *string `json:"protocol"`
	})

	if err := json.Unmarshal(data, partialData); err != nil {
		return nil, fmt.Errorf("failed parsing data field: %w", err)
	}

	return partialData.Protocol, nil
}
