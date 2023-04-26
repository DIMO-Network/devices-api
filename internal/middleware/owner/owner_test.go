package owner

import (
	"context"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	pb_devices "github.com/DIMO-Network/devices-api/pkg/grpc"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOwnerMiddleware(t *testing.T) {
	userID := "louxUser"
	userAddr := "0x1ABC7154748d1ce5144478cdeB574ae244b939B5"
	otherUserID := "stanleyUser"
	otherAddr := "0x3AC4f4Ae05b75b97bfC71Ea518913007FdCaab70"
	userDeviceID := "2OeRoU9VmbFVpgpPy3BjY2WsMMm"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, "../../../migrations")
	logger := test.Logger()

	usersClient := &test.UsersClient{}
	devicesClient := &test.DevicesClient{}
	middleware := New(pdb, usersClient, devicesClient, logger)

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
		DeviceUserID        string
		UserExists          bool
		UserEthereumAddress string
		DeviceOwnerAddress  string
		ExpectedCode        int
	}{
		{
			Name:         "NoDevice",
			ExpectedCode: 404,
		},
		{
			Name:             "UserIDMatch",
			UserExists:       true,
			UserDeviceUserID: userID,
			DeviceUserID:     userID,
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
			UserDeviceUserID:    userID,
			UserExists:          true,
			UserEthereumAddress: userAddr,
			ExpectedCode:        404,
		},
		{
			Name:                "UserIDMismatchEthereumAddressMatch",
			UserDeviceUserID:    otherUserID,
			DeviceUserID:        userID,
			DeviceOwnerAddress:  userAddr,
			UserExists:          true,
			UserEthereumAddress: userAddr,
			ExpectedCode:        200,
		},
		{
			Name:                "UserIDMismatchEthereumAddressMismatch",
			UserDeviceUserID:    otherUserID,
			DeviceOwnerAddress:  otherAddr,
			UserExists:          true,
			UserEthereumAddress: userAddr,
			ExpectedCode:        404,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := models.UserDevices().DeleteAll(ctx, pdb.DBS().Writer)
			require.NoError(t, err)

			usersClient.Store = map[string]*pb.User{}
			devicesClient.Store = map[string]*pb_devices.UserDevice{}

			if c.UserExists {
				u := &pb.User{Id: userID}
				if c.UserEthereumAddress != "" {
					u.EthereumAddress = &c.UserEthereumAddress
				}
				usersClient.Store[userID] = u
			}

			if c.DeviceUserID != "" {
				d := &pb_devices.UserDevice{Id: userDeviceID}
				if c.DeviceOwnerAddress != "" {
					d.OwnerAddress = common.Hex2Bytes(c.DeviceOwnerAddress)
				}
				d.UserId = userID
				devicesClient.Store[userDeviceID] = d
			}

			res, err := app.Test(request)
			require.Nil(t, err)
			assert.Equal(t, c.ExpectedCode, res.StatusCode)
		})
	}

	require.NoError(t, container.Terminate(ctx))
}
