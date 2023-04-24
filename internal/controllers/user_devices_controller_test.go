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

	"github.com/google/uuid"

	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/redis/mocks"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeEventService struct{}

func (f *fakeEventService) Emit(event *services.Event) error {
	fmt.Printf("Emitting %v\n", event)
	return nil
}

type UserDevicesControllerTestSuite struct {
	suite.Suite
	pdb             db.Store
	container       testcontainers.Container
	ctx             context.Context
	mockCtrl        *gomock.Controller
	app             *fiber.App
	deviceDefSvc    *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc *mock_services.MockDeviceDefinitionIntegrationService
	testUserID      string
	scTaskSvc       *mock_services.MockSmartcarTaskService
	nhtsaService    *mock_services.MockINHTSAService
	drivlyTaskSvc   *mock_services.MockDrivlyTaskService
	scClient        *mock_services.MockSmartcarClient
	redisClient     *mocks.MockCacheService
	autoPiSvc       *mock_services.MockAutoPiAPIService
	usersClient     *mock_services.MockUserServiceClient
}

// SetupSuite starts container db
func (s *UserDevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	logger := test.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(mockCtrl)
	s.scClient = mock_services.NewMockSmartcarClient(mockCtrl)
	s.scTaskSvc = mock_services.NewMockSmartcarTaskService(mockCtrl)
	teslaSvc := mock_services.NewMockTeslaService(mockCtrl)
	teslaTaskService := mock_services.NewMockTeslaTaskService(mockCtrl)
	s.nhtsaService = mock_services.NewMockINHTSAService(mockCtrl)
	autoPiIngest := mock_services.NewMockIngestRegistrar(mockCtrl)
	deviceDefinitionIngest := mock_services.NewMockDeviceDefinitionRegistrar(mockCtrl)
	autoPiTaskSvc := mock_services.NewMockAutoPiTaskService(mockCtrl)
	s.redisClient = mocks.NewMockCacheService(mockCtrl)
	s.autoPiSvc = mock_services.NewMockAutoPiAPIService(mockCtrl)
	s.usersClient = mock_services.NewMockUserServiceClient(mockCtrl)

	s.testUserID = "123123"
	testUserID2 := "3232451"
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: "prod"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc, &fakeEventService{}, s.scClient, s.scTaskSvc, teslaSvc, teslaTaskService, new(shared.ROT13Cipher), s.autoPiSvc,
		s.nhtsaService, autoPiIngest, deviceDefinitionIngest, autoPiTaskSvc, nil, nil, s.drivlyTaskSvc, nil, s.redisClient, nil, s.usersClient)
	app := test.SetupAppFiber(*logger)
	app.Post("/user/devices", test.AuthInjectorTestHandler(s.testUserID), c.RegisterDeviceForUser)
	app.Post("/user/devices/fromvin", test.AuthInjectorTestHandler(s.testUserID), c.RegisterDeviceForUserFromVIN)
	app.Post("/user/devices/fromsmartcar", test.AuthInjectorTestHandler(s.testUserID), c.RegisterDeviceForUserFromSmartcar)
	app.Post("/user/devices/second", test.AuthInjectorTestHandler(testUserID2), c.RegisterDeviceForUser) // for different test user
	app.Get("/user/devices/me", test.AuthInjectorTestHandler(s.testUserID), c.GetUserDevices)
	app.Patch("/user/devices/:userDeviceID/vin", test.AuthInjectorTestHandler(s.testUserID), c.UpdateVIN)
	app.Patch("/user/devices/:userDeviceID/name", test.AuthInjectorTestHandler(s.testUserID), c.UpdateName)
	app.Patch("/user/devices/:userDeviceID/image", test.AuthInjectorTestHandler(s.testUserID), c.UpdateImage)
	app.Get("/user/devices/:userDeviceID/offers", test.AuthInjectorTestHandler(s.testUserID), c.GetOffers)
	app.Get("/user/devices/:userDeviceID/valuations", test.AuthInjectorTestHandler(s.testUserID), c.GetValuations)
	app.Get("/user/devices/:userDeviceID/range", test.AuthInjectorTestHandler(s.testUserID), c.GetRange)
	app.Post("/user/devices/:userDeviceID/commands/refresh", test.AuthInjectorTestHandler(s.testUserID), c.RefreshUserDeviceStatus)

	s.app = app
}

