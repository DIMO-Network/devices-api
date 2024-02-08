package attestation

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/txaty/go-merkletree"
	mt "github.com/txaty/go-merkletree"
)

type onChainAttestation struct {
	data []byte
}

type leaf struct {
	VehicleTokenId         int
	DeviceDefinitionId     string
	VerifiableCredentialId string
	VCSignature            string
}

func (l *onChainAttestation) Serialize() ([]byte, error) {
	codeBytes, err := abiEncode([]interface{}{string(l.data)})
	if err != nil {
		return nil, err
	}

	return crypto.Keccak256(crypto.Keccak256(codeBytes)), nil
}

// TODO (ae): will this work for all types? need to check, might not want to infer
func abiEncode(values []interface{}) ([]byte, error) {
	argTypes := []reflect.Type{}
	for _, val := range values {
		argTypes = append(argTypes, reflect.TypeOf(val))
	}

	args := abi.Arguments{}
	for _, argType := range argTypes {
		abiType, err := abi.NewType(argType.Name(), "", nil)
		if err != nil {
			return nil, err
		}
		args = append(args, abi.Argument{Type: abiType})
	}

	return args.Pack(values...)
}

type attestor struct {
	config mt.Config
	pdb    db.Store
	client *ethclient.Client
}

func New(client *ethclient.Client, pdb db.Store) *attestor {
	return &attestor{
		config: mt.Config{
			SortSiblingPairs: true, // parameter for OpenZeppelin compatibility
			HashFunc: func(data []byte) ([]byte, error) {
				return crypto.Keccak256(data), nil
			},
			Mode: mt.ModeProofGenAndTreeBuild,
		},
		pdb:    pdb,
		client: client,
	}
}

func (m *attestor) Generate(ctx context.Context) (*merkletree.MerkleTree, error) {
	valid, err := models.VerifiableCredentials(models.VerifiableCredentialWhere.ExpirationDate.GT(time.Now().Add(-30*time.Hour*24))).All(ctx, m.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	blocks := []mt.DataBlock{}
	for n, vc := range valid {
		attestBts, err := json.Marshal(leaf{
			VehicleTokenId:         n,
			DeviceDefinitionId:     ksuid.New().String(),
			VerifiableCredentialId: vc.ClaimID,
			VCSignature:            gjson.GetBytes(vc.Credential.JSON, "proof.jws").String(),
		})
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, &onChainAttestation{data: attestBts})
	}

	return mt.New(&m.config, blocks)
}
