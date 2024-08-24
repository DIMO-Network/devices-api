package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/sdtask"
	"github.com/IBM/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source tesla_task_service.go -destination mocks/tesla_task_service_mock.go

type TeslaTaskService interface {
	StartPoll(udai *models.UserDeviceAPIIntegration, sd *models.SyntheticDevice) error
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

type TeslaTask struct {
	TaskID             string           `json:"taskId"`
	UserDeviceID       string           `json:"userDeviceId"`
	IntegrationID      string           `json:"integrationId"`
	Identifiers        TeslaIdentifiers `json:"identifiers"`
	OnlineIdleLastPoll bool             `json:"onlineIdleLastPoll"`
}

func (t *teslaTaskService) StartPoll(udai *models.UserDeviceAPIIntegration, sd *models.SyntheticDevice) error {
	var meta UserDeviceAPIIntegrationsMetadata
	err := udai.Metadata.Unmarshal(&meta)
	if err != nil {
		return fmt.Errorf("couldn't unmarshal metadata: %w", err)
	}

	id, err := strconv.Atoi(udai.ExternalID.String)
	if err != nil {
		return fmt.Errorf("couldn't parse Tesla id %q: %w", udai.ExternalID.String, err)
	}

	tt := shared.CloudEvent[TeslaTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.poll.scheduled",
		Data: TeslaTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: TeslaIdentifiers{
				ID:        id,
				VehicleID: meta.TeslaVehicleID,
			},
			OnlineIdleLastPoll: false,
		},
	}

	tokenID, _ := sd.TokenID.Int64()
	integrationTokenID, _ := sd.IntegrationTokenID.Int64()
	vehicleTokenID, _ := sd.VehicleTokenID.Int64()

	tc := shared.CloudEvent[sdtask.CredentialData]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.poll.credential.v2",
		Data: sdtask.CredentialData{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			AccessToken:   udai.AccessToken.String,
			Expiry:        udai.AccessExpiresAt.Time,
			RefreshToken:  udai.RefreshToken.String,
			Version:       meta.TeslaAPIVersion,
			SyntheticDevice: &sdtask.SyntheticDevice{
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

func (t *teslaTaskService) StopPoll(udai *models.UserDeviceAPIIntegration) error {
	var taskKey string
	if udai.TaskID.Valid {
		taskKey = udai.TaskID.String
	} else {
		taskKey = fmt.Sprintf("device/%s/integration/%s", udai.UserDeviceID, udai.IntegrationID)
	}

	tt := shared.CloudEvent[any]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.poll.stop",
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
