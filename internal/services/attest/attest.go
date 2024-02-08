package attest

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tidwall/gjson"
	"github.com/txaty/go-merkletree"
	mt "github.com/txaty/go-merkletree"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const validCredentialRange = time.Hour * 24 * 7

type leaf struct {
	data []byte
}

func (l *leaf) Serialize() ([]byte, error) {
	// since OpenZep will sort bytes before hashing
	// abi encoding and serialization is performed before
	// we pass any data to the merkle tree func

	// codeBytes, err := abiEncode([]interface{}{string(l.data)})
	// if err != nil {
	// 	return nil, err
	// }

	// return crypto.Keccak256(crypto.Keccak256(codeBytes)), nil
	return l.data, nil
}

func abiEncode(types []string, values []interface{}) ([]byte, error) {
	if len(types) != len(values) {
		return nil, fmt.Errorf("type array  and value array must be same length")

	}
	args := abi.Arguments{}
	for _, argType := range types {
		abiType, err := abi.NewType(argType, "", nil)
		if err != nil {
			return nil, err
		}
		args = append(args, abi.Argument{Type: abiType})
	}

	return args.Pack(values...)
}

type attestor struct {
	Config mt.Config
	pdb    db.Store
	client *ethclient.Client
}

func New(client *ethclient.Client, pdb db.Store) *attestor {
	return &attestor{
		Config: mt.Config{
			SortSiblingPairs: true, // parameter for OpenZeppelin compatibility
			HashFunc: func(data []byte) ([]byte, error) {
				return crypto.Keccak256(data), nil
			},
			Mode:               mt.ModeProofGenAndTreeBuild,
			DisableLeafHashing: true, // also required for OpenZeppelin compatilibity
		},
		pdb:    pdb,
		client: client,
	}
}

func (m *attestor) Generate(ctx context.Context) (*merkletree.MerkleTree, error) {
	attestations, err := models.UserDevices(
		qm.Load(
			qm.Rels(
				models.UserDeviceRels.VehicleNFT,
				models.VehicleNFTRels.Claim,
			),
		),
	).All(ctx, m.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	encodedLeaves := [][]byte{}
	for _, att := range attestations {
		if att.R.VehicleNFT == nil || att.R.VehicleNFT.Claim == nil {
			// log warning and continue
		}
		tokenID, ok := att.R.VehicleNFT.TokenID.Uint64()
		if !ok {
			// log warning and continue
		}

		attestation := []interface{}{
			tokenID,
			att.DeviceDefinitionID,
			att.R.VehicleNFT.R.Claim.ClaimID,
			gjson.GetBytes(att.R.VehicleNFT.R.Claim.Credential.JSON, "proof.jws").String(),
		}
		codeBytes, err := abiEncode([]string{"uint64", "string", "string", "string"}, attestation)
		if err != nil {
			panic(err)
		}
		encodedLeaves = append(encodedLeaves, crypto.Keccak256(crypto.Keccak256(codeBytes)))

	}

	// sort values for OpenZeppelin compatibility
	sort.Slice(encodedLeaves, func(i, j int) bool {
		return string(encodedLeaves[i]) < string(encodedLeaves[j])
	})

	blocks := []mt.DataBlock{}
	for _, vc := range encodedLeaves {
		blocks = append(blocks, &leaf{data: vc})
	}

	return mt.New(&m.Config, blocks)
}

func (m *attestor) VerifyAttestation(root, node string, proof []string) (bool, error) {
	dataBlock := &leaf{data: common.Hex2Bytes(node)}
	p := mt.Proof{}
	for _, prf := range proof {
		p.Siblings = append(p.Siblings, common.Hex2Bytes(prf))
	}
	r := common.Hex2Bytes(root)
	return mt.Verify(dataBlock, &p, r, &m.Config)

}
