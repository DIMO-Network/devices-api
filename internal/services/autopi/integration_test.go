package autopi

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"

	"math/big"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared/db"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context

	deviceDefSvc            *mock_services.MockDeviceDefinitionService
	ap                      *mock_services.MockAutoPiAPIService
	apTask                  *mock_services.MockAutoPiTaskService
	apReg                   *mock_services.MockIngestRegistrar
	eventer                 *mock_services.MockEventService
	ddRegistrar             *mock_services.MockDeviceDefinitionRegistrar
	hardwareTemplateService HardwareTemplateService
	integration             *Integration
}

const migrationsDirRelPath = "../../../migrations"

// SetupSuite starts container db
func (s *IntegrationTestSuite) SetupSuite() {
	mockCtrl := gomock.NewController(s.T())
	defer mockCtrl.Finish()
	logger := test.Logger()

	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(mockCtrl)
	s.ap = mock_services.NewMockAutoPiAPIService(mockCtrl)
	s.apTask = mock_services.NewMockAutoPiTaskService(mockCtrl)
	s.apReg = mock_services.NewMockIngestRegistrar(mockCtrl)
	s.eventer = mock_services.NewMockEventService(mockCtrl)
	s.ddRegistrar = mock_services.NewMockDeviceDefinitionRegistrar(mockCtrl)
	s.hardwareTemplateService = NewHardwareTemplateService(s.ap, s.pdb.DBS, logger)

	s.integration = NewIntegration(s.pdb.DBS, s.deviceDefSvc, s.ap, s.apTask, s.apReg, s.eventer, s.ddRegistrar, s.hardwareTemplateService, logger)
}

// TearDownTest after each test truncate tables
func (s *IntegrationTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *IntegrationTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) Test_Pair_With_DD_HardwareTemplate_Success() {
	// arrange
	const testUserID = "123123"
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	const vehicleID = 1
	const vin = ""
	const hardwareTemplateID = "1"

	deviceDefinitionID := ksuid.New().String()

	autoPiTokenID, _ := new(big.Int).SetString("0", 16)
	vehicleTokenID, _ := new(big.Int).SetString("0", 16)
	// todo: add code to test ud metadata with protocol
	md := []byte(`{"canProtocol":"06"}`)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, deviceDefinitionID, &md, "", s.pdb)

	_, apAddr, _ := test.GenerateWallet()

	autoPIUnit := test.SetupCreateMintedAutoPiUnit(s.T(), testUserID, unitID, autoPiTokenID, *apAddr, &ud.ID, s.pdb)
	vehicleNFT := test.SetupCreateVehicleNFT(s.T(), ud.ID, vin, vehicleTokenID, null.Bytes{}, s.pdb)

	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(deviceDefinitionID, "Ford", "F150", 2020, integration)

	dd[0].HardwareTemplateId = hardwareTemplateID

	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), gomock.Any()).Times(1).Return(integration, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)

	autoPiMock := &services.AutoPiDongleDevice{
		ID:     ksuid.New().String(),
		UnitID: unitID,
		IMEI:   ksuid.New().String(),
		Vehicle: services.AutoPiDongleVehicle{
			ID: vehicleID,
		},
	}

	s.ap.EXPECT().GetDeviceByUnitID(gomock.Any()).Times(1).Return(autoPiMock, nil)
	s.ap.EXPECT().PatchVehicleProfile(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	s.ap.EXPECT().AssociateDeviceToTemplate(autoPiMock.ID, gomock.Any()).Times(1).Return(nil)
	s.ap.EXPECT().ApplyTemplate(autoPiMock.ID, gomock.Any()).Times(1).Return(nil)

	autoPICommandResponseMock := &services.AutoPiCommandResponse{
		Jid: ksuid.New().String(),
	}

	s.ap.EXPECT().CommandSyncDevice(gomock.Any(), autoPiMock.UnitID, autoPiMock.ID, ud.ID).Times(1).Return(autoPICommandResponseMock, nil)
	s.apReg.EXPECT().Register(autoPiMock.UnitID, ud.ID, integration.Id).Times(1).Return(nil)
	s.apReg.EXPECT().Register2(&services.AftermarketDeviceVehicleMapping{
		AftermarketDevice: services.AftermarketDeviceVehicleMappingAftermarketDevice{
			Address:       *apAddr,
			Token:         autoPiTokenID,
			Serial:        unitID,
			IntegrationID: integration.Id,
		},
		Vehicle: services.AftermarketDeviceVehicleMappingVehicle{
			Token:        vehicleTokenID,
			UserDeviceID: ud.ID,
		},
	}).Times(1).Return(nil)

	taskID := ksuid.New().String()

	s.apTask.EXPECT().StartQueryAndUpdateVIN(autoPiMock.ID, autoPiMock.UnitID, ud.ID).Times(1).Return(taskID, nil)
	s.eventer.EXPECT().Emit(gomock.Any()).Times(1).Return(nil)

	s.ddRegistrar.EXPECT().Register(gomock.Any()).Times(1).Return(nil)

	err := s.integration.Pair(s.ctx, autoPiTokenID, vehicleTokenID)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), testUserID, ud.UserID)
	assert.Equal(s.T(), unitID, autoPIUnit.Serial)
	assert.Equal(s.T(), vin, vehicleNFT.Vin)
	udai, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.HWSerial.EQ(null.StringFrom(autoPIUnit.Serial))).
		One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "06", gjson.GetBytes(udai.Metadata.JSON, "canProtocol").String(), "canProtocol in metadata did not match expected")
}

