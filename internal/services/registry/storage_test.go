package registry

import (
	"context"
	"math/big"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type StorageTestSuite struct {
	suite.Suite
	dbs       db.Store
	container testcontainers.Container
}

func (s *StorageTestSuite) SetupSuite() {
	s.dbs, s.container = test.StartContainerDatabase(context.TODO(), s.T(), "../../../migrations")
}

func (s *StorageTestSuite) TearDownSuite() {
	s.container.Terminate(context.TODO())
}

func (s *StorageTestSuite) TearDownTest() {
	test.TruncateTables(s.dbs.DBS().Writer.DB, s.T())
}

func (s *StorageTestSuite) TestMintVehicle() {
	ctx := context.TODO()

	logger := zerolog.Nop()
	proc, err := NewProcessor(s.dbs.DBS, &logger, nil, nil, &config.Settings{Environment: "prod"})
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