// TearDownTest after each test truncate tables
func (s *UserDevicesControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserDevicesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
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
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
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

	s.deviceDefSvc.EXPECT().DecodeVIN(gomock.Any(), vinny).Times(1).Return(&grpc.DecodeVinResponse{
		DeviceMakeId:       dd[0].Make.Id,
		DeviceDefinitionId: dd[0].DeviceDefinitionId,
		DeviceStyleId:      "",
		Year:               dd[0].Type.Year,
	}, nil)
	s.deviceDefIntSvc.EXPECT().CreateDeviceDefinitionIntegration(gomock.Any(), "22N2xaPOq2WW2gAHBHd0Ikn4Zob", dd[0].DeviceDefinitionId, "Americas").Times(1).
		Return(nil, nil)
	s.redisClient.EXPECT().Set(gomock.Any(), buildSmartcarTokenKey(vinny, testUserID), gomock.Any(), time.Hour*2).Return(nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)
	request := test.BuildRequest("POST", "/user/devices/fromsmartcar", string(j))
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
	assert.Equal(s.T(), dd[0].DeviceDefinitionId, regUserResp.DeviceDefinition.DeviceDefinitionID)
	if assert.Len(s.T(), regUserResp.DeviceDefinition.CompatibleIntegrations, 2) == false {
		fmt.Println("resp body: " + string(body))
	}
	assert.Equal(s.T(), integration.Vendor, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Vendor)
	assert.Equal(s.T(), integration.Type, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Type)
	assert.Equal(s.T(), integration.Id, regUserResp.DeviceDefinition.CompatibleIntegrations[0].ID)

	userDevice, err := models.UserDevices().One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.NotNilf(s.T(), userDevice, "expected a user device in the database to exist")
	assert.Equal(s.T(), s.testUserID, userDevice.UserID)
	assert.Equal(s.T(), vinny, userDevice.VinIdentifier.String)
}

func (s *UserDevicesControllerTestSuite) TestPostUserDeviceFromSmartcar_SameUser_DuplicateVIN() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
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
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)
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
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
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
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	// act request
	const vinny = "4T3R6RFVXMU023395"
	reg := RegisterUserDeviceVIN{VIN: vinny, CountryCode: "USA", CANProtocol: "06"}
	j, _ := json.Marshal(reg)

	s.deviceDefSvc.EXPECT().DecodeVIN(gomock.Any(), vinny).Times(1).Return(&grpc.DecodeVinResponse{
		DeviceMakeId:       dd[0].Make.Id,
		DeviceDefinitionId: dd[0].DeviceDefinitionId,
		DeviceStyleId:      "",
		Year:               dd[0].Type.Year,
	}, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)
	apInteg := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 10)
	s.deviceDefIntSvc.EXPECT().GetAutoPiIntegration(gomock.Any()).Times(1).Return(apInteg, nil)
	s.deviceDefIntSvc.EXPECT().CreateDeviceDefinitionIntegration(gomock.Any(), apInteg.Id, dd[0].DeviceDefinitionId, "Americas")
	request := test.BuildRequest("POST", "/user/devices/fromvin", string(j))
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
	assert.Equal(s.T(), dd[0].DeviceDefinitionId, regUserResp.DeviceDefinition.DeviceDefinitionID)
	if assert.Len(s.T(), regUserResp.DeviceDefinition.CompatibleIntegrations, 2) == false {
		fmt.Println("resp body: " + string(body))
	}
	assert.Equal(s.T(), integration.Vendor, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Vendor)
	assert.Equal(s.T(), integration.Type, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Type)
	assert.Equal(s.T(), integration.Id, regUserResp.DeviceDefinition.CompatibleIntegrations[0].ID)

	userDevice, err := models.UserDevices().One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.NotNilf(s.T(), userDevice, "expected a user device in the database to exist")
	assert.Equal(s.T(), s.testUserID, userDevice.UserID)
	assert.Equal(s.T(), vinny, userDevice.VinIdentifier.String)
	assert.Equal(s.T(), "06", gjson.GetBytes(userDevice.Metadata.JSON, "canProtocol").Str)
}

