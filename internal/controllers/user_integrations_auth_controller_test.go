package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

type UserIntegrationAuthControllerTestSuite struct {
	suite.Suite
	pdb              db.Store
	controller       *UserIntegrationAuthController
	container        testcontainers.Container
	ctx              context.Context
	mockCtrl         *gomock.Controller
	app              *fiber.App
	deviceDefSvc     *mock_services.MockDeviceDefinitionService
	testUserID       string
	redisClient      *mocks.MockCacheService
	teslaFleetAPISvc *mock_services.MockTeslaFleetAPIService
	usersClient      *mock_services.MockUserServiceClient
	cipher           shared.Cipher
}

// SetupSuite starts container db
func (s *UserIntegrationAuthControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	logger := test.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(mockCtrl)
	s.teslaFleetAPISvc = mock_services.NewMockTeslaFleetAPIService(mockCtrl)
	s.redisClient = mocks.NewMockCacheService(mockCtrl)
	s.testUserID = "123123"
	s.cipher = new(shared.ROT13Cipher)
	s.usersClient = mock_services.NewMockUserServiceClient(mockCtrl)
	c := NewUserIntegrationAuthController(&config.Settings{
		Port:        "3000",
		Environment: "prod",
	}, s.pdb.DBS, logger, s.deviceDefSvc, s.teslaFleetAPISvc, s.redisClient, s.cipher, s.usersClient)
	app := test.SetupAppFiber(*logger)
	app.Post("/integration/:tokenID/credentials", test.AuthInjectorTestHandler(s.testUserID), c.CompleteOAuthExchange)

	s.controller = &c
	s.app = app
}

func (s *UserIntegrationAuthControllerTestSuite) SetupTest() {
	s.controller.Settings.Environment = "prod"
}

// TearDownTest after each test truncate tables
func (s *UserIntegrationAuthControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserIntegrationAuthControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish() // might need to do mockctrl on every test, and refactor setup into one method
}

// Test Runner
func TestUserIntegrationAuthControllerTestSuite(t *testing.T) {
	suite.Run(t, new(UserIntegrationAuthControllerTestSuite))
}

func (s *UserIntegrationAuthControllerTestSuite) TestCompleteOAuthExchanges() {
	mockAuthCode := "Mock_fd941f8da609db8cd66b1734f84ab289e2975b1889a5bedf478f02cf0cc4"
	mockRedirectURI := "https://mock-redirect.test.dimo.zone"
	mockRegion := "na"

	mockAuthCodeResp := &services.TeslaAuthCodeResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw",
		Expiry:       time.Now().Add(time.Hour * 1),
		Region:       mockRegion,
	}
	mockUserEthAddr := common.HexToAddress("1").String()
	s.usersClient.EXPECT().GetUser(gomock.Any(), &users.GetUserRequest{Id: s.testUserID}).Return(&users.User{EthereumAddress: &mockUserEthAddr}, nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), uint64(2)).Return(&ddgrpc.Integration{
		Vendor: constants.TeslaVendor,
	}, nil)
	s.teslaFleetAPISvc.EXPECT().CompleteTeslaAuthCodeExchange(gomock.Any(), mockAuthCode, mockRedirectURI, mockRegion).Return(mockAuthCodeResp, nil)
	s.deviceDefSvc.EXPECT().DecodeVIN(gomock.Any(), "1GBGC24U93Z337558", "", 0, "").Return(&ddgrpc.DecodeVinResponse{DeviceDefinitionId: "someID-1"}, nil)
	s.deviceDefSvc.EXPECT().DecodeVIN(gomock.Any(), "WAUAF78E95A553420", "", 0, "").Return(&ddgrpc.DecodeVinResponse{DeviceDefinitionId: "someID-2"}, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), "someID-1").Return(&ddgrpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "someID-1",
		Type: &ddgrpc.DeviceType{
			Make:  "Tesla",
			Model: "Y",
			Year:  2022,
		},
	}, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), "someID-2").Return(&ddgrpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "someID-2",
		Type: &ddgrpc.DeviceType{
			Make:  "Tesla",
			Model: "X",
			Year:  2020,
		},
	}, nil)

	tokenStr, err := json.Marshal(mockAuthCodeResp)
	s.Assert().NoError(err)

	encToken, err := s.cipher.Encrypt(string(tokenStr))
	s.Assert().NoError(err)

	cacheKey := fmt.Sprintf(teslaFleetAuthCacheKey, mockUserEthAddr)
	s.redisClient.EXPECT().Set(gomock.Any(), cacheKey, encToken, 5*time.Minute).Return(&redis.StatusCmd{})

	resp := []services.TeslaVehicle{
		{
			ID:  11114464922222,
			VIN: "1GBGC24U93Z337558",
		},
		{
			ID:  22222464999999,
			VIN: "WAUAF78E95A553420",
		},
	}
	s.teslaFleetAPISvc.EXPECT().GetVehicles(gomock.Any(), mockAuthCodeResp.AccessToken, mockRegion).Return(resp, nil)

	request := test.BuildRequest("POST", "/integration/2/credentials", fmt.Sprintf(`{
		"authorizationCode": "%s",
		"redirectUri": "%s",
		"region": "na"
	}`, mockAuthCode, mockRedirectURI))
	response, _ := s.app.Test(request)

	s.Assert().Equal(fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)

	expResp := &CompleteOAuthExchangeResponseWrapper{
		Vehicles: []CompleteOAuthExchangeResponse{
			{
				ExternalID: "11114464922222",
				VIN:        "1GBGC24U93Z337558",
				Definition: DeviceDefinition{
					Make:               "Tesla",
					Model:              "Y",
					Year:               2022,
					DeviceDefinitionID: "someID-1",
				},
			},
			{
				ExternalID: "22222464999999",
				VIN:        "WAUAF78E95A553420",
				Definition: DeviceDefinition{
					Make:               "Tesla",
					Model:              "X",
					Year:               2020,
					DeviceDefinitionID: "someID-2",
				},
			},
		},
	}

	expected, err := json.Marshal(expResp)
	s.Assert().NoError(err)

	s.Assert().Equal(expected, body)
}

