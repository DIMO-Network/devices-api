package services

import (
	"fmt"
	"io"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type DCNService interface {
	GetRecordExpiration(dcnAddress, name string) (uint64, error)
}

type dcnService struct {
	Settings   *config.Settings
	httpClient shared.HTTPClientWrapper
	//dbs        func() *db.ReaderWriter
}

const dimoWeb3URL = "https://multicall.dimo/"

func NewDcnService(settings *config.Settings) DCNService {
	client, _ := shared.NewHTTPClientWrapper(dimoWeb3URL, "", 20*time.Second, nil, true)
	return &dcnService{
		Settings:   settings,
		httpClient: client,
	}
}

func (ds *dcnService) GetRecordExpiration(dcnAddress, name string) (uint64, error) {
	nameHash := NameHash(name)
	reqBody := fmt.Sprintf(`{
    "network": 137,
        "calls": [
          {
            "to": "%s",
            "abi": "DcnRegistry",
            "function": "records",
            "inputs": {
              "node": "%s"
            }
          },
          {
            "to": "%s",
            "abi": "DcnRegistry",
            "function": "resolver",
            "inputs": {
              "node": "%s"
            }
          }
        ]    
}`, dcnAddress, nameHash, dcnAddress, nameHash)
	res, err := ds.httpClient.ExecuteRequest("v1/multicall", "POST", []byte(reqBody))
	if err != nil {
		return 0, errors.Wrap(err, "could not get dcn records from multicall")
	}
	defer res.Body.Close() // nolint
	respBytes, _ := io.ReadAll(res.Body)

	result := gjson.GetBytes(respBytes, "outputs.0.outputs.1.expires")
	if !result.Exists() {
		return 0, fmt.Errorf("unable to get expiration from multicall api")
	}
	return result.Uint(), nil
}

// NameHash Name of what?
func NameHash(name string) common.Hash {
	//node := strings.Repeat("00", 32)
	labelHash := crypto.Keccak256Hash([]byte(name))

	//node = ethers.utils.solidityKeccak256(
	//["uint256", "uint256"],
	//[node, labelHash]
	//);
	// todo, this is incomplete - function _namehash in dcn.ts, token-information
	return labelHash
}
