package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/pkg/errors"
	smartcar "github.com/smartcar/go-sdk"
)

//go:generate mockgen -source smartcar_client.go -destination mocks/smartcar_client_mock.go

type SmartcarClient interface {
	// The reason redirectURI is there is the frontend wanted flexibility. It's probably a
	// bad idea that the client can pass this in.
	ExchangeCode(ctx context.Context, code, redirectURI string) (*smartcar.Token, error)
	GetUserID(ctx context.Context, accessToken string) (string, error)
	GetExternalID(ctx context.Context, accessToken string) (string, error)
	GetEndpoints(ctx context.Context, accessToken string, id string) ([]string, error)
	HasDoorControl(ctx context.Context, accessToken string, id string) (bool, error)
	GetVIN(ctx context.Context, accessToken string, id string) (string, error)
	GetYear(ctx context.Context, accessToken string, id string) (int, error)
	GetInfo(ctx context.Context, accessToken string, id string) (*smartcar.Info, error)
}

type smartcarClient struct {
	settings       *config.Settings
	officialClient smartcar.Client
	exchangeURL    string
	baseURL        string
	httpClient     *http.Client
}

func NewSmartcarClient(settings *config.Settings) SmartcarClient {
	scClient := smartcar.NewClient()
	scClient.SetAPIVersion("2.0")
	return &smartcarClient{
		settings:       settings,
		officialClient: scClient,
		exchangeURL:    "https://auth.smartcar.com/oauth/token/",
		baseURL:        "https://api.smartcar.com/v2.0/",
		httpClient:     &http.Client{Timeout: time.Duration(310) * time.Second}, // Smartcar default.
	}
}

const smartcarDoorPermission = "control_security"

var scopeToEndpoints = map[string][]string{
	"read_engine_oil":   {"/engine/oil"},
	"read_battery":      {"/battery/capacity", "/battery"},
	"read_charge":       {"/charge/limit", "/charge"},
	"read_fuel":         {"/fuel"},
	"read_location":     {"/location"},
	"read_odometer":     {"/odometer"},
	"read_tires":        {"/tires/pressure"},
	"read_vehicle_info": {"/"},
	"read_vin":          {"/vin"},
}

type scExchangeRes struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

var scReqIDHeader = "SC-Request-Id"

type SmartcarError struct {
	RequestID string
	Code      int
	Body      []byte
}

func (e *SmartcarError) Error() string {
	return fmt.Sprintf("smartcar: status code %d", e.Code)
}

func (s *smartcarClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*smartcar.Token, error) {
	v := url.Values{}
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")
	v.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", s.exchangeURL, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.settings.SmartcarClientID+":"+s.settings.SmartcarClientSecret)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "DIMO/1.0")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	reqID := res.Header.Get(scReqIDHeader)

	bb, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, &SmartcarError{
			RequestID: reqID,
			Code:      res.StatusCode,
			Body:      bb,
		}
	}

	var t scExchangeRes
	if err := json.Unmarshal(bb, &t); err != nil {
		return nil, err
	}

	return &smartcar.Token{
		Access:       t.AccessToken,
		AccessExpiry: time.Now().Add(time.Duration(t.ExpiresIn) * time.Second),
		Refresh:      t.RefreshToken,
	}, nil
}

func (s *smartcarClient) GetUserID(ctx context.Context, accessToken string) (string, error) {
	id, err := s.officialClient.GetUserID(ctx, &smartcar.UserIDParams{Access: accessToken})
	if err != nil {
		return "", err
	}
	if id == nil {
		return "", errors.New("no error from Smartcar and yet no user id")
	}
	return *id, nil
}

func (s *smartcarClient) GetExternalID(ctx context.Context, accessToken string) (string, error) {
	ids, err := s.officialClient.GetVehicleIDs(ctx, &smartcar.VehicleIDsParams{Access: accessToken})
	if err != nil {
		return "", err
	}
	if ids == nil || len(*ids) != 1 {
		return "", errors.New("should only be one vehicle under the access token")
	}
	return (*ids)[0], nil
}

// GetEndpoints returns the Smartcar read endpoints granted to the access token.
func (s *smartcarClient) GetEndpoints(ctx context.Context, accessToken string, id string) ([]string, error) {
	v := s.officialClient.NewVehicle(&smartcar.VehicleParams{
		ID:          id,
		AccessToken: accessToken,
		UnitSystem:  smartcar.Metric,
	})
	perms, err := v.GetPermissions(ctx)
	if err != nil {
		return nil, err
	}
	if perms == nil {
		return nil, errors.New("nil permissions object")
	}

	endpoints := []string{}

	for _, perm := range perms.Permissions {
		scopeEndpoints, ok := scopeToEndpoints[perm]
		if !ok {
			continue
		}
		endpoints = append(endpoints, scopeEndpoints...)
	}

	return endpoints, nil
}

// HasDoorControl returns true if the access token can open and close doors.
// TODO(elffjs): Probably silly to have both this and GetEndpoints.
func (s *smartcarClient) HasDoorControl(ctx context.Context, accessToken string, id string) (bool, error) {
	v := s.officialClient.NewVehicle(&smartcar.VehicleParams{
		ID:          id,
		AccessToken: accessToken,
		UnitSystem:  smartcar.Metric,
	})
	perms, err := v.GetPermissions(ctx)
	if err != nil {
		return false, err
	}
	if perms == nil {
		return false, errors.New("nil permissions object")
	}

	for _, perm := range perms.Permissions {
		if perm == smartcarDoorPermission {
			return true, nil
		}
	}

	return false, nil
}

func (s *smartcarClient) GetVIN(ctx context.Context, accessToken string, id string) (string, error) {
	v := s.officialClient.NewVehicle(&smartcar.VehicleParams{
		ID:          id,
		AccessToken: accessToken,
		UnitSystem:  smartcar.Metric,
	})
	vin, err := v.GetVIN(ctx)
	if err != nil {
		return "", err
	}
	if vin == nil {
		return "", errors.New("nil VIN object")
	}
	return vin.VIN, nil
}

type VehicleAttributesRes struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
}

func (s *smartcarClient) GetYear(ctx context.Context, accessToken string, id string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/vehicles/%s", s.baseURL, id), nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	reqID := res.Header.Get(scReqIDHeader)

	bb, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, &SmartcarError{
			RequestID: reqID,
			Code:      res.StatusCode,
			Body:      bb,
		}
	}

	var v VehicleAttributesRes
	if err := json.Unmarshal(bb, &v); err != nil {
		return 0, err
	}

	return v.Year, nil
}

func (s *smartcarClient) GetInfo(ctx context.Context, accessToken string, id string) (*smartcar.Info, error) {
	v := s.officialClient.NewVehicle(&smartcar.VehicleParams{
		ID:          id,
		AccessToken: accessToken,
		UnitSystem:  smartcar.Metric,
	})
	info, err := v.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, errors.New("nil info object")
	}
	return info, nil
}
