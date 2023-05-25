package controllers

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/config"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type VirtualDevicesControllerTestSuite struct {
	suite.Suite
	pdb             db.Store
	container       testcontainers.Container
	ctx             context.Context
	mockCtrl        *gomock.Controller
	app             *fiber.App
	deviceDefSvc    *mock_services.MockDeviceDefinitionService
	userClient      *mock_services.MockUserServiceClient
	deviceDefIntSvc *mock_services.MockDeviceDefinitionIntegrationService
	vdc             VirtualDeviceController
}

// SetupSuite starts container db
func (s *VirtualDevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)

	s.mockCtrl = gomock.NewController(s.T())
	var err error

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	s.userClient = mock_services.NewMockUserServiceClient(s.mockCtrl)

	if err != nil {
		s.T().Fatal(err)
	}

	logger := test.Logger()
	c := NewVirtualDeviceController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefIntSvc, s.deviceDefSvc, s.userClient)
	s.vdc = c

	app := test.SetupAppFiber(*logger)

	app.Post("/v1/integration/:tokenID/mint-virtual-device", test.AuthInjectorTestHandler(testUserID), c.SignVirtualDeviceMintingPayload)
	app.Get("/v1/integration/:tokenID/mint-virtual-device", test.AuthInjectorTestHandler(testUserID), c.GetVirtualDeviceMintingPayload)

	s.app = app
}

// TearDownTest after each test truncate tables
func (s *VirtualDevicesControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *VirtualDevicesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())

	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

// Test Runner
func TestVirtualDevicesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(VirtualDevicesControllerTestSuite))
}

func (s *VirtualDevicesControllerTestSuite) TestGetVirtualDeviceMintingPayload() {
	_, addr, err := test.GenerateWallet()
	assert.NoError(s.T(), err)

	email := "some@email.com"
	eth := addr.Hex()

	user := test.BuildGetUserGRPC(testUserID, &email, &eth, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	integrations := test.BuildIntegrationForGRPCRequest(10, 0, uint64(1))
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(integrations, nil)

	_ = test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	_ = test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, testUserID, udID, 57, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/integration/%d/mint-virtual-device?vehicle_id=%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	rawExpectedResp := s.vdc.getVirtualDeviceMintPayload(int64(1), int64(57))
	expectedRespJSON, err := json.Marshal(rawExpectedResp)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	assert.Equal(s.T(), body, expectedRespJSON)
}

func (s *VirtualDevicesControllerTestSuite) TestGetVirtualDeviceMintingPayload_UserNotFound() {
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("User not found"))

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/integration/%d/mint-virtual-device?vehicle_id=%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusUnauthorized, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"error occurred when fetching user"}`, fiber.StatusUnauthorized)), body)
}

func (s *VirtualDevicesControllerTestSuite) TestGetVirtualDeviceMintingPayload_NoEthereumAddressForUser() {
	email := "some@email.com"
	user := test.BuildGetUserGRPC(testUserID, &email, nil, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/integration/%d/mint-virtual-device?vehicle_id=%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusUnauthorized, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"User does not have an Ethereum address on file."}`, fiber.StatusUnauthorized)), body)
}

func (s *VirtualDevicesControllerTestSuite) TestGetVirtualDeviceMintingPayload_NoIntegrationForID() {
	_, addr, err := test.GenerateWallet()
	assert.NoError(s.T(), err)

	eth := addr.Hex()
	email := "some@email.com"

	user := test.BuildGetUserGRPC(testUserID, &email, &eth, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(nil, errors.New("could not find integration"))

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/integration/%d/mint-virtual-device?vehicle_id=%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"failed to get integration"}`, fiber.StatusBadRequest)), body)
}

func (s *VirtualDevicesControllerTestSuite) TestGetVirtualDeviceMintingPayload_VehicleNodeNotOwnedByUserEthAddress() {
	_, addr, err := test.GenerateWallet()
	assert.NoError(s.T(), err)

	eth := addr.Hex()
	email := "some@email.com"

	user := test.BuildGetUserGRPC(testUserID, &email, &eth, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	integrations := test.BuildIntegrationForGRPCRequest(10, 0, uint64(1))
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(integrations, nil)

	request := test.BuildRequest("GET", fmt.Sprintf("/v1/integration/%d/mint-virtual-device?vehicle_id=%d", 1, 57), "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"user does not own vehicle node"}`, fiber.StatusBadRequest)), body)
}

func mockPayloadSignerHelper(privateKey *ecdsa.PrivateKey, payload []byte, t *testing.T) string {
	hash := crypto.Keccak256Hash(payload)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	assert.NoError(t, err)

	return hexutil.Encode(signature)
}

func (s *VirtualDevicesControllerTestSuite) TestSignVirtualDeviceMintingPayload() {
	pk, addr, err := test.GenerateWallet()
	assert.NoError(s.T(), err)

	email := "some@email.com"
	eth := addr.Hex()

	user := test.BuildGetUserGRPC(testUserID, &email, &eth, &users.UserReferrer{})
	s.userClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil)

	integrations := test.BuildIntegrationForGRPCRequest(10, 0, uint64(1))
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(integrations, nil)

	_ = test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Explorer", 2022, nil)

	udID := ksuid.New().String()
	_ = test.SetupCreateVehicleNFTForMiddleware(s.T(), *addr, testUserID, udID, 57, s.pdb)

	rawExpectedResp := s.vdc.getVirtualDeviceMintPayload(int64(1), int64(57))
	pJSON, err := json.Marshal(rawExpectedResp)
	assert.NoError(s.T(), err)

	sig := mockPayloadSignerHelper(pk, pJSON, s.T())
	req := fmt.Sprintf(`{
		"vehicleNode": %d,
		"credentials": {
			"authorizationCode": "a4d04dad-2b65-4778-94b7-f04996e89907"
		},
		"ownerSignature": "%s"
	}`, 57, sig[2:])

	request := test.BuildRequest("POST", fmt.Sprintf("/v1/integration/%d/mint-virtual-device?vehicle_id=%d", 1, 57), req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	assert.Equal(s.T(), "signature is valid", string(body))
}

func (s *VirtualDevicesControllerTestSuite) TestSignVirtualDeviceMintingPayload_BadSignatureFailure() {
	req := fmt.Sprintf(`{
		"vehicleNode": %d,
		"credentials": {
			"authorizationCode": "a4d04dad-2b65-4778-94b7-f04996e89907"
		},
		"ownerSignature": "%s"
	}`, 57, "Bad Signature")
	request := test.BuildRequest("POST", fmt.Sprintf("/v1/integration/%d/mint-virtual-device?vehicle_id=%d", 1, 57), req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"invalid signature provided"}`, fiber.StatusBadRequest)), body)
}

func (s *VirtualDevicesControllerTestSuite) TestSignVirtualDeviceMintingPayload_BadSignatureLengthFailure() {
	req := fmt.Sprintf(`{
		"vehicleNode": %d,
		"credentials": {
			"authorizationCode": "a4d04dad-2b65-4778-94b7-f04996e89907"
		},
		"ownerSignature": "%s"
	}`, 57, "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	request := test.BuildRequest("POST", fmt.Sprintf("/v1/integration/%d/mint-virtual-device?vehicle_id=%d", 1, 57), req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)

	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	assert.Equal(s.T(), []byte(fmt.Sprintf(`{"code":%d,"message":"invalid signature provided"}`, fiber.StatusBadRequest)), body)
}
