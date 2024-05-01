package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/shared/redis/mocks"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	pbuser "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	smock "github.com/Shopify/sarama/mocks"
	"github.com/gofiber/fiber/v2"
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
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.uber.org/mock/gomock"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/middleware/owner"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
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
	eventSvc                  *mock_services.MockEventService
	deviceDefinitionRegistrar *mock_services.MockDeviceDefinitionRegistrar
	deviceDefSvc              *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc           *mock_services.MockDeviceDefinitionIntegrationService
	redisClient               *mocks.MockCacheService
	userClient                *mock_services.MockUserServiceClient
	natsSvc                   *services.NATSService
	natsServer                *server.Server
	userDeviceSvc             *mock_services.MockUserDeviceService
	cipher                    shared.Cipher
	teslaFleetAPISvc          *mock_services.MockTeslaFleetAPIService
}

const testUserID = "123123"
const testUser2 = "someOtherUser2"

// SetupSuite starts container db
func (s *UserIntegrationsControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)

	s.mockCtrl = gomock.NewController(s.T(), gomock.WithOverridableExpectations())
	var err error

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	s.scClient = mock_services.NewMockSmartcarClient(s.mockCtrl)
	s.scTaskSvc = mock_services.NewMockSmartcarTaskService(s.mockCtrl)
	s.teslaSvc = mock_services.NewMockTeslaService(s.mockCtrl)
	s.teslaTaskService = mock_services.NewMockTeslaTaskService(s.mockCtrl)
	s.autopiAPISvc = mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	s.autoPiIngest = mock_services.NewMockIngestRegistrar(s.mockCtrl)
	s.deviceDefinitionRegistrar = mock_services.NewMockDeviceDefinitionRegistrar(s.mockCtrl)
	s.eventSvc = mock_services.NewMockEventService(s.mockCtrl)
	s.autoPiTaskService = mock_services.NewMockAutoPiTaskService(s.mockCtrl)
	s.redisClient = mocks.NewMockCacheService(s.mockCtrl)
	s.userClient = mock_services.NewMockUserServiceClient(s.mockCtrl)
	s.natsSvc, s.natsServer, err = mock_services.NewMockNATSService(natsStreamName)
	s.userDeviceSvc = mock_services.NewMockUserDeviceService(s.mockCtrl)
	s.teslaFleetAPISvc = mock_services.NewMockTeslaFleetAPIService(s.mockCtrl)
	s.cipher = new(shared.ROT13Cipher)

	if err != nil {
		s.T().Fatal(err)
	}

	logger := test.Logger()
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, s.eventSvc, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, s.cipher, s.autopiAPISvc,
		s.autoPiIngest, s.deviceDefinitionRegistrar, s.autoPiTaskService, nil, nil, nil, s.redisClient, nil, s.userClient, nil, s.natsSvc, nil, s.userDeviceSvc,
		s.teslaFleetAPISvc)

	app := test.SetupAppFiber(*logger)

	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	app.Delete("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.DeleteUserDeviceIntegration)

	app.Post("/user2/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUser2), c.RegisterDeviceIntegration)
	app.Get("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.GetUserDeviceIntegration)

	s.app = app
}

// TearDownTest after each test truncate tables
func (s *UserIntegrationsControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserIntegrationsControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	s.natsServer.Shutdown()
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

func (s *UserIntegrationsControllerTestSuite) TestPostSmartCarFailure() {
	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mach E", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)

	req := `{
			"code": "qxyz",
			"redirectURI": "http://dimo.zone/cb"
		}`
	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxyz", "http://dimo.zone/cb").Times(1).Return(nil, errors.New("failure communicating with Smartcar"))
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Times(1).Return(dd[0], nil)

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

func (s *UserIntegrationsControllerTestSuite) TestDeleteIntegration_BlockedBySyntheticDevice() {
	model := "Mach E"
	const vin = "CARVIN"
	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	vnft := test.SetupCreateVehicleNFT(s.T(), ud.ID, vin, big.NewInt(5), null.BytesFrom(common.HexToAddress("0xA1").Bytes()), s.pdb)

	mtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusConfirmed,
	}

	s.Require().NoError(mtr.Insert(context.TODO(), s.pdb.DBS().Writer, boil.Infer()))

	sd := models.SyntheticDevice{
		VehicleTokenID:     vnft.TokenID,
		IntegrationTokenID: types.NewDecimal(decimal.New(int64(integration.TokenId), 0)),
		MintRequestID:      mtr.ID,
		WalletChildNumber:  4,
		WalletAddress:      common.HexToAddress("0xB").Bytes(),
		TokenID:            types.NewNullDecimal(decimal.New(6, 0)),
	}
	s.Require().NoError(sd.Insert(context.TODO(), s.pdb.DBS().Writer, boil.Infer()))

	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)

	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "c005c7dd-9568-4083-8989-109205cdff28", ud.ID, integration.Id, s.pdb)

	request := test.BuildRequest("DELETE", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusConflict, response.StatusCode)
	fmt.Println(response, err)
}

func (s *UserIntegrationsControllerTestSuite) TestPostSmartCar_SuccessNewToken() {
	model := "Mach E"
	const vin = "CARVIN"
	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)

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
		func(event *shared.CloudEvent[any]) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), dd[0].DeviceDefinitionId, data.Device.DeviceDefinitionID)
			assert.Equal(s.T(), dd[0].Make.Name, data.Device.Make)
			assert.Equal(s.T(), dd[0].Type.Model, data.Device.Model)
			assert.Equal(s.T(), int(dd[0].Type.Year), data.Device.Year)
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
		Make:               dd[0].Make.Name,
		Model:              dd[0].Type.Model,
		Year:               int(dd[0].Type.Year),
		Region:             "Americas",
	}).Return(nil)

	// original device def
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Times(2).Return(dd[0], nil)
	s.scClient.EXPECT().GetUserID(gomock.Any(), "myAccess").Return(smartCarUserID, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "myAccess").Return("smartcar-idx", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "myAccess", "smartcar-idx").Return(vin, nil)
	s.scClient.EXPECT().GetEndpoints(gomock.Any(), "myAccess", "smartcar-idx").Return([]string{"/", "/vin"}, nil)
	s.scClient.EXPECT().HasDoorControl(gomock.Any(), "myAccess", "smartcar-idx").Return(false, nil)

	oUdai := &models.UserDeviceAPIIntegration{}
	s.scTaskSvc.EXPECT().StartPoll(gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
		func(udai *models.UserDeviceAPIIntegration) error {
			oUdai = udai
			return nil
		},
	)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success") == false {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "unexpected response: "+string(body))
	}
	apiInt, _ := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.Id)
	updatedUD, _ := models.FindUserDevice(s.ctx, s.pdb.DBS().Reader, ud.ID)

	assert.Equal(s.T(), "zlNpprff", apiInt.AccessToken.String)
	assert.True(s.T(), expiry.Equal(apiInt.AccessExpiresAt.Time))
	assert.Equal(s.T(), "PendingFirstData", apiInt.Status)
	assert.Equal(s.T(), "zlErserfu", apiInt.RefreshToken.String)
	assert.Equal(s.T(), vin, updatedUD.VinIdentifier.String)
	assert.Equal(s.T(), true, updatedUD.VinConfirmed)
}

