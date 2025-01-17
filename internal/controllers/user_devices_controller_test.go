package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"

	dagrpc "github.com/DIMO-Network/device-data-api/pkg/grpc"
	"github.com/ericlagergren/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats-server/v2/server"
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
	scTaskSvc       *mock_services.MockSmartcarTaskService
	scClient        *mock_services.MockSmartcarClient
	redisClient     *mocks.MockCacheService
	autoPiSvc       *mock_services.MockAutoPiAPIService
	usersClient     *mock_services.MockUserServiceClient
	natsService     *services.NATSService
	natsServer      *server.Server
	userDeviceSvc   *mock_services.MockUserDeviceService
	deviceDataSvc   *mock_services.MockDeviceDataService
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
	deviceDefinitionIngest := mock_services.NewMockDeviceDefinitionRegistrar(mockCtrl)
	s.redisClient = mocks.NewMockCacheService(mockCtrl)
	s.autoPiSvc = mock_services.NewMockAutoPiAPIService(mockCtrl)
	s.usersClient = mock_services.NewMockUserServiceClient(mockCtrl)
	s.natsService, s.natsServer, err = mock_services.NewMockNATSService(natsStreamName)
	s.userDeviceSvc = mock_services.NewMockUserDeviceService(mockCtrl)
	s.deviceDataSvc = mock_services.NewMockDeviceDataService(mockCtrl)
	if err != nil {
		s.T().Fatal(err)
	}

	s.testUserID = "123123"
	testUserID2 := "3232451"
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: "prod"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, teslaTaskService, new(shared.ROT13Cipher), s.autoPiSvc,
		autoPiIngest, deviceDefinitionIngest, nil, nil, s.redisClient, nil, s.usersClient, s.deviceDataSvc, s.natsService, nil, s.userDeviceSvc, nil, nil, nil)
	app := test.SetupAppFiber(*logger)
	app.Post("/user/devices", test.AuthInjectorTestHandler(s.testUserID), c.RegisterDeviceForUser)
	app.Post("/user/devices/fromvin", test.AuthInjectorTestHandler(s.testUserID), c.RegisterDeviceForUserFromVIN)
	app.Post("/user/devices/fromsmartcar", test.AuthInjectorTestHandler(s.testUserID), c.RegisterDeviceForUserFromSmartcar)
	app.Post("/user/devices/second", test.AuthInjectorTestHandler(testUserID2), c.RegisterDeviceForUser) // for different test user
	app.Get("/user/devices/me", test.AuthInjectorTestHandler(s.testUserID), c.GetUserDevices)
	app.Patch("/vehicle/:tokenID/vin", c.UpdateVINV2) // Auth done by the middleware.
	app.Post("/user/devices/:userDeviceID/commands/refresh", test.AuthInjectorTestHandler(s.testUserID), c.RefreshUserDeviceStatus)
	app.Get("/vehicle/:tokenID/commands/burn", test.AuthInjectorTestHandler(s.testUserID), c.GetBurnDevice)
	app.Post("/vehicle/:tokenID/commands/burn", test.AuthInjectorTestHandler(s.testUserID), c.PostBurnDevice)
	app.Delete("/user/devices/:userDeviceID", test.AuthInjectorTestHandler(s.testUserID), c.DeleteUserDevice)

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
func (s *UserDevicesControllerTestSuite) TestPostUserDeviceFromSmartcar() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(smartCarIntegrationID, constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	// act request
	vinny := "4T3R6RFVXMU023395"
	reg := RegisterUserDeviceSmartcar{Code: "XX", RedirectURI: "https://mobile-app", CountryCode: "USA"}
	j, _ := json.Marshal(reg)

	//ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, vinny, s.pdb)
	rot13 := new(shared.ROT13Cipher)

	scToken := smartcar.Token{
		Access:        "AA",
		AccessExpiry:  time.Now().Add(time.Hour),
		Refresh:       "RR",
		RefreshExpiry: time.Now().Add(time.Hour),
		ExpiresIn:     3600,
	}
	scTokenJSON, err := json.Marshal(scToken)
	require.NoError(s.T(), err)
	scTokenEnc, _ := rot13.Encrypt(string(scTokenJSON))

	s.scClient.EXPECT().ExchangeCode(gomock.Any(), reg.Code, reg.RedirectURI).Times(1).Return(&scToken, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), scToken.Access).Times(2).Return("123", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), scToken.Access, "123").Times(2).Return(vinny, nil)
	// called again below but with different response
	s.scClient.EXPECT().GetInfo(gomock.Any(), scToken.Access, "123").Times(1).Return(nil, nil)

	s.deviceDefSvc.EXPECT().DecodeVIN(gomock.Any(), vinny, "", 0, reg.CountryCode).Times(1).Return(&ddgrpc.DecodeVinResponse{
		DeviceMakeId:  dd[0].Make.Id,
		DefinitionId:  dd[0].Id,
		DeviceStyleId: "",
		Year:          dd[0].Year,
	}, nil)

	s.redisClient.EXPECT().Set(gomock.Any(), buildSmartcarTokenKey(vinny, testUserID), scTokenEnc, time.Hour*2).Return(nil)
	s.userDeviceSvc.EXPECT().CreateUserDevice(gomock.Any(), dd[0].Id, "", "USA", testUserID, &vinny, nil, false).
		Return(&models.UserDevice{
			ID:                 ksuid.New().String(),
			UserID:             testUserID,
			DeviceDefinitionID: dd[0].DeviceDefinitionId,
			VinIdentifier:      null.StringFrom(vinny),
			CountryCode:        null.StringFrom(reg.CountryCode),
			VinConfirmed:       false,
			DefinitionID:       dd[0].Id,
		}, dd[0], nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), smartCarIntegrationID).Return(integration, nil)
	// todo this one isn't getting called...
	//s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), dd[0].Id).Return(dd[0], nil)

	redisResponse := &redis.StringCmd{}
	redisResponse.SetVal(scTokenEnc)
	s.redisClient.EXPECT().Get(gomock.Any(), buildSmartcarTokenKey(vinny, testUserID)).Return(redisResponse)
	s.redisClient.EXPECT().Del(gomock.Any(), buildSmartcarTokenKey(vinny, testUserID)).Return(nil)

	s.scClient.EXPECT().GetUserID(gomock.Any(), scToken.Access).Return("123", nil)
	s.scClient.EXPECT().GetEndpoints(gomock.Any(), scToken.Access, "123").Return([]string{"https://smartcar.io/api"}, nil)
	s.scClient.EXPECT().HasDoorControl(gomock.Any(), scToken.Access, "123").Return(false, nil)
	s.scClient.EXPECT().GetInfo(gomock.Any(), scToken.Access, "123").Times(1).Return(&smartcar.Info{
		ID:              "1234567",
		Make:            "FORD",
		Model:           dd[0].Model,
		Year:            int(dd[0].Year),
		ResponseHeaders: smartcar.ResponseHeaders{},
	}, nil)
	encAccess, _ := rot13.Encrypt(scToken.Access)
	encRefresh, _ := rot13.Encrypt(scToken.Refresh)
	s.userDeviceSvc.EXPECT().CreateIntegration(gomock.Any(), gomock.Any(), gomock.Any(), integration.Id, "123", encAccess, gomock.Any(), encRefresh, gomock.Any()).Return(nil)

	request := test.BuildRequest("POST", "/user/devices/fromsmartcar", string(j))
	response, responseError := s.app.Test(request, 10000)
	fmt.Println(responseError)
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
	assert.Equal(s.T(), dd[0].DeviceDefinitionId, regUserResp.DeviceDefinition.DeviceDefinitionID)
	assert.Equal(s.T(), &vinny, regUserResp.VIN)
	// note: have removed any requirements on device definition integrations
}

