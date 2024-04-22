package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"testing"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/ericlagergren/decimal"
	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.uber.org/mock/gomock"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
)

type UserDevicesControllerV2Suite struct {
	suite.Suite
	pdb             db.Store
	controller      *UserDevicesControllerV2
	container       testcontainers.Container
	ctx             context.Context
	mockCtrl        *gomock.Controller
	app             *fiber.App
	deviceDefSvc    *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc *mock_services.MockDeviceDefinitionIntegrationService
	testUserID      string
}

// ClaimsInjectorTestHandler injects fake claims into context
func ClaimsInjectorTestHandler(claims []privileges.Privilege) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := privilegetoken.CustomClaims{
			PrivilegeIDs: claims,
		}
		c.Locals("tokenClaims", claims)
		return c.Next()
	}
}

func (s *UserDevicesControllerV2Suite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	logger := test.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(mockCtrl)

	c := NewUserDevicesControllerV2(&config.Settings{Port: "3000", Environment: "prod"}, s.pdb.DBS, logger, s.deviceDefSvc)
	app := test.SetupAppFiber(*logger)
	app.Use(ClaimsInjectorTestHandler([]privileges.Privilege{privileges.VehicleNonLocationData}))
	app.Get("/v2/vehicles/:tokenId/analytics/range", test.AuthInjectorTestHandler(s.testUserID), c.GetRange)
	s.controller = &c
	s.app = app
}

// TearDownTest after each test truncate tables
func (s *UserDevicesControllerV2Suite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserDevicesControllerV2Suite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish() // might need to do mockctrl on every test, and refactor setup into one method
}

// Test Runner
func TestUserDevicesControllerTestSuiteV2(t *testing.T) {
	suite.Run(t, new(UserDevicesControllerV2Suite))
}

//go:embed test_user_device_data.json
var testUserDeviceData []byte

func (s *UserDevicesControllerV2Suite) TestGetRange() {
	autoPiUnitID := "1234"
	autoPiDeviceID := "4321"
	tokenID := 4
	ddID := ksuid.New().String()
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	smartCarIntegration := test.BuildIntegrationGRPC(constants.SmartCarVendor, 10, 0)
	_ = test.SetupCreateAftermarketDevice(s.T(), testUserID, nil, autoPiUnitID, &autoPiDeviceID, s.pdb)
	_, addr, err := test.GenerateWallet()
	s.NoError(err)

	gddir := []*ddgrpc.GetDeviceDefinitionItemResponse{
		{
			DeviceAttributes: []*ddgrpc.DeviceTypeAttribute{
				{Name: "mpg", Value: "38.0"},
				{Name: "mpg_highway", Value: "40.0"},
				{Name: "fuel_tank_capacity_gal", Value: "14.5"},
			},
			Make: &ddgrpc.DeviceMake{
				Name: "Ford",
			},
			Name:               "F-150",
			DeviceDefinitionId: ddID,
		},
	}
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, ddID, nil, "", s.pdb)

	mint := models.MetaTransactionRequest{ID: ksuid.New().String(), Status: models.MetaTransactionRequestStatusConfirmed}
	s.Require().NoError(mint.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()))

	vnft := models.VehicleNFT{
		UserDeviceID:  null.StringFrom(ud.ID),
		Vin:           ud.VinIdentifier.String,
		TokenID:       types.NewNullDecimal(decimal.New(int64(tokenID), 0)),
		OwnerAddress:  null.BytesFrom(addr.Bytes()),
		MintRequestID: mint.ID,
	}
	s.Require().NoError(vnft.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()))

	test.SetupCreateUserDeviceAPIIntegration(s.T(), autoPiUnitID, autoPiDeviceID, ud.ID, integration.Id, s.pdb)
	udd := models.UserDeviceDatum{
		UserDeviceID:  ud.ID,
		Signals:       null.JSONFrom(testUserDeviceData),
		IntegrationID: integration.Id,
	}
	err = udd.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	udd2 := models.UserDeviceDatum{
		UserDeviceID:  ud.ID,
		Signals:       null.JSONFrom([]byte(`{"range": {"value": 380.14,"timestamp":"2022-06-18T04:02:11.544Z" } }`)),
		IntegrationID: smartCarIntegration.Id,
	}
	err = udd2.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ddID}).Return(gddir, nil)

	request := test.BuildRequest(http.MethodGet, fmt.Sprintf("/v2/vehicles/%d/analytics/range", tokenID), "")

	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)

	s.Assert().Equal(fiber.StatusOK, response.StatusCode)

	s.Assert().Equal(3, int(gjson.GetBytes(body, "rangeSets.#").Int()))
	s.Assert().Equal("2022-06-18T04:06:40Z", gjson.GetBytes(body, "rangeSets.0.updated").String())
	s.Assert().Equal("2022-06-18T04:06:40Z", gjson.GetBytes(body, "rangeSets.1.updated").String())
	s.Assert().Equal("2022-06-18T04:02:11Z", gjson.GetBytes(body, "rangeSets.2.updated").String())
	s.Assert().Equal("MPG", gjson.GetBytes(body, "rangeSets.0.rangeBasis").String())
	s.Assert().Equal("MPG Highway", gjson.GetBytes(body, "rangeSets.1.rangeBasis").String())
	s.Assert().Equal("Vehicle Reported", gjson.GetBytes(body, "rangeSets.2.rangeBasis").String())
	s.Assert().Equal(391, int(gjson.GetBytes(body, "rangeSets.0.rangeDistance").Int()))
	s.Assert().Equal(411, int(gjson.GetBytes(body, "rangeSets.1.rangeDistance").Int()))
	s.Assert().Equal(236, int(gjson.GetBytes(body, "rangeSets.2.rangeDistance").Int()))
	s.Assert().Equal("miles", gjson.GetBytes(body, "rangeSets.0.rangeUnit").String())
	s.Assert().Equal("miles", gjson.GetBytes(body, "rangeSets.1.rangeUnit").String())
	s.Assert().Equal("miles", gjson.GetBytes(body, "rangeSets.2.rangeUnit").String())
}
