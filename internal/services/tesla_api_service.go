package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/rs/zerolog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:generate mockgen -source tesla_api_service.go -destination mocks/tesla_api_service_mock.go
type TeslaAPIService interface {
	CompleteTeslaAuthCodeExchange(authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error)
	GetVehicles(token, region string) ([]TeslaVehicle, error)
}

type teslaAPIService struct {
	Settings   *config.Settings
	HTTPClient *http.Client
	log        *zerolog.Logger
}

func NewTeslaAPIService(settings *config.Settings, logger *zerolog.Logger) TeslaAPIService {
	return &teslaAPIService{
		Settings: settings,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		log: logger,
	}
}

func (t *teslaAPIService) CompleteTeslaAuthCodeExchange(authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error) {
	u, err := url.Parse(t.Settings.TeslaAuthorization.TokenAuthURL)
	if err != nil {
		return nil, fmt.Errorf("Could not complete tesla auth: %w", err)
	}
	f := url.Values{}
	f.Set("grant_type", "authorization_code")
	f.Set("client_id", t.Settings.TeslaAuthorization.ClientID)
	f.Set("client_secret", t.Settings.TeslaAuthorization.ClientSecret)
	f.Set("code", authCode)
	f.Set("redirect_uri", redirectURI)
	f.Set("scope", "openid offline_access "+strings.Join(teslaScopes, " "))
	f.Set("audience", fmt.Sprintf("https://fleet-api.prd.%s.vn.cloud.tesla.com", region))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(f.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errBody := new(TeslaError)
		if err := json.NewDecoder(res.Body).Decode(errBody); err != nil {
			t.log.
				Err(err).
				Str("url", t.Settings.TeslaAuthorization.TokenAuthURL).
				Msg("An error occurred when attempting to decode the error message from the api.")
			return nil, fmt.Errorf("error occurred completing authorization, invalid response received from during authorization process: %s", errBody.ErrorDescription)
		}
		return nil, fmt.Errorf("error occurred completing authorization: %s", errBody.ErrorDescription)
	}

	teslaAuth := new(TeslaAuthCodeResponse)
	if err := json.NewDecoder(res.Body).Decode(teslaAuth); err != nil {
		return nil, err
	}

	return teslaAuth, nil
}

func (t *teslaAPIService) GetVehicles(token, region string) ([]TeslaVehicle, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("fleet-api.prd.%s.vn.cloud.tesla.com", region),
		Path:   "api/1/vehicles",
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody := new(TeslaError)
		if err := json.NewDecoder(resp.Body).Decode(errBody); err != nil {
			t.log.
				Err(err).
				Str("url", fmt.Sprintf("fleet-api.prd.%s.vn.cloud.tesla.com", region)).
				Msg("An error occurred when attempting to decode the error message from the api.")
			return nil, fmt.Errorf("error occurred completing authorization, invalid response received from during authorization process: %s", errBody.ErrorDescription)
		}
		return nil, fmt.Errorf("error occurred completing authorization: %s", errBody.ErrorDescription)
	}

	vehicles := new(GetVehiclesResponse)
	if err := json.NewDecoder(resp.Body).Decode(vehicles); err != nil {
		return nil, err
	}

	if vehicles.Response == nil {
		return nil, fmt.Errorf("Error occurred fetching vehicles")
	}

	return vehicles.Response, nil
}
