package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/DIMO-Network/shared"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	es "github.com/DIMO-Network/devices-api/internal/elasticsearch"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type populateESRegionDataCmd struct {
	logger     zerolog.Logger
	settings   config.Settings
	pdb        db.Store
	esInstance es.ElasticSearch
	ddSvc      services.DeviceDefinitionService
}

func (*populateESRegionDataCmd) Name() string     { return "populate-es-region-data" }
func (*populateESRegionDataCmd) Synopsis() string { return "populate-es-region-data args to stdout." }
func (*populateESRegionDataCmd) Usage() string {
	return `populate-es-region-data:
	populate-es-region-data args.
  `
}

// nolint
func (p *populateESRegionDataCmd) SetFlags(f *flag.FlagSet) {

}

func (p *populateESRegionDataCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := populateESRegionData(ctx, &p.settings, p.esInstance, p.pdb, &p.logger, p.ddSvc)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running elastic search region update")
	}
	return subcommands.ExitSuccess
}

func populateESRegionData(ctx context.Context, settings *config.Settings, e es.ElasticSearch, pdb db.Store, logger *zerolog.Logger, ddSvc services.DeviceDefinitionService) error {
	db := pdb.DBS().Reader

	uAPIInt, err := models.UserDeviceAPIIntegrations(
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all user devices: %w", err)
	}

	for _, apiInt := range uAPIInt {
		ud := apiInt.R.UserDevice
		d, err := ddSvc.GetDeviceDefinitionBySlug(ctx, apiInt.R.UserDevice.DefinitionID)
		if err != nil {
			logger.Err(err).
				Str("userDeviceId", apiInt.UserDeviceID).
				Str("deviceDefinitionId", apiInt.R.UserDevice.DefinitionID).
				Msg("Failed to retrieve device definition.")
			continue
		}

		if ud.CountryCode.IsZero() {
			logger.Error().Str("userDeviceId", ud.ID).
				Str("deviceDefinitionId", ud.DefinitionID).
				Msg("Device missing country.")
			continue
		}

		md := services.UserDeviceMetadata{}
		if err = ud.Metadata.Unmarshal(&md); err != nil {
			logger.Error().Msgf("Could not unmarshal userdevice metadata for device: %s", ud.ID)
			continue
		}
		if !md.ElasticRegionSynced {
			c := constants.FindCountry(ud.CountryCode.String)
			if c == nil || c.Region == "" {
				logger.Error().Msgf("Could not get region from country informaton for - userDeviceID: %s, definition: %s", ud.ID, ud.DefinitionID)
				continue
			}

			dd := services.DeviceDefinitionDTO{
				DefinitionID: ud.DefinitionID,
				UserDeviceID: ud.ID,
				MakeSlug:     d.Make.NameSlug,
				ModelSlug:    shared.SlugString(d.Model),
				Region:       c.Region,
			}

			err = e.UpdateDeviceRegionsByQuery(dd, settings.ElasticDeviceStatusIndex)
			if err != nil {
				logger.Error().Msgf("error occurred during es update: %s", err)
				continue
			}
			md.ElasticRegionSynced = true

			err = ud.Metadata.Marshal(&md)
			if err != nil {
				logger.Error().Msgf("Could not marshal device metadata for, DefinitionId: %s", ud.DefinitionID)
				continue
			}

			if _, err := ud.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				logger.Err(err).
					Str("DefinitionId", ud.DefinitionID).
					Msg("Error updating device")
				continue
			}
		} else {
			logger.Debug().Msgf("Record has already been updated for, DefinitionId: %s", ud.DefinitionID)
		}
	}

	return nil
}
