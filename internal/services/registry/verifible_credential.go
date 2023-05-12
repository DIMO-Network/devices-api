package registry

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

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/google/uuid"
	"github.com/piprate/json-gold/ld"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
)

type issuerCredentials struct {
	Issuer             string
	PrivateKey         *ecdsa.PrivateKey
	KeyEnc             string
	VerificationMethod string
	NFTAddress         common.Address
	ChainID            *big.Int
}

var secp256k1Prefix = []byte{0xe7, 0x01}
var period byte = '.'

func IssuerCredentials(settings *config.Settings) (issuerCredentials, error) {
	db, err := base64.RawURLEncoding.DecodeString(settings.DIMOKeyD)
	if err != nil {
		return issuerCredentials{}, err
	}
	privateKey, err := crypto.ToECDSA(db)
	if err != nil {
		return issuerCredentials{}, err
	}

	v := secp256k1.CompressPubkey(privateKey.X, privateKey.Y)
	v = append(secp256k1Prefix, v...)

	keyEnc := "z" + base58.Encode(v)
	issuer := "did:key:" + keyEnc

	return issuerCredentials{
		Issuer:             issuer,
		PrivateKey:         privateKey,
		KeyEnc:             keyEnc,
		VerificationMethod: issuer + "#" + keyEnc,
		NFTAddress:         common.HexToAddress(settings.DIMORegistryAddr),
		ChainID:            big.NewInt(settings.DIMORegistryChainID),
	}, nil

}

//CreateVinCredential creates and signs credential using vin and tokenID
func (p *proc) CreateVinCredential(vin string, tokenID *big.Int) (string, error) {
	issuanceDate := time.Now().UTC().Format(time.RFC3339)
	credentialID := "urn:uuid:" + uuid.New().String()
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
		"issuer":       p.issuer.Issuer,
		"issuanceDate": issuanceDate,
		"credentialSubject": map[string]any{
			"id":                          fmt.Sprintf("did:nft:%d_erc721:%s_%d", p.issuer.ChainID, hexutil.Encode(p.issuer.NFTAddress[:]), tokenID),
			"vehicleIdentificationNumber": vin,
		},
	}

	proof := map[string]any{
		"@context":           "https://www.w3.org/2018/credentials/v1",
		"type":               "EcdsaSecp256k1Signature2019",
		"proofPurpose":       "assertionMethod",
		"verificationMethod": p.issuer.VerificationMethod,
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

	hb, err := json.Marshal(header)
	if err != nil {
		return credentialID, err
	}
	hb64 := make([]byte, base64.RawURLEncoding.EncodedLen(len(hb)))
	base64.RawURLEncoding.Encode(hb64, hb)

	jw2 := append(hb64, period)
	jw2 = append(jw2, digest...)
	enddig := sha256.Sum256(jw2)

	r, s, err := ecdsa.Sign(rand.Reader, p.issuer.PrivateKey, enddig[:])
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

	nft, err := models.VehicleNFTS(models.VehicleNFTWhere.Vin.EQ(vin)).One(context.Background(), p.DB().Reader)
	if err != nil {
		return credentialID, err
	}
	nft.ClaimID = null.StringFrom(credentialID)
	_, err = nft.Update(context.Background(), p.DB().Writer, boil.Infer())
	if err != nil {
		return credentialID, err
	}

	vc := models.VerifiableCredential{
		ClaimID:    credentialID,
		Credential: bj,
	}

	return credentialID, vc.Insert(context.Background(), p.DB().Writer, boil.Infer())
}