func (s *UserIntegrationsControllerTestSuite) TestPostSmartCar_FailureTestVIN() {
	model := "Mach E"
	const vin = "0SC12312312312312"
	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)

	const smartCarUserID = "smartCarUserId"
	req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
	expiry, _ := time.Parse(time.RFC3339, "2022-03-01T12:00:00Z")
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Return(dd[0], nil)
	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxy", "http://dimo.zone/cb").Times(1).Return(&smartcar.Token{
		Access:        "myAccess",
		AccessExpiry:  expiry,
		Refresh:       "myRefresh",
		RefreshExpiry: expiry.Add(24 * time.Hour),
	}, nil)
	s.scClient.EXPECT().GetUserID(gomock.Any(), "myAccess").Return(smartCarUserID, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "myAccess").Return("smartcar-idx", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "myAccess", "smartcar-idx").Return(vin, nil)

	logger := test.Logger()
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: "prod"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, s.eventSvc, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), s.autopiAPISvc,
		s.autoPiIngest, s.deviceDefinitionRegistrar, s.autoPiTaskService, nil, nil, nil, s.redisClient, nil, nil, nil, s.natsSvc, nil, s.userDeviceSvc, nil)

	app := test.SetupAppFiber(*logger)

	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	if assert.Equal(s.T(), fiber.StatusConflict, response.StatusCode, "should return failure") == false {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "unexpected response: "+string(body))
	}
}

