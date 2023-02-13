package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
)

//go:generate mockgen -source drivly_api_service.go -destination mocks/drivly_api_service_mock.go
type DrivlyAPIService interface {
	GetVINInfo(vin string) (map[string]interface{}, error)
	GetVINPricing(vin string, reqData *ValuationRequestData) (map[string]any, error)

	GetOffersByVIN(vin string, reqData *ValuationRequestData) (map[string]interface{}, error)
	GetAutocheckByVIN(vin string) (map[string]interface{}, error)
	GetBuildByVIN(vin string) (map[string]interface{}, error)
	GetCargurusByVIN(vin string) (map[string]interface{}, error)
	GetCarvanaByVIN(vin string) (map[string]interface{}, error)
	GetCarmaxByVIN(vin string) (map[string]interface{}, error)
	GetCarstoryByVIN(vin string) (map[string]interface{}, error)
	GetEdmundsByVIN(vin string) (map[string]interface{}, error)
	GetTMVByVIN(vin string) (map[string]interface{}, error)
	GetKBBByVIN(vin string) (map[string]interface{}, error)
	GetVRoomByVIN(vin string) (map[string]interface{}, error)

	GetExtendedOffersByVIN(vin string) (*DrivlyVINSummary, error)
}

type drivlyAPIService struct {
	Settings        *config.Settings
	httpClientVIN   shared.HTTPClientWrapper
	httpClientOffer shared.HTTPClientWrapper
	dbs             func() *db.ReaderWriter
}

func NewDrivlyAPIService(settings *config.Settings, dbs func() *db.ReaderWriter) DrivlyAPIService {
	if settings.DrivlyVINAPIURL == "" || settings.DrivlyAPIKey == "" || settings.DrivlyOfferAPIURL == "" {
		panic("Drivly configuration not set")
	}
	h := map[string]string{"x-api-key": settings.DrivlyAPIKey}
	hcwv, _ := shared.NewHTTPClientWrapper(settings.DrivlyVINAPIURL, "", 10*time.Second, h, true)
	hcwo, _ := shared.NewHTTPClientWrapper(settings.DrivlyOfferAPIURL, "", 10*time.Second, h, true)

	return &drivlyAPIService{
		Settings:        settings,
		httpClientVIN:   hcwv,
		httpClientOffer: hcwo,
		dbs:             dbs,
	}
}

