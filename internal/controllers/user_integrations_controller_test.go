package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/DIMO-Network/shared/db"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UserIntegrationsControllerTestSuite struct {
	suite.Suite
	pdb                       db.Store
	container                 testcontainers.Container
	ctx                       context.Context
	mockCtrl                  *gomock.Controller
	app                       *fiber.App
	scClient                  *mock_services.MockSmartcarClient
	scTaskSvc                 *mock_services.MockSmartcarTaskService
	teslaSvc                  *mock_services.MockTeslaService
	teslaTaskService          *mock_services.MockTeslaTaskService
	autopiAPISvc              *mock_services.MockAutoPiAPIService
	autoPiIngest              *mock_services.MockIngestRegistrar
	autoPiTaskService         *mock_services.MockAutoPiTaskService
	drivlyTaskSvc             *mock_services.MockDrivlyTaskService
	eventSvc                  *mock_services.MockEventService
	deviceDefinitionRegistrar *mock_services.MockDeviceDefinitionRegistrar
	deviceDefSvc              *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc           *mock_services.MockDeviceDefinitionIntegrationService
}

const testUserID = "123123"
const testUser2 = "someOtherUser2"

// SetupSuite starts container db
func (s *UserIntegrationsControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)

	s.mockCtrl = gomock.NewController(s.T())

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	s.scClient = mock_services.NewMockSmartcarClient(s.mockCtrl)
	s.scTaskSvc = mock_services.NewMockSmartcarTaskService(s.mockCtrl)
	s.teslaSvc = mock_services.NewMockTeslaService(s.mockCtrl)
	s.teslaTaskService = mock_services.NewMockTeslaTaskService(s.mockCtrl)
	s.autopiAPISvc = mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	s.autoPiIngest = mock_services.NewMockIngestRegistrar(s.mockCtrl)
	s.deviceDefinitionRegistrar = mock_services.NewMockDeviceDefinitionRegistrar(s.mockCtrl)
	s.drivlyTaskSvc = mock_services.NewMockDrivlyTaskService(s.mockCtrl)
	s.eventSvc = mock_services.NewMockEventService(s.mockCtrl)
	s.autoPiTaskService = mock_services.NewMockAutoPiTaskService(s.mockCtrl)

	logger := test.Logger()
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc,
		s.eventSvc, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), s.autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, s.autoPiTaskService, nil, nil, s.drivlyTaskSvc, nil)
	app := test.SetupAppFiber(*logger)
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	app.Post("/user2/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUser2), c.RegisterDeviceIntegration)
	app.Get("/integrations", c.GetIntegrations)
	app.Post("/user/devices/:userDeviceID/autopi/command", test.AuthInjectorTestHandler(testUserID), c.SendAutoPiCommand)
	app.Get("/user/devices/:userDeviceID/autopi/command/:jobID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiCommandStatus)
	s.app = app
}

// TearDownTest after each test truncate tables
func (s *UserIntegrationsControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserIntegrationsControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

// Test Runner
func TestUserIntegrationsControllerTestSuite(t *testing.T) {
	suite.Run(t, new(UserIntegrationsControllerTestSuite))
}

/* Actual Tests */
func (s *UserIntegrationsControllerTestSuite) TestGetIntegrations() {
	integrations := make([]*ddgrpc.Integration, 2)
	integrations[0] = test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	integrations[1] = test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	s.deviceDefSvc.EXPECT().GetIntegrations(gomock.Any()).Return(integrations, nil)

	request := test.BuildRequest("GET", "/integrations", "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	jsonIntegrations := gjson.GetBytes(body, "integrations")
	assert.True(s.T(), jsonIntegrations.IsArray())
	assert.Equal(s.T(), gjson.GetBytes(body, "integrations.0.id").Str, integrations[0].Id)
	assert.Equal(s.T(), gjson.GetBytes(body, "integrations.1.id").Str, integrations[1].Id)
}
func (s *UserIntegrationsControllerTestSuite) TestPostSmartCarFailure() {
	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mach E", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)

	req := `{
			"code": "qxyz",
			"redirectURI": "http://dimo.zone/cb"
		}`

	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxyz", "http://dimo.zone/cb").Times(1).Return(nil, errors.New("failure communicating with Smartcar"))
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).Times(1).Return(dd, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := s.app.Test(request, 60*1000)
	require.NoError(s.T(), err)
	if !assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should return bad request when given incorrect authorization code") {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "unexpected response: "+string(body))
	}
	exists, _ := models.UserDeviceAPIIntegrationExists(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.Id)
	assert.False(s.T(), exists, "no integration should have been created")
}
func (s *UserIntegrationsControllerTestSuite) TestPostSmartCar() {
	model := "Mach E"
	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	// corrected after query smartcar
	dd2 := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2022, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)

	const smartCarUserID = "smartCarUserId"
	req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
	expiry, _ := time.Parse(time.RFC3339, "2022-03-01T12:00:00Z")
	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxy", "http://dimo.zone/cb").Times(1).Return(&smartcar.Token{
		Access:        "myAccess",
		AccessExpiry:  expiry,
		Refresh:       "myRefresh",
		RefreshExpiry: expiry.Add(24 * time.Hour),
	}, nil)

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *services.Event) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), dd2[0].DeviceDefinitionId, data.Device.DeviceDefinitionID)
			assert.Equal(s.T(), dd2[0].Make.Name, data.Device.Make)
			assert.Equal(s.T(), dd2[0].Type.Model, data.Device.Model)
			assert.Equal(s.T(), int(dd2[0].Type.Year), data.Device.Year)
			assert.Equal(s.T(), "CARVIN", data.Device.VIN)
			assert.Equal(s.T(), ud.ID, data.Device.ID)

			assert.Equal(s.T(), "SmartCar", data.Integration.Vendor)
			assert.Equal(s.T(), integration.Id, data.Integration.ID)
			return nil
		},
	)

	s.deviceDefinitionRegistrar.EXPECT().Register(services.DeviceDefinitionDTO{
		IntegrationID:      integration.Id,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               dd2[0].Make.Name,
		Model:              dd2[0].Type.Model,
		Year:               int(dd2[0].Type.Year),
		Region:             "Americas",
	}).Return(nil)

	s.scClient.EXPECT().GetUserID(gomock.Any(), "myAccess").Return(smartCarUserID, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "myAccess").Return("smartcar-idx", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "myAccess", "smartcar-idx").Return("CARVIN", nil)
	s.scClient.EXPECT().GetEndpoints(gomock.Any(), "myAccess", "smartcar-idx").Return([]string{"/", "/vin"}, nil)
	s.scClient.EXPECT().HasDoorControl(gomock.Any(), "myAccess", "smartcar-idx").Return(false, nil)
	// return a different year than original to fix up device definition id
	s.scClient.EXPECT().GetYear(gomock.Any(), "myAccess", "smartcar-idx").Return(2022, nil)
	s.drivlyTaskSvc.EXPECT().StartDrivlyUpdate(dd[0].DeviceDefinitionId, ud.ID, "CARVIN").Return("task-id-123", nil)

	oUdai := &models.UserDeviceAPIIntegration{}
	s.scTaskSvc.EXPECT().StartPoll(gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
		func(udai *models.UserDeviceAPIIntegration) error {
			oUdai = udai
			return nil
		},
	)
	// original device def
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).Times(2).Return(dd, nil)
	// fixup device def with correct year
	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), "Ford", model, 2022).Return(dd2[0], nil)
	// fixed up device definition
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).Return(dd2, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success") == false {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "unexpected response: "+string(body))
	}
	apiInt, _ := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.Id)

	assert.Equal(s.T(), "zlNpprff", apiInt.AccessToken.String)
	assert.True(s.T(), expiry.Equal(apiInt.AccessExpiresAt.Time))
	assert.Equal(s.T(), "PendingFirstData", apiInt.Status)
	assert.Equal(s.T(), "zlErserfu", apiInt.RefreshToken.String)
}
func (s *UserIntegrationsControllerTestSuite) TestPostUnknownDevice() {
	req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
	request := test.BuildRequest("POST", "/user/devices/fakeDevice/integrations/"+"some-integration", req)
	response, _ := s.app.Test(request)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should fail")
}
func (s *UserIntegrationsControllerTestSuite) TestPostTesla() {
	integration := test.BuildIntegrationGRPC(constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)

	oV := &services.TeslaVehicle{}
	oUdai := &models.UserDeviceAPIIntegration{}

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *services.Event) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), dd[0].Make.Name, data.Device.Make)
			assert.Equal(s.T(), dd[0].Type.Model, data.Device.Model)
			assert.Equal(s.T(), int(dd[0].Type.Year), data.Device.Year)
			assert.Equal(s.T(), "5YJYGDEF9NF010423", data.Device.VIN)
			assert.Equal(s.T(), ud.ID, data.Device.ID)

			assert.Equal(s.T(), constants.TeslaVendor, data.Integration.Vendor)
			assert.Equal(s.T(), integration.Id, data.Integration.ID)
			return nil
		},
	)

	s.deviceDefinitionRegistrar.EXPECT().Register(services.DeviceDefinitionDTO{
		IntegrationID:      integration.Id,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               dd[0].Make.Name,
		Model:              dd[0].Type.Model,
		Year:               int(dd[0].Type.Year),
		Region:             "Americas",
	}).Return(nil)

	s.teslaTaskService.EXPECT().StartPoll(gomock.AssignableToTypeOf(oV), gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
		func(v *services.TeslaVehicle, udai *models.UserDeviceAPIIntegration) error {
			oV = v
			oUdai = udai
			return nil
		},
	)

	s.teslaSvc.EXPECT().GetVehicle("abc", 1145).Return(&services.TeslaVehicle{
		ID:        1145,
		VehicleID: 223,
		VIN:       "5YJYGDEF9NF010423",
	}, nil)
	s.teslaSvc.EXPECT().WakeUpVehicle("abc", 1145).Return(nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).Times(2).Return(dd, nil)
	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dd[0], nil)

	req := `{
			"accessToken": "abc",
			"externalId": "1145",
			"expiresIn": 600,
			"refreshToken": "fffg"
		}`
	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := s.app.Test(request, 60*1000)

	expectedExpiry := time.Now().Add(10 * time.Minute)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success")

	assert.Equal(s.T(), 1145, oV.ID)
	assert.Equal(s.T(), 223, oV.VehicleID)

	within := func(test, reference *time.Time, d time.Duration) bool {
		return test.After(reference.Add(-d)) && test.Before(reference.Add(d))
	}

	apiInt, err := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.Id)
	if err != nil {
		s.T().Fatalf("Couldn't find API integration record: %v", err)
	}
	assert.Equal(s.T(), "nop", apiInt.AccessToken.String)
	assert.Equal(s.T(), "1145", apiInt.ExternalID.String)
	assert.Equal(s.T(), "ssst", apiInt.RefreshToken.String)
	assert.True(s.T(), within(&apiInt.AccessExpiresAt.Time, &expectedExpiry, 15*time.Second), "access token expires at %s, expected something close to %s", apiInt.AccessExpiresAt, expectedExpiry)

}

