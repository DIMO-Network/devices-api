package controllers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
	smock "github.com/Shopify/sarama/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
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
	app.Get("/dcn/:tokenID", c.GetDcnNFTMetadata)

	t.Run("GET - dcn by token id decimal", func(t *testing.T) {
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

func (s *UserDevicesControllerTestSuite) TestGetBurn() {
	_, addr, err := test.GenerateWallet()
	s.Require().NoError(err)

	email := "some@email.com"
	eth := addr.Hex()
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             testUserID,
		DeviceDefinitionID: ksuid.New().String(),
		CountryCode:        null.StringFrom("USA"),
		Name:               null.StringFrom("Chungus"),
		VinConfirmed:       true,
		VinIdentifier:      null.StringFrom("4Y1SL65848Z411439"),
	}

	err = ud.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	test.SetupCreateVehicleNFT(s.T(), ud, big.NewInt(1), null.BytesFrom(addr.Bytes()), s.pdb)
	user := test.BuildGetUserGRPC(ud.UserID, &email, &eth, &pb.UserReferrer{})
	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: ud.UserID}).Return(user, nil)

	request := test.BuildRequest("GET", fmt.Sprintf("/vehicle/%s/commands/burn", "1"), "")
	response, err := s.app.Test(request)
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, response.StatusCode)
}

func (s *UserDevicesControllerTestSuite) TestPostBurn() {
	privKey, err := crypto.GenerateKey()
	s.Require().NoError(err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	email := "some@email.com"
	eth := addr.Hex()
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             testUserID,
		DeviceDefinitionID: ksuid.New().String(),
		CountryCode:        null.StringFrom("USA"),
		Name:               null.StringFrom("Chungus"),
		VinConfirmed:       true,
		VinIdentifier:      null.StringFrom("4Y1SL65848Z411439"),
	}

	err = ud.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	vnft := test.SetupCreateVehicleNFT(s.T(), ud, big.NewInt(1), null.BytesFrom(addr.Bytes()), s.pdb)
	user := test.BuildGetUserGRPC(ud.UserID, &email, &eth, &pb.UserReferrer{})
	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: ud.UserID}).Return(user, nil)
	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: ud.UserID}).Return(user, nil)

	sp := smock.NewSyncProducer(s.T(), nil)
	sp.ExpectSendMessageAndSucceed()
	s.controller.producer = sp

	getRequest := test.BuildRequest("GET", fmt.Sprintf("/vehicle/%s/commands/burn", "1"), "")
	getResp, err := s.app.Test(getRequest)
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, getResp.StatusCode)

	var td signer.TypedData
	s.Require().NoError(json.NewDecoder(getResp.Body).Decode(&td))

	tkn, ok := vnft.TokenID.Int64()
	s.Require().True(ok)

	bvs := registry.BurnVehicleSign{
		TokenID: big.NewInt(int64(tkn)),
	}

	client := registry.Client{
		Producer:     s.controller.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(s.controller.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(s.controller.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	hash, err := client.Hash(&bvs)
	s.Require().NoError(err)

	userSig, err := crypto.Sign(hash, privKey)
	s.Require().NoError(err)

	userSig[64] += 27

	br := new(BurnRequest)
	br.TokenID = bvs.TokenID
	br.Signature = hexutil.Encode(userSig)

	inp, err := json.Marshal(br)
	s.Require().NoError(err)

	request := test.BuildRequest("POST", fmt.Sprintf("/vehicle/%s/commands/burn", "1"), string(inp))
	response, err := s.app.Test(request)
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, response.StatusCode)

	if err := ud.Reload(context.Background(), s.pdb.DBS().Reader); err != nil {
		s.T().Fatal(err)
	}

	s.Require().NotEmpty(ud.BurnRequestID)

}
