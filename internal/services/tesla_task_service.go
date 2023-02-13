package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source tesla_task_service.go -destination mocks/tesla_task_service_mock.go

type TeslaTaskService interface {
	StartPoll(vehicle *TeslaVehicle, udai *models.UserDeviceAPIIntegration) error
	StopPoll(udai *models.UserDeviceAPIIntegration) error
	UnlockDoors(udai *models.UserDeviceAPIIntegration) (string, error)
	LockDoors(udai *models.UserDeviceAPIIntegration) (string, error)
	OpenTrunk(udai *models.UserDeviceAPIIntegration) (string, error)
	OpenFrunk(udai *models.UserDeviceAPIIntegration) (string, error)
}

func NewTeslaTaskService(settings *config.Settings, producer sarama.SyncProducer) TeslaTaskService {
	return &teslaTaskService{
		Producer: producer,
		Settings: settings,
	}
}

// Make sure we satisfy the interface.
var _ TeslaTaskService = &teslaTaskService{}

type teslaTaskService struct {
	Producer sarama.SyncProducer
	Settings *config.Settings
}

type TeslaIdentifiers struct {
	ID        int `json:"id"`
	VehicleID int `json:"vehicleId"`
}

type TeslaCredentialsV2 struct {
	TaskID        string    `json:"taskId"`
	UserDeviceID  string    `json:"userDeviceId"`
	IntegrationID string    `json:"integrationId"`
	AccessToken   string    `json:"accessToken"`
	Expiry        time.Time `json:"expiry"`
	RefreshToken  string    `json:"refreshToken"`
}

type TeslaTask struct {
	TaskID             string           `json:"taskId"`
	UserDeviceID       string           `json:"userDeviceId"`
	IntegrationID      string           `json:"integrationId"`
	Identifiers        TeslaIdentifiers `json:"identifiers"`
	OnlineIdleLastPoll bool             `json:"onlineIdleLastPoll"`
}

// CloudEventHeaders contains the fields common to all CloudEvent messages.
type CloudEventHeaders struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	SpecVersion string    `json:"specversion"`
	Subject     string    `json:"subject"`
	Time        time.Time `json:"time"`
	Type        string    `json:"type"`
}

type TeslaTaskCloudEvent struct {
	CloudEventHeaders
	Data TeslaTask `json:"data"`
}

type TeslaCredentialsCloudEventV2 struct {
	CloudEventHeaders
	Data TeslaCredentialsV2 `json:"data"`
}

func (t *teslaTaskService) StartPoll(vehicle *TeslaVehicle, udai *models.UserDeviceAPIIntegration) error {
	tt := TeslaTaskCloudEvent{
		CloudEventHeaders: CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.tesla.poll.scheduled",
		},
		Data: TeslaTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: TeslaIdentifiers{
				ID:        vehicle.ID,
				VehicleID: vehicle.VehicleID,
			},
			OnlineIdleLastPoll: false,
		},
	}

	tc := TeslaCredentialsCloudEventV2{
		CloudEventHeaders: CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.tesla.poll.credential.v2",
		},
		Data: TeslaCredentialsV2{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			AccessToken:   udai.AccessToken.String,
			Expiry:        udai.AccessExpiresAt.Time,
			RefreshToken:  udai.RefreshToken.String,
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

func (t *teslaTaskService) StopPoll(udai *models.UserDeviceAPIIntegration) error {
	var taskKey string
	if udai.TaskID.Valid {
		taskKey = udai.TaskID.String
	} else {
		taskKey = fmt.Sprintf("device/%s/integration/%s", udai.UserDeviceID, udai.IntegrationID)
	}

	tt := struct {
		CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.tesla.poll.stop",
		},
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

type TeslaDoorTask struct {
	TaskID        string           `json:"taskId"`
	SubTaskID     string           `json:"subTaskId"`
	UserDeviceID  string           `json:"userDeviceId"`
	IntegrationID string           `json:"integrationId"`
	Identifiers   TeslaIdentifiers `json:"identifiers"` // Don't actually need vehicleId.
	ChargeLimit   *float64         `json:"chargeLimit,omitempty"`
}

func (t *teslaTaskService) UnlockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	id, err := strconv.Atoi(udai.ExternalID.String)
	if err != nil {
		return "", err
	}

	tt := shared.CloudEvent[TeslaDoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.doors.unlock",
		Data: TeslaDoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: TeslaIdentifiers{
				ID: id,
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

func (t *teslaTaskService) LockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	id, err := strconv.Atoi(udai.ExternalID.String)
	if err != nil {
		return "", err
	}

	tt := shared.CloudEvent[TeslaDoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.doors.lock",
		Data: TeslaDoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: TeslaIdentifiers{
				ID: id,
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

func (t *teslaTaskService) OpenTrunk(udai *models.UserDeviceAPIIntegration) (string, error) {
	id, err := strconv.Atoi(udai.ExternalID.String)
	if err != nil {
		return "", err
	}

	tt := shared.CloudEvent[TeslaDoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.trunk.open",
		Data: TeslaDoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: TeslaIdentifiers{
				ID: id,
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

func (t *teslaTaskService) OpenFrunk(udai *models.UserDeviceAPIIntegration) (string, error) {
	id, err := strconv.Atoi(udai.ExternalID.String)
	if err != nil {
		return "", err
	}

	tt := shared.CloudEvent[TeslaDoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.frunk.open",
		Data: TeslaDoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: TeslaIdentifiers{
				ID: id,
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