func (s *UserIntegrationsControllerTestSuite) TestPostTeslaAndUpdateDD() {
	integration := test.BuildIntegrationGRPC(constants.TeslaVendor, 10, 20)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mach E", 2020, integration)

	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)

	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dd[0], nil)

	err := fixTeslaDeviceDefinition(s.ctx, test.Logger(), s.deviceDefSvc, s.pdb.DBS().Writer.DB, integration, &ud, "5YJRE1A31A1P01234")
	if err != nil {
		s.T().Fatalf("Got an error while fixing device definition: %v", err)
	}

	_ = ud.Reload(s.ctx, s.pdb.DBS().Writer.DB)
	// todo, we may need to point to new device def, or see how above fix method is implemented
	if ud.DeviceDefinitionID != dd[0].DeviceDefinitionId {
		s.T().Fatalf("Failed to switch device definition to the correct one")
	}
}

func (s *UserIntegrationsControllerTestSuite) TestPostAutoPiBlockedForDuplicateDeviceSameUser() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	logger := test.Logger()
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, s.autoPiTaskService, nil, nil, s.drivlyTaskSvc, nil)
	app := test.SetupAppFiber(*logger)
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	// arrange
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 34, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model 4", 2020, integration)

	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integration.Id, s.pdb)

	req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)
	// no calls should be made to autopi api

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).AnyTimes().Return(dd, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := app.Test(request, 1000*240)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should return failure")
	body, _ := io.ReadAll(response.Body)
	errorMsg := gjson.Get(string(body), "message").String()
	assert.Equal(s.T(),
		fmt.Sprintf("userDeviceId %s already has a user_device_api_integration with integrationId %s, please delete that first", ud.ID, integration.Id), errorMsg)
}
func (s *UserIntegrationsControllerTestSuite) TestPostAutoPiBlockedForDuplicateDeviceDifferentUser() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	logger := test.Logger()
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, s.autoPiTaskService, nil, nil, s.drivlyTaskSvc, nil)
	app := test.SetupAppFiber(*logger)
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUser2), c.RegisterDeviceIntegration)
	// arrange
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 34, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model 4", 2022, nil)
	// the other user that already claimed unit
	_ = test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	// test user
	ud2 := test.SetupCreateUserDevice(s.T(), testUser2, dd[0].DeviceDefinitionId, nil, s.pdb)

	req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd[0].DeviceDefinitionId}).Times(1).Return(dd, nil)

	// no calls should be made to autopi api
	request := test.BuildRequest("POST", "/user/devices/"+ud2.ID+"/integrations/"+integration.Id, req)
	response, err := app.Test(request, 2000)
	require.NoError(s.T(), err)
	if !assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should return bad request") {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "body response: "+string(body)+"\n")
	}
}

