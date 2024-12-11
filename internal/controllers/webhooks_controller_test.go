package controllers

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/constants"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
)

type WebHooksControllerTestSuite struct {
	suite.Suite
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
	mockCtrl  *gomock.Controller
}

// SetupSuite starts container db
func (s *WebHooksControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	s.mockCtrl = gomock.NewController(s.T())
}

// TearDownTest after each test truncate tables
func (s *WebHooksControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *WebHooksControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

// Test Runner
func TestWebHooksControllerTestSuite(t *testing.T) {
	suite.Run(t, new(WebHooksControllerTestSuite))
}

/* Actual Tests */
func (s *WebHooksControllerTestSuite) TestPostWebhook401InvalidSignature() {

	ddDefIntSvc := mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	autoAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)

	token := "BobbyHarry"
	c := NewWebhooksController(&config.Settings{AutoPiAPIToken: token}, s.pdb.DBS, test.Logger(), autoAPISvc, ddDefIntSvc)
	app := fiber.New()
	app.Post(constants.AutoPiWebhookPath, c.ProcessCommand)

	webhookJSON := `{
			"jid": "20220414153005426360",
    		"state": "COMMAND_EXECUTED",
    		"success": true,
    		"device_id": "26b1f359-1799-4a21-8e4e-1ad7607fe5af"
			}`

	request := test.BuildRequest("POST", constants.AutoPiWebhookPath, webhookJSON)
	request.Header.Set("X-Request-Signature", "") // copy this from replit python generated value
	response, _ := app.Test(request)
	// assert
	assert.Equal(s.T(), 401, response.StatusCode)
}
func (s *WebHooksControllerTestSuite) TestPostWebhook400BadPayload() {

	ddDefIntSvc := mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	autoAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)

	token := "BobbyHarry"
	c := NewWebhooksController(&config.Settings{AutoPiAPIToken: token}, s.pdb.DBS, test.Logger(), autoAPISvc, ddDefIntSvc)
	app := fiber.New()
	app.Post(constants.AutoPiWebhookPath, c.ProcessCommand)

	webhookJSON := `{"success": true,"device_id": "26b1f359-1799-4a21-8e4e-1ad7607fe5af"}`
	request := test.BuildRequest("POST", constants.AutoPiWebhookPath, webhookJSON)
	request.Header.Set("X-Request-Signature", "ade42bd1085401a581722e2003e995adea80ffd81dfb42877b185abc48ddc3fd") // copy this from replit python generated value
	response, _ := app.Test(request)
	// assert
	assert.Equal(s.T(), 400, response.StatusCode)
}
func (s *WebHooksControllerTestSuite) TestPostWebhookSyncCommand() {
	// arrange

	ddDefIntSvc := mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	autoAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)

	token := "BobbyHarry"
	c := NewWebhooksController(&config.Settings{AutoPiAPIToken: token}, s.pdb.DBS, test.Logger(), autoAPISvc, ddDefIntSvc)
	app := fiber.New()
	app.Post(constants.AutoPiWebhookPath, c.ProcessCommand)

	testUserID := ksuid.New().String()
	autoPiDeviceID := "123123"
	autoPiTemplateID := 987
	autoPiJobID := "AD111"
	integ := test.BuildIntegrationGRPC("autopi123", constants.AutoPiVendor, autoPiTemplateID, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model X", 2020, integ)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	autopiJob := test.SetupCreateAutoPiJob(s.T(), autoPiJobID, autoPiDeviceID, "state.sls pending", ud.ID, "COMMAND_EXECUTED", "", s.pdb)

	ddDefIntSvc.EXPECT().GetAutoPiIntegration(gomock.Any()).Return(integ, nil)

	autoAPISvc.EXPECT().UpdateJob(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(autopiJob, nil)
	ci := constants.FindCountry(ud.CountryCode.String)
	autoAPISvc.EXPECT().UpdateState(gomock.Any(), gomock.Any(), ud.CountryCode.String, ci.Region).Return(nil)

	udiai := models.UserDeviceAPIIntegration{
		// create user device api integration
		UserDeviceID:  ud.ID,
		IntegrationID: integ.Id,
		Status:        models.UserDeviceAPIIntegrationStatusPending,
		ExternalID:    null.StringFrom(autoPiDeviceID),
	}
	err := udiai.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	//_ = test.SetupCreateAutoPiJob(s.T(), autoPiJobID, autoPiDeviceID, "state.sls pending", ud.ID, s.pdb)

	// act
	webhookJSON := fmt.Sprintf(`{"jid": "%s","state": "COMMAND_EXECUTED","success": true,"device_id": "%s"}`, autoPiJobID, autoPiDeviceID)

	request := test.BuildRequest("POST", constants.AutoPiWebhookPath, webhookJSON)
	// signature generated by python per example code from autopi (for above payload)
	request.Header.Set("X-Request-Signature", "93c5e5e140fc132f7871f890790d0aa83509a9ba077a4a5fe9f6595f38dd470c")
	response, _ := app.Test(request)

	// assert
	require.Equal(s.T(), 204, response.StatusCode)
	// check the database has the expected change in status, and `auto_pi_sync_command_state` in metadata
	updatedUdiai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(ud.ID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integ.Id)).
		One(s.ctx, s.pdb.DBS().Writer)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), models.UserDeviceAPIIntegrationStatusPendingFirstData, updatedUdiai.Status)

	metadata := new(services.UserDeviceAPIIntegrationsMetadata)
	err = updatedUdiai.Metadata.Unmarshal(metadata)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), constants.TemplateConfirmed.String(), *metadata.AutoPiSubStatus)

	job, err := models.AutopiJobs(models.AutopiJobWhere.ID.EQ(autoPiJobID)).One(s.ctx, s.pdb.DBS().Reader)
	assert.NoError(s.T(), err)

	assert.NotNilf(s.T(), job, "autopi job should not be nil")
	assert.Equal(s.T(), "COMMAND_EXECUTED", job.State)
	assert.Equal(s.T(), autoPiJobID, job.ID)
	assert.NotEqual(s.T(), job.CommandLastUpdated.Time.String(), job.CreatedAt.String(),
		"expected updated job to have later time than when originally created")
}
func (s *WebHooksControllerTestSuite) TestPostWebhookRawCommand() {
	// arrange
	ddDefIntSvc := mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	autoAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)

	token := "BobbyHarry"
	c := NewWebhooksController(&config.Settings{AutoPiAPIToken: token}, s.pdb.DBS, test.Logger(), autoAPISvc, ddDefIntSvc)
	app := fiber.New()
	app.Post(constants.AutoPiWebhookPath, c.ProcessCommand)

	testUserID := ksuid.New().String()
	autoPiDeviceID := "123123"
	autoPiTemplateID := 987
	autoPiJobID := "AD111"
	integ := test.BuildIntegrationGRPC("autopi123", constants.AutoPiVendor, autoPiTemplateID, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model X", 2020, integ)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	// create user device api integration
	commandResult := `{ "value": "123", "type": "vin" }`

	autopiJob := test.SetupCreateAutoPiJob(s.T(), autoPiJobID, autoPiDeviceID, "some raw command", ud.ID, "COMMAND_EXECUTED", commandResult, s.pdb)

	//ddDefIntSvc.EXPECT().GetAutoPiIntegration(gomock.Any()).Return(integ, nil)

	autoAPISvc.EXPECT().UpdateJob(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(autopiJob, nil)
	//autoAPISvc.EXPECT().UpdateState(gomock.Any(), gomock.Any()).Return(nil)

	udiai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integ.Id,
		Status:        models.UserDeviceAPIIntegrationStatusPending, // assert this does not get changed since just raw command
		ExternalID:    null.StringFrom(autoPiDeviceID),
	}
	err := udiai.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	assert.NoError(s.T(), err)

	// act
	webhookJSON := fmt.Sprintf(`{"jid": "%s","state": "COMMAND_EXECUTED","success": true,"device_id": "%s", "response": { "tag": "salt/job/123", "data": { "return": { "value": "123", "_type": "vin" }}} }`, autoPiJobID, autoPiDeviceID)

	request := test.BuildRequest("POST", constants.AutoPiWebhookPath, webhookJSON)
	// signature generated by python per example code from autopi (for above payload) https://replit.com/@JamesReategui1/HmacPlayground#main.py
	request.Header.Set("X-Request-Signature", "5573091add7aecb98d3204944f8bb3887026ff51401f421de15ec3deccbad45b")
	response, _ := app.Test(request, 10000)

	// assert
	assert.Equal(s.T(), 204, response.StatusCode)
	// check the database has the expected change in status, and `auto_pi_sync_command_state` in metadata
	updatedUdiai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(ud.ID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integ.Id)).
		One(s.ctx, s.pdb.DBS().Writer)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), models.UserDeviceAPIIntegrationStatusPending, updatedUdiai.Status) // this should not change for regular raw commands

	job, err := models.AutopiJobs(models.AutopiJobWhere.ID.EQ(autoPiJobID)).One(s.ctx, s.pdb.DBS().Reader)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "COMMAND_EXECUTED", job.State)
	assert.Equal(s.T(), autoPiJobID, job.ID)
	assert.NotEqual(s.T(), job.CommandLastUpdated.Time.String(), udiai.UpdatedAt.String(),
		"expected updated job to have later time than original integration")
	require.True(s.T(), job.CommandResult.Valid)
	cmdResult := new(services.AutoPiCommandResult)
	err = job.CommandResult.Unmarshal(cmdResult)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "123", cmdResult.Value)
	assert.Equal(s.T(), "vin", cmdResult.Type)
}
