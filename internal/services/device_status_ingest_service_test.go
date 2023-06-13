package services

import (
	"context"
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type testEventService struct {
	Buffer []*Event
}

func (e *testEventService) Emit(event *Event) error {
	e.Buffer = append(e.Buffer, event)
	return nil
}

const migrationsDirRelPath = "../../migrations"

func TestVinValidation(t *testing.T) {

	testCases := []struct {
		Name           string
		Data           null.JSON
		ExpectedResult string
	}{
		{
			Name:           "Valid vin",
			Data:           null.JSONFrom([]byte(`{"vin": "JH4DB7560SS004122"}`)),
			ExpectedResult: "pass",
		},
		{
			Name:           "Vin is too short",
			Data:           null.JSONFrom([]byte(`{"vin": "JH4DB7560SS004"}`)),
			ExpectedResult: "",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {

			if _, err := extractVIN(c.Data.JSON); err != nil {
				if c.ExpectedResult != "pass" {
					return
				}
				log.Fatalf("expected test to pass but instead saw an error: %v", err)
			}

			if c.ExpectedResult == "pass" {
				return
			}
			log.Fatal("expected test to fail but it passed")
		})
	}
}

func TestIngestDeviceStatus(t *testing.T) {
	mes := &testEventService{
		Buffer: make([]*Event, 0),
	}
	deviceDefSvc := testDeviceDefSvc{}
	autoPISvc := testAutoPISvc{}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	integs, _ := deviceDefSvc.GetIntegrations(ctx)
	integrationID := integs[0].Id

	ingest := NewDeviceStatusIngestService(pdb.DBS, &logger, mes, deviceDefSvc, autoPISvc, &config.Settings{
		DIMORegistryChainID: 80001,
		DIMORegistryAddr:    "0x4De1bCf2B7E851E31216fC07989caA902A604784",
		VehicleNFTAddress:   "0x90c4d6113ec88dd4bdf12f26db2b3998fd13a144",
		IssuerPrivateKey:    "3qM17-6NmFbwY52XWtzgbtM1ku2ha3wQ6w6hFpE9o1j",
	})

	ud := test.SetupCreateUserDevice(t, "dylan", ksuid.New().String(), nil, "", pdb)

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integrationID,
		Status:        models.UserDeviceAPIIntegrationStatusPendingFirstData,
	}
	err := udai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	testCases := []struct {
		Name                string
		ExistingData        null.JSON
		NewData             null.JSON
		LastOdometerEventAt null.Time
		ExpectedEvent       null.Float64
	}{
		{
			Name:                "New reading, none prior",
			ExistingData:        null.JSON{},
			NewData:             null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			LastOdometerEventAt: null.Time{},
			ExpectedEvent:       null.Float64From(12.5),
		},
		{
			Name:                "Odometer changed, event off cooldown",
			ExistingData:        null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			NewData:             null.JSONFrom([]byte(`{"odometer": 14.5}`)),
			LastOdometerEventAt: null.TimeFrom(time.Now().Add(-2 * odometerCooldown)),
			ExpectedEvent:       null.Float64From(14.5),
		},
		{
			Name:                "Event off cooldown, odometer unchanged",
			ExistingData:        null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			NewData:             null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			LastOdometerEventAt: null.TimeFrom(time.Now().Add(-2 * odometerCooldown)),
			ExpectedEvent:       null.Float64{},
		},
		{
			Name:                "Odometer changed, but event on cooldown",
			ExistingData:        null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			NewData:             null.JSONFrom([]byte(`{"odometer": 14.5}`)),
			LastOdometerEventAt: null.TimeFrom(time.Now().Add(odometerCooldown / 2)),
			ExpectedEvent:       null.Float64{},
		},
	}

	tx := pdb.DBS().Writer

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			defer func() { mes.Buffer = nil }()

			datum := models.UserDeviceDatum{
				UserDeviceID:        ud.ID,
				Data:                c.ExistingData,
				LastOdometerEventAt: c.LastOdometerEventAt,
				IntegrationID:       integrationID,
			}

			err := datum.Upsert(ctx, tx, true, []string{models.UserDeviceDatumColumns.UserDeviceID, models.UserDeviceDatumColumns.IntegrationID},
				boil.Infer(), boil.Infer())
			if err != nil {
				t.Fatalf("Failed setting up existing data row: %v", err)
			}

			input := &DeviceStatusEvent{
				Source:      "dimo/integration/" + integrationID,
				Specversion: "1.0",
				Subject:     ud.ID,
				Type:        deviceStatusEventType,
				Data:        c.NewData.JSON,
			}

			var ctxGk goka.Context
			if err := ingest.processEvent(ctxGk, input); err != nil {
				t.Fatalf("Got an unexpected error processing status update: %v", err)
			}
			if c.ExpectedEvent.Valid {
				if len(mes.Buffer) != 1 {
					t.Fatalf("Expected one odometer event, but got %d", len(mes.Buffer))
				}
				// A bit ugly to have to cast like this.
				actualOdometer := mes.Buffer[0].Data.(OdometerEvent).Odometer
				if actualOdometer != c.ExpectedEvent.Float64 {
					t.Fatalf("Expected an odometer reading of %f but got %f", c.ExpectedEvent.Float64, actualOdometer)
				}
			} else if len(mes.Buffer) != 0 {
				t.Fatalf("Expected no odometer events, but got %d", len(mes.Buffer))
			}
		})
	}
}

