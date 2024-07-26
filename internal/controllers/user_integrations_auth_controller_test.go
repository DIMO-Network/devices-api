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
	"github.com/DIMO-Network/devices-api/internal/services/tmpcred"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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
	teslaFleetAPISvc *mock_services.MockTeslaFleetAPIService
	userAddr         common.Address
	credStore        *MockCredStore
}

// SetupSuite starts container db
func (s *UserIntegrationAuthControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	logger := test.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl

	s.credStore = NewMockCredStore(mockCtrl)
	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(mockCtrl)
	s.teslaFleetAPISvc = mock_services.NewMockTeslaFleetAPIService(mockCtrl)
	s.testUserID = "123123"
	c := NewUserIntegrationAuthController(&config.Settings{
		Port:        "3000",
		Environment: "prod",
	}, s.pdb.DBS, logger, s.deviceDefSvc, s.teslaFleetAPISvc, s.credStore)
	app := test.SetupAppFiber(*logger)
	s.userAddr = common.HexToAddress("1")
	app.Post("/integration/:tokenID/credentials", func(c *fiber.Ctx) error {
		// TODO(elffjs): Yes, yes, this is bad.
		c.Locals("ethereumAddress", s.userAddr)
		return c.Next()
	}, c.CompleteOAuthExchange)

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

	signingKey := []byte("xdd")

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"scp": []string{"vehicle_device_data", "vehicle_cmds", "vehicle_charging_cmds"},
	}).SignedString(signingKey)
	s.Require().NoError(err)

	mockAuthCodeResp := &services.TeslaAuthCodeResponse{
		AccessToken:  accessToken,
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw",
		Expiry:       time.Now().Add(time.Hour * 1),
		Region:       mockRegion,
	}
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), uint64(2)).Return(&ddgrpc.Integration{
		Vendor: constants.TeslaVendor,
	}, nil)
	s.teslaFleetAPISvc.EXPECT().CompleteTeslaAuthCodeExchange(gomock.Any(), mockAuthCode, mockRedirectURI).Return(mockAuthCodeResp, nil)
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

	s.credStore.EXPECT().Store(gomock.Any(), s.userAddr, &tmpcred.Credential{
		IntegrationID: 2,
		AccessToken:   mockAuthCodeResp.AccessToken,
		RefreshToken:  mockAuthCodeResp.RefreshToken,
		Expiry:        mockAuthCodeResp.Expiry,
	}).Return(nil)

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
	s.teslaFleetAPISvc.EXPECT().GetVehicles(gomock.Any(), mockAuthCodeResp.AccessToken).Return(resp, nil)

	request := test.BuildRequest("POST", "/integration/2/credentials", fmt.Sprintf(`{
		"authorizationCode": "%s",
		"redirectUri": "%s",
		"region": "%s"
	}`, mockAuthCode, mockRedirectURI, mockRegion))
	response, err := s.app.Test(request)
	s.Require().NoError(err)

	s.Equal(fiber.StatusOK, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	s.JSONEq(string(expected), string(body))
}

func (s *UserIntegrationAuthControllerTestSuite) TestMissingScope() {
	mockAuthCode := "Mock_fd941f8da609db8cd66b1734f84ab289e2975b1889a5bedf478f02cf0cc4"
	mockRedirectURI := "https://mock-redirect.test.dimo.zone"
	mockRegion := "na"

	signingKey := []byte("xdd")

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"scp": []string{"vehicle_cmds", "vehicle_charging_cmds"},
	}).SignedString(signingKey)
	s.Require().NoError(err)

	mockAuthCodeResp := &services.TeslaAuthCodeResponse{
		AccessToken:  accessToken,
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UWfqdcCvyzObpI2gaIGcx2r7CcDjlQ0IzGyk8N0_vqw",
		Expiry:       time.Now().Add(time.Hour * 1),
		Region:       mockRegion,
	}
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), uint64(2)).Return(&ddgrpc.Integration{
		Vendor: constants.TeslaVendor,
	}, nil)
	s.teslaFleetAPISvc.EXPECT().CompleteTeslaAuthCodeExchange(gomock.Any(), mockAuthCode, mockRedirectURI).Return(mockAuthCodeResp, nil)

	request := test.BuildRequest("POST", "/integration/2/credentials", fmt.Sprintf(`{
		"authorizationCode": "%s",
		"redirectUri": "%s",
		"region": "%s"
	}`, mockAuthCode, mockRedirectURI, mockRegion))
	response, _ := s.app.Test(request)
	defer response.Body.Close()

	s.Equal(fiber.StatusBadRequest, response.StatusCode)

	body, err := io.ReadAll(response.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "vehicle_device_data")
}

func (s *UserIntegrationAuthControllerTestSuite) TestMissingRefreshToken() {
	mockAuthCode := "Mock_fd941f8da609db8cd66b1734f84ab289e2975b1889a5bedf478f02cf0cc4"
	mockRedirectURI := "https://mock-redirect.test.dimo.zone"
	mockRegion := "na"

	signingKey := []byte("xdd")

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"scp": []string{"vehicle_device_data", "vehicle_cmds", "vehicle_charging_cmds"},
	}).SignedString(signingKey)
	s.Require().NoError(err)

	mockAuthCodeResp := &services.TeslaAuthCodeResponse{
		AccessToken:  accessToken,
		RefreshToken: "",
		Expiry:       time.Now().Add(time.Hour * 1),
		Region:       mockRegion,
	}
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), uint64(2)).Return(&ddgrpc.Integration{
		Vendor: constants.TeslaVendor,
	}, nil)
	s.teslaFleetAPISvc.EXPECT().CompleteTeslaAuthCodeExchange(gomock.Any(), mockAuthCode, mockRedirectURI).Return(mockAuthCodeResp, nil)

	request := test.BuildRequest("POST", "/integration/2/credentials", fmt.Sprintf(`{
		"authorizationCode": "%s",
		"redirectUri": "%s",
		"region": "%s"
	}`, mockAuthCode, mockRedirectURI, mockRegion))
	response, _ := s.app.Test(request)
	defer response.Body.Close()

	s.Equal(fiber.StatusBadRequest, response.StatusCode)

	body, err := io.ReadAll(response.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "offline_access")
}

func (s *UserIntegrationAuthControllerTestSuite) TestCompleteOAuthExchange_UnprocessableTokenID() {
	request := test.BuildRequest("POST", "/integration/wrongTokenID/credentials", fmt.Sprintf(`{
		"authorizationCode": "%s",
		"redirectUri": "%s",
		"region": "us-central"
	}`, "mockAuthCode", "mockRedirectURI"))
	response, _ := s.app.Test(request)

	s.Assert().Equal(fiber.StatusBadRequest, response.StatusCode)
}

func (s *UserIntegrationAuthControllerTestSuite) TestCompleteOAuthExchange_InvalidTokenID() {
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
}
