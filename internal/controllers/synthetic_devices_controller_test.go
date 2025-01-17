package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	smock "github.com/IBM/sarama/mocks"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.uber.org/mock/gomock"
)

var signature = "0xa4438e5cb667dc63ebd694167ae3ad83585f2834c9b04895dd890f805c4c459a024ed9df1b03872536b4ac0c7720d02cb787884a093cfcde5c3bd7f94657e30c1b"
var userEthAddress = "0xd64E249A06ee6263d989e43aBFe12748a2506f88"
var mockProducer *smock.SyncProducer
var mockUserID = ksuid.New().String()

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
}

func (s *SyntheticDevicesControllerTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())

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

	logger := test.Logger()

	c := NewSyntheticDevicesController(mockSettings, s.pdb.DBS, logger, s.deviceDefSvc, s.userClient, s.syntheticDeviceSigSvc, client)
	s.sdc = c

	app := test.SetupAppFiber(*logger)

	app.Post("/v1/user/devices/:userDeviceID/integrations/:integrationID/commands/mint", test.AuthInjectorTestHandler(mockUserID), c.MintSyntheticDevice)
	app.Get("/v1/user/devices/:userDeviceID/integrations/:integrationID/commands/mint", test.AuthInjectorTestHandler(mockUserID), c.GetSyntheticDeviceMintingPayload)

	s.app = app
}

// TearDownTest after each test truncate tables
func (s *SyntheticDevicesControllerTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *SyntheticDevicesControllerTestSuite) TearDownSuite() {
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

// Test Runner
func TestSyntheticDevicesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(SyntheticDevicesControllerTestSuite))
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	email := "some@email.com"
	eth := addr.Hex()

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &pb.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: mockUserID}).Return(user, nil)

	integration := test.BuildIntegrationForGRPCRequest(10, "SmartCar")
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), "ford_escape_2020").Return(&grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "ford_escape_2020",
		Id:                 "ford_escape_2020",
		Make: &grpc.DeviceMake{
			Name: "Ford",
			Id:   ksuid.New().String(),
		},
		Verified: true,
	}, nil)

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, integration.Id, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, integration.Id), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	rawExpectedResp := s.sdc.getEIP712Mint(int64(10), int64(57))
	expectedRespJSON, err := json.Marshal(rawExpectedResp)
	s.NoError(err)

	s.Equal(fiber.StatusOK, response.StatusCode)
	// TODO: This is a bit circular.
	s.JSONEq(string(body), string(expectedRespJSON))
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_UserDetailsError() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("User not found"))

	integration := test.BuildIntegrationForGRPCRequest(10, "SmartCar")

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, integration.Id, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, integration.Id), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusInternalServerError, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"User not found"}`, fiber.StatusInternalServerError)), body)
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_NoEthereumAddressForUser() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	email := "some@email.com"

	user := test.BuildGetUserGRPC(mockUserID, &email, nil, &pb.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	integration := test.BuildIntegrationForGRPCRequest(10, "SmartCar")

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, integration.Id, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, integration.Id), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	s.Equal(fiber.StatusUnauthorized, response.StatusCode)
	s.JSONEq(`{"code":401,"message":"User does not have an Ethereum address."}`, string(body))
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_GetIntegrationFailed() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	email := "some@email.com"
	eth := addr.Hex()

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &pb.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: mockUserID}).Return(user, nil)

	integration := test.BuildIntegrationForGRPCRequest(10, "SmartCar")
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(nil, errors.New("could not find integration"))

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, integration.Id, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, integration.Id), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	s.Equal(fiber.StatusInternalServerError, response.StatusCode)
	s.JSONEq(`{"code":500,"message":"could not find integration"}`, string(body))
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_VehicleNodeNotExist() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	email := "some@email.com"
	eth := addr.Hex()

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &pb.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: mockUserID}).Return(user, nil)

	integration := test.BuildIntegrationForGRPCRequest(10, "SmartCar")

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, integration.Id), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusNotFound, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"No vehicle with that id found."}`, fiber.StatusNotFound)), body)
}

