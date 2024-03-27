package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"

	"github.com/DIMO-Network/devices-api/internal/config"
)

//go:generate mockgen -source tesla_fleet_api_service.go -destination mocks/tesla_fleet_api_service_mock.go
type TeslaFleetAPIService interface {
	CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error)
	GetVehicles(ctx context.Context, token, region string) ([]TeslaVehicle, error)
	GetVehicle(ctx context.Context, token, region string, vehicleID int) (*TeslaVehicle, error)
	WakeUpVehicle(ctx context.Context, token, region string, vehicleID int) error
	RegisterToTelemetryServer(ctx context.Context, token, region, vin string) error
}

type Interval struct {
	IntervalSeconds int `json:"interval_seconds"`
}

type TelemetryConfigRequest struct {
	HostName            string              `json:"hostName"`
	PublicCACertificate string              `json:"ca"`
	Fields              map[string]Interval `json:"fields"`
	AlertTypes          []string            `json:"alert_types,omitempty"`
	Expiration          int64               `json:"exp"`
	Port                int                 `json:"port"`
}

type RegisterVehicleToTelemetryServerRequest struct {
	Vins   []string               `json:"vins"`
	Config TelemetryConfigRequest `json:"config"`
}

var teslaScopes = []string{"openid", "offline_access", "user_data", "vehicle_device_data", "vehicle_cmds", "vehicle_charging_cmds", "energy_device_data", "energy_device_data", "energy_cmds"}

type GetVehiclesResponse struct {
	Response []TeslaVehicle `json:"response"`
}

type GetSingleVehicleItemResponse struct {
	Response TeslaVehicle `json:"response"`
}

type TeslaFleetAPIError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ReferenceID      string `json:"referenceId"`
}

type TeslaAuthCodeResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type"`
	Region       string    `json:"region"`
}

type teslaFleetAPIService struct {
	Settings   *config.Settings
	HTTPClient *http.Client
	log        *zerolog.Logger
}

func NewTeslaFleetAPIService(settings *config.Settings, logger *zerolog.Logger) TeslaFleetAPIService {
	return &teslaFleetAPIService{
		Settings:   settings,
		HTTPClient: &http.Client{},
		log:        logger,
	}
}

// CompleteTeslaAuthCodeExchange calls Tesla Fleet API and exchange auth code for a new auth and refresh token
func (t *teslaFleetAPIService) CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error) {
	conf := oauth2.Config{
		ClientID:     t.Settings.TeslaClientID,
		ClientSecret: t.Settings.TeslaClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: t.Settings.TeslaTokenURL,
		},
		RedirectURL: redirectURI,
		Scopes:      teslaScopes,
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	tok, err := conf.Exchange(ctxTimeout, authCode, oauth2.SetAuthURLParam("audience", fmt.Sprintf(t.Settings.TeslaFleetURL, region)))
	if err != nil {
		var e *oauth2.RetrieveError
		errString := err.Error()
		if errors.As(err, &e) {
			errString = e.ErrorDescription
		}
		return nil, fmt.Errorf("error occurred completing authorization: %s", errString)
	}

	return &TeslaAuthCodeResponse{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
		TokenType:    tok.TokenType,
	}, nil
}

// GetVehicles calls Tesla Fleet API to get a list of vehicles using authorization token
func (t *teslaFleetAPIService) GetVehicles(ctx context.Context, token, region string) ([]TeslaVehicle, error) {
	baseURL := fmt.Sprintf(t.Settings.TeslaFleetURL, region)
	url := baseURL + "/api/1/vehicles"

	resp, err := t.performTeslaGetRequest(ctx, url, token)
	if err != nil {
		return nil, fmt.Errorf("could not fetch vehicles for user: %w", err)
	}
	defer resp.Body.Close()

	vehicles := new(GetVehiclesResponse)
	if err := json.NewDecoder(resp.Body).Decode(vehicles); err != nil {
		return nil, fmt.Errorf("invalid response encountered while fetching user vehicles: %w", err)
	}

	if vehicles.Response == nil {
		return nil, fmt.Errorf("error occurred fetching user vehicles")
	}

	return vehicles.Response, nil
}