func (s *UserDevicesControllerTestSuite) TestPostWithExistingDefinitionID() {
	// arrange DB
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	// act request
	reg := RegisterUserDevice{
		DeviceDefinitionID: &dd[0].DeviceDefinitionId,
		CountryCode:        "USA",
	}
	j, _ := json.Marshal(reg)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)
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
	assert.Equal(s.T(), dd[0].DeviceDefinitionId, regUserResp.DeviceDefinition.DeviceDefinitionID)
	if assert.Len(s.T(), regUserResp.DeviceDefinition.CompatibleIntegrations, 2) == false {
		fmt.Println("resp body: " + string(body))
	}
	assert.Equal(s.T(), integration.Vendor, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Vendor)
	assert.Equal(s.T(), integration.Type, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Type)
	assert.Equal(s.T(), integration.Id, regUserResp.DeviceDefinition.CompatibleIntegrations[0].ID)

	userDevice, err := models.UserDevices().One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.NotNilf(s.T(), userDevice, "expected a user device in the database to exist")
	assert.Equal(s.T(), s.testUserID, userDevice.UserID)
	assert.Nil(s.T(), userDevice.VinIdentifier.Ptr())
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
	if assert.Equal(s.T(), "deviceDefinitionId: cannot be blank.", errorMessage.String()) == false {
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
	grpcErr := status.Error(codes.NotFound, "dd not found")
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), invalidDD).Times(1).Return(nil, grpcErr)
	reg := RegisterUserDevice{
		DeviceDefinitionID: &invalidDD,
		CountryCode:        "USA",
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

	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	_ = test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integration.Id, s.pdb)

	addr := "67B94473D81D0cd00849D563C94d0432Ac988B49"
	_ = test.SetupCreateUserDeviceWithDeviceID(s.T(), userID2, deviceID2, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	_ = test.SetupCreateVehicleNFT(s.T(), deviceID2, "vin", big.NewInt(1), null.BytesFrom(common.Hex2Bytes(addr)), s.pdb)

	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: s.testUserID}).Return(&pb.User{Id: s.testUserID, EthereumAddress: &addr}, nil)
	s.deviceDefSvc.EXPECT().GetIntegrations(gomock.Any()).Return([]*grpc.Integration{integration}, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd[0].DeviceDefinitionId, dd[0].DeviceDefinitionId}).Times(1).Return(dd, nil)

	request := test.BuildRequest("GET", "/user/devices/me", "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	s.T().Log(string(body))
	result := gjson.Get(string(body), "userDevices.#.id")
	assert.Len(s.T(), result.Array(), 2)
	for n, id := range result.Array() {
		path := fmt.Sprintf("userDevices.%d.", n)
		assert.True(s.T(), id.Exists(), "expected to find the ID")
		if id.String() == s.testUserID {
			assert.Equal(s.T(), ud.ID, id.String(), "expected user device ID to match")
			assert.Equal(s.T(), integration.Id, gjson.GetBytes(body, path+"integrations.0.integrationId").String())
			assert.Equal(s.T(), "device123", gjson.GetBytes(body, path+"integrations.0.externalId").String())
			assert.Equal(s.T(), integration.Vendor, gjson.GetBytes(body, path+"integrations.0.integrationVendor").String())
		}
		if id.String() == userID2 {
			assert.Equal(s.T(), "device2                    ", gjson.GetBytes(body, path+"id").String())
		}
	}
}