func (s *UserDevicesControllerTestSuite) TestPostUserDeviceFromSmartcar_Fail_Decode() {
	// arrange DB
	_ = test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	// act request
	const vinny = "4T3R6RFVXMU023395"
	reg := RegisterUserDeviceSmartcar{Code: "XX", RedirectURI: "https://mobile-app", CountryCode: "USA"}
	j, _ := json.Marshal(reg)

	s.scClient.EXPECT().ExchangeCode(gomock.Any(), reg.Code, reg.RedirectURI).Times(1).Return(&smartcar.Token{
		Access:        "AA",
		AccessExpiry:  time.Now().Add(time.Hour),
		Refresh:       "RR",
		RefreshExpiry: time.Now().Add(time.Hour),
		ExpiresIn:     3600,
	}, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "AA").Times(1).Return("123", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "AA", "123").Times(1).Return(vinny, nil)
	s.redisClient.EXPECT().Set(gomock.Any(), buildSmartcarTokenKey(vinny, testUserID), gomock.Any(), time.Hour*2).Return(nil)
	s.scClient.EXPECT().GetInfo(gomock.Any(), "AA", "123").Times(1).Return(nil, nil)
	grpcErr := status.Error(codes.InvalidArgument, "failed to decode vin")
	s.deviceDefSvc.EXPECT().DecodeVIN(gomock.Any(), vinny, "", 0, reg.CountryCode).Times(1).Return(nil,
		grpcErr)

	request := test.BuildRequest("POST", "/user/devices/fromsmartcar", string(j))
	response, responseError := s.app.Test(request)
	fmt.Println(responseError)

	// assert we get bad request and not 500
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
}

