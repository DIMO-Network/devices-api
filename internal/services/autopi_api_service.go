package services

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/tidwall/gjson"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

//go:generate mockgen -source autopi_api_service.go -destination mocks/autopi_api_service_mock.go
type AutoPiAPIService interface {
	GetUserDeviceIntegrationByUnitID(ctx context.Context, unitID string) (*models.UserDeviceAPIIntegration, error)
	GetDeviceByUnitID(unitID string) (*AutoPiDongleDevice, error)
	GetDeviceByID(deviceID string) (*AutoPiDongleDevice, error)
	PatchVehicleProfile(vehicleID int, profile PatchVehicleProfile) error
	UnassociateDeviceTemplate(deviceID string, templateID int) error
	AssociateDeviceToTemplate(deviceID string, templateID int) error
	CreateNewTemplate(templateName string, parent int, description string) (int, error)
	SetTemplateICEPowerSettings(templateID int) error
	ApplyTemplate(deviceID string, templateID int) error
	CommandQueryVIN(ctx context.Context, unitID, deviceID, userDeviceID string) (*AutoPiCommandResponse, error)
	CommandSyncDevice(ctx context.Context, unitID, deviceID, userDeviceID string) (*AutoPiCommandResponse, error)
	CommandRaw(ctx context.Context, unitID, deviceID, command, userDeviceID string) (*AutoPiCommandResponse, error)
	GetCommandStatus(ctx context.Context, jobID string) (*AutoPiCommandJob, *models.AutopiJob, error)
	GetCommandStatusFromAutoPi(deviceID string, jobID string) ([]byte, error)
	UpdateJob(ctx context.Context, jobID, newState string, result *AutoPiCommandResult) (*models.AutopiJob, error)
}

type autoPiAPIService struct {
	Settings   *config.Settings
	httpClient shared.HTTPClientWrapper
	dbs        func() *db.ReaderWriter
}

var ErrNotFound = errors.New("not found")

func NewAutoPiAPIService(settings *config.Settings, dbs func() *db.ReaderWriter) AutoPiAPIService {
	h := map[string]string{"Authorization": "APIToken " + settings.AutoPiAPIToken}
	hcw, _ := shared.NewHTTPClientWrapper(settings.AutoPiAPIURL, "", 10*time.Second, h, true) // ok to ignore err since only used for tor check

	return &autoPiAPIService{
		Settings:   settings,
		httpClient: hcw,
		dbs:        dbs,
	}
}

func (a *autoPiAPIService) GetUserDeviceIntegrationByUnitID(ctx context.Context, unitID string) (*models.UserDeviceAPIIntegration, error) {
	udai, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.AutopiUnitID.EQ(null.StringFrom(unitID)),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice)).
		One(ctx, a.dbs().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return udai, nil
}

// GetDeviceByUnitID calls /dongle/devices/by_unit_id/{unit_id}/ to get the device for the unitID.
// Errors if it finds none or more than one device, as there should only be one device attached to a unit.
func (a *autoPiAPIService) GetDeviceByUnitID(unitID string) (*AutoPiDongleDevice, error) {
	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/dongle/devices/by_unit_id/%s/", unitID), "GET", nil)
	if err != nil {
		if _, ok := err.(shared.HTTPResponseError); !ok {
			return nil, errors.Wrapf(err, "error calling autopi api to get unit with ID %s", unitID)
		}
	}
	defer res.Body.Close() // nolint
	if res.StatusCode == 404 {
		return nil, ErrNotFound
	}

	u := new(AutoPiDongleDevice)
	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding json from autopi api to get device by unitID %s", unitID)
	}

	return u, nil
}

// GetDeviceByID calls https://api.dimo.autopi.io/dongle/devices/{DEVICE_ID}/ Note that the deviceID is the autoPi one. This brings us the templateID
func (a *autoPiAPIService) GetDeviceByID(deviceID string) (*AutoPiDongleDevice, error) {
	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/dongle/devices/%s/", deviceID), "GET", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling autopi api to get device %s", deviceID)
	}
	defer res.Body.Close() // nolint

	d := new(AutoPiDongleDevice)
	err = json.NewDecoder(res.Body).Decode(d)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding json from autopi api to get device %s", deviceID)
	}
	return d, nil
}