func (s *UserIntegrationsControllerTestSuite) TestPostAutoPiCommand() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, nil)
	app := fiber.New()
	app.Post("/user/devices/:userDeviceID/autopi/command", test.AuthInjectorTestHandler(testUserID), c.SendAutoPiCommand)
	// arrange
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 34, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model 4", 2022, nil)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	udapiInt := test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integration.Id, s.pdb)

	udAPIMetadata := services.UserDeviceAPIIntegrationsMetadata{
		AutoPiUnitID: func(s string) *string { return &s }(unitID),
	}
	_ = udapiInt.Metadata.Marshal(udAPIMetadata)
	_, err := udapiInt.Update(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	autoPiJob := models.AutopiJob{
		ID:                 "somepreviousjobId",
		AutopiDeviceID:     deviceID,
		Command:            "raw",
		State:              "COMMAND_EXECUTED",
		CommandLastUpdated: null.TimeFrom(time.Now().UTC()),
		UserDeviceID:       null.StringFrom(ud.ID),
	}
	err = autoPiJob.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	// test job can be retrieved
	apSvc := services.NewAutoPiAPIService(&config.Settings{}, s.pdb.DBS)
	status, _, err := apSvc.GetCommandStatus(s.ctx, "somepreviousjobId")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "somepreviousjobId", status.CommandJobID)
	assert.Equal(s.T(), autoPiJob.State, status.CommandState)
	assert.Equal(s.T(), "raw", status.CommandRaw)

	// test sending a command from api
	const jobID = "123"
	// mock expectations
	const cmd = "raw test"
	autopiAPISvc.EXPECT().CommandRaw(gomock.Any(), unitID, deviceID, cmd, ud.ID).Return(&services.AutoPiCommandResponse{
		Jid:     jobID,
		Minions: nil,
	}, nil)
	// act: send request
	req := fmt.Sprintf(`{
			"command": "%s"
		}`, cmd)

	s.deviceDefIntSvc.EXPECT().FindUserDeviceAutoPiIntegration(gomock.Any(), gomock.Any(), ud.ID, testUserID).Times(1).Return(&udapiInt, &udAPIMetadata, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/autopi/command", req)
	response, _ := app.Test(request, 20000)
	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	jid := gjson.GetBytes(body, "jid")
	assert.Equal(s.T(), jobID, jid.String())
}

