package test

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/docker/go-connections/nat"
	"github.com/ericlagergren/decimal"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

const testDbName = "devices_api"

// StartContainerDatabase starts postgres container with default test settings, and migrates the db. Caller must terminate container.
func StartContainerDatabase(ctx context.Context, t *testing.T, migrationsDirRelPath string) (db.Store, testcontainers.Container) {
	settings := getTestDbSettings()
	pgPort := "5432/tcp"
	dbURL := func(port nat.Port) string {
		return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", settings.DB.User, settings.DB.Password, port.Port(), settings.DB.Name)
	}
	cr := testcontainers.ContainerRequest{
		Image:        "postgres:12.9-alpine",
		Env:          map[string]string{"POSTGRES_USER": settings.DB.User, "POSTGRES_PASSWORD": settings.DB.Password, "POSTGRES_DB": settings.DB.Name},
		ExposedPorts: []string{pgPort},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
		WaitingFor:   wait.ForSQL(nat.Port(pgPort), "postgres", dbURL).Timeout(time.Second * 15),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: cr,
		Started:          true,
	})
	if err != nil {
		return handleContainerStartErr(ctx, err, pgContainer, t)
	}
	mappedPort, err := pgContainer.MappedPort(ctx, nat.Port(pgPort))
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to get container external port"), pgContainer, t)
	}
	fmt.Printf("postgres container session %s ready and running at port: %s \n", pgContainer.SessionID(), mappedPort)
	//defer pgContainer.Terminate(ctx) // this should be done by the caller

	settings.DB.Port = mappedPort.Port()
	pdb := db.NewDbConnectionForTest(ctx, &settings.DB, false)
	for !pdb.IsReady() {
		time.Sleep(500 * time.Millisecond)
	}
	// can't connect to db, dsn=user=postgres password=postgres dbname=devices_api host=localhost port=49395 sslmode=disable search_path=devices_api, err=EOF
	// error happens when calling here
	_, err = pdb.DBS().Writer.Exec(`
		grant usage on schema public to public;
		grant create on schema public to public;
		CREATE SCHEMA IF NOT EXISTS devices_api;
		ALTER USER postgres SET search_path = devices_api, public;
		SET search_path = devices_api, public;
		`)
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrapf(err, "failed to apply schema. session: %s, port: %s",
			pgContainer.SessionID(), mappedPort.Port()), pgContainer, t)
	}
	// add truncate tables func
	_, err = pdb.DBS().Writer.Exec(`
CREATE OR REPLACE FUNCTION truncate_tables() RETURNS void AS $$
DECLARE
    statements CURSOR FOR
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'devices_api' and tablename != 'migrations';
BEGIN
    FOR stmt IN statements LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
    END LOOP;
END;
$$ LANGUAGE plpgsql;
`)
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to create truncate func"), pgContainer, t)
	}

	goose.SetTableName("devices_api.migrations")
	if err := goose.Run("up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to apply goose migrations for test"), pgContainer, t)
	}

	return pdb, pgContainer
}

func handleContainerStartErr(ctx context.Context, err error, container testcontainers.Container, t *testing.T) (db.Store, testcontainers.Container) {
	if err != nil {
		fmt.Println("start container error: " + err.Error())
		if container != nil {
			container.Terminate(ctx) //nolint
		}
		t.Fatal(err)
	}
	return db.Store{}, container
}

// getTestDbSettings builds test db config.settings object
func getTestDbSettings() config.Settings {
	dbSettings := db.Settings{
		Name:               testDbName,
		Host:               "localhost",
		Port:               "6669",
		User:               "postgres",
		Password:           "postgres",
		MaxOpenConnections: 2,
		MaxIdleConnections: 2,
	}
	settings := config.Settings{
		LogLevel:    "info",
		DB:          dbSettings,
		ServiceName: "devices-api",
	}
	return settings
}

// SetupAppFiber sets up app fiber with defaults for testing, like our production error handler.
func SetupAppFiber(logger zerolog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return helpers.ErrorHandler(c, err, logger, "test")
		},
	})
	return app
}

func BuildRequest(method, url, body string) *http.Request {
	req, _ := http.NewRequest(
		method,
		url,
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	return req
}

// AuthInjectorTestHandler injects fake jwt with sub
func AuthInjectorTestHandler(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userID,
			"nbf": time.Now().Unix(),
		})

		c.Locals("user", token)
		return c.Next()
	}
}

// TruncateTables truncates tables for the test db, useful to run as teardown at end of each DB dependent test.
func TruncateTables(db *sql.DB, t *testing.T) {
	_, err := db.Exec(`SELECT truncate_tables();`)
	if err != nil {
		fmt.Println("truncating tables failed.")
		t.Fatal(err)
	}
}

