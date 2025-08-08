package rpc

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"

	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
)

const migrationsDirRelPath = "../../migrations"
const autoPiIntegrationID = "autopi123aaaaaaaaaaaaaaaaaa"

func populateDB(ctx context.Context, pdb db.Store) (string, error) {
	integration := test.BuildIntegrationGRPC(autoPiIntegrationID, constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	vin := "W1N2539531F907299"
	deviceStyleID := "24GE7Mlc4c9o4j5P4mcD1Fzinx1"
	userID := ksuid.New().String()
	ownerAddress := null.BytesFrom(common.Hex2Bytes("448cF8Fd88AD914e3585401241BC434FbEA94bbb"))
	_, childWallet, _ := test.GenerateWallet()

	ud := models.UserDevice{
		ID:            ksuid.New().String(),
		UserID:        userID,
		DefinitionID:  dd[0].Id,
		VinIdentifier: null.StringFrom(vin),
		CountryCode:   null.StringFrom("USA"),
		VinConfirmed:  true,
		Metadata:      null.JSONFrom([]byte(`{ "powertrainType": "ICE", "canProtocol": "6" }`)),
		DeviceStyleID: null.StringFrom(deviceStyleID),
		TokenID:       types.NewNullDecimal(decimal.New(4, 0)),
		OwnerAddress:  null.BytesFrom(common.BigToAddress(big.NewInt(7)).Bytes()),
		MintRequestID: null.StringFrom(ksuid.New().String()),
	}

	ad := models.AftermarketDevice{
		UserID:                    null.StringFrom(ud.ID),
		OwnerAddress:              ownerAddress,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		TokenID:                   types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(13), 0)),
		VehicleTokenID:            ud.TokenID,
		Beneficiary:               null.BytesFrom(common.BytesToAddress([]byte{uint8(1)}).Bytes()),
		EthereumAddress:           ownerAddress.Bytes,
		DeviceManufacturerTokenID: types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(42), 0)),
	}

	sd := models.SyntheticDevice{
		VehicleTokenID:     ud.TokenID,
		IntegrationTokenID: types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(19), 0)),
		MintRequestID:      ksuid.New().String(),
		WalletChildNumber:  100,
		TokenID:            types.NewNullDecimal(decimal.New(6, 0)),
		WalletAddress:      childWallet.Bytes(),
	}

	metaTxUd := models.MetaTransactionRequest{
		ID:     ud.MintRequestID.String,
		Status: models.MetaTransactionRequestStatusConfirmed,
	}

	metaTxSd := models.MetaTransactionRequest{
		ID:     sd.MintRequestID,
		Status: models.MetaTransactionRequestStatusConfirmed,
	}

	if err := metaTxUd.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	if err := metaTxSd.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	if err := ud.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	if err := ad.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	if err := sd.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	return ud.ID, nil
}

func TestGetUserDevice_AftermarketDeviceObj_NotNil(t *testing.T) {
	// AftermarketDevice obj is not nil when valid data is present
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			assert.NoError(err)
		}
	}()

	userDeviceID, err := populateDB(ctx, pdb)
	assert.NoError(err)

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, userDeviceSvc, nil)

	udResult, err := udService.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{Id: userDeviceID})
	assert.NoError(err)

	// AftermarketDevice obj set correctly (not nil)
	assert.NotNil(udResult.AftermarketDevice)
	assert.Equal(*udResult.AftermarketDevice.UserId, userDeviceID)
}

func TestGetUserDevice_AftermarketDeviceObj_Nil(t *testing.T) {
	// AftermarketDevice obj is nil when no associated AD is set
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			assert.NoError(err)
		}
	}()

	userDeviceID, err := populateDB(ctx, pdb)
	assert.NoError(err)

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, userDeviceSvc, nil)

	_, err = models.AftermarketDevices(
		models.AftermarketDeviceWhere.UserID.EQ(null.StringFrom(userDeviceID)),
	).DeleteAll(ctx, pdb.DBS().Writer)
	assert.NoError(err)

	udResult, err := udService.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{Id: userDeviceID})
	assert.NoError(err)
	assert.Nil(udResult.AftermarketDevice)
}

func TestGetUserDevice_PopulateDeprecatedFields(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	userDeviceID, err := populateDB(ctx, pdb)
	assert.NoError(err)

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, userDeviceSvc, nil)

	udResult, err := udService.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{Id: userDeviceID})
	assert.NoError(err)

	// Deprecated fields still populated
	assert.Equal(udResult.AftermarketDevice.Beneficiary, udResult.AftermarketDeviceBeneficiaryAddress) //nolint:staticcheck
	assert.Equal(udResult.AftermarketDevice.TokenId, *udResult.AftermarketDeviceTokenId)               //nolint:staticcheck
	assert.NotEmpty(udResult.AftermarketDeviceBeneficiaryAddress)                                      //nolint:staticcheck
}

func TestGetUserDevice_PopulateSyntheticDeviceFields(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	userDeviceID, err := populateDB(ctx, pdb)
	assert.NoError(err)

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, userDeviceSvc, nil)

	udResult, err := udService.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{Id: userDeviceID})
	assert.NoError(err)

	assert.Equal(udResult.SyntheticDevice.TokenId, uint64(6))
	assert.Equal(udResult.SyntheticDevice.IntegrationTokenId, uint64(19))
}

func TestGetUserDevice_NoSyntheticDeviceFields_WhenNoTokenID(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	userDeviceID, err := populateDB(ctx, pdb)
	assert.NoError(err)

	sd, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(6, 0))),
		models.SyntheticDeviceWhere.IntegrationTokenID.EQ(types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(19), 0))),
	).One(ctx, pdb.DBS().Reader)
	assert.NoError(err)

	sd.TokenID = types.NullDecimal{}

	_, err = sd.Update(ctx, pdb.DBS().Writer, boil.Whitelist(models.SyntheticDeviceColumns.TokenID))
	assert.NoError(err)

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, userDeviceSvc, nil)

	udResult, err := udService.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{Id: userDeviceID})
	assert.NoError(err)

	assert.Nil(udResult.SyntheticDevice)
}
