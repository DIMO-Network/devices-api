package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source device_definition_registrar.go -destination mocks/device_definition_registrar_mock.go

// DeviceDefinitionRegistrar
type DeviceDefinitionRegistrar interface {
	Register(d DeviceDefinitionDTO) error
}

type DeviceDefinitionDTO struct {
	UserDeviceID       string
	DeviceDefinitionID string
	Make               string
	Model              string
	Year               int
	IntegrationID      string
	Region             string
	MakeSlug           string
	ModelSlug          string
}

type DeviceDefinitionIDEventData struct {
	UserDeviceID       string `json:"userDeviceId"`
	DeviceDefinitionID string `json:"deviceDefinitionId"`
}

type DeviceDefinitionMetadataEventData struct {
	Make      string `json:"make"`
	Model     string `json:"model"`
	Year      int    `json:"year"`
	Region    string `json:"region"`
	MakeSlug  string `json:"makeSlug"`
	ModelSlug string `json:"modelSlug"`
}

type deviceDefinitionRegistrar struct {
	producer sarama.SyncProducer
	settings *config.Settings
}

func NewDeviceDefinitionRegistrar(producer sarama.SyncProducer, settings *config.Settings) DeviceDefinitionRegistrar {
	return &deviceDefinitionRegistrar{
		producer: producer,
		settings: settings,
	}
}

func (s *deviceDefinitionRegistrar) emitDeviceDefinitionIDEvent(d DeviceDefinitionDTO) error {
	topic := s.settings.DeviceDefinitionTopic
	eventType := "zone.dimo.device.definition"

	value := shared.CloudEvent[DeviceDefinitionIDEventData]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + d.IntegrationID,
		Subject:     d.UserDeviceID,
		SpecVersion: "1.0",
		Time:        time.Now(),
		Type:        eventType,
		Data: DeviceDefinitionIDEventData{
			UserDeviceID:       d.UserDeviceID,
			DeviceDefinitionID: d.DeviceDefinitionID,
		},
	}

	return s.emitKafka(topic, d.UserDeviceID, value)
}

func (s *deviceDefinitionRegistrar) emitDeviceDefinitionMetadataEvent(d DeviceDefinitionDTO) error {
	topic := s.settings.DeviceDefinitionMetadataTopic
	eventType := "zone.dimo.device.definition.metadata"

	value := shared.CloudEvent[DeviceDefinitionMetadataEventData]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + d.IntegrationID,
		Subject:     d.DeviceDefinitionID,
		SpecVersion: "1.0",
		Time:        time.Now(),
		Type:        eventType,
		Data: DeviceDefinitionMetadataEventData{
			Make:      d.Make,
			Model:     d.Model,
			Year:      d.Year,
			Region:    d.Region,
			MakeSlug:  d.MakeSlug,
			ModelSlug: d.ModelSlug,
		},
	}

	return s.emitKafka(topic, d.DeviceDefinitionID, value)
}

func (s *deviceDefinitionRegistrar) emitKafka(topic, key string, value any) error {
	valueb, err := json.Marshal(value)
	if err != nil {
		return err
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(valueb),
	}

	_, _, err = s.producer.SendMessage(message)
	return err
}

func (s *deviceDefinitionRegistrar) Register(d DeviceDefinitionDTO) error {
	err := s.emitDeviceDefinitionIDEvent(d)
	if err != nil {
		return fmt.Errorf("failed to emit device definition id event: %w", err)
	}

	err = s.emitDeviceDefinitionMetadataEvent(d)
	if err != nil {
		return fmt.Errorf("failed to emit device definition metadata event: %w", err)
	}

	return nil
}
