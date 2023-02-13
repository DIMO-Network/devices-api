package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gofiber/fiber/v2"
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
	cio          CIOClient
}

type CIOClient interface {
	Track(customerID string, eventName string, data map[string]interface{}) error
}

type TaskStatusData struct {
	TaskID        string `json:"taskId"`
	SubTaskID     string `json:"subTaskId"`
	UserDeviceID  string `json:"userDeviceId"`
	IntegrationID string `json:"integrationId"`
	Status        string `json:"status"`
}

func NewTaskStatusListener(db func() *db.ReaderWriter, log *zerolog.Logger, cio CIOClient, ddSvc DeviceDefinitionService) *TaskStatusListener {
	return &TaskStatusListener{db: db, log: log, cio: cio, DeviceDefSvc: ddSvc}
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
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, i.db().Writer)
	if err != nil {
		return fmt.Errorf("couldn't find device integration for device %s and integration %s: %w", userDeviceID, integrationID, err)
	}

	userDevice := udai.R.UserDevice

	i.log.Info().Str("userDeviceId", userDeviceID).Msg("Setting Smartcar integration to failed because credentials have changed.")

	if udai.TaskID.Valid && udai.TaskID.String == event.Data.TaskID {
		// Maybe you've restarted the task with new credentials already.
		// TODO: Delete credentials entry?
		udai.TaskID = null.String{}
	}
	udai.Status = models.UserDeviceAPIIntegrationStatusAuthenticationFailure
	if _, err := udai.Update(context.Background(), i.db().Writer, boil.Infer()); err != nil {
		return err
	}

	deviceDefinitionResponse, err := i.DeviceDefSvc.GetDeviceDefinitionsByIDs(ctx, []string{userDevice.DeviceDefinitionID})

	if err != nil {
		return helpers.GrpcErrorToFiber(err, fmt.Sprintf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID))
	}

	if len(deviceDefinitionResponse) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "could not find device definition id: "+userDevice.DeviceDefinitionID)
	}

	dd := deviceDefinitionResponse[0]
	data := map[string]interface{}{
		"deviceId":     userDeviceID,
		"make_name":    dd.Make.Name,
		"model_name":   dd.Type.Model,
		"model_year":   dd.Type.Year,
		"country_code": userDevice.CountryCode.String,
	}

	if err := i.cio.Track(userDevice.UserID, "smartcar.Reauth.Required", data); err != nil {
		i.log.Err(err).Str("userId", userDevice.UserID).Str("userDeviceId", userDeviceID).Msg("Failed to emit reauthentication Customer.io event.")
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
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, i.db().Writer)
	if err != nil {
		return fmt.Errorf("couldn't find device integration for device %s and integration %s: %w", userDeviceID, integrationID, err)
	}

	userDevice := udai.R.UserDevice

	i.log.Info().Str("userDeviceId", userDeviceID).Msg("Setting Tesla integration to failed because credentials have changed.")

	if udai.TaskID.Valid && udai.TaskID.String == event.Data.TaskID {
		// Maybe you've restarted the task with new credentials already.
		// TODO: Delete credentials entry?
		udai.TaskID = null.String{}
	}
	udai.Status = models.UserDeviceAPIIntegrationStatusAuthenticationFailure
	if _, err := udai.Update(context.Background(), i.db().Writer, boil.Infer()); err != nil {
		return err
	}

	dd, err := i.DeviceDefSvc.GetDeviceDefinitionByID(ctx, userDevice.DeviceDefinitionID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, fmt.Sprintf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID))
	}

	data := map[string]interface{}{
		"deviceId":     userDeviceID,
		"make_name":    dd.Make.Name,
		"model_name":   dd.Type.Model,
		"model_year":   dd.Type.Year,
		"country_code": userDevice.CountryCode.String,
	}

	if err := i.cio.Track(userDevice.UserID, "tesla.Reauth.Required", data); err != nil {
		i.log.Err(err).Str("userId", userDevice.UserID).Str("userDeviceId", userDeviceID).Msg("Failed to emit reauthentication Customer.io event.")
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