// PatchVehicleProfile https://api.dimo.autopi.io/vehicle/profile/{device.vehicle.id}/ driveType: {"ICE", "BEV", "PHEV", "HEV"}
func (a *autoPiAPIService) PatchVehicleProfile(vehicleID int, profile PatchVehicleProfile) error {
	j, _ := json.Marshal(profile)
	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/vehicle/profile/%d/", vehicleID), "PATCH", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to patch vehicle profile for vehicleID %d", vehicleID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// UnassociateDeviceTemplate Unassociate the device from the existing templateID.
func (a *autoPiAPIService) UnassociateDeviceTemplate(deviceID string, templateID int) error {
	p := postDeviceIDs{
		Devices:         []string{deviceID},
		UnassociateOnly: false,
	}
	j, _ := json.Marshal(p)
	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/dongle/templates/%d/unassociate_devices/", templateID), "POST", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to unassociate_devices. template %d", templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// AssociateDeviceToTemplate set a new templateID on the device by doing a Patch request
func (a *autoPiAPIService) AssociateDeviceToTemplate(deviceID string, templateID int) error {
	p := postDeviceIDs{
		Devices: []string{deviceID},
	}
	j, _ := json.Marshal(p)
	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/dongle/templates/%d/", templateID), "PATCH", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to associate device %s with new template %d", deviceID, templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// CreateNewTemplate create a new template on the AutoPi cloud by doing a put request.
//
//	Parent is optional(setting to 0 creates template with no parent)
func (a *autoPiAPIService) CreateNewTemplate(templateName string, parent int, description string) (int, error) {
	p := postNewTemplateRequest{
		TemplateName: templateName,
		Description:  description,
		Devices:      []string{}, // must not be null, but not required to have entries
	}
	if parent > 0 {
		p.Parent = parent
	}
	j, _ := json.Marshal(p)
	res, err := a.httpClient.ExecuteRequest("/dongle/templates/", "POST", j)
	if err != nil {
		return 0, errors.Wrapf(err, "error calling autopi api to create new template")
	}

	respBytes, _ := io.ReadAll(res.Body)
	idResult := gjson.GetBytes(respBytes, "id")
	defer res.Body.Close() // nolint
	if idResult.Exists() && idResult.Int() > 0 {
		return int(idResult.Int()), nil
	}
	return 0, errors.New("did not find a template id in the response: " + string(respBytes))
}

//go:embed generic_ice_power_settings.json
var powerSettings string

func (a *autoPiAPIService) SetTemplateICEPowerSettings(templateID int) error {
	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/dongle/settings/?template_id=%d", templateID), "POST", []byte(powerSettings))
	if err != nil {
		println(res.Body)
		return errors.Wrapf(err, "Could not apply power settings to template %d", templateID)
	}
	return nil
}

// AddDefaultPIDsToTemplate adds all default PIDs to template for 2008+, does not add odometer PID
func (a *autoPiAPIService) AddDefaultPIDsToTemplate(templateID int) error {
	pids := make([]*addPIDRequest, 11)
	pids[0] = newAddPIDReqWithDefaults(templateID, 32, 30)    // runtime
	pids[1] = newAddPIDReqWithDefaults(templateID, 52, 5)     //bar pressure
	pids[2] = newAddPIDReqWithDefaults(templateID, 18, 5)     // throttle pos
	pids[3] = newAddPIDReqWithDefaults(templateID, 14, 5)     //speed
	pids[4] = newAddPIDReqWithDefaults(templateID, 5, 5)      // engine load
	pids[5] = newAddPIDReqWithDefaults(templateID, 71, 5)     // ambiente air
	pids[6] = newAddPIDReqWithDefaults(templateID, 16, 5)     // intake temp
	pids[7] = newAddPIDReqWithDefaults(templateID, 48, 5)     // fuel level
	pids[8] = newAddPIDReqWithDefaults(templateID, 6, 5)      // coolant temp
	pids[9] = newAddPIDReqWithDefaults(templateID, 2900, 600) // vin 2008+
	pids[10] = newAddPIDReqWithDefaults(templateID, 13, 3)    // rpm
	pids[10].Trigger = []string{"rpm_engine_event"}

	for _, pid := range pids {
		req, _ := json.Marshal(pid)
		res, err := a.httpClient.ExecuteRequest("/obd/loggers/pid/", "POST", req)
		if err != nil {
			println(res.Body)
			return errors.Wrapf(err, "Could not add PID %d to template %d", pid.Pid, templateID)
		}
	}
	return nil
}

// ApplyTemplate When device awakes, it checks if it has templates to be applied. If device is awake, this won't do anything until next cycle.
func (a *autoPiAPIService) ApplyTemplate(deviceID string, templateID int) error {
	p := postDeviceIDs{
		Devices: []string{deviceID},
	}
	j, _ := json.Marshal(p)
	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/dongle/templates/%d/apply_explicit/", templateID), "POST", j)
	if err != nil {
		return errors.Wrapf(err, "error calling autopi api to apply template for device %s with new template %d", deviceID, templateID)
	}
	defer res.Body.Close() // nolint

	return nil
}

// CommandQueryVIN sends raw command to autopi to get the vin in the webhook response after. only works if device is online.
func (a *autoPiAPIService) CommandQueryVIN(ctx context.Context, unitID, deviceID, userDeviceID string) (*AutoPiCommandResponse, error) {
	return a.CommandRaw(ctx, unitID, deviceID,
		"obd.query vin mode=09 pid=02 header=7DF bytes=20 formula='messages[0].data[3:].decode(\"ascii\")' baudrate=500000 protocol=auto verify=false force=true",
		userDeviceID)
}

// CommandSyncDevice sends raw command to autopi only if it is online. Invokes syncing the pending changes (eg. template change) on the device.
func (a *autoPiAPIService) CommandSyncDevice(ctx context.Context, unitID, deviceID, userDeviceID string) (*AutoPiCommandResponse, error) {
	return a.CommandRaw(ctx, unitID, deviceID, "state.sls pending", userDeviceID)
}

// CommandRaw sends raw command to autopi and saves in autopi_jobs. If device is offline command will eventually timeout.
func (a *autoPiAPIService) CommandRaw(ctx context.Context, unitID, deviceID, command, userDeviceID string) (*AutoPiCommandResponse, error) {
	// todo: whitelist command
	v, unitID := ValidateAndCleanUUID(unitID)
	if !v {
		return nil, errors.New("send command failed, invalid unitId: " + unitID)
	}
	webhookURL := fmt.Sprintf("%s/v1%s", a.Settings.DeploymentBaseURL, constants.AutoPiWebhookPath)
	syncCommand := autoPiCommandRequest{
		Command:     command,
		CallbackURL: &webhookURL,
	}

	j, err := json.Marshal(syncCommand)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshall json for autoPiCommandRequest")
	}

	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/dongle/devices/%s/execute_raw/", deviceID), "POST", j)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling autopi api execute_raw command %s for deviceId %s", command, deviceID)
	}
	defer res.Body.Close() // nolint

	d := new(AutoPiCommandResponse)
	err = json.NewDecoder(res.Body).Decode(d)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to decode responde from autopi api execute_raw command")
	}

	// insert job
	autoPiJob := models.AutopiJob{
		ID:             d.Jid,
		Command:        command,
		AutopiDeviceID: deviceID,
		AutopiUnitID:   null.StringFrom(unitID),
	}
	if len(userDeviceID) > 0 {
		autoPiJob.UserDeviceID = null.StringFrom(userDeviceID)
	}
	err = autoPiJob.Insert(ctx, a.dbs().Writer, boil.Infer())
	if err != nil {
		return nil, err
	}

	return d, nil
}

