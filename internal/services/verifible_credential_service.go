package services

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/gofiber/fiber"
	"github.com/iden3/go-circuits"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/keccak256"
	"github.com/iden3/go-merkletree-sql"
	"github.com/iden3/go-merkletree-sql/db/memory"
	"github.com/iden3/iden3comm/protocol"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type IdentityMerkleTreeService struct {
	dbs *db.Store
}

// NewIdentityMerkleTrees generates a new merkle tree service
func NewIdentityMerkleTreeService(dbs *db.Store) *IdentityMerkleTreeService {
	return &IdentityMerkleTreeService{
		dbs: dbs,
	}
}

func (imts *IdentityMerkleTreeService) CreateIdentity(ctx context.Context, userDeviceID string) error {

	babyJubjubPrivKey := babyjub.NewRandPrivKey()
	babyJubjubPubKey := babyJubjubPrivKey.Public()

	revNonce := uint64(1)

	authClaim, _ := core.NewClaim(core.AuthSchemaHash,
		core.WithIndexDataInts(babyJubjubPubKey.X, babyJubjubPubKey.Y),
		core.WithRevocationNonce(revNonce))

	clt, err := merkletree.NewMerkleTree(ctx, memory.NewMemoryStorage(), 32)
	if err != nil {
		return err
	}
	ret, err := merkletree.NewMerkleTree(ctx, memory.NewMemoryStorage(), 32)
	if err != nil {
		return err
	}
	rot, err := merkletree.NewMerkleTree(ctx, memory.NewMemoryStorage(), 32)
	if err != nil {
		return err
	}

	hIndex, hValue, err := authClaim.HiHv()
	if err != nil {
		return err
	}

	clt.Add(ctx, hIndex, hValue)

	state, err := merkletree.HashElems(
		clt.Root().BigInt(),
		ret.Root().BigInt(),
		rot.Root().BigInt())
	if err != nil {
		return err
	}

	id, err := core.IdGenesisFromIdenState(core.TypeDefault, state.BigInt())
	if err != nil {
		return err
	}

	// store empty state-- should we add the vin claim immediately?
	vc := models.VerifiableCredential{
		UserDeviceID:   userDeviceID,
		ClaimsRoot:     null.BytesFrom([]byte(clt.Root().Hex())), // TODO: type
		RevocationRoot: null.BytesFrom([]byte(ret.Root().Hex())),
		RootOfRoots:    null.BytesFrom([]byte(rot.Root().Hex())),
		State:          null.BytesFrom([]byte(state.Hex())),
		ID:             null.BytesFrom(id.Bytes()),
	}
	// genesisTreeState := circuits.TreeState{
	// 	State:          state,
	// 	ClaimsRoot:     clt.Root(),
	// 	RevocationRoot: ret.Root(),
	// 	RootOfRoots:    rot.Root(),
	// }

	// rot.Add(ctx, clt.Root().BigInt(), big.NewInt(0))

	return vc.Insert(ctx, imts.dbs.DBS().Writer, boil.Infer())

}

func createSchemaHashHex(path_to_schema_json string) (string, error) {
	res, err := http.Get(path_to_schema_json)
	if err != nil {
		return "", err
	}

	if res.StatusCode >= 400 {
		return "", errors.New("invalid request")
	}

	schemaBytes, err := ioutil.ReadAll(res.Body)

	var sHash core.SchemaHash
	h := keccak256.Hash(schemaBytes, []byte("DIMO-Vin-KYC"))

	// copy in hash minus 16 characters?
	copy(sHash[:], h[len(h)-16:])

	sHashHex, err := sHash.MarshalText()
	if err != nil {
		return "", err
	}

	authClaim, _ := core.NewClaim(core.AuthSchemaHash)

	authClaim.HiHv()

	return string(sHashHex), nil
}

func ProofRequest(c *fiber.Ctx) error {
	var params RequestParameters
	err := c.QueryParser(params)
	if err != nil {
		return err
	}

	// Generate request for basic authentication
	var request protocol.AuthorizationRequestMessage

	request.ID = "7f38a193-0918-4a48-9fac-36adfdb8b542" // ids?
	request.ThreadID = "7f38a193-0918-4a48-9fac-36adfdb8b542"

	// Add request for a specific proof
	var mtpProofRequest protocol.ZeroKnowledgeProofRequest
	mtpProofRequest.ID = 1
	mtpProofRequest.CircuitID = string(circuits.AtomicQuerySigV2CircuitID)
	mtpProofRequest.Query = map[string]interface{}{
		"allowedIssuers": []string{"*"},
		"credentialSubject": map[string]interface{}{
			"vin": map[string]interface{}{
				"$eq": params.Vin,
			},
		},
		"context": "https://raw.githubusercontent.com/DIMO-Network/dimo-vin-kyc/main/dimo-vin-kyc.jsonld",
		"type":    "DIMO-Vin-KYC",
	}
	request.Body.Scope = append(request.Body.Scope, mtpProofRequest)

	return c.JSON(request)
}

func CreateVinClaim(c *fiber.Ctx) error {
	var params RequestParameters
	err := c.QueryParser(params)

	hash, err := createSchemaHashHex("https://raw.githubusercontent.com/DIMO-Network/dimo-vin-kyc/main/dimo-vin-kyc.json")
	if err != nil {
		return err
	}

	// set schema
	dimoVinKYC, err := core.NewSchemaHashFromHex(hash)
	if err != nil {
		return err
	}

	// set ID of the claim subject
	subjectID, err := core.IDFromInt(big.NewInt(1))
	if err != nil {
		return err
	}

	// create claim
	claim, err := core.NewClaim(dimoVinKYC, core.WithIndexID(subjectID), core.WithIndexDataBytes([]byte(params.Vin), []byte(params.TokenID)))
	if err != nil {
		return err
	}

	// transform claim from bytes array to json
	claimToMarshal, err := json.Marshal(claim)
	if err != nil {
		return err
	}

	return c.JSON(string(claimToMarshal))
}

type RequestParameters struct {
	Vin     string `json:"vin"`
	TokenID string `json:"token_id"`
	UserID  string `json:"user_id"`
}
