package main

import (
	"context"
	"fmt"

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
		d, err := ddSvc.GetDeviceDefinitionByID(ctx, apiInt.R.UserDevice.DeviceDefinitionID)
		if err != nil {
			logger.Err(err).
				Str("userDeviceId", apiInt.UserDeviceID).
				Str("deviceDefinitionId", apiInt.R.UserDevice.DeviceDefinitionID).
				Msg("Failed to retrieve device definition.")
			continue
		}

		if ud.CountryCode.IsZero() {
			logger.Error().Str("userDeviceId", ud.ID).
				Str("deviceDefinitionId", ud.DeviceDefinitionID).
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
				logger.Error().Msgf("Could not get region from country informaton for - userDeviceID: %s, deviceDefinition: %s", ud.ID, ud.DeviceDefinitionID)
				continue
			}

			dd := services.DeviceDefinitionDTO{
				DeviceDefinitionID: ud.DeviceDefinitionID,
				UserDeviceID:       ud.ID,
				MakeSlug:           d.Type.MakeSlug,
				ModelSlug:          d.Type.ModelSlug,
				Region:             c.Region,
			}

			err = e.UpdateDeviceRegionsByQuery(dd, settings.ElasticDeviceStatusIndex)
			if err != nil {
				logger.Error().Msgf("error occurred during es update: %s", err)
				continue
			}
			md.ElasticRegionSynced = true

			err = ud.Metadata.Marshal(&md)
			if err != nil {
				logger.Error().Msgf("Could not marshal device metadata for, DeviceDefinitionId: %s", ud.DeviceDefinitionID)
				continue
			}

			if _, err := ud.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				logger.Err(err).
					Str("DeviceDefinitionId", ud.DeviceDefinitionID).
					Msg("Error updating device")
				continue
			}
		} else {
			logger.Debug().Msgf("Record has already been updated for, DeviceDefinitionId: %s", ud.DeviceDefinitionID)
		}
	}

	return nil
}