func (s *UserIntegrationsControllerTestSuite) TestPostSmartCar_SuccessCachedToken() {
	model := "Mach E"
	const vin = "CARVIN"
	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	ud.VinIdentifier = null.StringFrom(vin)
	ud.VinConfirmed = true
	_, err := ud.Update(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)

	const smartCarUserID = "smartCarUserId"
	req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
	expiry, _ := time.Parse(time.RFC3339, "2022-03-01T12:00:00Z")
	// token found in cache
	token := &smartcar.Token{
		Access:        "some-access-code",
		AccessExpiry:  expiry,
		Refresh:       "some-refresh-code",
		RefreshExpiry: expiry,
		ExpiresIn:     3000,
	}
	tokenJSON, err := json.Marshal(token)
	require.NoError(s.T(), err)
	cipher := new(shared.ROT13Cipher)
	encrypted, err := cipher.Encrypt(string(tokenJSON))
	require.NoError(s.T(), err)
	s.redisClient.EXPECT().Get(gomock.Any(), buildSmartcarTokenKey(vin, testUserID)).Return(redis.NewStringResult(encrypted, nil))
	s.redisClient.EXPECT().Del(gomock.Any(), buildSmartcarTokenKey(vin, testUserID)).Return(redis.NewIntResult(1, nil))
	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *shared.CloudEvent[any]) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), dd[0].DeviceDefinitionId, data.Device.DeviceDefinitionID)
			assert.Equal(s.T(), dd[0].Make.Name, data.Device.Make)
			assert.Equal(s.T(), dd[0].Type.Model, data.Device.Model)
			assert.Equal(s.T(), int(dd[0].Type.Year), data.Device.Year)
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
		Make:               dd[0].Make.Name,
		Model:              dd[0].Type.Model,
		Year:               int(dd[0].Type.Year),
		Region:             "Americas",
	}).Return(nil)

	s.scClient.EXPECT().GetUserID(gomock.Any(), token.Access).Return(smartCarUserID, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), token.Access).Return("smartcar-idx", nil)
	s.scClient.EXPECT().GetEndpoints(gomock.Any(), token.Access, "smartcar-idx").Return([]string{"/", "/vin"}, nil)
	s.scClient.EXPECT().HasDoorControl(gomock.Any(), token.Access, "smartcar-idx").Return(false, nil)

	oUdai := &models.UserDeviceAPIIntegration{}
	s.scTaskSvc.EXPECT().StartPoll(gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
		func(udai *models.UserDeviceAPIIntegration) error {
			oUdai = udai
			return nil
		},
	)
	// original device def
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Times(2).Return(dd[0], nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success") == false {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "unexpected response: "+string(body))
	}
	apiInt, _ := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.Id)

	assert.Equal(s.T(), "fbzr-npprff-pbqr", apiInt.AccessToken.String)
	assert.True(s.T(), expiry.Equal(apiInt.AccessExpiresAt.Time))
	assert.Equal(s.T(), "PendingFirstData", apiInt.Status)
	assert.Equal(s.T(), "fbzr-erserfu-pbqr", apiInt.RefreshToken.String)
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
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)

	oV := &services.TeslaVehicle{}
	oUdai := &models.UserDeviceAPIIntegration{}

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *shared.CloudEvent[any]) error {
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

	s.teslaTaskService.EXPECT().StartPoll(gomock.AssignableToTypeOf(oV), gomock.AssignableToTypeOf(oUdai), 1, "").DoAndReturn(
		func(v *services.TeslaVehicle, udai *models.UserDeviceAPIIntegration, _ int, _ string) error {
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
	s.teslaSvc.EXPECT().GetAvailableCommands().Return(&services.UserDeviceAPIIntegrationsMetadataCommands{Enabled: []string{constants.DoorsUnlock, constants.DoorsLock}})
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Times(2).Return(dd[0], nil)
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

	meta := &services.UserDeviceAPIIntegrationsMetadata{}
	err = apiInt.Metadata.Unmarshal(&meta)
	s.Assert().NoError(err)

	s.Assert().Equal(1, meta.TeslaAPIVersion)
}

func (s *UserIntegrationsControllerTestSuite) TestPostTeslaAndUpdateDD() {
	integration := test.BuildIntegrationGRPC(constants.TeslaVendor, 10, 20)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mach E", 2020, integration)

	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)

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
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, s.autoPiTaskService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := test.SetupAppFiber(*logger)
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	// arrange
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 34, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model 4", 2020, integration)

	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAftermarketDevice(s.T(), testUserID, nil, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integration.Id, s.pdb)

	req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)
	// no calls should be made to autopi api

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).AnyTimes().Return(dd[0], nil)

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
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, s.autoPiTaskService, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := test.SetupAppFiber(*logger)
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUser2), c.RegisterDeviceIntegration)
	// arrange
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 34, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Testla", "Model 4", 2022, nil)
	// the other user that already claimed unit
	_ = test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAftermarketDevice(s.T(), testUserID, nil, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	// test user
	ud2 := test.SetupCreateUserDevice(s.T(), testUser2, dd[0].DeviceDefinitionId, nil, "", s.pdb)

	req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)

	// no calls should be made to autopi api
	request := test.BuildRequest("POST", "/user/devices/"+ud2.ID+"/integrations/"+integration.Id, req)
	response, err := app.Test(request, 2000)
	require.NoError(s.T(), err)
	if !assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should return bad request") {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "body response: "+string(body)+"\n")
	}
}

