package fingerprint

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services/issuer"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.uber.org/mock/gomock"
)

type ConsumerTestSuite struct {
	suite.Suite
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
	mockCtrl  *gomock.Controller
	topic     string
	iss       *issuer.Issuer
	cons      *Consumer
}

const migrationsDirRelPath = "../../../migrations"

// SetupSuite starts container db
func (s *ConsumerTestSuite) SetupSuite() {
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

	iss, err := issuer.New(issuer.Config{
		PrivateKey:        pk,
		ChainID:           big.NewInt(137),
		VehicleNFTAddress: common.HexToAddress("00f1"),
		DBS:               s.pdb,
	},
		&logger)
	s.iss = iss
	s.cons = &Consumer{
		logger: &logger,
		iss:    iss,
		DBS:    s.pdb,
	}

	s.Require().NoError(err)
}

// TearDownSuite cleanup at end by terminating container
func (s *ConsumerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

func TestConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(ConsumerTestSuite))
}

func (s *ConsumerTestSuite) TestVinCredentialerHandler_DeviceFingerprint() {
	deviceID := ksuid.New().String()
	ownerAddress := null.BytesFrom(common.Hex2Bytes("448cF8Fd88AD914e3585401241BC434FbEA94bbb"))
	vin := "W1N2539531F907299"
	ctx := context.Background()
	tokenID := big.NewInt(3)
	userDeviceID := "userDeviceID1"
	mtxReq := ksuid.New().String()
	deiceDefID := "deviceDefID"
	claimID := "claimID1"

	// tables used in tests
	aftermarketDevice := models.AftermarketDevice{
		UserID:          null.StringFrom("SomeID"),
		OwnerAddress:    ownerAddress,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		TokenID:         types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(13), 0)),
		VehicleTokenID:  types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
		Beneficiary:     null.BytesFrom(common.BytesToAddress([]byte{uint8(1)}).Bytes()),
		EthereumAddress: ownerAddress.Bytes,
	}

	userDevice := models.UserDevice{
		ID:                 deviceID,
		UserID:             userDeviceID,
		DeviceDefinitionID: deiceDefID,
		VinConfirmed:       true,
		VinIdentifier:      null.StringFrom(vin),
	}

	metaTx := models.MetaTransactionRequest{
		ID:     mtxReq,
		Status: models.MetaTransactionRequestStatusConfirmed,
	}

	credential := models.VerifiableCredential{
		ClaimID:        claimID,
		Credential:     []byte{},
		ExpirationDate: time.Now().AddDate(0, 0, 7),
	}

	nft := models.VehicleNFT{
		MintRequestID: mtxReq,
		UserDeviceID:  null.StringFrom(deviceID),
		Vin:           vin,
		TokenID:       types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
		OwnerAddress:  ownerAddress,
		ClaimID:       null.StringFrom(claimID),
	}

	msg :=
		`{
	"data": {"rpiUptimeSecs":39,"batteryVoltage":13.49,"timestamp":1688136702634,"vin":"W1N2539531F907299","protocol":"7"},
	"id": "2RvhwjUbtoePjmXN7q9qfjLQgwP",
	"signature": "7c31e54ddcffc2a548ccaf10ed64b7e4bdd239bbaa3e5f6dba41d3e4051d930b7fbdf184724c2fb8d3b2ac8ac82662d2ed74e881dd01c09c4b2a9b4e62ede5db1b",
	"source": "aftermarket/device/fingerprint",
	"specversion": "1.0",
	"subject": "0x448cF8Fd88AD914e3585401241BC434FbEA94bbb",
	"type": "zone.dimo.aftermarket.device.fingerprint"
}`

	cases := []struct {
		Name              string
		ReturnsError      bool
		ExpectedResponse  string
		UserDeviceTable   models.UserDevice
		MetaTxTable       models.MetaTransactionRequest
		VCTable           models.VerifiableCredential
		VehicleNFT        models.VehicleNFT
		AftermarketDevice models.AftermarketDevice
	}{
		{
			Name:             "No corresponding aftermarket device for address",
			ReturnsError:     true,
			ExpectedResponse: "sql: no rows in result set",
		},
		{
			Name:              "active credential",
			ReturnsError:      false,
			UserDeviceTable:   userDevice,
			MetaTxTable:       metaTx,
			VCTable:           credential,
			VehicleNFT:        nft,
			AftermarketDevice: aftermarketDevice,
		},
		{
			Name:            "inactive credential",
			ReturnsError:    false,
			UserDeviceTable: userDevice,
			MetaTxTable:     metaTx,
			VCTable: models.VerifiableCredential{
				ClaimID:        claimID,
				Credential:     []byte{},
				ExpirationDate: time.Now().AddDate(0, 0, -10),
			},
			VehicleNFT:        nft,
			AftermarketDevice: aftermarketDevice,
		},
		{
			Name:            "invalid token id",
			ReturnsError:    false,
			UserDeviceTable: userDevice,
			MetaTxTable:     metaTx,
			VCTable:         credential,
			VehicleNFT:      nft,
			AftermarketDevice: models.AftermarketDevice{
				UserID:          null.StringFrom("SomeID"),
				OwnerAddress:    ownerAddress,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				TokenID:         types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(13), 0)),
				Beneficiary:     null.BytesFrom(common.BytesToAddress([]byte{uint8(1)}).Bytes()),
				EthereumAddress: ownerAddress.Bytes,
			},
		},
	}

	for _, c := range cases {
		s.T().Run(c.Name, func(t *testing.T) {
			test.TruncateTables(s.pdb.DBS().Writer.DB, t)
			err := c.UserDeviceTable.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
			require.NoError(t, err)

			err = c.MetaTxTable.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
			require.NoError(t, err)

			err = c.VCTable.Insert(ctx, s.pdb.DBS().Reader, boil.Infer())
			require.NoError(t, err)

			err = c.VehicleNFT.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
			require.NoError(t, err)

			err = c.AftermarketDevice.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
			require.NoError(t, err)

			var event Event
			err = json.Unmarshal([]byte(msg), &event)
			require.NoError(t, err)
			err = s.cons.HandleDeviceFingerprint(s.ctx, &event)

			if c.ReturnsError {
				assert.ErrorContains(t, err, c.ExpectedResponse)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func (s *ConsumerTestSuite) TestVinCredentialerHandler_SyntheticFingerprint() {
	ctx := context.Background()
	userDeviceID := "userDeviceID1"
	userID := "userID6"
	vin := "W1N2539531F907299"
	claimID := "claimID1"
	tokenID := big.NewInt(3)
	mtxReq := ksuid.New().String()
	walletAddr := null.BytesFrom(common.FromHex("0x5d25D4891fdb93DFb88f8F9AAB66F6d2f714eD8f"))
	ownerAddr := null.BytesFrom(common.FromHex("0x6e15D4891fdb93DFb88f8F9AAB66F6d2f714eD8f"))

	metaTx := models.MetaTransactionRequest{
		ID:     mtxReq,
		Status: models.MetaTransactionRequestStatusConfirmed,
	}

	userDevice := models.UserDevice{
		ID:                 userDeviceID,
		UserID:             userDeviceID,
		DeviceDefinitionID: userID,
		VinConfirmed:       true,
		VinIdentifier:      null.StringFrom(vin),
	}

	eventTime, err := time.Parse(time.RFC3339Nano, "2023-07-04T00:00:00Z")
	s.Require().NoError(err)

	credential := models.VerifiableCredential{
		ClaimID:        claimID,
		Credential:     []byte{},
		ExpirationDate: eventTime.AddDate(0, 0, 7),
	}

	nft := models.VehicleNFT{
		MintRequestID: mtxReq,
		UserDeviceID:  null.StringFrom(userDeviceID),
		Vin:           vin,
		TokenID:       types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
		OwnerAddress:  ownerAddr,
		ClaimID:       null.StringFrom(claimID),
	}

	synthDevice := models.SyntheticDevice{
		WalletAddress:  walletAddr.Bytes,
		MintRequestID:  metaTx.ID,
		VehicleTokenID: nft.TokenID,
	}

	msg := fmt.Sprintf(`{
	"data": {"vin":%q,"week":7},
	"id": "2RvhwjUbtoePjmXN7q9qfjLQgwP",
	"source": "aftermarket/synthetic/fingerprint",
	"specversion": "1.0",
	"subject": %q,
	"time": %q,
	"type": "zone.dimo.aftermarket.synthetic.fingerprint"
}`, vin, userDeviceID, eventTime.Format(time.RFC3339))

	cases := []struct {
		Name                 string
		ReturnsError         bool
		ExpectedResponse     string
		SyntheticDeviceTable *models.SyntheticDevice
		MetaTxTable          *models.MetaTransactionRequest
		VCTable              *models.VerifiableCredential
		VehicleNFT           *models.VehicleNFT
		AftermarketDevice    *models.AftermarketDevice
		UserDeviceTable      *models.UserDevice
		ExpiresAt            time.Time
	}{
		{
			Name:         "No corresponding device for id",
			ReturnsError: true,
		},
		{
			Name:                 "active credential",
			ReturnsError:         false,
			SyntheticDeviceTable: &synthDevice,
			MetaTxTable:          &metaTx,
			VCTable:              &credential,
			UserDeviceTable:      &userDevice,
			VehicleNFT:           &nft,
			ExpiresAt:            credential.ExpirationDate,
		},
		{
			Name:                 "inactive credential",
			ReturnsError:         false,
			SyntheticDeviceTable: &synthDevice,
			MetaTxTable:          &metaTx,
			UserDeviceTable:      &userDevice,
			VCTable: &models.VerifiableCredential{
				ClaimID:        claimID,
				Credential:     []byte{},
				ExpirationDate: eventTime.AddDate(0, 0, -10),
			},
			VehicleNFT: &nft,
			ExpiresAt:  eventTime.AddDate(0, 0, 8),
		},
	}

	for _, c := range cases {
		s.T().Run(c.Name, func(t *testing.T) {
			test.TruncateTables(s.pdb.DBS().Writer.DB, t)

			if c.UserDeviceTable != nil {
				err := c.UserDeviceTable.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
				require.NoError(t, err)
			}

			if c.MetaTxTable != nil {
				err := c.MetaTxTable.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
				require.NoError(t, err)
			}

			if c.VCTable != nil {
				err := c.VCTable.Insert(ctx, s.pdb.DBS().Reader, boil.Infer())
				require.NoError(t, err)
			}

			if c.VehicleNFT != nil {
				err := c.VehicleNFT.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
				require.NoError(t, err)
			}

			if c.SyntheticDeviceTable != nil {
				err := c.SyntheticDeviceTable.Insert(ctx, s.pdb.DBS().Writer, boil.Infer())
				require.NoError(t, err)
			}

			var event Event
			err := json.Unmarshal([]byte(msg), &event)
			require.NoError(t, err)

			err = s.cons.HandleSyntheticFingerprint(s.ctx, &event)

			if c.ReturnsError {
				assert.ErrorContains(t, err, c.ExpectedResponse)
			} else {
				require.NoError(t, err)
				s.Require().NoError(c.VehicleNFT.Reload(s.ctx, s.pdb.DBS().Reader))
				s.Require().True(c.VehicleNFT.ClaimID.Valid)

				vc, err := models.FindVerifiableCredential(s.ctx, s.pdb.DBS().Reader.DB, c.VehicleNFT.ClaimID.String)
				s.Require().NoError(err)
				s.Require().Equal(c.ExpiresAt, vc.ExpirationDate)

			}
		})
	}

}

func (s *ConsumerTestSuite) TestSignatureValidation() {

	cases := []struct {
		Data string
	}{
		{
			Data: `{
				"data": {"rpiUptimeSecs":39,"batteryVoltage":13.49,"timestamp":1688136702634,"vin":"W1N2539531F907299","protocol":"7"},
				"id": "2RvhwjUbtoePjmXN7q9qfjLQgwP",
				"signature": "7c31e54ddcffc2a548ccaf10ed64b7e4bdd239bbaa3e5f6dba41d3e4051d930b7fbdf184724c2fb8d3b2ac8ac82662d2ed74e881dd01c09c4b2a9b4e62ede5db1b",
				"source": "aftermarket/device/fingerprint",
				"specversion": "1.0",
				"subject": "0x448cF8Fd88AD914e3585401241BC434FbEA94bbb",
				"type": "zone.dimo.aftermarket.device.fingerprint"
			}`,
		},
		{
			Data: `{
				"data": {"rpiUptimeSecs":36,"batteryVoltage":13.73,"timestamp":1688760445189,"vin":"LRBFXCSA5KD124854","protocol":"6"},
				"id": "2SG6Cu2NWOcu7LvhadPtmDGb65S",
				"signature": "5fb985f758c6224ab45630d055c7aca163329b88accfb8fd76a0dbb13b2ebcfe3c5bd8b801851f683f7a288c174a11ed8fc2631d95929c3b3cc85c75fb10ea001c",
				"source": "aftermarket/device/fingerprint",
				"specversion": "1.0",
				"subject": "0x06fF8E7A4A159EA388da7c133DC5F79727868d83",
				"type": "zone.dimo.aftermarket.device.fingerprint"
			}`,
		},
	}

	for _, c := range cases {
		var event Event
		err := json.Unmarshal([]byte(c.Data), &event)
		require.NoError(s.T(), err)
		data, err := json.Marshal(event.Data)
		require.NoError(s.T(), err)

		signature := common.FromHex(event.Signature)
		addr := common.HexToAddress(event.Subject)
		hash := crypto.Keccak256Hash(data)
		recAddr, err := helpers.Ecrecover(hash.Bytes(), signature)
		s.NoError(err)
		s.Equal(addr, recAddr)
	}
}

func (s *ConsumerTestSuite) TestInvalidSignature() {
	msg := `{
		"data": {"rpiUptimeSecs":36,"batteryVoltage":13.73,"timestamp":1688760445189,"vin":"LRBFXCSA5KD124854","protocol":"6"},
		"id": "2SG6Cu2NWOcu7LvhadPtmDGb65S",
		"signature": "5fb985f758c6224ab45630d055c7aca163329b88accfb8fd76a0dbb13b2ebcfe3c5bd8b801851f683f7a288c174a11ed8fc2631d95929c3b3cc85c75fb10ea001a",
		"source": "aftermarket/device/fingerprint",
		"specversion": "1.0",
		"subject": "0x06fF8E7A4A159EA388da7c133DC5F79727868d83",
		"type": "zone.dimo.aftermarket.device.fingerprint"
	}`

	var event Event
	err := json.Unmarshal([]byte(msg), &event)
	require.NoError(s.T(), err)
	data, err := json.Marshal(event.Data)
	require.NoError(s.T(), err)

	signature := common.FromHex(event.Signature)
	addr := common.HexToAddress(event.Subject)
	hash := crypto.Keccak256Hash(data)
	recAddr, err := helpers.Ecrecover(hash.Bytes(), signature)
	s.Error(err)
	s.Equal(err.Error(), "invalid signature recovery id")
	s.NotEqual(recAddr, addr)
}
