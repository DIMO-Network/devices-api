package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"time"
)

//go:generate mockgen -source tesla_api_service.go -destination mocks/tesla_api_service_mock.go
type TeslaAPIService interface {
	CompleteTeslaAuthCodeExchange(authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error)
	GetVehicles(token, region string) ([]TeslaVehicle, error)
}

var teslaScopes = []string{"openid offline_access", "user_data", "vehicle_device_data", "vehicle_cmds", "vehicle_charging_cmds", "energy_device_data", "energy_device_data", "energy_cmds"}

type GetVehiclesResponse struct {
	Response []TeslaVehicle `json:"response"`
}

type TeslaError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ReferenceID      string `json:"ReferenceID"`
}

type TeslaAuthCodeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	State        string `json:"state"`
	TokenType    string `json:"token_type"`
}

type teslaAPIService struct {
	Settings   *config.Settings
	HTTPClient *http.Client
	log        *zerolog.Logger
}

func NewTeslaAPIService(settings *config.Settings, logger *zerolog.Logger) TeslaAPIService {
	return &teslaAPIService{
		Settings:   settings,
		HTTPClient: &http.Client{},
		log:        logger,
	}
}

func (t *teslaAPIService) CompleteTeslaAuthCodeExchange(authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error) {
	conf := oauth2.Config{
		ClientID:     t.Settings.Tesla.ClientID,
		ClientSecret: t.Settings.Tesla.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: t.Settings.Tesla.TokenAuthURL,
		},
		RedirectURL: redirectURI,
		Scopes:      teslaScopes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	tok, err := conf.Exchange(ctx, authCode, oauth2.SetAuthURLParam("audience", fmt.Sprintf("https://fleet-api.prd.%s.vn.cloud.tesla.com", region)), oauth2.SetAuthURLParam("grant_type", "authorization_code"))
	if err != nil {
		var e *oauth2.RetrieveError
		errors.As(err, &e)
		return nil, fmt.Errorf("error occurred completing authorization: %s", e.ErrorDescription)
	}

	return &TeslaAuthCodeResponse{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresIn:    int(tok.Expiry.Unix()),
		TokenType:    tok.TokenType,
	}, nil
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

	resp, err := t.HTTPClient.Do(req)
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
		return nil, fmt.Errorf("error occurred fetching vehicles")
	}

	return vehicles.Response, nil
}