func (s *UserDevicesControllerTestSuite) TestPatchVIN() {
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 4)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Escape", 2020, integration)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	s.deviceDefSvc.EXPECT().GetIntegrations(gomock.Any()).Return([]*grpc.Integration{integration}, nil)

	evID := "4"
	s.nhtsaService.EXPECT().DecodeVIN("5YJYGDEE5MF085533").Return(&services.NHTSADecodeVINResponse{
		Results: []services.NHTSAResult{
			{
				VariableID: 126,
				ValueID:    &evID,
			},
		},
	}, nil)
	payload := `{ "vin": "5YJYGDEE5MF085533" }`
	request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/vin", payload)
	response, responseError := s.app.Test(request)
	require.NoError(s.T(), responseError)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd[0].DeviceDefinitionId}).Times(1).Return(dd, nil)

	request = test.BuildRequest("GET", "/user/devices/me", "")
	response, responseError = s.app.Test(request)
	require.NoError(s.T(), responseError)

	body, _ := io.ReadAll(response.Body)
	fmt.Println(string(body))
	pt := gjson.GetBytes(body, "userDevices.0.metadata.powertrainType").String()
	assert.Equal(s.T(), "BEV", pt)
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

func (s *UserDevicesControllerTestSuite) TestNameValidate() {

	type test struct {
		name   string
		want   bool
		reason string
	}

	tests := []test{
		{name: "ValidNameHere", want: true, reason: "valid name"},
		{name: "MyCar2022", want: true, reason: "valid name"},
		{name: "16CharactersLong", want: true, reason: "valid name"},
		{name: "12345", want: true, reason: "valid name"},
		{name: "a", want: true, reason: "valid name"},
		{name: "เร็ว", want: true, reason: "valid name"},
		{name: "快速地", want: true, reason: "valid name"},
		{name: "швидко", want: true, reason: "valid name"},
		{name: "سريع", want: true, reason: "valid name"},
		{name: "Dimo's Fav Car", want: true, reason: "valid name"},
		{name: "My Car: 2022", want: true, reason: "valid name"},
		{name: "Car #2", want: true, reason: "valid name"},
		{name: `Sally "Speed Demon" Sedan`, want: true, reason: "valid name"},
		{name: "Valid Car Name", want: true, reason: "valid name"},
		{name: " Invalid Name", want: false, reason: "starts with space"},
		{name: "My Car!!!", want: true, reason: "valid name with !"},
		{name: "", want: false, reason: "empty name"},
		{name: "ThisNameIsTooLong--CanOnlyBe40CharactersInLengthxdd", want: false, reason: "too long"},
		{name: "Audi E-tron Sportback Atanas", want: true, reason: "up to 40 characters"},
		{name: "no\nNewLine", want: false, reason: "no new lines allowed"},
		{name: "RC Kia eNiro 4+", want: true, reason: "+ is okay"},
		{name: "Tesla (Alaska)", want: true, reason: "Parentheses allowed"},
	}

	for _, tc := range tests {
		vinReq := UpdateNameReq{Name: &tc.name}
		err := vinReq.validate()
		if tc.want {
			assert.NoError(s.T(), err, tc.reason)
		} else {
			assert.Error(s.T(), err, tc.reason)
		}
	}
}

func (s *UserDevicesControllerTestSuite) TestPatchName() {
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, ksuid.New().String(), nil, "", s.pdb)
	deviceID := uuid.New().String()
	apunit := test.SetupCreateAutoPiUnit(s.T(), s.testUserID, uuid.NewString(), &deviceID, s.pdb)
	autoPiIntID := ksuid.New().String()
	vehicleID := 3214
	_ = test.SetupCreateUserDeviceAPIIntegration(s.T(), apunit.AutopiUnitID, deviceID, ud.ID, autoPiIntID, s.pdb)

	// nil check test
	payload := `{}`
	request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/name", payload)
	response, _ := s.app.Test(request)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	// name with spaces happy path test
	testName := "Queens Charriot,.@!$’"
	payload = fmt.Sprintf(`{ "name": " %s " }`, testName) // intentionally has spaces to test trimming

	s.autoPiSvc.EXPECT().GetDeviceByUnitID(apunit.AutopiUnitID).Times(1).Return(&services.AutoPiDongleDevice{
		ID: deviceID, UnitID: apunit.AutopiUnitID, Vehicle: services.AutoPiDongleVehicle{ID: vehicleID},
	}, nil)
	s.autoPiSvc.EXPECT().PatchVehicleProfile(vehicleID, services.PatchVehicleProfile{
		CallName: &testName,
	}).Times(1).Return(nil)
	request = test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/name", payload)
	response, _ = s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}
	require.NoError(s.T(), ud.Reload(s.ctx, s.pdb.DBS().Reader))
	assert.Equal(s.T(), testName, ud.Name.String)
}

