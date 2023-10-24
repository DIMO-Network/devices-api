package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/google/subcommands"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type syncDeviceTemplatesCmd struct {
	logger             zerolog.Logger
	settings           config.Settings
	pdb                db.Store
	moveFromTemplateID *string
	targetTemplateID   *string
}

func (*syncDeviceTemplatesCmd) Name() string { return "sync-device-templates" }
func (*syncDeviceTemplatesCmd) Synopsis() string {
	return "iterate through all UD's and set the template to what our config says should be, or filter down impact with options"
}
func (*syncDeviceTemplatesCmd) Usage() string {
	return `sync-device-templates [-move-from-template] <template ID, 0 to move from any>
									[-target-template] <template ID>
  `
}

func (p *syncDeviceTemplatesCmd) SetFlags(f *flag.FlagSet) {
	p.moveFromTemplateID = f.String("move-from-template", "10", "By default only moves devices in template 10, specify this to change or 0 for any")
	p.targetTemplateID = f.String("target-template", "", "Filter device definitions to apply for where the template is this value. Good for moving only certain autopi's.")
}

func (p *syncDeviceTemplatesCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	moveFromTemplateID := "10" // default
	if p.moveFromTemplateID != nil {
		moveFromTemplateID = *p.moveFromTemplateID
	}

	p.logger.Info().Msgf("starting syncing device templates based on device definition setting."+
		"\n Only moving from template ID: %s. To change specify --move-from-template XX. Set to 0 for none.", moveFromTemplateID)
	autoPiSvc := services.NewAutoPiAPIService(&p.settings, p.pdb.DBS)
	hardwareTemplateService := autopi.NewHardwareTemplateService(autoPiSvc, p.pdb.DBS, &p.logger)
	err := syncDeviceTemplates(ctx, &p.logger, &p.settings, p.pdb, hardwareTemplateService, moveFromTemplateID, p.targetTemplateID)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("failed to sync all devices with their templates")
	}
	p.logger.Info().Msg("success")
	return subcommands.ExitSuccess
}

// syncDeviceTemplates looks for DD's with a templateID set, and then compares to all UD's connected and Applies the template if doesn't match.
// If onlyMoveFromTemplate is > 0, then only apply the template if the current template is this value.
func syncDeviceTemplates(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb db.Store, autoPiHWSvc autopi.HardwareTemplateService, onlyMoveFromTemplate string, targetTemplateID *string) error {
	conn, err := grpc.Dial(settings.DefinitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	resp, err := definitionsClient.GetDeviceDefinitionsWithHardwareTemplate(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	if targetTemplateID != nil && len(*targetTemplateID) > 0 {
		fmt.Printf("Selected Target Template: %s\n", *targetTemplateID)
	}

	// group by template id
	templateXDefinitions := map[string][]*ddgrpc.GetDevicesMMYItemResponse{}

	for _, dd := range resp.Device {
		if targetTemplateID != nil && len(*targetTemplateID) > 0 {
			if dd.HardwareTemplateId != *targetTemplateID {
				// skip all template ID's that do not match the target
				continue
			}
		}
		// we currently only allow integer type template ID's
		tIDInt, err := strconv.Atoi(dd.HardwareTemplateId)
		if tIDInt == 0 || err != nil {
			continue
		}
		templateXDefinitions[dd.HardwareTemplateId] = append(templateXDefinitions[dd.HardwareTemplateId], dd)
	}

	// loop by each template
	for templateID, dds := range templateXDefinitions {
		fmt.Printf("\nFound %d device definitions for template %s\n", len(dds), templateID)

		query := fmt.Sprintf(`select ud.id, udai.serial, (udai.metadata -> 'autoPiTemplateApplied')::text template_id from user_devices ud 
        inner join user_device_api_integrations udai on ud.id = udai.user_device_id
        where udai.integration_id = '27qftVRWQYpVDcO5DltO5Ojbjxk' and udai.metadata -> 'autoPiTemplateApplied' != '%s'`, templateID)

		ids := make([]string, len(dds))
		for i, dd := range dds {
			ids[i] = dd.Id
		}
		appendIn := " and ud.device_definition_id in ('" + strings.Join(ids, "','") + "')"

		type Result struct {
			UserDeviceID    string `boil:"id"`
			AutoPiUnitID    string `boil:"serial"`
			CurrentTemplate string `boil:"template_id"`
		}
		var userDevices []Result
		err := queries.Raw(query+appendIn).Bind(ctx, pdb.DBS().Reader, &userDevices)
		if err != nil {
			logger.Err(err).Msg("Database failure retrieving user devices")
			return err
		}
		fmt.Printf("found total of %d impacted user_devices to move to template %s\n", len(userDevices), templateID)

		for i, ud := range userDevices {
			if onlyMoveFromTemplate != "0" && ud.CurrentTemplate != onlyMoveFromTemplate {
				fmt.Printf("%d Skipped ud: %s because it is not currently in template %s\n", i+1, ud.UserDeviceID, onlyMoveFromTemplate)
				continue
			}
			fmt.Printf("%d Update template for ud: %s from template %s to template %s", i+1, ud.UserDeviceID, ud.CurrentTemplate, templateID)
			_, err = autoPiHWSvc.ApplyHardwareTemplate(ctx, &pb.ApplyHardwareTemplateRequest{
				UserDeviceId:       ud.UserDeviceID,
				AutoApiUnitId:      ud.AutoPiUnitID,
				HardwareTemplateId: templateID,
			})
			if err != nil {
				fmt.Printf(" : failed\n")
				logger.Err(err).Str("user_device_id", ud.UserDeviceID).Msg("failed to update template")
			} else {
				fmt.Printf(" : ok\n")
			}
			time.Sleep(time.Millisecond * 400)
		}

	}

	return nil
}
