package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/DIMO-Network/shared/redis/mocks"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/rs/zerolog"

	pbuser "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/middleware/owner"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/services/tmpcred"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
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
	teslaTaskService          *mock_services.MockTeslaTaskService
	autopiAPISvc              *mock_services.MockAutoPiAPIService
	autoPiIngest              *mock_services.MockIngestRegistrar
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

// TODO(elffjs): This shouldn't be necessary anymore. Need to work with an interface.
const teslaFleetAuthCacheKey = "integration_credentials_%s"

// integration ID's - must be 27 chars to match ksuid length
const (
	smartCarIntegrationID = "22N2xaPOq2WW2gAHBHd0Ikn4Zob"
	teslaIntegrationID    = "tesla123aaaaaaaaaaaaaaaaaaa"
	autoPiIntegrationID   = "autopi123aaaaaaaaaaaaaaaaaa"
)

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
	s.teslaTaskService = mock_services.NewMockTeslaTaskService(s.mockCtrl)
	s.autopiAPISvc = mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	s.autoPiIngest = mock_services.NewMockIngestRegistrar(s.mockCtrl)
	s.deviceDefinitionRegistrar = mock_services.NewMockDeviceDefinitionRegistrar(s.mockCtrl)
	s.eventSvc = mock_services.NewMockEventService(s.mockCtrl)
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
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, s.eventSvc, s.scClient, s.scTaskSvc, s.teslaTaskService, s.cipher, s.autopiAPISvc,
		s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, s.redisClient, nil, s.userClient, nil, s.natsSvc, nil, s.userDeviceSvc,
		s.teslaFleetAPISvc, nil, nil)

	app := test.SetupAppFiber(*logger)

	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	app.Delete("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.DeleteUserDeviceIntegration)

	app.Post("/user2/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUser2), c.RegisterDeviceIntegration)
	app.Get("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.GetUserDeviceIntegration)
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/telemetry/subscribe",
		test.AuthInjectorTestHandler(testUserID),
		c.TelemetrySubscribe,
	)

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
	integration := test.BuildIntegrationGRPC(smartCarIntegrationID, constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mach E", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)

	req := `{
			"code": "qxyz",
			"redirectURI": "http://dimo.zone/cb"
		}`
	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxyz", "http://dimo.zone/cb").Times(1).Return(nil, errors.New("failure communicating with Smartcar"))
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), ud.DefinitionID).Times(1).Return(dd[0], nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Times(1).Return(integration, nil)

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
	integration := test.BuildIntegrationGRPC(smartCarIntegrationID, constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)
	vnft := test.SetupCreateVehicleNFT(s.T(), ud, big.NewInt(5), null.BytesFrom(common.HexToAddress("0xA1").Bytes()), s.pdb)

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
	integration := test.BuildIntegrationGRPC(smartCarIntegrationID, constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)

	const smartCarUserID = "smartCarUserId"
	req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
	expiry, _ := time.Parse(time.RFC3339, "2022-03-01T12:00:00Z")

	token := smartcar.Token{
		Access:        "myAccess",
		AccessExpiry:  expiry,
		Refresh:       "myRefresh",
		RefreshExpiry: expiry.Add(24 * time.Hour),
	}
	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxy", "http://dimo.zone/cb").Times(1).Return(&token, nil)

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *shared.CloudEvent[any]) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), dd[0].Id, data.Device.DefinitionID)
			assert.Equal(s.T(), dd[0].Make.Name, data.Device.Make)
			assert.Equal(s.T(), dd[0].Model, data.Device.Model)
			assert.Equal(s.T(), int(dd[0].Year), data.Device.Year)
			assert.Equal(s.T(), "CARVIN", data.Device.VIN)
			assert.Equal(s.T(), ud.ID, data.Device.ID)

			assert.Equal(s.T(), "SmartCar", data.Integration.Vendor)
			assert.Equal(s.T(), integration.Id, data.Integration.ID)
			return nil
		},
	)

	// original device def
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), ud.DefinitionID).Times(2).Return(dd[0], nil)
	s.scClient.EXPECT().GetUserID(gomock.Any(), "myAccess").Return(smartCarUserID, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "myAccess").Return("smartcar-idx", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "myAccess", "smartcar-idx").Return(vin, nil)
	s.scClient.EXPECT().GetEndpoints(gomock.Any(), "myAccess", "smartcar-idx").Return([]string{"/", "/vin"}, nil)
	s.scClient.EXPECT().HasDoorControl(gomock.Any(), "myAccess", "smartcar-idx").Return(false, nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)

	rot13 := new(shared.ROT13Cipher)
	encAccess, _ := rot13.Encrypt(token.Access)
	encRefresh, _ := rot13.Encrypt(token.Refresh)
	s.userDeviceSvc.EXPECT().CreateIntegration(gomock.Any(), gomock.Any(), gomock.Any(), integration.Id, "smartcar-idx",
		encAccess, gomock.Any(), encRefresh, gomock.Any()).Return(nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success") == false {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "unexpected response: "+string(body))
	}
	// no actual db record gets created anymore
	//apiInt, _ := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.Id)
	//updatedUD, _ := models.FindUserDevice(s.ctx, s.pdb.DBS().Reader, ud.ID)
	//
	//assert.Equal(s.T(), "zlNpprff", apiInt.AccessToken.String)
	//assert.True(s.T(), expiry.Equal(apiInt.AccessExpiresAt.Time))
	//assert.Equal(s.T(), "PendingFirstData", apiInt.Status)
	//assert.Equal(s.T(), "zlErserfu", apiInt.RefreshToken.String)
	//assert.Equal(s.T(), vin, updatedUD.VinIdentifier.String)
	//assert.Equal(s.T(), true, updatedUD.VinConfirmed)
}

