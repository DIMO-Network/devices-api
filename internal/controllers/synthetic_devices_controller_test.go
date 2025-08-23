package controllers

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/shared/pkg/payloads"
	pb_oracle "github.com/DIMO-Network/tesla-oracle/pkg/grpc"
	smock "github.com/IBM/sarama/mocks"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source synthetic_devices_controller.go -destination synthetic_devices_controller_mock_test.go -package controllers

var rawPrivateRandomKey = "df17cb3eac8df0fdfc96d44a7423a952e068d169e295af7f1c2607e98fb190e4"

// var signature = "0xa4438e5cb667dc63ebd694167ae3ad83585f2834c9b04895dd890f805c4c459a024ed9df1b03872536b4ac0c7720d02cb787884a093cfcde5c3bd7f94657e30c1b"
var mockProducer *smock.SyncProducer

var mockUserID = ksuid.New().String()

// var userEthAddress = common.HexToAddress("0xd64E249A06ee6263d989e43aBFe12748a2506f88")

type SyntheticDevicesControllerTestSuite struct {
	suite.Suite
	pdb                   db.Store
	container             testcontainers.Container
	ctx                   context.Context
	mockCtrl              *gomock.Controller
	app                   *fiber.App
	deviceDefSvc          *mock_services.MockDeviceDefinitionService
	sdc                   SyntheticDevicesController
	syntheticDeviceSigSvc *mock_services.MockSyntheticWalletInstanceService
	userPrivKey           *ecdsa.PrivateKey
	userAddr              common.Address
	mockOracle            *MockTeslaOracleClient
}

// SetupSuite starts container db
func (s *SyntheticDevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
}

func (s *SyntheticDevicesControllerTestSuite) SetupTest() {
	privKey, err := crypto.HexToECDSA(rawPrivateRandomKey)
	if err != nil {
		s.Require().NoError(err)
	}

	s.userPrivKey = privKey

	s.userAddr = crypto.PubkeyToAddress(privKey.PublicKey)

	s.mockCtrl = gomock.NewController(s.T())

	s.mockOracle = NewMockTeslaOracleClient(s.mockCtrl)

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.syntheticDeviceSigSvc = mock_services.NewMockSyntheticWalletInstanceService(s.mockCtrl)

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

	c := NewSyntheticDevicesController(mockSettings, s.pdb.DBS, logger, s.deviceDefSvc, s.syntheticDeviceSigSvc, client, s.mockOracle)
	s.sdc = c

	app := test.SetupAppFiber(*logger)

	app.Post("/v1/user/devices/:userDeviceID/integrations/:integrationID/commands/mint", test.AuthInjectorTestHandler(mockUserID, &s.userAddr), c.MintSyntheticDevice)
	app.Get("/v1/user/devices/:userDeviceID/integrations/:integrationID/commands/mint", test.AuthInjectorTestHandler(mockUserID, &s.userAddr), c.GetSyntheticDeviceMintingPayload)

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

const teslaKSUID = "26A5Dk3vvvQutjSyF0Jka2DP5lg"

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload() {
	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model Y", 2022, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), s.userAddr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, teslaKSUID, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, teslaKSUID), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	rawExpectedResp := s.sdc.getEIP712Mint(2, 57)
	expectedRespJSON, err := json.Marshal(rawExpectedResp)
	s.NoError(err)

	s.Equal(fiber.StatusOK, response.StatusCode)
	// TODO: This is a bit circular.
	s.JSONEq(string(body), string(expectedRespJSON))
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_GetIntegrationFailed() {
	unknownIntegrationID := ksuid.New().String()

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	test.SetupCreateVehicleNFTForMiddleware(s.T(), s.userAddr, mockUserID, udID, 57, s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, unknownIntegrationID, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, unknownIntegrationID), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	s.Equal(fiber.StatusBadRequest, response.StatusCode)
	s.JSONEq(`{"code":400,"message":"Cannot mint this integration with devices-api."}`, string(body))
}

func (s *SyntheticDevicesControllerTestSuite) TestGetSyntheticDeviceMintingPayload_VehicleNodeNotExist() {
	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Model 3", 2022, nil)

	udID := ksuid.New().String()

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, teslaKSUID), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusNotFound, response.StatusCode)
	assert.Equal(s.T(), fmt.Sprintf(`{"code":%d,"message":"No vehicle with that id found."}`, fiber.StatusNotFound), string(body))
}

