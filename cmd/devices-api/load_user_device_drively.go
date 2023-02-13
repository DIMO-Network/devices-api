package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/volatiletech/null/v8"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
)

// loadUserDeviceDrivly iterates over user_devices with vin verified in USA and tries pulling data from drivly
func loadUserDeviceDrivly(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, forceSetAll bool, wmi string, pdb db.Store) error {
	// get all devices from DB.
	all, err := models.UserDevices(
		models.UserDeviceWhere.VinConfirmed.EQ(true),
		models.UserDeviceWhere.CountryCode.EQ(null.StringFrom("USA"))).
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
	logger.Info().Msgf("processing %d user_devices with verified VINs in the USA only", len(all))

	deviceDefinitionSvc := services.NewDeviceDefinitionService(pdb.DBS, logger, nil, settings)
	statsAggr := map[services.DrivlyDataStatusEnum]int{}
	for _, ud := range all {
		status, err := deviceDefinitionSvc.PullDrivlyData(ctx, ud.ID, ud.DeviceDefinitionID, ud.VinIdentifier.String, forceSetAll)
		if err != nil {
			logger.Err(err).Str("vin", ud.VinIdentifier.String).Msg("error pulling drivly data")
		} else {
			logger.Info().Msgf("%s vin: %s", status, ud.VinIdentifier.String)
		}
		statsAggr[status]++
	}
	fmt.Println("-------------------RUN SUMMARY--------------------------")
	// colorize each result
	fmt.Printf("Total VINs processed: %d \n", len(all))
	fmt.Printf("New Drivly Pulls (vin + valuations): %d \n", statsAggr[services.PulledAllDrivlyStatus])
	fmt.Printf("Pulled New Pricing & Offers: %d \n", statsAggr[services.PulledValuationDrivlyStatus])
	fmt.Printf("SkippedDrivlyStatus due to biz logic: %d \n", statsAggr[services.SkippedDrivlyStatus])
	fmt.Printf("SkippedDrivlyStatus due to error: %d \n", statsAggr[""])
	fmt.Println("--------------------------------------------------------")
	return nil
}