func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_ShouldUpdate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID), owner.AftermarketDevice(s.pdb, s.userClient, &logger), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	test.SetupCreateAftermarketDevice(s.T(), "", nil, unitID, nil, s.pdb)
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

	s.deviceDefSvc.EXPECT().GetMakeByTokenID(gomock.Any(), gomock.Any()).Return(&ddgrpc.DeviceMake{Name: "AutoPi"}, nil)

	// act
	request := test.BuildRequest("GET", "/aftermarket/device/by-serial/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	// assert
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
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID), owner.AftermarketDevice(s.pdb, s.userClient, &logger), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	test.SetupCreateAftermarketDevice(s.T(), "", nil, unitID, nil, s.pdb)
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		IsUpdated:         true,
		UnitID:            unitID,
		ID:                "4321",
		HwRevision:        "1.23",
		Template:          10,
		LastCommunication: time.Now(),
		Release: struct {
			Version string `json:"version"`
		}(struct{ Version string }{Version: "1.22.8"}),
	}, nil)

	s.deviceDefSvc.EXPECT().GetMakeByTokenID(gomock.Any(), gomock.Any()).Return(&ddgrpc.DeviceMake{Name: "AutoPi"}, nil)

	// act
	request := test.BuildRequest("GET", "/aftermarket/device/by-serial/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), true, gjson.GetBytes(body, "isUpdated").Bool())
	assert.Equal(s.T(), "1.22.8", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), false, gjson.GetBytes(body, "shouldUpdate").Bool()) // returned version is 1.21.9 which is our cutoff
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_FutureUpdate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID), owner.AftermarketDevice(s.pdb, s.userClient, &logger), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	test.SetupCreateAftermarketDevice(s.T(), "", nil, unitID, nil, s.pdb)
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

	s.deviceDefSvc.EXPECT().GetMakeByTokenID(gomock.Any(), gomock.Any()).Return(&ddgrpc.DeviceMake{Name: "AutoPi"}, nil)

	// act
	request := test.BuildRequest("GET", "/aftermarket/device/by-serial/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), false, gjson.GetBytes(body, "isUpdated").Bool())
	assert.Equal(s.T(), "1.23.1", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), false, gjson.GetBytes(body, "shouldUpdate").Bool())
}

func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_ShouldUpdate_Semver() {
	// as of jun 12 23, versions are now correctly semverd starting with "v", so test for this too
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID), owner.AftermarketDevice(s.pdb, s.userClient, &logger), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	test.SetupCreateAftermarketDevice(s.T(), "", nil, unitID, nil, s.pdb)
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		IsUpdated:         false,
		UnitID:            unitID,
		ID:                "4321",
		HwRevision:        "1.23",
		Template:          10,
		LastCommunication: time.Now(),
		Release: struct {
			Version string `json:"version"`
		}(struct{ Version string }{Version: "v1.22.8"}),
	}, nil)

	s.deviceDefSvc.EXPECT().GetMakeByTokenID(gomock.Any(), gomock.Any()).Return(&ddgrpc.DeviceMake{Name: "AutoPi"}, nil)

	// act
	request := test.BuildRequest("GET", "/aftermarket/device/by-serial/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), unitID, gjson.GetBytes(body, "unitId").String())
	assert.Equal(s.T(), "v1.22.8", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), false, gjson.GetBytes(body, "shouldUpdate").Bool()) // this because releaseVersion below 1.21.9
}

