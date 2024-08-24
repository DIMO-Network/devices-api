package services

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/IBM/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source event_service.go -destination mocks/event_service_mock.go -package mock_services

//type Event struct {
//	Type    string
//	Subject string
//	Source  string
//	Data    any
//}

type EventService interface {
	Emit(event *shared.CloudEvent[any]) error
}

type eventService struct {
	Settings *config.Settings
	Logger   *zerolog.Logger
	Producer sarama.SyncProducer
}

func NewEventService(logger *zerolog.Logger, settings *config.Settings, producer sarama.SyncProducer) EventService {
	return &eventService{
		Settings: settings,
		Logger:   logger,
		Producer: producer,
	}
}

func (e *eventService) Emit(event *shared.CloudEvent[any]) error {
	msgBytes, err := json.Marshal(shared.CloudEvent[any]{
		ID:          ksuid.New().String(),
		Source:      event.Source,
		SpecVersion: "1.0",
		Subject:     event.Subject,
		Time:        time.Now(),
		Type:        event.Type,
		Data:        event.Data,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal CloudEvent: %w", err)
	}
	msg := &sarama.ProducerMessage{
		Topic: e.Settings.EventsTopic,
		Key:   sarama.StringEncoder(event.Subject),
		Value: sarama.ByteEncoder(msgBytes),
	}
	_, _, err = e.Producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to produce CloudEvent to Kafka: %w", err)
	}
	return nil
}

type UserDeviceEvent struct {
	Timestamp time.Time             `json:"timestamp"`
	UserID    string                `json:"userId"`
	Device    UserDeviceEventDevice `json:"device"`
}

type UserDeviceEventDevice struct {
	ID                 string `json:"id"`
	DeviceDefinitionID string `json:"device_definition_id"`
	Make               string `json:"make"`
	Model              string `json:"model"`
	Year               int    `json:"year"`
	VIN                string `json:"vin"`
}

type UserDeviceEventIntegration struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Style  string `json:"style"`
	Vendor string `json:"vendor"`
}

type UserDeviceIntegrationEvent struct {
	Timestamp   time.Time                  `json:"timestamp"`
	UserID      string                     `json:"userId"`
	Device      UserDeviceEventDevice      `json:"device"`
	Integration UserDeviceEventIntegration `json:"integration"`
}

type UserDeviceEventNFT struct {
	TokenID *big.Int       `json:"tokenId"`
	Owner   common.Address `json:"address"`
	TxHash  common.Hash    `json:"txHash"`
}

type UserDeviceMintEvent struct {
	Timestamp time.Time             `json:"timestamp"`
	UserID    string                `json:"userId"`
	Device    UserDeviceEventDevice `json:"device"`
	NFT       UserDeviceEventNFT    `json:"nft"`
}
