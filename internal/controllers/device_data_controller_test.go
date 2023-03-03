package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const migrationsDirRelPath = "../../migrations"

func TestUserDevicesController_GetUserDeviceStatus(t *testing.T) {
	// arrange global db and route setup
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

	deviceDefIntSvc := mock_services.NewMockDeviceDefinitionIntegrationService(mockCtrl)
	deviceDefSvc := mock_services.NewMockDeviceDefinitionService(mockCtrl)
	scClient := mock_services.NewMockSmartcarClient(mockCtrl)
	scTaskSvc := mock_services.NewMockSmartcarTaskService(mockCtrl)
	teslaSvc := mock_services.NewMockTeslaService(mockCtrl)
	teslaTaskService := mock_services.NewMockTeslaTaskService(mockCtrl)
	nhtsaService := mock_services.NewMockINHTSAService(mockCtrl)
	autoPiIngest := mock_services.NewMockIngestRegistrar(mockCtrl)
	deviceDefinitionIngest := mock_services.NewMockDeviceDefinitionRegistrar(mockCtrl)
	autoPiTaskSvc := mock_services.NewMockAutoPiTaskService(mockCtrl)
	drivlyTaskSvc := mock_services.NewMockDrivlyTaskService(mockCtrl)

	testUserID := "123123"
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger, deviceDefSvc, deviceDefIntSvc, &fakeEventService{},
		scClient, scTaskSvc, teslaSvc, teslaTaskService, nil, nil, nhtsaService, autoPiIngest, deviceDefinitionIngest,
		autoPiTaskSvc, nil, nil, drivlyTaskSvc, nil)
	app := fiber.New()
	app.Get("/user/devices/:userDeviceID/status", test.AuthInjectorTestHandler(testUserID), c.GetUserDeviceStatus)

	t.Run("GET - device status merge autopi and smartcar", func(t *testing.T) {
		// arrange db, insert some user_devices
		autoPiInteg := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
		smartCarInt := test.BuildIntegrationGRPC(constants.SmartCarVendor, 0, 0)
		dd := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "Mach E", 2020, autoPiInteg)
		ud := test.SetupCreateUserDevice(t, testUserID, dd[0].DeviceDefinitionId, nil, pdb)
		const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
		const deviceID = "device123"
		_ = test.SetupCreateAutoPiUnit(t, testUserID, unitID, func(s string) *string { return &s }(deviceID), pdb)
		_ = test.SetupCreateUserDeviceAPIIntegration(t, unitID, deviceID, ud.ID, autoPiInteg.Id, pdb)
		_ = test.SetupCreateUserDeviceAPIIntegration(t, unitID, deviceID, ud.ID, smartCarInt.Id, pdb)
		// SC data setup to  older
		smartCarData := models.UserDeviceDatum{
			UserDeviceID:        ud.ID,
			Data:                null.JSONFrom([]byte(`{"oil": 0.6859999895095825, "range": 187.79, "tires": {"backLeft": 244, "backRight": 280, "frontLeft": 244, "frontRight": 252}, "charging": false, "latitude": 33.675048828125, "odometer": 195677.59375, "longitude": -117.85894775390625, "timestamp": "2022-05-18T16:49:37.879182265Z", "vehicleId": "0ef49636-28be-4f0f-8b08-4121137f0d5d", "fuelPercentRemaining": 0.4}`)),
			CreatedAt:           time.Now().Add(time.Minute * -5),
			UpdatedAt:           time.Now().Add(time.Minute * -5),
			LastOdometerEventAt: null.TimeFrom(time.Now().Add(time.Minute * -5)),
			IntegrationID:       smartCarInt.Id,
		}
		err := smartCarData.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
		// newer autopi data, expect to replace lat/long
		autoPiData := models.UserDeviceDatum{
			UserDeviceID:  ud.ID,
			Data:          null.JSONFrom([]byte(`{"latitude": 33.75, "longitude": -117.91}`)),
			CreatedAt:     time.Now().Add(time.Minute * -1),
			UpdatedAt:     time.Now().Add(time.Minute * -1),
			IntegrationID: autoPiInteg.Id,
		}
		err = autoPiData.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)

		request := test.BuildRequest("GET", "/user/devices/"+ud.ID+"/status", "")
		response, _ := app.Test(request)
		body, _ := io.ReadAll(response.Body)

		if assert.Equal(t, fiber.StatusOK, response.StatusCode) == false {
			fmt.Println("response body: " + string(body))
		}

		snapshot := new(DeviceSnapshot)
		err = json.Unmarshal(body, snapshot)
		assert.NoError(t, err)

		assert.Equal(t, 187.79, *snapshot.Range)
		assert.Equal(t, false, *snapshot.Charging)
		assert.Equal(t, 244.0, snapshot.TirePressure.BackLeft)
		assert.Equal(t, 195677.59375, *snapshot.Odometer)
		assert.Equal(t, 33.75, *snapshot.Latitude, "expected autopi latitude")
		assert.Equal(t, -117.91, *snapshot.Longitude, "expected autopi longitude")

		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
}

func Test_sortByJSONOdometerAsc(t *testing.T) {
	udd := models.UserDeviceDatumSlice{
		&models.UserDeviceDatum{
			UserDeviceID:  "123",
			Data:          null.JSONFrom([]byte(`{ "odometer": 100 }`)),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			IntegrationID: "123",
		},
		&models.UserDeviceDatum{
			UserDeviceID:  "123",
			Data:          null.JSONFrom([]byte(`{ "odometer": 90}`)),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			IntegrationID: "345",
		},
	}
	// validate setup is ok
	assert.Equal(t, float64(100), gjson.GetBytes(udd[0].Data.JSON, "odometer").Float())
	assert.Equal(t, float64(90), gjson.GetBytes(udd[1].Data.JSON, "odometer").Float())
	// sort and validate
	sortByJSONOdometerAsc(udd)
	assert.Equal(t, float64(90), gjson.GetBytes(udd[0].Data.JSON, "odometer").Float())
	assert.Equal(t, float64(100), gjson.GetBytes(udd[1].Data.JSON, "odometer").Float())
}
