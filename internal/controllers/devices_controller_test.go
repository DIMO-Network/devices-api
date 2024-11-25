package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared/db"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

type DevicesControllerTestSuite struct {
	suite.Suite
	pdb             db.Store
	container       testcontainers.Container
	ctx             context.Context
	mockCtrl        *gomock.Controller
	deviceDefSvc    *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc *mock_services.MockDeviceDefinitionIntegrationService
}

// SetupSuite starts container db
func (s *DevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	s.mockCtrl = gomock.NewController(s.T())

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
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

func TestNewDeviceDefinitionFromGrpc(t *testing.T) {
	dbDevice := &grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "123",
		Model:              "R500",
		Year:               2020,

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
}
