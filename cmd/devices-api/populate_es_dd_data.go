package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/shared/db"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	es "github.com/DIMO-Network/devices-api/internal/elasticsearch"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type populateESDDDataCmd struct {
	logger     zerolog.Logger
	settings   config.Settings
	pdb        db.Store
	esInstance es.ElasticSearch
	ddSvc      services.DeviceDefinitionService
}

func (*populateESDDDataCmd) Name() string     { return "populate-es-dd-data" }
func (*populateESDDDataCmd) Synopsis() string { return "populate-es-dd-data args to stdout." }
func (*populateESDDDataCmd) Usage() string {
	return `populate-es-dd-data:
	populate-es-dd-data args.
  `
}

// nolint
func (p *populateESDDDataCmd) SetFlags(f *flag.FlagSet) {

}

func (p *populateESDDDataCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := populateESDDData(ctx, &p.settings, p.esInstance, p.pdb, &p.logger, p.ddSvc)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running elastic search dd update")
	}
	return subcommands.ExitSuccess
}

func populateESDDData(ctx context.Context, settings *config.Settings, e es.ElasticSearch, pdb db.Store, logger *zerolog.Logger, ddSvc services.DeviceDefinitionService) error {
	db := pdb.DBS().Reader

	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.ExternalID.IsNotNull(),
		qm.Where("metadata ->> 'elasticDefinitionSynced' IS NULL OR metadata ->> 'elasticDefinitionSynced' = ?", false),
	).All(ctx, db)

	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}

	ids := make([]string, len(apiInts))
	for _, d := range apiInts {
		ids = append(ids, d.R.UserDevice.DeviceDefinitionID)
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

	for _, apiInt := range apiInts {

		dd, err := filterDeviceDefinition(apiInt.R.UserDevice.DeviceDefinitionID, deviceDefinitionResponse)
		if err != nil {
			logger.Fatal().Err(err)
			continue
		}

		makeRel := dd.Make
		ddID := dd.DeviceDefinitionId

		md := services.UserDeviceMetadata{}
		if err = apiInt.R.UserDevice.Metadata.Unmarshal(&md); err != nil {
			logger.Error().Msgf("Could not unmarshal userdevice metadata for device: %s", apiInt.R.UserDevice.ID)
			continue
		}

		if !md.ElasticDefinitionSynced {
			dd := services.DeviceDefinitionDTO{
				DeviceDefinitionID: ddID,
				UserDeviceID:       apiInt.R.UserDevice.ID,
				Make:               makeRel.Name,
				Model:              dd.Model,
				Year:               int(dd.Year),
			}
			err = e.UpdateAutopiDevicesByQuery(dd, settings.ElasticDeviceStatusIndex)
			if err != nil {
				logger.Error().Msgf("error occurred during es update: %s", err)
				continue
			}

			md.ElasticDefinitionSynced = true
			err = apiInt.R.UserDevice.Metadata.Marshal(&md)
			if err != nil {
				logger.Error().Msgf("could not marshal userdevice metadata for device: %s", apiInt.R.UserDevice.ID)
				continue
			}

			if _, err := apiInt.R.UserDevice.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				logger.Err(err).Str("userDeviceId", apiInt.UserDeviceID).Msg("Could not update metadata for device.")
				continue
			}
		} else {
			logger.Debug().Msgf("device record has already been updated for user device %s", apiInt.R.UserDevice.ID)
		}
	}

	return nil
}
