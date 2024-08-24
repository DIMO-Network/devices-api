package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
)

type populateSDFingerprintTable struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	producer  sarama.SyncProducer
	container dependencyContainer
}

func (*populateSDFingerprintTable) Name() string { return "populate-sd-fingerprint-table" }
func (*populateSDFingerprintTable) Synopsis() string {
	return "populate-sd-fingerprint-table args to stdout."
}
func (*populateSDFingerprintTable) Usage() string {
	return `populate-sd-fingerprint-table:
	populate-sd-fingerprint-table args.
  `
}

// nolint
func (p *populateSDFingerprintTable) SetFlags(f *flag.FlagSet) {

}

func (p *populateSDFingerprintTable) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	p.producer = p.container.getKafkaProducer()
	err := remakeSDFingerprintTable(p.pdb, p.producer, &p.logger)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running SD fingerprint backfill.")
	}
	return subcommands.ExitSuccess
}

func remakeSDFingerprintTable(pdb db.Store, producer sarama.SyncProducer, logger *zerolog.Logger) error {
	logger.Info().Msgf("Starting synthetic fingerprint backfill to %s.", "synthetic-fingerprint-event-table")

	ctx := context.Background()

	sds, err := models.SyntheticDevices(
		models.SyntheticDeviceWhere.TokenID.IsNotNull(),
		qm.OrderBy(models.SyntheticDeviceColumns.TokenID),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return fmt.Errorf("failed to minted devices: %w", err)
	}

	type TableMsg struct {
		IntegrationID     string         `json:"integrationId"`
		ExternalID        string         `json:"externalId"`
		VehicleTokenID    int            `json:"vehicleTokenId"`
		Address           common.Address `json:"address"`
		WalletChildNumber uint32         `json:"walletChildNumber"`
	}

	for _, sd := range sds {
		var integrationID string
		if sd.IntegrationTokenID.Cmp(decimal.New(1, 0)) == 0 {
			integrationID = "22N2xaPOq2WW2gAHBHd0Ikn4Zob"
		} else {
			integrationID = "26A5Dk3vvvQutjSyF0Jka2DP5lg"
		}

		ud, err := models.UserDevices(
			models.UserDeviceWhere.TokenID.EQ(sd.VehicleTokenID),
			qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID)),
		).One(ctx, pdb.DBS().Reader)
		if err != nil {
			return err
		}

		if len(ud.R.UserDeviceAPIIntegrations) == 0 {
			logger.Warn().Str("userDeviceId", ud.ID).Msg("Vehicle has a synthetic but no integration.")
			// This is bad.
			continue
		}

		vehicleToken, _ := sd.VehicleTokenID.Int64()

		tm := TableMsg{
			IntegrationID:     integrationID,
			ExternalID:        ud.R.UserDeviceAPIIntegrations[0].ExternalID.String,
			VehicleTokenID:    int(vehicleToken),
			Address:           common.BytesToAddress(sd.WalletAddress),
			WalletChildNumber: uint32(sd.WalletChildNumber),
		}
		b, err := json.Marshal(tm)
		if err != nil {
			return err
		}

		msg := &sarama.ProducerMessage{
			Topic: "synthetic-fingerprint-event-table",
			Key:   sarama.StringEncoder(ud.ID),
			Value: sarama.ByteEncoder(b),
		}
		if _, _, err := producer.SendMessage(msg); err != nil {
			return err
		}
	}

	return nil
}
