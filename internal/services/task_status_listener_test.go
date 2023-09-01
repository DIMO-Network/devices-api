package services

//
//import (
//	"context"
//	"github.com/DIMO-Network/devices-api/internal/constants"
//	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
//	"go.uber.org/mock/gomock"
//	"os"
//	"testing"
//
//	"github.com/DIMO-Network/devices-api/internal/test"
//	"github.com/DIMO-Network/devices-api/models"
//	"github.com/DIMO-Network/shared"
//	"github.com/rs/zerolog"
//	"github.com/segmentio/ksuid"
//	"github.com/stretchr/testify/assert"
//	"github.com/volatiletech/sqlboiler/v4/boil"
//)
//
//type fakeCIOEvent struct {
//	CustomerID string
//	EventName  string
//	Data       map[string]interface{}
//}
//
//type fakeCIO struct {
//	Events []fakeCIOEvent
//}
//
//func (c *fakeCIO) Track(customerID string, eventName string, data map[string]interface{}) error {
//	c.Events = append(c.Events, fakeCIOEvent{customerID, eventName, data})
//	return nil
//}
//
//func TestTaskStatusListener(t *testing.T) {
//	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "devices-api").Logger()
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//
//	deviceDefSvc := mock_services.NewMockDeviceDefinitionService(mockCtrl)
//
//	ctx := context.Background()
//	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
//	defer func() {
//		if err := container.Terminate(ctx); err != nil {
//			t.Fatal(err)
//		}
//	}()
//
//	cio := new(fakeCIO)
//	ingest := NewTaskStatusListener(pdb.dbs, &logger, cio, deviceDefSvc)
//
//	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
//	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2021, integration)
//	ud := test.SetupCreateUserDevice(t, "dylan", dd[0].DeviceDefinitionId, nil, pdb)
//
//	udai := models.UserDeviceAPIIntegration{
//		UserDeviceID:  ud.ID,
//		IntegrationID: integration.ID,
//		Status:        models.UserDeviceAPIIntegrationStatusActive,
//	}
//	err := udai.Insert(ctx, pdb.dbs().Writer, boil.Infer())
//	assert.NoError(t, err)
//
//	input := &shared.CloudEvent[TaskStatusData]{
//		Source:      "dimo/integration/" + integration.ID,
//		SpecVersion: "1.0",
//		Subject:     ud.ID,
//		Type:        "zone.dimo.task.smartcar.poll.status.update",
//		Data: TaskStatusData{
//			TaskID:        ksuid.New().String(),
//			UserDeviceID:  ud.ID,
//			IntegrationID: integration.ID,
//			Status:        "AuthenticationFailure",
//		},
//	}
//
//	if err := ingest.processEvent(input); err != nil {
//		t.Fatalf("Got an unexpected error processing status update: %v", err)
//	}
//
//	if err := udai.Reload(ctx, pdb.dbs().Writer); err != nil {
//		t.Fatalf("Couldn't reload UDAI: %v", err)
//	}
//
//	assert.Equal(t, models.UserDeviceAPIIntegrationStatusAuthenticationFailure, udai.Status, "New status should be AuthenticationFailure.")
//
//	assert.Len(t, cio.Events, 1, "Should have emitted one CIO event.")
//
//	event := cio.Events[0]
//
//	assert.Equal(t, "dylan", event.CustomerID)
//	assert.Equal(t, "smartcar.Reauth.Required", event.EventName)
//	assert.Equal(t, map[string]any{
//		"deviceId":     ud.ID,
//		"make_name":    "Tesla",
//		"model_name":   "Model Y",
//		"model_year":   int16(2021),
//		"country_code": "USA",
//	}, event.Data)
//}
