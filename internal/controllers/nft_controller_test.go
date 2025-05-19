package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"go.uber.org/mock/gomock"
)

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

func (s *UserDevicesControllerTestSuite) TestUpdateVINV2_japanChasisNumber() {
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
		VIN:         "AGH30-0397617",
		CountryCode: "JPN",
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

	s.Equal("JPN", userDevice.CountryCode.String)
	s.Equal(`{"canProtocol": "7", "postal_code": null, "powertrainType": "ICE", "geoDecodedCountry": null, "geoDecodedStateProv": null}`,
		string(userDevice.Metadata.JSON))
	s.Equal("AGH30-0397617", userDevice.VinIdentifier.String)
}