func (s *UserDevicesControllerTestSuite) TestPatchImageURL() {
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, ksuid.New().String(), nil, "", s.pdb)

	payload := `{ "imageUrl": "https://ipfs.com/planetary/car.jpg" }`
	request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/image", payload)
	response, _ := s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}
}

//go:embed test_drivly_pricing_by_vin.json
var testDrivlyPricingJSON string

//go:embed test_drivly_pricing2.json
var testDrivlyPricing2JSON string

//go:embed test_vincario_valuation.json
var testVincarioValuationJSON string

func (s *UserDevicesControllerTestSuite) TestGetDeviceValuations_Format1() {
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, ddID, nil, "", s.pdb)
	_ = test.SetupCreateExternalVINData(s.T(), ddID, &ud, map[string][]byte{
		"PricingMetadata": []byte(testDrivlyPricingJSON),
	}, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", ud.ID), "")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "valuationSets.#").Int()))
	assert.Equal(s.T(), 49957, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).mileage").Int()))
	assert.Equal(s.T(), 49957, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).odometer").Int()))
	assert.Equal(s.T(), "miles", gjson.GetBytes(body, "valuationSets.#(vendor=drivly).odometerUnit").String())
	assert.Equal(s.T(), 54123, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).retail").Int()))
	//54123 + 50151 / 2
	assert.Equal(s.T(), 52137, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).userDisplayPrice").Int()))
	assert.Equal(s.T(), "USD", gjson.GetBytes(body, "valuationSets.#(vendor=drivly).currency").String())
	// 49040 + 52173 + 49241 / 3 = 50151
	assert.Equal(s.T(), 50151, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).tradeIn").Int()))
	assert.Equal(s.T(), 50151, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).tradeInAverage").Int()))
}

func (s *UserDevicesControllerTestSuite) TestGetDeviceValuations_Format2() {
	// this is the other format we're seeing coming from drivly for pricing
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, ddID, nil, "", s.pdb)
	_ = test.SetupCreateExternalVINData(s.T(), ddID, &ud, map[string][]byte{
		"PricingMetadata": []byte(testDrivlyPricing2JSON),
	}, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", ud.ID), "")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "valuationSets.#").Int()))
	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(s.T(), 50702, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).mileage").Int()))
	assert.Equal(s.T(), 40611, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).tradeIn").Int()))
	assert.Equal(s.T(), 50803, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).retail").Int()))
}

func (s *UserDevicesControllerTestSuite) TestGetDeviceValuations_Vincario() {
	// this is the other format we're seeing coming from drivly for pricing
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, ddID, nil, "", s.pdb)
	_ = test.SetupCreateExternalVINData(s.T(), ddID, &ud, map[string][]byte{
		"VincarioMetadata": []byte(testVincarioValuationJSON),
	}, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", ud.ID), "")
	response, _ := s.app.Test(request, 2000)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "valuationSets.#").Int()))
	// mileage comes from request metadata, but it is also sometimes returned by payload
	assert.Equal(s.T(), 30137, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).mileage").Int()))
	assert.Equal(s.T(), 30137, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).odometer").Int()))
	assert.Equal(s.T(), "km", gjson.GetBytes(body, "valuationSets.#(vendor=vincario).odometerUnit").String())
	assert.Equal(s.T(), "EUR", gjson.GetBytes(body, "valuationSets.#(vendor=vincario).currency").String())

	assert.Equal(s.T(), 44800, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).tradeIn").Int()))
	assert.Equal(s.T(), 55200, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).retail").Int()))
	assert.Equal(s.T(), 51440, int(gjson.GetBytes(body, "valuationSets.#(vendor=vincario).userDisplayPrice").Int()))
}