// GetVINInfo is the basic enriched VIN call, that is pretty standard now. Looks in multiple sources in their backend.
func (ds *drivlyAPIService) GetVINInfo(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientVIN, fmt.Sprintf("/api/%s/", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetVINPricing mileage is not sent if nil and zipcode is not sent if length is not equal to 5
func (ds *drivlyAPIService) GetVINPricing(vin string, reqData *ValuationRequestData) (map[string]any, error) {
	params := url.Values{}
	if reqData.Mileage != nil {
		params.Add("mileage", fmt.Sprint(int(*reqData.Mileage)))
	}
	if reqData.ZipCode != nil && len(*reqData.ZipCode) == 5 { // US 5 digit zip codes only
		params.Add("zipcode", *reqData.ZipCode)
	}
	res, err := executeAPI(ds.httpClientVIN, fmt.Sprintf("/api/%s/Pricing?"+params.Encode(), vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetOffersByVIN mileage is not sent if nil and zipcode is not sent if length is not equal to 5
func (ds *drivlyAPIService) GetOffersByVIN(vin string, reqData *ValuationRequestData) (map[string]interface{}, error) {
	params := url.Values{}
	if reqData.Mileage != nil {
		params.Add("mileage", fmt.Sprint(int(*reqData.Mileage)))
	}
	if reqData.ZipCode != nil && len(*reqData.ZipCode) == 5 { // US 5 digit zip codes only
		params.Add("zipcode", *reqData.ZipCode)
	}
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s?"+params.Encode(), vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetAutocheckByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/autocheck", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetBuildByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/build", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetCargurusByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/cargurus", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetCarmaxByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/carmax", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetCarstoryByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/carstory", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetCarvanaByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/carvana", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetEdmundsByVIN one of their raw data sources, the style_id they return may or not may be perfect.
func (ds *drivlyAPIService) GetEdmundsByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/edmunds", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetTMVByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/tmv", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetKBBByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/kbb", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ds *drivlyAPIService) GetVRoomByVIN(vin string) (map[string]interface{}, error) {
	res, err := executeAPI(ds.httpClientOffer, fmt.Sprintf("/api/vin/%s/tmv", vin))

	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetExtendedOffersByVIN calls all apis for offers and build info except the VIN info endpoint
func (ds *drivlyAPIService) GetExtendedOffersByVIN(vin string) (*DrivlyVINSummary, error) {
	result := new(DrivlyVINSummary)

	pricingRes, err := ds.GetVINPricing(vin, nil)
	if err != nil {
		return nil, err
	}

	offerRes, err := ds.GetOffersByVIN(vin, nil)
	if err != nil {
		return nil, err
	}

	autoCheckRes, err := ds.GetAutocheckByVIN(vin)
	if err != nil {
		return nil, err
	}

	buildRes, err := ds.GetBuildByVIN(vin)
	if err != nil {
		return nil, err
	}

	cargurusRes, err := ds.GetCargurusByVIN(vin)
	if err != nil {
		return nil, err
	}

	carmaxRes, err := ds.GetCarmaxByVIN(vin)
	if err != nil {
		return nil, err
	}

	carstoryRes, err := ds.GetCarstoryByVIN(vin)
	if err != nil {
		return nil, err
	}

	carvanaRes, err := ds.GetCarvanaByVIN(vin)
	if err != nil {
		return nil, err
	}

	edmundsRes, err := ds.GetEdmundsByVIN(vin)
	if err != nil {
		return nil, err
	}

	tmvRes, err := ds.GetTMVByVIN(vin)
	if err != nil {
		return nil, err
	}

	kbbRes, err := ds.GetKBBByVIN(vin)
	if err != nil {
		return nil, err
	}

	vroomRes, err := ds.GetVRoomByVIN(vin)
	if err != nil {
		return nil, err
	}

	result.Pricing = pricingRes
	result.Offers = offerRes
	result.AutoCheck = autoCheckRes
	result.Build = buildRes
	result.Cargurus = cargurusRes
	result.Carmax = carmaxRes
	result.Carstory = carstoryRes
	result.Carvana = carvanaRes
	result.Edmunds = edmundsRes
	result.TMV = tmvRes
	result.KBB = kbbRes
	result.VRoom = vroomRes

	return result, nil
}

type DrivlyVINSummary struct {
	Pricing   map[string]interface{}
	Offers    map[string]interface{}
	AutoCheck map[string]interface{}
	Build     map[string]interface{}
	Cargurus  map[string]interface{}
	Carvana   map[string]interface{}
	Carmax    map[string]interface{}
	Carstory  map[string]interface{}
	Edmunds   map[string]interface{}
	TMV       map[string]interface{}
	KBB       map[string]interface{}
	VRoom     map[string]interface{}
}

// todo add tests to this
func executeAPI(httpClient shared.HTTPClientWrapper, path string) (map[string]interface{}, error) {
	res, err := httpClient.ExecuteRequest(path, "GET", nil)
	if res == nil {
		if err != nil {
			return nil, errors.Wrapf(err, "error calling driv.ly api => %s", path)
		}
		return nil, fmt.Errorf("received error with no response when calling GET to %s", path)
	}

	if err != nil && res.StatusCode != 404 {
		return nil, errors.Wrapf(err, "error calling driv.ly api => %s", path)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var result map[string]interface{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing driv.ly api data => %s", path)
	}

	return result, nil
}
