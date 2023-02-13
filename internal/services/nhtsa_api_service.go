package services

//go:generate mockgen -source nhtsa_api_service.go -destination mocks/nhtsa_api_service_mock.go

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type INHTSAService interface {
	DecodeVIN(vin string) (*NHTSADecodeVINResponse, error)
}

type NHTSAService struct {
	baseURL    string
	httpClient *http.Client
}

func NewNHTSAService() INHTSAService {
	return &NHTSAService{
		baseURL: "https://vpic.nhtsa.dot.gov/api/",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (ns *NHTSAService) DecodeVIN(vin string) (*NHTSADecodeVINResponse, error) {
	url := fmt.Sprintf("%s/vehicles/decodevinextended/%s?format=json", ns.baseURL, vin)
	res, err := ns.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received a non 200 response from nhtsa api. status code: %d", res.StatusCode)
	}

	decodedVin := NHTSADecodeVINResponse{}
	err = json.NewDecoder(res.Body).Decode(&decodedVin)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body from nhtsa api")
	}

	for _, r := range decodedVin.Results {
		if r.Variable == "Error Code" {
			if r.Value != "0" && r.Value != "" {
				return nil, fmt.Errorf("nhtsa api responded with error code %s", r.Value)
			}
			break
		}
	}

	return &decodedVin, nil
}

type NHTSAResult struct {
	Value      string  `json:"Value"`
	ValueID    *string `json:"ValueId"`
	Variable   string  `json:"Variable"`
	VariableID int     `json:"VariableId"`
}

type NHTSADecodeVINResponse struct {
	Count          int           `json:"Count"`
	Message        string        `json:"Message"`
	SearchCriteria string        `json:"SearchCriteria"`
	Results        []NHTSAResult `json:"Results"`
}

// LookupValue looks up value in nhtsa object, and uppercase the resulting value if not make or model
func (n *NHTSADecodeVINResponse) LookupValue(variableName string) string {
	for _, result := range n.Results {
		if result.Variable == variableName {
			if variableName == "Make" || variableName == "Model" {
				return result.Value
			}
			return strings.ToUpper(result.Value)
		}
	}
	return ""
}

const nhtsaElectrificationLevelVariableID = 126

func (n *NHTSADecodeVINResponse) DriveType() (PowertrainType, error) {
	for _, result := range n.Results {
		if result.VariableID == nhtsaElectrificationLevelVariableID {
			dt, err := nhtsaElectrificationLevelToPowertrain(result.ValueID)
			if err != nil {
				return "", err
			}

			return dt, nil
		}
	}
	return ICE, nil
}

// nhtsaElectrificationLevelToPowertrain converts the value id for the NHTSA electrification type
// field to our vehicle drive type. For values:
//
// https://vpic.nhtsa.dot.gov/api/vehicles/getvehiclevariablevalueslist/126?format=json
func nhtsaElectrificationLevelToPowertrain(el *string) (PowertrainType, error) {
	if el == nil {
		return ICE, nil
	}
	switch *el {
	case "1", "2", "9":
		return HEV, nil
	case "3":
		return PHEV, nil
	case "4":
		return BEV, nil
	case "5":
		return PHEV, nil
	default:
		return "", fmt.Errorf("unrecognized electrification level id: %d", el)
	}
}
