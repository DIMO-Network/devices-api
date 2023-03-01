package main

import (
	"context"
	"flag"

	"github.com/Shopify/sarama"

	"math/big"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type web2PairCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
	producer sarama.SyncProducer
	ddSvc    services.DeviceDefinitionService
}

func (*web2PairCmd) Name() string     { return "web2-pair" }
func (*web2PairCmd) Synopsis() string { return "web2-pair args to stdout." }
func (*web2PairCmd) Usage() string {
	return `web2-pair [] <some text>:
	web2-pair args.
  `
}

func (p *web2PairCmd) SetFlags(f *flag.FlagSet) {

}

func (p *web2PairCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if len(f.Args()[2:]) != 2 {
		p.logger.Fatal().Msg("Requires aftermarket_token_id vehicle_token_id")
	}

	amToken, ok := new(big.Int).SetString(f.Args()[2], 10)
	if !ok {
		p.logger.Fatal().Msgf("Couldn't parse aftermarket_token_id %q", f.Args()[2])
	}

	vToken, ok := new(big.Int).SetString(f.Args()[3], 10)
	if !ok {
		p.logger.Fatal().Msgf("Couldn't parse vehicle_token_id %q", f.Args()[3])
	}

	p.logger.Info().Msgf("Attempting to web2 pair am device %s to vehicle %s.", amToken, vToken)

	autoPiSvc := services.NewAutoPiAPIService(&p.settings, p.pdb.DBS)
	autoPiTaskService := services.NewAutoPiTaskService(&p.settings, autoPiSvc, p.pdb.DBS, p.logger)
	autoPiIngest := services.NewIngestRegistrar(services.AutoPi, p.producer)
	eventService := services.NewEventService(&p.logger, &p.settings, p.producer)
	deviceDefinitionRegistrar := services.NewDeviceDefinitionRegistrar(p.producer, &p.settings)
	hardwareTemplateService := autopi.NewHardwareTemplateService(autoPiSvc, p.pdb.DBS, &p.logger)

	i := autopi.NewIntegration(p.pdb.DBS, p.ddSvc, autoPiSvc, autoPiTaskService, autoPiIngest, eventService, deviceDefinitionRegistrar, hardwareTemplateService, &p.logger)

	err := i.Pair(ctx, amToken, vToken)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Pairing failure.")
	}

	p.logger.Info().Msg("Pairing success.")

	return subcommands.ExitSuccess
}
