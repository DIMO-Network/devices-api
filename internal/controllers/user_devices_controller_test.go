package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"testing"

	"github.com/testcontainers/testcontainers-go"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeEventService struct{}

func (f *fakeEventService) Emit(event *shared.CloudEvent[any]) error {
	fmt.Printf("Emitting %v\n", event)
	return nil
}

type UserDevicesControllerTestSuite struct {
	suite.Suite
	pdb             db.Store
	controller      *UserDevicesController
	container       testcontainers.Container
	ctx             context.Context
	mockCtrl        *gomock.Controller
	app             *fiber.App
	deviceDefSvc    *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc *mock_services.MockDeviceDefinitionIntegrationService
	testUserID      string
	testUserEthAddr common.Address
	scTaskSvc       *mock_services.MockSmartcarTaskService
	scClient        *mock_services.MockSmartcarClient
	redisClient     *mocks.MockCacheService
	autoPiSvc       *mock_services.MockAutoPiAPIService
	natsService     *services.NATSService
	natsServer      *server.Server
	userDeviceSvc   *mock_services.MockUserDeviceService
}

const natsStreamName = "test-stream"

// SetupSuite starts container db
func (s *UserDevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	logger := test.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl
	var err error

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(mockCtrl)
	s.scClient = mock_services.NewMockSmartcarClient(mockCtrl)
	s.scTaskSvc = mock_services.NewMockSmartcarTaskService(mockCtrl)
	teslaTaskService := mock_services.NewMockTeslaTaskService(mockCtrl)
	autoPiIngest := mock_services.NewMockIngestRegistrar(mockCtrl)
	s.redisClient = mocks.NewMockCacheService(mockCtrl)
	s.autoPiSvc = mock_services.NewMockAutoPiAPIService(mockCtrl)
	s.natsService, s.natsServer, err = mock_services.NewMockNATSService(natsStreamName)
	s.userDeviceSvc = mock_services.NewMockUserDeviceService(mockCtrl)
	if err != nil {
		s.T().Fatal(err)
	}

	s.testUserID = "123123"
	testUserID2 := "3232451"
	s.testUserEthAddr = common.HexToAddress("0x1231231231231231231231231231231231231231")
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: "prod"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, teslaTaskService, nil, new(shared.ROT13Cipher), s.autoPiSvc,
		autoPiIngest, nil, nil, s.redisClient, nil, s.natsService, nil, s.userDeviceSvc, nil, nil, nil)
	app := test.SetupAppFiber(*logger)
	app.Post("/user/devices", test.AuthInjectorTestHandler(s.testUserID, nil), c.RegisterDeviceForUser)
	app.Post("/user/devices/fromvin", test.AuthInjectorTestHandler(s.testUserID, &s.testUserEthAddr), c.RegisterDeviceForUserFromVIN)
	app.Post("/user/devices/fromsmartcar", test.AuthInjectorTestHandler(s.testUserID, nil), c.RegisterDeviceForUserFromSmartcar)
	app.Post("/user/devices/second", test.AuthInjectorTestHandler(testUserID2, nil), c.RegisterDeviceForUser) // for different test user
	app.Get("/user/devices/me", test.AuthInjectorTestHandler(s.testUserID, &s.testUserEthAddr), c.GetUserDevices)
	app.Patch("/vehicle/:tokenID/vin", test.AuthInjectorTestHandler(s.testUserID, &s.testUserEthAddr), c.UpdateVINV2) // Auth done by the middleware.
	app.Post("/user/devices/:userDeviceID/commands/refresh", test.AuthInjectorTestHandler(s.testUserID, nil), c.RefreshUserDeviceStatus)
	app.Delete("/user/devices/:userDeviceID", test.AuthInjectorTestHandler(s.testUserID, nil), c.DeleteUserDevice)

	s.controller = &c
	s.app = app
}

func (s *UserDevicesControllerTestSuite) SetupTest() {
	s.controller.Settings.Environment = "prod"
}