func (s *UserIntegrationsControllerTestSuite) TestPairAftermarketNoLegacy() {
	privateKey, err := crypto.GenerateKey()
	s.Require().NoError(err)

	kprod := smock.NewSyncProducer(s.T(), nil)

	var kb []byte

	kprod.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
		kb = val
		return nil
	})

	userAddr := crypto.PubkeyToAddress(privateKey.PublicKey)

	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", DIMORegistryChainID: 1337, DIMORegistryAddr: common.BigToAddress(big.NewInt(7)).Hex()}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, kprod, nil, nil, nil, nil, s.userClient, nil, nil, nil, nil, nil)
	s.deviceDefIntSvc.EXPECT().GetAutoPiIntegration(gomock.Any()).Return(&ddgrpc.Integration{Id: ksuid.New().String()}, nil).AnyTimes()

	userID := "louxUser"
	userAddrStr := userAddr.String()
	unitID := uuid.New().String()
	deviceID := uuid.New().String()

	s.userClient.EXPECT().GetUser(gomock.Any(), &pbuser.GetUserRequest{Id: userID}).Return(&pbuser.User{EthereumAddress: &userAddrStr}, nil).AnyTimes()

	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)
	ud := test.SetupCreateUserDevice(s.T(), userID, dd[0].DeviceDefinitionId, nil, "4Y1SL65848Z411439", s.pdb)

	mint := models.MetaTransactionRequest{ID: ksuid.New().String(), Status: models.MetaTransactionRequestStatusConfirmed}
	s.Require().NoError(mint.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()))

	vnft := models.VehicleNFT{
		UserDeviceID:  null.StringFrom(ud.ID),
		Vin:           ud.VinIdentifier.String,
		TokenID:       types.NewNullDecimal(decimal.New(4, 0)),
		OwnerAddress:  null.BytesFrom(userAddr.Bytes()),
		MintRequestID: mint.ID,
	}
	s.Require().NoError(vnft.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()))

	aftermarketDevice := test.SetupCreateAftermarketDevice(s.T(), testUserID, common.BigToAddress(big.NewInt(2)).Bytes(), unitID, &deviceID, s.pdb)
	aftermarketDevice.TokenID = types.NewDecimal(decimal.New(5, 0))
	aftermarketDevice.OwnerAddress = null.BytesFrom(userAddr.Bytes())
	row, errAMD := aftermarketDevice.Update(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Assert().Equal(int64(1), row)
	s.Require().NoError(errAMD)

	app := fiber.New()
	app.Use(test.AuthInjectorTestHandler(userID))
	app.Get("/:userDeviceID/pair", c.GetAutoPiPairMessage)
	app.Post("/:userDeviceID/pair", c.PostPairAutoPi)

	req := test.BuildRequest("GET", "/"+ud.ID+"/pair?external_id="+aftermarketDevice.Serial, "")

	res, err := app.Test(req)
	s.Require().NoError(err)
	s.Require().Equal(fiber.StatusOK, res.StatusCode) // todo issue - this is returning 409 instead of 200? due to change in how get unit?
	defer res.Body.Close()

	var td signer.TypedData
	s.Require().NoError(json.NewDecoder(res.Body).Decode(&td))

	b, _, err := signer.TypedDataAndHash(td)
	s.Require().NoError(err)

	userSig, err := crypto.Sign(b, privateKey)
	s.Require().NoError(err)
	userSig[64] += 27

	in := map[string]any{
		"externalId": aftermarketDevice.Serial,
		"signature":  hexutil.Bytes(userSig).String(),
	}

	inp, err := json.Marshal(in)
	s.Require().NoError(err)

	req = test.BuildRequest("POST", "/"+ud.ID+"/pair", string(inp))
	res, err = app.Test(req)
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Require().Equal(200, res.StatusCode)

	kprod.Close()

	var ce shared.CloudEvent[registry.RequestData]

	err = json.Unmarshal(kb, &ce)
	s.Require().NoError(err)

	abi, err := contracts.RegistryMetaData.GetAbi()
	s.Require().NoError(err)

	method := abi.Methods["pairAftermarketDeviceSign0"]

	callData := ce.Data.Data

	s.EqualValues(method.ID, callData[:4])

	o, err := method.Inputs.Unpack(callData[4:])
	s.Require().NoError(err)

	amID := o[0].(*big.Int)
	vID := o[1].(*big.Int)
	u2Sig := o[2].([]byte)

	s.Equal(big.NewInt(5), amID)
	s.Equal(big.NewInt(4), vID)
	s.Equal(userSig, u2Sig)
}

