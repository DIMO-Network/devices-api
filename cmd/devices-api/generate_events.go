package main

import (
	"context"
	"flag"

	"github.com/DIMO-Network/shared"

	"github.com/DIMO-Network/devices-api/internal/constants"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
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
	container    dependencyContainer
	generate     bool
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

func (p *generateEventCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if !p.generate {
		return subcommands.ExitSuccess
	}

	p.eventService = services.NewEventService(&p.logger, &p.settings, p.container.getKafkaProducer())
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

	deviceDefinitionResponse := make([]*ddgrpc.GetDeviceDefinitionItemResponse, len(devices))
	for i, d := range devices {
		dd, err := ddSvc.GetDeviceDefinitionBySlug(ctx, d.DefinitionID)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to retrieve all devices and definitions for event generation from grpc")
		}
		deviceDefinitionResponse[i] = dd
	}

	filterDeviceDefinition := func(id string, items []*ddgrpc.GetDeviceDefinitionItemResponse) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
		for _, dd := range items {
			if id == dd.Id {
				return dd, nil
			}
		}
		return nil, errors.Errorf("no device definition %s", id)
	}

	for _, device := range devices {

		dd, err := filterDeviceDefinition(device.DefinitionID, deviceDefinitionResponse)

		if err != nil {
			logger.Fatal().Err(err)
			continue
		}

		err = eventService.Emit(
			&shared.CloudEvent[any]{
				Type:    constants.UserDeviceCreationEventType,
				Subject: device.UserID,
				Source:  "devices-api",
				Data: services.UserDeviceEvent{
					Timestamp: device.CreatedAt,
					UserID:    device.UserID,
					Device: services.UserDeviceEventDevice{
						ID:           device.ID,
						Make:         dd.Make.Name,
						Model:        dd.Model,
						Year:         int(dd.Year),
						DefinitionID: dd.Id,
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

	for _, scInteg := range scIntegs {
		if !scInteg.R.UserDevice.VinIdentifier.Valid {
			logger.Warn().Msgf("Device %s has an active integration but no VIN", scInteg.UserDeviceID)
			continue
		}
		if !scInteg.R.UserDevice.VinConfirmed {
			logger.Warn().Msgf("Device %s has an active integration but the VIN %s is unconfirmed", scInteg.UserDeviceID, scInteg.R.UserDevice.VinIdentifier.String)
			continue
		}

		dd, err := filterDeviceDefinition(scInteg.R.UserDevice.DefinitionID, deviceDefinitionResponse)

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
			&shared.CloudEvent[any]{
				Type:    "com.dimo.zone.device.integration.create",
				Subject: scInteg.UserDeviceID,
				Source:  "devices-api",
				Data: services.UserDeviceIntegrationEvent{
					Timestamp: scInteg.CreatedAt,
					UserID:    scInteg.R.UserDevice.UserID,
					Device: services.UserDeviceEventDevice{
						ID:           scInteg.UserDeviceID,
						Make:         dd.Make.Name,
						Model:        dd.Model,
						Year:         int(dd.Year),
						VIN:          scInteg.R.UserDevice.VinIdentifier.String,
						DefinitionID: dd.Id,
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