// TearDownTest after each test truncate tables
func (s *UserDevicesControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserDevicesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	s.natsServer.Shutdown() // shuts down nats test server
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish() // might need to do mockctrl on every test, and refactor setup into one method
}

// Test Runner
func TestUserDevicesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(UserDevicesControllerTestSuite))
}

/* Actual Tests */
func (s *UserDevicesControllerTestSuite) TestPostUserDeviceFromVIN() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	// act request
	const deviceStyleID = "24GE7Mlc4c9o4j5P4mcD1Fzinx1"
	vinny := "4T3R6RFVXMU023395"
	canProtocol := "06"
	reg := RegisterUserDeviceVIN{VIN: vinny, CountryCode: "USA", CANProtocol: canProtocol}
	j, _ := json.Marshal(reg)

	s.deviceDefSvc.EXPECT().DecodeVIN(gomock.Any(), vinny, "", 0, reg.CountryCode).Times(1).Return(&ddgrpc.DecodeVinResponse{
		DeviceMakeId:  dd[0].Make.Id,
		DefinitionId:  dd[0].Id,
		DeviceStyleId: deviceStyleID,
		Year:          dd[0].Year,
	}, nil)

	apInteg := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 10)
	s.deviceDefIntSvc.EXPECT().GetAutoPiIntegration(gomock.Any()).Times(1).Return(apInteg, nil)
	s.userDeviceSvc.EXPECT().CreateUserDevice(gomock.Any(), dd[0].Id, deviceStyleID, "USA", s.testUserID, &vinny, &canProtocol, false).Times(1).
		Return(&models.UserDevice{
			ID:            ksuid.New().String(),
			UserID:        s.testUserID,
			DefinitionID:  dd[0].Id,
			VinIdentifier: null.StringFrom(vinny),
			CountryCode:   null.StringFrom("USA"),
			VinConfirmed:  true,
			Metadata:      null.JSONFrom([]byte(`{ "powertrainType": "ICE", "canProtocol": "6" }`)),
			DeviceStyleID: null.StringFrom(deviceStyleID),
		}, dd[0], nil)

	request := test.BuildRequest("POST", "/user/devices/fromvin", string(j))
	response, responseError := s.app.Test(request, 10000)
	require.NoError(s.T(), responseError)
	body, _ := io.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
	regUserResp := UserDeviceFull{}
	jsonUD := gjson.Get(string(body), "userDevice")
	_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

	assert.Len(s.T(), regUserResp.ID, 27)
	assert.Equal(s.T(), dd[0].Id, regUserResp.DeviceDefinition.DefinitionID)

	assert.Equal(s.T(), "USA", *regUserResp.CountryCode)
	assert.Equal(s.T(), vinny, *regUserResp.VIN)
	assert.Equal(s.T(), true, regUserResp.VINConfirmed)
	require.NotNil(s.T(), regUserResp.Metadata.CANProtocol)
	assert.Equal(s.T(), "6", *regUserResp.Metadata.CANProtocol)
	assert.EqualValues(s.T(), "ICE", *regUserResp.Metadata.PowertrainType)

	msg, responseError := s.natsService.JetStream.GetMsg(natsStreamName, 1)
	assert.NoError(s.T(), responseError, "expected no error from nats")
	vinResult := gjson.GetBytes(msg.Data, "vin")
	assert.Equal(s.T(), vinny, vinResult.Str)
}

