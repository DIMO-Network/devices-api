package services

import (
	"context"

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
}

type smartcarClient struct {
	settings       *config.Settings
	officialClient smartcar.Client
}

func NewSmartcarClient(settings *config.Settings) SmartcarClient {
	scClient := smartcar.NewClient()
	scClient.SetAPIVersion("2.0")
	return &smartcarClient{
		settings:       settings,
		officialClient: scClient,
	}
}

var smartcarScopes = []string{
	"control_security",
	"control_charge",
	"read_engine_oil",
	"read_battery",
	"read_charge",
	"control_charge",
	"read_fuel",
	"read_location",
	"read_odometer",
	"read_tires",
	"read_vehicle_info",
	"read_vin",
}

const smartcarDoorPermission = "control_security"

var scopeToEndpoints = map[string][]string{
	"read_engine_oil":   {"/engine/oil"},
	"read_battery":      {"/battery/capacity", "/battery"},
	"read_charge":       {"/charge", "/battery"},
	"read_fuel":         {"/fuel"},
	"read_location":     {"/location"},
	"read_odometer":     {"/odometer"},
	"read_tires":        {"/tires/pressure"},
	"read_vehicle_info": {"/"},
	"read_vin":          {"/vin"},
}

func (s *smartcarClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*smartcar.Token, error) {
	params := &smartcar.AuthParams{
		ClientID:     s.settings.SmartcarClientID,
		ClientSecret: s.settings.SmartcarClientSecret,
		RedirectURI:  redirectURI,
		Scope:        smartcarScopes,
	}
	return s.officialClient.NewAuth(params).ExchangeCode(ctx, &smartcar.ExchangeCodeParams{Code: code})
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

func (s *smartcarClient) GetYear(ctx context.Context, accessToken string, id string) (int, error) {
	v := s.officialClient.NewVehicle(&smartcar.VehicleParams{
		ID:          id,
		AccessToken: accessToken,
		UnitSystem:  smartcar.Metric,
	})
	info, err := v.GetInfo(ctx)
	if err != nil {
		return 0, err
	}
	if info == nil {
		return 0, errors.New("nil info object")
	}
	return info.Year, nil
}
