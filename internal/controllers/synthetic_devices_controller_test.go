package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	smock "github.com/Shopify/sarama/mocks"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/types"
)

var signature = "0xa4438e5cb667dc63ebd694167ae3ad83585f2834c9b04895dd890f805c4c459a024ed9df1b03872536b4ac0c7720d02cb787884a093cfcde5c3bd7f94657e30c1b"
var userEthAddress = "0xd64E249A06ee6263d989e43aBFe12748a2506f88"
var mockProducer *smock.SyncProducer
var mockUserID = ksuid.New().String()
var mockSmartClientToken = &smartcar.Token{
	Access:       "mockAccessToken",
	AccessExpiry: time.Now().Add(time.Hour + 1),
	Refresh:      "mockRefreshToken",
}

type SyntheticDevicesControllerTestSuite struct {
	suite.Suite
	pdb                   db.Store
	container             testcontainers.Container
	ctx                   context.Context
	mockCtrl              *gomock.Controller
	app                   *fiber.App
	deviceDefSvc          *mock_services.MockDeviceDefinitionService
	userClient            *mock_services.MockUserServiceClient
	sdc                   SyntheticDevicesController
	syntheticDeviceSigSvc *mock_services.MockSyntheticWalletInstanceService
	smartcarClient        *mock_services.MockSmartcarClient
}

// SetupSuite starts container db
func (s *SyntheticDevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)

	s.mockCtrl = gomock.NewController(s.T())
	var err error

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.userClient = mock_services.NewMockUserServiceClient(s.mockCtrl)
	s.syntheticDeviceSigSvc = mock_services.NewMockSyntheticWalletInstanceService(s.mockCtrl)
	s.smartcarClient = mock_services.NewMockSmartcarClient(s.mockCtrl)

	mockProducer = smock.NewSyncProducer(s.T(), nil)

	mockSettings := &config.Settings{Port: "3000", DIMORegistryChainID: 80001, DIMORegistryAddr: common.HexToAddress("0x4De1bCf2B7E851E31216fC07989caA902A604784").Hex()}
	mockSettings.DB.Name = "devices_api"

	client := registry.Client{
		Producer:     mockProducer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(mockSettings.DIMORegistryChainID),
			Address: common.HexToAddress(mockSettings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	if err != nil {
		s.T().Fatal(err)
	}

	logger := test.Logger()

	c := NewSyntheticDevicesController(mockSettings, s.pdb.DBS, logger, s.deviceDefSvc, s.userClient, s.syntheticDeviceSigSvc, client, s.smartcarClient, new(shared.ROT13Cipher))
	s.sdc = c

	app := test.SetupAppFiber(*logger)

	app.Post("/v1/synthetic/device/mint/:integrationNode/:vehicleNode", test.AuthInjectorTestHandler(mockUserID), c.MintSyntheticDevice)
	app.Get("/v1/synthetic/device/mint/:integrationNode/:vehicleNode", test.AuthInjectorTestHandler(mockUserID), c.GetSyntheticDeviceMintingPayload)

	s.app = app
}

// TearDownTest after each test truncate tables
func (s *SyntheticDevicesControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *SyntheticDevicesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())

	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

// Test Runner
func TestSyntheticDevicesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(SyntheticDevicesControllerTestSuite))
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload() {
	_, addr, err := test.GenerateWallet()
	assert.NoError(s.T(), err)

	email := "some@email.com"
	eth := addr.Hex()

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	integrations := test.BuildIntegrationForGRPCRequest(10, uint64(1))
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(integrations, nil)

	_ = test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	_ = test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	rawExpectedResp := s.sdc.getEIP712(int64(1), int64(57))
	expectedRespJSON, err := json.Marshal(rawExpectedResp)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	assert.Equal(s.T(), body, expectedRespJSON)
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_UserNotFound() {
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("User not found"))

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusInternalServerError, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"error occurred when fetching user: User not found"}`, fiber.StatusInternalServerError)), body)
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_NoEthereumAddressForUser() {
	email := "some@email.com"
	user := test.BuildGetUserGRPC(mockUserID, &email, nil, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusUnauthorized, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"User does not have an Ethereum address on file."}`, fiber.StatusUnauthorized)), body)
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_NoIntegrationForID() {
	_, addr, err := test.GenerateWallet()
	assert.NoError(s.T(), err)

	eth := addr.Hex()
	email := "some@email.com"

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(nil, errors.New("could not find integration"))

	udID := ksuid.New().String()
	_ = test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusInternalServerError, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"failed to get integration: could not find integration"}`, fiber.StatusInternalServerError)), body)
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_VehicleNodeNotOwnedByUserEthAddress() {
	_, addr, err := test.GenerateWallet()
	assert.NoError(s.T(), err)

	eth := addr.Hex()
	email := "some@email.com"

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusNotFound, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"user does not own vehicle node"}`, fiber.StatusNotFound)), body)
}

