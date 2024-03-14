package rpc

import (
	"context"
	"database/sql"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/shared/db"
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

func populateDB(ctx context.Context, pdb db.Store) (string, error) {
	integration := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "F150", 2020, integration)
	vin := "W1N2539531F907299"
	deviceStyleID := "24GE7Mlc4c9o4j5P4mcD1Fzinx1"
	userID := ksuid.New().String()
	ownerAddress := null.BytesFrom(common.Hex2Bytes("448cF8Fd88AD914e3585401241BC434FbEA94bbb"))
	claimID := ksuid.New().String()
	_, childWallet, _ := test.GenerateWallet()

	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             userID,
		DeviceDefinitionID: dd[0].DeviceDefinitionId,
		VinIdentifier:      null.StringFrom(vin),
		CountryCode:        null.StringFrom("USA"),
		VinConfirmed:       true,
		Metadata:           null.JSONFrom([]byte(`{ "powertrainType": "ICE", "canProtocol": "6" }`)),
		DeviceStyleID:      null.StringFrom(deviceStyleID),
	}

	vnft := models.VehicleNFT{
		UserDeviceID:  null.StringFrom(ud.ID),
		Vin:           ud.VinIdentifier.String,
		TokenID:       types.NewNullDecimal(decimal.New(4, 0)),
		OwnerAddress:  null.BytesFrom(common.BigToAddress(big.NewInt(7)).Bytes()),
		MintRequestID: ksuid.New().String(),
		ClaimID:       null.StringFrom(claimID),
	}

	ad := models.AftermarketDevice{
		UserID:                    null.StringFrom(ud.ID),
		OwnerAddress:              ownerAddress,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		TokenID:                   types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(13), 0)),
		VehicleTokenID:            vnft.TokenID,
		Beneficiary:               null.BytesFrom(common.BytesToAddress([]byte{uint8(1)}).Bytes()),
		EthereumAddress:           ownerAddress.Bytes,
		DeviceManufacturerTokenID: types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(42), 0)),
	}

	sd := models.SyntheticDevice{
		VehicleTokenID:     vnft.TokenID,
		IntegrationTokenID: types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(19), 0)),
		MintRequestID:      vnft.MintRequestID,
		WalletChildNumber:  100,
		TokenID:            types.NewNullDecimal(decimal.New(6, 0)),
		WalletAddress:      childWallet.Bytes(),
	}

	credential := models.VerifiableCredential{
		ClaimID:        claimID,
		ExpirationDate: time.Now().AddDate(0, 0, 7),
	}

	metaTx := models.MetaTransactionRequest{
		ID:     vnft.MintRequestID,
		Status: models.MetaTransactionRequestStatusConfirmed,
	}

	if err := ud.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	if err := metaTx.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	if err := credential.Insert(ctx, pdb.DBS().Reader, boil.Infer()); err != nil {
		return "", err
	}

	if err := vnft.Insert(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
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
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

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
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

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
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

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
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

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
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

	udResult, err := udService.GetUserDevice(ctx, &pb_devices.GetUserDeviceRequest{Id: userDeviceID})
	assert.NoError(err)

	assert.Nil(udResult.SyntheticDevice)
}

func TestClearMetaTransactionRequests(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	mtID := ksuid.New().String()

	currTime := time.Now()
	fifteenminsAgo := currTime.Add(-time.Minute * 15)
	metaTx := models.MetaTransactionRequest{
		ID:        mtID,
		Status:    models.MetaTransactionRequestStatusConfirmed,
		CreatedAt: fifteenminsAgo,
	}

	err := metaTx.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

	resp, err := udService.ClearMetaTransactionRequests(ctx, nil)
	assert.NoError(err)

	assert.Equal(mtID, resp.Id)

	_, err = models.MetaTransactionRequests(models.MetaTransactionRequestWhere.ID.EQ(mtID)).One(ctx, pdb.DBS().Reader)
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestClearMetaTransactionRequests_MultipleRecords(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	mtID := []string{ksuid.New().String(), ksuid.New().String()}

	currTime := time.Now()
	fifteenminsAgo := currTime.Add(-time.Minute * 15)
	sixteenMins := currTime.Add(-time.Minute * 16)
	metaTx := []models.MetaTransactionRequest{
		{
			ID:        mtID[0],
			Status:    models.MetaTransactionRequestStatusConfirmed,
			CreatedAt: fifteenminsAgo,
		},
		{
			ID:        mtID[1],
			Status:    models.MetaTransactionRequestStatusConfirmed,
			CreatedAt: sixteenMins,
		},
	}

	for _, m := range metaTx {
		err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(err)
	}

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

	resp, err := udService.ClearMetaTransactionRequests(ctx, nil)
	assert.NoError(err)
	assert.Equal(mtID[1], resp.Id)

	_, err = models.MetaTransactionRequests(models.MetaTransactionRequestWhere.ID.EQ(mtID[1])).One(ctx, pdb.DBS().Reader)
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestClearMetaTransactionRequests_MultipleRecords_Dates(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	mtID := []string{ksuid.New().String(), ksuid.New().String()}

	currTime := time.Now()
	fifteenminsAgo := currTime.Add(-time.Minute * 15)
	metaTx := []models.MetaTransactionRequest{
		{
			ID:        mtID[0],
			Status:    models.MetaTransactionRequestStatusConfirmed,
			CreatedAt: fifteenminsAgo,
		},
		{
			ID:     mtID[1],
			Status: models.MetaTransactionRequestStatusConfirmed,
		},
	}

	for _, m := range metaTx {
		err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(err)
	}

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

	resp, err := udService.ClearMetaTransactionRequests(ctx, nil)
	assert.NoError(err)
	assert.Equal(mtID[0], resp.Id)

	_, err = models.MetaTransactionRequests(models.MetaTransactionRequestWhere.ID.EQ(resp.Id)).One(ctx, pdb.DBS().Reader)
	assert.ErrorIs(err, sql.ErrNoRows)
}

func TestClearMetaTransactionRequests_NotExpired(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	mtID := ksuid.New().String()
	currTime := time.Now()
	expiryTime := currTime.Add(-time.Minute * 14)
	metaTx := models.MetaTransactionRequest{
		ID:        mtID,
		Status:    models.MetaTransactionRequestStatusUnsubmitted,
		CreatedAt: expiryTime,
	}

	err := metaTx.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	logger := zerolog.Logger{}
	userDeviceSvc := services.NewUserDeviceService(nil, logger, pdb.DBS, nil, nil)
	udService := NewUserDeviceRPCService(pdb.DBS, nil, nil, nil, nil, nil, nil, userDeviceSvc)

	resp, err := udService.ClearMetaTransactionRequests(ctx, nil)
	assert.Nil(resp)

	assert.Error(err)

	mt, err := models.MetaTransactionRequests(models.MetaTransactionRequestWhere.ID.EQ(mtID)).One(ctx, pdb.DBS().Reader)
	assert.NoError(err)

	assert.Equal(mt.ID, mtID)
}
