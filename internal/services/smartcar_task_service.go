package services

import (
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/IBM/sarama"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source smartcar_task_service.go -destination mocks/smartcar_task_service_mock.go

type SmartcarTaskService interface {
	StopPoll(udai *models.UserDeviceAPIIntegration) error
}

func NewSmartcarTaskService(settings *config.Settings, producer sarama.SyncProducer) SmartcarTaskService {
	return &smartcarTaskService{
		Producer: producer,
		Settings: settings,
	}
}

type smartcarTaskService struct {
	Producer sarama.SyncProducer
	Settings *config.Settings
}

type SmartcarIdentifiers struct {
	ID string `json:"id"`
}

type SmartcarTask struct {
	TaskID        string              `json:"taskId"`
	UserDeviceID  string              `json:"userDeviceId"`
	IntegrationID string              `json:"integrationId"`
	Identifiers   SmartcarIdentifiers `json:"identifiers"`
	Paths         []string            `json:"paths"`
}

func (t *smartcarTaskService) StopPoll(udai *models.UserDeviceAPIIntegration) error {
	var taskKey = udai.TaskID.String

	tt := shared.CloudEvent[any]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.smartcar.poll.stop",
		Data: struct {
			TaskID        string `json:"taskId"`
			UserDeviceID  string `json:"userDeviceId"`
			IntegrationID string `json:"integrationId"`
		}{
			TaskID:        taskKey,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	err = t.Producer.SendMessages(
		[]*sarama.ProducerMessage{
			{
				Topic: t.Settings.TaskStopTopic,
				Key:   sarama.StringEncoder(taskKey),
				Value: sarama.ByteEncoder(ttb),
			},
			{
				Topic: t.Settings.TaskCredentialTopic,
				Key:   sarama.StringEncoder(taskKey),
				Value: nil,
			},
		},
	)

	return err
}