func TestAutoPiStatusMerge(t *testing.T) {
	assert := assert.New(t)

	mes := &testEventService{
		Buffer: make([]*Event, 0),
	}
	deviceDefSvc := testDeviceDefSvc{}
	autoPISvc := testAutoPISvc{}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	// Only making use the last parameter.
	ddID := ksuid.New().String()
	integs, _ := deviceDefSvc.GetIntegrations(ctx)
	integrationID := integs[0].Id

	ingest := NewDeviceStatusIngestService(pdb.DBS, &logger, mes, deviceDefSvc, autoPISvc, &config.Settings{
		DIMORegistryChainID: 80001,
		DIMORegistryAddr:    "0x4De1bCf2B7E851E31216fC07989caA902A604784",
		VehicleNFTAddress:   "0x90c4d6113ec88dd4bdf12f26db2b3998fd13a144",
		IssuerPrivateKey:    "3qM17-6NmFbwY52XWtzgbtM1ku2ha3wQ6w6hFpE9o1j",
	})

	ud := test.SetupCreateUserDevice(t, "dylan", ddID, nil, "", pdb)

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integrationID,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
	}

	err := udai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	tx := pdb.DBS().Writer

	dat1 := models.UserDeviceDatum{
		UserDeviceID:        ud.ID,
		Data:                null.JSONFrom([]byte(`{"odometer": 45.22, "latitude": 11.0, "longitude": -7.0}`)),
		LastOdometerEventAt: null.TimeFrom(time.Now().Add(-10 * time.Second)),
		IntegrationID:       integrationID,
	}

	err = dat1.Insert(ctx, tx, boil.Infer())
	assert.NoError(err)

	input := &DeviceStatusEvent{
		Source:      "dimo/integration/" + integrationID,
		Specversion: "1.0",
		Subject:     ud.ID,
		Type:        deviceStatusEventType,
		Time:        time.Now(),
		Data:        []byte(`{"latitude": 2.0, "longitude": 3.0}`),
	}

	var ctxGk goka.Context
	err = ingest.processEvent(ctxGk, input)
	require.NoError(t, err)

	err = dat1.Reload(ctx, tx)
	require.NoError(t, err)

	assert.JSONEq(`{"odometer": 45.22, "latitude": 2.0, "longitude": 3.0}`, string(dat1.Data.JSON))
}

