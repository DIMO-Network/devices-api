package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
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

func (p *vinDecodeCompareCmd) SetFlags(_ *flag.FlagSet) {
	//	p.targetTemplateID = f.String("target-template", "", "Filter device definitions to apply for where the template is this value. Good for moving only certain autopi's.")
	//f.BoolVar(&p.moveAllDevices, "move-all-devices", false, "move all devices in autopi to the template specified.")
}

func (p *vinDecodeCompareCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	ddSvc := services.NewDeviceDefinitionService(p.pdb.DBS, &p.logger, &p.settings)

	file, err := os.OpenFile("/tmp/vin-decode-compare.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close() // nolint
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to open file")
		return subcommands.ExitFailure
	}
	lastProcessedTokenID := getLatestTokenID(file)

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
	if lastProcessedTokenID > 0 {
		zero = big.NewInt(int64(lastProcessedTokenID))
	}
	markerTokenID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(zero, 0))

	for keepGoing {
		uds, err := models.UserDevices(models.UserDeviceWhere.TokenID.IsNotNull(), models.UserDeviceWhere.VinConfirmed.EQ(true),
			models.UserDeviceWhere.TokenID.GT(markerTokenID),
			qm.OrderBy(models.UserDeviceColumns.TokenID), qm.Limit(1000)).All(ctx, p.pdb.DBS().Reader)
		if err != nil {
			p.logger.Error().Err(err).Msg("failed to get user devices")
			return subcommands.ExitFailure
		}
		if len(uds) < 999 {
			keepGoing = false
		}
		tk, _ := markerTokenID.Uint64()
		p.logger.Info().Msgf("processing %d user devices, starting at token_id: %d", len(uds), tk)

		markerTokenID = uds[len(uds)-1].TokenID

		for _, ud := range uds {
			decodeVIN, err := ddSvc.DecodeVIN(ctx, ud.VinIdentifier.String, "", 0, ud.CountryCode.String)
			t, _ := ud.TokenID.Uint64()
			if err != nil {
				p.logger.Error().Uint64("token_id", t).Err(err).Msg("failed to decode vin")
				errW := writer.Write([]string{ud.TokenID.String(), ud.VinIdentifier.String, ud.DefinitionID, "failed"})
				if errW != nil {
					p.logger.Error().Err(errW).Msg("failed to write CSV row")
				}
				continue
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

func getLatestTokenID(file *os.File) uint64 {

	reader := csv.NewReader(file)
	// Read all rows from the CSV file
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV file:", err) //Error reading CSV file: read /tmp/vin-decode-compare.csv: bad file descriptor
		return 0
	}
	// Check if there are rows in the file
	if len(rows) == 0 {
		fmt.Println("CSV file is empty")
		return 0
	}
	// Get the last row
	lastRow := rows[len(rows)-1]

	// Get the value in the first column of the last row
	if len(lastRow) > 0 {
		tid := lastRow[0]
		value, _ := strconv.ParseUint(tid, 10, 64)
		return value
	}

	return 0
}
