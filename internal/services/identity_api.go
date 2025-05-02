package services

import (
	"encoding/json"
	"io"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrBadRequest = errors.New("bad request")

type identityAPIService struct {
	httpClient     shared.HTTPClientWrapper
	logger         zerolog.Logger
	identityAPIURL string
}

//go:generate mockgen -source identity_api.go -destination mocks/identity_api_mock.go -package mock_services
type IdentityAPI interface {
	GetManufacturer(slug string) (*Manufacturer, error)
	GetDefinition(definitionID string) (*DeviceDefinitionIdentity, error)
}

// NewIdentityAPIService creates a new instance of IdentityAPI, initializing it with the provided logger, settings, and HTTP client.
// httpClient is used for testing really
func NewIdentityAPIService(logger *zerolog.Logger, settings *config.Settings) IdentityAPI {
	h := map[string]string{}
	h["Content-Type"] = "application/json"
	httpClient, _ := shared.NewHTTPClientWrapper("", "", 10*time.Second, h, false) // ok to ignore err since only used for tor check

	return &identityAPIService{
		httpClient:     httpClient,
		logger:         *logger,
		identityAPIURL: settings.IdentiyAPIURL.String(),
	}
}

// GetManufacturer from identity-api by the name - must match exactly. Returns the token id and other on chain info
func (i *identityAPIService) GetManufacturer(name string) (*Manufacturer, error) {
	query := `{
  manufacturer(by: {name: "` + name + `"}) {
    	tokenId
    	name
    	tableId
    	owner
  	  }
	}`
	var wrapper struct {
		Data struct {
			Manufacturer Manufacturer `json:"manufacturer"`
		} `json:"data"`
	}
	err := i.fetchWithQuery(query, &wrapper)
	if err != nil {
		return nil, err
	}
	if wrapper.Data.Manufacturer.Name == "" {
		return nil, errors.Wrapf(ErrNotFound, "identity-api did not find manufacturer with name: %s", name)
	}
	return &wrapper.Data.Manufacturer, nil
}

func (i *identityAPIService) GetDefinition(definitionID string) (*DeviceDefinitionIdentity, error) {
	query := `{
  deviceDefinition(by: {id: "` + definitionID + `"}) {
    model,
    year,
    legacyId,
    manufacturer {
      tokenId
      name
    },
    imageURI,
    deviceType,
    attributes {
      name,
      value
    }
  }
}`
	var wrapper struct {
		Data struct {
			DeviceDefinition DeviceDefinitionIdentity `json:"deviceDefinition"`
		} `json:"data"`
	}
	err := i.fetchWithQuery(query, &wrapper)
	if err != nil {
		return nil, err
	}
	if wrapper.Data.DeviceDefinition.Model == "" {
		return nil, errors.Wrapf(ErrNotFound, "identity-api did not find device definition with name: %s", definitionID)
	}
	return &wrapper.Data.DeviceDefinition, nil
}

func (i *identityAPIService) fetchWithQuery(query string, result interface{}) error {
	// GraphQL request
	requestPayload := GraphQLRequest{Query: query}
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return err
	}

	// POST request
	res, err := i.httpClient.ExecuteRequest(i.identityAPIURL, "POST", payloadBytes)
	if err != nil {
		i.logger.Err(err).Str("func", "fetchWithQuery").Msgf("request payload: %s", string(payloadBytes))
		if _, ok := err.(shared.HTTPResponseError); !ok {
			return errors.Wrapf(err, "error calling identity api from url %s", i.identityAPIURL)
		}
	}
	defer res.Body.Close() // nolint

	if res.StatusCode == 404 {
		return ErrNotFound
	}
	if res.StatusCode == 400 {
		return ErrBadRequest
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrapf(err, "error reading response body from url %s", i.identityAPIURL)
	}

	if err := json.Unmarshal(bodyBytes, result); err != nil {
		return err
	}

	return nil
}

type Manufacturer struct {
	TokenID int    `json:"tokenId"`
	Name    string `json:"name"`
	TableID int    `json:"tableId"`
	Owner   string `json:"owner"`
}

type GraphQLRequest struct {
	Query string `json:"query"`
}

type DeviceDefinitionIdentity struct {
	Model        string `json:"model"`
	Year         int    `json:"year"`
	LegacyID     string `json:"legacyId"`
	Manufacturer struct {
		TokenID int    `json:"tokenId"`
		Name    string `json:"name"`
	} `json:"manufacturer"`
	ImageURI   string `json:"imageURI"`
	DeviceType string `json:"deviceType"`
	Attributes []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"attributes"`
}