// TestAutoPiStatusWithSignals tests that the signals column is getting updated correctly merging any existing data and setting timestamps
func TestAutoPiStatusWithSignals(t *testing.T) {
	assert := assert.New(t)

	mes := &testEventService{
		Buffer: make([]*Event, 0),
	}
	deviceDefSvc := testDeviceDefSvc{}
	autoPISvc := testAutoPISvc{}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	// Only making use the last parameter.
	ddID := ksuid.New().String()
	integs, _ := deviceDefSvc.GetIntegrations(ctx)
	integrationID := integs[0].Id

	ingest := NewDeviceStatusIngestService(pdb.DBS, &logger, mes, deviceDefSvc, autoPISvc, &config.Settings{
		DIMORegistryChainID: 80001,
		DIMORegistryAddr:    "0x4De1bCf2B7E851E31216fC07989caA902A604784",
		VehicleNFTAddress:   "0x90c4d6113ec88dd4bdf12f26db2b3998fd13a144",
		IssuerPrivateKey:    "3qM17-6NmFbwY52XWtzgbtM1ku2ha3wQ6w6hFpE9o1j",
	})

	ud := test.SetupCreateUserDevice(t, "rvivanco", ddID, nil, "", pdb)

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integrationID,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
	}

	err := udai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	tx := pdb.DBS().Writer

	dat1 := models.UserDeviceDatum{
		UserDeviceID:        ud.ID,
		Data:                null.JSONFrom([]byte(`{"odometer": 45.22, "signal_name_version_1": 23.4}`)),
		Signals:             null.JSONFrom([]byte(`{"signal_name_version_1": {"timestamp": "xx", "value": 23.4}}`)),
		LastOdometerEventAt: null.TimeFrom(time.Now().Add(-10 * time.Second)),
		IntegrationID:       integrationID,
	}

	err = dat1.Insert(ctx, tx, boil.Infer())
	assert.NoError(err)

	input := &DeviceStatusEvent{
		Source:      "dimo/integration/" + integrationID,
		Specversion: "1.0",
		Subject:     ud.ID,
		Type:        deviceStatusEventType,
		Time:        time.Now(),
		Data:        []byte(`{"odometer": 45.22, "signal_name_version_2": 12.3}`),
	}

	var ctxGk goka.Context
	err = ingest.processEvent(ctxGk, input)
	require.NoError(t, err)

	err = dat1.Reload(ctx, tx)
	require.NoError(t, err)

	// validate signals were updated, or not updated, as expected
	assert.Equal("xx", gjson.GetBytes(dat1.Signals.JSON, "signal_name_version_1.timestamp").Str, "signal 1 ts should not change and be present")
	assert.Equal(23.4, gjson.GetBytes(dat1.Signals.JSON, "signal_name_version_1.value").Num, "signal 1 value should not change and be present")
	// assume UTC tz
	assert.Equal(input.Time.Format("2006-01-02T15:04:05Z"), gjson.GetBytes(dat1.Signals.JSON, "odometer.timestamp").Str, "odometer ts should be updated from latest event")
	assert.Equal(45.22, gjson.GetBytes(dat1.Signals.JSON, "odometer.value").Num, "odometer value should be updated from latest event")

	assert.Equal(input.Time.Format("2006-01-02T15:04:05Z"), gjson.GetBytes(dat1.Signals.JSON, "signal_name_version_2.timestamp").Str, "signal 2 ts should be updated from latest event")
	assert.Equal(12.3, gjson.GetBytes(dat1.Signals.JSON, "signal_name_version_2.value").Num, "signal 2 value should be updated from latest event")
}

type testDeviceDefSvc struct {
}

func (t testDeviceDefSvc) GetDeviceDefinitionByID(ctx context.Context, id string) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	dd, err := t.GetDeviceDefinitionsByIDs(ctx, []string{id})
	return dd[0], err
}

