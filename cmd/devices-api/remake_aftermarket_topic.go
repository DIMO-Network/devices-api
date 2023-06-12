package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
)

type remakeAftermarketTopicCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	container dependencyContainer
}

func (*remakeAftermarketTopicCmd) Name() string { return "remake-aftermarket-topic" }
func (*remakeAftermarketTopicCmd) Synopsis() string {
	return "remake-aftermarket-topic args to stdout."
}
func (*remakeAftermarketTopicCmd) Usage() string {
	return `remake-aftermarket-topic:
	remake-autopi-topic args.
  `
}

// nolint
func (p *remakeAftermarketTopicCmd) SetFlags(f *flag.FlagSet) {

}

func (p *remakeAftermarketTopicCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := remakeAftermarketTopic(ctx, p.pdb, p.container.getKafkaProducer(), p.container.getDeviceDefinitionService())
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running AutoPi Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

// remakeAftermarketTopic re-populates the autopi ingest registrar topic based on data we have in user_device_api_integrations
func remakeAftermarketTopic(ctx context.Context, pdb db.Store, producer sarama.SyncProducer, ddSvc services.DeviceDefinitionService) error {
	reg := services.NewIngestRegistrar(producer)
	db := pdb.DBS().Reader

	integ, err := ddSvc.GetIntegrationByVendor(ctx, constants.AutoPiVendor)
	if err != nil {
		return fmt.Errorf("failed to retrieve AutoPi integration: %w", err)
	}

	aps, err := models.AutopiUnits(
		models.AutopiUnitWhere.VehicleTokenID.IsNotNull(),
		qm.Load(models.AutopiUnitRels.VehicleToken),
	).All(ctx, db)
	if err != nil {
		return err
	}

	for _, ap := range aps {
		if !ap.R.VehicleToken.UserDeviceID.Valid {
			continue
		}

		if err := reg.Register2(&services.AftermarketDeviceVehicleMapping{
			AftermarketDevice: services.AftermarketDeviceVehicleMappingAftermarketDevice{
				Address:       common.BytesToAddress(ap.EthereumAddress.Bytes),
				Token:         ap.TokenID.Int(nil),
				Serial:        ap.AutopiUnitID,
				IntegrationID: integ.Id,
			},
			Vehicle: services.AftermarketDeviceVehicleMappingVehicle{
				Token:        ap.VehicleTokenID.Int(nil),
				UserDeviceID: ap.R.VehicleToken.UserDeviceID.String,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
