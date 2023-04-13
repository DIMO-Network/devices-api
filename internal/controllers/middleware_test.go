package controllers

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	validUserDeviceID   = "validUserDeviceID"
	invalidUserDeviceID = "invalidUserDeviceID"
	validNFTOwner       = "validNFTOwner"
)

func Middleware_UserDeviceIDTest(t *testing.T) {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	logger := test.Logger()
	userClient := &test.FakeUserClient{Response: []test.UserClientResp{
		{
			ID:      testUserID,
			Address: test.MkAddr(1),
		},
		{
			ID:      validNFTOwner,
			Address: test.MkAddr(2),
		},
	}}
	middleware := NewMiddleware(nil, pdb.DBS, userClient, logger)

	app := test.SetupAppFiber(*logger)

	grouped := app.Group("/user/devices/:userDeviceID", test.AuthInjectorTestHandler(testUserID), middleware.DeviceOwnershipMiddleware)
	grouped.Get("/validUserDeviceID", func(c *fiber.Ctx) error {
		udi := c.Params("userDeviceID")
		return c.JSON(udi)
	})

	uDevice := models.UserDevice{
		ID:                 validUserDeviceID,
		UserID:             testUserID,
		DeviceDefinitionID: "deviceDefinition" + testUserID,
	}
	err := uDevice.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	nftOwner := models.VehicleNFT{
		UserDeviceID: null.StringFrom(validNFTOwner),
		OwnerAddress: null.BytesFrom(common.FromHex(test.MkAddr(1).String())),
	}
	err = nftOwner.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/validUserDeviceID", validUserDeviceID), "")
	response, err := app.Test(request)
	assert.Equal(t, fiber.StatusOK, response.StatusCode)
	assert.Nil(t, err)

	request = test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/validUserDeviceID", invalidUserDeviceID), "")
	response, err = app.Test(request)
	assert.Equal(t, fiber.ErrNotFound, response.StatusCode)
	assert.Nil(t, err)

	request = test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/validUserDeviceID", invalidUserDeviceID), "")
	response, err = app.Test(request)
	assert.Equal(t, fiber.ErrNotFound, response.StatusCode)
	assert.Nil(t, err)

	fmt.Printf("shutting down postgres at with session: %s \n", container.SessionID())
	if err := container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}

}

func Middleware_NFTOwnershipTest(t *testing.T) {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	logger := test.Logger()
	userClient := &test.FakeUserClient{Response: []test.UserClientResp{
		{
			ID:      validNFTOwner,
			Address: test.MkAddr(2),
		},
	}}
	middleware := NewMiddleware(nil, pdb.DBS, userClient, logger)

	app := test.SetupAppFiber(*logger)

	grouped := app.Group("/user/devices/:userDeviceID", test.AuthInjectorTestHandler(validNFTOwner), middleware.DeviceOwnershipMiddleware)
	grouped.Get("/validUserDeviceID", func(c *fiber.Ctx) error {
		udi := c.Params("userDeviceID")
		return c.JSON(udi)
	})

	nftOwner := models.VehicleNFT{
		UserDeviceID: null.StringFrom(validNFTOwner),
		OwnerAddress: null.BytesFrom(common.FromHex(test.MkAddr(1).String())),
	}
	err := nftOwner.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/validUserDeviceID", validNFTOwner), "")
	response, err := app.Test(request)
	assert.Equal(t, fiber.ErrNotFound, response.StatusCode)
	assert.Nil(t, err)

	fmt.Printf("shutting down postgres at with session: %s \n", container.SessionID())
	if err := container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}

}
