package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
)

//go:generate mockgen -source tesla_service.go -destination mocks/tesla_service_mock.go
type TeslaService interface {
	GetVehicle(ownerAccessToken string, id int) (*TeslaVehicle, error)
	WakeUpVehicle(ownerAccessToken string, id int) error
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
