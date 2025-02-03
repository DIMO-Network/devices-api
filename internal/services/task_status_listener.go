package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services/cio"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	teslaStatusEventType    = "zone.dimo.task.tesla.poll.status.update"
	smartcarStatusEventType = "zone.dimo.task.smartcar.poll.status.update"
	commandStatusEventType  = "zone.dimo.task.device.integration.command.status"
)

type TaskStatusListener struct {
	db           func() *db.ReaderWriter
	log          *zerolog.Logger
	DeviceDefSvc DeviceDefinitionService
	prod         sarama.SyncProducer
	cioSvc       *cio.Service
	settings     *config.Settings
}

type TaskStatusData struct {
	TaskID        string `json:"taskId"`
	SubTaskID     string `json:"subTaskId"`
	UserDeviceID  string `json:"userDeviceId"`
	IntegrationID string `json:"integrationId"`
	Status        string `json:"status"`
}

func NewTaskStatusListener(db func() *db.ReaderWriter, log *zerolog.Logger, ddSvc DeviceDefinitionService, prod sarama.SyncProducer, cioSvc *cio.Service, settings *config.Settings) *TaskStatusListener {
	return &TaskStatusListener{db: db, log: log, DeviceDefSvc: ddSvc, prod: prod, cioSvc: cioSvc, settings: settings}
}

func (i *TaskStatusListener) ProcessTaskUpdates(messages <-chan *message.Message) {
	for msg := range messages {
		err := i.processMessage(msg)
		if err != nil {
			i.log.Err(err).Msg("error processing task status message")
		}
	}
}

func (i *TaskStatusListener) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	event := new(shared.CloudEvent[TaskStatusData])
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		return errors.Wrap(err, "error parsing task status payload")
	}

	return i.processEvent(event)
}

func (i *TaskStatusListener) processEvent(event *shared.CloudEvent[TaskStatusData]) error {
	switch event.Type {
	case smartcarStatusEventType:
		return i.processSmartcarPollStatusEvent(event)
	case teslaStatusEventType:
		return i.processTeslaPollStatusEvent(event)
	case commandStatusEventType:
		return i.processCommandStatusEvent(event)
	default:
		return fmt.Errorf("unexpected event type %s", event.Type)
	}
}

func (i *TaskStatusListener) processSmartcarPollStatusEvent(event *shared.CloudEvent[TaskStatusData]) error {
	var (
		ctx          = context.Background()
		userDeviceID = event.Subject
	)

	// Should we use data.integrationId instead?
	if !strings.HasPrefix(event.Source, sourcePrefix) {
		return fmt.Errorf("unexpected event source format: %s", event.Source)
	}
	integrationID := strings.TrimPrefix(event.Source, sourcePrefix)

	// Just one case for now.
	if event.Data.Status != models.UserDeviceAPIIntegrationStatusAuthenticationFailure {
		return fmt.Errorf("unexpected task status %s", event.Data.Status)
	}

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).One(ctx, i.db().Writer)
	if err != nil {
		return fmt.Errorf("couldn't find device integration for device %s and integration %s: %w", userDeviceID, integrationID, err)
	}

	i.log.Info().Str("userDeviceId", userDeviceID).Msg("Setting Smartcar integration to failed because credentials have changed.")

	if udai.TaskID.Valid && udai.TaskID.String == event.Data.TaskID {
		// Maybe you've restarted the task with new credentials already.
		// TODO: Delete credentials entry?
		udai.TaskID = null.String{}

		_, _, err := i.prod.SendMessage(&sarama.ProducerMessage{
			Topic: i.settings.TaskCredentialTopic,
			Key:   sarama.StringEncoder(event.Data.TaskID),
			Value: nil,
		})
		if err != nil {
			i.log.Err(err).Msg("Failed to null out credential message for failed job.")
		}
	}

	udai.Status = models.UserDeviceAPIIntegrationStatusAuthenticationFailure
	if _, err = udai.Update(context.Background(), i.db().Writer, boil.Infer()); err != nil {
		i.log.Err(err).Str("userDeviceID", userDeviceID).Str("integrationID", integrationID).Msg("failed up update user device api integration with failure status")
	}

	if err := i.cioSvc.SoftwareDisconnectionEvent(ctx, udai); err != nil {
		i.log.Err(err).Str("userDeviceID", userDeviceID).Str("integrationID", integrationID).Msg("failed to send CIO software disconnection event")
		return err
	}

	return nil
}

func (i *TaskStatusListener) processTeslaPollStatusEvent(event *shared.CloudEvent[TaskStatusData]) error {
	var (
		ctx          = context.Background()
		userDeviceID = event.Subject
	)

	// Should we use data.integrationId instead?
	if !strings.HasPrefix(event.Source, sourcePrefix) {
		return fmt.Errorf("unexpected event source format: %s", event.Source)
	}
	integrationID := strings.TrimPrefix(event.Source, sourcePrefix)

	// Just one case for now.
	if event.Data.Status != models.UserDeviceAPIIntegrationStatusAuthenticationFailure {
		return fmt.Errorf("unexpected task status %s", event.Data.Status)
	}

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).One(ctx, i.db().Writer)
	if err != nil {
		return fmt.Errorf("couldn't find device integration for device %s and integration %s: %w", userDeviceID, integrationID, err)
	}

	i.log.Info().Str("userDeviceId", userDeviceID).Msg("Setting Tesla integration to failed because credentials have changed.")

	if udai.TaskID.Valid && udai.TaskID.String == event.Data.TaskID {
		// Maybe you've restarted the task with new credentials already.
		// TODO: Delete credentials entry?
		udai.TaskID = null.String{}

		_, _, err := i.prod.SendMessage(&sarama.ProducerMessage{
			Topic: i.settings.TaskCredentialTopic,
			Key:   sarama.StringEncoder(event.Data.TaskID),
			Value: nil,
		})
		if err != nil {
			i.log.Err(err).Msg("Failed to null out credential message for failed job.")
		}
	}
	udai.Status = models.UserDeviceAPIIntegrationStatusAuthenticationFailure
	if _, err = udai.Update(context.Background(), i.db().Writer, boil.Infer()); err != nil {
		i.log.Err(err).Str("userDeviceID", userDeviceID).Str("integrationID", integrationID).Msg("failed up update user device api integration with failure status")
	}

	if err := i.cioSvc.SoftwareDisconnectionEvent(ctx, udai); err != nil {
		i.log.Err(err).Str("userDeviceID", userDeviceID).Str("integrationID", integrationID).Msg("failed to send CIO software disconnection event")
		return err
	}

	return nil
}

func (i *TaskStatusListener) processCommandStatusEvent(event *shared.CloudEvent[TaskStatusData]) error {
	dcr, err := models.FindDeviceCommandRequest(context.Background(), i.db().Writer, event.Data.SubTaskID)
	if err != nil {
		return fmt.Errorf("failed to find command request: %w", err)
	}

	if event.Data.Status != models.DeviceCommandRequestStatusComplete && event.Data.Status != models.DeviceCommandRequestStatusFailed {
		return fmt.Errorf("unexpected command status %q", dcr.Status)
	}

	dcr.Status = event.Data.Status
	_, err = dcr.Update(context.Background(), i.db().Writer, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to update command request: %w", err)
	}

	i.log.Info().
		Str("subTaskId", event.Data.SubTaskID).
		Str("command", dcr.Command).
		Str("userDeviceId", dcr.UserDeviceID).
		Str("integrationId", dcr.IntegrationID).
		Str("status", dcr.Status).
		Msg("Updated command request status.")

	return nil
}
