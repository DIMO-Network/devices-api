package attestor

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

const migrationsDirRelPath = "../../../migrations"

func TestCreateAttestation(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	logger := zerolog.Logger{}

	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			assert.NoError(err)
		}
	}()

	// create attestor
	witness := New(nil, pdb, &logger)
	if err := populateDB(ctx, pdb); err != nil {
		assert.NoError(err)
	}

	devices, err := models.UserDevices(
		qm.Load(
			qm.Rels(
				models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.Claim,
			),
		),
	).All(ctx, pdb.DBS().Reader)
	assert.NoError(err)

	attestations := [][]interface{}{}
	for _, device := range devices {
		if device.R.VehicleNFT == nil {
			logger.Warn().Str("userDevice", device.ID).Msg("vehicle not minted")
			continue
		}

		if device.R.VehicleNFT.R.Claim == nil {
			logger.Warn().Str("userDevice", device.ID).Str("vin", device.R.VehicleNFT.Vin).Msg("associated credential not found")
			continue
		}

		tokenID, ok := device.R.VehicleNFT.TokenID.Uint64()
		if !ok {
			logger.Warn().Str("userDevice", device.ID).Str("vin", device.R.VehicleNFT.Vin).Msg("invalid vehicle token id, this should never happen")
			continue
		}

		attestation := []interface{}{
			tokenID,
			device.DeviceDefinitionID,
			device.R.VehicleNFT.R.Claim.ClaimID,
			gjson.GetBytes(device.R.VehicleNFT.R.Claim.Credential.JSON, "proof.jws").String(),
		}
		attestations = append(attestations, attestation)
	}

	dataTypes := []string{"uint64", "string", "string", "string"}
	_, err = witness.GenerateMerkleTree(ctx, attestations, dataTypes)
	assert.NoError(err)

}

func populateDB(ctx context.Context, pdb db.Store) error {

	for i := 1; i < 10; i++ {
		ud := models.UserDevice{
			ID:                 ksuid.New().String(),
			UserID:             ksuid.New().String(),
			DeviceDefinitionID: ksuid.New().String(),
			VinIdentifier:      null.StringFrom(fmt.Sprintf("W1N2539531F90729%d", i)),
			CountryCode:        null.StringFrom("USA"),
			VinConfirmed:       true,
		}

		tokenID := big.NewInt(int64(i))
		vnft := models.VehicleNFT{
			UserDeviceID:  null.StringFrom(ud.ID),
			Vin:           ud.VinIdentifier.String,
			TokenID:       types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
			OwnerAddress:  null.BytesFrom(common.BigToAddress(big.NewInt(int64(i))).Bytes()),
			MintRequestID: ksuid.New().String(),
			ClaimID:       null.StringFrom(ksuid.New().String()),
		}

		ad := models.AftermarketDevice{
			Serial:                    ksuid.New().String(),
			UserID:                    null.StringFrom(ud.ID),
			OwnerAddress:              null.BytesFrom(common.BigToAddress(big.NewInt(int64(i))).Bytes()),
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
			TokenID:                   types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(int64(i)), 0)),
			VehicleTokenID:            vnft.TokenID,
			DeviceManufacturerTokenID: types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(int64(i)), 0)),
		}
		ad.EthereumAddress = ad.OwnerAddress.Bytes

		credential := models.VerifiableCredential{
			ClaimID:        vnft.ClaimID.String,
			ExpirationDate: time.Now().AddDate(0, 0, 7),
		}

		metaTx := models.MetaTransactionRequest{
			ID:     vnft.MintRequestID,
			Status: models.MetaTransactionRequestStatusConfirmed,
		}

		if err := ud.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}

		if err := metaTx.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}

		if err := credential.Insert(ctx, pdb.DBS().Reader, boil.Infer()); err != nil {
			return err
		}

		if err := vnft.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}

		if err := ad.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}

	}

	return nil
}
