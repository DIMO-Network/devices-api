package services

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type AutoPiAPIServiceTestSuite struct {
	suite.Suite
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
}

// SetupSuite starts container db
func (s *AutoPiAPIServiceTestSuite) SetupSuite() {
	mockCtrl := gomock.NewController(s.T())
	defer mockCtrl.Finish()

	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
}

// TearDownTest after each test truncate tables
func (s *AutoPiAPIServiceTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *AutoPiAPIServiceTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func TestAutoPiApiServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AutoPiAPIServiceTestSuite))
}

func (s *AutoPiAPIServiceTestSuite) TestGetUserDeviceIntegrationByUnitID() {
	// arrange
	const testUserID = "123123"
	autoPiUnitID := "456"

	ud := test.SetupCreateUserDevice(s.T(), testUserID, ksuid.New().String(), nil, "", s.pdb)

	unit := &models.AftermarketDevice{
		Serial: autoPiUnitID,
	}

	require.NoError(s.T(), unit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()))

	apUdai := &models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: ksuid.New().String(),
		Status:        models.UserDeviceAPIIntegrationStatusActive,
		ExternalID:    null.StringFrom("autoPiDeviceID"),
		Serial:        null.StringFrom(autoPiUnitID),
	}

	err := apUdai.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	assert.NoError(s.T(), err)
	// act
	autoPiSvc := NewAutoPiAPIService(&config.Settings{AutoPiAPIToken: "fdff"}, s.pdb.DBS)
	udai, err := autoPiSvc.GetUserDeviceIntegrationByUnitID(context.Background(), autoPiUnitID)
	// assert
	require.NoError(s.T(), err)
	require.NotNilf(s.T(), udai, "user device integration must not be nil")
	assert.Equal(s.T(), testUserID, udai.R.UserDevice.UserID)
}

func (s *AutoPiAPIServiceTestSuite) TestCommandRaw() {
	// arrange
	const (
		testUserID = "123123"
		unitID     = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
		deviceID   = "device123"
		apiURL     = "https://mock.town"
		jobID      = "321"
	)
	_ = test.SetupCreateAftermarketDevice(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	// http client mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	respJSON := fmt.Sprintf(`{ "jid": "%s", "minions": ["minion"]}`, jobID)

	url := fmt.Sprintf("%s/dongle/devices/%s/execute_raw/", apiURL, deviceID)
	httpmock.RegisterResponder(http.MethodPost, url, httpmock.NewStringResponder(200, respJSON))

	autoPiSvc := NewAutoPiAPIService(&config.Settings{AutoPiAPIToken: "fdff", AutoPiAPIURL: apiURL}, s.pdb.DBS)
	// call method
	commandResponse, err := autoPiSvc.CommandRaw(context.Background(), unitID, deviceID, "command", "")
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), jobID, commandResponse.Jid)
	assert.Len(s.T(), commandResponse.Minions, 1)

	apJob, err := models.FindAutopiJob(context.Background(), s.pdb.DBS().Writer, jobID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), unitID, apJob.AutopiUnitID.String)
	assert.Equal(s.T(), "", apJob.UserDeviceID.String)
	assert.Equal(s.T(), "command", apJob.Command)
	assert.Equal(s.T(), deviceID, apJob.AutopiDeviceID)
}

func (s *AutoPiAPIServiceTestSuite) TestGetDeviceByUnitID_Should_Be_NotFound() {
	// arrange
	const (
		unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
		apiURL = "https://mock.town"
	)

	// http client mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := fmt.Sprintf("%s/dongle/devices/by_unit_id/%s/", apiURL, unitID)
	httpmock.RegisterResponder(http.MethodGet, url, httpmock.NewStringResponder(404, `{ "status": false}`))

	// act
	autoPiSvc := NewAutoPiAPIService(&config.Settings{AutoPiAPIToken: "fdff", AutoPiAPIURL: apiURL}, s.pdb.DBS)
	_, err := autoPiSvc.GetDeviceByUnitID(unitID)

	// assert
	require.ErrorIs(s.T(), err, ErrNotFound)
}
