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

	err = ud.Insert(ctx, s.dbs.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	mtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusMined,
	}
	err = mtr.Insert(ctx, s.dbs.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	vnft := models.VehicleNFT{
		MintRequestID: mtr.ID,
		UserDeviceID:  null.StringFrom(ud.ID),
	}

	err = vnft.Insert(ctx, s.dbs.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	err = proc.Handle(context.TODO(), &ceData{
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
	})
	s.Require().NoError(err)

	err = vnft.Reload(ctx, s.dbs.DBS().Writer)
	s.Require().NoError(err)

	s.Zero(vnft.TokenID.Int(nil).Cmp(big.NewInt(14443)))
	s.Equal(common.HexToAddress("7e74d0f663d58d12817b8bef762bcde3af1f63d6"), common.BytesToAddress(vnft.OwnerAddress.Bytes))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
