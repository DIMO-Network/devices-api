package services

import (
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/segmentio/ksuid"

	"encoding/json"
	"fmt"
	"time"
)

const ingestSmartcarRegistrationTopic = "table.device.integration.smartcar"
const smartcarRegistrationEventType = "zone.dimo.device.integration.smartcar.register"

const ingestAutoPiRegistrationTopic = "table.device.integration.autopi"
const autoPiRegistrationEventType = "zone.dimo.device.integration.autopi.register"

//go:generate mockgen -source ingest_registrar.go -destination mocks/ingest_registrar_mock.go
type IngestRegistrar interface {
	Register(externalID, userDeviceID, integrationID string) error
	Deregister(externalID, userDeviceID, integrationID string) error
}

func NewIngestRegistrar(integrationType IntegrationTypeEnum, producer sarama.SyncProducer) IngestRegistrar {
	eventType := ""
	ingestTopic := ""
	switch integrationType {
	case Smartcar:
		eventType = smartcarRegistrationEventType
		ingestTopic = ingestSmartcarRegistrationTopic
	case AutoPi:
		eventType = autoPiRegistrationEventType
		ingestTopic = ingestAutoPiRegistrationTopic
	}

	return &ingestRegistrar{eventType: eventType, ingestTopic: ingestTopic, Producer: producer}
}

// IngestRegistrar is an interface to the table.device.integration.smartcar/autopi topic, a
// compacted Kafka topic keyed by Smartcar vehicle ID or autoPi Device ID. The ingest service needs to match
// these IDs to our device IDs.
type ingestRegistrar struct {
	eventType   string
	ingestTopic string
	Producer    sarama.SyncProducer
}

type deviceIDLink struct {
	DeviceID   string `json:"deviceId"`
	ExternalID string `json:"externalId"`
}

func (s *ingestRegistrar) Register(externalID, userDeviceID, integrationID string) error {
	value := shared.CloudEvent[deviceIDLink]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + integrationID,
		Subject:     userDeviceID,
		SpecVersion: "1.0",
		Time:        time.Now(),
		Type:        s.eventType,
		Data: deviceIDLink{
			DeviceID:   userDeviceID,
			ExternalID: externalID,
		},
	}
	valueb, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize JSON body: %w", err)
	}
	message := &sarama.ProducerMessage{
		Topic: s.ingestTopic,
		Key:   sarama.StringEncoder(externalID),
		Value: sarama.ByteEncoder(valueb),
	}
	_, _, err = s.Producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed sending to Kafka: %w", err)
	}

	return nil
}

func (s *ingestRegistrar) Deregister(smartcarID, userDeviceID, integrationID string) error {
	message := &sarama.ProducerMessage{
		Topic: s.ingestTopic,
		Key:   sarama.StringEncoder(smartcarID),
		Value: nil, // Delete from compacted topic.
	}
	_, _, err := s.Producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed sending to Kafka: %w", err)
	}

	return nil
}

type IntegrationTypeEnum string

const (
	Smartcar IntegrationTypeEnum = "Smartcar"
	AutoPi   IntegrationTypeEnum = "AutoPi"
)

func (r IntegrationTypeEnum) String() string {
	return string(r)
}