// Tesla Fleet API Tests
func (s *UserIntegrationsControllerTestSuite) TestPostTesla_V2() {
	integration := test.BuildIntegrationGRPC(constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)

	oV := &services.TeslaVehicle{}
	oUdai := &models.UserDeviceAPIIntegration{}

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *shared.CloudEvent[any]) error {
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

	s.teslaTaskService.EXPECT().StartPoll(gomock.AssignableToTypeOf(oV), gomock.AssignableToTypeOf(oUdai), 2, "na").DoAndReturn(
		func(v *services.TeslaVehicle, udai *models.UserDeviceAPIIntegration, _ int, _ string) error {
			oV = v
			oUdai = udai
			return nil
		},
	)

	s.teslaFleetAPISvc.EXPECT().GetVehicle(gomock.Any(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", "na", 1145).Return(&services.TeslaVehicle{
		ID:        1145,
		VehicleID: 223,
		VIN:       "5YJYGDEF9NF010423",
	}, nil)
	s.teslaFleetAPISvc.EXPECT().WakeUpVehicle(gomock.Any(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", "na", 1145).Return(nil)
	s.teslaFleetAPISvc.EXPECT().GetAvailableCommands().Return(&services.UserDeviceAPIIntegrationsMetadataCommands{
		Enabled:  []string{constants.DoorsUnlock, constants.DoorsLock, constants.TrunkOpen, constants.FrunkOpen, constants.ChargeLimit},
		Disabled: []string{constants.TelemetrySubscribe},
	})
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Times(2).Return(dd[0], nil)
	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dd[0], nil)

	userEthAddr := common.HexToAddress("1").String()
	s.userClient.EXPECT().GetUser(gomock.Any(), &pbuser.GetUserRequest{Id: testUserID}).Return(&pbuser.User{EthereumAddress: &userEthAddr}, nil).AnyTimes()

	expectedExpiry := time.Now().Add(10 * time.Minute)
	teslaResp := services.TeslaAuthCodeResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw",
		IDToken:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.ouLgsgz-xUWN7lLuo8qE2nueNgJIrBz49QLr_GLHRno",
		Expiry:       expectedExpiry,
		Region:       "na",
	}
	tokenStr, err := json.Marshal(teslaResp)
	s.Assert().NoError(err)

	encTeslaAuth, err := s.cipher.Encrypt(string(tokenStr))
	s.Assert().NoError(err)

	cacheKey := fmt.Sprintf(teslaFleetAuthCacheKey, userEthAddr)
	s.redisClient.EXPECT().Get(gomock.Any(), cacheKey).Return(redis.NewStringResult(encTeslaAuth, nil))
	s.redisClient.EXPECT().Del(gomock.Any(), cacheKey).Return(redis.NewIntResult(1, nil))

	in := `{
		"externalId": "1145",
		"version": 2
	}`
	request := test.BuildRequest("POST", fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), in)
	_, err = s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	intd, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom("1145"))).One(s.ctx, s.pdb.DBS().Reader)
	s.Assert().NoError(err)
	s.Assert().NotEmpty(intd.Metadata)

	encAccessToken, err := s.cipher.Encrypt(teslaResp.AccessToken)
	s.Assert().NoError(err)

	meta := &services.UserDeviceAPIIntegrationsMetadata{}
	err = intd.Metadata.Unmarshal(&meta)
	s.Assert().NoError(err)

	encRefreshToken, err := s.cipher.Encrypt(teslaResp.RefreshToken)
	s.Assert().NoError(err)
	s.Assert().Equal(1145, oV.ID)
	s.Assert().Equal(223, oV.VehicleID)
	s.Assert().Equal(null.StringFrom("1145"), intd.ExternalID)
	s.Assert().Equal(encAccessToken, intd.AccessToken.String)
	s.Assert().Equal(encRefreshToken, intd.RefreshToken.String)
	s.Assert().Equal(2, meta.TeslaAPIVersion)
}

