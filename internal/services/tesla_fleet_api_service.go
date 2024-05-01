package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
)

//go:generate mockgen -source tesla_fleet_api_service.go -destination mocks/tesla_fleet_api_service_mock.go
type TeslaFleetAPIService interface {
	CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error)
	GetVehicles(ctx context.Context, token, region string) ([]TeslaVehicle, error)
	GetVehicle(ctx context.Context, token, region string, vehicleID int) (*TeslaVehicle, error)
	WakeUpVehicle(ctx context.Context, token, region string, vehicleID int) error
	GetAvailableCommands() *UserDeviceAPIIntegrationsMetadataCommands
	VirtualKeyConnectionStatus(ctx context.Context, token, region, vin string) (bool, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TeslaAuthCodeResponse, error)
	SubscribeForTelemetryData(ctx context.Context, token, region, vin string) error
}

var teslaScopes = []string{"openid", "offline_access", "user_data", "vehicle_device_data", "vehicle_cmds", "vehicle_charging_cmds", "energy_device_data", "energy_device_data", "energy_cmds"}

type TeslaResponseWrapper[A any] struct {
	Response A `json:"response"`
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

type VirtualKeyConnectionStatusResponse struct {
	Response VirtualKeyConnectionStatus `json:"response"`
}

type VirtualKeyConnectionStatus struct {
	UnpairedVINs  []string `json:"unpaired_vins"`
	KeyPairedVINs []string `json:"key_paired_vins"`
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
	HostName            string          `json:"hostName"`
	PublicCACertificate string          `json:"ca"`
	Fields              TelemetryFields `json:"fields"`
	Port                int             `json:"port"`
}

type SkippedVehicles struct {
	MissingKey          []string `json:"missing_key"`
	UnsupportedHardware []string `json:"unsupported_hardware"`
	UnsupportedFirmware []string `json:"unsupported_firmware"`
}

type SubscribeForTelemetryDataResponse struct {
	UpdatedVehicles int             `json:"updated_vehicles"`
	SkippedVehicles SkippedVehicles `json:"skipped_vehicles"`
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
	url, err := url.JoinPath(t.fleetURLForRegion(region), "/api/1/vehicles")
	if err != nil {
		return nil, err
	}

	body, err := t.performRequest(ctx, url, token, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch vehicles for user: %w", err)
	}

	var vehicles TeslaResponseWrapper[[]TeslaVehicle]
	err = json.Unmarshal(body, &vehicles)
	if err != nil {
		return nil, fmt.Errorf("invalid response encountered while fetching user vehicles: %w", err)
	}

	if vehicles.Response == nil {
		return nil, fmt.Errorf("error occurred fetching user vehicles")
	}

	return vehicles.Response, nil
}

// GetVehicle calls Tesla Fleet API to get a single vehicle by ID
func (t *teslaFleetAPIService) GetVehicle(ctx context.Context, token, region string, vehicleID int) (*TeslaVehicle, error) {
	url, err := url.JoinPath(t.fleetURLForRegion(region), "/api/1/vehicles", strconv.Itoa(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("error constructing URL: %w", err)
	}

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
func (t *teslaFleetAPIService) WakeUpVehicle(ctx context.Context, token, region string, vehicleID int) error {
	url, err := url.JoinPath(t.fleetURLForRegion(region), "/api/1/vehicles", strconv.Itoa(vehicleID), "wake_up")
	if err != nil {
		return fmt.Errorf("error constructing URL: %w", err)
	}

	_, err = t.performRequest(ctx, url, token, http.MethodGet, nil)
	if err != nil {
		return fmt.Errorf("could not fetch vehicles for user: %w", err)
	}

	return err
}

func (t *teslaFleetAPIService) GetAvailableCommands() *UserDeviceAPIIntegrationsMetadataCommands {
	return &UserDeviceAPIIntegrationsMetadataCommands{
		Enabled: []string{constants.DoorsUnlock, constants.DoorsLock, constants.TrunkOpen, constants.FrunkOpen, constants.ChargeLimit},
		Capable: []string{constants.DoorsUnlock, constants.DoorsLock, constants.TrunkOpen, constants.FrunkOpen, constants.ChargeLimit, constants.TelemetrySubscribe},
	}
}

// VirtualKeyConnectionStatus Checks whether vehicles can accept Tesla commands protocol for the partner's public key
func (t *teslaFleetAPIService) VirtualKeyConnectionStatus(ctx context.Context, token, region, vin string) (bool, error) {
	url, err := url.JoinPath(t.fleetURLForRegion(region), "/api/1/vehicles/fleet_status")
	if err != nil {
		return false, fmt.Errorf("error constructing URL: %w", err)
	}

	jsonBody := fmt.Sprintf(`{"vins": [%q]}`, vin)
	inBody := strings.NewReader(jsonBody)

	body, err := t.performRequest(ctx, url, token, http.MethodPost, inBody)
	if err != nil {
		return false, fmt.Errorf("could not fetch vehicles for user: %w", err)
	}

	var keyConn TeslaResponseWrapper[VirtualKeyConnectionStatus]
	err = json.Unmarshal(body, &keyConn)
	if err != nil {
		return false, fmt.Errorf("error occurred decoding connection status %w", err)
	}

	isConnected := slices.Contains(keyConn.Response.KeyPairedVINs, vin)

	return isConnected, nil
}

func (t *teslaFleetAPIService) RefreshToken(ctx context.Context, refreshToken string) (*TeslaAuthCodeResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", t.Settings.TeslaClientID)
	data.Set("refresh_token", refreshToken)

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, "POST", t.Settings.TeslaTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if code := resp.StatusCode; code != http.StatusOK {
		return nil, fmt.Errorf("status code %d", code)
	}

	var tokResp TeslaAuthCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokResp); err != nil {
		return nil, err
	}

	return &tokResp, nil
}

var fields = TelemetryFields{
	"ChargeState":         {IntervalSeconds: 300},
	"Location":            {IntervalSeconds: 10},
	"OriginLocation":      {IntervalSeconds: 300},
	"DestinationLocation": {IntervalSeconds: 300},
	"DestinationName":     {IntervalSeconds: 300},
	"EnergyRemaining":     {IntervalSeconds: 300},
	"VehicleSpeed":        {IntervalSeconds: 60},
	"Odometer":            {IntervalSeconds: 300},
	"EstBatteryRange":     {IntervalSeconds: 300},
	"Soc":                 {IntervalSeconds: 300},
	"BatteryLevel":        {IntervalSeconds: 60},
}

func (t *teslaFleetAPIService) fleetURLForRegion(region string) string {
	return fmt.Sprintf(t.Settings.TeslaFleetURL, region)
}

func (t *teslaFleetAPIService) SubscribeForTelemetryData(ctx context.Context, token, region, vin string) error {
	u, err := url.JoinPath(t.fleetURLForRegion(region), "/api/1/vehicles/fleet_telemetry_config")
	if err != nil {
		return fmt.Errorf("error constructing URL: %w", err)
	}

	r := SubscribeForTelemetryDataRequest{
		VINs: []string{vin},
		Config: TelemetryConfigRequest{
			HostName:            t.Settings.TeslaTelemetryHostName,
			PublicCACertificate: t.Settings.TeslaTelemetryCACertificate,
			Port:                t.Settings.TeslaTelemetryPort,
			Fields:              fields,
		},
	}

	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	body, err := t.performRequest(ctx, u, token, http.MethodPost, bytes.NewReader(b))
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
		return fmt.Errorf("vehicle has not approved virtual token connection")
	}

	if slices.Contains(subResp.Response.SkippedVehicles.UnsupportedHardware, vin) {
		return fmt.Errorf("vehicle hardware not supported")
	}

	if slices.Contains(subResp.Response.SkippedVehicles.UnsupportedFirmware, vin) {
		return fmt.Errorf("vehicle firmware not supported")
	}

	return nil
}

// performRequest a helper function for making http requests, it adds a timeout context and parses error response
func (t *teslaFleetAPIService) performRequest(ctx context.Context, url, token, method string, body io.Reader) ([]byte, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error occurred calling tesla fleet api: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		t.log.Info().Str("teslaError", string(b)).Int("code", resp.StatusCode).Msg("xdd")
		var errBody TeslaFleetAPIError
		if err := json.Unmarshal(b, &errBody); err != nil {
			t.log.
				Err(err).
				Str("url", url).
				Msg("An error occurred when attempting to decode the error message from the api.")
			return nil, fmt.Errorf("couldn't parse Tesla error response body: %w", err)
		}
		return nil, fmt.Errorf("error occurred calling Tesla api: %s", errBody.ErrorDescription)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return b, nil
}
