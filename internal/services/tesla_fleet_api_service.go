package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type TeslaVehicle struct {
	ID        int    `json:"id"`
	VehicleID int    `json:"vehicle_id"`
	VIN       string `json:"vin"`
}

//go:generate mockgen -source tesla_fleet_api_service.go -destination mocks/tesla_fleet_api_service_mock.go
type TeslaFleetAPIService interface {
	CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI string) (*TeslaAuthCodeResponse, error)
	GetVehicles(ctx context.Context, token string) ([]TeslaVehicle, error)
	GetVehicle(ctx context.Context, token string, vehicleID int) (*TeslaVehicle, error)
	WakeUpVehicle(ctx context.Context, token string, vehicleID int) error
	GetAvailableCommands(token string) (*UserDeviceAPIIntegrationsMetadataCommands, error)
	VirtualKeyConnectionStatus(ctx context.Context, token, vin string) (*VehicleFleetStatus, error)
	SubscribeForTelemetryData(ctx context.Context, token, vin string) error
	GetTelemetrySubscriptionStatus(ctx context.Context, token, vin string) (*VehicleTelemetryStatus, error)
}

var teslaScopes = []string{"openid", "offline_access", "user_data", "vehicle_device_data", "vehicle_cmds", "vehicle_charging_cmds"}

type TeslaResponseWrapper[A any] struct {
	Response   A `json:"response"`
	Pagination struct {
		Next int `json:"next"`
	} `json:"pagination"`
}

// ErrWrongRegion is returned when the Tesla proxy chooses the wrong region for a request.
var ErrWrongRegion = errors.New("tesla: incorrect region")

type TeslaFleetAPIError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ReferenceID      string `json:"referenceId"`
}

type TeslaAuthCodeResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type"`
	Region       string    `json:"region"`
}

type fleetStatusResponse struct {
	KeyPairedVINs []string `json:"key_paired_vins"`
	UnpairedVINs  []string `json:"unpaired_vins"`
	VehicleInfo   map[string]struct {
		FirmwareVersion                    string `json:"firmware_version"`
		VehicleCommandProtocolRequired     bool   `json:"vehicle_command_protocol_required"`
		DiscountedDeviceData               bool   `json:"discounted_device_data"`
		FleetTelemetryVersion              string `json:"fleet_telemetry_version"`
		TotalNumberOfKeys                  int    `json:"total_number_of_keys"`
		SafetyScreenStreamingToggleEnabled *bool  `json:"safety_screen_streaming_toggle_enabled"`
	} `json:"vehicle_info"`
}

type VehicleFleetStatus struct {
	KeyPaired                          bool
	VehicleCommandProtocolRequired     bool
	FirmwareVersion                    string
	DiscountedDeviceData               bool
	FleetTelemetryVersion              string
	NumberOfKeys                       int
	SafetyScreenStreamingToggleEnabled *bool
}

type VehicleTelemetryStatus struct {
	Synced       bool
	Configured   bool
	LimitReached bool
	KeyPaired    bool
}

type SubscribeForTelemetryDataRequest struct {
	VINs   []string               `json:"vins"`
	Config TelemetryConfigRequest `json:"config"`
}

type Interval struct {
	IntervalSeconds int `json:"interval_seconds"`
}

type TelemetryFields map[string]Interval

type TelemetryConfigRequest struct {
	HostName    string          `json:"hostname"`
	CA          string          `json:"ca"`
	Fields      TelemetryFields `json:"fields"`
	Port        int             `json:"port"`
	PreferTyped bool            `json:"prefer_typed"`
}

type TelemetryConfigStatusResponse struct {
	Synced       bool                    `json:"synced"`
	Config       *TelemetryConfigRequest `json:"config"`
	LimitReached bool                    `json:"limit_reached"`
	KeyPaired    bool                    `json:"key_paired"`
}

type SkippedVehicles struct {
	MissingKey          []string `json:"missing_key"`
	UnsupportedHardware []string `json:"unsupported_hardware"`
	UnsupportedFirmware []string `json:"unsupported_firmware"`
	MaxConfigs          []string `json:"max_configs"`
}

type SubscribeForTelemetryDataResponse struct {
	UpdatedVehicles int             `json:"updated_vehicles"`
	SkippedVehicles SkippedVehicles `json:"skipped_vehicles"`
}

type teslaFleetAPIService struct {
	Settings   *config.Settings
	HTTPClient *http.Client
	log        *zerolog.Logger
	FleetBase  *url.URL
}