// UpdateJob updates the state of a autopi job on our end
func (a *autoPiAPIService) UpdateJob(ctx context.Context, jobID, newState string, result *AutoPiCommandResult) (*models.AutopiJob, error) {
	autopiJob, err := models.AutopiJobs(models.AutopiJobWhere.ID.EQ(jobID)).One(ctx, a.dbs().Reader)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding autopi job")
	}
	// update the job state
	autopiJob.State = newState
	autopiJob.CommandLastUpdated = null.TimeFrom(time.Now().UTC())
	if result != nil {
		err = autopiJob.CommandResult.Marshal(result)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal command result to save in db %+v", result)
		}
	}

	_, err = autopiJob.Update(ctx, a.dbs().Writer, boil.Infer())
	if err != nil {
		return nil, errors.Wrapf(err, "error updating autopi job")
	}
	return autopiJob, nil
}

// GetCommandStatusFromAutoPi gets the status of a previously sent command by calling autopi. returns raw body since it can change depending on command
func (a *autoPiAPIService) GetCommandStatusFromAutoPi(deviceID string, jobID string) ([]byte, error) {
	res, err := a.httpClient.ExecuteRequest(fmt.Sprintf("/dongle/devices/%s/command_result/%s/", deviceID, jobID), "GET", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling autopi api to get command status for deviceId %s", deviceID)
	}
	defer res.Body.Close() // nolint
	body, _ := io.ReadAll(res.Body)

	return body, nil
}

// GetCommandStatus gets job state from our database, which is updated by autopi webhooks.
func (a *autoPiAPIService) GetCommandStatus(ctx context.Context, jobID string) (*AutoPiCommandJob, *models.AutopiJob, error) {
	autoPiJob, err := models.AutopiJobs(models.AutopiJobWhere.ID.EQ(jobID)).One(ctx, a.dbs().Reader)
	if err != nil {
		return nil, nil, err
	}

	job := &AutoPiCommandJob{
		CommandJobID: autoPiJob.ID,
		CommandState: autoPiJob.State,
		CommandRaw:   autoPiJob.Command,
		LastUpdated:  autoPiJob.CommandLastUpdated.Ptr(),
	}
	if autoPiJob.CommandResult.Valid {
		err = autoPiJob.CommandResult.Unmarshal(job.Result)
		return nil, nil, err
	}
	return job, autoPiJob, nil
}