func (s *UserIntegrationsControllerTestSuite) TestPostTesla_V2_PartialCredentials() {
	integration := test.BuildIntegrationGRPC(constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Return(dd[0], nil).AnyTimes()

	userEthAddr := common.HexToAddress("1").String()
	s.userClient.EXPECT().GetUser(gomock.Any(), &pbuser.GetUserRequest{Id: testUserID}).Return(&pbuser.User{EthereumAddress: &userEthAddr}, nil).AnyTimes()

	teslaResp := services.TeslaAuthCodeResponse{
		AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		IDToken:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.ouLgsgz-xUWN7lLuo8qE2nueNgJIrBz49QLr_GLHRno",
	}

	tokenStr, err := json.Marshal(teslaResp)
	s.Assert().NoError(err)

	encTeslaAuth, err := s.cipher.Encrypt(string(tokenStr))
	s.Assert().NoError(err)

	cacheKey := fmt.Sprintf(teslaFleetAuthCacheKey, userEthAddr)
	s.redisClient.EXPECT().Get(gomock.Any(), cacheKey).Return(redis.NewStringResult(encTeslaAuth, nil))

	in := `{
		"externalId": "1145",
		"version": 2
	}`
	request := test.BuildRequest("POST", fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), in)
	res, _ := s.app.Test(request, 60*1000)

	s.Assert().True(res.StatusCode == fiber.StatusBadRequest)
	body, _ := io.ReadAll(res.Body)

	defer res.Body.Close()

	_, err = models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom("1145"))).One(s.ctx, s.pdb.DBS().Reader)
	s.Assert().Equal(err.Error(), sql.ErrNoRows.Error())

	s.Assert().Equal("Couldn't retrieve stored credentials: missing tesla auth credentials", gjson.GetBytes(body, "message").String())
}

func (s *UserIntegrationsControllerTestSuite) TestPostTesla_V2_MissingCredentials() {
	integration := test.BuildIntegrationGRPC(constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Return(dd[0], nil).AnyTimes()

	userEthAddr := common.HexToAddress("1").String()
	s.userClient.EXPECT().GetUser(gomock.Any(), &pbuser.GetUserRequest{Id: testUserID}).Return(&pbuser.User{EthereumAddress: &userEthAddr}, nil).AnyTimes()

	cacheKey := fmt.Sprintf(teslaFleetAuthCacheKey, userEthAddr)
	s.redisClient.EXPECT().Get(gomock.Any(), cacheKey).Return(redis.NewStringResult("", nil))

	in := `{
		"externalId": "1145",
		"version": 2
	}`
	request := test.BuildRequest("POST", fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), in)
	res, _ := s.app.Test(request, 60*1000)

	s.Assert().True(res.StatusCode == fiber.StatusBadRequest)
	body, _ := io.ReadAll(res.Body)

	defer res.Body.Close()

	_, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom("1145"))).One(s.ctx, s.pdb.DBS().Reader)
	s.Assert().Equal(err.Error(), sql.ErrNoRows.Error())

	s.Assert().Equal("Couldn't retrieve stored credentials: no credential found", gjson.GetBytes(body, "message").String())
}

