package main

import (
	"context"
	"encoding/csv"
	"flag"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"math/big"
	"os"
)

type vinDecodeCompareCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*vinDecodeCompareCmd) Name() string { return "vin-decode-compare" }
func (*vinDecodeCompareCmd) Synopsis() string {
	return "iterate through all minted vehicles and decode their VIN comparing resulting definition id"
}
func (*vinDecodeCompareCmd) Usage() string {
	return `vin-decode-compare`
}

func (p *vinDecodeCompareCmd) SetFlags(f *flag.FlagSet) {
	//	p.targetTemplateID = f.String("target-template", "", "Filter device definitions to apply for where the template is this value. Good for moving only certain autopi's.")
	//f.BoolVar(&p.moveAllDevices, "move-all-devices", false, "move all devices in autopi to the template specified.")
}

func (p *vinDecodeCompareCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	ddSvc := services.NewDeviceDefinitionService(p.pdb.DBS, &p.logger, &p.settings)

	file, err := os.OpenFile("vin-decode-compare.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close() // nolint
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to open file")
		return subcommands.ExitFailure
	}
	// todo: check if file exists and what token left off on

	writer := csv.NewWriter(file)
	defer writer.Flush()
	// Write header
	err = writer.Write([]string{"Token ID", "VIN", "Original Definition ID", "New Definition ID"})
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to write CSV header")
		return subcommands.ExitFailure
	}

	keepGoing := true
	zero := big.NewInt(int64(0))
	markerTokenId := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(zero, 0))

	for keepGoing {
		uds, err := models.UserDevices(models.UserDeviceWhere.TokenID.IsNotNull(), models.UserDeviceWhere.VinConfirmed.EQ(true),
			models.UserDeviceWhere.TokenID.GT(markerTokenId),
			qm.OrderBy(models.UserDeviceColumns.TokenID), qm.Limit(1000)).All(ctx, p.pdb.DBS().Reader)
		if err != nil {
			p.logger.Error().Err(err).Msg("failed to get user devices")
			return subcommands.ExitFailure
		}
		if len(uds) < 999 {
			keepGoing = false
		}
		tk, _ := markerTokenId.Uint64()
		p.logger.Info().Msgf("processing %d user devices, starting at token_id: %d", len(uds), tk)

		markerTokenId = uds[len(uds)-1].TokenID

		for _, ud := range uds {
			decodeVIN, err := ddSvc.DecodeVIN(ctx, ud.VinIdentifier.String, "", 0, ud.CountryCode.String)
			t, _ := ud.TokenID.Uint64()
			if err != nil {
				p.logger.Error().Uint64("token_id", t).Err(err).Msg("failed to decode vin")
			}
			if ud.DefinitionID != decodeVIN.DefinitionId {
				// log and csv
				p.logger.Info().Uint64("token_id", t).Str("vin", ud.VinIdentifier.String).
					Msgf("vin decode mismatch. original: %s, new: %s", ud.DefinitionID, decodeVIN.DefinitionId)

				errW := writer.Write([]string{ud.TokenID.String(), ud.VinIdentifier.String, ud.DefinitionID, decodeVIN.DefinitionId})
				if errW != nil {
					p.logger.Error().Err(errW).Msg("failed to write CSV row")
				}
			}
		}
	}

	return subcommands.ExitSuccess
}