// AutoPiDongleDevice https://api.dimo.autopi.io/#/dongle/dongle_devices_read
type AutoPiDongleDevice struct {
	ID                string              `json:"id"`
	UnitID            string              `json:"unit_id"`
	Token             string              `json:"token"`
	CallName          string              `json:"callName"`
	Owner             int                 `json:"owner"`
	Vehicle           AutoPiDongleVehicle `json:"vehicle"`
	Display           string              `json:"display"`
	LastCommunication time.Time           `json:"last_communication"`
	IsUpdated         bool                `json:"is_updated"`
	EthereumAddress   string              `json:"ethereum_address"`
	Release           struct {
		Version string `json:"version"`
	} `json:"release"`
	OpenAlerts struct {
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Critical int `json:"critical"`
		Low      int `json:"low"`
	} `json:"open_alerts"`
	IMEI     string `json:"imei"`
	Template int    `json:"template"`
	Warnings []struct {
		DeviceHasNoMakeModel struct {
			Header  string `json:"header"`
			Message string `json:"message"`
		} `json:"device_has_no_make_model"`
	} `json:"warnings"`
	KeyState           string `json:"key_state"`
	Access             string `json:"access"`
	DockerReleases     []int  `json:"docker_releases"`
	DataUsage          int    `json:"data_usage"`
	PhoneNumber        string `json:"phone_number"`
	Icc                string `json:"icc"`
	MaxDataUsage       int    `json:"max_data_usage"`
	IsBlockedByRelease bool   `json:"is_blocked_by_release"`
	// only exists when get by unitID
	HwRevision string   `json:"hw_revision"`
	Tags       []string `json:"tags"`
}

type AutoPiDongleVehicle struct {
	ID                    int    `json:"id"`
	Vin                   string `json:"vin"`
	Display               string `json:"display"`
	CallName              string `json:"callName"`
	LicensePlate          string `json:"licensePlate"`
	Model                 int    `json:"model"`
	Make                  int    `json:"make"`
	Year                  int    `json:"year"`
	Type                  string `json:"type"`
	BatteryNominalVoltage int    `json:"battery_nominal_voltage"`
}

// PatchVehicleProfile used to update vehicle profile https://api.dimo.autopi.io/#/vehicle/vehicle_profile_partial_update
type PatchVehicleProfile struct {
	Vin      string `json:"vin,omitempty"`
	CallName string `json:"callName,omitempty"`
	Year     int    `json:"year,omitempty"`
	Type     string `json:"type,omitempty"`
}

// used to post an array of device ID's, for template and command operations
type postDeviceIDs struct {
	Devices         []string `json:"devices"`
	UnassociateOnly bool     `json:"unassociate_only,omitempty"`
}

// used to create a new AutoPi template on the cloud
type postNewTemplateRequest struct {
	TemplateName string   `json:"name"`
	Parent       int      `json:"parent,omitempty"`
	Description  string   `json:"description"`
	Devices      []string `json:"devices"`
}

type autoPiCommandRequest struct {
	Command     string  `json:"command"`
	CallbackURL *string `json:"callback_url,omitempty"`
	// CallbackTimeout default is 120 seconds
	CallbackTimeout *int `json:"callback_timeout,omitempty"`
}

type AutoPiCommandResponse struct {
	Jid     string   `json:"jid"`
	Minions []string `json:"minions"`
}

type AutoPiCommandResult struct {
	// corresponds to webhook response.data.return._type
	Type string `json:"type"`
	// corresponds to webhook response.data.return.value
	Value string `json:"value"`
	// corresponds to webhook response.tag
	Tag string `json:"tag"`
}

type addPIDRequest struct {
	ID        int      `json:"id"`
	Converter string   `json:"converter"`
	Trigger   []string `json:"trigger"`
	Filter    string   `json:"filter"`
	Returner  []string `json:"returner"`
	Interval  int      `json:"interval"`
	Pid       int      `json:"pid"`
	Verify    bool     `json:"verify"`
	Enabled   bool     `json:"enabled"`
	Template  int      `json:"template"`
}

func newAddPIDReqWithDefaults(templateID, pid, interval int) *addPIDRequest {
	req := addPIDRequest{
		ID:        0,
		Converter: "",
		Trigger:   nil,
		Filter:    "alternating_readout",
		Returner:  []string{"context_returner_data"},
		Interval:  interval,
		Pid:       pid,
		Verify:    false,
		Enabled:   true,
		Template:  templateID,
	}
	return &req
}