func (s *SyntheticDevicesControllerTestSuite) Test_MintSyntheticDeviceTesla() {

	deviceEthAddr := common.HexToAddress("11")

	test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	ud := test.SetupCreateVehicleNFTForMiddleware(s.T(), s.userAddr, mockUserID, udID, 57, s.pdb)

	ud.VinConfirmed = true
	ud.VinIdentifier = null.StringFrom("5YJ3E1EB6KF511862")

	_, err := ud.Update(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	test.SetupCreateUserDeviceAPIIntegration(s.T(), "", "xddL", udID, teslaKSUID, s.pdb)

	vehicleSig := common.BytesToHash(common.HexToAddress("20").Bytes()).Bytes()
	s.syntheticDeviceSigSvc.EXPECT().SignHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(vehicleSig, nil).AnyTimes()
	s.syntheticDeviceSigSvc.EXPECT().GetAddress(gomock.Any(), gomock.Any()).Return(deviceEthAddr.Bytes(), nil).AnyTimes()

	rawExpectedResp := s.sdc.getEIP712Mint(2, 57)

	tdHash, _, err := signer.TypedDataAndHash(*rawExpectedResp)
	if err != nil {
		panic(err)
	}

	signature, err := crypto.Sign(tdHash, s.userPrivKey)
	s.Require().NoError(err)

	signature[64] += 27

	var kb []byte
	mockProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
		kb = val
		return nil
	})

	req := fmt.Sprintf(`{
		"signature": "%s"
	}`, hexutil.Encode(signature))

	s.mockOracle.EXPECT().RegisterNewSyntheticDeviceV2(gomock.Any(), &pb_oracle.RegisterNewSyntheticDeviceRequest{
		Vin: ud.VinIdentifier.String,
	})

	request := test.BuildRequest("POST", fmt.Sprintf("/v1/user/devices/%s/integrations/%s/commands/mint", udID, teslaKSUID), req)
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	body, _ := io.ReadAll(response.Body)

	s.Equal(fiber.StatusOK, response.StatusCode)

	mockProducer.Close()

	assert.Equal(s.T(), "{\"message\":\"Submitted synthetic device mint request.\"}", string(body))

	var me payloads.CloudEvent[registry.RequestData]

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
		VehicleOwnerSig:     signature,
		SyntheticDeviceAddr: deviceEthAddr,
		SyntheticDeviceSig:  vehicleSig,
	}

	vnID := types.NewNullDecimal(decimal.New(57, 0))
	syntDevice, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.VehicleTokenID.EQ(vnID),
		models.SyntheticDeviceWhere.IntegrationTokenID.EQ(types.NewDecimal(decimal.New(2, 0))),
	).One(s.ctx, s.pdb.DBS().Reader)
	assert.NoError(s.T(), err)

	assert.ObjectsAreEqual(expectedMnInput, actualMnInput)

	metaTrxReq, err := models.MetaTransactionRequests(
		models.MetaTransactionRequestWhere.ID.EQ(syntDevice.MintRequestID),
		models.MetaTransactionRequestWhere.Status.EQ(models.MetaTransactionRequestStatusUnsubmitted),
	).Exists(s.ctx, s.pdb.DBS().Reader)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), metaTrxReq, true)
}

func (s *SyntheticDevicesControllerTestSuite) TestSignSyntheticDeviceMintingPayload_BadSignatureFailure() {
	s.T().Skip()
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	integration := test.BuildIntegrationForGRPCRequest(10, "Tesla")
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
}