func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiCommand() {
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, nil)
	app := fiber.New()
	app.Get("/user/devices/:userDeviceID/autopi/command/:jobID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiCommandStatus)
	//arrange
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 34, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model 4", 2022, nil)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		jobID    = "somepreviousjobId"
	)
	_ = test.SetupCreateUserDeviceAPIIntegration(s.T(), "", deviceID, ud.ID, integration.Id, s.pdb)

	lastUpdated := time.Now()

	autopiAPISvc.EXPECT().GetCommandStatus(gomock.Any(), jobID).Return(&services.AutoPiCommandJob{
		CommandJobID: jobID,
		CommandState: "COMMAND_EXECUTED",
		CommandRaw:   "raw",
		LastUpdated:  &lastUpdated,
	}, &models.AutopiJob{
		ID:                 jobID,
		AutopiDeviceID:     deviceID,
		Command:            "raw",
		State:              "COMMAND_EXECUTED",
		CommandLastUpdated: null.TimeFrom(lastUpdated),
		UserDeviceID:       null.StringFrom(ud.ID),
	}, nil)

	// act: send request
	request := test.BuildRequest("GET", "/user/devices/"+ud.ID+"/autopi/command/"+jobID, "")
	response, _ := app.Test(request)
	require.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), jobID, gjson.GetBytes(body, "commandJobId").String())
	assert.Equal(s.T(), "COMMAND_EXECUTED", gjson.GetBytes(body, "commandState").String())
	assert.Equal(s.T(), "raw", gjson.GetBytes(body, "commandRaw").String())

}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiCommandNoResults400() {
	//arrange
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 34, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model 4", 2022, nil)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, s.pdb)
	const (
		jobID    = "somepreviousjobId2"
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integration.Id, s.pdb)

	s.autopiAPISvc.EXPECT().GetCommandStatus(gomock.Any(), jobID).Return(nil, nil, sql.ErrNoRows)

	// act: send request
	request := test.BuildRequest("GET", "/user/devices/"+ud.ID+"/autopi/command/"+jobID, "")
	response, _ := s.app.Test(request)
	//assert
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_ShouldUpdate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, nil)
	app := fiber.New()
	app.Get("/autopi/unit/:unitID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		IsUpdated:         false,
		UnitID:            unitID,
		ID:                "4321",
		HwRevision:        "1.23",
		Template:          10,
		LastCommunication: time.Now(),
		Release: struct {
			Version string `json:"version"`
		}(struct{ Version string }{Version: "1.21.6"}),
	}, nil)
	autopiAPISvc.EXPECT().GetUserDeviceIntegrationByUnitID(gomock.Any(), unitID).Return(nil, nil)
	// act
	request := test.BuildRequest("GET", "/autopi/unit/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), false, gjson.GetBytes(body, "isUpdated").Bool())
	assert.Equal(s.T(), unitID, gjson.GetBytes(body, "unitId").String())
	assert.Equal(s.T(), "4321", gjson.GetBytes(body, "deviceId").String())
	assert.Equal(s.T(), "1.23", gjson.GetBytes(body, "hwRevision").String())
	assert.Equal(s.T(), "1.21.6", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), true, gjson.GetBytes(body, "shouldUpdate").Bool()) // this because releaseVersion below 1.21.9
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_UpToDate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, nil)
	app := fiber.New()
	app.Get("/autopi/unit/:unitID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		IsUpdated:         true,
		UnitID:            unitID,
		ID:                "4321",
		HwRevision:        "1.23",
		Template:          10,
		LastCommunication: time.Now(),
		Release: struct {
			Version string `json:"version"`
		}(struct{ Version string }{Version: "1.21.9"}),
	}, nil)
	autopiAPISvc.EXPECT().GetUserDeviceIntegrationByUnitID(gomock.Any(), unitID).Return(nil, nil)
	// act
	request := test.BuildRequest("GET", "/autopi/unit/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), true, gjson.GetBytes(body, "isUpdated").Bool())
	assert.Equal(s.T(), "1.21.9", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), false, gjson.GetBytes(body, "shouldUpdate").Bool()) // returned version is 1.21.9 which is our cutoff
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_FutureUpdate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, nil)
	app := fiber.New()
	app.Get("/autopi/unit/:unitID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		IsUpdated:         false,
		UnitID:            unitID,
		ID:                "4321",
		HwRevision:        "1.23",
		Template:          10,
		LastCommunication: time.Now(),
		Release: struct {
			Version string `json:"version"`
		}(struct{ Version string }{Version: "1.23.1"}),
	}, nil)
	autopiAPISvc.EXPECT().GetUserDeviceIntegrationByUnitID(gomock.Any(), unitID).Return(nil, nil)
	// act
	request := test.BuildRequest("GET", "/autopi/unit/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), false, gjson.GetBytes(body, "isUpdated").Bool())
	assert.Equal(s.T(), "1.23.1", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), false, gjson.GetBytes(body, "shouldUpdate").Bool())
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoMatchUDAI() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, nil)
	app := fiber.New()
	app.Get("/autopi/unit/:unitID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 34, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model 4", 2022, nil)
	ud := test.SetupCreateUserDevice(s.T(), "some-other-user", dd[0].DeviceDefinitionId, nil, s.pdb)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }("1234"), s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, "321", ud.ID, integration.Id, s.pdb)

	udai := models.UserDeviceAPIIntegration{}
	udai.R = udai.R.NewStruct()
	udai.R.UserDevice = &ud
	autopiAPISvc.EXPECT().GetUserDeviceIntegrationByUnitID(gomock.Any(), unitID).Return(&udai, nil)

	// act
	request := test.BuildRequest("GET", "/autopi/unit/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusForbidden, response.StatusCode)
}
