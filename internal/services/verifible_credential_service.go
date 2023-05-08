package services

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"time"

	ethC "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/mr-tron/base58"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/twinj/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type VerifiableCredentialService struct {
	dbs    db.Store
	ID     string
	Issuer string
}

var randReader io.Reader = rand.Reader

// NewVerifiableCredentialService generates vc service
func NewVerifiableCredentialService(dbs db.Store) (*VerifiableCredentialService, error) {
	return &VerifiableCredentialService{
		dbs:    dbs,
		ID:     "did:dimo_example:abfe13f712120431c276e12ecab",
		Issuer: "dimo.zone",
	}, nil
}

type Claim struct {
	ID      string
	Vin     string
	TokenID string
}

type Proof struct {
	TypeOfProof string
	Created     time.Time
	Creator     crypto.PublicKey
	Signature   struct {
		R, S *big.Int
	}
}

type Credential struct {
	Context           []string
	ID                string
	TypeOfCredential  []string
	Issuer            string
	IssuanceDate      time.Time
	CredentialSubject Claim
	Proof             Proof
	Status            Status
}

type Status struct {
	ID     string
	Status string
}

var secp256k1Prefix = []byte{0xe7, 0x01}

func (vcs *VerifiableCredentialService) generateKeysFromUserData(tokenID string) (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(ethC.S256(), randReader)
	if err != nil {
		return privateKey, err
	}
	v := secp256k1.CompressPubkey(privateKey.X, privateKey.Y)
	v = append(secp256k1Prefix, v...)
	keyEnc := "z" + base58.Encode(v)
	identity := "did:key:" + keyEnc
	vcData := models.VerifiableCredential{
		TokenID:  tokenID,
		X:        privateKey.X.Bytes(),
		Y:        privateKey.Y.Bytes(),
		D:        privateKey.D.Bytes(),
		Identity: identity,
	}
	return privateKey, vcData.Insert(context.Background(), vcs.dbs.DBS().Writer, boil.Infer())
}

//create credential
func (vcs *VerifiableCredentialService) createVinCredential(vin, tokenID string) (*Credential, error) {
	privKey, err := vcs.generateKeysFromUserData(tokenID)
	if err != nil {
		return nil, err
	}
	claim := Claim{
		ID:      "did:uuid:" + uuid.NewV1().String(),
		Vin:     vin,
		TokenID: tokenID,
	}
	credential := Credential{
		Context:           []string{"https://raw.githubusercontent.com/DIMO-Network/dimo-vin-kyc/main/dimo-vin-kyc.json", "https://raw.githubusercontent.com/DIMO-Network/dimo-vin-kyc/main/dimo-vin-kyc.jsonld"},
		ID:                vcs.ID,
		TypeOfCredential:  []string{"VerifiableCredential", "VinVerification"},
		Issuer:            vcs.Issuer,
		IssuanceDate:      time.Now(),
		CredentialSubject: claim,
	}
	r, s, err := ecdsa.Sign(randReader, privKey, []byte(fmt.Sprintf("%v", credential)))
	if err != nil {
		return nil, err
	}
	//create proof
	proof := Proof{
		TypeOfProof: "ed25519",
		Created:     time.Now(),
		Creator:     privKey.PublicKey,
		Signature: struct {
			R *big.Int
			S *big.Int
		}{R: r, S: s},
	}
	//add proof
	credential.Proof = proof
	return &credential, nil
}

func (vcs *VerifiableCredentialService) VerifyCredential(tokenID string, credential Credential) (bool, error) {
	vcUserData, err := models.VerifiableCredentials(models.VerifiableCredentialWhere.TokenID.EQ(tokenID)).One(context.Background(), vcs.dbs.DBS().Reader)
	if err != nil {
		return false, err
	}
	publicKey := ecdsa.PublicKey{
		Curve: ethC.S256(),
		X:     new(big.Int).SetBytes(vcUserData.X),
		Y:     new(big.Int).SetBytes(vcUserData.Y),
	}
	return ecdsa.Verify(&publicKey, []byte(fmt.Sprintf("%v", credential)), credential.Proof.Signature.R, credential.Proof.Signature.S), nil
}

// //create presentation
// func createPresentation(keyPair KeyPair, metadata PresenterMetadata, credential Credential) Presentation {
// 	presentation := Presentation{
// 		context:           metadata.context,
// 		typeOfPresentaion: metadata.typeOfPresentation,
// 		credential:        credential,
// 	}

// 	//create proof
// 	proofOfPresentaton := Proof{
// 		typeOfProof: "ed25519",
// 		created:     time.Now(),
// 		creator:     keyPair.publicKey,
// 		signature:   ed25519.Sign(keyPair.privateKey, []byte(fmt.Sprintf("%v", presentation))),
// 	}

// 	presentation.proof = proofOfPresentaton

// 	return presentation

// }

// //verify presentation
// func verifyPresentation(publicKey ed25519.PublicKey, presentation Presentation) bool {
// 	//verify if the public key is the same as in the credential
// 	if string(publicKey) != string(presentation.proof.creator) {
// 		return false
// 	}
// 	proofObj := presentation.proof
// 	presentation.proof = Proof{}
// 	//verify signature
// 	return ed25519.Verify(publicKey, []byte(fmt.Sprintf("%v", presentation)), proofObj.signature)
// }