// GetVehicle calls Tesla Fleet API to get a single vehicle by ID
func (t *teslaFleetAPIService) GetVehicle(ctx context.Context, token, region string, vehicleID int) (*TeslaVehicle, error) {
	baseURL := fmt.Sprintf(t.Settings.TeslaFleetURL, region)
	url := fmt.Sprintf("%s/api/1/vehicles/%d", baseURL, vehicleID)

	resp, err := t.performTeslaGetRequest(ctx, url, token)
	if err != nil {
		return nil, fmt.Errorf("could not fetch vehicles for user: %w", err)
	}
	defer resp.Body.Close()

	vehicle := new(GetSingleVehicleItemResponse)
	if err := json.NewDecoder(resp.Body).Decode(vehicle); err != nil {
		return nil, fmt.Errorf("invalid response encountered while fetching user vehicles: %w", err)
	}

	return &vehicle.Response, nil
}

// WakeUpVehicle Calls Tesla Fleet API to wake a vehicle from sleep
func (t *teslaFleetAPIService) WakeUpVehicle(ctx context.Context, token, region string, vehicleID int) error {
	baseURL := fmt.Sprintf(t.Settings.TeslaFleetURL, region)
	url := fmt.Sprintf("%s/api/1/vehicles/%d/wake_up", baseURL, vehicleID)

	resp, err := t.performTeslaGetRequest(ctx, url, token)
	if err != nil {
		return fmt.Errorf("could not fetch vehicles for user: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status code %d waking up vehicle %d", resp.StatusCode, vehicleID)
	}

	return nil
}

func (t *teslaFleetAPIService) RegisterToTelemetryServer(ctx context.Context, token, region, vin string) error {
	baseURL := fmt.Sprintf(t.Settings.TeslaFleetURL, region)
	u, err := url.Parse(fmt.Sprintf("%s/api/1/vehicles/fleet_telemetry_config", baseURL))
	if err != nil {
		return err
	}

	exp := time.Now().AddDate(0, 0, 364).Unix()

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	r := RegisterVehicleToTelemetryServerRequest{
		Vins: []string{vin},
		Config: TelemetryConfigRequest{
			HostName:            t.Settings.TeslaTelemetryHostName,
			PublicCACertificate: t.Settings.TeslaTelemetryCACertificate,
			Expiration:          exp,
			Port:                t.Settings.TeslaTelemetryPort,
			Fields:              make(map[string]Interval),
			AlertTypes:          []string{"service"},
		},
	}
	for _, v := range fields {
		r.Config.Fields[v] = Interval{
			IntervalSeconds: 1800,
		}
	}

	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errBody := new(TeslaFleetAPIError)
		if err := json.NewDecoder(resp.Body).Decode(errBody); err != nil {
			t.log.
				Err(err).
				Str("url", u.String()).
				Msg("error occurred registering vehicle to telemetry server.")
			return fmt.Errorf("invalid response encountered while registring vehicle to telemetry server: %s", errBody.ErrorDescription)
		}
		return fmt.Errorf("error occurred registering vehicle to telemetry server.: %s", errBody.ErrorDescription)
	}

	return nil
}

// performTeslaGetRequest a helper function for making http requests, it adds a timeout context and parses error response
func (t *teslaFleetAPIService) performTeslaGetRequest(ctx context.Context, url, token string) (*http.Response, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error occurred calling tesla fleet api: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errBody := new(TeslaFleetAPIError)
		if err := json.NewDecoder(resp.Body).Decode(errBody); err != nil {
			t.log.
				Err(err).
				Str("url", url).
				Msg("An error occurred when attempting to decode the error message from the api.")
			return nil, fmt.Errorf("invalid response encountered while fetching user vehicles: %s", errBody.ErrorDescription)
		}
		return nil, fmt.Errorf("error occurred fetching user vehicles: %s", errBody.ErrorDescription)
	}

	return resp, nil
}

