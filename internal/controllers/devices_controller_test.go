package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared/db"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
)

type DevicesControllerTestSuite struct {
	suite.Suite
	pdb             db.Store
	container       testcontainers.Container
	ctx             context.Context
	mockCtrl        *gomock.Controller
	app             *fiber.App
	deviceDefSvc    *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc *mock_services.MockDeviceDefinitionIntegrationService
}

// SetupSuite starts container db
func (s *DevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	s.mockCtrl = gomock.NewController(s.T())
	logger := test.Logger()

	nhtsaSvc := mock_services.NewMockINHTSAService(s.mockCtrl)
	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	c := NewDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, nhtsaSvc, s.deviceDefSvc, s.deviceDefIntSvc)

	// routes
	app := fiber.New()
	app.Get("/device-definitions/:id", c.GetDeviceDefinitionByID)
	app.Get("/device-definitions/:id/integrations", c.GetDeviceIntegrationsByID)

	s.app = app

	// note we do not want to truncate tables after each test for this one
}

// TearDownSuite cleanup at end by terminating container
func (s *DevicesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

func TestDevicesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(DevicesControllerTestSuite))
}

/* Actual tests*/

func (s *DevicesControllerTestSuite) TestGetDeviceDefinitionById() {
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	ddGrpc := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Ford", 2020, integration)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ddGrpc[0].DeviceDefinitionId}).Times(1).Return(ddGrpc, nil)

	request, _ := http.NewRequest("GET", "/device-definitions/"+ddGrpc[0].DeviceDefinitionId, nil)
	response, errRes := s.app.Test(request)
	require.NoError(s.T(), errRes)

	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)

	var dd services.DeviceDefinition
	v := gjson.GetBytes(body, "deviceDefinition")
	err := json.Unmarshal([]byte(v.Raw), &dd)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), ddGrpc[0].DeviceDefinitionId, dd.DeviceDefinitionID)
	if assert.True(s.T(), len(dd.CompatibleIntegrations) >= 2, "should be atleast 2 integrations for autopi") {
		assert.Equal(s.T(), constants.AutoPiVendor, dd.CompatibleIntegrations[0].Vendor)
		assert.Equal(s.T(), "Americas", dd.CompatibleIntegrations[0].Region)
		assert.Equal(s.T(), constants.AutoPiVendor, dd.CompatibleIntegrations[1].Vendor)
		assert.Equal(s.T(), "Europe", dd.CompatibleIntegrations[1].Region)
	} else {
		fmt.Printf("found integrations: %+v", dd.CompatibleIntegrations)
	}
}

func (s *DevicesControllerTestSuite) TestGetDeviceDefinitionDoesNotAddAutoPiForOldCars() {
	ddOldCar := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Odlie", 1998, nil)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ddOldCar[0].DeviceDefinitionId}).Times(1).
		Return(ddOldCar, nil)

	request, _ := http.NewRequest("GET", "/device-definitions/"+ddOldCar[0].DeviceDefinitionId, nil)
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)
	v := gjson.GetBytes(body, "deviceDefinition")
	var dd services.DeviceDefinition
	err := json.Unmarshal([]byte(v.Raw), &dd)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), ddOldCar[0].DeviceDefinitionId, dd.DeviceDefinitionID)
	assert.Len(s.T(), dd.CompatibleIntegrations, 0, "vehicles before 2020 should not auto inject autopi integrations")
}

func (s *DevicesControllerTestSuite) TestGetDeviceDefinitionDoesNotAddAutoPiForTesla() {
	ddTesla := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Tesla", "Odlie", 2020, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ddTesla[0].DeviceDefinitionId}).Times(1).
		Return(ddTesla, nil)

	request, _ := http.NewRequest("GET", "/device-definitions/"+ddTesla[0].DeviceDefinitionId, nil)
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)
	v := gjson.GetBytes(body, "deviceDefinition")
	var dd services.DeviceDefinition
	err := json.Unmarshal([]byte(v.Raw), &dd)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), dd.CompatibleIntegrations, 0, "vehicles before 2012 should not auto inject autopi integrations")
}

