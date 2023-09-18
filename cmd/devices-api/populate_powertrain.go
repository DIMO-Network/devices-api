package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/constants"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type populateUSAPowertrainCmd struct {
	logger       zerolog.Logger
	settings     config.Settings
	pdb          db.Store
	nhtsaService services.INHTSAService
	deviceDefSvc services.DeviceDefinitionService

	useNHTSA bool
}

func (*populateUSAPowertrainCmd) Name() string { return "populate-powertrain" }
func (*populateUSAPowertrainCmd) Synopsis() string {
	return "populate-powertrain runs through all user devices trying to set the powertrain"
}
func (*populateUSAPowertrainCmd) Usage() string {
	return `populate-powertrain[-useNHTSA].
  `
}

// nolint
func (p *populateUSAPowertrainCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.useNHTSA, "useNHTSA", false, "Use useNHTSA")
}

func (p *populateUSAPowertrainCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.logger.Info().Msg("Iterating over all confirmed user devices to set their Powertrain")
	err := populateUSAPowertrain(ctx, &p.logger, p.pdb, p.nhtsaService, p.deviceDefSvc, p.useNHTSA)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error filling in powertrain data.")
	}
	return subcommands.ExitSuccess
}

func populateUSAPowertrain(ctx context.Context, logger *zerolog.Logger, pdb db.Store, nhtsaService services.INHTSAService, deviceDefSvc services.DeviceDefinitionService, useNHTSA bool) error {
	devices, err := models.UserDevices(
		models.UserDeviceWhere.CountryCode.IsNotNull(),
		models.UserDeviceWhere.VinIdentifier.IsNotNull(),
		models.UserDeviceWhere.VinConfirmed.EQ(true),
	).All(ctx, pdb.DBS().Writer)
	if err != nil {
		return err
	}

	// todo: if metadta.powertrain is null, just set it. Otherwise compare what we have vs. what we get from DD.
	// prompt user to select which one to use.
	// do we still want to use nhtsa - keep optional?
	processFromNTHSA := func(device *models.UserDevice) error {
		resp, err := nhtsaService.DecodeVIN(device.VinIdentifier.String)
		if err != nil {
			return err
		}
		dt, err := resp.DriveType()
		if err != nil {
			return err
		}

		md := new(services.UserDeviceMetadata)
		if err := device.Metadata.Unmarshal(md); err != nil {
			return err
		}

		md.PowertrainType = &dt
		if err := device.Metadata.Marshal(md); err != nil {
			return err
		}
		if _, err := device.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}
		return nil
	}

	processFromDeviceDefinition := func(device *models.UserDevice) error {
		dd, err := deviceDefSvc.GetDeviceDefinitionByID(ctx, device.DeviceDefinitionID)
		if err != nil {
			return err
		}

		md := new(services.UserDeviceMetadata)
		if err := device.Metadata.Unmarshal(md); err != nil {
			return err
		}

		if len(dd.DeviceAttributes) > 0 {
			// Find device attribute (powertrain_type)
			for _, item := range dd.DeviceAttributes {
				if item.Name == constants.PowerTrainTypeKey {
					powertrainType := services.ConvertPowerTrainStringToPowertrain(item.Value)
					if md.PowertrainType != nil {
						if !strings.EqualFold(powertrainType.String(), md.PowertrainType.String()) {
							fmt.Println("------------------------------------------------")
							fmt.Println(dd.Name)
							fmt.Println("Current powertrain is different than what Device Definitions proposes:")
							fmt.Println("Current:" + md.PowertrainType.String())
							fmt.Println("Proposed:" + powertrainType.String())
							fmt.Println("y/n accept proposed? [n]")
							accept := "n"
							_, _ = fmt.Scanln(accept)
							if strings.ToLower(strings.TrimSpace(accept)) == "n" {
								fmt.Println("leaving Powertrain as is.")
								return nil
							} else {
								fmt.Println("updating Powertrain to proposed.")
							}
						}
					}

					md.PowertrainType = &powertrainType
					if err := device.Metadata.Marshal(md); err != nil {
						return err
					}

					fmt.Printf("Updating powertrain for %s to %s \n", dd.Name, md.PowertrainType.String())

					break
				}
			}
		}

		if _, err := device.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}
		return nil
	}

	for _, device := range devices {
		if useNHTSA {
			if err := processFromNTHSA(device); err != nil {
				logger.Err(err).Str("userDeviceId", device.ID).Str("vin", device.VinIdentifier.String).Msg("Failed to update powertrain metadata NHTSA.")
			}
		}

		if !useNHTSA {
			if err := processFromDeviceDefinition(device); err != nil {
				logger.Err(err).Str("userDeviceId", device.ID).Str("vin", device.VinIdentifier.String).Msg("Failed to update powertrain metadata from DeviceDefinition.")
			}
		}

	}

	return nil
}
