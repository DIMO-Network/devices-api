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

func (s *ConsumerTestSuite) TestVinCredentialerHandler() {
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
		EthereumAddress: ownerAddress,
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

	tm, _ := time.Parse(time.RFC1123, "2023-06-30T14:51:42.63507585Z")
	msg := DeviceFingerprintCloudEvent{
		CloudEventHeaders: CloudEventHeaders{
			ID:          "2RvhwjUbtoePjmXN7q9qfjLQgwP",
			Subject:     "0x448cF8Fd88AD914e3585401241BC434FbEA94bbb",
			Source:      "aftermarket/device/fingerprint",
			SpecVersion: "1.0",
			Time:        tm,
			Type:        "zone.dimo.aftermarket.device.fingerprint",
			Signature:   "7c31e54ddcffc2a548ccaf10ed64b7e4bdd239bbaa3e5f6dba41d3e4051d930b7fbdf184724c2fb8d3b2ac8ac82662d2ed74e881dd01c09c4b2a9b4e62ede5db1b",
		},
		Data: Data{
			CommonData: CommonData{
				BatteryVoltage: 13.49,
				Timestamp:      1688136702634,
				RpiUptimeSecs:  39,
			},
			Vin:      "W1N2539531F907299",
			Protocol: "7",
		},
	}

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
				EthereumAddress: ownerAddress,
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

			err = s.cons.Handle(s.ctx, &msg)

			if c.ReturnsError {
				assert.ErrorContains(t, err, c.ExpectedResponse)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func (s *ConsumerTestSuite) TestSignatureValidation() {

	cases := []struct {
		Data      DeviceFingerprintCloudEvent
		Signature string
		Time      string
	}{
		{
			Data: DeviceFingerprintCloudEvent{
				CloudEventHeaders: CloudEventHeaders{
					ID:          "2RvhwjUbtoePjmXN7q9qfjLQgwP",
					Subject:     "0x448cF8Fd88AD914e3585401241BC434FbEA94bbb",
					Source:      "aftermarket/device/fingerprint",
					SpecVersion: "1.0",
					Type:        "zone.dimo.aftermarket.device.fingerprint",
				},
				Data: Data{
					CommonData: CommonData{
						BatteryVoltage: 13.49,
						Timestamp:      1688136702634,
						RpiUptimeSecs:  39,
					},
					Vin:      "W1N2539531F907299",
					Protocol: "7",
				},
			},
			Signature: "7c31e54ddcffc2a548ccaf10ed64b7e4bdd239bbaa3e5f6dba41d3e4051d930b7fbdf184724c2fb8d3b2ac8ac82662d2ed74e881dd01c09c4b2a9b4e62ede5db1b",
			Time:      "2023-06-30T14:51:42.63507585Z",
		},
		{
			Data: DeviceFingerprintCloudEvent{
				CloudEventHeaders: CloudEventHeaders{
					ID:          "2S9zmhIVydI2mjPjqxOFiV8WYwN",
					Subject:     "0xa3CF0f4670F557D1F2f5d38F35645363c5c06f8d",
					Source:      "aftermarket/device/fingerprint",
					SpecVersion: "1.0",
					Type:        "zone.dimo.aftermarket.device.fingerprint",
				},
				Data: Data{
					CommonData: CommonData{
						BatteryVoltage: 13.43,
						Timestamp:      1688573744829,
						RpiUptimeSecs:  37,
					},
					Vin:      "4T1BF1FK5EU372595",
					Protocol: "6",
				},
			},
			Signature: "e6d4fb3b81c2c533d9aaa9eb6131f2e0492e7e83fae3462fe64604d45a6eafa178193b47a6621e1ce7c29a5bb95adf294bcb5bae41e1244a9021c7f6c4a071621b",
			Time:      "2023-07-05T16:15:44.829649462Z",
		},
	}

	for _, c := range cases {

		tm, _ := time.Parse(time.RFC1123, c.Time)
		c.Data.CloudEventHeaders.Time = tm
		data, err := json.Marshal(c.Data.Data)
		require.NoError(s.T(), err)

		signature := common.FromHex(c.Signature)
		addr := common.HexToAddress(c.Data.CloudEventHeaders.Subject)
		hash := crypto.Keccak256Hash(data)
		recAddr, err := helpers.Ecrecover(hash, signature)
		s.NoError(err)
		s.Equal(addr, recAddr)
	}
}

func (s *ConsumerTestSuite) TestInvalidSignature() {

	tm, _ := time.Parse(time.RFC1123, "2023-06-30T14:51:42.63507585Z")

	msg := DeviceFingerprintCloudEvent{
		CloudEventHeaders: CloudEventHeaders{
			ID:          "2RvhwjUbtoePjmXN7q9qfjLQgwP",
			Subject:     "0x448cF8Fd88AD914e3585401241BC434FbEA94bbb",
			Source:      "aftermarket/device/fingerprint",
			SpecVersion: "1.0",
			Time:        tm,
			Type:        "zone.dimo.aftermarket.device.fingerprint",
		},
		Data: Data{
			CommonData: CommonData{
				BatteryVoltage: 13.49,
				Timestamp:      1688136702634,
				RpiUptimeSecs:  39,
			},
			Vin:      "W1N2539531F907299",
			Protocol: "7",
		},
	}

	sig := "7c31e54ddcffc2a548ccaf10ed64b7e4bdd239bbaa3e5f6dba41d3e4051d930b7fbdf184724c2fb8d3b2ac8ac82662d2ed74e881dd01c09c4b2a9b4e62ede5db1a"
	data, err := json.Marshal(msg.Data)
	require.NoError(s.T(), err)

	signature := common.FromHex(sig)
	addr := common.HexToAddress(msg.CloudEventHeaders.Subject)
	hash := crypto.Keccak256Hash(data)
	recAddr, err := helpers.Ecrecover(hash, signature)
	s.Error(err)
	s.Equal(err.Error(), "invalid signature recovery id")
	s.NotEqual(recAddr, addr)
}
