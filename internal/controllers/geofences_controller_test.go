package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/devices-api/internal/config"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/IBM/sarama"
	saramamocks "github.com/IBM/sarama/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

type partialFenceCloudEvent struct {
	Data struct {
		H3Indexes []string `json:"h3Indexes"`
	} `json:"data"`
}

func checkForDeviceAndH3(key string, h3Indexes []string) func(*sarama.ProducerMessage) error {
	return func(msg *sarama.ProducerMessage) error {
		kb, _ := msg.Key.Encode()
		if string(kb) != key {
			return fmt.Errorf("expected message to be keyed with %s but got %s", key, string(kb))
		}

		if len(h3Indexes) == 0 {
			if msg.Value != nil {
				return fmt.Errorf("non-nil body when nil was expected")
			}
			return nil
		}

		ev := new(partialFenceCloudEvent)
		vb, _ := msg.Value.Encode()
		if err := json.Unmarshal(vb, ev); err != nil {
			return err
		}
		if len(ev.Data.H3Indexes) != len(h3Indexes) {
			return fmt.Errorf("expected %d H3 indices but got %d", len(h3Indexes), len(ev.Data.H3Indexes))
		}

		set := shared.NewStringSet()
		for _, ind := range h3Indexes {
			set.Add(ind)
		}

		for _, ind := range ev.Data.H3Indexes {
			if !set.Contains(ind) {
				return fmt.Errorf("message contained unexpected H3 index %s", ind)
			}
		}

		return nil
	}
}

type GeofencesControllerTestSuite struct {
	suite.Suite
	pdb          db.Store
	container    testcontainers.Container
	ctx          context.Context
	logger       *zerolog.Logger
	deviceDefSvc *mock_services.MockDeviceDefinitionService
	mockCtrl     *gomock.Controller
}

// SetupSuite starts container db
func (s *GeofencesControllerTestSuite) SetupSuite() {
	s.mockCtrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)

	s.logger = test.Logger()
}

// TearDownTest after each test cleanup eg. truncate tables
func (s *GeofencesControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *GeofencesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

func TestGeofencesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(GeofencesControllerTestSuite))
}

/* Actual Tests */
func (s *GeofencesControllerTestSuite) TestPostGeofence() {
	injectedUserID := ksuid.New().String()
	usersClient := mock_services.NewMockUserServiceClient(s.mockCtrl)
	usersClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(2).Return(&users.User{}, nil)
	producer := saramamocks.NewSyncProducer(s.T(), sarama.NewConfig())
	c := NewGeofencesController(&config.Settings{Port: "3000"}, s.pdb.DBS, s.logger, producer, s.deviceDefSvc, usersClient)
	app := fiber.New()
	app.Post("/user/geofences", test.AuthInjectorTestHandler(injectedUserID), c.Create)
	ud := test.SetupCreateUserDevice(s.T(), injectedUserID, ksuid.New().String(), "", nil, "", s.pdb)
	ud.TokenID = types.NewNullDecimal(decimal.New(1, 0))
	_, err := ud.Update(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	req := CreateGeofence{
		Name:          "Home",
		Type:          "PrivacyFence",
		H3Indexes:     []string{"123", "321"},
		UserDeviceIDs: []string{ud.ID},
	}
	j, _ := json.Marshal(req)

	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.ID, []string{"123", "321"}))
	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.TokenID.String(), []string{"123", "321"}))

	request := test.BuildRequest("POST", "/user/geofences", string(j))
	response, _ := app.Test(request)
	body, _ := io.ReadAll(response.Body)
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("error message: " + string(body))
		assert.Fail(s.T(), "could not create geofence")
	}
	createdID := gjson.Get(string(body), "id").String()
	assert.Len(s.T(), createdID, 27)

	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.ID, []string{"123", "321"}))
	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.TokenID.String(), []string{"123", "321"}))

	// create one without h3 indexes required
	req = CreateGeofence{
		Name:          "Work",
		Type:          "PrivacyFence",
		UserDeviceIDs: []string{ud.ID},
	}
	j, _ = json.Marshal(req)
	request = test.BuildRequest("POST", "/user/geofences", string(j))
	response, _ = app.Test(request)
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode, "expected create OK without h3 indexes") == false {
		body, _ = io.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}
	_ = producer.Close()
}

