package registry

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.uber.org/mock/gomock"
)

type StorageTestSuite struct {
	suite.Suite
	ctx              context.Context
	dbs              db.Store
	container        testcontainers.Container
	mockCtrl         *gomock.Controller
	scTaskSvc        *mock_services.MockSmartcarTaskService
	teslaTaskService *mock_services.MockTeslaTaskService
	deviceDefSvc     *mock_services.MockDeviceDefinitionService
	eventSvc         *mock_services.MockEventService

	proc StatusProcessor
}

const migrationsDirRelPath = "../../../migrations"

func (s *StorageTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.dbs, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	logger := test.Logger()
	s.mockCtrl = gomock.NewController(s.T())

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.scTaskSvc = mock_services.NewMockSmartcarTaskService(s.mockCtrl)
	s.teslaTaskService = mock_services.NewMockTeslaTaskService(s.mockCtrl)
	s.eventSvc = mock_services.NewMockEventService(s.mockCtrl)
	proc, err := NewProcessor(s.dbs.DBS, logger, nil, &config.Settings{Environment: "prod"}, s.eventSvc)
	if err != nil {
		s.T().Fatal(err)
	}
	s.proc = proc
}

func (s *StorageTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())

	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

func (s *StorageTestSuite) TearDownTest() {
	test.TruncateTables(s.dbs.DBS().Writer.DB, s.T())
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (s *StorageTestSuite) Test_SyntheticMintSetsID() {
	vehicleID := int64(54)
	integrationNode := int64(1)
	childKeyNumber := 300
	syntheticDeviceAddr := common.HexToAddress("4")
	cipher := new(shared.ROT13Cipher)
	ownerAddr := common.HexToAddress("1000")
	integrationID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusMined,
	}
	s.MustInsert(&mtr)

	ud := models.UserDevice{
		ID:            ksuid.New().String(),
		MintRequestID: null.StringFrom(mtr.ID),
		TokenID:       types.NewNullDecimal(decimal.New(vehicleID, 0)),
		OwnerAddress:  null.BytesFrom(ownerAddr.Bytes()),
	}
	s.MustInsert(&ud)

	syntMtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusMined,
	}
	s.MustInsert(&syntMtr)

	vnID := types.NewNullDecimal(decimal.New(vehicleID, 0))
	syntheticDevice := models.SyntheticDevice{
		VehicleTokenID:     vnID,
		IntegrationTokenID: types.NewDecimal(decimal.New(integrationNode, 0)),
		WalletChildNumber:  childKeyNumber,
		WalletAddress:      syntheticDeviceAddr.Bytes(),
		MintRequestID:      syntMtr.ID,
	}
	s.MustInsert(&syntheticDevice)

	acToken, err := cipher.Encrypt("mockAccessToken")
	s.NoError(err)
	refToken, err := cipher.Encrypt("mockRefreshToken")
	s.NoError(err)

	udi := models.UserDeviceAPIIntegration{
		IntegrationID:   integrationID,
		UserDeviceID:    ud.ID,
		Status:          models.UserDeviceAPIIntegrationStatusPending,
		AccessToken:     null.StringFrom(acToken),
		AccessExpiresAt: null.TimeFrom(time.Now()),
		RefreshToken:    null.StringFrom(refToken),
	}
	s.MustInsert(&udi)

	a, _ := contracts.RegistryMetaData.GetAbi()

	err = s.proc.Handle(context.TODO(), &ceData{
		RequestID: syntMtr.ID,
		Type:      "Confirmed",
		Transaction: ceTx{
			Hash: "0x28db529e841dc0bc46c27a5a43ae7db8ed43294c1b97a8b81b142b8fd6763f43",
			Logs: []ceLog{
				{
					Topics: []common.Hash{
						a.Events["SyntheticDeviceNodeMinted"].ID,
						common.BigToHash(big.NewInt(vehicleID)),
						common.BytesToHash(syntheticDeviceAddr.Bytes()),
						common.BytesToHash(ownerAddr.Bytes()),
					},
					Data: common.FromHex(
						"0000000000000000000000000000000000000000000000000000000000000001" +
							"000000000000000000000000000000000000000000000000000000000000001e",
					),
				},
			},
		},
	})
	s.NoError(err)

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.MintRequestID.EQ(syntMtr.ID),
		qm.Load(models.SyntheticDeviceRels.VehicleToken),
	).One(s.ctx, s.dbs.DBS().Reader)
	s.NoError(err)

	tkID := types.NewNullDecimal(decimal.New(30, 0))
	s.Equal(tkID, sd.TokenID)
}

func (s *StorageTestSuite) TestMintVehicle() {
	ctx := context.TODO()

	logger := zerolog.Nop()
	s.mockCtrl = gomock.NewController(s.T())

	proc, err := NewProcessor(s.dbs.DBS, &logger, nil, &config.Settings{Environment: "prod"}, s.eventSvc)
	s.Require().NoError(err)

	mtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusMined,
	}
	s.MustInsert(&mtr)

	var emEv *shared.CloudEvent[any]
	s.eventSvc.EXPECT().Emit(gomock.Any()).Do(func(event *shared.CloudEvent[any]) {
		emEv = event
	})

	ud := models.UserDevice{
		ID:            ksuid.New().String(),
		MintRequestID: null.StringFrom(mtr.ID),
	}
	s.MustInsert(&ud)

	s.Require().NoError(proc.Handle(context.TODO(), &ceData{
		RequestID: mtr.ID,
		Type:      "Confirmed",
		Transaction: ceTx{
			Hash: "0x45556dbb377e6287c939d565aa785385d80a2945f2075225980b63d1488ff85b",
			Logs: []ceLog{
				{
					Topics: []common.Hash{
						// keccack256("VehicleNodeMinted(uint256,address)")
						common.HexToHash("0xd471ae8ab3c01edc986909c344bb50f982b21772fcac173103ef8b9924375ec6"),
					},
					Data: common.FromHex(
						"000000000000000000000000000000000000000000000000000000000000386b" +
							"000000000000000000000000000000000000000000000000000000000000386b" +
							"0000000000000000000000007e74d0f663d58d12817b8bef762bcde3af1f63d6",
					),
				},
			},
		},
	}))

	s.Require().NoError(ud.Reload(ctx, s.dbs.DBS().Writer))

	s.Zero(ud.TokenID.Int(nil).Cmp(big.NewInt(14443)))
	s.Equal(common.HexToAddress("7e74d0f663d58d12817b8bef762bcde3af1f63d6"), common.BytesToAddress(ud.OwnerAddress.Bytes))

	s.Equal(ud.ID, emEv.Subject)
}

func (s *StorageTestSuite) MustInsert(o boilInsertable) {
	s.Require().NoError(o.Insert(context.TODO(), s.dbs.DBS().Writer, boil.Infer()))
}

type boilInsertable interface {
	Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error
}