func (s *UserIntegrationsControllerTestSuite) TestPostSmartCar_FailureTestVIN() {
	model := "Mach E"
	const vin = "0SC12312312312312"
	integration := test.BuildIntegrationGRPC(smartCarIntegrationID, constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)

	const smartCarUserID = "smartCarUserId"
	req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
	expiry, _ := time.Parse(time.RFC3339, "2022-03-01T12:00:00Z")
	//s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), dd[0].Id).Return(dd[0], nil)
	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxy", "http://dimo.zone/cb").Times(1).Return(&smartcar.Token{
		Access:        "myAccess",
		AccessExpiry:  expiry,
		Refresh:       "myRefresh",
		RefreshExpiry: expiry.Add(24 * time.Hour),
	}, nil)
	s.scClient.EXPECT().GetUserID(gomock.Any(), "myAccess").Return(smartCarUserID, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "myAccess").Return("smartcar-idx", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "myAccess", "smartcar-idx").Return(vin, nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)

	logger := test.Logger()
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: "prod"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, s.eventSvc, s.scClient, s.scTaskSvc, s.teslaTaskService, new(shared.ROT13Cipher), s.autopiAPISvc,
		s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, s.redisClient, nil, nil, nil, s.natsSvc, nil, s.userDeviceSvc, nil, nil, nil)

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
	integration := test.BuildIntegrationGRPC(smartCarIntegrationID, constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", model, 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)
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

			assert.Equal(s.T(), dd[0].Id, data.Device.DefinitionID)
			assert.Equal(s.T(), dd[0].Make.Name, data.Device.Make)
			assert.Equal(s.T(), dd[0].Model, data.Device.Model)
			assert.Equal(s.T(), int(dd[0].Year), data.Device.Year)
			assert.Equal(s.T(), "CARVIN", data.Device.VIN)
			assert.Equal(s.T(), ud.ID, data.Device.ID)

			assert.Equal(s.T(), "SmartCar", data.Integration.Vendor)
			assert.Equal(s.T(), integration.Id, data.Integration.ID)
			return nil
		},
	)

	s.scClient.EXPECT().GetUserID(gomock.Any(), token.Access).Return(smartCarUserID, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), token.Access).Return("smartcar-idx", nil)
	s.scClient.EXPECT().GetEndpoints(gomock.Any(), token.Access, "smartcar-idx").Return([]string{"/", "/vin"}, nil)
	s.scClient.EXPECT().HasDoorControl(gomock.Any(), token.Access, "smartcar-idx").Return(false, nil)

	// original device def
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), ud.DefinitionID).Times(1).Return(dd[0], nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)

	rot13 := new(shared.ROT13Cipher)
	encAccess, _ := rot13.Encrypt(token.Access)
	encRefresh, _ := rot13.Encrypt(token.Refresh)
	s.userDeviceSvc.EXPECT().CreateIntegration(gomock.Any(), gomock.Any(), gomock.Any(), integration.Id, "smartcar-idx",
		encAccess, gomock.Any(), encRefresh, gomock.Any()).Return(nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.Id, req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success") == false {
		body, _ := io.ReadAll(response.Body)
		assert.FailNow(s.T(), "unexpected response: "+string(body))
	}
	// no actual db record gets created anymore
	//apiInt, _ := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.Id)
	//
	//assert.Equal(s.T(), "fbzr-npprff-pbqr", apiInt.AccessToken.String)
	//assert.True(s.T(), expiry.Equal(apiInt.AccessExpiresAt.Time))
	//assert.Equal(s.T(), "PendingFirstData", apiInt.Status)
	//assert.Equal(s.T(), "fbzr-erserfu-pbqr", apiInt.RefreshToken.String)
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

