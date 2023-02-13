package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type SmartCarService struct {
	baseURL      string
	DBS          func() *db.ReaderWriter
	log          zerolog.Logger // can't remember if best practice with this logger is to use *
	deviceDefSvc DeviceDefinitionService
}

func NewSmartCarService(dbs func() *db.ReaderWriter, logger zerolog.Logger, settings *config.Settings) SmartCarService {
	return SmartCarService{
		baseURL:      "https://api.smartcar.com/v2.0/",
		DBS:          dbs,
		log:          logger,
		deviceDefSvc: NewDeviceDefinitionService(dbs, &logger, nil, settings), // not using nhtsa service or settings
	}
}

// ParseSmartCarYears parses out the years format in the smartcar document and returns an array of years
func ParseSmartCarYears(yearsPtr *string) ([]int, error) {
	if yearsPtr == nil || len(*yearsPtr) == 0 {
		return nil, errors.New("years string was nil")
	}
	years := *yearsPtr
	if len(years) > 4 {
		var rangeYears []int
		startYear := years[:4]
		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			return nil, errors.Errorf("could not parse start year from: %s", years)
		}
		endYear := time.Now().Year()
		if strings.Contains(years, "-") {
			eyStr := years[5:]
			endYear, err = strconv.Atoi(eyStr)
			if err != nil {
				return nil, errors.Errorf("could not parse end year from: %s", years)
			}
		}
		for y := startYearInt; y <= endYear; y++ {
			rangeYears = append(rangeYears, y)
		}
		return rangeYears, nil
	}
	y, err := strconv.Atoi(years)
	if err != nil {
		return nil, errors.Errorf("could not parse single year from: %s", years)
	}
	return []int{y}, nil
}

func (s *SmartCarService) GetOrCreateSmartCarIntegration(ctx context.Context) (string, error) {
	const (
		smartCarType  = "API"
		smartCarStyle = constants.IntegrationStyleWebhook
	)

	integration, err := s.deviceDefSvc.GetIntegrationByFilter(ctx, smartCarType, constants.SmartCarVendor, smartCarStyle)

	if err != nil {
		return "", errors.Wrap(err, "error fetching smart car integration from grpc")
	}

	if integration == nil {
		// create
		integration, err = s.deviceDefSvc.CreateIntegration(ctx, smartCarType, constants.SmartCarVendor, smartCarStyle)

		if err != nil {
			return "", errors.Wrap(err, "error insert smart car integration grpc")
		}
	}
	return integration.Id, nil
}

// GetSmartCarVehicleData gets all smartcar data on compatibility from their website
func GetSmartCarVehicleData() (*SmartCarCompatibilityData, error) {
	const url = "https://smartcar.com/page-data/product/compatible-vehicles/page-data.json"
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received a non 200 response from smart car page. status code: %d", res.StatusCode)
	}

	compatibleVehicles := SmartCarCompatibilityData{}
	err = json.NewDecoder(res.Body).Decode(&compatibleVehicles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal json from smart car")
	}
	return &compatibleVehicles, nil
}

type SmartCarCompatibilityData struct {
	ComponentChunkName string `json:"componentChunkName"`
	Path               string `json:"path"`
	Result             struct {
		Data struct {
			AllMakesTable struct {
				Edges []struct {
					Node struct {
						CompatibilityData map[string][]struct {
							Name    string `json:"name"`
							Headers []struct {
								Text    string  `json:"text"`
								Tooltip *string `json:"tooltip"`
							} `json:"headers"`
							Rows [][]struct {
								Color       *string `json:"color"`
								Subtext     *string `json:"subtext"`
								Text        *string `json:"text"`
								Type        *string `json:"type"`
								VehicleType *string `json:"vehicleType"`
							} `json:"rows"`
						} `json:"compatibilityData"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"allMakesTable"`
		} `json:"data"`
	} `json:"result"`
}

// IntegrationCapabilities gets stored on the association table btw a device_definition and the integrations, device_integrations
type IntegrationCapabilities struct {
	Location          bool `json:"location"`
	Odometer          bool `json:"odometer"`
	LockUnlock        bool `json:"lock_unlock"`
	EVBattery         bool `json:"ev_battery"`
	EVChargingStatus  bool `json:"ev_charging_status"`
	EVStartStopCharge bool `json:"ev_start_stop_charge"`
	FuelTank          bool `json:"fuel_tank"`
	TirePressure      bool `json:"tire_pressure"`
	EngineOilLife     bool `json:"engine_oil_life"`
	VehicleAttributes bool `json:"vehicle_attributes"`
	VIN               bool `json:"vin"`
}
