package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/pkg/errors"
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
		proposedPowertrain := services.ICE
		// try to get powertrain from style level
		if ds != nil && len(ds.DeviceAttributes) > 0 {
			for _, item := range ds.DeviceAttributes {
				if item.Name == constants.PowerTrainTypeKey {
					proposedPowertrain = services.ConvertPowerTrainStringToPowertrain(item.Value)
					break
				}
			}
		} else if len(dd.DeviceAttributes) > 0 {
			// otherwise powertrain from the device definition
			// Find device attribute (powertrain_type)
			for _, item := range dd.DeviceAttributes {
				if item.Name == constants.PowerTrainTypeKey {
					proposedPowertrain = services.ConvertPowerTrainStringToPowertrain(item.Value)
					break
				}
			}
		} else {
			fmt.Println("no attributes found, using default powertrain ICE for: " + dd.Name)
		}
		if initialPowertrain == &proposedPowertrain {
			return nil // no need to update in this case
		}
		// find case where we have a mismatch and want user input
		if initialPowertrain != nil {
			if !strings.EqualFold(proposedPowertrain.String(), initialPowertrain.String()) {
				fmt.Println("------------------------------------------------")
				fmt.Println(dd.Name)
				fmt.Println("Current powertrain is different than what Device Definitions proposes:")
				fmt.Println("Current:" + md.PowertrainType.String())
				fmt.Println("Proposed:" + proposedPowertrain.String())
				if ds != nil {
					fmt.Println("Powertrain came from Device Style")
				}
				fmt.Println("y/n accept proposed? [n]")
				accept := "n"
				_, err = fmt.Scanln(accept)
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

		md.PowertrainType = &proposedPowertrain
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