func (s *SyntheticDevicesControllerTestSuite) Test_MintSyntheticDeviceSmartcar() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	email := "some@email.com"
	eth := userEthAddress
	// addr := common.HexToAddress(userEthAddress)
	deviceEthAddr := common.HexToAddress("11")

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &pb.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: mockUserID}).Return(user, nil)

	integration := test.BuildIntegrationForGRPCRequest(1, "SmartCar")
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, integration.Id, s.pdb)

	vehicleSig := common.BytesToHash(common.HexToAddress("20").Bytes()).Bytes()
	s.syntheticDeviceSigSvc.EXPECT().SignHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(vehicleSig, nil).AnyTimes()
	s.syntheticDeviceSigSvc.EXPECT().GetAddress(gomock.Any(), gomock.Any()).Return(deviceEthAddr.Bytes(), nil).AnyTimes()

	var kb []byte
	mockProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
		kb = val
		return nil
	})

	req := fmt.Sprintf(`{
		"signature": "%s"
	}`, signature)

	request := test.BuildRequest("POST", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, integration.Id), req)
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	mockProducer.Close()

	assert.Equal(s.T(), "{\"message\":\"Submitted synthetic device mint request.\"}", string(body))

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

	vnID := types.NewNullDecimal(decimal.New(57, 0))
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
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_FailToMintKia() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	email := "some@email.com"
	eth := addr.Hex()

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &pb.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: mockUserID}).Return(user, nil)

	integration := test.BuildIntegrationForGRPCRequest(10, "SmartCar")
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), "ford_escape_2020").Return(&grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "ford_escape_2020",
		Id:                 "ford_escape_2020",
		Year:               2020,
		Make: &grpc.DeviceMake{
			Name: "Kia",
			Id:   "2681cSm2zmTmGHzqK3ldzoTLZIw",
		},
		Verified: true,
	}, nil)

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Escape", 2020, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, integration.Id, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, integration.Id), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)
	s.Equal(fiber.StatusFailedDependency, response.StatusCode)

	body, _ := io.ReadAll(response.Body)

	var resp map[string]interface{}
	err = json.Unmarshal(body, &resp)
	s.Require().NoError(err)

	s.Equal("Kia vehicles connected via smartcar cannot be manually minted.", resp["message"])
}

func (s *SyntheticDevicesControllerTestSuite) TestSignSyntheticDeviceMintingPayload_BadSignatureFailure() {
	s.T().Skip()
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	email := "some@email.com"
	eth := addr.Hex()

	user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &pb.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: mockUserID}).Return(user, nil)

	integration := test.BuildIntegrationForGRPCRequest(10, "SmartCar")
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.Id).Return(integration, nil)

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, integration.Id, s.pdb)

	req := `{
		"signature": "0xa3438e5cb667dc63ebd694167ae3ad83585f2834c9b04895dd890f805c4c459a024ed9df1b03872536b4ac0c7720d02cb787884a093cfcde5c3bd7f94657e30c1b"
	}`
	request := test.BuildRequest("POST", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, integration.Id), req)
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	msg := struct {
		Message string `json:"message"`
	}{}
	err = json.Unmarshal(body, &msg)
	s.NoError(err)
	assert.Equal(s.T(), fiber.StatusInternalServerError, response.StatusCode)
	assert.Equal(s.T(), secp256k1.ErrRecoverFailed.Error(), msg.Message)

	// //

	// email := "some@email.com"
	// eth := userEthAddress

	// user := test.BuildGetUserGRPC(mockUserID, &email, &eth, &pb.UserReferrer{})
	// s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	// scInt := &ddgrpc.Integration{
	// 	ID:      "22N2xaPOq2WW2gAHBHd0Ikn4Zob",
	// 	Vendor:  "SmartCar",
	// 	TokenId: 1,
	// }

	// s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), uint64(1)).Return(scInt, nil)

	// udID := ksuid.New().String()
	// _ = test.SetupCreateVehicleNFTForMiddleware(s.T(), common.HexToAddress(userEthAddress), mockUserID, udID, 57, s.pdb)

	// req := `{
	// 	"ownerSignature": "0xa3438e5cb667dc63ebd694167ae3ad83585f2834c9b04895dd890f805c4c459a024ed9df1b03872536b4ac0c7720d02cb787884a093cfcde5c3bd7f94657e30c1b"
	// }`
	// request := test.BuildRequest("POST", fmt.Sprintf("/v1/synthetic/device/mint/%d/%d", 1, 57), req)
	// response, err := s.app.Test(request)
	// require.NoError(s.T(), err)

	// body, _ := io.ReadAll(response.Body)

	// msg := struct {
	// 	Message string `json:"message"`
	// }{}

}