//go:embed test_drivly_offers_by_vin.json
var testDrivlyOffersJSON string

func (s *UserDevicesControllerTestSuite) TestGetDeviceOffers() {
	// arrange db, insert some user_devices
	ddID := ksuid.New().String()
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, ddID, nil, "", s.pdb)
	_ = test.SetupCreateExternalVINData(s.T(), ddID, &ud, map[string][]byte{
		"OfferMetadata": []byte(testDrivlyOffersJSON),
		// "PricingMetadata":   nil,
		// "BlackbookMetadata": nil,
	}, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/offers", ud.ID), "")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "offerSets.#").Int()))
	assert.Equal(s.T(), "drivly", gjson.GetBytes(body, "offerSets.0.source").String())
	assert.Equal(s.T(), 3, int(gjson.GetBytes(body, "offerSets.0.offers.#").Int()))
	assert.Equal(s.T(), "Error in v1/acquisition/appraisal POST",
		gjson.GetBytes(body, "offerSets.0.offers.#(vendor=vroom).error").String())
	assert.Equal(s.T(), 10123, int(gjson.GetBytes(body, "offerSets.0.offers.#(vendor=carvana).price").Int()))
	assert.Equal(s.T(), "Make[Ford],Model[Mustang Mach-E],Year[2022] is not eligible for offer.",
		gjson.GetBytes(body, "offerSets.0.offers.#(vendor=carmax).declineReason").String())
}

//go:embed test_user_device_data.json
var testUserDeviceData []byte

func (s *UserDevicesControllerTestSuite) TestGetRange() {
	// arrange db, insert some user_devices
	autoPiUnitID := "1234"
	autoPiDeviceID := "4321"
	ddID := ksuid.New().String()
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	smartCarIntegration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, autoPiUnitID, &autoPiDeviceID, s.pdb)

	gddir := []*grpc.GetDeviceDefinitionItemResponse{
		{
			DeviceAttributes: []*grpc.DeviceTypeAttribute{
				{Name: "mpg", Value: "38.0"},
				{Name: "mpg_highway", Value: "40.0"},
				{Name: "fuel_tank_capacity_gal", Value: "14.5"},
			},
			Make: &grpc.DeviceMake{
				Name: "Ford",
			},
			Name:               "F-150",
			DeviceDefinitionId: ddID,
		},
	}
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, ddID, nil, "", s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), autoPiUnitID, autoPiDeviceID, ud.ID, integration.Id, s.pdb)
	udd := models.UserDeviceDatum{
		UserDeviceID:  ud.ID,
		Data:          null.JSONFrom(testUserDeviceData),
		IntegrationID: integration.Id,
	}
	err := udd.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	udd2 := models.UserDeviceDatum{
		UserDeviceID:  ud.ID,
		Data:          null.JSONFrom([]byte(`{"range":380.14,"timestamp":"2022-06-18T04:02:11.544Z"}`)),
		IntegrationID: smartCarIntegration.Id,
	}
	err = udd2.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ddID}).Return(gddir, nil)
	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/range", ud.ID), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 3, int(gjson.GetBytes(body, "rangeSets.#").Int()))
	assert.Equal(s.T(), "2022-06-18T04:06:40Z", gjson.GetBytes(body, "rangeSets.0.updated").String())
	assert.Equal(s.T(), "2022-06-18T04:06:40Z", gjson.GetBytes(body, "rangeSets.1.updated").String())
	assert.Equal(s.T(), "2022-06-18T04:02:11Z", gjson.GetBytes(body, "rangeSets.2.updated").String())
	assert.Equal(s.T(), "MPG", gjson.GetBytes(body, "rangeSets.0.rangeBasis").String())
	assert.Equal(s.T(), "MPG Highway", gjson.GetBytes(body, "rangeSets.1.rangeBasis").String())
	assert.Equal(s.T(), "Vehicle Reported", gjson.GetBytes(body, "rangeSets.2.rangeBasis").String())
	assert.Equal(s.T(), 391, int(gjson.GetBytes(body, "rangeSets.0.rangeDistance").Int()))
	assert.Equal(s.T(), 411, int(gjson.GetBytes(body, "rangeSets.1.rangeDistance").Int()))
	assert.Equal(s.T(), 236, int(gjson.GetBytes(body, "rangeSets.2.rangeDistance").Int()))
	assert.Equal(s.T(), "miles", gjson.GetBytes(body, "rangeSets.0.rangeUnit").String())
	assert.Equal(s.T(), "miles", gjson.GetBytes(body, "rangeSets.1.rangeUnit").String())
	assert.Equal(s.T(), "miles", gjson.GetBytes(body, "rangeSets.2.rangeUnit").String())
}

