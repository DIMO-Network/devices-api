package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/pkg/errors"

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
		// todo what about device style - i think that also gets powertrain more specific
		dd, err := deviceDefSvc.GetDeviceDefinitionByID(ctx, device.DeviceDefinitionID)
		if err != nil {
			return err
		}
		var ds *grpc.DeviceStyle
		if device.DeviceStyleID.Valid {
			ds, err = deviceDefSvc.GetDeviceStyleByID(ctx, device.DeviceStyleID.String)
			if err != nil {
				return errors.Wrapf(err, "failed to get device style for id: %s", device.DeviceStyleID.String)
			}
		}

		md := new(services.UserDeviceMetadata)
		if err := device.Metadata.Unmarshal(md); err != nil {
			return err
		}
		initialPowertrain := md.PowertrainType
		powertrainFromStyle := false
		var proposedPowertrain *services.PowertrainType
		// try to get powertrain from style level
		if ds != nil && len(ds.DeviceAttributes) > 0 {
			for _, item := range ds.DeviceAttributes {
				if item.Name == constants.PowerTrainTypeKey {
					fmt.Println("found powertrain value from device_style: " + item.Value)
					pt := services.ConvertPowerTrainStringToPowertrain(item.Value)
					proposedPowertrain = &pt
					powertrainFromStyle = true
					break
				}
			}
		} else if len(dd.DeviceAttributes) > 0 {
			// otherwise powertrain from the device definition
			// Find device attribute (powertrain_type)
			for _, item := range dd.DeviceAttributes {
				if item.Name == constants.PowerTrainTypeKey {
					pt := services.ConvertPowerTrainStringToPowertrain(item.Value)
					proposedPowertrain = &pt
					break
				}
			}
		} else {
			fmt.Println("no attributes found, using default powertrain ICE for: " + dd.Name)
			ice := services.ICE
			proposedPowertrain = &ice
		}
		// short circuit if can't propose anything - very edge case
		if proposedPowertrain == nil {
			return nil
		}
		// short circuit if don't need to update
		if initialPowertrain != nil && strings.EqualFold(proposedPowertrain.String(), initialPowertrain.String()) {
			return nil
		}
		// find case where we have a mismatch and want user input
		if initialPowertrain != nil {
			if !strings.EqualFold(proposedPowertrain.String(), initialPowertrain.String()) {
				fmt.Println("------------------------------------------------")
				fmt.Printf("%s https://admin.team.dimo.zone/device-definitions/%s \n", dd.Name, dd.DeviceDefinitionId)
				fmt.Println("Current powertrain is different than what Device Definitions proposes:")
				fmt.Println("Current:" + md.PowertrainType.String())
				fmt.Println("Proposed:" + proposedPowertrain.String())
				if ds != nil {
					if powertrainFromStyle {
						fmt.Printf("Powertrain came from Device Style: %s : %s\n", ds.Name, ds.Id)
					} else {
						fmt.Printf("Issue: No Powertrain found in Device Style: %s : %s\n", ds.Name, ds.Id)
					}
				}
				fmt.Println("y/n accept proposed? [n]")
				var accept string
				_, err := fmt.Scanln(&accept)
				if err != nil {
					fmt.Println("Error:", err)
					return err
				}
				if strings.ToLower(strings.TrimSpace(accept)) == "n" {
					fmt.Println("leaving Powertrain as is.")
					return nil
				}
				fmt.Println("updating Powertrain to proposed.")
				fmt.Println("------------------------------------------------")
			}
		}

		md.PowertrainType = proposedPowertrain
		if err := device.Metadata.Marshal(md); err != nil {
			return err
		}

		fmt.Printf("Updating powertrain for %s to %s. before: %s \n", dd.Name, md.PowertrainType.String(), initialPowertrain)

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
