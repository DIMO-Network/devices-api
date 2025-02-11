package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/ethereum/go-ethereum/crypto"
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
		DefinitionID:       "ford_escape_2020",
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

func (s *UserDevicesControllerTestSuite) TestUpdateVINV2_setCountryAndProtocol() {
	privKey, err := crypto.GenerateKey()
	s.Require().NoError(err)
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	//email := "some@email.com"
	//eth := addr.Hex()
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Escape", 2020, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionBySlug(gomock.Any(), dd[0].DeviceDefinitionId).Return(dd[0], nil)

	userDevice := test.SetupCreateUserDevice(s.T(), testUserID, dd[0].DeviceDefinitionId, nil, "", s.pdb)
	_ = test.SetupCreateVehicleNFT(s.T(), userDevice, big.NewInt(1), null.BytesFrom(addr.Bytes()), s.pdb)

	input := &UpdateVINReq{
		VIN:         "4Y1SL65848Z411439",
		CountryCode: "USA",
		CANProtocol: "7",
		Signature:   "",
	}
	marshal, _ := json.Marshal(input)
	request := test.BuildRequest("PATCH", fmt.Sprintf("/vehicle/%s/vin", "1"), string(marshal))

	response, err := s.app.Test(request)
	s.Require().NoError(err)
	s.Equal(204, response.StatusCode)

	if err := userDevice.Reload(context.Background(), s.pdb.DBS().Reader); err != nil {
		s.T().Fatal(err)
	}

	s.Equal("USA", userDevice.CountryCode.String)
	s.Equal(`{"canProtocol": "7", "postal_code": null, "powertrainType": "ICE", "geoDecodedCountry": null, "geoDecodedStateProv": null}`,
		string(userDevice.Metadata.JSON))
	s.Equal("4Y1SL65848Z411439", userDevice.VinIdentifier.String)
}
