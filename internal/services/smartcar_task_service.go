package services

import (
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source smartcar_task_service.go -destination mocks/smartcar_task_service_mock.go

type SmartcarTaskService interface {
	StartPoll(udai *models.UserDeviceAPIIntegration, sd *models.SyntheticDevice) error
	StopPoll(udai *models.UserDeviceAPIIntegration) error
	Refresh(udai *models.UserDeviceAPIIntegration) error
	UnlockDoors(udai *models.UserDeviceAPIIntegration) (string, error)
	LockDoors(udai *models.UserDeviceAPIIntegration) (string, error)
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

func (t *smartcarTaskService) StartPoll(udai *models.UserDeviceAPIIntegration, sd *models.SyntheticDevice) error {
	m := new(UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(m); err != nil {
		return err
	}

	tt := shared.CloudEvent[SmartcarTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.smartcar.poll.scheduled",
		Data: SmartcarTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: SmartcarIdentifiers{
				ID: udai.ExternalID.String,
			},
			Paths: m.SmartcarEndpoints,
		},
	}

	tokenID, _ := sd.TokenID.Int64()
	integrationTokenID, _ := sd.IntegrationTokenID.Int64()
	vehicleTokenID, _ := sd.VehicleTokenID.Int64()

	tc := shared.CloudEvent[SyntheticTaskCredentialData]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.smartcar.poll.credential",
		Data: SyntheticTaskCredentialData{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			AccessToken:   udai.AccessToken.String,
			Expiry:        udai.AccessExpiresAt.Time,
			RefreshToken:  udai.RefreshToken.String,
			SyntheticDevice: &CredsSynthetic{
				TokenID:            int(tokenID),
				Address:            common.BytesToAddress(sd.WalletAddress),
				IntegrationTokenID: int(integrationTokenID),
				WalletChildNumber:  sd.WalletChildNumber,
				VehicleTokenID:     int(vehicleTokenID),
			},
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	tcb, err := json.Marshal(tc)
	if err != nil {
		return err
	}

	err = t.Producer.SendMessages(
		[]*sarama.ProducerMessage{
			{
				Topic: t.Settings.TaskRunNowTopic,
				Key:   sarama.StringEncoder(udai.TaskID.String),
				Value: sarama.ByteEncoder(ttb),
			},
			{
				Topic: t.Settings.TaskCredentialTopic,
				Key:   sarama.StringEncoder(udai.TaskID.String),
				Value: sarama.ByteEncoder(tcb),
			},
		},
	)

	return err
}

func (t *smartcarTaskService) Refresh(udai *models.UserDeviceAPIIntegration) error {
	m := new(UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(m); err != nil {
		return err
	}

	tt := shared.CloudEvent[SmartcarTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.smartcar.poll.refresh",
		Data: SmartcarTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: SmartcarIdentifiers{
				ID: udai.ExternalID.String,
			},
			Paths: m.SmartcarEndpoints,
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	_, _, err = t.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: t.Settings.TaskRunNowTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return err
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

type SmartcarDoorTask struct {
	TaskID        string              `json:"taskId"`
	SubTaskID     string              `json:"subTaskId"`
	UserDeviceID  string              `json:"userDeviceId"`
	IntegrationID string              `json:"integrationId"`
	Identifiers   SmartcarIdentifiers `json:"identifiers"`
}

func (t *smartcarTaskService) UnlockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	tt := shared.CloudEvent[SmartcarDoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.smartcar.doors.unlock",
		Data: SmartcarDoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: SmartcarIdentifiers{
				ID: udai.ExternalID.String,
			},
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return "", err
	}

	_, _, err = t.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: t.Settings.TaskRunNowTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return tt.Data.SubTaskID, err
}

func (t *smartcarTaskService) LockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	tt := shared.CloudEvent[SmartcarDoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.smartcar.doors.lock",
		Data: SmartcarDoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: SmartcarIdentifiers{
				ID: udai.ExternalID.String,
			},
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return "", err
	}

	_, _, err = t.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: t.Settings.TaskRunNowTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return tt.Data.SubTaskID, err
}
