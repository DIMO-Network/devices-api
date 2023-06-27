package issuer

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/piprate/json-gold/ld"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
)

var secp256k1Prefix = []byte{0xe7, 0x01}
var period byte = '.'

type Config struct {
	PrivateKey        []byte
	ChainID           *big.Int
	VehicleNFTAddress common.Address
	DBS               db.Store
}

type Issuer struct {
	PrivateKey         *ecdsa.PrivateKey
	ChainID            *big.Int
	VehicleNFTAddress  common.Address
	DBS                db.Store
	KeyEnc             string
	IssuerDID          string
	VerificationMethod string
	LDProcessor        *ld.JsonLdProcessor
	LDOptions          *ld.JsonLdOptions
	log                *zerolog.Logger
}

func New(c Config, log *zerolog.Logger) (*Issuer, error) {
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
		log:                log,
	}, nil
}

func (i *Issuer) VIN(vin string, tokenID *big.Int) (id string, err error) {
	id = uuid.New().String()
	issuanceDate := time.Now().UTC().Format(time.RFC3339)
	expirationDate := time.Now().Add(time.Hour * 24 * 7).UTC()

	credential := map[string]any{
		"@context": []any{
			"https://www.w3.org/2018/credentials/v1",
			"https://schema.org/",
		},
		"id":             "urn:uuid:" + id,
		"type":           []any{"VerifiableCredential", "Vehicle"},
		"issuer":         i.IssuerDID,
		"issuanceDate":   issuanceDate,
		"expirationDate": expirationDate.Format(time.RFC3339),
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

	tx, err := i.DBS.DBS().Writer.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint

	vc := models.VerifiableCredential{
		ClaimID:        id,
		Credential:     signedBytes,
		ExpirationDate: expirationDate,
	}

	if err := vc.Insert(context.Background(), i.DBS.DBS().Writer, boil.Infer()); err != nil {
		return "", err
	}

	nft, err := models.VehicleNFTS(models.VehicleNFTWhere.Vin.EQ(vin)).One(context.Background(), tx)
	if err != nil {
		return "", err
	}

	nft.ClaimID = null.StringFrom(id)
	if _, err := nft.Update(context.Background(), i.DBS.DBS().Writer, boil.Whitelist(models.VehicleNFTColumns.ClaimID)); err != nil {
		return "", err
	}

	return id, tx.Commit()
}

func (i *Issuer) Handle(ctx context.Context, event *ADVinCredentialEvent) error {
	observedVIN, err := services.ExtractVIN(event.CloudEvent.Data)
	if err != nil {
		i.log.Info().Err(err).Msg("could not extract vin from payload")
		return err
	}

	logger := i.log.With().Str("device-address", event.Subject).Str("vin", observedVIN).Logger()
	logger.Info().Msg("got vin credentialer event")

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.EthereumAddress.EQ(
			null.BytesFrom(common.FromHex(event.Subject)),
		),
		qm.Load(models.AftermarketDeviceRels.VehicleToken),
	).One(ctx, i.DBS.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			i.log.Info().Err(err).Msg("no corresponding aftermarket device for address")
			return err
		}
		i.log.Info().Err(err).Msg("database failure retrieving aftermarket device")
		return err
	}

	vnft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(ad.VehicleTokenID),
		qm.Load(models.VehicleNFTRels.UserDevice),
		qm.Load(models.VehicleNFTRels.Claim, qm.OrderBy(models.VerifiableCredentialColumns.ExpirationDate+" DESC")),
	).One(ctx, i.DBS.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Err(err).Msg("no associated vehicle nft found for device")
			return err
		}
		i.log.Info().Err(err).Msg("database failure retrieving nft associated with device")
		return err
	}

	if !vnft.R.UserDevice.VinConfirmed {
		err := errors.New("invalid vin")
		logger.Err(err).Msg("vin associated with device not confirmed")
		return err
	}

	if vnft.R.UserDevice.VinIdentifier.String != observedVIN {
		// do we want to do anything here?
		logger.Info().Msg("observed vin does not match confirmed vin")
		return nil
	}

	if vnft.R.Claim != nil {
		if vnft.R.Claim.ExpirationDate.After(time.Now()) {
			logger.Info().Str("claimID", vnft.R.Claim.ClaimID).Msg("valid claim already exists")
			return nil
		}
	}

	tkn, ok := vnft.TokenID.Int64()
	if !ok {
		err := errors.New("invalid token id")
		logger.Err(err).Msg("unable to convert token id to int")
		return err
	}
	claimID, err := i.VIN(observedVIN, big.NewInt(tkn))
	logger.Info().Str("claim id", claimID).Msg("credential issued")
	return err
}

type ADVinCredentialEvent struct {
	shared.CloudEvent[json.RawMessage]
	Signature string `json:"signature"`
}

type VinEligibilityStatus struct {
	VIN                      string `json:"vin"`
	LatestEligibleRewardWeek int    `json:"latestEligibleRewardWeek"`
}

var issuanceStartTime = time.Date(2022, time.January, 31, 5, 0, 0, 0, time.UTC)
var weekDuration = 7 * 24 * time.Hour