func (s *UserDevicesControllerTestSuite) TestPostUserDeviceFromVIN_FailDecode() {
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	_ = test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)

	vinny := "4T3R6RFVXMU023395"
	canProtocol := "06"
	reg := RegisterUserDeviceVIN{VIN: vinny, CountryCode: "USA", CANProtocol: canProtocol}
	j, _ := json.Marshal(reg)

	grpcErr := status.Error(codes.InvalidArgument, "failed to decode vin")

	s.deviceDefSvc.EXPECT().DecodeVIN(gomock.Any(), vinny, "", 0, reg.CountryCode).Times(1).
		Return(nil, grpcErr)

	apInteg := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 10)
	s.deviceDefIntSvc.EXPECT().GetAutoPiIntegration(gomock.Any()).Times(1).Return(apInteg, nil)

	request := test.BuildRequest("POST", "/user/devices/fromvin", string(j))
	response, responseError := s.app.Test(request, 10000)
	require.NoError(s.T(), responseError)
	body, _ := io.ReadAll(response.Body)
	fmt.Println("resp body: " + string(body))
	// assert we get bad request and not 500
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	assert.Equal(s.T(), "failed to decode vin. unable to decode vin: 4T3R6RFVXMU023395", gjson.GetBytes(body, "message").String())
}

func (s *UserDevicesControllerTestSuite) TestPostUserDeviceFromVIN_SameUser_DuplicateVIN() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	// act request
	const vinny = "4T3R6RFVXMU023395"
	reg := RegisterUserDeviceVIN{VIN: vinny, CountryCode: "USA", CANProtocol: "06"}
	j, _ := json.Marshal(reg)
	// existing UserDevice with same VIN
	existingUD := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, vinny, s.pdb)
	// if the vin already exists for this user, do not expect decode request

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), dd[0].Id).Times(1).Return(dd[0], nil)
	apInteg := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 10)
	s.deviceDefIntSvc.EXPECT().GetAutoPiIntegration(gomock.Any()).Times(1).Return(apInteg, nil)

	request := test.BuildRequest("POST", "/user/devices/fromvin", string(j))
	response, responseError := s.app.Test(request, 10000)
	fmt.Println(responseError)
	body, _ := io.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
	regUserResp := UserDeviceFull{}
	jsonUD := gjson.Get(string(body), "userDevice")
	_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

	assert.Len(s.T(), regUserResp.ID, 27)
	assert.Equal(s.T(), existingUD.ID, regUserResp.ID, "expected to return existing user_device")
	assert.Equal(s.T(), dd[0].Id, regUserResp.DeviceDefinition.DefinitionID)

	msg, responseError := s.natsService.JetStream.GetMsg(natsStreamName, 1)
	assert.NoError(s.T(), responseError, "expected no error from nats")
	vinResult := gjson.GetBytes(msg.Data, "vin")
	assert.Equal(s.T(), vinny, vinResult.Str)

	userDevice, err := models.UserDevices().One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.NotNilf(s.T(), userDevice, "expected a user device in the database to exist")
	assert.Equal(s.T(), s.testUserID, userDevice.UserID)
	assert.Equal(s.T(), vinny, userDevice.VinIdentifier.String)
	// CAN Protocol to be updated on each request, assuming
}

func (s *UserDevicesControllerTestSuite) TestPostWithExistingDefinitionID() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	// act request
	reg := RegisterUserDevice{
		DefinitionID: dd[0].Id,
		CountryCode:  "USA",
	}
	j, _ := json.Marshal(reg)

	s.userDeviceSvc.EXPECT().CreateUserDevice(gomock.Any(), dd[0].Id, "", "USA", s.testUserID, nil, nil, false).Times(1).
		Return(&models.UserDevice{
			ID:           ksuid.New().String(),
			UserID:       testUserID,
			DefinitionID: dd[0].Id,
			CountryCode:  null.StringFrom("USA"),
			VinConfirmed: true,
			Metadata:     null.JSONFrom([]byte(`{ "powertrainType": "ICE" }`)),
		}, dd[0], nil)

	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, responseError := s.app.Test(request)
	fmt.Println(responseError)
	body, _ := io.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
	regUserResp := UserDeviceFull{}
	jsonUD := gjson.Get(string(body), "userDevice")
	_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

	assert.Len(s.T(), regUserResp.ID, 27)
	assert.Len(s.T(), regUserResp.DeviceDefinition.DeviceDefinitionID, 27)
	assert.Equal(s.T(), dd[0].Id, regUserResp.DeviceDefinition.DefinitionID)
}

