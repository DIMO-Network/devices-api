package test

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/shared"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/docker/go-connections/nat"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const testDbName = "devices_api"

// StartContainerDatabase starts postgres container with default test settings, and migrates the db. Caller must terminate container.
func StartContainerDatabase(ctx context.Context, t *testing.T, migrationsDirRelPath string) (db.Store, testcontainers.Container) {
	settings := getTestDbSettings()
	pgPort := "5432/tcp"
	dbURL := func(_ string, port nat.Port) string {
		return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", settings.DB.User, settings.DB.Password, port.Port(), settings.DB.Name)
	}
	cr := testcontainers.ContainerRequest{
		Image:        "postgres:16.6-alpine",
		Env:          map[string]string{"POSTGRES_USER": settings.DB.User, "POSTGRES_PASSWORD": settings.DB.Password, "POSTGRES_DB": settings.DB.Name},
		ExposedPorts: []string{pgPort},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
		WaitingFor:   wait.ForSQL(nat.Port(pgPort), "postgres", dbURL),
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
	if err := goose.RunContext(ctx, "up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
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
			// copied from controllers.helpers.ErrorHandler - but temporarily in here to see if resolved circular deps issue
			code := fiber.StatusInternalServerError // Default 500 statuscode

			e, fiberTypeErr := err.(*fiber.Error)
			if fiberTypeErr {
				// Override status code if fiber.Error type
				code = e.Code
			}
			logger.Err(err).Str("httpStatusCode", strconv.Itoa(code)).
				Str("httpMethod", c.Method()).
				Str("httpPath", c.Path()).
				Msg("caught an error from http request")

			return c.Status(code).JSON(fiber.Map{
				"code":    code,
				"message": err.Error(),
			})
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
func AuthInjectorTestHandler(userID string, userEthAddr *common.Address) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := jwt.MapClaims{
			"sub": userID,
			"nbf": time.Now().Unix(),
		}
		if userEthAddr != nil {
			claims["ethereum_address"] = userEthAddr.Hex()
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

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

func SetupCreateUserDevice(t *testing.T, testUserID string, ddID string, metadata *[]byte, vin string, pdb db.Store) models.UserDevice {
	ud := models.UserDevice{
		ID:           ksuid.New().String(),
		UserID:       testUserID,
		DefinitionID: ddID,
		CountryCode:  null.StringFrom("USA"),
		Name:         null.StringFrom("Chungus"),
	}
	if len(vin) == 17 {
		ud.VinIdentifier = null.StringFrom(vin)
		ud.VinConfirmed = true
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

func SetupCreateUserDeviceWithDeviceID(t *testing.T, testUserID string, deviceID string, definitionID string, metadata *[]byte, vin string, pdb db.Store) models.UserDevice {
	ud := models.UserDevice{
		ID:           deviceID,
		UserID:       testUserID,
		DefinitionID: definitionID,
		CountryCode:  null.StringFrom("USA"),
		Name:         null.StringFrom("Chungus"),
	}
	if len(vin) == 17 {
		ud.VinIdentifier = null.StringFrom(vin)
		ud.VinConfirmed = true
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

func SetupCreateUserDeviceWithTokenID(t *testing.T, testUserID string, tokenID *big.Int, definitionID string, metadata *[]byte, vin string, pdb db.Store) models.UserDevice {
	ud := models.UserDevice{
		ID:           ksuid.New().String(),
		TokenID:      types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
		UserID:       testUserID,
		DefinitionID: definitionID,
		CountryCode:  null.StringFrom("USA"),
		Name:         null.StringFrom("Chungus"),
	}
	if len(vin) == 17 {
		ud.VinIdentifier = null.StringFrom(vin)
		ud.VinConfirmed = true
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

func SetupCreateAftermarketDevice(t *testing.T, userID string, bytes []byte, unitID string, deviceID *string, pdb db.Store) *models.AftermarketDevice {
	amd := models.AftermarketDevice{
		EthereumAddress: bytes, // pkey
		Serial:          unitID,
		UserID:          null.StringFrom(userID),
	}
	if deviceID != nil {
		amdMD := map[string]any{"autoPiDeviceId": *deviceID}
		_ = amd.Metadata.Marshal(amdMD)
	}
	err := amd.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return &amd
}

func SetupCreateMintedAftermarketDevice(t *testing.T, userID, unitID string, tokenID *big.Int, addr common.Address, deviceID *string, pdb db.Store) *models.AftermarketDevice {
	amd := models.AftermarketDevice{
		Serial:          unitID,
		UserID:          null.StringFrom(userID),
		TokenID:         types.NewDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0)),
		EthereumAddress: addr.Bytes(),
	}
	if deviceID != nil {
		amdMD := map[string]any{"autoPiDeviceId": *deviceID}
		_ = amd.Metadata.Marshal(amdMD)
	}
	err := amd.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return &amd
}

func SetupCreateVehicleNFT(t *testing.T, userDevice models.UserDevice, tokenID *big.Int, ownerAddr null.Bytes, pdb db.Store) *models.UserDevice {

	mint := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusConfirmed,
	}
	err := mint.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	userDevice.TokenID = types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tokenID, 0))
	userDevice.OwnerAddress = ownerAddr
	userDevice.MintRequestID = null.StringFrom(mint.ID)

	_, err = userDevice.Update(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return &userDevice
}

func SetupCreateVehicleNFTForMiddleware(t *testing.T, addr common.Address, userID, userDeviceID string, tokenID int64, pdb db.Store) *models.UserDevice {
	mint := models.MetaTransactionRequest{
		ID: ksuid.New().String(),
	}
	err := mint.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	ud := models.UserDevice{
		ID:                 userDeviceID,
		UserID:             userID,
		DeviceDefinitionID: "ddID",
		DefinitionID:       "ford_escape_2020",
		CountryCode:        null.StringFrom("USA"),
		Name:               null.StringFrom("Chungus"),
		VinIdentifier:      null.StringFrom("00000000000000001"),
		MintRequestID:      null.StringFrom(mint.ID),
		OwnerAddress:       null.BytesFrom(common.FromHex(addr.String())),
		VinConfirmed:       true,
		TokenID:            types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0)),
	}
	err = ud.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	return &ud
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
		udapiInt.Serial = null.StringFrom(autoPiUnitID)
		_ = udapiInt.Metadata.UnmarshalJSON([]byte(md))
	}
	err := udapiInt.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return udapiInt
}

var MkAddr = func(i int) common.Address {
	return common.BigToAddress(big.NewInt(int64(i)))
}

type UsersClient struct {
	Store map[string]*pb.User
}

func (c *UsersClient) GetUser(_ context.Context, in *pb.GetUserRequest, _ ...grpc.CallOption) (*pb.User, error) {
	u, ok := c.Store[in.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "No user with that id found.")
	}
	return u, nil
}

func SetupCreateAutoPiJob(t *testing.T, jobID, deviceID, cmd, userDeviceID, state, commandResult string, pdb db.Store) *models.AutopiJob {
	autopiJob := models.AutopiJob{
		ID:             jobID,
		AutopiDeviceID: deviceID,
		Command:        cmd,
		State:          state,
		UserDeviceID:   null.StringFrom(userDeviceID),
	}

	if commandResult != "" {
		_ = autopiJob.CommandResult.UnmarshalJSON([]byte(commandResult))
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

// BuildIntegrationDefaultGRPC depending on integration vendor, defines an integration object with typical settings. Smartcar refresh limit default is 100 seconds.
func BuildIntegrationDefaultGRPC(id, integrationVendor string, autoPiDefaultTemplateID int, bevTemplateID int, includeAutoPiPowertrainTemplate bool) *ddgrpc.Integration {
	var integration *ddgrpc.Integration
	switch integrationVendor {
	case constants.AutoPiVendor:
		integration = &ddgrpc.Integration{
			Id:                      id,
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
			Id:               id,
			Type:             constants.IntegrationTypeAPI,
			Style:            constants.IntegrationStyleWebhook,
			Vendor:           constants.SmartCarVendor,
			RefreshLimitSecs: 100,
			TokenId:          1,
		}
	case constants.TeslaVendor:
		integration = &ddgrpc.Integration{
			Id:      id,
			Type:    constants.IntegrationTypeAPI,
			Style:   constants.IntegrationStyleOEM,
			Vendor:  constants.TeslaVendor,
			TokenId: 2,
		}
	}
	return integration
}

// BuildIntegrationGRPC depending on integration vendor, defines an integration object with typical settings. Smartcar refresh limit default is 100 seconds.
func BuildIntegrationGRPC(id, integrationVendor string, autoPiDefaultTemplateID int, bevTemplateID int) *ddgrpc.Integration {
	return BuildIntegrationDefaultGRPC(id, integrationVendor, autoPiDefaultTemplateID, bevTemplateID, false)
}

// BuildDeviceDefinitionGRPC generates an array with single device definition, adds integration to response if integration passed in not nil. uses Americas region
func BuildDeviceDefinitionGRPC(deviceDefinitionID string, mk string, model string, year int, integration *ddgrpc.Integration) []*ddgrpc.GetDeviceDefinitionItemResponse {
	// todo can we get rid of deviceDefinitionID?
	// can we get rid of integrations?
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
		Id:                 shared.SlugString(mk) + "_" + shared.SlugString(model) + "_" + strconv.Itoa(year),
		Name:               "Name",
		Ksuid:              deviceDefinitionID,
		Make: &ddgrpc.DeviceMake{
			Id:       ksuid.New().String(),
			Name:     mk,
			NameSlug: shared.SlugString(mk),
		},
		Model:    model,
		Year:     int32(year),
		Verified: true,
	}
	if integration != nil {
		rp.DeviceIntegrations = integrationsToAdd //nolint
	}

	rp.DeviceAttributes = append(rp.DeviceAttributes, &ddgrpc.DeviceTypeAttribute{
		Name:  "powertrain_type",
		Value: "ICE",
	})

	return []*ddgrpc.GetDeviceDefinitionItemResponse{rp}
}

func BuildGetUserGRPC(id string, email *string, ethereumAddress *string, referredBy *pb.UserReferrer) *pb.User {
	return &pb.User{
		Id:              id,
		EthereumAddress: ethereumAddress,
		EmailAddress:    email,
		ReferredBy:      referredBy,
	}
}

func GenerateWallet() (*ecdsa.PrivateKey, *common.Address, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	userAddr := crypto.PubkeyToAddress(privateKey.PublicKey)

	return privateKey, &userAddr, nil
}

// BuildIntegrationForGRPCRequest includes tokenID when creating mock integration
func BuildIntegrationForGRPCRequest(tokenID uint64, vendor string) *ddgrpc.Integration {
	integration := &ddgrpc.Integration{
		Id:                      ksuid.New().String(),
		Type:                    constants.IntegrationTypeHardware,
		Style:                   constants.IntegrationStyleAddon,
		Vendor:                  vendor,
		AutoPiDefaultTemplateId: 0,
		TokenId:                 tokenID,
	}

	return integration
}
