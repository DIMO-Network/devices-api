package autopi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/shared/db"
	"github.com/testcontainers/testcontainers-go"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"go.uber.org/mock/gomock"
)

type HardwareTemplateServiceTestSuite struct {
	suite.Suite
	hardwareTemplateService HardwareTemplateService
	ap                      *mock_services.MockAutoPiAPIService
	pdb                     db.Store
	container               testcontainers.Container
	context                 context.Context
}

func (s *HardwareTemplateServiceTestSuite) SetupSuite() {
	mockCtrl := gomock.NewController(s.T())
	defer mockCtrl.Finish()
	s.context = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.context, s.T(), migrationsDirRelPath)
	logger := test.Logger()

	s.ap = mock_services.NewMockAutoPiAPIService(mockCtrl)

	s.hardwareTemplateService = NewHardwareTemplateService(s.ap, s.pdb.DBS, logger)
}

func (s *HardwareTemplateServiceTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

func (s *HardwareTemplateServiceTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.context); err != nil {
		s.T().Fatal(err)
	}
}

func TestHardwareTemplateServiceTestSuite(t *testing.T) {
	suite.Run(t, new(HardwareTemplateServiceTestSuite))
}

func (s *HardwareTemplateServiceTestSuite) Test_GetTemplateID() {
	type tableTestCases struct {
		description string
		expected    string
		ud          *models.UserDevice
		dd          *ddgrpc.GetDeviceDefinitionItemResponse
		integ       *ddgrpc.Integration
	}
	const (
		tIDIntegrationDefault = "10"
		tIDDeviceStyle        = "11"
		tIDDeviceStyleFromUD  = "111"
		tIDDeviceDef          = "12"
		tIDDeviceMake         = "13"
		tIDBEVPowertrainUD    = "14"
	)
	def, _ := strconv.Atoi(tIDIntegrationDefault)
	bev, _ := strconv.Atoi(tIDBEVPowertrainUD)
	integration := test.BuildIntegrationDefaultGRPC(constants.AutoPiVendor, def, bev, true)
	integrationWithoutAutoPiPowertrainTemplate := test.BuildIntegrationDefaultGRPC(constants.AutoPiVendor, def, 0, false)

	ddWithTID := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, "ford-f150", integration)[0]
	ddWithDeviceStyleTID := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, "ford-f150", integration)[0]
	ddWithMakeTID := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, "ford-f150", integration)[0]
	ddNoTIDs := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, "ford-f150", integration)[0]
	ddWithDeviceStyleInUD := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, "ford-f150", integration)[0]

	ddWithTID.HardwareTemplateId = tIDDeviceDef
	ddWithDeviceStyleTID.DeviceStyles = append(ddWithDeviceStyleTID.DeviceStyles, &ddgrpc.DeviceStyle{
		Id:                 ksuid.New().String(),
		HardwareTemplateId: tIDDeviceStyle,
	})
	ddWithDeviceStyleInUD.DeviceStyles = append(ddWithDeviceStyleInUD.DeviceStyles, &ddgrpc.DeviceStyle{
		Id:                 ksuid.New().String(),
		HardwareTemplateId: tIDDeviceStyleFromUD,
	})
	ddWithMakeTID.Make.HardwareTemplateId = tIDDeviceMake

	pt := services.BEV
	udmdBEVPT := services.UserDeviceMetadata{
		PowertrainType: &pt,
	}
	udmdBEVPTjson, _ := json.Marshal(udmdBEVPT)

	for _, scenario := range []tableTestCases{
		{
			description: "Should get hardware template id from style id in User Device",
			ud: &models.UserDevice{
				ID:                 ksuid.New().String(),
				UserID:             "testUserID",
				DeviceDefinitionID: ddWithDeviceStyleInUD.DeviceDefinitionId,
				CountryCode:        null.StringFrom("USA"),
				Name:               null.StringFrom("Chungus"),
				DeviceStyleID:      null.StringFrom(ddWithDeviceStyleInUD.DeviceStyles[0].Id),
			},
			integ:    integration,
			dd:       ddWithDeviceStyleInUD,
			expected: tIDDeviceStyleFromUD,
		},
		{
			description: "Should get hardware template id from DD",
			ud: &models.UserDevice{
				ID:                 ksuid.New().String(),
				UserID:             "testUserID",
				DeviceDefinitionID: ddWithTID.DeviceDefinitionId,
				CountryCode:        null.StringFrom("USA"),
				Name:               null.StringFrom("Chungus"),
			},
			integ:    integration,
			dd:       ddWithTID,
			expected: tIDDeviceDef,
		},
		{
			description: "Should NOT get template id from DD with styles when no style id in UD",
			ud: &models.UserDevice{
				ID:                 ksuid.New().String(),
				UserID:             "testUserID",
				DeviceDefinitionID: ddWithDeviceStyleTID.DeviceDefinitionId,
				CountryCode:        null.StringFrom("USA"),
				Name:               null.StringFrom("Chungus"),
			},
			integ:    integration,
			dd:       ddWithDeviceStyleTID,
			expected: tIDIntegrationDefault,
		},
		{
			description: "Should get hardware template id from Device Make in DD",
			ud: &models.UserDevice{
				ID:                 ksuid.New().String(),
				UserID:             "testUserID",
				DeviceDefinitionID: ddWithMakeTID.DeviceDefinitionId,
				CountryCode:        null.StringFrom("USA"),
				Name:               null.StringFrom("Chungus"),
			},
			integ:    integration,
			dd:       ddWithMakeTID,
			expected: tIDDeviceMake,
		},
		{
			description: "Should get hardware template id from AutoPi integration AutoPiPowertrainTemplate in UD",
			ud: &models.UserDevice{
				ID:                 ksuid.New().String(),
				UserID:             "testUserID",
				DeviceDefinitionID: ddNoTIDs.DeviceDefinitionId,
				CountryCode:        null.StringFrom("USA"),
				Name:               null.StringFrom("Chungus"),
				Metadata:           null.JSONFrom(udmdBEVPTjson),
			},
			integ:    integration,
			dd:       ddNoTIDs,
			expected: tIDBEVPowertrainUD,
		},
		{
			description: "Should get hardware template id from AutoPi DefaultTemplate",
			ud: &models.UserDevice{
				ID:                 ksuid.New().String(),
				UserID:             "testUserID",
				DeviceDefinitionID: ddNoTIDs.DeviceDefinitionId,
				CountryCode:        null.StringFrom("USA"),
				Name:               null.StringFrom("Chungus"),
			},
			integ:    integrationWithoutAutoPiPowertrainTemplate,
			dd:       ddNoTIDs,
			expected: tIDIntegrationDefault,
		},
	} {
		s.T().Run(scenario.description, func(t *testing.T) {
			id, _ := s.hardwareTemplateService.GetTemplateID(scenario.ud, scenario.dd, scenario.integ)
			assert.Equal(t, scenario.expected, id)
		})
	}
}