/** Test Setup functions. At some point may want to move elsewhere more generic **/

func Logger() *zerolog.Logger {
	l := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()
	return &l
}

func SetupCreateUserDevice(t *testing.T, testUserID string, ddID string, metadata *[]byte, pdb db.Store) models.UserDevice {
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             testUserID,
		DeviceDefinitionID: ddID,
		CountryCode:        null.StringFrom("USA"),
		Name:               null.StringFrom("Chungus"),
	}
	if metadata == nil {
		// note cannot import enum from services
		md := []byte(`{"powertrainType":"ICE"}`)
		metadata = &md
	}
	ud.Metadata = null.JSONFrom(*metadata)
	err := ud.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return ud
}

func SetupCreateAutoPiUnit(t *testing.T, userID, unitID string, deviceID *string, pdb db.Store) *models.AutopiUnit {
	au := models.AutopiUnit{
		AutopiUnitID:   unitID,
		UserID:         null.StringFrom(userID),
		AutopiDeviceID: null.StringFromPtr(deviceID),
	}
	err := au.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return &au
}

func SetupCreateAutoPiUnitWithToken(t *testing.T, userID, unitID string, tokenID *big.Int, deviceID *string, pdb db.Store) *models.AutopiUnit {
	au := models.AutopiUnit{
		AutopiUnitID:   unitID,
		UserID:         null.StringFrom(userID),
		AutopiDeviceID: null.StringFromPtr(deviceID),
		TokenID:        types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
	}
	err := au.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return &au
}

func SetupCreateVehicleNFT(t *testing.T, userDeviceID, vin string, tokenID *big.Int, pdb db.Store) *models.VehicleNFT {

	mint := models.MetaTransactionRequest{
		ID: ksuid.New().String(),
	}
	err := mint.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	vehicle := models.VehicleNFT{
		Vin:           vin,
		MintRequestID: mint.ID,
		UserDeviceID:  null.StringFrom(userDeviceID),
		TokenID:       types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
	}
	err = vehicle.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return &vehicle
}

// SetupCreateUserDeviceAPIIntegration status set to Active, autoPiUnitId is optional
func SetupCreateUserDeviceAPIIntegration(t *testing.T, autoPiUnitID, externalID, userDeviceID, integrationID string, pdb db.Store) models.UserDeviceAPIIntegration {
	udapiInt := models.UserDeviceAPIIntegration{
		UserDeviceID:  userDeviceID,
		IntegrationID: integrationID,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
		ExternalID:    null.StringFrom(externalID),
	}
	if autoPiUnitID != "" {
		md := fmt.Sprintf(`{"autoPiUnitId": "%s"}`, autoPiUnitID)
		udapiInt.AutopiUnitID = null.StringFrom(autoPiUnitID)
		_ = udapiInt.Metadata.UnmarshalJSON([]byte(md))
	}
	err := udapiInt.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return udapiInt
}

func SetupCreateAutoPiJob(t *testing.T, jobID, deviceID, cmd, userDeviceID string, pdb db.Store) *models.AutopiJob {
	autopiJob := models.AutopiJob{
		ID:             jobID,
		AutopiDeviceID: deviceID,
		Command:        cmd,
		State:          "sent",
		UserDeviceID:   null.StringFrom(userDeviceID),
	}
	err := autopiJob.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return &autopiJob
}

