package owner

import (
	"context"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

func TestUserDeviceOwnerMiddleware(t *testing.T) {
	userID := "louxUser"
	userAddr := "0x1ABC7154748d1ce5144478cdeB574ae244b939B5"
	otherUserID := "stanleyUser"
	userDeviceID1 := ksuid.New().String()
	userDeviceID2 := ksuid.New().String()

	addr := common.HexToAddress(userAddr)

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, "../../../migrations")
	logger := test.Logger()

	middleware := UserDevice(pdb, logger)

	ud := []models.UserDevice{
		{
			ID:                 userDeviceID1,
			UserID:             userID,
			DeviceDefinitionID: ksuid.New().String(),
			DefinitionID:       "ford_escape_2020",
			OwnerAddress:       null.BytesFrom(common.HexToAddress(userAddr).Bytes()),
		},
		{
			ID:                 userDeviceID2,
			UserID:             otherUserID,
			DeviceDefinitionID: ksuid.New().String(),
			DefinitionID:       "ford_escape_2020",
			OwnerAddress:       null.BytesFrom(common.HexToAddress(userAddr).Bytes()),
		},
	}

	for _, u := range ud {
		err := u.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		require.NoError(t, err)
	}

	cases := []struct {
		Name             string
		UserDeviceUserID string
		UserID           string
		OwnerAddress     string
		ExpectedCode     int
		Error            error
	}{
		{
			Name:             "user-id-udid-match",
			UserDeviceUserID: userDeviceID1,
			UserID:           userID,
			ExpectedCode:     200,
		},
		{
			Name:             "user-owners-ud",
			UserDeviceUserID: userDeviceID2,
			UserID:           userID,
			ExpectedCode:     200,
		},
		{
			Name:             "device-does-not-exist",
			UserDeviceUserID: ksuid.New().String(),
			UserID:           userID,
			ExpectedCode:     404,
			Error:            errNotFound,
		},
		{
			Name:             "invalid-eth-addr",
			UserDeviceUserID: ksuid.New().String(),
			UserID:           otherUserID,
			ExpectedCode:     404,
			Error:            errNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			app := test.SetupAppFiber(*logger)
			app.Get("/:userDeviceID", test.AuthInjectorTestHandler(c.UserID, &addr), middleware, func(c *fiber.Ctx) error {
				logger := c.Locals("logger").(*zerolog.Logger)
				logger.Info().Msg("Omega croggers.")
				return nil
			})

			res, err := app.Test(test.BuildRequest("GET", "/"+c.UserDeviceUserID, ""))
			require.Nil(t, err)
			assert.Equal(t, c.ExpectedCode, res.StatusCode)
		})
	}

	require.NoError(t, container.Terminate(ctx))
}

func TestAutoPiOwnerMiddleware(t *testing.T) {
	userID := "louxUser"
	userAddr := "0x9eaD03F7136Fc6b4bDb0780B00a1c14aE5A8B6d0"
	unitID := "4a12c37b-b662-4fad-68e6-7e74f9ce658c"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, "../../../migrations")
	logger := test.Logger()

	middleware := AftermarketDevice(pdb, logger)

	addr := common.HexToAddress(userAddr)

	app := test.SetupAppFiber(*logger)
	app.Get("/:serial", test.AuthInjectorTestHandler(userID, &addr), middleware, func(c *fiber.Ctx) error {
		logger := c.Locals("logger").(*zerolog.Logger)
		logger.Info().Msg("Omega croggers.")
		return nil
	})

	request := test.BuildRequest("GET", "/"+unitID, "")

	cases := []struct {
		Name              string
		UserEthAddr       *string
		AftermarketDevice models.AftermarketDevice
		UserDevice        models.UserDevice
		ExpectedCode      int
	}{
		{
			Name:         "AftermarketDevice not minted, or unit ID invalid.",
			ExpectedCode: 404,
			UserDevice: models.UserDevice{
				ID:     ksuid.New().String(),
				UserID: userID,
			},
		},
		{
			Name:         "Token ID is null, device is not paired",
			ExpectedCode: 200,
			UserDevice: models.UserDevice{
				ID:     ksuid.New().String(),
				UserID: userID,
			},
			AftermarketDevice: models.AftermarketDevice{
				Serial: unitID,
			},
		},
		{
			Name:         "Check if user is web2 owner",
			ExpectedCode: 200,
			UserDevice: models.UserDevice{
				ID:     ksuid.New().String(),
				UserID: userID,
			},
			AftermarketDevice: models.AftermarketDevice{
				UserID: null.StringFrom(userID),
				Serial: unitID,
			},
		},
		{
			Name:         "user does not have a valid ethereum address",
			ExpectedCode: 403,
			UserDevice: models.UserDevice{
				ID:      ksuid.New().String(),
				UserID:  userID,
				TokenID: types.NewNullDecimal(decimal.New(int64(1), 0)),
			},
			AftermarketDevice: models.AftermarketDevice{
				Serial:         unitID,
				VehicleTokenID: types.NewNullDecimal(decimal.New(int64(1), 0)),
			},
		},
		{
			Name:         "user is not owner of paired vehicle or AftermarketDevice",
			ExpectedCode: 403,
			UserEthAddr:  &userAddr,
			AftermarketDevice: models.AftermarketDevice{
				Serial:         unitID,
				VehicleTokenID: types.NewNullDecimal(decimal.New(int64(1), 0)),
			},
			UserDevice: models.UserDevice{
				ID:           ksuid.New().String(),
				UserID:       userID,
				TokenID:      types.NewNullDecimal(decimal.New(int64(1), 0)),
				OwnerAddress: null.BytesFrom(common.Hex2Bytes("1ABC7154748d1ce5144478cdeB574ae244b939B5")),
			},
		},
		{
			Name:         "user is owner",
			ExpectedCode: 200,
			UserEthAddr:  &userAddr,
			AftermarketDevice: models.AftermarketDevice{
				Serial:         unitID,
				VehicleTokenID: types.NewNullDecimal(decimal.New(int64(1), 0)),
			},
			UserDevice: models.UserDevice{
				ID:           ksuid.New().String(),
				UserID:       userID,
				TokenID:      types.NewNullDecimal(decimal.New(int64(1), 0)),
				OwnerAddress: null.BytesFrom(common.HexToAddress(userAddr).Bytes()),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := models.AftermarketDevices().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)
			_, err = models.UserDevices().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)
			_, err = models.MetaTransactionRequests().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)

			err = c.UserDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
			require.NoError(t, err)

			err = c.AftermarketDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())
			require.NoError(t, err)

			t.Log(c.Name)
			res, err := app.Test(request)
			require.Nil(t, err)
			assert.Equal(t, c.ExpectedCode, res.StatusCode)
		})
	}

	require.NoError(t, container.Terminate(ctx))
}
