package services

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/jarcoal/httpmock"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
)

const migrationsDirRelPath = "../../migrations"

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
	_ = test.SetupCreateAftermarketDevice(s.T(), testUserID, nil, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
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
	require.Error(s.T(), err)
}

//go:embed testDongleDeviceResp.json
var testDongleDeviceResp string

func TestUpdateState(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// http mock
	url := "https://mock.town/dongle/devices/1c030237-af16-492c-9020-a183dad2797b/"
	httpmock.RegisterResponder(http.MethodGet, url, httpmock.NewStringResponder(200, testDongleDeviceResp))
	httpmock.RegisterResponder(http.MethodPatch, url, httpmock.NewStringResponder(200, `{}`))

	apSvc := NewAutoPiAPIService(&config.Settings{AutoPiAPIURL: "https://mock.town"}, nil)
	err := apSvc.UpdateState("1c030237-af16-492c-9020-a183dad2797b", "Failed", "", "")
	assert.NoError(t, err)
}

func TestBuildCallName(t *testing.T) {
	callName := "supercar"
	dd := &ddgrpc.GetDeviceDefinitionItemResponse{
		Year: 2024,
		Make: &ddgrpc.DeviceMake{
			Name:     "Ford",
			NameSlug: "ford",
		},
		Model: "escape",
	}
	type args struct {
		callName *string
		dd       *ddgrpc.GetDeviceDefinitionItemResponse
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "both dd and callname",
			args: args{
				callName: &callName,
				dd:       dd,
			},
			want: "supercar:2024 ford escape",
		},
		{
			name: "only dd",
			args: args{
				dd: dd,
			},
			want: ":2024 ford escape",
		},
		{
			name: "only callname",
			args: args{
				callName: &callName,
			},
			want: "supercar",
		},
		{
			name: "neither dd or callname",
			args: args{},
			want: "any",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nm := BuildCallName(tt.args.callName, tt.args.dd)
			if tt.want == "any" {
				assert.Len(t, nm, 4)
			} else {
				if tt.want[0:1] == ":" {
					tt.want = nm[0:4] + tt.want
				}
				assert.Equalf(t, tt.want, nm, "BuildCallName(%v, %v)", tt.args.callName, tt.args.dd)
			}
		})
	}
}