func (s *UserIntegrationsControllerTestSuite) TestPostTeslaAndUpdateDD() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 20)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mach E", 2020, integration)

	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)

	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dd[0], nil)

	err := fixTeslaDeviceDefinition(s.ctx, test.Logger(), s.deviceDefSvc, s.pdb.DBS().Writer.DB, integration, &ud, "5YJRE1A31A1P01234")
	if err != nil {
		s.T().Fatalf("Got an error while fixing device definition: %v", err)
	}

	_ = ud.Reload(s.ctx, s.pdb.DBS().Writer.DB)
	// todo, we may need to point to new device def, or see how above fix method is implemented
	if ud.DefinitionID != dd[0].Id {
		s.T().Fatalf("Failed to switch device definition to the correct one")
	}
}

func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_ShouldUpdate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID), owner.AftermarketDevice(s.pdb, s.userClient, &logger), c.GetAftermarketDeviceInfo)
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
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID), owner.AftermarketDevice(s.pdb, s.userClient, &logger), c.GetAftermarketDeviceInfo)
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
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID), owner.AftermarketDevice(s.pdb, s.userClient, &logger), c.GetAftermarketDeviceInfo)
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
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID), owner.AftermarketDevice(s.pdb, s.userClient, &logger), c.GetAftermarketDeviceInfo)
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

// Tesla Fleet API Tests
func (s *UserIntegrationsControllerTestSuite) TestPostTesla_V2() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *shared.CloudEvent[any]) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), dd[0].Make.Name, data.Device.Make)
			assert.Equal(s.T(), dd[0].Model, data.Device.Model)
			assert.Equal(s.T(), int(dd[0].Year), data.Device.Year)
			assert.Equal(s.T(), "5YJYGDEF9NF010423", data.Device.VIN)
			assert.Equal(s.T(), ud.ID, data.Device.ID)

			assert.Equal(s.T(), constants.TeslaVendor, data.Integration.Vendor)
			assert.Equal(s.T(), integration.Id, data.Integration.ID)
			return nil
		},
	)

	s.teslaFleetAPISvc.EXPECT().GetVehicle(gomock.Any(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", 1145).Return(&services.TeslaVehicle{
		ID:        1145,
		VehicleID: 223,
		VIN:       "5YJYGDEF9NF010423",
	}, nil)
	s.teslaFleetAPISvc.EXPECT().WakeUpVehicle(gomock.Any(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", 1145).Return(nil)
	s.teslaFleetAPISvc.EXPECT().GetAvailableCommands(gomock.Any()).Return(&services.UserDeviceAPIIntegrationsMetadataCommands{
		Enabled:  []string{constants.DoorsUnlock, constants.DoorsLock, constants.TrunkOpen, constants.FrunkOpen, constants.ChargeLimit},
		Disabled: []string{constants.TelemetrySubscribe},
	}, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), ud.DefinitionID).Times(1).Return(dd[0], nil)
	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dd[0], nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Times(1).Return(integration, nil)

	userEthAddr := common.HexToAddress("1").String()
	s.userClient.EXPECT().GetUser(gomock.Any(), &pbuser.GetUserRequest{Id: testUserID}).Return(&pbuser.User{EthereumAddress: &userEthAddr}, nil).AnyTimes()

	expectedExpiry := time.Now().Add(10 * time.Minute)
	teslaResp := tmpcred.Credential{
		IntegrationID: 2,
		AccessToken:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		RefreshToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw",
		Expiry:        expectedExpiry,
	}
	tokenStr, err := json.Marshal(teslaResp)
	s.Assert().NoError(err)

	encTeslaAuth, err := s.cipher.Encrypt(string(tokenStr))
	s.Assert().NoError(err)

	cacheKey := fmt.Sprintf(teslaFleetAuthCacheKey, userEthAddr)
	s.redisClient.EXPECT().Get(gomock.Any(), cacheKey).Return(redis.NewStringResult(encTeslaAuth, nil))
	s.redisClient.EXPECT().Del(gomock.Any(), cacheKey).AnyTimes().Return(redis.NewIntResult(1, nil))

	in := `{
		"externalId": "1145",
		"version": 2
	}`
	request := test.BuildRequest("POST", fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), in)
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Equal(fiber.StatusNoContent, res.StatusCode)

	intd, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom("1145"))).One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().NotEmpty(intd.Metadata)

	encAccessToken, err := s.cipher.Encrypt(teslaResp.AccessToken)
	s.Assert().NoError(err)

	meta := &services.UserDeviceAPIIntegrationsMetadata{}
	err = intd.Metadata.Unmarshal(&meta)
	s.Assert().NoError(err)

	encRefreshToken, err := s.cipher.Encrypt(teslaResp.RefreshToken)
	s.Assert().NoError(err)
	s.Assert().Equal(null.StringFrom("1145"), intd.ExternalID)
	s.Assert().Equal(encAccessToken, intd.AccessToken.String)
	s.Assert().Equal(encRefreshToken, intd.RefreshToken.String)
	s.Assert().Equal(2, meta.TeslaAPIVersion)
}

