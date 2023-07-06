package issuer

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type CredentialTestSuite struct {
	suite.Suite
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
	mockCtrl  *gomock.Controller
	topic     string
	iss       *Issuer
}

const migrationsDirRelPath = "../../../migrations"

// SetupSuite starts container db
func (s *CredentialTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(context.Background(), s.T(), migrationsDirRelPath)
	s.mockCtrl = gomock.NewController(s.T())
	s.topic = "topic.fingerprint"

	pk, err := base64.RawURLEncoding.DecodeString("2pN28-5VmEavX46XWszjasN0kx4ha3wQ6w6hGqD8o0k")
	require.NoError(s.T(), err)

	gitSha1 := os.Getenv("GIT_SHA1")
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Str("git-sha1", gitSha1).
		Logger()

	iss, err := New(Config{
		PrivateKey:        pk,
		ChainID:           big.NewInt(137),
		VehicleNFTAddress: common.HexToAddress("00f1"),
		DBS:               s.pdb,
	},
		&logger)
	s.iss = iss

	s.Require().NoError(err)
}

// TearDownSuite cleanup at end by terminating container
func (s *CredentialTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

func TestCredentialTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialTestSuite))
}

func (s *CredentialTestSuite) TestVerifiableCredential() {
	ctx := context.Background()
	vin := "1G6AL1RY2K0111939"
	tokenID := big.NewInt(3)
	userDeviceID := "userDeviceID1"
	mtxReq := ksuid.New().String()
	deviceID := ksuid.New().String()

	udd := models.UserDevice{
		ID:                 deviceID,
		UserID:             userDeviceID,
		DeviceDefinitionID: "deviceDefID",
	}
	err := udd.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)

	tx := models.MetaTransactionRequest{
		ID:     mtxReq,
		Status: "Confirmed",
	}
	err = tx.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)

	nft := models.VehicleNFT{
		MintRequestID: mtxReq,
		UserDeviceID:  null.StringFrom(deviceID),
		Vin:           vin,
		TokenID:       types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
		OwnerAddress:  null.BytesFrom(common.Hex2Bytes("ab8438a18d83d41847dffbdc6101d37c69c9a2fc")),
	}

	err = nft.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)

	s.Require().NoError(err)

	credentialID, err := s.iss.VIN(vin, tokenID)
	s.Require().NoError(err)

	vc, err := models.VerifiableCredentials(models.VerifiableCredentialWhere.ClaimID.EQ(credentialID)).One(context.Background(), s.pdb.DBS().Reader)
	require.NoError(s.T(), err)

	assert.NotEqual(s.T(), vc.Credential, []byte{})
}
