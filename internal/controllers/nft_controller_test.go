package controllers

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
)

func TestNFTController_GetDcnNFTMetadata(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()
	deviceDefSvc := mock_services.NewMockDeviceDefinitionService(mockCtrl)

	c := NewNFTController(&config.Settings{Port: "3000"}, pdb.DBS, &logger,
		nil, deviceDefSvc, nil, nil, nil, nil, nil)

	app := fiber.New()
	app.Get("/dcn/:nodeID", c.GetDcnNFTMetadata)

	t.Run("GET - dcn by node id decimal", func(t *testing.T) {
		ndhex, _ := hex.DecodeString("37B0403A1C4B24E0865A97B4C64206E478444EC9B9D21947048DFDC31BE9DC7F")
		ownerHex, _ := hex.DecodeString("B8E514DA5E7B2918AEBC139AE7CBEFC3727F05D3")
		// setup data
		dcn := models.DCN{
			NFTNodeID:              ndhex,
			OwnerAddress:           null.BytesFrom(ownerHex),
			Name:                   null.StringFrom("reddy.dimo"),
			Expiration:             null.TimeFrom(time.Now()),
			NFTNodeBlockCreateTime: null.TimeFrom(time.Now()),
		}
		err := dcn.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		require.NoError(t, err)

		request := test.BuildRequest("GET", "/dcn/25188615033903404929663096463794904537890497498749865140845237535414961167487", "")
		response, _ := app.Test(request)
		body, _ := io.ReadAll(response.Body)

		if assert.Equal(t, fiber.StatusOK, response.StatusCode) == false {
			fmt.Println("response body: " + string(body))
		}
		fmt.Println(string(body))

		assert.Equal(t, "reddy.dimo", gjson.GetBytes(body, "name").String())
		assert.Equal(t, "reddy.dimo, a DCN name.", gjson.GetBytes(body, "description").String())
		assert.Equal(t, "/v1/dcn/25188615033903404929663096463794904537890497498749865140845237535414961167487/image", gjson.GetBytes(body, "image").String())
		assert.Equal(t, "Creation Date", gjson.GetBytes(body, "attributes.0.trait_type").String())
		assert.Equal(t, "Registration Date", gjson.GetBytes(body, "attributes.1.trait_type").String())
		assert.Equal(t, "Expiration Date", gjson.GetBytes(body, "attributes.2.trait_type").String())
		assert.Equal(t, "Character Set", gjson.GetBytes(body, "attributes.3.trait_type").String())
		assert.Equal(t, "Length", gjson.GetBytes(body, "attributes.4.trait_type").String())
		assert.Equal(t, "Nodehash", gjson.GetBytes(body, "attributes.5.trait_type").String())
	})
}