func (s *GeofencesControllerTestSuite) TestPostGeofenceRespectsWallet() {
	injectedUserID := ksuid.New().String()
	usersClient := mock_services.NewMockUserServiceClient(s.mockCtrl)
	addr := "0x00000000219ab540356cbb839cbe05303d7705fa"
	usersClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(&users.User{EthereumAddress: &addr}, nil)
	producer := saramamocks.NewSyncProducer(s.T(), sarama.NewConfig())
	c := NewGeofencesController(&config.Settings{Port: "3000"}, s.pdb.DBS, s.logger, producer, s.deviceDefSvc, usersClient)
	app := fiber.New()
	app.Post("/user/geofences", test.AuthInjectorTestHandler(injectedUserID), c.Create)
	someOtherUserID := ksuid.New().String()
	ud := test.SetupCreateUserDevice(s.T(), someOtherUserID, ksuid.New().String(), nil, "", s.pdb)
	ud.TokenID = types.NewNullDecimal(decimal.New(1, 0))
	ud.OwnerAddress = null.BytesFrom(common.HexToAddress(addr).Bytes())
	_, err := ud.Update(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	req := CreateGeofence{
		Name:          "Home",
		Type:          "PrivacyFence",
		H3Indexes:     []string{"123", "321"},
		UserDeviceIDs: []string{ud.ID},
	}
	j, _ := json.Marshal(req)

	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.ID, []string{"123", "321"}))
	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.TokenID.String(), []string{"123", "321"}))

	request := test.BuildRequest("POST", "/user/geofences", string(j))
	response, err := app.Test(request)
	s.Require().NoError(err)
	s.Equal(fiber.StatusCreated, response.StatusCode)

}

