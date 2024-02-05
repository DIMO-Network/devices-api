package attestation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/crypto"
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
	// return crypto.Keccak256(crypto.Keccak256(l.data)), nil
	return l.data, nil
}

type attestor struct { // idk if this makes any sense.. bc witnesses attest to things? seems too cheeky
	config mt.Config
	pdb    db.Store
}

func New(pdb db.Store) *attestor {
	return &attestor{
		config: mt.Config{
			SortSiblingPairs: true,
			HashFunc: func(data []byte) ([]byte, error) {
				return crypto.Keccak256(crypto.Keccak256(data)), nil
			},
			Mode: mt.ModeProofGenAndTreeBuild,
		},
		pdb: pdb,
	}
}

func (m *attestor) Generate(ctx context.Context) (*merkletree.MerkleTree, error) {
	valid, err := models.VerifiableCredentials(models.VerifiableCredentialWhere.ExpirationDate.GT(time.Now().Add(-7*time.Hour*24))).All(ctx, m.pdb.DBS().Reader)
	if err != nil {
		return nil, err
	}

	// sorting here?

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
