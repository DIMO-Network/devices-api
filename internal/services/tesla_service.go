package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
)

//go:generate mockgen -source tesla_service.go -destination mocks/tesla_service_mock.go
type TeslaService interface {
	GetVehicle(ownerAccessToken string, id int) (*TeslaVehicle, error)
	WakeUpVehicle(ownerAccessToken string, id int) error
	CompleteTeslaAuthCodeExchange(authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error)
	GetVehicles(token, region string) ([]TeslaVehicle, error)
}

type teslaService struct {
	Settings   *config.Settings
	HTTPClient *http.Client
}

func NewTeslaService(settings *config.Settings) TeslaService {
	return &teslaService{
		Settings: settings,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (t *teslaService) GetVehicle(ownerAccessToken string, id int) (*TeslaVehicle, error) {
	u := fmt.Sprintf("https://owner-api.teslamotors.com/api/1/vehicles/%d", id)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ownerAccessToken)
	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d retrieving vehicle %d", resp.StatusCode, id)
	}

	respBody := new(struct {
		Response TeslaVehicle `json:"response"`
	})

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	return &respBody.Response, nil
}

func (t *teslaService) WakeUpVehicle(ownerAccessToken string, id int) error {
	u := fmt.Sprintf("https://owner-api.teslamotors.com/api/1/vehicles/%d/wake_up", id)
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+ownerAccessToken)
	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status code %d waking up vehicle %d", resp.StatusCode, id)
	}

	return nil
}

type TeslaVehicle struct {
	ID        int    `json:"id"`
	VehicleID int    `json:"vehicle_id"`
	VIN       string `json:"vin"`
}

type GetVehiclesResponse struct {
	Response *[]TeslaVehicle `json:"response"`
}

type TeslaError struct {
	Error            string
	ErrorDescription string `json:"error_description"`
	ReferenceID      string
}

var teslaScopes = []string{"user_data", "vehicle_device_data", "vehicle_cmds", "vehicle_charging_cmds", "energy_device_data", "energy_device_data", "energy_cmds"}

type TeslaAuthCodeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	State        string `json:"state"`
	TokenType    string `json:"token_type"`
}

func (t *teslaService) CompleteTeslaAuthCodeExchange(authCode, redirectURI, region string) (*TeslaAuthCodeResponse, error) {
	u, _ := url.Parse("https://auth.tesla.com/oauth2/v3/token")
	f := url.Values{}
	f.Set("grant_type", "authorization_code")
	f.Set("client_id", t.Settings.TeslaAuthorization.ClientID)
	f.Set("client_secret", t.Settings.TeslaAuthorization.ClientSecret)
	f.Set("code", authCode)
	f.Set("redirect_uri", redirectURI)
	f.Set("scope", "openid offline_access "+strings.Join(teslaScopes, " "))
	f.Set("audience", fmt.Sprintf("https://fleet-api.prd.%s.vn.cloud.tesla.com", region))

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(f.Encode()))
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
			return nil, err
		}
		return nil, fmt.Errorf("error occurred completing authorization: %s", errBody.ErrorDescription)
	}

	teslaAuth := new(TeslaAuthCodeResponse)
	if err := json.NewDecoder(res.Body).Decode(teslaAuth); err != nil {
		return nil, err
	}

	return teslaAuth, nil
}

func (t *teslaService) GetVehicles(token, region string) ([]TeslaVehicle, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("fleet-api.prd.%s.vn.cloud.tesla.com", region),
		Path:   "api/1/vehicles",
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return []TeslaVehicle{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []TeslaVehicle{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody := new(TeslaError)
		if err := json.NewDecoder(resp.Body).Decode(errBody); err != nil {
			return []TeslaVehicle{}, err
		}
		return []TeslaVehicle{}, fmt.Errorf("error occurred completing authorization: %s", errBody.ErrorDescription)
	}

	vehicles := new(GetVehiclesResponse)
	if err := json.NewDecoder(resp.Body).Decode(vehicles); err != nil {
		return []TeslaVehicle{}, err
	}

	if vehicles.Response == nil {
		return []TeslaVehicle{}, fmt.Errorf("Error occurred fetching vehicles")
	}

	return *vehicles.Response, nil
}
