package registry

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"

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

func (ds *dcnService) GetExpiration(name string) (uint64, error) {
	const dcnAddress = "0xE9F4dfE02f895DC17E2e146e578873c9095bA293"
	nameHash := NameHash(name)
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
}`, dcnAddress, nameHash)
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

// NameHash returns a hash of the vehicle Name, eg. james.eth -> hash
func NameHash(name string) string {
	labels := strings.Split(name, ".")
	// todo this is not returning the expected address
	node := strings.Repeat("00", 32)

	if len(labels) > 0 {
		for i := len(labels) - 1; i >= 0; i-- {
			labelHash := crypto.Keccak256Hash([]byte(labels[i]))
			solHash := solsha3.SoliditySHA3(
				// types
				[]string{"uint256", "uint256"},
				// values
				[]interface{}{
					node, // note this is something that we overwrite each time
					labelHash.String(),
				},
			)

			node = hex.EncodeToString(solHash)
		}
	}

	return node
}