func (s *SyntheticDevicesControllerTestSuite) Test_MintSyntheticDevice() {
	email := "some@email.com"
	eth := userEthAddress
	addr := common.HexToAddress(userEthAddress)
	deviceEthAddr := common.HexToAddress("11")

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	integration := test.BuildIntegrationForGRPCRequest(10, uint64(1))
	integration.Vendor = constants.SmartCarVendor
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(integration, nil)

	_ = test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	vehicle := test.SetupCreateVehicleNFTForMiddleware(s.T(), addr, mockUserID, udID, 57, s.pdb)

	vehicleSig := common.HexToAddress("20").Hash().Bytes()
	s.syntheticDeviceSigSvc.EXPECT().SignHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(vehicleSig, nil).AnyTimes()
	s.syntheticDeviceSigSvc.EXPECT().GetAddress(gomock.Any(), gomock.Any()).Return(deviceEthAddr.Bytes(), nil).AnyTimes()

	s.smartcarClient.EXPECT().ExchangeCode(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSmartClientToken, nil)
	s.smartcarClient.EXPECT().GetExternalID(gomock.Any(), gomock.Any()).Return("smartcarVehicleId", nil)
	s.smartcarClient.EXPECT().GetEndpoints(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"/odometer", "/location"}, nil)

	var kb []byte
	mockProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
		kb = val
		return nil
	})

	req := fmt.Sprintf(`{
		"credentials": {
			"authorizationCode": "a4d04dad-2b65-4778-94b7-f04996e89907"
		},
		"ownerSignature": "%s"
	}`, signature)

	request := test.BuildRequest("POST", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	mockProducer.Close()

	assert.Equal(s.T(), "synthetic device mint request successful", string(body))

	var me shared.CloudEvent[registry.RequestData]

	err = json.Unmarshal(kb, &me)
	s.Require().NoError(err)

	abi, err := contracts.RegistryMetaData.GetAbi()
	s.Require().NoError(err)

	method := abi.Methods["mintSyntheticDeviceSign"]

	callData := me.Data.Data

	s.EqualValues(method.ID, callData[:4])

	o, err := method.Inputs.Unpack(callData[4:])
	s.Require().NoError(err)

	actualMnInput := o[0].(struct {
		IntegrationNode     *big.Int       `json:"integrationNode"`
		VehicleNode         *big.Int       `json:"vehicleNode"`
		SyntheticDeviceSig  []uint8        `json:"syntheticDeviceSig"`
		VehicleOwnerSig     []uint8        `json:"vehicleOwnerSig"`
		SyntheticDeviceAddr common.Address `json:"syntheticDeviceAddr"`
		AttrInfoPairs       []struct {
			Attribute string `json:"attribute"`
			Info      string `json:"info"`
		} `json:"attrInfoPairs"`
	})

	expectedMnInput := contracts.MintSyntheticDeviceInput{
		IntegrationNode:     new(big.Int).SetUint64(1),
		VehicleNode:         new(big.Int).SetUint64(57),
		VehicleOwnerSig:     common.FromHex(signature),
		SyntheticDeviceAddr: deviceEthAddr,
		SyntheticDeviceSig:  vehicleSig,
	}

	vnID := types.NewDecimal(decimal.New(57, 0))
	syntDevice, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.VehicleTokenID.EQ(vnID),
		models.SyntheticDeviceWhere.IntegrationTokenID.EQ(types.NewDecimal(decimal.New(1, 0))),
	).One(s.ctx, s.pdb.DBS().Reader)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), syntDevice.IntegrationTokenID, types.NewDecimal(decimal.New(1, 0)))
	assert.Equal(s.T(), syntDevice.VehicleTokenID, vnID)

	assert.ObjectsAreEqual(expectedMnInput, actualMnInput)

	metaTrxReq, err := models.MetaTransactionRequests(
		models.MetaTransactionRequestWhere.ID.EQ(syntDevice.MintRequestID),
		models.MetaTransactionRequestWhere.Status.EQ(models.MetaTransactionRequestStatusUnsubmitted),
	).Exists(s.ctx, s.pdb.DBS().Reader)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), metaTrxReq, true)

	udis, err := models.UserDeviceAPIIntegrations().All(s.ctx, s.pdb.DBS().Reader)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 1, len(udis))

	decAccessToken, err := s.sdc.cipher.Decrypt(udis[0].AccessToken.String)
	assert.NoError(s.T(), err)

	decRefreshToken, err := s.sdc.cipher.Decrypt(udis[0].RefreshToken.String)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), integration.Id, udis[0].IntegrationID)
	assert.Equal(s.T(), vehicle.UserDeviceID.String, udis[0].UserDeviceID)
	assert.Equal(s.T(), mockSmartClientToken.Access, decAccessToken)
	assert.Equal(s.T(), mockSmartClientToken.Refresh, decRefreshToken)
}

