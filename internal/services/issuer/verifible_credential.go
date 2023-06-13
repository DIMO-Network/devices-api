package issuer

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
	"github.com/google/uuid"
	"github.com/piprate/json-gold/ld"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
)

var secp256k1Prefix = []byte{0xe7, 0x01}
var period byte = '.'

type Config struct {
	PrivateKey        []byte
	ChainID           *big.Int
	VehicleNFTAddress common.Address
	DBS               func() *db.ReaderWriter
}

type Issuer struct {
	PrivateKey         *ecdsa.PrivateKey
	ChainID            *big.Int
	VehicleNFTAddress  common.Address
	DBS                func() *db.ReaderWriter
	KeyEnc             string
	IssuerDID          string
	VerificationMethod string
	LDProcessor        *ld.JsonLdProcessor
	LDOptions          *ld.JsonLdOptions
}

func New(c Config) (*Issuer, error) {
	privateKey, err := crypto.ToECDSA(c.PrivateKey)
	if err != nil {
		return nil, err
	}

	keyEnc := "z" + base58.Encode(append(secp256k1Prefix, crypto.CompressPubkey(&privateKey.PublicKey)...))
	issuer := "did:key:" + keyEnc
	verificationMethod := issuer + "#" + keyEnc

	ldProc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Format = "application/n-quads"
	options.Algorithm = ld.AlgorithmURDNA2015

	return &Issuer{
		PrivateKey:         privateKey,
		KeyEnc:             keyEnc,
		ChainID:            c.ChainID,
		VehicleNFTAddress:  c.VehicleNFTAddress,
		DBS:                c.DBS,
		IssuerDID:          issuer,
		VerificationMethod: verificationMethod,
		LDProcessor:        ldProc,
		LDOptions:          options,
	}, nil
}

func (i *Issuer) VIN(vin string, tokenID *big.Int) (id string, err error) {
	id = uuid.New().String()
	issuanceDate := time.Now().UTC().Format(time.RFC3339)
	expirationDate := time.Now().Add(time.Hour * 24 * 7).UTC().Format(time.RFC3339)

	credential := map[string]any{
		"@context": []any{
			"https://www.w3.org/2018/credentials/v1",
			"https://schema.org/",
		},
		"id":             "urn:uuid:" + id,
		"type":           []any{"VerifiableCredential", "Vehicle"},
		"issuer":         i.IssuerDID,
		"issuanceDate":   issuanceDate,
		"expirationDate": expirationDate,
		"credentialSubject": map[string]any{
			"id":                          fmt.Sprintf("did:nft:%d_erc721:%s_%d", i.ChainID, hexutil.Encode(i.VehicleNFTAddress.Bytes()), tokenID),
			"vehicleIdentificationNumber": vin,
		},
	}

	proof := map[string]any{
		"@context":           "https://www.w3.org/2018/credentials/v1",
		"type":               "EcdsaSecp256k1Signature2019",
		"proofPurpose":       "assertionMethod",
		"verificationMethod": i.VerificationMethod,
		"created":            issuanceDate,
	}

	docNorm, err := i.LDProcessor.Normalize(credential, i.LDOptions)
	if err != nil {
		return "", err
	}
	docNormStr := docNorm.(string)

	preProofNorm, err := i.LDProcessor.Normalize(proof, i.LDOptions)
	if err != nil {
		return "", err
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
		return "", err
	}

	hb64 := make([]byte, base64.RawURLEncoding.EncodedLen(len(hb)))
	base64.RawURLEncoding.Encode(hb64, hb)

	jw2 := append(hb64, period)
	jw2 = append(jw2, digest...)
	enddig := sha256.Sum256(jw2)

	r, s, err := ecdsa.Sign(rand.Reader, i.PrivateKey, enddig[:])
	if err != nil {
		return "", err
	}

	outb := make([]byte, 64)
	r.FillBytes(outb[:32])
	s.FillBytes(outb[32:])

	proof["jws"] = string(hb64) + ".." + base64.RawURLEncoding.EncodeToString(outb)
	credential["proof"] = proof

	signedBytes, err := json.Marshal(credential)
	if err != nil {
		return "", err
	}

	tx, err := i.DBS().Writer.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint

	vc := models.VerifiableCredential{
		ClaimID:    id,
		Credential: signedBytes,
	}

	if err := vc.Insert(context.Background(), i.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	nft, err := models.VehicleNFTS(models.VehicleNFTWhere.Vin.EQ(vin)).One(context.Background(), tx)
	if err != nil {
		return "", err
	}

	nft.ClaimID = null.StringFrom(id)
	if _, err := nft.Update(context.Background(), i.DBS().Writer, boil.Whitelist(models.VehicleNFTColumns.ClaimID)); err != nil {
		return "", err
	}

	return id, tx.Commit()
}
