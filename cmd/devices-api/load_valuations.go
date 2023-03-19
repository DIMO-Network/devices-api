package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
)

type loadValuationsCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*loadValuationsCmd) Name() string     { return "valuations-pull" }
func (*loadValuationsCmd) Synopsis() string { return "valuations-pull args to stdout." }
func (*loadValuationsCmd) Usage() string {
	return `valuations-pull:
	valuations-pull args.
  `
}

// nolint
func (p *loadValuationsCmd) SetFlags(f *flag.FlagSet) {

}

func (p *loadValuationsCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.logger.Info().Msgf("Pull VIN info, valuations and pricing from driv.ly for USA and valuations from Vincario for EUR")
	setAll := false
	wmi := ""
	if len(os.Args) > 2 {
		setAll = os.Args[2] == "--set-all"
		// parse out vin WMI code to filter on
		for i, a := range os.Args {
			if a == "--wmi" {
				wmi = os.Args[i+1]
				break
			}
		}
	}
	err := loadValuations(ctx, &p.logger, &p.settings, setAll, wmi, p.pdb)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("error trying to pull valuations")
	}

	return subcommands.ExitSuccess
}

// loadValuations iterates over user_devices with vin verified and tries pulling data from drivly in USA & CAN and vincario for rest of world
func loadValuations(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, forceSetAll bool, wmi string, pdb db.Store) error {
	// get all devices from DB.
	all, err := models.UserDevices(
		models.UserDeviceWhere.VinConfirmed.EQ(true)).
		All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	if len(wmi) == 3 {
		wmi = strings.ToUpper(wmi)
		logger.Info().Msgf("WMI filter set: %s", wmi)
		filtered := models.UserDeviceSlice{}
		for _, device := range all {
			if len(device.VinIdentifier.String) > 3 && device.VinIdentifier.String[:3] == wmi {
				filtered = append(filtered, device)
			}
		}
		all = filtered
	}
	logger.Info().Msgf("processing %d user_devices with verified VINs in ALL regions", len(all))

	deviceDefinitionSvc := services.NewDeviceDefinitionService(pdb.DBS, logger, nil, settings)
	statsAggr := map[services.DataPullStatusEnum]int{}
	for _, ud := range all {
		if ud.CountryCode.String == "USA" || ud.CountryCode.String == "CAN" || ud.CountryCode.String == "MEX" {
			status, err := deviceDefinitionSvc.PullDrivlyData(ctx, ud.ID, ud.DeviceDefinitionID, ud.VinIdentifier.String, forceSetAll)
			if err != nil {
				logger.Err(err).Str("vin", ud.VinIdentifier.String).Msg("error pulling drivly data")
			} else {
				logger.Info().Msgf("Drivly   %s vin: %s, country: %s", status, ud.VinIdentifier.String, ud.CountryCode.String)
			}
			statsAggr[status]++
		} else {
			status, err := deviceDefinitionSvc.PullVincarioValuation(ctx, ud.ID, ud.DeviceDefinitionID, ud.VinIdentifier.String)
			if err != nil {
				logger.Err(err).Str("vin", ud.VinIdentifier.String).Msg("error pulling vincario data")
			} else {
				logger.Info().Msgf("Vincario %s vin: %s, country: %s", status, ud.VinIdentifier.String, ud.CountryCode.String)
			}
			statsAggr[status]++
		}
	}
	fmt.Println("-------------------RUN SUMMARY--------------------------")
	// colorize each result
	fmt.Printf("Total VINs processed: %d \n", len(all))
	fmt.Printf("New Drivly Pulls (vin + valuations): %d \n", statsAggr[services.PulledInfoAndValuationStatus])
	fmt.Printf("Pulled New Pricing & Offers: %d \n", statsAggr[services.PulledValuationDrivlyStatus])
	fmt.Printf("Skipped VIN due to biz logic: %d \n", statsAggr[services.SkippedDataPullStatus])
	fmt.Printf("Pulled New Vincario Market Valuation: %d \n", statsAggr[services.PulledValuationVincarioStatus])
	fmt.Printf("Skipped VIN due to error: %d \n", statsAggr[""])
	fmt.Println("--------------------------------------------------------")
	return nil
}
