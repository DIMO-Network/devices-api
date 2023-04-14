package controllers

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const genericUserID = "genericUserID"

func TestMiddleware_DeviceDoesNotExist(t *testing.T) {

	deviceDoesntExist := "deviceDoesntExist"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	logger := test.Logger()

	userClient := &test.FakeUserClient{Response: []test.UserClientResp{}}
	middleware := NewMiddleware(nil, pdb.DBS, userClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/user/devices/:userDeviceID/test", test.AuthInjectorTestHandler(genericUserID), middleware.DeviceOwnershipMiddleware, func(c *fiber.Ctx) error {
		return nil
	})

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/test", deviceDoesntExist), "")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, response.StatusCode, fiber.StatusNotFound)
	body, _ := io.ReadAll(response.Body)
	assert.Equal(t, string(body), `{"code":404,"message":"user not found in users api mock"}`)

	fmt.Printf("shutting down postgres at with session: %s \n", container.SessionID())
	if err := container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestMiddleware_ExistsInUsersDevicesTable(t *testing.T) {

	deviceOwner := "deviceOwner"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	logger := test.Logger()

	userClient := &test.FakeUserClient{Response: []test.UserClientResp{
		{
			ID:      genericUserID,
			Address: test.MkAddr(1),
		},
	}}
	middleware := NewMiddleware(nil, pdb.DBS, userClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/user/devices/:userDeviceID/test", test.AuthInjectorTestHandler(genericUserID), middleware.DeviceOwnershipMiddleware, func(c *fiber.Ctx) error {
		return nil
	})

	test.SetupCreateUserDeviceForMiddleware(t, deviceOwner, genericUserID, "deviceDefID", nil, "00000000000000001", pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/test", deviceOwner), "")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode)

	fmt.Printf("shutting down postgres at with session: %s \n", container.SessionID())
	if err := container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestMiddleware_ExistsInVehicleNFTTable(t *testing.T) {

	nftOwnerDeviceID := "nftOwnerDeviceID"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	logger := test.Logger()

	userClient := &test.FakeUserClient{Response: []test.UserClientResp{
		{
			ID:      nftOwnerDeviceID,
			Address: test.MkAddr(1),
		},
	}}
	middleware := NewMiddleware(nil, pdb.DBS, userClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/user/devices/:userDeviceID/test", test.AuthInjectorTestHandler(genericUserID), middleware.DeviceOwnershipMiddleware, func(c *fiber.Ctx) error {
		return nil
	})

	test.SetupCreateVehicleNFTForMiddleware(t, test.MkAddr(1), genericUserID, nftOwnerDeviceID, big.NewInt(1), pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/test", nftOwnerDeviceID), "")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode)

	fmt.Printf("shutting down postgres at with session: %s \n", container.SessionID())
	if err := container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestMiddleware_DeviceNotMinted(t *testing.T) {

	nftOwnerDeviceID := "nftOwnerDeviceID"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	logger := test.Logger()

	userClient := &test.FakeUserClient{Response: []test.UserClientResp{
		{
			ID:      genericUserID,
			Address: test.MkAddr(1),
		},
	}}
	middleware := NewMiddleware(nil, pdb.DBS, userClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/user/devices/:userDeviceID/test", test.AuthInjectorTestHandler(genericUserID), middleware.DeviceOwnershipMiddleware, func(c *fiber.Ctx) error {
		return nil
	})

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/test", nftOwnerDeviceID), "")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, response.StatusCode, fiber.StatusNotFound)
	body, _ := io.ReadAll(response.Body)
	assert.Equal(t, string(body), `{"code":404,"message":"User does not own device or nft associated with userDeviceID."}`)

	fmt.Printf("shutting down postgres at with session: %s \n", container.SessionID())
	if err := container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestMiddleware_InvalidAddress(t *testing.T) {

	nftOwnerDeviceID := "nftOwnerDeviceID"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	logger := test.Logger()

	userClient := &test.FakeUserClient{Response: []test.UserClientResp{
		{
			ID: genericUserID,
		},
	}}
	middleware := NewMiddleware(nil, pdb.DBS, userClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/user/devices/:userDeviceID/test", test.AuthInjectorTestHandler(genericUserID), middleware.DeviceOwnershipMiddleware, func(c *fiber.Ctx) error {
		return nil
	})

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/test", nftOwnerDeviceID), "")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, response.StatusCode, fiber.StatusNotFound)
	body, _ := io.ReadAll(response.Body)
	t.Log(string(body))
	assert.Equal(t, string(body), `{"code":404,"message":"User does not have an Ethereum address and does not own device associated with userDeviceID."}`)

	fmt.Printf("shutting down postgres at with session: %s \n", container.SessionID())
	if err := container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestMiddleware_Locals(t *testing.T) {

	deviceOwner := "deviceOwner"

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	logger := test.Logger()

	userClient := &test.FakeUserClient{Response: []test.UserClientResp{
		{
			ID:      genericUserID,
			Address: test.MkAddr(1),
		},
	}}
	middleware := NewMiddleware(nil, pdb.DBS, userClient, logger)

	app := test.SetupAppFiber(*logger)
	app.Get("/user/devices/:userDeviceID/test", test.AuthInjectorTestHandler(genericUserID), middleware.DeviceOwnershipMiddleware, func(c *fiber.Ctx) error {
		udi := c.Locals("userDeviceId")
		userID := c.Locals("userId")

		if udi != deviceOwner || userID != genericUserID {
			return errors.New("incorrect vars from local")
		}

		return nil
	})

	test.SetupCreateUserDeviceForMiddleware(t, deviceOwner, genericUserID, "deviceDefID", nil, "00000000000000001", pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%+v/test", deviceOwner), "")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode)

	fmt.Printf("shutting down postgres at with session: %s \n", container.SessionID())
	if err := container.Terminate(ctx); err != nil {
		t.Fatal(err)
	}
}