func (s *IntegrationTestSuite) Test_Pair_With_Make_HardwareTemplate_Success() {
	// arrange
	const testUserID = "123123"
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	const vehicleID = 1
	const vin = ""
	const hardwareTemplateID = "1"

	deviceDefinitionID := ksuid.New().String()

	autoPiTokenID, _ := new(big.Int).SetString("0", 16)
	vehicleTokenID, _ := new(big.Int).SetString("0", 16)

	_, apAddr, _ := test.GenerateWallet()
	ud := test.SetupCreateUserDevice(s.T(), testUserID, deviceDefinitionID, nil, "", s.pdb)
	autoPIUnit := test.SetupCreateMintedAutoPiUnit(s.T(), testUserID, unitID, autoPiTokenID, *apAddr, &ud.ID, s.pdb)
	vehicleNFT := test.SetupCreateVehicleNFT(s.T(), ud.ID, vin, vehicleTokenID, null.Bytes{}, s.pdb)

	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(deviceDefinitionID, "Ford", "F150", 2020, integration)

	dd[0].Make.HardwareTemplateId = hardwareTemplateID

	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), gomock.Any()).Times(1).Return(integration, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)

	autoPiMock := &services.AutoPiDongleDevice{
		ID:     ksuid.New().String(),
		UnitID: unitID,
		IMEI:   ksuid.New().String(),
		Vehicle: services.AutoPiDongleVehicle{
			ID: vehicleID,
		},
	}

	s.ap.EXPECT().GetDeviceByUnitID(gomock.Any()).Times(1).Return(autoPiMock, nil)
	s.ap.EXPECT().PatchVehicleProfile(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	s.ap.EXPECT().AssociateDeviceToTemplate(autoPiMock.ID, gomock.Any()).Times(1).Return(nil)
	s.ap.EXPECT().ApplyTemplate(autoPiMock.ID, gomock.Any()).Times(1).Return(nil)

	autoPICommandResponseMock := &services.AutoPiCommandResponse{
		Jid: ksuid.New().String(),
	}

	s.ap.EXPECT().CommandSyncDevice(gomock.Any(), autoPiMock.UnitID, autoPiMock.ID, ud.ID).Times(1).Return(autoPICommandResponseMock, nil)
	s.apReg.EXPECT().Register(autoPiMock.UnitID, ud.ID, integration.Id).Times(1).Return(nil)
	s.apReg.EXPECT().Register2(&services.AftermarketDeviceVehicleMapping{
		AftermarketDevice: services.AftermarketDeviceVehicleMappingAftermarketDevice{
			Address:       *apAddr,
			Token:         autoPiTokenID,
			Serial:        unitID,
			IntegrationID: integration.Id,
		},
		Vehicle: services.AftermarketDeviceVehicleMappingVehicle{
			Token:        vehicleTokenID,
			UserDeviceID: ud.ID,
		},
	}).Times(1).Return(nil)

	taskID := ksuid.New().String()

	s.apTask.EXPECT().StartQueryAndUpdateVIN(autoPiMock.ID, autoPiMock.UnitID, ud.ID).Times(1).Return(taskID, nil)
	s.eventer.EXPECT().Emit(gomock.Any()).Times(1).Return(nil)

	s.ddRegistrar.EXPECT().Register(gomock.Any()).Times(1).Return(nil)

	err := s.integration.Pair(s.ctx, autoPiTokenID, vehicleTokenID)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testUserID, ud.UserID)
	assert.Equal(s.T(), unitID, autoPIUnit.Serial)
	assert.Equal(s.T(), vin, vehicleNFT.Vin)

}

