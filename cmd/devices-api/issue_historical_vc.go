package main

import (
	"context"
	"encoding/base64"
	"flag"
	"math/big"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services/issuer"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type issueHistorialCredentialCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*issueHistorialCredentialCmd) Name() string     { return "issue-historical-vc" }
func (*issueHistorialCredentialCmd) Synopsis() string { return "issue-historical-vc args to stdout." }
func (*issueHistorialCredentialCmd) Usage() string {
	return `issue-historical-vc:
	issue verifiable credential to devices minted before vc creation was deployed.
  `
}

// nolint
func (p *issueHistorialCredentialCmd) SetFlags(f *flag.FlagSet) {

}

func (hcmd *issueHistorialCredentialCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	pk, err := base64.RawURLEncoding.DecodeString(hcmd.settings.IssuerPrivateKey)
	if err != nil {
		hcmd.logger.Fatal().Err(err).Msg("Couldn't parse issuer private key.")
	}

	issuer, err := issuer.New(
		issuer.Config{
			PrivateKey:        pk,
			ChainID:           big.NewInt(hcmd.settings.DIMORegistryChainID),
			VehicleNFTAddress: common.HexToAddress(hcmd.settings.VehicleNFTAddress),
			DBS:               hcmd.pdb,
		},
	)
	if err != nil {
		hcmd.logger.Err(err).Msg("unable to instantiate issuer")
	}

	claims := make([]string, 0)
	vcs, err := models.VerifiableCredentials().All(context.Background(), hcmd.pdb.DBS().Reader)
	if err != nil {
		hcmd.logger.Err(err).Msg("unable to query verifiable credential table")
	}

	for _, vc := range vcs {
		claims = append(claims, vc.ClaimID)
	}

	mintedDevices, err := models.VehicleNFTS(
		models.VehicleNFTWhere.ClaimID.NIN(claims),
	).All(context.Background(), hcmd.pdb.DBS().Reader)
	if err != nil {
		hcmd.logger.Err(err).Msg("unable to query vehicle nft table")
	}

	for _, device := range mintedDevices {
		tid := device.TokenID.Big.Int(new(big.Int))
		vcID, err := issuer.VIN(device.Vin, tid)
		if err != nil {
			hcmd.logger.Err(err).Str("error spot", vcID).Str("vin", device.Vin).Int64("tokenID", tid.Int64()).Msgf("unable to issue vc")
			continue
		}

		hcmd.logger.Info().Str("vin", device.Vin).Int64("tokenID", tid.Int64()).Str("claimID", vcID).Msgf("vc issued")
	}

	return subcommands.ExitSuccess
}
