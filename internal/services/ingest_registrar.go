package services

import (
	"math/big"

	"github.com/DIMO-Network/shared"
	"github.com/IBM/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"

	"encoding/json"
	"fmt"
	"time"
)

const ingestAutoPiRegistrationTopic = "table.device.integration.autopi"
const autoPiRegistrationEventType = "zone.dimo.device.integration.autopi.register"

const aftermarketDeviceIntegrationTopic = "table.aftermarket.device.integration"
const aftermarketDeviceIntegrationEventType = "zone.dimo.aftermarket.device.integration"

//go:generate mockgen -source ingest_registrar.go -destination mocks/ingest_registrar_mock.go
type IngestRegistrar interface {
	Register(externalID, userDeviceID, integrationID string) error
	Deregister(externalID, userDeviceID, integrationID string) error

	Register2(data *AftermarketDeviceVehicleMapping) error
	Deregister2(addr common.Address) error
}

func NewIngestRegistrar(producer sarama.SyncProducer) IngestRegistrar {
	return &ingestRegistrar{Producer: producer}
}

// IngestRegistrar is an interface to the table.device.integration.smartcar/autopi topic, a
// compacted Kafka topic keyed by Smartcar vehicle ID or autoPi Device ID. The ingest service needs to match
// these IDs to our device IDs.
type ingestRegistrar struct {
	Producer sarama.SyncProducer
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
		Type:        autoPiRegistrationEventType,
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
		Topic: ingestAutoPiRegistrationTopic,
		Key:   sarama.StringEncoder(externalID),
		Value: sarama.ByteEncoder(valueb),
	}
	_, _, err = s.Producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed sending to Kafka: %w", err)
	}

	return nil
}

func (s *ingestRegistrar) Deregister(externalID, _, _ string) error {
	message := &sarama.ProducerMessage{
		Topic: ingestAutoPiRegistrationTopic,
		Key:   sarama.StringEncoder(externalID),
		Value: nil, // Delete from compacted topic.
	}
	_, _, err := s.Producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed sending to Kafka: %w", err)
	}

	return nil
}

func (s *ingestRegistrar) Register2(data *AftermarketDeviceVehicleMapping) error {
	value := shared.CloudEvent[AftermarketDeviceVehicleMapping]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + data.AftermarketDevice.IntegrationID,
		Subject:     data.Vehicle.UserDeviceID,
		SpecVersion: "1.0",
		Time:        time.Now(),
		Type:        aftermarketDeviceIntegrationEventType,
		Data:        *data,
	}
	valueb, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize JSON body: %w", err)
	}
	message := &sarama.ProducerMessage{
		Topic: aftermarketDeviceIntegrationTopic,
		Key:   sarama.StringEncoder(data.AftermarketDevice.Address.Hex()), // Must be checksummed
		Value: sarama.ByteEncoder(valueb),
	}
	_, _, err = s.Producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed sending to Kafka: %w", err)
	}

	return nil
}

func (s *ingestRegistrar) Deregister2(addr common.Address) error {
	message := &sarama.ProducerMessage{
		Topic: aftermarketDeviceIntegrationTopic,
		Key:   sarama.StringEncoder(addr.Hex()),
		Value: nil, // Delete from compacted topic.
	}
	_, _, err := s.Producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed sending to Kafka: %w", err)
	}

	return nil
}

type AftermarketDeviceVehicleMapping struct {
	AftermarketDevice AftermarketDeviceVehicleMappingAftermarketDevice `json:"aftermarketDevice"`
	Vehicle           AftermarketDeviceVehicleMappingVehicle           `json:"vehicle"`
}

type AftermarketDeviceVehicleMappingAftermarketDevice struct {
	Address       common.Address `json:"address"`
	Token         *big.Int       `json:"token"`
	Serial        string         `json:"serial"`
	IntegrationID string         `json:"integrationId"`
}

type AftermarketDeviceVehicleMappingVehicle struct {
	Token        *big.Int `json:"token"`
	UserDeviceID string   `json:"userDeviceId"`
}