func (s *UserIntegrationsControllerTestSuite) TestGetUserDeviceIntegration() {
	integration := test.BuildIntegrationGRPC(constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model S", 2012, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "5YJSA1CN0CFP02439", s.pdb)

	accessTk := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	refreshTk := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw"
	extID := "SomeID"
	expectedExpiry := time.Now().Add(10 * time.Minute)
	region := "na"

	encAccessTk, err := s.cipher.Encrypt(accessTk)
	s.Require().NoError(err)
	encRefreshTk, err := s.cipher.Encrypt(refreshTk)
	s.Require().NoError(err)

	apIntd := models.UserDeviceAPIIntegration{
		UserDeviceID:    ud.ID,
		IntegrationID:   integration.Id,
		Status:          models.UserDeviceAPIIntegrationStatusActive,
		AccessToken:     null.StringFrom(encAccessTk),
		AccessExpiresAt: null.TimeFrom(expectedExpiry),
		RefreshToken:    null.StringFrom(encRefreshTk),
		ExternalID:      null.StringFrom(extID),
		Metadata:        null.JSONFrom([]byte(fmt.Sprintf(`{"teslaRegion":%q, "teslaApiVersion": 2}`, region))),
	}
	err = apIntd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)
	s.teslaFleetAPISvc.EXPECT().VirtualKeyConnectionStatus(gomock.Any(), accessTk, region, ud.VinIdentifier.String).Return(true, nil)

	request := test.BuildRequest(http.MethodGet, fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Require().Equal(res.StatusCode, fiber.StatusOK)
	body, _ := io.ReadAll(res.Body)

	defer res.Body.Close()

	actual := GetUserDeviceIntegrationResponse{}
	s.Require().NoError(json.Unmarshal(body, &actual))

	s.Assert().True(actual.Tesla.IsVirtualKeyConnected)
	s.Assert().Equal(models.UserDeviceAPIIntegrationStatusActive, actual.Status)
	s.Assert().Equal(extID, actual.ExternalID.String)
}

type deviceIntegrationCredentialsMatcher struct {
	accessToken  string
	refreshToken string
	expireAt     time.Time
}

func (m *deviceIntegrationCredentialsMatcher) String() string {
	return ""
}

func (m *deviceIntegrationCredentialsMatcher) Matches(x interface{}) bool {
	creds := x.(*models.UserDeviceAPIIntegration)

	if creds.AccessToken.String != m.accessToken {
		return false
	}

	if creds.RefreshToken.String != m.refreshToken {
		return false
	}

	if creds.AccessExpiresAt.Time != m.expireAt {
		return false
	}

	return true
}

func (s *UserIntegrationsControllerTestSuite) TestGetUserDeviceIntegration_RefreshToken() {
	integration := test.BuildIntegrationGRPC(constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model S", 2012, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "5YJSA1CN0CFP02439", s.pdb)

	accessTk := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	refreshTk := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw"
	extID := "SomeID"
	expectedExpiry := time.Now().Add(-10 * time.Minute)
	region := "na"

	newCredentials := &services.TeslaAuthCodeResponse{
		AccessToken:  accessTk,
		RefreshToken: refreshTk,
		Expiry:       time.Now().Add(10 * time.Hour),
		Region:       region,
	}

	expiredRefreshTk := "SomeExpRefTk"
	encExpiredAccessTk, err := s.cipher.Encrypt("SomeAccessTk")
	s.Require().NoError(err)
	encExpiredRefreshTk, err := s.cipher.Encrypt("SomeExpRefTk")
	s.Require().NoError(err)

	encAccessTk, err := s.cipher.Encrypt(newCredentials.AccessToken)
	s.Require().NoError(err)
	encRefreshTk, err := s.cipher.Encrypt(newCredentials.RefreshToken)
	s.Require().NoError(err)

	apIntd := models.UserDeviceAPIIntegration{
		UserDeviceID:    ud.ID,
		IntegrationID:   integration.Id,
		Status:          models.UserDeviceAPIIntegrationStatusActive,
		AccessToken:     null.StringFrom(encExpiredAccessTk),
		AccessExpiresAt: null.TimeFrom(expectedExpiry),
		RefreshToken:    null.StringFrom(encExpiredRefreshTk),
		ExternalID:      null.StringFrom(extID),
		Metadata:        null.JSONFrom([]byte(fmt.Sprintf(`{"teslaRegion":"%s", "teslaApiVersion":2}`, region))),
	}
	err = apIntd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)
	s.teslaFleetAPISvc.EXPECT().VirtualKeyConnectionStatus(gomock.Any(), accessTk, region, ud.VinIdentifier.String).Return(true, nil)

	s.teslaTaskService.EXPECT().UpdateCredentials(&deviceIntegrationCredentialsMatcher{
		accessToken:  encAccessTk,
		refreshToken: encRefreshTk,
		expireAt:     newCredentials.Expiry,
	}, constants.TeslaAPIV2, region).Return(nil)
	s.teslaFleetAPISvc.EXPECT().RefreshToken(gomock.Any(), expiredRefreshTk).Return(newCredentials, nil)

	request := test.BuildRequest(http.MethodGet, fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Assert().True(res.StatusCode == fiber.StatusOK)
	body, _ := io.ReadAll(res.Body)

	defer res.Body.Close()

	actual := GetUserDeviceIntegrationResponse{}
	s.Assert().NoError(json.Unmarshal(body, &actual))

	s.Assert().True(actual.Tesla.IsVirtualKeyConnected)
	s.Assert().Equal(models.UserDeviceAPIIntegrationStatusActive, actual.Status)
	s.Assert().Equal(extID, actual.ExternalID.String)

	newAPIInt, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(ud.ID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integration.Id),
	).One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)

	s.Assert().Equal(encRefreshTk, newAPIInt.RefreshToken.String)
	s.Assert().Equal(encAccessTk, newAPIInt.AccessToken.String)
}