func (s *IntegrationTestSuite) Test_Pair_With_DD_DeviceStyle_HardwareTemplate_Success() {
	// arrange
	const testUserID = "123123"
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	const vehicleID = 1
	const vin = ""
	const hardwareTemplateID = "1"

	deviceDefinitionID := ksuid.New().String()

	autoPiTokenID, _ := new(big.Int).SetString("0", 16)
	vehicleTokenID, _ := new(big.Int).SetString("0", 16)

	ud := test.SetupCreateUserDevice(s.T(), testUserID, deviceDefinitionID, nil, "", s.pdb)

	_, apAddr, _ := test.GenerateWallet()
	autoPIUnit := test.SetupCreateMintedAutoPiUnit(s.T(), testUserID, unitID, autoPiTokenID, *apAddr, &ud.ID, s.pdb)
	vehicleNFT := test.SetupCreateVehicleNFT(s.T(), ud.ID, vin, vehicleTokenID, null.Bytes{}, s.pdb)

	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(deviceDefinitionID, "Ford", "F150", 2020, integration)

	dd[0].DeviceStyles = append(dd[0].DeviceStyles, &grpc.DeviceStyle{
		Id:                 ksuid.New().String(),
		Name:               "Test",
		HardwareTemplateId: hardwareTemplateID,
		DeviceDefinitionId: deviceDefinitionID,
		Source:             "Source",
		SubModel:           "Sub-Model",
	})

	dd[0].DeviceStyles[0].HardwareTemplateId = hardwareTemplateID

	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), gomock.Any()).Times(1).Return(integration, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)

	autoPiMock := &services.AutoPiDongleDevice{
		ID:     ksuid.New().String(),
		UnitID: unitID,
		IMEI:   ksuid.New().String(),
		Vehicle: services.AutoPiDongleVehicle{
			ID: vehicleID,
		},
	}

	s.ap.EXPECT().GetDeviceByUnitID(gomock.Any()).Times(1).Return(autoPiMock, nil)
	s.ap.EXPECT().PatchVehicleProfile(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	s.ap.EXPECT().AssociateDeviceToTemplate(autoPiMock.ID, gomock.Any()).Times(1).Return(nil)
	s.ap.EXPECT().ApplyTemplate(autoPiMock.ID, gomock.Any()).Times(1).Return(nil)

	autoPICommandResponseMock := &services.AutoPiCommandResponse{
		Jid: ksuid.New().String(),
	}

	s.ap.EXPECT().CommandSyncDevice(gomock.Any(), autoPiMock.UnitID, autoPiMock.ID, ud.ID).Times(1).Return(autoPICommandResponseMock, nil)
	s.apReg.EXPECT().Register(autoPiMock.UnitID, ud.ID, integration.Id).Times(1).Return(nil)
	s.apReg.EXPECT().Register2(&services.AftermarketDeviceVehicleMapping{
		AftermarketDevice: services.AftermarketDeviceVehicleMappingAftermarketDevice{
			Address:       *apAddr,
			Token:         autoPiTokenID,
			Serial:        unitID,
			IntegrationID: integration.Id,
		},
		Vehicle: services.AftermarketDeviceVehicleMappingVehicle{
			Token:        vehicleTokenID,
			UserDeviceID: ud.ID,
		},
	}).Times(1).Return(nil)

	taskID := ksuid.New().String()

	s.apTask.EXPECT().StartQueryAndUpdateVIN(autoPiMock.ID, autoPiMock.UnitID, ud.ID).Times(1).Return(taskID, nil)
	s.eventer.EXPECT().Emit(gomock.Any()).Times(1).Return(nil)

	s.ddRegistrar.EXPECT().Register(gomock.Any()).Times(1).Return(nil)

	err := s.integration.Pair(s.ctx, autoPiTokenID, vehicleTokenID)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testUserID, ud.UserID)
	assert.Equal(s.T(), unitID, autoPIUnit.Serial)
	assert.Equal(s.T(), vin, vehicleNFT.Vin)

}

