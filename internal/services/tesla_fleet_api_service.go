package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	VirtualTokenConnectionStatus(ctx context.Context, token, region, vin string) (bool, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TeslaAuthCodeResponse, error)
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

type VirtualTokenConnectionStatusResponse struct {
	Response VirtualTokenConnectionStatus `json:"response"`
}

type VirtualTokenConnectionStatus struct {
	UnpairedVins  []string `json:"unpaired_vins"`
	KeyPairedVins []string `json:"key_paired_vins"`
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

	resp, err := t.performRequest(ctx, url, token, http.MethodGet, nil)
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

	resp, err := t.performRequest(ctx, url, token, http.MethodGet, nil)
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

	resp, err := t.performRequest(ctx, url, token, http.MethodGet, nil)
	if err != nil {
		return fmt.Errorf("could not fetch vehicles for user: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status code %d waking up vehicle %d", resp.StatusCode, vehicleID)
	}

	return nil
}

func (t *teslaFleetAPIService) GetAvailableCommands() *UserDeviceAPIIntegrationsMetadataCommands {
	return &UserDeviceAPIIntegrationsMetadataCommands{
		Enabled: []string{constants.DoorsUnlock, constants.DoorsLock, constants.TrunkOpen, constants.FrunkOpen, constants.ChargeLimit},
		Capable: []string{constants.DoorsUnlock, constants.DoorsLock, constants.TrunkOpen, constants.FrunkOpen, constants.ChargeLimit, constants.TelemetrySubscribe},
	}
}

// VirtualTokenConnectionStatus Checks whether vehicles can accept Tesla commands protocol for the partner's public key
func (t *teslaFleetAPIService) VirtualTokenConnectionStatus(ctx context.Context, token, region, vin string) (bool, error) {
	baseURL := fmt.Sprintf(t.Settings.TeslaFleetURL, region)
	url := fmt.Sprintf("%s/api/1/vehicles/fleet_status", baseURL)

	jsonBody := fmt.Sprintf(`{"vins": [%q]}`, vin)
	body := strings.NewReader(jsonBody)
	// bytes.NewReader(jsonBody)

	resp, err := t.performRequest(ctx, url, token, http.MethodPost, body)
	if err != nil {
		return false, fmt.Errorf("could not fetch vehicles for user: %w", err)
	}

	defer resp.Body.Close()

	var v VirtualTokenConnectionStatusResponse
	bd, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("could not verify connection status %w", err)
	}

	err = json.Unmarshal(bd, &v)
	if err != nil {
		return false, fmt.Errorf("error occurred decoding connection status %w", err)
	}

	isConnected := slices.Contains(v.Response.KeyPairedVins, vin)

	return isConnected, nil
}

func (t *teslaFleetAPIService) RefreshToken(ctx context.Context, refreshToken string) (*TeslaAuthCodeResponse, error) {
	reqs := struct {
		GrantType    string `json:"grant_type"`
		ClientID     string `json:"client_id"`
		RefreshToken string `json:"refresh_token"`
	}{
		GrantType:    "refresh_token",
		ClientID:     t.Settings.TeslaClientID,
		RefreshToken: refreshToken,
	}

	reqb, err := json.Marshal(reqs)
	if err != nil {
		return nil, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, "POST", t.Settings.TeslaTokenURL, bytes.NewBuffer(reqb))
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

	tokResp := new(TeslaAuthCodeResponse)
	if err := json.NewDecoder(resp.Body).Decode(tokResp); err != nil {
		return nil, err
	}

	return tokResp, nil
}

// performRequest a helper function for making http requests, it adds a timeout context and parses error response
func (t *teslaFleetAPIService) performRequest(ctx context.Context, url, token, method string, body *strings.Reader) (*http.Response, error) {
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

	if resp.StatusCode != http.StatusOK {
		errBody := new(TeslaFleetAPIError)
		if err := json.NewDecoder(resp.Body).Decode(errBody); err != nil {
			t.log.
				Err(err).
				Str("url", url).
				Msg("An error occurred when attempting to decode the error message from the api.")
			return nil, fmt.Errorf("invalid response encountered while fetching user vehicles: %s", errBody.ErrorDescription)
		}
		return nil, fmt.Errorf("error occurred calling tesla api: %s", errBody.ErrorDescription)
	}

	return resp, nil
}
