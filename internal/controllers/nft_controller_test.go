package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"

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
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
)

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
	s.usersClient.EXPECT().GetUser(gomock.Any(), &pb.GetUserRequest{Id: ud.UserID}).Times(2).Return(user, nil)

	sp := smock.NewSyncProducer(s.T(), nil)
	sp.ExpectSendMessageAndSucceed()
	s.controller.registryClient.Producer = sp

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
		},
	}

	hash, err := client.Hash(&bvs)
	s.Require().NoError(err)

	userSig, err := crypto.Sign(hash, privKey)
	s.Require().NoError(err)

	userSig[64] += 27

	br := new(BurnRequest)
	br.Signature = hexutil.Encode(userSig)

	inp, err := json.Marshal(br)
	s.Require().NoError(err)

	request := test.BuildRequest("POST", fmt.Sprintf("/vehicle/%s/commands/burn", "1"), string(inp))
	response, err := s.app.Test(request)
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, response.StatusCode)

	b, err := io.ReadAll(response.Body)
	fmt.Println(string(b))

	if err := ud.Reload(context.Background(), s.pdb.DBS().Reader); err != nil {
		s.T().Fatal(err)
	}

	s.Require().NotEmpty(ud.BurnRequestID)

}