func (s *GeofencesControllerTestSuite) TestPostGeofence400IfSameName() {
	injectedUserID := ksuid.New().String()
	c := NewGeofencesController(&config.Settings{Port: "3000"}, s.pdb.DBS, s.logger, nil, s.deviceDefSvc, nil)
	app := fiber.New()
	app.Post("/user/geofences", test.AuthInjectorTestHandler(injectedUserID), c.Create)
	ud := test.SetupCreateUserDevice(s.T(), injectedUserID, ksuid.New().String(), "", nil, "", s.pdb)
	test.SetupCreateGeofence(s.T(), injectedUserID, "Home", &ud, s.pdb)

	req := CreateGeofence{
		Name:          "Home",
		Type:          models.GeofenceTypePrivacyFence,
		UserDeviceIDs: []string{ud.ID},
	}
	j, _ := json.Marshal(req)
	request := test.BuildRequest("POST", "/user/geofences", string(j))
	response, _ := app.Test(request)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "expected bad request on duplicate name")
}
func (s *GeofencesControllerTestSuite) TestPostGeofence400IfNotYourDevice() {
	injectedUserID := ksuid.New().String()
	usersClient := mock_services.NewMockUserServiceClient(s.mockCtrl)
	usersClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(&users.User{}, nil)
	c := NewGeofencesController(&config.Settings{Port: "3000"}, s.pdb.DBS, s.logger, nil, s.deviceDefSvc, usersClient)
	app := fiber.New()
	app.Post("/user/geofences", test.AuthInjectorTestHandler(injectedUserID), c.Create)
	otherUserID := "7734"
	ud := test.SetupCreateUserDevice(s.T(), otherUserID, ksuid.New().String(), "", nil, "", s.pdb)

	req := CreateGeofence{
		Name:          "Home",
		Type:          models.GeofenceTypePrivacyFence,
		UserDeviceIDs: []string{ud.ID},
	}
	j, _ := json.Marshal(req)
	request := test.BuildRequest("POST", "/user/geofences", string(j))
	response, _ := app.Test(request)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "expected bad request when trying to attach a fence to a device that isn't ours")

}
func (s *GeofencesControllerTestSuite) TestGetAllUserGeofences() {
	injectedUserID := ksuid.New().String()
	c := NewGeofencesController(&config.Settings{Port: "3000"}, s.pdb.DBS, s.logger, nil, s.deviceDefSvc, nil)
	app := fiber.New()
	app.Get("/user/geofences", test.AuthInjectorTestHandler(injectedUserID), c.GetAll)
	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "escaped", 2020, "ford-escaped", nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd[0].DeviceDefinitionId}).Return(dd, nil)
	ud := test.SetupCreateUserDevice(s.T(), injectedUserID, dd[0].DeviceDefinitionId, "", nil, "", s.pdb)
	test.SetupCreateGeofence(s.T(), injectedUserID, "Home", &ud, s.pdb)

	request, _ := http.NewRequest("GET", "/user/geofences", nil)
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	get := gjson.Get(string(body), "geofences")
	if assert.True(s.T(), get.IsArray()) == false {
		fmt.Println("body: " + string(body))
	}
	assert.Len(s.T(), get.Array(), 1, "expected to find one item in response")
}
func (s *GeofencesControllerTestSuite) TestPutGeofence() {
	injectedUserID := ksuid.New().String()
	usersClient := mock_services.NewMockUserServiceClient(s.mockCtrl)
	usersClient.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(&users.User{}, nil)
	producer := saramamocks.NewSyncProducer(s.T(), sarama.NewConfig())
	c := NewGeofencesController(&config.Settings{Port: "3000"}, s.pdb.DBS, s.logger, producer, s.deviceDefSvc, usersClient)
	app := fiber.New()
	app.Get("/user/geofences", test.AuthInjectorTestHandler(injectedUserID), c.GetAll)
	app.Put("/user/geofences/:geofenceID", test.AuthInjectorTestHandler(injectedUserID), c.Update)

	dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "escaped", 2020, "ford-escaped", nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd[0].DeviceDefinitionId}).Return(dd, nil)
	ud := test.SetupCreateUserDevice(s.T(), injectedUserID, dd[0].DeviceDefinitionId, "", nil, "", s.pdb)
	ud.TokenID = types.NewNullDecimal(decimal.New(1, 0))
	_, err := ud.Update(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	gf := test.SetupCreateGeofence(s.T(), injectedUserID, "something", &ud, s.pdb)

	// The fence is being detached from the device and it has type TriggerEntry anyway.
	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.ID, []string{}))
	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.TokenID.String(), []string{}))

	req := CreateGeofence{
		Name:          "School",
		Type:          "TriggerEntry",
		H3Indexes:     []string{"123", "321", "1234555"},
		UserDeviceIDs: nil,
	}
	j, _ := json.Marshal(req)
	request := test.BuildRequest("PUT", "/user/geofences/"+gf.ID, string(j))
	response, _ := app.Test(request)
	body, _ := io.ReadAll(response.Body)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
		fmt.Println("id: " + gf.ID)
	}
	// validate update was performed
	request, _ = http.NewRequest("GET", "/user/geofences", nil)
	response, _ = app.Test(request)
	body, _ = io.ReadAll(response.Body)
	// assert changes
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	get := gjson.Get(string(body), "geofences").Array()
	assert.Len(s.T(), get, 1)
	// assert against second item in array, which was the created one
	assert.Equal(s.T(), req.Name, get[0].Get("name").String(), "expected name to be updated")
	assert.Equal(s.T(), req.Type, get[0].Get("type").String(), "expected type to be updated")
	assert.Len(s.T(), get[0].Get("h3Indexes").Array(), 3)
	_ = producer.Close()
}
func (s *GeofencesControllerTestSuite) TestDeleteGeofence() {
	injectedUserID := ksuid.New().String()
	producer := saramamocks.NewSyncProducer(s.T(), sarama.NewConfig())
	c := NewGeofencesController(&config.Settings{Port: "3000"}, s.pdb.DBS, s.logger, producer, s.deviceDefSvc, nil)
	app := fiber.New()
	app.Delete("/user/geofences/:geofenceID", test.AuthInjectorTestHandler(injectedUserID), c.Delete)
	ud := test.SetupCreateUserDevice(s.T(), injectedUserID, ksuid.New().String(), "", nil, "", s.pdb)
	gf := test.SetupCreateGeofence(s.T(), injectedUserID, "something", &ud, s.pdb)

	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(ud.ID, []string{}))

	request, _ := http.NewRequest("DELETE", "/user/geofences/"+gf.ID, nil)
	response, _ := app.Test(request)
	// assert
	assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode)

	_ = producer.Close()
}