func (s *UserDevicesControllerTestSuite) TestPostWithExistingDeviceDefinitionID() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	// act request
	reg := RegisterUserDevice{
		DefinitionID: dd[0].Id,
		CountryCode:  "USA",
	}
	j, _ := json.Marshal(reg)

	s.userDeviceSvc.EXPECT().CreateUserDevice(gomock.Any(), dd[0].Id, "", "USA", s.testUserID, nil, nil, false).Times(1).
		Return(&models.UserDevice{
			ID:           ksuid.New().String(),
			UserID:       testUserID,
			DefinitionID: dd[0].Id,
			CountryCode:  null.StringFrom("USA"),
			VinConfirmed: true,
			Metadata:     null.JSONFrom([]byte(`{ "powertrainType": "ICE" }`)),
		}, dd[0], nil)

	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, responseError := s.app.Test(request)
	fmt.Println(responseError)
	body, _ := io.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
	regUserResp := UserDeviceFull{}
	jsonUD := gjson.Get(string(body), "userDevice")
	_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

	assert.Len(s.T(), regUserResp.ID, 27)
	assert.Len(s.T(), regUserResp.DeviceDefinition.DeviceDefinitionID, 27)
	assert.Equal(s.T(), dd[0].Id, regUserResp.DeviceDefinition.DefinitionID)
}

func (s *UserDevicesControllerTestSuite) TestPostWithoutDefinitionID_BadRequest() {
	// act request
	reg := RegisterUserDevice{
		CountryCode: "USA",
	}
	j, _ := json.Marshal(reg)
	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, err := s.app.Test(request, 10*1000)
	require.NoError(s.T(), err)
	// assert
	require.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)

	body, err := io.ReadAll(response.Body)
	require.NoError(s.T(), err)

	errorMessage := gjson.Get(string(body), "message")
	if assert.Equal(s.T(), "definitionId is required", errorMessage.String()) == false {
		fmt.Println(string(body))
	}
}

func (s *UserDevicesControllerTestSuite) TestPostBadPayload() {
	request := test.BuildRequest("POST", "/user/devices", "{}")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	msg := gjson.Get(string(body), "message").String()
	assert.Contains(s.T(), msg, "cannot be blank")
}

func (s *UserDevicesControllerTestSuite) TestPostInvalidDefinitionID() {
	invalidDD := "caca"
	grpcErr := status.Error(codes.NotFound, "dd not found: "+invalidDD)
	s.userDeviceSvc.EXPECT().CreateUserDevice(gomock.Any(), invalidDD, "", "USA", s.testUserID, nil, nil, false).
		Return(nil, nil, grpcErr)
	reg := RegisterUserDevice{
		CountryCode:  "USA",
		DefinitionID: invalidDD,
	}
	j, _ := json.Marshal(reg)

	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)
	assert.Equal(s.T(), fiber.StatusNotFound, response.StatusCode)
	msg := gjson.Get(string(body), "message").String()
	fmt.Println("message: " + msg)
	assert.Contains(s.T(), msg, invalidDD)
}

