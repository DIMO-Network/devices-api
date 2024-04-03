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

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
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
	test.SetupCreateVehicleNFT(s.T(), ud.ID, ud.VinIdentifier.String, big.NewInt(1), null.BytesFrom(addr.Bytes()), s.pdb)
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

	vnft := test.SetupCreateVehicleNFT(s.T(), ud.ID, ud.VinIdentifier.String, big.NewInt(1), null.BytesFrom(addr.Bytes()), s.pdb)
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

	if err := vnft.Reload(context.Background(), s.pdb.DBS().Reader); err != nil {
		s.T().Fatal(err)
	}

	s.Require().NotEmpty(vnft.BurnRequestID)

}

func (s *UserDevicesControllerTestSuite) TestGetPostMint() {
	privKey, err := crypto.GenerateKey()
	s.Require().NoError(err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	email := "some@email.com"
	eth := addr.Hex()
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             "123123",
		DeviceDefinitionID: ksuid.New().String(),
		CountryCode:        null.StringFrom("USA"),
		Name:               null.StringFrom("Chungus"),
		VinConfirmed:       true,
		VinIdentifier:      null.StringFrom("4Y1SL65848Z411439"),
	}

	err = ud.Insert(context.Background(), s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	user := test.BuildGetUserGRPC(ud.UserID, &email, &eth, &pb.UserReferrer{})

	attrs := []*grpc.DeviceTypeAttribute{
		{
			Name:  "fuel_tank_capacity_gal",
			Value: "15",
		},
		{
			Name:  "mpg",
			Value: "20",
		},
	}
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Times(2).Return(&grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: ud.DeviceDefinitionID,
		Verified:           true,
		DeviceAttributes:   attrs,
		Make: &grpc.DeviceMake{
			TokenId: 7,
			Name:    "Toyota",
		},
		Type: &grpc.DeviceType{
			Model: "Camry",
			Year:  2023,
		},
	}, nil)

	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: ud.UserID}).Times(2).Return(user, nil)
	s.s3 = &mockS3Client{}
	sp := smock.NewSyncProducer(s.T(), nil)
	sp.ExpectSendMessageAndSucceed()
	s.controller.producer = sp

	getRequest := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/commands/mint", ud.ID), "")
	getResp, err := s.app.Test(getRequest)
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, getResp.StatusCode)

	var td signer.TypedData
	s.Require().NoError(json.NewDecoder(getResp.Body).Decode(&td))

	bvs := registry.MintVehicleSign{
		ManufacturerNode: big.NewInt(7),
		Owner:            common.HexToAddress(*user.EthereumAddress),
		Attributes:       []string{"Make", "Model", "Year"},
		Infos:            []string{"Toyota", "Camry", "2023"},
	}

	c := registry.Client{
		Producer:     s.controller.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(0),
			Address: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			Name:    "DIMO",
			Version: "1",
		},
	}

	hash, err := c.Hash(&bvs)
	s.Require().NoError(err)

	userSig, err := crypto.Sign(hash, privKey)
	s.Require().NoError(err)

	userSig[64] += 27

	br := new(MintRequest)
	br.Signature = hexutil.Encode(userSig)
	br.ImageData = "data:image/png;base64,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"
	br.ImageDataTransparent = br.ImageData

	inp, err := json.Marshal(br)
	s.Require().NoError(err)

	request := test.BuildRequest("POST", fmt.Sprintf("/user/devices/%s/commands/mint", ud.ID), string(inp))
	response, err := s.app.Test(request)
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, response.StatusCode)

}
