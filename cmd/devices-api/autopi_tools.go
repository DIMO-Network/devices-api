package main

import (
	"context"
	"flag"
	"github.com/DIMO-Network/devices-api/models"
	"os"

	"strconv"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/devices-api/internal/services"
)

type autopiToolsCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*autopiToolsCmd) Name() string     { return "autopi-tools" }
func (*autopiToolsCmd) Synopsis() string { return "autopi-tools args to stdout." }
func (*autopiToolsCmd) Usage() string {
	return `autopi-tools [] <some text>:
	autopi-tools args.
  `
}

// nolint
func (p *autopiToolsCmd) SetFlags(f *flag.FlagSet) {

}

func (p *autopiToolsCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	autoPiSvc := services.NewAutoPiAPIService(&p.settings, p.pdb.DBS)
	autopiTools(os.Args, autoPiSvc)

	return subcommands.ExitSuccess
}

func autopiTools(args []string, autoPiSvc services.AutoPiAPIService) {
	if len(args) > 3 {
		templateName := args[2]
		var parent int
		var description string

		if args[3] == "-p" {
			parent, _ = strconv.Atoi(args[4])
			description = args[5]
		} else {
			parent = 0
			description = args[3]
		}
		newTemplateID, err := autoPiSvc.CreateNewTemplate(templateName, parent, description)
		if err == nil && newTemplateID > 0 {
			println("template created: " + strconv.Itoa(newTemplateID) + " : " + templateName + " : " + description)
			err := autoPiSvc.SetTemplateICEPowerSettings(newTemplateID)
			if err != nil {
				println(err.Error())
			} else {
				println("Set ICE Template PowerSettings set on template: " + templateName + " (" + strconv.Itoa(newTemplateID) + ")")
			}
			err = autoPiSvc.AddDefaultPIDsToTemplate(newTemplateID)
			if err != nil {
				println(err.Error())
			} else {
				println("Add default PIDs to template")
			}
		} else {
			println(err.Error())
		}
	} else {
		// "incorrect argument count"
		println("Incorrect parameter count. Please use following syntax:")
		println("\"thisEXECUTABLE  autopi-tools  templateName  [-p  parentIndex]  description\"")
	}
}

// clearVINFromAutopi iterates over all our known AutoPi units and sets the vin to blank, only if it has a value in their autopi profile.
func clearVINFromAutopi(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb db.Store) {
	// instantiate
	autoPiSvc := services.NewAutoPiAPIService(settings, pdb.DBS)

	// iterate all autopi units
	all, err := models.AutopiUnits().All(ctx, pdb.DBS().Reader)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to query db")
	}
	logger.Info().Msgf("processing %d autopi units", len(all))

	for _, unit := range all {
		innerLogger := logger.With().Str("autopiUnitID", unit.AutopiUnitID).Logger()

		autoPiDevice, err := autoPiSvc.GetDeviceByUnitID(unit.AutopiUnitID)
		if err != nil {
			innerLogger.Err(err).Msg("failed to call autopi api to get autoPiDevice")
		}
		if len(autoPiDevice.Vehicle.Vin) > 1 {
			// call api svc to update profile, setting vin = ""
			err = autoPiSvc.PatchVehicleProfile(autoPiDevice.Vehicle.ID, services.PatchVehicleProfile{
				Vin: "",
			})
			if err != nil {
				// uh oh spaghettie oh
				innerLogger.Err(err).Msg("failed to set VIN on autopi service")
			}
		}
	}

	logger.Info().Msg("all done")
}