func (s *UserDevicesControllerTestSuite) TestPostRefreshSmartCar() {
	smartCarInt := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Escape", 2020, smartCarInt)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), constants.SmartCarVendor).Return(smartCarInt, nil)
	// arrange some additional data for this to work

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
	udd := models.UserDeviceDatum{
		UserDeviceID:  ud.ID,
		Data:          null.JSONFrom([]byte(`{"odometer": 123.223}`)),
		IntegrationID: smartCarInt.Id,
		CreatedAt:     time.Now().UTC().Add(time.Hour * -4),
		UpdatedAt:     time.Now().UTC().Add(time.Hour * -4),
	}
	err = udd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
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
	integration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mache", 2022, integration)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), constants.SmartCarVendor).Return(integration, nil)

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
	// arrange data to cause condition
	udd := models.UserDeviceDatum{
		UserDeviceID:  ud.ID,
		Data:          null.JSON{},
		IntegrationID: integration.Id,
	}
	err = udd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	payload := `{}`
	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/commands/refresh", payload)
	response, _ := s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusTooManyRequests, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body))
	}
}

func TestEIP712Hash(t *testing.T) {
	td := &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintDevice": {
				{Name: "rootNode", Type: "uint256"},
				{Name: "attributes", Type: "string[]"},
				{Name: "infos", Type: "string[]"},
			},
		},
		PrimaryType: "MintDevice",
		Domain: signer.TypedDataDomain{
			Name:              "DIMO",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x5fbdb2315678afecb367f032d93f642f64180aa3",
		},
		Message: signer.TypedDataMessage{
			"rootNode":   math.NewHexOrDecimal256(7), // Just hardcoding this. We need a node for each make, and to keep these in sync.
			"attributes": []any{"Make", "Model", "Year"},
			"infos":      []any{"Tesla", "Model 3", "2020"},
		},
	}
	hash, err := computeTypedDataHash(td)
	if assert.NoError(t, err) {
		realHash := common.HexToHash("0x8258cd28afb13c201c07bf80c717d55ce13e226b725dd8a115ae5ab064e537da")
		assert.Equal(t, realHash, hash)
	}
}

func TestEIP712Recover(t *testing.T) {
	td := &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintDevice": {
				{Name: "rootNode", Type: "uint256"},
				{Name: "attributes", Type: "string[]"},
				{Name: "infos", Type: "string[]"},
			},
		},
		PrimaryType: "MintDevice",
		Domain: signer.TypedDataDomain{
			Name:              "DIMO",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x5fbdb2315678afecb367f032d93f642f64180aa3",
		},
		Message: signer.TypedDataMessage{
			"rootNode":   math.NewHexOrDecimal256(7), // Just hardcoding this. We need a node for each make, and to keep these in sync.
			"attributes": []any{"Make", "Model", "Year"},
			"infos":      []any{"Tesla", "Model 3", "2020"},
		},
	}
	sig := common.FromHex("0x558266d4d8cd994c9eab2dee0efeb3ee33c839e4ce77c64da544679a85bd4a864805dd1fab769e9888fdfc0ed6502f685dc43ddda1add760febd749acfcd517b1b")
	addr, err := recoverAddress(td, sig)
	if assert.NoError(t, err) {
		realAddr := common.HexToAddress("0x969602c4f39D345Cbe47E7fe0dd8F1f16f984D65")
		assert.Equal(t, realAddr, addr)
	}
}
