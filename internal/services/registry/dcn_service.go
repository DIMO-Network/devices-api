package registry

import (
	"fmt"
	"io"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type DCNService interface {
	GetExpiration(name string) (uint64, error)
}

type dcnService struct {
	Settings   *config.Settings
	httpClient shared.HTTPClientWrapper
	//dbs        func() *db.ReaderWriter
}

func NewDcnService(settings *config.Settings) DCNService {
	client, _ := shared.NewHTTPClientWrapper(settings.DIMOContractAPIURL, "", 20*time.Second, nil, true)
	return &dcnService{
		Settings:   settings,
		httpClient: client,
	}
}

func (ds *dcnService) GetExpiration(dcnNodeAddress string) (uint64, error) {
	const dcnAddress = "0xE9F4dfE02f895DC17E2e146e578873c9095bA293"

	reqBody := fmt.Sprintf(`{
    "network": 137,
        "calls": [
          {
            "to": "%s",
            "abi": "DcnRegistry",
            "function": "expires",
            "inputs": {
              "node": "%s"
            }
          }
        ]    
}`, dcnAddress, dcnNodeAddress)
	res, err := ds.httpClient.ExecuteRequest("v1/multicall", "POST", []byte(reqBody))
	if err != nil {
		return 0, errors.Wrap(err, "could not get dcn records from multicall")
	}
	defer res.Body.Close() // nolint
	respBytes, _ := io.ReadAll(res.Body)

	result := gjson.GetBytes(respBytes, "outputs.0.outputs.0.expires_")
	if !result.Exists() {
		return 0, fmt.Errorf("unable to get expiration from multicall api")
	}
	return result.Uint(), nil
}