func SetupCreateGeofence(t *testing.T, userID, name string, ud *models.UserDevice, pdb db.Store) *models.Geofence {
	gf := models.Geofence{
		ID:     ksuid.New().String(),
		UserID: userID,
		Name:   name,
		Type:   models.GeofenceTypePrivacyFence,
	}
	err := gf.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	if ud != nil {
		udtgf := models.UserDeviceToGeofence{
			UserDeviceID: ud.ID,
			GeofenceID:   gf.ID,
		}
		err = udtgf.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	return &gf
}

func SetupCreateExternalVINData(t *testing.T, ddID string, ud *models.UserDevice, md map[string][]byte, pdb db.Store) *models.ExternalVinDatum {
	evd := models.ExternalVinDatum{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(ddID),
		Vin:                ud.VinIdentifier.String,
		UserDeviceID:       null.StringFrom(ud.ID),
		RequestMetadata:    null.JSONFrom([]byte(`{"mileage":49957,"zipCode":"48216"}`)), // default request metadata
	}
	if rmd, ok := md["RequestMetadata"]; ok {
		evd.RequestMetadata = null.JSONFrom(rmd)
	}
	if omd, ok := md["OfferMetadata"]; ok {
		evd.OfferMetadata = null.JSONFrom(omd)
	}
	if pmd, ok := md["PricingMetadata"]; ok {
		evd.PricingMetadata = null.JSONFrom(pmd)
	}
	if vmd, ok := md["VincarioMetadata"]; ok {
		evd.VincarioMetadata = null.JSONFrom(vmd)
	}
	if bmd, ok := md["BlackbookMetadata"]; ok {
		evd.BlackbookMetadata = null.JSONFrom(bmd)
	}
	err := evd.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	return &evd
}

// BuildIntegrationGRPC depending on integration vendor, defines an integration object with typical settings. Smartcar refresh limit default is 100 seconds.
func BuildIntegrationDefaultGRPC(integrationVendor string, autoPiDefaultTemplateID int, bevTemplateID int, includeAutoPiPowertrainTemplate bool) *ddgrpc.Integration {
	var integration *ddgrpc.Integration
	switch integrationVendor {
	case constants.AutoPiVendor:
		integration = &ddgrpc.Integration{
			Id:                      ksuid.New().String(),
			Type:                    constants.IntegrationTypeHardware,
			Style:                   constants.IntegrationStyleAddon,
			Vendor:                  constants.AutoPiVendor,
			AutoPiDefaultTemplateId: int32(autoPiDefaultTemplateID),
		}

		if includeAutoPiPowertrainTemplate {
			integration.AutoPiPowertrainTemplate = &ddgrpc.Integration_AutoPiPowertrainTemplate{
				BEV:  int32(bevTemplateID),
				HEV:  10,
				ICE:  10,
				PHEV: 4,
			}
		}
	case constants.SmartCarVendor:
		integration = &ddgrpc.Integration{
			Id:               ksuid.New().String(),
			Type:             constants.IntegrationTypeAPI,
			Style:            constants.IntegrationStyleWebhook,
			Vendor:           constants.SmartCarVendor,
			RefreshLimitSecs: 100,
		}
	case constants.TeslaVendor:
		integration = &ddgrpc.Integration{
			Id:     ksuid.New().String(),
			Type:   constants.IntegrationTypeAPI,
			Style:  constants.IntegrationStyleOEM,
			Vendor: constants.TeslaVendor,
		}
	}
	return integration
}

// BuildIntegrationWithOutAutoPiPowertrainTemplateGRPC depending on integration vendor, defines an integration object with typical settings. Smartcar refresh limit default is 100 seconds.
func BuildIntegrationGRPC(integrationVendor string, autoPiDefaultTemplateID int, bevTemplateID int) *ddgrpc.Integration {
	return BuildIntegrationDefaultGRPC(integrationVendor, autoPiDefaultTemplateID, bevTemplateID, false)
}

// BuildDeviceDefinitionGRPC generates an array with single device definition, adds integration to response if integration passed in not nil. uses Americas region
func BuildDeviceDefinitionGRPC(deviceDefinitionID string, mk string, model string, year int, integration *ddgrpc.Integration) []*ddgrpc.GetDeviceDefinitionItemResponse {
	// todo can we get rid of deviceDefinitionID?
	integrationsToAdd := make([]*ddgrpc.DeviceIntegration, 2)
	if integration != nil {
		integrationsToAdd[0] = &ddgrpc.DeviceIntegration{
			Integration: integration,
			Region:      constants.AmericasRegion.String(),
		}
		integrationsToAdd[1] = &ddgrpc.DeviceIntegration{
			Integration: integration,
			Region:      constants.EuropeRegion.String(),
		}
	}

	rp := &ddgrpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: deviceDefinitionID,
		Name:               "Name",
		Make: &ddgrpc.DeviceMake{
			Id:   ksuid.New().String(),
			Name: mk,
		},
		Type: &ddgrpc.DeviceType{
			Type:  "Vehicle",
			Make:  mk,
			Model: model,
			Year:  int32(year),
		},
		VehicleData: &ddgrpc.VehicleInfo{
			MPG:                 1,
			MPGHighway:          1,
			MPGCity:             1,
			FuelTankCapacityGal: 1,
			FuelType:            "gas",
			Base_MSRP:           1,
			DrivenWheels:        "1",
			NumberOfDoors:       1,
			EPAClass:            "class",
			VehicleType:         "Vehicle",
		},
		//Metadata: dd.Metadata,
		Verified: true,
	}
	if integration != nil {
		rp.DeviceIntegrations = integrationsToAdd
	}

	return []*ddgrpc.GetDeviceDefinitionItemResponse{rp}
}
