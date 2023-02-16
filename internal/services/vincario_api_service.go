package services

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/devices-api/internal/config"
	"reflect"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source vincario_api_service.go -destination mocks/vincario_api_service_mock.go -package mock_services
type VincarioAPIService interface {
	GetMarketValuation(vin string) (*VincarioMarketValueResponse, error)
}

type vincarioAPIService struct {
	settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
	log           *zerolog.Logger
}

func NewVincarioAPIService(settings *config.Settings, log *zerolog.Logger) VincarioAPIService {
	if settings.VincarioAPIURL == "" || settings.VincarioAPISecret == "" {
		panic("Vincario configuration not set")
	}
	hcwv, _ := shared.NewHTTPClientWrapper(settings.VincarioAPIURL, "", 10*time.Second, nil, false)

	return &vincarioAPIService{
		settings:      settings,
		httpClientVIN: hcwv,
		log:           log,
	}
}

func (va *vincarioAPIService) GetMarketValuation(vin string) (*VincarioMarketValueResponse, error) {
	id := "vehicle-market-value"

	urlPath := vincarioPathBuilder(vin, id, va.settings.VincarioAPIKey, va.settings.VincarioAPISecret)
	// url with api access
	resp, err := va.httpClientVIN.ExecuteRequest(urlPath, "GET", nil)
	if err != nil {
		return nil, err
	}

	// decode JSON from response body
	var data VincarioMarketValueResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &data, nil
}

func vincarioPathBuilder(vin, id, key, secret string) string {
	s := vin + "|" + id + "|" + key + "|" + secret

	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	controlSum := hex.EncodeToString(bs[0:5])

	return "/" + key + "/" + controlSum + "/" + id + "/" + vin + ".json"
}

func setStructPropertiesByMetadataKey(structPtr interface{}, key string, value interface{}) error {
	structValue := reflect.ValueOf(structPtr).Elem()
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		if fieldType.Tag.Get("key") == key {
			if !field.CanSet() {
				return fmt.Errorf("field %s is unexported and cannot be set", fieldType.Name)
			}

			fieldValue := reflect.ValueOf(value)

			if !fieldValue.Type().AssignableTo(field.Type()) {
				if fieldValue.Kind() == reflect.Float64 || fieldValue.Kind() == reflect.Float32 {
					f := fieldValue.Float()
					if field.Kind() == reflect.Int {
						field.Set(reflect.ValueOf(int(f)))
					}
				} else {
					return fmt.Errorf("value %v is not assignable to field %s of type %s", value, fieldType.Name, field.Type())
				}
			} else {
				field.Set(fieldValue)
			}
			return nil
		}
	}

	return fmt.Errorf("struct does not contain a field with key %s", key)
}

type VincarioMarketValueResponse struct {
	Vin     string `json:"vin"`
	Vehicle struct {
		Make                  string `json:"make"`
		Model                 string `json:"model"`
		ModelYear             int    `json:"model_year"`
		Transmission          string `json:"transmission"`
		EngineDisplacementCcm int    `json:"engine_displacement_ccm"`
		FuelTypePrimary       string `json:"fuel_type_primary"`
	} `json:"vehicle"`
	Period struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"period"`
	MarketPrice struct {
		PriceCount    int    `json:"price_count"`
		PriceCurrency string `json:"price_currency"`
		PriceBelow    int    `json:"price_below"`
		PriceMean     int    `json:"price_mean"`
		PriceAvg      int    `json:"price_avg"`
		PriceAbove    int    `json:"price_above"`
		PriceStdev    int    `json:"price_stdev"`
	} `json:"market_price"`
	MarketOdometer struct {
		OdometerCount int    `json:"odometer_count"`
		OdometerUnit  string `json:"odometer_unit"`
		OdometerBelow int    `json:"odometer_below"`
		OdometerMean  int    `json:"odometer_mean"`
		OdometerAvg   int    `json:"odometer_avg"`
		OdometerAbove int    `json:"odometer_above"`
		OdometerStdev int    `json:"odometer_stdev"`
	} `json:"market_odometer"`
}
