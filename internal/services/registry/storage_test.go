package registry

import (
	"context"
	"math/big"
	"testing"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type StorageTestSuite struct {
	suite.Suite
	ctx          context.Context
	dbs          db.Store
	container    testcontainers.Container
	mockCtrl     *gomock.Controller
	scTaskSvc    *mock_services.MockSmartcarTaskService
	deviceDefSvc *mock_services.MockDeviceDefinitionService

	proc StatusProcessor
}

func (s *StorageTestSuite) SetupSuite() {
	s.ctx = context.TODO()
	s.mockCtrl = gomock.NewController(s.T())
	logger := zerolog.Nop()

	s.dbs, s.container = test.StartContainerDatabase(context.TODO(), s.T(), "../../../migrations")

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.scTaskSvc = mock_services.NewMockSmartcarTaskService(s.mockCtrl)
	proc, err := NewProcessor(s.dbs.DBS, &logger, nil, &config.Settings{Environment: "prod"}, s.scTaskSvc, s.deviceDefSvc)
	if err != nil {
		s.T().Fatal(err)
	}
	s.proc = proc
}

func (s *StorageTestSuite) TearDownSuite() {
	s.Require().NoError(s.container.Terminate(context.TODO()))
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

func (s *StorageTestSuite) Test_Mint_SyntheticDevice() {
	vehicleID := int64(54)
	integrationNode := int64(1)
	childKeyNumber := 300
	syntheticDeviceAddr := common.HexToAddress("4")
	cipher := new(shared.ROT13Cipher)
	ownerAddr := common.HexToAddress("1000")
	integrationID := ksuid.New().String()

	udArgs := &models.UserDeviceAPIIntegration{}
	s.scTaskSvc.EXPECT().StartPoll(gomock.Any()).Return(nil).Do(func(arg *models.UserDeviceAPIIntegration) {
		udArgs = arg
	})

	ud := models.UserDevice{
		ID: ksuid.New().String(),
	}
	s.MustInsert(&ud)

	mtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusMined,
	}
	s.MustInsert(&mtr)

	vnft := models.VehicleNFT{
		MintRequestID: mtr.ID,
		UserDeviceID:  null.StringFrom(ud.ID),
		TokenID:       types.NewNullDecimal(decimal.New(vehicleID, 0)),
		OwnerAddress:  null.BytesFrom(ownerAddr.Bytes()),
	}
	s.MustInsert(&vnft)

	syntMtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusMined,
	}
	s.MustInsert(&syntMtr)

	vnID := types.NewDecimal(decimal.New(vehicleID, 0))
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

	integration := &ddgrpc.Integration{Id: integrationID}
	var intTokenID uint64
	s.deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), gomock.Any()).Return(integration, nil).Do(func(ct context.Context, arg uint64) {
		intTokenID = arg
	})

	a, _ := contracts.RegistryMetaData.GetAbi()

	err = s.proc.Handle(context.TODO(), &ceData{
		RequestID: syntMtr.ID,
		Type:      "Confirmed",
		Transaction: ceTx{
			Hash: "0x28db529e841dc0bc46c27a5a43ae7db8ed43294c1b97a8b81b142b8fd6763f43",
			Logs: []ceLog{
				{
					Topics: []common.Hash{
						/*
							event SyntheticDeviceNodeMinted(
								uint256 integrationNode,
								uint256 syntheticDeviceNode,
								uint256 indexed vehicleNode,
								address indexed syntheticDeviceAddress,
								address indexed owner
							)
						*/
						a.Events["SyntheticDeviceNodeMinted"].ID,
						common.BigToHash(big.NewInt(vehicleID)),
						syntheticDeviceAddr.Hash(),
						ownerAddr.Hash(),
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

	s.Equal(udi.UserDeviceID, udArgs.UserDeviceID)
	s.Equal(uint64(integrationNode), intTokenID)

	tkID := types.NewNullDecimal(decimal.New(30, 0))
	s.Equal(tkID, sd.TokenID)
}

func (s *StorageTestSuite) TestMintVehicle() {
	ctx := context.TODO()

	logger := zerolog.Nop()
	s.mockCtrl = gomock.NewController(s.T())

	proc, err := NewProcessor(s.dbs.DBS, &logger, nil, &config.Settings{Environment: "prod"}, s.scTaskSvc, s.deviceDefSvc)
	s.Require().NoError(err)

	ud := models.UserDevice{
		ID: ksuid.New().String(),
	}
	s.MustInsert(&ud)

	mtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusMined,
	}
	s.MustInsert(&mtr)

	vnft := models.VehicleNFT{
		MintRequestID: mtr.ID,
		UserDeviceID:  null.StringFrom(ud.ID),
	}
	s.MustInsert(&vnft)

	s.Require().NoError(proc.Handle(context.TODO(), &ceData{
		RequestID: mtr.ID,
		Type:      "Confirmed",
		Transaction: ceTx{
			Hash: "0x45556dbb377e6287c939d565aa785385d80a2945f2075225980b63d1488ff85b",
			Logs: []ceLog{
				{
					Topics: []common.Hash{
						// keccack256("VehicleNodeMinted(uint256,address)")
						common.HexToHash("0x09ec7fe5281be92443463e1061ce315afc1142b6c31c98a90b711012a54cc32f"),
					},
					Data: common.FromHex(
						"000000000000000000000000000000000000000000000000000000000000386b" +
							"0000000000000000000000007e74d0f663d58d12817b8bef762bcde3af1f63d6",
					),
				},
			},
		},
	}))

	s.Require().NoError(vnft.Reload(ctx, s.dbs.DBS().Writer))

	s.Zero(vnft.TokenID.Int(nil).Cmp(big.NewInt(14443)))
	s.Equal(common.HexToAddress("7e74d0f663d58d12817b8bef762bcde3af1f63d6"), common.BytesToAddress(vnft.OwnerAddress.Bytes))
}

func (s *StorageTestSuite) MustInsert(o interface {
	Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error
}) {
	s.Require().NoError(o.Insert(context.TODO(), s.dbs.DBS().Writer, boil.Infer()))
}
