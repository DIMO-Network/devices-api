package smartcar

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
	smartcar "github.com/smartcar/go-sdk"
)

//go:generate mockgen -source smartcar_client.go -destination mocks/smartcar_client_mock.go

type SmartcarClient struct {
	settings    *config.Settings
	exchangeURL string
	baseURL     string
	httpClient  *http.Client
}

func New() *SmartcarClient {
	return &SmartcarClient{
		exchangeURL: "https://auth.smartcar.com/oauth/token/",
		baseURL:     "https://api.smartcar.com/v2.0",
		httpClient:  &http.Client{Timeout: time.Duration(310) * time.Second}, // Smartcar default.
	}
}

var permEndpoints = map[string][]string{
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

func DataEndpoints(permissions []string) []string {
	out := []string{}
	for _, p := range permissions {
		out = append(out, permEndpoints[p]...)
	}
	return out
}

type scExchangeRes struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

var scReqIDHeader = "SC-Request-Id"

type Error struct {
	RequestID string
	Code      int
	Body      []byte
}

func (e *Error) Error() string {
	return fmt.Sprintf("smartcar: status code %d", e.Code)
}

func (s *SmartcarClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*smartcar.Token, error) {
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

	bb, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, &Error{
			RequestID: res.Header.Get(scReqIDHeader),
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

type scListResp struct {
	Vehicles []string `json:"vehicles"`
}

func (s *SmartcarClient) GetExternalID(ctx context.Context, accessToken string) (string, error) {
	var r scListResp
	if err := s.do(ctx, accessToken, "/vehicles", &r); err != nil {
		return "", err
	}

	if len(r.Vehicles) != 1 {
		return "", fmt.Errorf("expected one vehicle in list, but got %d", len(r.Vehicles))
	}

	return r.Vehicles[0], nil
}

type scPermissions struct {
	Permissions []string `json:"permissions"`
}

func (s *SmartcarClient) GetPermissions(ctx context.Context, accessToken string, id string) ([]string, error) {
	var r scPermissions
	if err := s.do(ctx, accessToken, fmt.Sprintf("/vehicles/%s/permissions", id), &r); err != nil {
		return nil, err
	}

	return r.Permissions, nil
}

type scVIN struct {
	VIN string `json:"vin"`
}

func (s *SmartcarClient) GetVIN(ctx context.Context, accessToken string, id string) (string, error) {
	var r scVIN
	if err := s.do(ctx, accessToken, fmt.Sprintf("/vehicles/%s/vin", id), &r); err != nil {
		return "", err
	}
	return r.VIN, nil
}

type MMY struct {
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
}

func (s *SmartcarClient) GetInfo(ctx context.Context, accessToken string, id string) (*MMY, error) {
	var r MMY
	if err := s.do(ctx, accessToken, fmt.Sprintf("/vehicles/%s", id), &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *SmartcarClient) do(ctx context.Context, accessToken string, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	bb, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return &Error{
			RequestID: res.Header.Get(scReqIDHeader),
			Code:      res.StatusCode,
			Body:      bb,
		}
	}

	return json.Unmarshal(bb, v)
}