func (s *UserIntegrationAuthControllerTestSuite) TestCompleteOAuthExchange_InvalidRegion() {
	mockUserEthAddr := common.HexToAddress("1").String()
	s.usersClient.EXPECT().GetUser(gomock.Any(), &users.GetUserRequest{Id: s.testUserID}).Return(&users.User{EthereumAddress: &mockUserEthAddr}, nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), uint64(2)).Return(&ddgrpc.Integration{
		Vendor: constants.TeslaVendor,
	}, nil)
	request := test.BuildRequest("POST", "/integration/2/credentials", fmt.Sprintf(`{
		"authorizationCode": "%s",
		"redirectUri": "%s",
		"region": "us-central"
	}`, "mockAuthCode", "mockRedirectURI"))
	response, _ := s.app.Test(request)

	s.Assert().Equal(fiber.StatusBadRequest, response.StatusCode)
	body, _ := io.ReadAll(response.Body)

	s.Assert().Equal(`{"code":400,"message":"invalid value provided for region, only na and eu are allowed"}`, string(body))
}

func (s *UserIntegrationAuthControllerTestSuite) TestCompleteOAuthExchange_UnprocessableTokenID() {
	mockUserEthAddr := common.HexToAddress("1").String()
	s.usersClient.EXPECT().GetUser(gomock.Any(), &users.GetUserRequest{Id: s.testUserID}).Return(&users.User{EthereumAddress: &mockUserEthAddr}, nil)
	request := test.BuildRequest("POST", "/integration/wrongTokenID/credentials", fmt.Sprintf(`{
		"authorizationCode": "%s",
		"redirectUri": "%s",
		"region": "us-central"
	}`, "mockAuthCode", "mockRedirectURI"))
	response, _ := s.app.Test(request)

	s.Assert().Equal(fiber.StatusBadRequest, response.StatusCode)
	body, _ := io.ReadAll(response.Body)

	s.Assert().Equal(`{"code":400,"message":"could not process the provided tokenId!"}`, string(body))
}

func (s *UserIntegrationAuthControllerTestSuite) TestCompleteOAuthExchange_InvalidTokenID() {
	mockUserEthAddr := common.HexToAddress("1").String()
	s.usersClient.EXPECT().GetUser(gomock.Any(), &users.GetUserRequest{Id: s.testUserID}).Return(&users.User{EthereumAddress: &mockUserEthAddr}, nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), uint64(1)).Return(&ddgrpc.Integration{
		Vendor: constants.SmartCarVendor,
	}, nil)

	request := test.BuildRequest("POST", "/integration/1/credentials", fmt.Sprintf(`{
		"authorizationCode": "%s",
		"redirectUri": "%s",
		"region": "us-central"
	}`, "mockAuthCode", "mockRedirectURI"))
	response, _ := s.app.Test(request)

	s.Assert().Equal(fiber.StatusBadRequest, response.StatusCode)
	body, _ := io.ReadAll(response.Body)

	s.Assert().Equal(`{"code":400,"message":"invalid value provided for tokenId!"}`, string(body))
}

func (s *UserIntegrationAuthControllerTestSuite) TestPersistOauthCredentials() {
	mockAuthCodeResp := &services.TeslaAuthCodeResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw",
		Expiry:       time.Now().Add(time.Hour * 1),
	}
	tokenStr, err := json.Marshal(mockAuthCodeResp)
	s.Assert().NoError(err)

	mockUserEthAddr := common.HexToAddress("1").String()

	encToken, err := s.cipher.Encrypt(string(tokenStr))
	s.Assert().NoError(err)

	cacheKey := fmt.Sprintf(teslaFleetAuthCacheKey, mockUserEthAddr)
	s.redisClient.EXPECT().Set(gomock.Any(), cacheKey, encToken, 5*time.Minute).Return(&redis.StatusCmd{})

	intCtrl := NewUserIntegrationAuthController(&config.Settings{
		Port:        "3000",
		Environment: "prod",
	}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.teslaFleetAPISvc, s.redisClient, s.cipher, s.usersClient)

	err = intCtrl.persistOauthCredentials(s.ctx, *mockAuthCodeResp, mockUserEthAddr)
	s.Assert().NoError(err)
}