func NewTeslaFleetAPIService(settings *config.Settings, logger *zerolog.Logger) (TeslaFleetAPIService, error) {
	u, err := url.ParseRequestURI(settings.TeslaFleetURL)
	if err != nil {
		return nil, err
	}

	return &teslaFleetAPIService{
		Settings:   settings,
		HTTPClient: &http.Client{},
		log:        logger,
		FleetBase:  u,
	}, nil
}

var ErrInvalidAuthCode = errors.New("authorization code invalid, expired, or revoked")

// CompleteTeslaAuthCodeExchange calls Tesla Fleet API and exchange auth code for a new auth and refresh token
func (t *teslaFleetAPIService) CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI string) (*TeslaAuthCodeResponse, error) {
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

	// TODO(elffjs): Tesla says on their site that audience is required, but the token always has
	// both na and eu audiences and omitting the parameter results in no errors.
	tok, err := conf.Exchange(ctxTimeout, authCode)
	if err != nil {
		var e *oauth2.RetrieveError
		errString := err.Error()
		if errors.As(err, &e) {
			// Non-standard error code from Tesla. See RFC 6749.
			if e.ErrorCode == "invalid_auth_code" {
				return nil, ErrInvalidAuthCode
			}
			t.log.Info().Str("error", e.ErrorCode).Str("errorDescription", e.ErrorDescription).Msg("Code exchange failure.")
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
func (t *teslaFleetAPIService) GetVehicles(ctx context.Context, token string) ([]TeslaVehicle, error) {
	out := make([]TeslaVehicle, 0)
	page := 1

	listStart := time.Now()
	for {
		url := t.FleetBase.JoinPath("api/1/vehicles")

		v := url.Query()
		v.Set("page_size", "5")
		v.Set("page", strconv.Itoa(page))
		url.RawQuery = v.Encode()

		body, err := t.performRequest(ctx, url, token, http.MethodGet, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list vehicles: %w", err)
		}

		var vehicles TeslaResponseWrapper[[]TeslaVehicle]
		err = json.Unmarshal(body, &vehicles)
		if err != nil {
			return nil, fmt.Errorf("invalid response encountered while fetching user vehicles: %w", err)
		}

		out = append(out, vehicles.Response...)

		if vehicles.Pagination.Next == 0 {
			t.log.Info().Msgf("Took %s to page through %d vehicles.", time.Since(listStart), len(out))
			return out, nil
		}

		page = vehicles.Pagination.Next
	}
}

// GetVehicle calls Tesla Fleet API to get a single vehicle by ID
func (t *teslaFleetAPIService) GetVehicle(ctx context.Context, token string, vehicleID int) (*TeslaVehicle, error) {
	url := t.FleetBase.JoinPath("api/1/vehicles", strconv.Itoa(vehicleID))

	body, err := t.performRequest(ctx, url, token, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch vehicles for user: %w", err)
	}

	var vehicle TeslaResponseWrapper[TeslaVehicle]
	err = json.Unmarshal(body, &vehicle)
	if err != nil {
		return nil, fmt.Errorf("invalid response encountered while fetching vehicles: %w", err)
	}

	return &vehicle.Response, nil
}

// WakeUpVehicle Calls Tesla Fleet API to wake a vehicle from sleep
func (t *teslaFleetAPIService) WakeUpVehicle(ctx context.Context, token string, vehicleID int) error {
	url := t.FleetBase.JoinPath("api/1/vehicles", strconv.Itoa(vehicleID), "wake_up")

	if _, err := t.performRequest(ctx, url, token, http.MethodPost, nil); err != nil {
		return fmt.Errorf("could not wake vehicle: %w", err)
	}

	return nil
}

// TODO(elffjs): This being here is a bad sign.
type partialTeslaClaims struct {
	jwt.RegisteredClaims
	Scopes []string `json:"scp"`
}

const (
	teslaCommandScope  = "vehicle_cmds"
	teslaChargingScope = "vehicle_charging_cmds"
)

func (t *teslaFleetAPIService) GetAvailableCommands(token string) (*UserDeviceAPIIntegrationsMetadataCommands, error) {
	var claims partialTeslaClaims
	_, _, err := jwt.NewParser().ParseUnverified(token, &claims)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse JWT: %w", err)
	}

	enabled := []string{constants.TelemetrySubscribe} // TODO(elffjs): Maybe not a safe assumption.
	disabled := []string{}

	if slices.Contains(claims.Scopes, teslaCommandScope) {
		enabled = append(enabled, constants.DoorsLock, constants.DoorsUnlock, constants.TrunkOpen, constants.FrunkOpen)
	} else {
		disabled = append(disabled, constants.DoorsLock, constants.DoorsUnlock, constants.TrunkOpen, constants.FrunkOpen)
	}

	if slices.Contains(claims.Scopes, teslaCommandScope) || slices.Contains(claims.Scopes, teslaChargingScope) {
		enabled = append(enabled, constants.ChargeLimit)
	} else {
		disabled = append(disabled, constants.ChargeLimit)
	}

	return &UserDeviceAPIIntegrationsMetadataCommands{
		Enabled:  enabled,
		Disabled: disabled,
	}, nil
}

func (t *teslaFleetAPIService) VirtualKeyConnectionStatus(ctx context.Context, token, vin string) (*VehicleFleetStatus, error) {
	url := t.FleetBase.JoinPath("api/1/vehicles/fleet_status")

	jsonBody := fmt.Sprintf(`{"vins": [%q]}`, vin)
	inBody := strings.NewReader(jsonBody)

	body, err := t.performRequest(ctx, url, token, http.MethodPost, inBody)
	if err != nil {
		t.log.Warn().Str("body", jsonBody).Msg("Virtual key status request failure.")
		return nil, fmt.Errorf("error requesting key status: %w", err)
	}

	var keyConn TeslaResponseWrapper[fleetStatusResponse]
	err = json.Unmarshal(body, &keyConn)
	if err != nil {
		return nil, fmt.Errorf("error decoding key status %w", err)
	}

	vi := keyConn.Response.VehicleInfo[vin]

	return &VehicleFleetStatus{
		KeyPaired:                          len(keyConn.Response.KeyPairedVINs) == 1,
		FirmwareVersion:                    vi.FirmwareVersion,
		DiscountedDeviceData:               vi.DiscountedDeviceData,
		FleetTelemetryVersion:              vi.FleetTelemetryVersion,
		NumberOfKeys:                       vi.TotalNumberOfKeys,
		VehicleCommandProtocolRequired:     vi.VehicleCommandProtocolRequired,
		SafetyScreenStreamingToggleEnabled: vi.SafetyScreenStreamingToggleEnabled,
	}, nil
}

// This is Jeremy's "Advanced" set of fields and intervals. We will get this out of Go soon.
var fields = TelemetryFields{
	"ACChargingEnergyIn":              {IntervalSeconds: 60},
	"ACChargingPower":                 {IntervalSeconds: 60},
	"AutomaticEmergencyBrakingOff":    {IntervalSeconds: 1},
	"BlindSpotCollisionWarningChime":  {IntervalSeconds: 1},
	"BrickVoltageMax":                 {IntervalSeconds: 300},
	"BrickVoltageMin":                 {IntervalSeconds: 300},
	"CarType":                         {IntervalSeconds: 21599},
	"ChargeAmps":                      {IntervalSeconds: 60},
	"ChargeLimitSoc":                  {IntervalSeconds: 3600},
	"ChargerVoltage":                  {IntervalSeconds: 300},
	"ChargingCableType":               {IntervalSeconds: 300},
	"CruiseFollowDistance":            {IntervalSeconds: 60},
	"CruiseSetSpeed":                  {IntervalSeconds: 60},
	"CurrentLimitMph":                 {IntervalSeconds: 1},
	"DCChargingEnergyIn":              {IntervalSeconds: 60},
	"DCChargingPower":                 {IntervalSeconds: 60},
	"DetailedChargeState":             {IntervalSeconds: 60},
	"DoorState":                       {IntervalSeconds: 1},
	"EmergencyLaneDepartureAvoidance": {IntervalSeconds: 1},
	"EnergyRemaining":                 {IntervalSeconds: 60},
	"EstBatteryRange":                 {IntervalSeconds: 300},
	"FastChargerPresent":              {IntervalSeconds: 300},
	"FdWindow":                        {IntervalSeconds: 1},
	"ForwardCollisionWarning":         {IntervalSeconds: 1},
	"FpWindow":                        {IntervalSeconds: 1},
	"GuestModeEnabled":                {IntervalSeconds: 3600},
	"IdealBatteryRange":               {IntervalSeconds: 20},
	"LaneDepartureAvoidance":          {IntervalSeconds: 1},
	"Location":                        {IntervalSeconds: 1},
	"Locked":                          {IntervalSeconds: 300},
	"Odometer":                        {IntervalSeconds: 300},
	"OutsideTemp":                     {IntervalSeconds: 60},
	"RdWindow":                        {IntervalSeconds: 1},
	"RpWindow":                        {IntervalSeconds: 1},
	"Soc":                             {IntervalSeconds: 60},
	"SoftwareUpdateVersion":           {IntervalSeconds: 21599},
	"SpeedLimitWarning":               {IntervalSeconds: 1},
	"TpmsPressureFl":                  {IntervalSeconds: 300},
	"TpmsPressureFr":                  {IntervalSeconds: 300},
	"TpmsPressureRl":                  {IntervalSeconds: 300},
	"TpmsPressureRr":                  {IntervalSeconds: 300},
	"Trim":                            {IntervalSeconds: 21599},
	"VehicleName":                     {IntervalSeconds: 21599},
	"VehicleSpeed":                    {IntervalSeconds: 20},
	"Version":                         {IntervalSeconds: 21599},
}

type TeslaSubscriptionErrorType int

const (
	KeyUnpaired TeslaSubscriptionErrorType = iota
	UnsupportedVehicle
	UnsupportedFirmware
	MaxConfigs
)

// TeslaSubscriptionError is an error containing text suitable for showing to the user.
// It indicates user error.
type TeslaSubscriptionError struct {
	internal string
	Type     TeslaSubscriptionErrorType
}

func (e *TeslaSubscriptionError) Error() string {
	return e.internal
}

func (t *teslaFleetAPIService) SubscribeForTelemetryData(ctx context.Context, token string, vin string) error {
	url := t.FleetBase.JoinPath("api/1/vehicles/fleet_telemetry_config")

	r := SubscribeForTelemetryDataRequest{
		VINs: []string{vin},
		Config: TelemetryConfigRequest{
			HostName:    t.Settings.TeslaTelemetryHostName,
			CA:          t.Settings.TeslaTelemetryCACertificate,
			Port:        t.Settings.TeslaTelemetryPort,
			Fields:      fields,
			PreferTyped: true,
		},
	}

	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	body, err := t.performRequest(ctx, url, token, http.MethodPost, bytes.NewReader(b))
	if err != nil {
		return err
	}

	var subResp TeslaResponseWrapper[SubscribeForTelemetryDataResponse]
	err = json.Unmarshal(body, &subResp)
	if err != nil {
		return err
	}

	if subResp.Response.UpdatedVehicles == 1 {
		return nil
	}

	if slices.Contains(subResp.Response.SkippedVehicles.MissingKey, vin) {
		return &TeslaSubscriptionError{internal: "virtual key not added to vehicle", Type: KeyUnpaired}
	}

	if slices.Contains(subResp.Response.SkippedVehicles.UnsupportedHardware, vin) {
		return &TeslaSubscriptionError{internal: "vehicle hardware not supported", Type: UnsupportedVehicle}
	}

	if slices.Contains(subResp.Response.SkippedVehicles.UnsupportedFirmware, vin) {
		return &TeslaSubscriptionError{internal: "vehicle firmware not supported", Type: UnsupportedFirmware}
	}

	if slices.Contains(subResp.Response.SkippedVehicles.MaxConfigs, vin) {
		return &TeslaSubscriptionError{internal: "vehicle firmware not supported", Type: MaxConfigs}
	}

	return nil
}

func (t *teslaFleetAPIService) GetTelemetrySubscriptionStatus(ctx context.Context, token, vin string) (*VehicleTelemetryStatus, error) {
	u := t.FleetBase.JoinPath("api/1/vehicles", vin, "fleet_telemetry_config")

	body, err := t.performRequest(ctx, u, token, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var statResp TeslaResponseWrapper[TelemetryConfigStatusResponse]
	err = json.Unmarshal(body, &statResp)
	if err != nil {
		return nil, err
	}

	return &VehicleTelemetryStatus{
		KeyPaired:    statResp.Response.KeyPaired,
		Synced:       statResp.Response.Synced,
		Configured:   statResp.Response.Config != nil,
		LimitReached: statResp.Response.LimitReached,
	}, nil
}

var ErrUnauthorized = errors.New("unauthorized")

// performRequest a helper function for making http requests, it adds a timeout context and parses error response
func (t *teslaFleetAPIService) performRequest(ctx context.Context, url *url.URL, token, method string, body io.Reader) ([]byte, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, method, url.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusMisdirectedRequest {
			return nil, ErrWrongRegion
		}
		if typ, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type")); err != nil {
			return nil, fmt.Errorf("status code %d and unparseable content type %q: %w", resp.StatusCode, resp.Header.Get("Content-Type"), err)
		} else if typ != "application/json" {
			return nil, fmt.Errorf("status code %d and non-JSON content type %s", resp.StatusCode, resp.Header.Get("Content-Type"))
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		var errBody TeslaFleetAPIError
		if err := json.Unmarshal(b, &errBody); err != nil {
			return nil, fmt.Errorf("couldn't parse Tesla error response body: %w", err)
		}
		t.log.Info().Int("code", resp.StatusCode).Str("error", errBody.Error).Str("errorDescription", errBody.ErrorDescription).Str("url", url.String()).Msg("Tesla error.")

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			return nil, ErrUnauthorized
		}

		return nil, fmt.Errorf("error occurred calling Tesla api: %s", errBody.Error)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return b, nil
}
