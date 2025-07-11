package main

import (
	"context"
	"flag"
	"os"

	"math/big"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/genericad"
	"github.com/DIMO-Network/devices-api/internal/utils"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type web2PairCmd struct {
	logger    zerolog.Logger
	settings  config.Settings
	pdb       db.Store
	container dependencyContainer
}

func (*web2PairCmd) Name() string     { return "web2-pair" }
func (*web2PairCmd) Synopsis() string { return "web2-pair args to stdout." }
func (*web2PairCmd) Usage() string {
	return `web2-pair [] <some text>:
	web2-pair args.
  `
}

// nolint
func (p *web2PairCmd) SetFlags(f *flag.FlagSet) {

}

func (p *web2PairCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	producer := p.container.getKafkaProducer()
	ddSvc := p.container.getDeviceDefinitionService()

	if len(os.Args[2:]) != 2 {
		p.logger.Fatal().Msg("Requires aftermarket_token_id vehicle_token_id")
	}

	amToken, ok := new(big.Int).SetString(os.Args[2], 10)
	if !ok {
		p.logger.Fatal().Msgf("Couldn't parse aftermarket_token_id %q", os.Args[2])
	}

	vToken, ok := new(big.Int).SetString(os.Args[3], 10)
	if !ok {
		p.logger.Fatal().Msgf("Couldn't parse vehicle_token_id %q", os.Args[3])
	}

	am, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.TokenID.EQ(utils.BigToDecimal(amToken)),
	).One(context.TODO(), p.container.dbs().Reader)
	if err != nil {
		p.logger.Fatal().Msgf("Can't find aftermarket device %d.", am.TokenID)
	}

	p.logger.Info().Msgf("Attempting to web2 pair am device %s to vehicle %s.", amToken, vToken)

	autoPiIngest := services.NewIngestRegistrar(producer)
	eventService := services.NewEventService(&p.logger, &p.settings, producer)

	integ := genericad.NewIntegration(p.pdb.DBS, ddSvc, autoPiIngest, eventService, &p.logger)
	if err := integ.Pair(ctx, amToken, vToken); err != nil {
		p.logger.Fatal().Err(err).Msg("Pairing failure.")
	}

	p.logger.Info().Msg("Pairing success.")

	return subcommands.ExitSuccess
}