func (s *UserDevicesControllerTestSuite) TestPostUserDeviceFromSmartcar_SameUser_DuplicateVIN() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)

	// act request
	const vinny = "4T3R6RFVXMU023395"
	reg := RegisterUserDeviceSmartcar{Code: "XX", RedirectURI: "https://mobile-app", CountryCode: "USA"}
	j, _ := json.Marshal(reg)
	test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, vinny, s.pdb)

	s.scClient.EXPECT().ExchangeCode(gomock.Any(), reg.Code, reg.RedirectURI).Times(1).Return(&smartcar.Token{
		Access:        "AA",
		AccessExpiry:  time.Now().Add(time.Hour),
		Refresh:       "RR",
		RefreshExpiry: time.Now().Add(time.Hour),
		ExpiresIn:     3600,
	}, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "AA").Times(1).Return("123", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "AA", "123").Times(1).Return(vinny, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)
	s.redisClient.EXPECT().Set(gomock.Any(), buildSmartcarTokenKey(vinny, testUserID), gomock.Any(), time.Hour*2).Return(nil)

	request := test.BuildRequest("POST", "/user/devices/fromsmartcar", string(j))
	response, responseError := s.app.Test(request)
	require.NoError(s.T(), responseError)
	body, _ := io.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusOK, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
}

