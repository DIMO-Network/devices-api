package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared/db"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type DeviceDefinitionIntegrationTestSuite struct {
	suite.Suite
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
}

// SetupSuite starts container db
func (s *DeviceDefinitionIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
}

// TearDownTest after each test truncate tables
func (s *DeviceDefinitionIntegrationTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *DeviceDefinitionIntegrationTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func TestDeviceDefinitionIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(DeviceDefinitionIntegrationTestSuite))
}

// todo: to be removed per PLA-260
//func (s *DeviceDefinitionIntegrationTestSuite) TestAppendAutoPiCompatibility() {
//	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
//	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "MachE", 2020, s.pdb)
//	var dcs []DeviceCompatibility
//	compatibility, err := AppendAutoPiCompatibility(s.ctx, dcs, dd.ID, s.pdb.dbs().Writer)
//
//	assert.NoError(s.T(), err)
//	assert.Len(s.T(), compatibility, 2)
//	all, err := models.DeviceIntegrations().All(s.ctx, s.pdb.dbs().Reader)
//	assert.NoError(s.T(), err)
//	assert.Len(s.T(), all, 2)
//
//	test.TruncateTables(s.pdb.dbs().Writer.DB, s.T())
//}
