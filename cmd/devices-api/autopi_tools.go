package main

import (
	"context"
	"flag"
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
