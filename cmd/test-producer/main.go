package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

func main() {
	integrationID := ""
	vehicleID := ""

	//ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api-test-producer").
		Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load settings")
	}

	// todo: learn better way to do flags
	if len(os.Args) != 3 {
		logger.Fatal().Msg("must pass int integrationId and vehicleId as parameters, eg. test-producer <iid> <vid>")
	}
	integrationID = os.Args[1]
	vehicleID = os.Args[2]

	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_6_0_0
	clusterConfig.Producer.Return.Successes = true
	clusterConfig.Producer.Return.Errors = true

	syncProducer, err := sarama.NewSyncProducer(strings.Split(settings.KafkaBrokers, ","), clusterConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not start test sync producer")
	}
	msgID := ksuid.New().String()
	testMessage := buildTestMessage(msgID, integrationID, vehicleID)
	bytes, err := json.Marshal(testMessage)
	if err != nil {
		logger.Fatal().Err(err).Msg("error marshalling test event to json")
	}
	message := &sarama.ProducerMessage{
		Topic: settings.DeviceStatusTopic,
		Value: sarama.StringEncoder(bytes),
		Key:   sarama.StringEncoder(msgID),
	}

	partition, offset, err := syncProducer.SendMessage(message)
	if err != nil {
		logger.Err(err).Msg("could not produce message to topic")
	}
	logger.Info().Msgf("succesfully published message on topic. partition: %d offset: %d", partition, offset)
}

func buildTestMessage(id, sourceIntegrationID, subjectVehicleID string) services.DeviceStatusEvent {
	j := json.RawMessage{}
	_ = j.UnmarshalJSON(testVehicleData())
	e := services.DeviceStatusEvent{
		ID:          id,
		Source:      sourceIntegrationID,
		Specversion: "1.0",
		Subject:     subjectVehicleID,
		Time:        time.Now().UTC(),
		Type:        "zone.dimo.device.status.update",
		Data:        j,
	}
	return e
}

func testVehicleData() []byte {
	d := `{
        "vin": "0SCWEBHOOKTEST000",
        "make": "MOCK",
        "model": "Webhook Test",
        "year": 2020,
        "batteryCapacity": 57.86,
        "charging": false,
        "errors": [],
        "latitude": 39.0272216796875,
        "longitude": -105.93428802490234,
        "odometer": 109931.921875,
        "oil": 0.6700000166893005,
        "range": 473.23,
        "soc": 0.49,
        "tires": {
            "backLeft": 214.03509521484375,
            "backRight": 197.893798828125,
            "frontLeft": 187.8616943359375,
            "frontRight": 197.83590698242188
        }
    }`
	return []byte(d)
}