func (s *UserDevicesControllerTestSuite) TestPostUserDeviceFromSmartcar_Fail_DuplicateVIN() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)

	// act request
	const vinny = "4T3R6RFVXMU023395"
	reg := RegisterUserDeviceSmartcar{Code: "XX", RedirectURI: "https://mobile-app", CountryCode: "USA"}
	j, _ := json.Marshal(reg)
	test.SetupCreateUserDevice(s.T(), "09098877", dd[0].DeviceDefinitionId, nil, vinny, s.pdb)

	s.scClient.EXPECT().ExchangeCode(gomock.Any(), reg.Code, reg.RedirectURI).Times(1).Return(&smartcar.Token{
		Access:        "AA",
		AccessExpiry:  time.Now().Add(time.Hour),
		Refresh:       "RR",
		RefreshExpiry: time.Now().Add(time.Hour),
		ExpiresIn:     3600,
	}, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "AA").Times(1).Return("123", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "AA", "123").Times(1).Return(vinny, nil)

	request := test.BuildRequest("POST", "/user/devices/fromsmartcar", string(j))
	response, responseError := s.app.Test(request)
	require.NoError(s.T(), responseError)
	body, _ := io.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusConflict, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
}

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
			ID:                 ksuid.New().String(),
			UserID:             s.testUserID,
			DeviceDefinitionID: dd[0].Ksuid,
			DefinitionID:       dd[0].Id,
			VinIdentifier:      null.StringFrom(vinny),
			CountryCode:        null.StringFrom("USA"),
			VinConfirmed:       true,
			Metadata:           null.JSONFrom([]byte(`{ "powertrainType": "ICE", "canProtocol": "6" }`)),
			DeviceStyleID:      null.StringFrom(deviceStyleID),
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
			ID:                 ksuid.New().String(),
			UserID:             testUserID,
			DeviceDefinitionID: dd[0].DeviceDefinitionId,
			DefinitionID:       dd[0].Id,
			CountryCode:        null.StringFrom("USA"),
			VinConfirmed:       true,
			Metadata:           null.JSONFrom([]byte(`{ "powertrainType": "ICE" }`)),
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
			ID:                 ksuid.New().String(),
			UserID:             testUserID,
			DeviceDefinitionID: dd[0].DeviceDefinitionId,
			DefinitionID:       dd[0].Id,
			CountryCode:        null.StringFrom("USA"),
			VinConfirmed:       true,
			Metadata:           null.JSONFrom([]byte(`{ "powertrainType": "ICE" }`)),
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

	addr := "67B94473D81D0cd00849D563C94d0432Ac988B49"
	ud2 := test.SetupCreateUserDeviceWithDeviceID(s.T(), userID2, deviceID2, dd[0].Id, nil, "", s.pdb)
	_ = test.SetupCreateVehicleNFT(s.T(), ud2, big.NewInt(1), null.BytesFrom(common.Hex2Bytes(addr)), s.pdb)

	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: s.testUserID}).Return(&pb.User{Id: s.testUserID, EthereumAddress: &addr}, nil)
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

	addr := "67B94473D81D0cd00849D563C94d0432Ac988B49"

	_ = test.SetupCreateVehicleNFT(s.T(), ud, big.NewInt(1), null.BytesFrom(common.Hex2Bytes(addr)), s.pdb)

	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: s.testUserID}).Return(&pb.User{Id: s.testUserID, EthereumAddress: &addr}, nil)
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

func (s *UserDevicesControllerTestSuite) TestPatchVIN() {
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 4)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Escape", 2020, integration)

	const powertrainType = "powertrain_type"
	powertrainValue := "BEV"
	for _, item := range dd[0].DeviceAttributes {
		if item.Name == powertrainType {
			item.Value = powertrainValue
			break
		}
	}

	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].Id, nil, "", s.pdb)
	ud.TokenID = types.NewNullDecimal(new(decimal.Big).SetUint64(40))
	_, err := ud.Update(context.TODO(), s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	s.deviceDefSvc.EXPECT().GetIntegrations(gomock.Any()).Return([]*ddgrpc.Integration{integration}, nil)

	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: s.testUserID}).Return(&pb.User{Id: s.testUserID, EthereumAddress: nil}, nil)
	// validates that if country=USA we update the powertrain based on what the NHTSA vin decoder says

	// seperate request to validate info persisted to user_device table
	//s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), []string{dd[0].Id}).Times(1).
	//	Return(dd[0], nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), dd[0].Id).Times(2).
		Return(dd[0], nil)

	payload := `{ "vin": "5YJYGDEE5MF085533" }`
	request := test.BuildRequest("PATCH", "/vehicle/40/vin", payload)
	response, responseError := s.app.Test(request)
	require.NoError(s.T(), responseError)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}

	request = test.BuildRequest("GET", "/user/devices/me", "")
	response, responseError = s.app.Test(request, 120*1000)
	require.NoError(s.T(), responseError)

	body, _ := io.ReadAll(response.Body)
	fmt.Println(string(body))
	pt := gjson.GetBytes(body, "userDevices.0.metadata.powertrainType").String()
	assert.Equal(s.T(), powertrainValue, pt)
}

func (s *UserDevicesControllerTestSuite) TestVINValidate() {

	type test struct {
		vin    string
		want   bool
		reason string
	}

	tests := []test{
		{vin: "5YJYGDEE5MF085533", want: true, reason: "valid vin number"},
		{vin: "5YJYGDEE5MF08553", want: false, reason: "too short"},
		{vin: "JA4AJ3AUXKU602608", want: true, reason: "valid vin number"},
		{vin: "2T1BU4EE2DC071057", want: true, reason: "valid vin number"},
		{vin: "5YJ3E1EA1NF156661", want: true, reason: "valid vin number"},
		{vin: "7AJ3E1EB3JF110865", want: true, reason: "valid vin number"},
		{vin: "", want: false, reason: "empty vin string"},
		{vin: "7FJ3E1EB3JF1108651234", want: false, reason: "vin string too long"},
	}

	for _, tc := range tests {
		vinReq := UpdateVINReq{VIN: tc.vin}
		err := vinReq.validate()
		if tc.want == true {
			assert.NoError(s.T(), err, tc.reason)
		} else {
			assert.Error(s.T(), err, tc.reason)
		}
	}
}

