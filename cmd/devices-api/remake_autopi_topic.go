package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

type remakeAutoPiTopicCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	producer  sarama.SyncProducer
	ddSvc     services.DeviceDefinitionService
	container dependencyContainer
}

func (*remakeAutoPiTopicCmd) Name() string     { return "remake-autopi-topic" }
func (*remakeAutoPiTopicCmd) Synopsis() string { return "remake-autopi-topic args to stdout." }
func (*remakeAutoPiTopicCmd) Usage() string {
	return `remake-autopi-topic:
	remake-autopi-topic args.
  `
}

// nolint
func (p *remakeAutoPiTopicCmd) SetFlags(f *flag.FlagSet) {

}

func (p *remakeAutoPiTopicCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	err := remakeAutoPiTopic(ctx, p.pdb, p.producer, p.ddSvc)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running AutoPi Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

// remakeAutoPiTopic re-populates the autopi ingest registrar topic based on data we have in user_device_api_integrations
func remakeAutoPiTopic(ctx context.Context, pdb db.Store, producer sarama.SyncProducer, ddSvc services.DeviceDefinitionService) error {
	reg := services.NewIngestRegistrar(producer)
	db := pdb.DBS().Reader

	// Grab the Smartcar integration ID, there should be exactly one.
	var apIntID string
	integ, err := ddSvc.GetIntegrationByVendor(ctx, constants.AutoPiVendor)
	if err != nil {
		return fmt.Errorf("failed to retrieve AutoPi integration: %w", err)
	}
	apIntID = integ.Id

	// Find all integration instances that have acquired Smartcar ids.
	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(apIntID),
		models.UserDeviceAPIIntegrationWhere.ExternalID.NEQ(null.StringFromPtr(nil)),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}

	// For each of these send a new registration message, keyed by autopi device ID.
	for _, apiInt := range apiInts {
		md := new(services.UserDeviceAPIIntegrationsMetadata)
		err := apiInt.Metadata.Unmarshal(md)
		if err != nil {
			return errors.Wrap(err, "unable to unmarshall userDeviceAPIINtegrations Metadata")
		}
		if md.AutoPiUnitID == nil {
			return fmt.Errorf("failed to register AutoPi-DIMO id link for device %s. autoPi unitID is nil", apiInt.UserDeviceID)
		}
		if err := reg.Register(*md.AutoPiUnitID, apiInt.UserDeviceID, apIntID); err != nil {
			return fmt.Errorf("failed to register AutoPi-DIMO id link for device %s: %w", apiInt.UserDeviceID, err)
		}
	}

	return nil
}
