package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

//go:generate mockgen -source tesla_fleet_api_service.go -destination mocks/tesla_fleet_api_service_mock.go
type TeslaFleetAPIService interface {
	CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error)
	GetVehicles(ctx context.Context, token, region string) ([]TeslaVehicle, error)
}

var teslaScopes = []string{"openid", "offline_access", "user_data", "vehicle_device_data", "vehicle_cmds", "vehicle_charging_cmds", "energy_device_data", "energy_device_data", "energy_cmds"}

type GetVehiclesResponse struct {
	Response []TeslaVehicle `json:"response"`
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

// CompleteTeslaAuthCodeExchange godoc
// @Description Call Tesla Fleet API and exchange auth code for a new auth and refresh token
// @Param       authCode - authorization code to exchange
// @Param       redirectURI - redirect uri to pass on as part of the request to for oauth exchange
// @Param       region - API region which is used to determine which fleet api to call
// @Returns     {object} services.TeslaAuthCodeResponse
func (t *teslaFleetAPIService) CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error) {
	conf := oauth2.Config{
		ClientID:     t.Settings.Tesla.ClientID,
		ClientSecret: t.Settings.Tesla.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: t.Settings.Tesla.TokenURL,
		},
		RedirectURL: redirectURI,
		Scopes:      teslaScopes,
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	tok, err := conf.Exchange(ctxTimeout, authCode, oauth2.SetAuthURLParam("audience", fmt.Sprintf(t.Settings.Tesla.FleetAPI, region)))
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

// GetVehicles godoc
// @Description Call Tesla Fleet API to get a list of vehicles using authorization token
// @Param       token - authorization token to be used as bearer token
// @Param       region - API region which is used to determine which fleet api to call
// @Returns     {array} []services.TeslaVehicle
func (t *teslaFleetAPIService) GetVehicles(ctx context.Context, token, region string) ([]TeslaVehicle, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf(t.Settings.Tesla.FleetAPI, region),
		Path:   "api/1/vehicles",
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch vehicles for user: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch vehicles for user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody := new(TeslaFleetAPIError)
		if err := json.NewDecoder(resp.Body).Decode(errBody); err != nil {
			t.log.
				Err(err).
				Str("url", fmt.Sprintf(t.Settings.Tesla.FleetAPI, region)).
				Msg("An error occurred when attempting to decode the error message from the api.")
			return nil, fmt.Errorf("invalid response encountered while fetching user vehicles: %s", errBody.ErrorDescription)
		}
		return nil, fmt.Errorf("error occurred fetching user vehicles: %s", errBody.ErrorDescription)
	}

	vehicles := new(GetVehiclesResponse)
	if err := json.NewDecoder(resp.Body).Decode(vehicles); err != nil {
		return nil, fmt.Errorf("invalid response encountered while fetching user vehicles: %w", err)
	}

	if vehicles.Response == nil {
		return nil, fmt.Errorf("error occurred fetching user vehicles")
	}

	return vehicles.Response, nil
}