var fields = []string{
	"Unknown",
	"DriveRail",
	"ChargeState",
	"BmsFullchargecomplete",
	"VehicleSpeed",
	"Odometer",
	"PackVoltage",
	"PackCurrent",
	"Soc",
	"DCDCEnable",
	"Gear",
	"IsolationResistance",
	"PedalPosition",
	"BrakePedal",
	"DiStateR",
	"DiHeatsinkTR",
	"DiAxleSpeedR",
	"DiTorquemotor",
	"DiStatorTempR",
	"DiVBatR",
	"DiMotorCurrentR",
	"Location",
	"GpsState",
	"GpsHeading",
	"NumBrickVoltageMax",
	"BrickVoltageMax",
	"NumBrickVoltageMin",
	"BrickVoltageMin",
	"NumModuleTempMax",
	"ModuleTempMax",
	"NumModuleTempMin",
	"ModuleTempMin",
	"RatedRange",
	"Hvil",
	"DCChargingEnergyIn",
	"DCChargingPower",
	"ACChargingEnergyIn",
	"ACChargingPower",
	"ChargeLimitSoc",
	"FastChargerPresent",
	"EstBatteryRange",
	"IdealBatteryRange",
	"BatteryLevel",
	"TimeToFullCharge",
	"ScheduledChargingStartTime",
	"ScheduledChargingPending",
	"ScheduledDepartureTime",
	"PreconditioningEnabled",
	"ScheduledChargingMode",
	"ChargeAmps",
	"ChargeEnableRequest",
	"ChargerPhases",
	"ChargePortColdWeatherMode",
	"ChargeCurrentRequest",
	"ChargeCurrentRequestMax",
	"BatteryHeaterOn",
	"NotEnoughPowerToHeat",
	"SuperchargerSessionTripPlanner",
	"DoorState",
	"Locked",
	"FdWindow",
	"FpWindow",
	"RdWindow",
	"RpWindow",
	"VehicleName",
	"SentryMode",
	"SpeedLimitMode",
	"CurrentLimitMph",
	"Version",
	"TpmsPressureFl",
	"TpmsPressureFr",
	"TpmsPressureRl",
	"TpmsPressureRr",
	"SemitruckTpmsPressureRe1L0",
	"SemitruckTpmsPressureRe1L1",
	"SemitruckTpmsPressureRe1R0",
	"SemitruckTpmsPressureRe1R1",
	"SemitruckTpmsPressureRe2L0",
	"SemitruckTpmsPressureRe2L1",
	"SemitruckTpmsPressureRe2R0",
	"SemitruckTpmsPressureRe2R1",
	"TpmsLastSeenPressureTimeFl",
	"TpmsLastSeenPressureTimeFr",
	"TpmsLastSeenPressureTimeRl",
	"TpmsLastSeenPressureTimeRr",
	"InsideTemp",
	"OutsideTemp",
	"SeatHeaterLeft",
	"SeatHeaterRight",
	"SeatHeaterRearLeft",
	"SeatHeaterRearRight",
	"SeatHeaterRearCenter",
	"AutoSeatClimateLeft",
	"AutoSeatClimateRight",
	"DriverSeatBelt",
	"PassengerSeatBelt",
	"DriverSeatOccupied",
	"SemitruckPassengerSeatFoldPosition",
	"LateralAcceleration",
	"LongitudinalAcceleration",
	"CruiseState",
	"CruiseSetSpeed",
	"LifetimeEnergyUsed",
	"LifetimeEnergyUsedDrive",
	"SemitruckTractorParkBrakeStatus",
	"SemitruckTrailerParkBrakeStatus",
	"BrakePedalPos",
	"RouteLastUpdated",
	"RouteLine",
	"MilesToArrival",
	"MinutesToArrival",
	"OriginLocation",
	"DestinationLocation",
	"CarType",
	"Trim",
	"ExteriorColor",
	"RoofColor",
	"ChargePort",
	"ChargePortLatch",
	"Experimental_1",
	"Experimental_2",
	"Experimental_3",
	"Experimental_4",
	"GuestModeEnabled",
	"PinToDriveEnabled",
	"PairedPhoneKeyAndKeyFobQty",
	"CruiseFollowDistance",
	"AutomaticBlindSpotCamera",
	"BlindSpotCollisionWarningChime",
	"SpeedLimitWarning",
	"ForwardCollisionWarning",
	"LaneDepartureAvoidance",
	"EmergencyLaneDepartureAvoidance",
	"AutomaticEmergencyBrakingOff",
	"LifetimeEnergyGainedRegen",
	"DiStateF",
	"DiStateREL",
	"DiStateRER",
	"DiHeatsinkTF",
	"DiHeatsinkTREL",
	"DiHeatsinkTRER",
	"DiAxleSpeedF",
	"DiAxleSpeedREL",
	"DiAxleSpeedRER",
	"DiSlaveTorqueCmd",
	"DiTorqueActualR",
	"DiTorqueActualF",
	"DiTorqueActualREL",
	"DiTorqueActualRER",
	"DiStatorTempF",
	"DiStatorTempREL",
	"DiStatorTempRER",
	"DiVBatF",
	"DiVBatREL",
	"DiVBatRER",
	"DiMotorCurrentF",
	"DiMotorCurrentREL",
	"DiMotorCurrentRER",
	"EnergyRemaining",
	"ServiceMode",
	"BMSState",
	"GuestModeMobileAccessState",
	"Deprecated_1",
	"DestinationName",
}