func (s *UserDevicesControllerTestSuite) TestGetMyUserDevices() {
	// arrange db, insert some user_devices
	const (
		// Device 1
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
		deviceID = "device1"

		// Device 2
		userID2   = "user2"
		deviceID2 = "device2"
	)

	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd[0].Id, nil, "", s.pdb)
	_ = test.SetupCreateAftermarketDevice(s.T(), testUserID, nil, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	_ = test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integration.Id, s.pdb)

	ud2 := test.SetupCreateUserDeviceWithDeviceID(s.T(), userID2, deviceID2, dd[0].Id, nil, "", s.pdb)
	_ = test.SetupCreateVehicleNFT(s.T(), ud2, big.NewInt(1), null.BytesFrom(s.testUserEthAddr.Bytes()), s.pdb)

	s.deviceDefSvc.EXPECT().GetIntegrations(gomock.Any()).Return([]*ddgrpc.Integration{integration}, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), dd[0].Id).Times(2).Return(dd[0], nil)

	s.controller.Settings.Environment = "dev"
	request := test.BuildRequest("GET", "/user/devices/me", "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	result := gjson.Get(string(body), "userDevices.#.id")
	assert.Len(s.T(), result.Array(), 2)
	for _, id := range result.Array() {
		assert.True(s.T(), id.Exists(), "expected to find the ID")
	}

	assert.Equal(s.T(), integration.Id, gjson.GetBytes(body, "userDevices.1.integrations.0.integrationId").String())
	assert.Equal(s.T(), deviceID, gjson.GetBytes(body, "userDevices.1.integrations.0.externalId").String())
	assert.Equal(s.T(), integration.Vendor, gjson.GetBytes(body, "userDevices.1.integrations.0.integrationVendor").String())
	assert.Equal(s.T(), ud.ID, gjson.GetBytes(body, "userDevices.1.id").String())
	assert.Equal(s.T(), "device2                    ", gjson.GetBytes(body, "userDevices.0.id").String())
}

func (s *UserDevicesControllerTestSuite) TestGetMyUserDevicesNoDuplicates() {
	// arrange db, insert some user_devices
	const (
		// User
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
		deviceID = "device1                    "
		userID   = "userID"
	)
	s.controller.Settings.Environment = "dev"

	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	ud := test.SetupCreateUserDeviceWithDeviceID(s.T(), userID, deviceID, dd[0].Id, nil, "", s.pdb)
	_ = test.SetupCreateAftermarketDevice(s.T(), userID, nil, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	_ = test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integration.Id, s.pdb)

	_ = test.SetupCreateVehicleNFT(s.T(), ud, big.NewInt(1), null.BytesFrom(s.testUserEthAddr.Bytes()), s.pdb)

	s.deviceDefSvc.EXPECT().GetIntegrations(gomock.Any()).Return([]*ddgrpc.Integration{integration}, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), dd[0].Id).Times(1).Return(dd[0], nil)

	request := test.BuildRequest("GET", "/user/devices/me", "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	result := gjson.Get(string(body), "userDevices.#.id")
	assert.Len(s.T(), result.Array(), 1)

	for _, id := range result.Array() {
		assert.True(s.T(), id.Exists(), "expected to find the ID")
	}

	assert.Equal(s.T(), integration.Id, gjson.GetBytes(body, "userDevices.0.integrations.0.integrationId").String())
	assert.Equal(s.T(), integration.Vendor, gjson.GetBytes(body, "userDevices.0.integrations.0.integrationVendor").String())
	assert.Equal(s.T(), ud.ID, gjson.GetBytes(body, "userDevices.0.id").String())
}

func (s *UserDevicesControllerTestSuite) TestDeleteUserDevice_ErrNFTNotBurned() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	ud := models.UserDevice{
		ID:            ksuid.New().String(),
		UserID:        testUserID,
		DefinitionID:  "ford_escape_2020",
		CountryCode:   null.StringFrom("USA"),
		Name:          null.StringFrom("Chungus"),
		VinConfirmed:  true,
		VinIdentifier: null.StringFrom("4Y1SL65848Z411439"),
	}

	err = ud.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	test.SetupCreateVehicleNFT(s.T(), ud, big.NewInt(1), null.BytesFrom(addr.Bytes()), s.pdb)

	request := test.BuildRequest("DELETE", "/user/devices/"+ud.ID, "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)
	body, _ := io.ReadAll(response.Body)

	var resp map[string]interface{}
	err = json.Unmarshal(body, &resp)
	s.Require().NoError(err)

	s.Equal(fiber.StatusBadRequest, response.StatusCode)
}