// nolint
func (t testDeviceDefSvc) GetIntegrationByTokenID(ctx context.Context, tokenID uint64) (*ddgrpc.Integration, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) PullVincarioValuation(ctx context.Context, userDeiceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) DecodeVIN(ctx context.Context, vin string, model string, year int, code string) (*ddgrpc.DecodeVinResponse, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) FindDeviceDefinitionByMMY(ctx context.Context, mk, model string, year int) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) CheckAndSetImage(ctx context.Context, dd *ddgrpc.GetDeviceDefinitionItemResponse, overwrite bool) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID string, vin string) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) PullDrivlyData(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) (DataPullStatusEnum, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) GetMakeByTokenID(ctx context.Context, tokenID *big.Int) (*ddgrpc.DeviceMake, error) {
	return nil, nil
}

// nolint
func (t testDeviceDefSvc) PullBlackbookData(ctx context.Context, userDeviceID, deviceDefinitionID string, vin string) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*ddgrpc.DeviceMake, error) {
	//TODO implement me
	panic("implement me")
}

var testDeviceDefs []*ddgrpc.GetDeviceDefinitionItemResponse

func (t testDeviceDefSvc) GetDeviceDefinitionsByIDs(_ context.Context, _ []string) ([]*ddgrpc.GetDeviceDefinitionItemResponse, error) {
	if len(testDeviceDefs) > 0 {
		return testDeviceDefs, nil
	}
	d1 := test.BuildDeviceDefinitionGRPC(ksuid.New().String(), "Ford", "escape", 2022, nil)
	testDeviceDefs = d1
	return testDeviceDefs, nil
}

var testIntegs []*ddgrpc.Integration

func (t testDeviceDefSvc) GetIntegrations(_ context.Context) ([]*ddgrpc.Integration, error) {
	if len(testIntegs) > 0 {
		return testIntegs, nil
	}
	i1 := test.BuildIntegrationGRPC(constants.AutoPiVendor, 10, 0)
	testIntegs = []*ddgrpc.Integration{i1}
	return testIntegs, nil
}

// nolint
func (t testDeviceDefSvc) GetIntegrationByID(ctx context.Context, id string) (*ddgrpc.Integration, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) GetIntegrationByVendor(ctx context.Context, vendor string) (*ddgrpc.Integration, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) GetIntegrationByFilter(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testDeviceDefSvc) CreateIntegration(ctx context.Context, integrationType string, vendor string, style string) (*ddgrpc.Integration, error) {
	//TODO implement me
	panic("implement me")
}

type testAutoPISvc struct {
}

// nolint
func (t testAutoPISvc) GetUserDeviceIntegrationByUnitID(ctx context.Context, unitID string) (*models.UserDeviceAPIIntegration, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) GetDeviceByUnitID(unitID string) (*AutoPiDongleDevice, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) GetDeviceByID(deviceID string) (*AutoPiDongleDevice, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) GetDeviceByEthAddress(ethAddress string) (*AutoPiDongleDevice, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) PatchVehicleProfile(vehicleID int, profile PatchVehicleProfile) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) UnassociateDeviceTemplate(deviceID string, templateID int) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) AssociateDeviceToTemplate(deviceID string, templateID int) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) CreateNewTemplate(templateName string, parent int, description string) (int, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) AddDefaultPIDsToTemplate(templateID int) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) SetTemplateICEPowerSettings(templateID int) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) ApplyTemplate(deviceID string, templateID int) error {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) CommandQueryVIN(ctx context.Context, unitID, deviceID, userDeviceID string) (*AutoPiCommandResponse, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) CommandSyncDevice(ctx context.Context, unitID, deviceID, userDeviceID string) (*AutoPiCommandResponse, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) CommandRaw(ctx context.Context, unitID, deviceID, command, userDeviceID string) (*AutoPiCommandResponse, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) GetCommandStatus(ctx context.Context, jobID string) (*AutoPiCommandJob, *models.AutopiJob, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) GetCommandStatusFromAutoPi(deviceID string, jobID string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) UpdateJob(ctx context.Context, jobID, newState string, result *AutoPiCommandResult) (*models.AutopiJob, error) {
	//TODO implement me
	panic("implement me")
}

// nolint
func (t testAutoPISvc) UpdateState(deviceID string, state string) error {
	return nil
}