func (s *UserDevicesControllerTestSuite) TestPostRefreshSmartCar() {
	smartCarInt := test.BuildIntegrationGRPC(smartCarIntegrationID, constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Escape", 2020, smartCarInt)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), constants.SmartCarVendor).Return(smartCarInt, nil)
	s.deviceDataSvc.EXPECT().GetRawDeviceData(gomock.Any(), ud.ID, smartCarInt.Id).Return(&dagrpc.RawDeviceDataResponse{Items: []*dagrpc.RawDeviceDataResponseItem{
		{
			IntegrationId:   smartCarInt.Id,
			SignalsJsonData: []byte(`{"odometer": { "value": 123.223, "timestamp": "2022-06-18T04:06:40.200Z" } }`),
			RecordUpdatedAt: timestamppb.New(time.Now().UTC().Add(time.Hour * -4)),
			UserDeviceId:    ud.ID,
		},
	}}, nil)

	udiai := models.UserDeviceAPIIntegration{
		UserDeviceID:    ud.ID,
		IntegrationID:   smartCarInt.Id,
		Status:          models.UserDeviceAPIIntegrationStatusActive,
		AccessToken:     null.StringFrom("caca-token"),
		AccessExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(10) * time.Hour)),
		RefreshToken:    null.StringFrom("caca-refresh"),
		ExternalID:      null.StringFrom("caca-external-id"),
		TaskID:          null.StringFrom(ksuid.New().String()),
	}
	err := udiai.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)

	var oUdai *models.UserDeviceAPIIntegration

	// arrange mock
	s.scTaskSvc.EXPECT().Refresh(gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
		func(udai *models.UserDeviceAPIIntegration) error {
			oUdai = udai
			return nil
		},
	)

	payload := `{}`
	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/commands/refresh", payload)
	response, err := s.app.Test(request)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), ud.ID, oUdai.UserDeviceID)

	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body))
	}
}

func (s *UserDevicesControllerTestSuite) TestPostRefreshSmartCarRateLimited() {
	integration := test.BuildIntegrationGRPC(smartCarIntegrationID, constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mache", 2022, integration)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), constants.SmartCarVendor).Return(integration, nil)
	// arrange data to cause condition
	s.deviceDataSvc.EXPECT().GetRawDeviceData(gomock.Any(), ud.ID, integration.Id).Return(&dagrpc.RawDeviceDataResponse{Items: []*dagrpc.RawDeviceDataResponseItem{
		{
			IntegrationId:   integration.Id,
			RecordUpdatedAt: timestamppb.New(time.Now().UTC()),
			UserDeviceId:    ud.ID,
		},
	}}, nil)

	udiai := models.UserDeviceAPIIntegration{
		UserDeviceID:    ud.ID,
		IntegrationID:   integration.Id,
		Status:          models.UserDeviceAPIIntegrationStatusActive,
		AccessToken:     null.StringFrom("caca-token"),
		AccessExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(10) * time.Hour)),
		RefreshToken:    null.StringFrom("caca-refresh"),
		ExternalID:      null.StringFrom("caca-external-id"),
	}
	err := udiai.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)

	payload := `{}`
	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/commands/refresh", payload)
	response, _ := s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusTooManyRequests, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body))
	}
}

func (s *UserDevicesControllerTestSuite) TestDeleteUserDevice_ErrNFTNotBurned() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             testUserID,
		DeviceDefinitionID: ksuid.New().String(),
		DefinitionID:       "ford_escape_2020",
		CountryCode:        null.StringFrom("USA"),
		Name:               null.StringFrom("Chungus"),
		VinConfirmed:       true,
		VinIdentifier:      null.StringFrom("4Y1SL65848Z411439"),
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
