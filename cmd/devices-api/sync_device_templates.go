package main

import (
	"context"
	"flag"
	"fmt"
	"os"
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
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*syncDeviceTemplatesCmd) Name() string     { return "sync-device-templates" }
func (*syncDeviceTemplatesCmd) Synopsis() string { return "sync-device-templates args to stdout." }
func (*syncDeviceTemplatesCmd) Usage() string {
	return `sync-device-templates [] <some text>:
	sync-device-templates args.
  `
}

// nolint
func (p *syncDeviceTemplatesCmd) SetFlags(f *flag.FlagSet) {

}

func (p *syncDeviceTemplatesCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	moveFromTemplateID := "10" // default
	if len(os.Args) > 2 {
		// parse out custom move from template ID option
		for i, a := range os.Args {
			if a == "--move-from-template" {
				moveFromTemplateID = os.Args[i+1]
				break
			}
		}
	}

	p.logger.Info().Msgf("starting syncing device templates based on device definition setting."+
		"\n Only moving from template ID: %s. To change specify --move-from-template XX. Set to 0 for none.", moveFromTemplateID)
	autoPiSvc := services.NewAutoPiAPIService(&p.settings, p.pdb.DBS)
	hardwareTemplateService := autopi.NewHardwareTemplateService(autoPiSvc, p.pdb.DBS, &p.logger)
	err := syncDeviceTemplates(ctx, &p.logger, &p.settings, p.pdb, hardwareTemplateService, moveFromTemplateID)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("failed to sync all devices with their templates")
	}
	p.logger.Info().Msg("success")
	return subcommands.ExitSuccess
}

// syncDeviceTemplates looks for DD's with a templateID set, and then compares to all UD's connected and Applies the template if doesn't match.
// If onlyMoveFromTemplate is > 0, then only apply the template if the current template is this value.
func syncDeviceTemplates(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb db.Store, autoPiHWSvc autopi.HardwareTemplateService,
	onlyMoveFromTemplate string) error {
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

	// group by template id
	templateXDefinitions := map[string][]*ddgrpc.GetDevicesMMYItemResponse{}

	for _, dd := range resp.Device {
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

		query := fmt.Sprintf(`select ud.id, udai.autopi_unit_id, (udai.metadata -> 'autoPiTemplateApplied')::text template_id from user_devices ud 
        inner join user_device_api_integrations udai on ud.id = udai.user_device_id
        where udai.integration_id = '27qftVRWQYpVDcO5DltO5Ojbjxk' and udai.metadata -> 'autoPiTemplateApplied' != '%s'`, templateID)

		ids := make([]string, len(dds))
		for i, dd := range dds {
			ids[i] = dd.Id
		}
		appendIn := " and ud.device_definition_id in ('" + strings.Join(ids, "','") + "')"

		type Result struct {
			UserDeviceID    string `boil:"id"`
			AutoPiUnitID    string `boil:"autopi_unit_id"`
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
			fmt.Printf("%d Update template for ud: %s from template %s to template %s\n", i+1, ud.UserDeviceID, ud.CurrentTemplate, templateID)
			_, err = autoPiHWSvc.ApplyHardwareTemplate(ctx, &pb.ApplyHardwareTemplateRequest{
				UserDeviceId:       ud.UserDeviceID,
				AutoApiUnitId:      ud.AutoPiUnitID,
				HardwareTemplateId: templateID,
			})
			if err != nil {
				logger.Err(err).Str("user_device_id", ud.UserDeviceID).Msg("failed to update template")
			}
			time.Sleep(time.Millisecond * 400)
		}

	}

	return nil
}
