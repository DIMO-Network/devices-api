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
	"github.com/segmentio/ksuid"
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
	natsSvc                   *services.NATSService
	natsServer                *server.Server
	userDeviceSvc             *mock_services.MockUserDeviceService
	cipher                    shared.Cipher
	teslaFleetAPISvc          *mock_services.MockTeslaFleetAPIService
	user1EthAddr              common.Address
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
	s.natsSvc, s.natsServer, err = mock_services.NewMockNATSService(natsStreamName)
	s.userDeviceSvc = mock_services.NewMockUserDeviceService(s.mockCtrl)
	s.teslaFleetAPISvc = mock_services.NewMockTeslaFleetAPIService(s.mockCtrl)
	s.cipher = new(shared.ROT13Cipher)
	s.user1EthAddr = common.HexToAddress("1")

	if err != nil {
		s.T().Fatal(err)
	}

	logger := test.Logger()
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, s.eventSvc, s.scClient, s.scTaskSvc, s.teslaTaskService, nil, s.cipher, s.autopiAPISvc,
		s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, s.redisClient, nil, s.natsSvc, nil, s.userDeviceSvc,
		s.teslaFleetAPISvc, nil, nil)

	app := test.SetupAppFiber(*logger)

	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID, &s.user1EthAddr), c.RegisterDeviceIntegration)
	app.Delete("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID, &s.user1EthAddr), c.DeleteUserDeviceIntegration)

	app.Post("/user2/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUser2, nil), c.RegisterDeviceIntegration)
	app.Get("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID, &s.user1EthAddr), c.GetUserDeviceIntegration)
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/telemetry/subscribe",
		test.AuthInjectorTestHandler(testUserID, nil),
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

	err := fixTeslaDeviceDefinition(s.ctx, test.Logger(), s.pdb.DBS().Writer.DB, integration, &ud, "5YJRE1A31A1P01234")
	if err != nil {
		s.T().Fatalf("Got an error while fixing device definition: %v", err)
	}

	_ = ud.Reload(s.ctx, s.pdb.DBS().Writer.DB)
	if ud.DefinitionID != "tesla_roadster_2010" { // based on the above VIN decoding
		s.T().Fatalf("Failed to switch device definition to the correct one")
	}
}

func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_ShouldUpdate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaTaskService, nil, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID, nil), owner.AftermarketDevice(s.pdb, &logger), c.GetAftermarketDeviceInfo)
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
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaTaskService, nil, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID, nil), owner.AftermarketDevice(s.pdb, &logger), c.GetAftermarketDeviceInfo)
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
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaTaskService, nil, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID, nil), owner.AftermarketDevice(s.pdb, &logger), c.GetAftermarketDeviceInfo)
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
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaTaskService, nil, new(shared.ROT13Cipher), autopiAPISvc, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	app := fiber.New()
	logger := zerolog.Nop()
	app.Get("/aftermarket/device/by-serial/:serial", test.AuthInjectorTestHandler(testUserID, nil), owner.AftermarketDevice(s.pdb, &logger), c.GetAftermarketDeviceInfo)
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
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2022, integration)
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
	s.teslaFleetAPISvc.EXPECT().VirtualKeyConnectionStatus(gomock.Any(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", "5YJYGDEF9NF010423").Return(&services.VehicleFleetStatus{
		DiscountedDeviceData: false,
	}, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), ud.DefinitionID).Times(1).Return(dd[0], nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Times(1).Return(integration, nil)

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

	cacheKey := fmt.Sprintf(teslaFleetAuthCacheKey, s.user1EthAddr.Hex())
	s.redisClient.EXPECT().Get(gomock.Any(), cacheKey).Return(redis.NewStringResult(encTeslaAuth, nil))
	s.redisClient.EXPECT().Del(gomock.Any(), cacheKey).AnyTimes().Return(redis.NewIntResult(1, nil))

	in := `{
		"externalId": "1145",
		"version": 2
	}`
	request := test.BuildRequest("POST", fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), in)
	res, err := s.app.Test(request, 120*1000)
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
	s.Require().NotNil(meta.TeslaDiscountedData)
	s.Assert().False(*meta.TeslaDiscountedData)
}

func (s *UserIntegrationsControllerTestSuite) TestPostTesla_V2_PartialCredentials() {
	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), ud.DefinitionID).Return(dd[0], nil).AnyTimes()
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil).AnyTimes()

	userEthAddr := common.HexToAddress("1").String()

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
	vin := "5YJSA1CN0CFP02439"

	integration := test.BuildIntegrationGRPC(teslaIntegrationID, constants.TeslaVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model S", 2012, integration)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, vin, s.pdb)

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
	s.teslaFleetAPISvc.EXPECT().GetTelemetrySubscriptionStatus(gomock.Any(), accessTk, vin).Return(&services.VehicleTelemetryStatus{}, nil)

	s.teslaFleetAPISvc.EXPECT().VirtualKeyConnectionStatus(gomock.Any(), accessTk, vin).Return(&services.VehicleFleetStatus{DiscountedDeviceData: true}, nil)

	request := test.BuildRequest(http.MethodGet, fmt.Sprintf("/user/devices/%s/integrations/%s", ud.ID, integration.Id), "")
	res, err := s.app.Test(request, 60*1000)
	s.Assert().NoError(err)

	s.Require().Equal(res.StatusCode, fiber.StatusOK)
	body, _ := io.ReadAll(res.Body)

	defer res.Body.Close()

	actual := GetUserDeviceIntegrationResponse{}
	s.Require().NoError(json.Unmarshal(body, &actual))

	s.Assert().False(actual.Tesla.VirtualKeyAdded)
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

func TestTeslaFirmewareCheck(t *testing.T) {
	cases := []struct {
		vs      string
		capable bool
	}{
		{"2025.8", true},
		{"2025.2.8", true},
		{"2025.2.6.2", true},
		{"2024.44.4", true},
		{"2024.26", true},
		{"2024.20.6", false},
		{"2024.14.200.1", false},
	}

	for _, tc := range cases {
		if b, _ := IsFirmwareFleetTelemetryCapable(tc.vs); b != tc.capable {
			if tc.capable {
				t.Errorf("expected %q to be capable", tc.vs)
			} else {
				t.Errorf("expected %q to be incapable", tc.vs)
			}
		}
	}
}
