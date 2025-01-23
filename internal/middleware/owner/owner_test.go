package owner

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
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

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, "../../../migrations")
	logger := test.Logger()

	usersClient := &test.UsersClient{}
	middleware := UserDevice(pdb, usersClient, logger)

	usersClient.Store = map[string]*pb.User{}
	usersClient.Store[userID] = &pb.User{
		Id:              userID,
		EthereumAddress: &userAddr,
	}
	usersClient.Store[otherUserID] = &pb.User{
		Id:              otherUserID,
		EthereumAddress: nil,
	}

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
			app.Get("/:userDeviceID", test.AuthInjectorTestHandler(c.UserID, nil), middleware, func(c *fiber.Ctx) error {
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

	usersClient := &test.UsersClient{}
	middleware := AftermarketDevice(pdb, usersClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/:serial", test.AuthInjectorTestHandler(userID, nil), middleware, func(c *fiber.Ctx) error {
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

			usersClient.Store = map[string]*pb.User{}
			u := &pb.User{Id: userID}
			u.EthereumAddress = c.UserEthAddr
			usersClient.Store[userID] = u

			t.Log(c.Name)
			res, err := app.Test(request)
			require.Nil(t, err)
			assert.Equal(t, c.ExpectedCode, res.StatusCode)
		})
	}

	require.NoError(t, container.Terminate(ctx))
}

func TestVehicleTokenOwnerMiddleware(t *testing.T) {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, "../../../migrations")
	logger := test.Logger()

	usersClient := &test.UsersClient{}
	middleware := VehicleToken(pdb, usersClient, logger)
	app := test.SetupAppFiber(*logger)

	userID := ksuid.New().String()
	app.Get("/user/vehicle/:tokenID/commands/burn", test.AuthInjectorTestHandler(userID, nil), middleware, func(c *fiber.Ctx) error {
		logger := c.Locals("logger").(*zerolog.Logger)
		logger.Info().Msg("Omega croggers.")
		return nil
	})

	cases := []struct {
		Name         string
		UserID       string
		OwnerAddress common.Address
		TokenID      *big.Int
		ExpectedCode int
	}{
		{
			Name:         "valid-user-id/valid-addr",
			UserID:       userID,
			OwnerAddress: common.HexToAddress("0x1ABC7154748d1ce5144478cdeB574ae244b939B5"),
			TokenID:      big.NewInt(5),
			ExpectedCode: 200,
		},
		{
			Name:         "no-eth-addr",
			UserID:       userID,
			ExpectedCode: errNotFound.Code,
			OwnerAddress: common.HexToAddress(""),
			TokenID:      big.NewInt(5),
		},
		{
			Name:         "user-not-found",
			UserID:       ksuid.New().String(),
			ExpectedCode: errNotFound.Code,
			OwnerAddress: common.HexToAddress("0x1ABC7154748d1ce5144478cdeB574ae244b939B5"),
			TokenID:      big.NewInt(6),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := models.UserDevices().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)
			_, err = models.UserDevices().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)

			ud := test.SetupCreateUserDevice(t, c.UserID, "ddID", nil, "vin", pdb)

			usersClient.Store = map[string]*pb.User{}
			u := &pb.User{Id: userID}

			if c.OwnerAddress != common.HexToAddress("") {
				_ = test.SetupCreateVehicleNFT(t, ud, big.NewInt(5), null.BytesFrom(c.OwnerAddress.Bytes()), pdb)

				addr := c.OwnerAddress.Hex()
				u.EthereumAddress = &addr
			}

			usersClient.Store[userID] = u
			request := test.BuildRequest("GET", fmt.Sprintf("/user/vehicle/%s/commands/burn", c.TokenID.String()), "")
			res, err := app.Test(request)
			require.Nil(t, err)
			assert.Equal(t, c.ExpectedCode, res.StatusCode)
		})
	}

	require.NoError(t, container.Terminate(ctx))
}
