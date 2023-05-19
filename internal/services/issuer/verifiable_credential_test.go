package issuer

import (
	"context"
	"encoding/base64"
	"math/big"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
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
}

const migrationsDirRelPath = "../../../migrations"

// SetupSuite starts container db
func (s *CredentialTestSuite) SetupSuite() {
	s.pdb, s.container = test.StartContainerDatabase(context.Background(), s.T(), migrationsDirRelPath)
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

	pk, err := base64.RawURLEncoding.DecodeString("2pN28-5VmEavX46XWszjasN0kx4ha3wQ6w6hGqD8o0k")
	require.NoError(s.T(), err)

	iss, err := New(Config{
		PrivateKey:        pk,
		ChainID:           big.NewInt(137),
		VehicleNFTAddress: common.HexToAddress("00f1"),
		DBS:               s.pdb,
	})
	s.Require().NoError(err)

	credentialID, err := iss.VIN(vin, tokenID)
	s.Require().NoError(err)

	vc, err := models.VerifiableCredentials(models.VerifiableCredentialWhere.ClaimID.EQ(credentialID)).One(context.Background(), s.pdb.DBS().Reader)
	require.NoError(s.T(), err)

	assert.NotEqual(s.T(), vc.Credential, []byte{})
}
