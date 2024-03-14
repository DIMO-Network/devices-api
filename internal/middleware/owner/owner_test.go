package owner

import (
	"context"
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
	otherAddr := "0x3AC4f4Ae05b75b97bfC71Ea518913007FdCaab70"
	userDeviceID := "2OeRoU9VmbFVpgpPy3BjY2WsMMm"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, "../../../migrations")
	logger := test.Logger()

	usersClient := &test.UsersClient{}
	middleware := UserDevice(pdb, usersClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/:userDeviceID", test.AuthInjectorTestHandler(userID), middleware, func(c *fiber.Ctx) error {
		logger := c.Locals("logger").(*zerolog.Logger)
		logger.Info().Msg("Omega croggers.")
		return nil
	})

	request := test.BuildRequest("GET", "/"+userDeviceID, "")

	cases := []struct {
		Name                string
		UserDeviceUserID    string
		NFTOwnerAddr        string
		UserExists          bool
		UserEthereumAddress string
		ExpectedCode        int
	}{
		{
			Name:         "NoDevice",
			ExpectedCode: 404,
		},
		{
			Name:             "UserIDMatch",
			UserDeviceUserID: userID,
			ExpectedCode:     200,
		},
		{
			Name:             "UserIDMismatchNoAccount",
			UserDeviceUserID: otherUserID,
			ExpectedCode:     404,
		},
		{
			Name:             "UserIDMismatchNoEthereumAddress",
			UserDeviceUserID: otherUserID,
			UserExists:       true,
			ExpectedCode:     404,
		},
		{
			Name:                "UserIDMismatchNotMinted",
			UserDeviceUserID:    otherUserID,
			UserExists:          true,
			UserEthereumAddress: userAddr,
			ExpectedCode:        404,
		},
		{
			Name:                "UserIDMismatchEthereumAddressMatch",
			UserDeviceUserID:    otherUserID,
			NFTOwnerAddr:        userAddr,
			UserExists:          true,
			UserEthereumAddress: userAddr,
			ExpectedCode:        200,
		},
		{
			Name:                "UserIDMismatchEthereumAddressMismatch",
			UserDeviceUserID:    otherUserID,
			NFTOwnerAddr:        otherAddr,
			UserExists:          true,
			UserEthereumAddress: userAddr,
			ExpectedCode:        404,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := models.UserDevices().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)

			if c.UserDeviceUserID != "" {
				ud := models.UserDevice{ID: userDeviceID, UserID: c.UserDeviceUserID}
				require.NoError(t, ud.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

				if c.NFTOwnerAddr != "" {
					mtr := models.MetaTransactionRequest{
						ID: ksuid.New().String(),
					}

					require.NoError(t, mtr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

					ud.MintRequestID = null.StringFrom(mtr.ID)
					ud.OwnerAddress = null.BytesFrom(common.FromHex(c.NFTOwnerAddr))

					_, err = ud.Update(ctx, pdb.DBS().Writer, boil.Whitelist(models.UserDeviceColumns.MintRequestID, models.UserDeviceColumns.OwnerAddress))
				}
			}

			usersClient.Store = map[string]*pb.User{}

			if c.UserExists {
				u := &pb.User{Id: userID}
				if c.UserEthereumAddress != "" {
					u.EthereumAddress = &c.UserEthereumAddress
				}
				usersClient.Store[userID] = u
			}

			res, err := app.Test(request)
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
	app.Get("/:serial", test.AuthInjectorTestHandler(userID), middleware, func(c *fiber.Ctx) error {
		logger := c.Locals("logger").(*zerolog.Logger)
		logger.Info().Msg("Omega croggers.")
		return nil
	})

	request := test.BuildRequest("GET", "/"+unitID, "")

	cases := []struct {
		Name               string
		AutoPiVehicleToken types.NullDecimal
		AutoPiUserID       null.String
		AutoPiUnitID       string
		UserEthAddr        *string
		vNFTAddr           string
		ExpectedCode       int
	}{
		{
			Name:         "AftermarketDevice not minted, or unit ID invalid.",
			ExpectedCode: 404,
		},
		{
			Name:         "Token ID is null, device is not paired",
			ExpectedCode: 200,
			AutoPiUnitID: unitID,
		},
		{
			Name:         "Check if user is web2 owner",
			ExpectedCode: 200,
			AutoPiUnitID: unitID,
			AutoPiUserID: null.StringFrom(userID),
		},
		{
			Name:               "user does not have a valid ethereum address",
			ExpectedCode:       403,
			AutoPiVehicleToken: types.NewNullDecimal(decimal.New(int64(1), 0)),
			AutoPiUnitID:       unitID,
		},
		{
			Name:               "user is not owner of paired vehicle or AftermarketDevice",
			ExpectedCode:       403,
			AutoPiVehicleToken: types.NewNullDecimal(decimal.New(int64(1), 0)),
			AutoPiUnitID:       unitID,
			UserEthAddr:        &userAddr,
		},
		{
			Name:               "user is owner",
			ExpectedCode:       200,
			AutoPiVehicleToken: types.NewNullDecimal(decimal.New(int64(1), 0)),
			AutoPiUnitID:       unitID,
			UserEthAddr:        &userAddr,
			vNFTAddr:           userAddr,
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

			mtx := models.MetaTransactionRequest{
				ID:     ksuid.New().String(),
				Status: models.MetaTransactionRequestStatusConfirmed,
			}
			err = mtx.Insert(ctx, pdb.DBS().Writer, boil.Infer())
			require.NoError(t, err)

			ud := models.UserDevice{
				ID:                 ksuid.New().String(),
				MintRequestID:      null.StringFrom(mtx.ID),
				TokenID:            c.AutoPiVehicleToken,
				UserID:             userID,
				DeviceDefinitionID: ksuid.New().String(),
			}

			if c.vNFTAddr != "" {
				ud.OwnerAddress = null.BytesFrom(common.FromHex(c.vNFTAddr))
			}

			err = ud.Insert(ctx,
				pdb.DBS().Writer,
				boil.Whitelist(
					models.UserDeviceColumns.ID,
					models.UserDeviceColumns.UserID,
					models.UserDeviceColumns.DeviceDefinitionID,
					models.UserDeviceColumns.MintRequestID,
					models.UserDeviceColumns.TokenID,
					models.UserDeviceColumns.OwnerAddress))
			require.NoError(t, err)

			ap := models.AftermarketDevice{
				Serial:                        c.AutoPiUnitID,
				VehicleTokenID:                c.AutoPiVehicleToken,
				UserID:                        c.AutoPiUserID,
				ClaimMetaTransactionRequestID: null.StringFrom(mtx.ID),
			}
			err = ap.Insert(ctx, pdb.DBS().Writer, boil.Infer())
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