func (s *UserIntegrationsControllerTestSuite) TestPostTesla_V2_PartialCredentials() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), ud.DefinitionID).Return(dd[0], nil).AnyTimes()
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil).AnyTimes()

	userEthAddr := common.HexToAddress("1").String()
	s.userClient.EXPECT().GetUser(gomock.Any(), &pbuser.GetUserRequest{Id: testUserID}).Return(&pbuser.User{EthereumAddress: &userEthAddr}, nil).AnyTimes()

	teslaResp := services.TeslaAuthCodeResponse{
		AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
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

	s.Equal(fiber.StatusInternalServerError, res.StatusCode)

	_, err = models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom("1145"))).One(s.ctx, s.pdb.DBS().Reader)
	s.Assert().Equal(err.Error(), sql.ErrNoRows.Error())
}

func (s *UserIntegrationsControllerTestSuite) TestPostTesla_V2_MissingCredentials() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), ud.DefinitionID).Return(dd[0], nil).AnyTimes()
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil).AnyTimes()

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

	s.Assert().Equal(fiber.StatusInternalServerError, res.StatusCode)

	_, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom("1145"))).One(s.ctx, s.pdb.DBS().Reader)
	s.Assert().ErrorIs(err, sql.ErrNoRows)
}

func (s *UserIntegrationsControllerTestSuite) TestGetUserDeviceIntegration() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model S", 2012, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "5YJSA1CN0CFP02439", s.pdb)

	accessTk := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	refreshTk := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw"
	extID := 1337
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
		ExternalID:      null.StringFrom(strconv.Itoa(extID)),
		Metadata:        null.JSONFrom([]byte(fmt.Sprintf(`{"teslaRegion":%q, "teslaApiVersion": 2}`, region))),
	}
	err = apIntd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)
	s.teslaFleetAPISvc.EXPECT().VirtualKeyConnectionStatus(gomock.Any(), accessTk, ud.VinIdentifier.String).Return(true, nil)
	s.teslaFleetAPISvc.EXPECT().GetTelemetrySubscriptionStatus(gomock.Any(), accessTk, extID).Return(false, nil)

	request := test.BuildRequest(http.MethodGet, fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Require().Equal(res.StatusCode, fiber.StatusOK)
	body, _ := io.ReadAll(res.Body)

	defer res.Body.Close()

	actual := GetUserDeviceIntegrationResponse{}
	s.Require().NoError(json.Unmarshal(body, &actual))

	s.Assert().True(actual.Tesla.VirtualKeyAdded)
	s.Assert().False(actual.Tesla.TelemetrySubscribed)
	s.Assert().Equal(models.UserDeviceAPIIntegrationStatusActive, actual.Status)
	s.Assert().Equal(strconv.Itoa(extID), actual.ExternalID.String)
}

