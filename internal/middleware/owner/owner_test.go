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
			}

			if c.NFTOwnerAddr != "" {
				mtr := models.MetaTransactionRequest{
					ID: ksuid.New().String(),
				}

				require.NoError(t, mtr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

				vnft := models.VehicleNFT{
					MintRequestID: mtr.ID,
					UserDeviceID:  null.StringFrom(userDeviceID),
					OwnerAddress:  null.BytesFrom(common.FromHex(c.NFTOwnerAddr)),
				}

				require.NoError(t, vnft.Insert(ctx, pdb.DBS().Writer, boil.Infer()))
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
	userAddr := "1ABC7154748d1ce5144478cdeB574ae244b939B5"
	unitID := "4a12c37b-b662-4fad-68e6-7e74f9ce658c"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, "../../../migrations")
	logger := test.Logger()

	usersClient := &test.UsersClient{}
	middleware := AutoPi(pdb, usersClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/:unitID", test.AuthInjectorTestHandler(userID), middleware, func(c *fiber.Ctx) error {
		logger := c.Locals("logger").(*zerolog.Logger)
		logger.Info().Msg("Omega croggers.")
		return nil
	})

	request := test.BuildRequest("GET", "/"+unitID, "")

	cases := []struct {
		Name            string
		UserID          string
		AutoPiUnitID    string
		UserEthAddr     string
		AutoPiOwnerAddr string
		VehicleNFTAddr  string
		TokenID         int
		ExpectedCode    int
	}{
		{
			Name:         "NoDevice",
			ExpectedCode: 404,
		},
		{
			Name:         "AutoPiNotPaired",
			UserID:       userID,
			AutoPiUnitID: unitID,
			ExpectedCode: 200,
		},
		{
			Name:         "AutoPiPairedAddrsDontMatch",
			UserID:       userID,
			AutoPiUnitID: unitID,
			UserEthAddr:  userAddr,
			TokenID:      1,
			ExpectedCode: 403,
		},
		{
			Name:           "AutoPiVehicleNFTAddrMatches",
			UserID:         userID,
			AutoPiUnitID:   unitID,
			VehicleNFTAddr: userAddr,
			UserEthAddr:    userAddr,
			TokenID:        1,
			ExpectedCode:   200,
		},
		{
			Name:            "AutoPiPairedOwnerAddrMatches",
			UserID:          userID,
			AutoPiUnitID:    unitID,
			AutoPiOwnerAddr: userAddr,
			UserEthAddr:     userAddr,
			TokenID:         1,
			ExpectedCode:    200,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := models.AutopiUnits().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)
			_, err = models.VehicleNFTS().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)
			_, err = models.MetaTransactionRequests().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)

			if c.AutoPiUnitID != "" {
				var apEth null.Bytes
				var apOwnr null.Bytes
				var pairReq null.String
				var unpairReq null.String
				var tknID types.NullDecimal
				if c.VehicleNFTAddr != "" {
					mtxReq := "abcdefghijklmnopqrstuvwxyz1"
					mtr := models.MetaTransactionRequest{
						ID:     mtxReq,
						Status: models.MetaTransactionRequestStatusConfirmed,
					}
					err = mtr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
					require.NoError(t, err)

					vnft := models.VehicleNFT{
						MintRequestID: mtxReq,
						TokenID:       types.NewNullDecimal(decimal.New(int64(c.TokenID), 0)),
						OwnerAddress:  null.BytesFrom(common.Hex2Bytes(c.VehicleNFTAddr)),
					}
					err := vnft.Insert(ctx, pdb.DBS().Writer, boil.Infer())
					require.NoError(t, err)
				}

				if c.AutoPiOwnerAddr != "" {
					apOwnr = null.BytesFrom(common.FromHex(c.AutoPiOwnerAddr))
				}

				if c.TokenID > 0 {
					tknID = types.NewNullDecimal(decimal.New(int64(c.TokenID), 0))
					t.Log(tknID)
				}

				ap := models.AutopiUnit{
					AutopiUnitID:    c.AutoPiUnitID,
					EthereumAddress: apEth,
					OwnerAddress:    apOwnr,
					PairRequestID:   pairReq,
					UnpairRequestID: unpairReq,
					TokenID:         tknID,
				}
				require.NoError(t, ap.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

				usersClient.Store = map[string]*pb.User{}
				u := &pb.User{Id: userID}
				if c.UserEthAddr != "" {
					u.EthereumAddress = &c.UserEthAddr
				}
				usersClient.Store[userID] = u
			}

			t.Log(c.Name)
			res, err := app.Test(request)
			require.Nil(t, err)
			assert.Equal(t, c.ExpectedCode, res.StatusCode)
		})
	}

	require.NoError(t, container.Terminate(ctx))
}