func (s *DevicesControllerTestSuite) TestGetDeviceIntegrationsById() {
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "model etc", 2020, integration)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd[0].DeviceDefinitionId}).Return(dd, nil)

	request, _ := http.NewRequest("GET", "/device-definitions/"+dd[0].DeviceDefinitionId+"/integrations", nil)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)
	v := gjson.GetBytes(body, "compatibleIntegrations")
	var dc []services.DeviceCompatibility
	err = json.Unmarshal([]byte(v.Raw), &dc)
	assert.NoError(s.T(), err)
	if assert.True(s.T(), len(dc) >= 2, "should be atleast 2 integrations for autopi") {
		assert.Equal(s.T(), constants.AutoPiVendor, dc[0].Vendor)
		assert.Equal(s.T(), "Americas", dc[0].Region)
		assert.Equal(s.T(), constants.AutoPiVendor, dc[1].Vendor)
		assert.Equal(s.T(), "Europe", dc[1].Region)
	}
}

func (s *DevicesControllerTestSuite) TestGetDeviceDefinitionWithInvalidID() {
	request, _ := http.NewRequest("GET", "/device-definitions/caca", nil)
	response, _ := s.app.Test(request)
	// assert
	assert.Equal(s.T(), 400, response.StatusCode)
}

func (s *DevicesControllerTestSuite) TestGetDeviceDefIntegrationWithInvalidID() {
	request, _ := http.NewRequest("GET", "/device-definitions/caca/integrations", nil)
	response, _ := s.app.Test(request)
	// assert
	assert.Equal(s.T(), 400, response.StatusCode)
}

func TestNewDeviceDefinitionFromGrpc(t *testing.T) {
	subModels := []string{"AMG"}
	dbDevice := &grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "123",
		Type: &grpc.DeviceType{
			Type:      "Vehicle",
			Model:     "R500",
			Year:      2020,
			Make:      "Mercedes",
			SubModels: subModels,
		},
		DeviceAttributes: []*grpc.DeviceTypeAttribute{
			{Name: "fuel_type", Value: "gas"},
			{Name: "driven_wheels", Value: "4"},
			{Name: "number_of_doors", Value: "5"},
		},
		Make: &grpc.DeviceMake{
			Id:   "1",
			Name: "Mercedes",
		},
		DeviceIntegrations: append([]*grpc.DeviceIntegration{}, &grpc.DeviceIntegration{
			Integration: &grpc.Integration{
				Id:     "123",
				Vendor: "Autopi",
			},
		}),
		//Metadata:     null.JSONFrom([]byte(`{"vehicle_info": {"fuel_type": "gas", "driven_wheels": "4", "number_of_doors":"5" } }`)),
	}

	dd, err := NewDeviceDefinitionFromGRPC(dbDevice)

	assert.NoError(t, err)
	assert.Equal(t, "123", dd.DeviceDefinitionID)
	assert.Contains(t, dd.DeviceAttributes, services.DeviceAttribute{Name: "fuel_type", Value: "gas"})
	assert.Contains(t, dd.DeviceAttributes, services.DeviceAttribute{Name: "driven_wheels", Value: "4"})
	assert.Contains(t, dd.DeviceAttributes, services.DeviceAttribute{Name: "number_of_doors", Value: "5"})
	assert.Equal(t, "Vehicle", dd.Type.Type)
	assert.Equal(t, 2020, dd.Type.Year)
	assert.Equal(t, "Mercedes", dd.Type.Make)
	assert.Equal(t, "R500", dd.Type.Model)
	assert.Contains(t, dd.Type.SubModels, "AMG")

	assert.Len(t, dd.CompatibleIntegrations, 1)
	assert.Equal(t, "Autopi", dd.CompatibleIntegrations[0].Vendor)
}

func TestNewDeviceDefinitionFromDatabase_Error(t *testing.T) {
	dbDevice := &grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "123",
		VehicleData: &grpc.VehicleInfo{
			FuelType:      "gas",
			DrivenWheels:  "4",
			NumberOfDoors: 5,
		},
	}
	_, err := NewDeviceDefinitionFromGRPC(dbDevice)
	assert.Error(t, err)
}