func (s *UserIntegrationsControllerTestSuite) TestTelemetrySubscribe() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model S", 2012, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "5YJSA1CN0CFP02439", s.pdb)

	accessTk := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	accessTkEnc, _ := s.cipher.Encrypt(accessTk)
	refreshTk := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw"
	extID := "SomeID"
	expectedExpiry := time.Now().Add(10 * time.Minute)

	mtd, err := json.Marshal(services.UserDeviceAPIIntegrationsMetadata{
		Commands: &services.UserDeviceAPIIntegrationsMetadataCommands{
			Enabled: []string{constants.TelemetrySubscribe},
		},
	})
	s.Require().NoError(err)

	apIntd := models.UserDeviceAPIIntegration{
		UserDeviceID:    ud.ID,
		IntegrationID:   integration.Id,
		Status:          models.UserDeviceAPIIntegrationStatusActive,
		AccessToken:     null.StringFrom(accessTkEnc),
		AccessExpiresAt: null.TimeFrom(expectedExpiry),
		RefreshToken:    null.StringFrom(refreshTk),
		ExternalID:      null.StringFrom(extID),
		Metadata:        null.JSONFrom(mtd),
	}
	err = apIntd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)
	s.teslaFleetAPISvc.EXPECT().SubscribeForTelemetryData(gomock.Any(), accessTk, ud.VinIdentifier.String).Return(nil)

	request := test.BuildRequest(http.MethodPost, fmt.Sprintf("/user/devices/%s/integrations/%s/commands/telemetry/subscribe", ud.ID, integration.Id), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Assert().True(res.StatusCode == fiber.StatusOK)

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integration.Id),
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(ud.ID),
	).One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)

	md := new(services.UserDeviceAPIIntegrationsMetadata)
	err = udai.Metadata.Unmarshal(md)
	s.Require().NoError(err)

	s.T().Log(md.Commands.Enabled, "-0------")
	s.Assert().Equal(md.Commands.Enabled, []string{constants.TelemetrySubscribe})
}

func (s *UserIntegrationsControllerTestSuite) Test_NoUserDevice_TelemetrySubscribe() {
	request := test.BuildRequest(http.MethodPost, fmt.Sprintf("/user/devices/%s/integrations/%s/commands/telemetry/subscribe", "mockUserDeviceID", "mockIntID"), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Assert().True(res.StatusCode == fiber.StatusNotFound)
}

func (s *UserIntegrationsControllerTestSuite) Test_InactiveIntegration_TelemetrySubscribe() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model S", 2012, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "5YJSA1CN0CFP02439", s.pdb)

	apIntd := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integration.Id,
		Status:        models.DeviceCommandRequestStatusPending,
	}
	err := apIntd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	request := test.BuildRequest(http.MethodPost, fmt.Sprintf("/user/devices/%s/integrations/%s/commands/telemetry/subscribe", ud.ID, integration.Id), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Equal(fiber.StatusBadRequest, res.StatusCode)
}

func (s *UserIntegrationsControllerTestSuite) Test_MissingRegionAndCapable_TelemetrySubscribe() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model S", 2012, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "5YJSA1CN0CFP02439", s.pdb)

	apIntd := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integration.Id,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
	}
	err := apIntd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	request := test.BuildRequest(http.MethodPost, fmt.Sprintf("/user/devices/%s/integrations/%s/commands/telemetry/subscribe", ud.ID, integration.Id), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Assert().True(res.StatusCode == fiber.StatusBadRequest)
}

func (s *UserIntegrationsControllerTestSuite) Test_TelemetrySubscribe_NotCapable() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model S", 2012, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "5YJSA1CN0CFP02439", s.pdb)

	mtd, err := json.Marshal(services.UserDeviceAPIIntegrationsMetadata{})
	s.Require().NoError(err)
	apIntd := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integration.Id,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
		Metadata:      null.JSONFrom(mtd),
	}
	err = apIntd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	request := test.BuildRequest(http.MethodPost, fmt.Sprintf("/user/devices/%s/integrations/%s/commands/telemetry/subscribe", ud.ID, integration.Id), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Assert().True(res.StatusCode == fiber.StatusBadRequest)
}