func (s *IntegrationTestSuite) Test_Pair_With_UserDeviceStyle_HardwareTemplate_Success() {
	// arrange
	const testUserID = "123123"
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	const vehicleID = 1
	const vin = ""
	const hardwareTemplateID = "1"
	deviceDefinitionID := ksuid.New().String()

	autoPiTokenID, _ := new(big.Int).SetString("0", 16)
	vehicleTokenID, _ := new(big.Int).SetString("0", 16)

	_, apAddr, _ := test.GenerateWallet()
	ud := test.SetupCreateUserDevice(s.T(), testUserID, deviceDefinitionID, nil, "", s.pdb)
	autoPIUnit := test.SetupCreateMintedAutoPiUnit(s.T(), testUserID, unitID, autoPiTokenID, *apAddr, &ud.ID, s.pdb)
	vehicleNFT := test.SetupCreateVehicleNFT(s.T(), ud.ID, vin, vehicleTokenID, null.Bytes{}, s.pdb)

	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(deviceDefinitionID, "Ford", "F150", 2020, integration)

	dd[0].HardwareTemplateId = hardwareTemplateID

	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), gomock.Any()).Times(1).Return(integration, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)

	autoPiMock := &services.AutoPiDongleDevice{
		ID:     ksuid.New().String(),
		UnitID: unitID,
		IMEI:   ksuid.New().String(),
		Vehicle: services.AutoPiDongleVehicle{
			ID: vehicleID,
		},
	}

	s.ap.EXPECT().GetDeviceByUnitID(gomock.Any()).Times(1).Return(autoPiMock, nil)
	s.ap.EXPECT().PatchVehicleProfile(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	s.ap.EXPECT().AssociateDeviceToTemplate(autoPiMock.ID, gomock.Any()).Times(1).Return(nil)
	s.ap.EXPECT().ApplyTemplate(autoPiMock.ID, gomock.Any()).Times(1).Return(nil)

	autoPICommandResponseMock := &services.AutoPiCommandResponse{
		Jid: ksuid.New().String(),
	}

	s.ap.EXPECT().CommandSyncDevice(gomock.Any(), autoPiMock.UnitID, autoPiMock.ID, ud.ID).Times(1).Return(autoPICommandResponseMock, nil)
	s.apReg.EXPECT().Register(autoPiMock.UnitID, ud.ID, integration.Id).Times(1).Return(nil)
	s.apReg.EXPECT().Register2(&services.AftermarketDeviceVehicleMapping{
		AftermarketDevice: services.AftermarketDeviceVehicleMappingAftermarketDevice{
			Address:       *apAddr,
			Token:         autoPiTokenID,
			Serial:        unitID,
			IntegrationID: integration.Id,
		},
		Vehicle: services.AftermarketDeviceVehicleMappingVehicle{
			Token:        vehicleTokenID,
			UserDeviceID: ud.ID,
		},
	}).Times(1).Return(nil)

	taskID := ksuid.New().String()

	s.apTask.EXPECT().StartQueryAndUpdateVIN(autoPiMock.ID, autoPiMock.UnitID, ud.ID).Times(1).Return(taskID, nil)
	s.eventer.EXPECT().Emit(gomock.Any()).Times(1).Return(nil)

	s.ddRegistrar.EXPECT().Register(gomock.Any()).Times(1).Return(nil)

	err := s.integration.Pair(s.ctx, autoPiTokenID, vehicleTokenID)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testUserID, ud.UserID)
	assert.Equal(s.T(), unitID, autoPIUnit.Serial)
	assert.Equal(s.T(), vin, vehicleNFT.Vin)

}

//
//func (s *IntegrationTestSuite) Test_Pair_HardwareTemplate_Exception() {
//	// arrange
//	const testUserID = "123123"
//	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
//	const vehicleID = 1
//	const vin = ""
//
//	deviceDefinitionID := ksuid.New().String()
//
//	autoPiTokenID, _ := new(big.Int).SetString("0", 16)
//	vehicleTokenID, _ := new(big.Int).SetString("0", 16)
//
//	ud := test.SetupCreateUserDevice(s.T(), testUserID, deviceDefinitionID, nil, s.pdb)
//	_ = test.SetupCreateAutoPiUnitWithToken(s.T(), testUserID, unitID, autoPiTokenID, &ud.ID, s.pdb)
//	_ = test.SetupCreateVehicleNFT(s.T(), ud.ID, vin, vehicleTokenID, s.pdb)
//
//	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 0, 0)
//	dd := test.BuildDeviceDefinitionGRPC(deviceDefinitionID, "Ford", "F150", 2020, integration)
//
//	s.deviceDefSvc.EXPECT().GetIntegrationByVendor(gomock.Any(), gomock.Any()).Times(1).Return(integration, nil)
//	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd[0].DeviceDefinitionId).Times(1).Return(dd[0], nil)
//
//	autoPiMock := &services.AutoPiDongleDevice{
//		ID:     ksuid.New().String(),
//		UnitID: unitID,
//		IMEI:   ksuid.New().String(),
//		Vehicle: services.AutoPiDongleVehicle{
//			ID: vehicleID,
//		},
//	}
//
//	s.ap.EXPECT().GetDeviceByUnitID(gomock.Any()).Times(1).Return(autoPiMock, nil)
//
//	err := s.integration.Pair(s.ctx, autoPiTokenID, vehicleTokenID)
//
//	assert.Error(s.T(), err)
//}
