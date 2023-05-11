package services

//go:generate mockgen -source verifible_credential_service.go -destination mocks/verifible_credential_service_mock.go

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethC "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/mr-tron/base58"
	"github.com/piprate/json-gold/ld"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/twinj/uuid"
)

type VerifiableCredentialService struct {
	dbs                db.Store
	Issuer             string
	PublicKey          ecdsa.PublicKey
	PrivateKey         ecdsa.PrivateKey
	KeyEnc             string
	VerificationMethod string
	NFTAddress         common.Address
	ChainID            *big.Int
}

var secp256k1Prefix = []byte{0xe7, 0x01}
var period byte = '.'

type VCService interface {
	CreateVinCredential(vin string, tokenID *big.Int) (string, error)
}

// NewVerifiableCredentialService generates vc service
func NewVerifiableCredentialService(dbs db.Store, settings *config.Settings) (VerifiableCredentialService, error) {
	xb, err := base64.RawURLEncoding.DecodeString(settings.DIMOKeyX)
	if err != nil {
		return VerifiableCredentialService{}, err
	}

	x := new(big.Int).SetBytes(xb)

	yb, err := base64.RawURLEncoding.DecodeString(settings.DIMOKeyY)
	if err != nil {
		return VerifiableCredentialService{}, err
	}
	y := new(big.Int).SetBytes(yb)

	db, err := base64.RawURLEncoding.DecodeString(settings.DIMOKeyD)
	if err != nil {
		return VerifiableCredentialService{}, err
	}
	d := new(big.Int).SetBytes(db)

	publicKey := ecdsa.PublicKey{
		Curve: ethC.S256(),
		X:     x,
		Y:     y,
	}
	privateKey := ecdsa.PrivateKey{
		PublicKey: publicKey,
		D:         d,
	}

	v := secp256k1.CompressPubkey(x, y)
	v = append(secp256k1Prefix, v...)

	keyEnc := "z" + base58.Encode(v)
	issuer := "did:key:" + keyEnc

	chainID := big.NewInt(settings.ChainID)
	nftAddr := common.HexToAddress(settings.NFTAddress)

	return VerifiableCredentialService{
		dbs:                dbs,
		Issuer:             issuer,
		PublicKey:          publicKey,
		PrivateKey:         privateKey,
		KeyEnc:             keyEnc,
		VerificationMethod: issuer + "#" + keyEnc,
		NFTAddress:         nftAddr,
		ChainID:            chainID,
	}, nil
}

//CreateVinCredential creates and signs credential using vin and tokenID
func (vcs VerifiableCredentialService) CreateVinCredential(vin string, tokenID *big.Int) (string, error) {
	issuanceDate := time.Now().UTC().Format(time.RFC3339)
	credentialID := "urn:uuid:" + uuid.NewV1().String()
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/n-quads"
	options.Algorithm = ld.AlgorithmURDNA2015

	credential := map[string]any{
		"@context": []any{
			"https://www.w3.org/2018/credentials/v1",
			"https://schema.org/",
		},
		"id":           credentialID,
		"type":         []any{"VerifiableCredential", "Vehicle"},
		"issuer":       vcs.Issuer,
		"issuanceDate": issuanceDate,
		"credentialSubject": map[string]any{
			"id":                          fmt.Sprintf("did:nft:%d_erc721:%s_%d", vcs.ChainID, hexutil.Encode(vcs.NFTAddress[:]), tokenID),
			"vehicleIdentificationNumber": vin,
		},
	}

	proof := map[string]any{
		"@context":           "https://www.w3.org/2018/credentials/v1",
		"type":               "EcdsaSecp256k1Signature2019",
		"proofPurpose":       "assertionMethod",
		"verificationMethod": vcs.VerificationMethod,
		"created":            issuanceDate,
	}

	docNorm, err := proc.Normalize(credential, options)
	if err != nil {
		return credentialID, err
	}
	docNormStr := docNorm.(string)

	preProofNorm, err := proc.Normalize(proof, options)
	if err != nil {
		return credentialID, err
	}

	preProofNormStr := preProofNorm.(string)
	preProofDigest := sha256.Sum256([]byte(preProofNormStr))
	docDigest := sha256.Sum256([]byte(docNormStr))
	digest := append(preProofDigest[:], docDigest[:]...)
	header := map[string]any{
		"alg":  "ES256K",
		"crit": []any{"b64"},
		"b64":  false,
	}

	hb, _ := json.Marshal(header)
	hb64 := make([]byte, base64.RawURLEncoding.EncodedLen(len(hb)))
	base64.RawURLEncoding.Encode(hb64, hb)

	jw2 := append(hb64, period)
	jw2 = append(jw2, digest...)
	enddig := sha256.Sum256(jw2)

	r, s, err := ecdsa.Sign(rand.Reader, &vcs.PrivateKey, enddig[:])
	if err != nil {
		return credentialID, err
	}

	outb := make([]byte, 64)
	r.FillBytes(outb[:32])
	s.FillBytes(outb[32:])

	proof["jws"] = string(hb64) + ".." + base64.RawURLEncoding.EncodeToString(outb)
	credential["proof"] = proof
	bj, err := json.MarshalIndent(credential, "", "  ")
	if err != nil {
		return credentialID, err
	}

	nft, err := models.VehicleNFTS(models.VehicleNFTWhere.Vin.EQ(vin)).One(context.Background(), vcs.dbs.DBS().Reader)
	if err != nil {
		return credentialID, err
	}
	nft.ClaimID = null.StringFrom(credentialID)
	_, err = nft.Update(context.Background(), vcs.dbs.DBS().Writer, boil.Infer())
	if err != nil {
		return credentialID, err
	}

	vc := models.VerifiableCredential{
		ClaimID: credentialID,
		Proof:   bj,
	}

	return credentialID, vc.Insert(context.Background(), vcs.dbs.DBS().Writer, boil.Infer())
}