func (s *SyntheticDevicesControllerTestSuite) TestSignSyntheticDeviceMintingPayload_BadSignatureFailure() {
	integration := test.BuildIntegrationForGRPCRequest(10, uint64(1))
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(integration, nil)

	req := fmt.Sprintf(`{
		"credentials": {
			"authorizationCode": "a4d04dad-2b65-4778-94b7-f04996e89907"
		},
		"ownerSignature": "%s"
	}`, "Bad Signature")
	request := test.BuildRequest("POST", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"invalid signature provided"}`, fiber.StatusBadRequest)), body)
}

func (s *SyntheticDevicesControllerTestSuite) TestSignSyntheticDeviceMintingPayload_BadSignatureLengthFailure() {
	integration := test.BuildIntegrationForGRPCRequest(10, uint64(1))
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(integration, nil)

	req := fmt.Sprintf(`{
		"credentials": {
			"authorizationCode": "a4d04dad-2b65-4778-94b7-f04996e89907"
		},
		"ownerSignature": "%s"
	}`, "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	request := test.BuildRequest("POST", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"invalid signature provided"}`, fiber.StatusBadRequest)), body)
}

func (s *SyntheticDevicesControllerTestSuite) Test_Synthetic_Device_Sequence() {
	childKeyNumber, err := s.sdc.generateNextChildKeyNumber(s.ctx)
	assert.NoError(s.T(), err)

	startSeq := 2 // We start from 2 because the initial mint test would have generated sequence 1 already

	assert.Equal(s.T(), startSeq, childKeyNumber)

	childKeyNumber, err = s.sdc.generateNextChildKeyNumber(s.ctx)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), startSeq+1, childKeyNumber)
}

// handleDeviceApiIntegrationCreation
func (s *SyntheticDevicesControllerTestSuite) Test_Device_API_Integration_Creation() {
	ctx := context.Background()
	addr := common.HexToAddress(userEthAddress)

	integration := test.BuildIntegrationForGRPCRequest(10, uint64(1))
	integration.Vendor = constants.SmartCarVendor

	udID := ksuid.New().String()
	vehicle := test.SetupCreateVehicleNFTForMiddleware(s.T(), addr, mockUserID, udID, 57, s.pdb)

	s.smartcarClient.EXPECT().ExchangeCode(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockSmartClientToken, nil)
	s.smartcarClient.EXPECT().GetExternalID(gomock.Any(), "mockAccessToken").Return("smartcarVehicleId", nil)
	s.smartcarClient.EXPECT().GetEndpoints(gomock.Any(), "mockAccessToken", "smartcarVehicleId").Return([]string{"/odometer", "/location"}, nil)

	tx, err := s.sdc.DBS().Writer.DB.BeginTx(ctx, nil)
	assert.NoError(s.T(), err)

	err = s.sdc.handleDeviceAPIIntegrationCreation(ctx, tx, &MintSyntheticDeviceRequest{
		Credentials: struct {
			AuthorizationCode string `json:"authorizationCode"`
		}{AuthorizationCode: "mockAuthCode"},
	}, vehicle.UserDeviceID.String, integration)
	assert.NoError(s.T(), err)

	udis, err := models.UserDeviceAPIIntegrations().All(ctx, tx)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), 1, len(udis))

	decAccessToken, err := s.sdc.cipher.Decrypt(udis[0].AccessToken.String)
	assert.NoError(s.T(), err)

	decRefreshToken, err := s.sdc.cipher.Decrypt(udis[0].RefreshToken.String)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), integration.Id, udis[0].IntegrationID)
	assert.Equal(s.T(), vehicle.UserDeviceID.String, udis[0].UserDeviceID)
	assert.Equal(s.T(), mockSmartClientToken.Access, decAccessToken)
	assert.Equal(s.T(), mockSmartClientToken.Refresh, decRefreshToken)
	s.Equal("smartcarVehicleId", udis[0].ExternalID.String)
	s.JSONEq(`{"smartcarEndpoints": ["/odometer", "/location"]}`, string(udis[0].Metadata.JSON))

	defer tx.Rollback() //nolint
}
