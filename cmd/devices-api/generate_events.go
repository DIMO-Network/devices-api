package main

import (
	"context"
	"flag"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type generateEventCmd struct {
	logger       zerolog.Logger
	settings     config.Settings
	pdb          db.Store
	eventService services.EventService
	ddSvc        services.DeviceDefinitionService

	generate bool
}

func (*generateEventCmd) Name() string     { return "events" }
func (*generateEventCmd) Synopsis() string { return "events args to stdout." }
func (*generateEventCmd) Usage() string {
	return `events [-generate] <some text>:
	events args.
  `
}

func (p *generateEventCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.generate, "generate", true, "generate events")
}

func (p *generateEventCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if !p.generate {
		return subcommands.ExitSuccess
	}

	generateEvents(p.logger, p.pdb, p.eventService, p.ddSvc)
	return subcommands.ExitSuccess
}

func generateEvents(logger zerolog.Logger, pdb db.Store, eventService services.EventService, ddSvc services.DeviceDefinitionService) {
	ctx := context.Background()
	tx, err := pdb.DBS().Reader.BeginTx(ctx, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create transaction")
	}
	defer tx.Rollback() //nolint
	devices, err := models.UserDevices().All(ctx, tx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve all devices and definitions for event generation")
	}

	ids := make([]string, len(devices))
	for _, d := range devices {
		ids = append(ids, d.DeviceDefinitionID)
	}

	deviceDefinitionResponse, err := ddSvc.GetDeviceDefinitionsByIDs(ctx, ids)

	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve all devices and definitions for event generation from grpc")
	}

	filterDeviceDefinition := func(id string, items []*ddgrpc.GetDeviceDefinitionItemResponse) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
		for _, dd := range items {
			if id == dd.DeviceDefinitionId {
				return dd, nil
			}
		}
		return nil, errors.Errorf("no device definition %s", id)
	}

	for _, device := range devices {

		dd, err := filterDeviceDefinition(device.DeviceDefinitionID, deviceDefinitionResponse)

		if err != nil {
			logger.Fatal().Err(err)
			continue
		}

		err = eventService.Emit(
			&services.Event{
				Type:    controllers.UserDeviceCreationEventType,
				Subject: device.UserID,
				Source:  "devices-api",
				Data: controllers.UserDeviceEvent{
					Timestamp: device.CreatedAt,
					UserID:    device.UserID,
					Device: services.UserDeviceEventDevice{
						ID:    device.ID,
						Make:  dd.Make.Name,
						Model: dd.Type.Model,
						Year:  int(dd.Type.Year),
					},
				},
			},
		)
		if err != nil {
			logger.Err(err).Msgf("Failed to emit creation event for device %s", device.ID)
		}
	}

	scIntegs, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
	).All(ctx, tx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve all active integrations")
	}

	deviceDefinitionResponse, err = ddSvc.GetDeviceDefinitionsByIDs(ctx, ids)

	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve all devices and definitions for event generation from grpc")
	}

	for _, scInteg := range scIntegs {
		if !scInteg.R.UserDevice.VinIdentifier.Valid {
			logger.Warn().Msgf("Device %s has an active integration but no VIN", scInteg.UserDeviceID)
			continue
		}
		if !scInteg.R.UserDevice.VinConfirmed {
			logger.Warn().Msgf("Device %s has an active integration but the VIN %s is unconfirmed", scInteg.UserDeviceID, scInteg.R.UserDevice.VinIdentifier.String)
			continue
		}

		dd, err := filterDeviceDefinition(scInteg.R.UserDevice.DeviceDefinitionID, deviceDefinitionResponse)

		if err != nil {
			logger.Fatal().Err(err)
			continue
		}

		integration, err := ddSvc.GetIntegrationByID(ctx, scInteg.IntegrationID)

		if err != nil {
			logger.Fatal().Err(err)
			continue
		}

		err = eventService.Emit(
			&services.Event{
				Type:    "com.dimo.zone.device.integration.create",
				Subject: scInteg.UserDeviceID,
				Source:  "devices-api",
				Data: services.UserDeviceIntegrationEvent{
					Timestamp: scInteg.CreatedAt,
					UserID:    scInteg.R.UserDevice.UserID,
					Device: services.UserDeviceEventDevice{
						ID:    scInteg.UserDeviceID,
						Make:  dd.Make.Name,
						Model: dd.Type.Model,
						Year:  int(dd.Type.Year),
						VIN:   scInteg.R.UserDevice.VinIdentifier.String,
					},
					Integration: services.UserDeviceEventIntegration{
						ID:     integration.Id,
						Type:   integration.Type,
						Style:  integration.Style,
						Vendor: integration.Vendor,
					},
				},
			},
		)
		if err != nil {
			logger.Err(err).Msgf("Failed to emit integration creation event for device %s", scInteg.UserDeviceID)
		}
	}

	err = tx.Commit()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to commit (kinda dumb)")
	}
}
